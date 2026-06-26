package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"
)

// fullBuild is a Thief build with two distinct skill palette slots (one
// duplicated across terrestrial/aquatic), one skill override, and two
// specializations set — enough to exercise all three resolve categories.
func fullBuild() chatlinks.BuildTemplate {
	return chatlinks.BuildTemplate{
		Profession:       "Thief",
		SkillPaletteIDs:  [10]int{100, 100, 200, 0, 0, 0, 0, 0, 0, 300},
		SkillOverrideIDs: []int{400},
		Specializations: [3]chatlinks.SpecializationChoice{
			{SpecializationID: 7},
			{SpecializationID: 0}, // unset — must not be requested
			{SpecializationID: 8},
		},
	}
}

func TestResolveBuildTemplate_FullBuildUsesThreeRequests(t *testing.T) {
	var profCalls, skillCalls, specCalls int
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/professions/Thief":
			profCalls++
			_, _ = w.Write([]byte(`{"skills_by_palette": [[100,1001],[200,1002],[300,1003]]}`))
		case "/skills":
			skillCalls++
			// palette-derived 1001,1002,1003 + override 400, deduped & sorted.
			if got := r.URL.Query().Get("ids"); got != "400,1001,1002,1003" {
				t.Errorf("skills ids = %q, want 400,1001,1002,1003", got)
			}
			_, _ = w.Write([]byte(`[{"id":400,"name":"Override"},{"id":1001,"name":"Heal"},` +
				`{"id":1002,"name":"Util"},{"id":1003,"name":"Elite"}]`))
		case "/specializations":
			specCalls++
			if got := r.URL.Query().Get("ids"); got != "7,8" {
				t.Errorf("spec ids = %q, want 7,8", got)
			}
			_, _ = w.Write([]byte(`[{"id":7,"name":"Acrobatics"},{"id":8,"name":"Trickery"}]`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	res, err := client.ResolveBuildTemplate(context.Background(), fullBuild())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profCalls != 1 || skillCalls != 1 || specCalls != 1 {
		t.Errorf("requests = prof:%d skill:%d spec:%d, want exactly 1 each", profCalls, skillCalls, specCalls)
	}
	if res.PaletteToSkillID[100] != 1001 || res.PaletteToSkillID[200] != 1002 || res.PaletteToSkillID[300] != 1003 {
		t.Errorf("PaletteToSkillID = %v", res.PaletteToSkillID)
	}
	for id, want := range map[int]string{400: "Override", 1001: "Heal", 1002: "Util", 1003: "Elite"} {
		if res.SkillNames[id] != want {
			t.Errorf("SkillNames[%d] = %q, want %q", id, res.SkillNames[id], want)
		}
	}
	if res.SpecializationNames[7] != "Acrobatics" || res.SpecializationNames[8] != "Trickery" {
		t.Errorf("SpecializationNames = %v", res.SpecializationNames)
	}
}

func TestResolveBuildTemplate_EmptyBuildMakesNoRequest(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected HTTP request for an empty build: %s", r.URL.Path)
	})
	res, err := client.ResolveBuildTemplate(context.Background(), chatlinks.BuildTemplate{Profession: "Thief"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.PaletteToSkillID) != 0 || len(res.SkillNames) != 0 || len(res.SpecializationNames) != 0 {
		t.Errorf("expected all-empty result, got %+v", res)
	}
}

// TestResolveBuildTemplate_UnrecognizedIDsAbsentNotError confirms that IDs the
// API doesn't know (206/404 from the batch endpoints) come back as missing
// map keys, not as an error.
func TestResolveBuildTemplate_UnrecognizedIDsAbsentNotError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/professions/Thief":
			_, _ = w.Write([]byte(`{"skills_by_palette": [[100,1001]]}`)) // 200 unmatched
		case "/skills":
			_, _ = w.Write([]byte(`[{"id":1001,"name":"Heal"}]`)) // override 400 absent
		case "/specializations":
			w.WriteHeader(http.StatusNotFound) // all-invalid batch
			_, _ = w.Write([]byte(`{"text":"all ids provided are invalid"}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	res, err := client.ResolveBuildTemplate(context.Background(), fullBuild())
	if err != nil {
		t.Fatalf("unexpected error for unrecognized ids: %v", err)
	}
	if _, ok := res.PaletteToSkillID[200]; ok {
		t.Error("palette 200 should be absent (unmatched in profession doc)")
	}
	if _, ok := res.SkillNames[400]; ok {
		t.Error("override skill 400 should be absent (unrecognized)")
	}
	if len(res.SpecializationNames) != 0 {
		t.Errorf("specs should be empty for an all-invalid batch, got %v", res.SpecializationNames)
	}
}

// TestResolveBuildTemplate_PartialFailureIsAggregated confirms one category
// failing (a 500) doesn't blank out the others, and the error is surfaced.
func TestResolveBuildTemplate_PartialFailureIsAggregated(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/professions/Thief":
			_, _ = w.Write([]byte(`{"skills_by_palette": [[100,1001],[200,1002],[300,1003]]}`))
		case "/skills":
			_, _ = w.Write([]byte(`[{"id":1001,"name":"Heal"},{"id":1002,"name":"Util"},` +
				`{"id":1003,"name":"Elite"},{"id":400,"name":"Override"}]`))
		case "/specializations":
			w.WriteHeader(http.StatusInternalServerError) // this category fails hard
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	res, err := client.ResolveBuildTemplate(context.Background(), fullBuild())
	if err == nil {
		t.Fatal("expected a non-nil aggregated error when the spec lookup 500s")
	}
	if !strings.Contains(err.Error(), "specializations") {
		t.Errorf("error should mention the failing /specializations path, got: %v", err)
	}
	// Skills still resolved despite the spec failure.
	if res.SkillNames[1001] != "Heal" || res.SkillNames[400] != "Override" {
		t.Errorf("skills should still resolve despite spec failure, got %v", res.SkillNames)
	}
	if len(res.SpecializationNames) != 0 {
		t.Errorf("failed spec category should yield an empty (non-nil) map, got %v", res.SpecializationNames)
	}
}

// TestResolveBuildTemplate_RateLimitErrorPropagates confirms a 429 surfaces as
// the typed *RateLimitError through the aggregated error.
func TestResolveBuildTemplate_RateLimitErrorPropagates(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	_, err := client.ResolveBuildTemplate(context.Background(), fullBuild())
	if err == nil {
		t.Fatal("expected an error when every category is rate limited")
	}
	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Errorf("expected a *RateLimitError in the aggregated error, got: %v", err)
	}
}

// TestResolveBuildTemplate_RevenantNilSafe confirms a Revenant build (with a
// RevenantLegends pointer set) resolves without panicking; the orchestrator
// doesn't touch legend fields, but the build is a realistic shape to pass in.
func TestResolveBuildTemplate_RevenantNilSafe(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/professions/Revenant":
			_, _ = w.Write([]byte(`{"skills_by_palette": [[100,9001]]}`))
		case "/skills":
			_, _ = w.Write([]byte(`[{"id":9001,"name":"Empowering Misery"}]`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	bt := chatlinks.BuildTemplate{
		Profession:      "Revenant",
		SkillPaletteIDs: [10]int{100, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		RevenantLegends: &chatlinks.RevenantLegends{TerrestrialActive: 1, TerrestrialInactive: 2},
	}
	res, err := client.ResolveBuildTemplate(context.Background(), bt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SkillNames[9001] != "Empowering Misery" {
		t.Errorf("SkillNames[9001] = %q, want Empowering Misery", res.SkillNames[9001])
	}
}
