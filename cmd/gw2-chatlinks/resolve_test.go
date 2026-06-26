package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Ev3nt1ne/gw2-chatlinks-go/api"
	"github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"
)

// TestRun_BuildTemplate_Resolve_UsesBatchAPI is the first test to cover the
// CLI's --resolve path against a mock server. It proves the build-template
// resolve flow makes a small, constant number of requests (1 profession-doc
// fetch + 1 batched skill-name lookup) rather than one request per skill
// slot/override, regardless of how many slots/overrides there are.
func TestRun_BuildTemplate_Resolve_UsesBatchAPI(t *testing.T) {
	skillNames := map[int]string{1001: "Heal Thing", 1002: "Util Thing", 1003: "Elite Thing", 400: "Override Thing"}

	var requestCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch r.URL.Path {
		case "/professions/Thief":
			_, _ = w.Write([]byte(`{"skills_by_palette": [[100,1001],[200,1002],[300,1003]]}`))
		case "/skills":
			ids := strings.Split(r.URL.Query().Get("ids"), ",")
			var entries []string
			for _, idStr := range ids {
				id, err := strconv.Atoi(idStr)
				if err != nil {
					t.Fatalf("non-integer id in batch request: %q", idStr)
				}
				name, ok := skillNames[id]
				if !ok {
					t.Fatalf("unexpected skill id in batch request: %d", id)
				}
				entries = append(entries, fmt.Sprintf(`{"id":%d,"name":%q}`, id, name))
			}
			_, _ = w.Write([]byte("[" + strings.Join(entries, ",") + "]"))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	code, err := chatlinks.EncodeBuildTemplate(chatlinks.BuildTemplate{
		ProfessionID: 5, // Thief
		// heal_terrestrial == heal_aquatic (duplicate palette id, as real
		// build templates commonly have) plus 2 more distinct slots.
		SkillPaletteIDs:  [10]int{100, 100, 200, 0, 0, 0, 0, 0, 0, 300},
		SkillOverrideIDs: []int{400},
	})
	if err != nil {
		t.Fatalf("unexpected error building test fixture: %v", err)
	}

	var buf bytes.Buffer
	client := &api.Client{BaseURL: srv.URL}
	if err := run(&buf, code, options{resolve: true}, client); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("requestCount = %d, want exactly 2 (1 profession fetch + 1 batched skill lookup) regardless of slot/override count", requestCount)
	}

	out := buf.String()
	for _, want := range []string{"Heal Thing", "Util Thing", "Elite Thing", "Override Thing"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing resolved name %q:\n%s", want, out)
		}
	}
}

// TestRun_SimpleIDLink_Resolve covers --resolve for skill/trait/item links,
// previously untested (only the recipe-unsupported error branch had
// coverage). Confirms printSimpleIDLink actually calls the right Resolve*
// method per link type via the injected client.
func TestRun_SimpleIDLink_Resolve(t *testing.T) {
	tests := []struct {
		linkType string
		wantPath string
		wantName string
	}{
		{"skill", "/skills/1", "Real Skill"},
		{"trait", "/traits/1", "Real Trait"},
		{"item", "/items/1", "Real Item"},
	}
	for _, tt := range tests {
		t.Run(tt.linkType, func(t *testing.T) {
			code, err := chatlinks.EncodeSimpleIDLink(chatlinks.SimpleIDLink{LinkType: tt.linkType, ID: 1})
			if err != nil {
				t.Fatalf("unexpected error building test fixture: %v", err)
			}

			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = fmt.Fprintf(w, `{"name":%q}`, tt.wantName)
			}))
			t.Cleanup(srv.Close)

			var buf bytes.Buffer
			client := &api.Client{BaseURL: srv.URL}
			if err := run(&buf, code, options{resolve: true}, client); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotPath != tt.wantPath {
				t.Errorf("request path = %q, want %q", gotPath, tt.wantPath)
			}
			if !strings.Contains(buf.String(), tt.wantName) {
				t.Errorf("output missing resolved name %q:\n%s", tt.wantName, buf.String())
			}
		})
	}
}

// TestRun_BuildTemplate_Resolve_UnresolvedPaletteShowsPlaceholder confirms
// the "?" placeholder convention survives the move to the batch API: a
// palette id absent from the profession document still shows "?", exactly
// as the old single-ID PaletteIDToSkillID "ok=false" path did.
func TestRun_BuildTemplate_Resolve_UnresolvedPaletteShowsPlaceholder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"skills_by_palette": []}`))
	}))
	t.Cleanup(srv.Close)

	code, err := chatlinks.EncodeBuildTemplate(chatlinks.BuildTemplate{
		ProfessionID:    5,
		SkillPaletteIDs: [10]int{999, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})
	if err != nil {
		t.Fatalf("unexpected error building test fixture: %v", err)
	}

	var buf bytes.Buffer
	client := &api.Client{BaseURL: srv.URL}
	if err := run(&buf, code, options{resolve: true}, client); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "heal_terrestrial: ? (palette=999, skill_id=0)") {
		t.Errorf("output missing '?' placeholder for an unresolved palette id:\n%s", buf.String())
	}
}
