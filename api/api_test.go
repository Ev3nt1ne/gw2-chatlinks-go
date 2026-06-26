package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL}
}

func TestResolveSkillName(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/skills/3876" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("v") != schemaVersion {
			t.Errorf("missing/wrong schema version query param: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"name": "Skelk Venom"}`))
	})

	name, err := client.ResolveSkillName(context.Background(), 3876)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Skelk Venom" {
		t.Errorf("name = %q, want Skelk Venom", name)
	}
}

func TestResolveTraitName(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"name": "Some Trait"}`))
	})
	name, err := client.ResolveTraitName(context.Background(), 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Some Trait" {
		t.Errorf("name = %q, want Some Trait", name)
	}
}

func TestResolveItemName(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"name": "Some Item"}`))
	})
	name, err := client.ResolveItemName(context.Background(), 456)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Some Item" {
		t.Errorf("name = %q, want Some Item", name)
	}
}

func TestResolveSpecializationName(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"name": "Some Spec"}`))
	})
	name, err := client.ResolveSpecializationName(context.Background(), 789)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Some Spec" {
		t.Errorf("name = %q, want Some Spec", name)
	}
}

func TestPaletteIDToSkillID_Found(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/professions/Thief" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"skills_by_palette": [[3876, 12345], [9999, 1]]}`))
	})
	skillID, ok, err := client.PaletteIDToSkillID(context.Background(), "Thief", 3876)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || skillID != 12345 {
		t.Errorf("got (%d, %v), want (12345, true)", skillID, ok)
	}
}

func TestPaletteIDToSkillID_NotFound(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"skills_by_palette": [[1, 2]]}`))
	})
	_, ok, err := client.PaletteIDToSkillID(context.Background(), "Thief", 9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected ok=false for an unmatched palette id")
	}
}

func TestPaletteIDToSkillID_ZeroIsUnsetSlot(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make an HTTP request for paletteID == 0")
	})
	_, ok, err := client.PaletteIDToSkillID(context.Background(), "Thief", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected ok=false for paletteID 0")
	}
}

func TestGet_NonOKStatus(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	if _, err := client.ResolveSkillName(context.Background(), 1); err == nil {
		t.Error("expected error for non-200 response, got nil")
	}
}

func TestGet_InvalidJSON(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})
	if _, err := client.ResolveSkillName(context.Background(), 1); err == nil {
		t.Error("expected error for invalid JSON response, got nil")
	}
}

func TestResolveSkillNames_EmptyInputNoRequest(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make an HTTP request for empty input")
	})
	names, err := client.ResolveSkillNames(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("names = %v, want empty map", names)
	}
}

func TestResolveSkillNames_AllRecognized(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/skills" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("ids"); got != "1110,1115" {
			t.Errorf("ids query = %q, want 1110,1115", got)
		}
		_, _ = w.Write([]byte(`[{"id":1110,"name":"Throw Gunk"},{"id":1115,"name":"Branch Leap"}]`))
	})
	names, err := client.ResolveSkillNames(context.Background(), []int{1110, 1115})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if names[1110] != "Throw Gunk" || names[1115] != "Branch Leap" {
		t.Errorf("names = %v, want {1110: Throw Gunk, 1115: Branch Leap}", names)
	}
}

// TestResolveSkillNames_PartialContent mirrors the live API behavior
// verified empirically 2026-06-26: a batch with one unrecognized id among
// recognized ones returns 206 with only the recognized entries in the body.
func TestResolveSkillNames_PartialContent(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Warning", `299 api.guildwars2.com "Unknown id 999999999999"`)
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte(`[{"id":1110,"name":"Throw Gunk"}]`))
	})
	names, err := client.ResolveSkillNames(context.Background(), []int{1110, 999999999999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 || names[1110] != "Throw Gunk" {
		t.Errorf("names = %v, want exactly {1110: Throw Gunk}", names)
	}
	if _, ok := names[999999999999]; ok {
		t.Error("unrecognized id should be absent from the result map, not present")
	}
}

// TestResolveSkillNames_AllUnrecognized mirrors the live API behavior
// verified empirically: a batch where every id is unrecognized returns 404
// with a non-array body ({"text":"all ids provided are invalid"}) — this
// must be treated as zero results, not an error.
func TestResolveSkillNames_AllUnrecognized(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"text":"all ids provided are invalid"}`))
	})
	names, err := client.ResolveSkillNames(context.Background(), []int{999999999999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("names = %v, want empty map", names)
	}
}

func TestResolveSkillNames_ChunksAt200(t *testing.T) {
	var requestCount int
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		ids := r.URL.Query().Get("ids")
		var entities []string
		for _, id := range strings.Split(ids, ",") {
			entities = append(entities, fmt.Sprintf(`{"id":%s,"name":"n%s"}`, id, id))
		}
		_, _ = w.Write([]byte("[" + strings.Join(entities, ",") + "]"))
	})

	ids := make([]int, 250)
	for i := range ids {
		ids[i] = i + 1
	}
	names, err := client.ResolveSkillNames(context.Background(), ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 2 {
		t.Errorf("requestCount = %d, want 2 (250 ids chunked at 200)", requestCount)
	}
	if len(names) != 250 {
		t.Errorf("len(names) = %d, want 250", len(names))
	}
}

func TestResolveSkillNames_ServerError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := client.ResolveSkillNames(context.Background(), []int{1}); err == nil {
		t.Error("expected error for a 500 response, got nil")
	}
}

func TestResolveSkillNames_InvalidJSON(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})
	if _, err := client.ResolveSkillNames(context.Background(), []int{1}); err == nil {
		t.Error("expected error for invalid JSON response, got nil")
	}
}

func TestPaletteIDsToSkillIDs_PropagatesProfessionFetchError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := client.PaletteIDsToSkillIDs(context.Background(), "Thief", []int{1}); err == nil {
		t.Error("expected error to propagate from the underlying profession fetch, got nil")
	}
}

func TestResolveTraitNamesItemNamesSpecializationNames_Collections(t *testing.T) {
	tests := []struct {
		name       string
		call       func(c *Client) (map[int]string, error)
		wantPrefix string
	}{
		{"traits", func(c *Client) (map[int]string, error) { return c.ResolveTraitNames(context.Background(), []int{1}) }, "/traits"},
		{"items", func(c *Client) (map[int]string, error) { return c.ResolveItemNames(context.Background(), []int{1}) }, "/items"},
		{"specializations", func(c *Client) (map[int]string, error) {
			return c.ResolveSpecializationNames(context.Background(), []int{1})
		}, "/specializations"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantPrefix {
					t.Errorf("path = %s, want %s", r.URL.Path, tt.wantPrefix)
				}
				_, _ = w.Write([]byte(`[{"id":1,"name":"X"}]`))
			})
			names, err := tt.call(client)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if names[1] != "X" {
				t.Errorf("names = %v, want {1: X}", names)
			}
		})
	}
}

func TestPaletteIDsToSkillIDs_SingleFetchServesMultipleIDs(t *testing.T) {
	var requestCount int
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/professions/Thief" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"skills_by_palette": [[3876, 12345], [9999, 1], [42, 7]]}`))
	})
	skillIDs, err := client.PaletteIDsToSkillIDs(context.Background(), "Thief", []int{3876, 42, 0, 555})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("requestCount = %d, want exactly 1 fetch regardless of how many palette ids were requested", requestCount)
	}
	want := map[int]int{3876: 12345, 42: 7}
	if len(skillIDs) != len(want) || skillIDs[3876] != 12345 || skillIDs[42] != 7 {
		t.Errorf("skillIDs = %v, want %v (0 and unmatched 555 absent)", skillIDs, want)
	}
	if _, ok := skillIDs[0]; ok {
		t.Error("palette id 0 (unset slot) should be absent from the result map")
	}
	if _, ok := skillIDs[555]; ok {
		t.Error("unmatched palette id should be absent from the result map")
	}
}

func TestPaletteIDsToSkillIDs_AllZeroNoRequest(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make an HTTP request when every palette id is 0")
	})
	skillIDs, err := client.PaletteIDsToSkillIDs(context.Background(), "Thief", []int{0, 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skillIDs) != 0 {
		t.Errorf("skillIDs = %v, want empty map", skillIDs)
	}
}

func TestUserAgent_DefaultWhenUnset(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != defaultUserAgent {
			t.Errorf("User-Agent = %q, want default %q", got, defaultUserAgent)
		}
		_, _ = w.Write([]byte(`{"name":"X"}`))
	})
	if _, err := client.ResolveSkillName(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserAgent_ExplicitOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "my-app/1.0" {
			t.Errorf("User-Agent = %q, want my-app/1.0", got)
		}
		_, _ = w.Write([]byte(`{"name":"X"}`))
	}))
	t.Cleanup(srv.Close)
	client := &Client{BaseURL: srv.URL, UserAgent: "my-app/1.0"}
	if _, err := client.ResolveSkillName(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
