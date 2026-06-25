package chatlinks

import (
	"encoding/json"
	"os"
	"testing"
)

// fixture mirrors one entry of testdata/realworld_fixtures.json. Code is
// only present for types where the live GW2 API exposes a chat_link field
// directly (item, skill, recipe) — see testdata/gather_fixtures.py. For
// achievement/map, only a real id is available; those fixtures verify
// self-consistent round-tripping of a real id, not a match against an
// independently-published code string.
type fixture struct {
	Type string `json:"type"`
	Code string `json:"code,omitempty"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func loadFixtures(t *testing.T) []fixture {
	t.Helper()
	data, err := os.ReadFile("testdata/realworld_fixtures.json")
	if err != nil {
		t.Fatalf("reading testdata/realworld_fixtures.json: %v", err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("parsing testdata/realworld_fixtures.json: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("testdata/realworld_fixtures.json contained no fixtures")
	}
	return fixtures
}

// TestRealWorldFixtures_GroundTruth validates every fixture that carries a
// real, externally-sourced chat_link string (item/skill/recipe, pulled
// straight from the live GW2 API — see testdata/gather_fixtures.py): the
// decoded ID must match the API's own id for that object, and re-encoding
// must reproduce the exact original bytes.
func TestRealWorldFixtures_GroundTruth(t *testing.T) {
	fixtures := loadFixtures(t)
	tested := 0
	for _, fx := range fixtures {
		if fx.Code == "" {
			continue
		}
		fx := fx
		t.Run(fx.Type+"/"+fx.Name, func(t *testing.T) {
			link, err := DecodeSimpleIDLink(fx.Code)
			if err != nil {
				t.Fatalf("DecodeSimpleIDLink(%q): %v", fx.Code, err)
			}
			if link.LinkType != fx.Type {
				t.Errorf("LinkType = %q, want %q", link.LinkType, fx.Type)
			}
			if link.ID != fx.ID {
				t.Errorf("ID = %d, want %d (name=%q)", link.ID, fx.ID, fx.Name)
			}
			if fx.Type == "item" && link.Quantity != 1 {
				t.Errorf("Quantity = %d, want 1 (the API's chat_link always encodes quantity 1)", link.Quantity)
			}

			reencoded, err := EncodeSimpleIDLink(link)
			if err != nil {
				t.Fatalf("EncodeSimpleIDLink: %v", err)
			}
			wantRaw, err := DecodeRaw(fx.Code)
			if err != nil {
				t.Fatalf("DecodeRaw(%q): %v", fx.Code, err)
			}
			gotRaw, err := DecodeRaw(reencoded)
			if err != nil {
				t.Fatalf("DecodeRaw(reencoded): %v", err)
			}
			if string(gotRaw) != string(wantRaw) {
				t.Errorf("round-trip mismatch: got %x, want %x", gotRaw, wantRaw)
			}
		})
		tested++
	}
	if tested == 0 {
		t.Fatal("no fixtures had a code field to test against — testdata may be stale or malformed")
	}
	t.Logf("validated %d real (code, id) pairs sourced directly from the live GW2 API", tested)
}

// TestRealWorldFixtures_SelfConsistency covers fixture types where the
// public API doesn't expose a chat_link field (achievement, map points of
// interest) — see testdata/gather_fixtures.py. These only have a real id,
// so this checks that encoding it and decoding the result recovers the
// same id, not a match against an independently-published code.
func TestRealWorldFixtures_SelfConsistency(t *testing.T) {
	fixtures := loadFixtures(t)
	tested := 0
	for _, fx := range fixtures {
		if fx.Code != "" {
			continue
		}
		fx := fx
		t.Run(fx.Type+"/"+fx.Name, func(t *testing.T) {
			code, err := EncodeSimpleIDLink(SimpleIDLink{LinkType: fx.Type, ID: fx.ID})
			if err != nil {
				t.Fatalf("EncodeSimpleIDLink: %v", err)
			}
			link, err := DecodeSimpleIDLink(code)
			if err != nil {
				t.Fatalf("DecodeSimpleIDLink(%q): %v", code, err)
			}
			if link.ID != fx.ID {
				t.Errorf("round-trip ID = %d, want %d (name=%q)", link.ID, fx.ID, fx.Name)
			}
		})
		tested++
	}
	if tested == 0 {
		t.Fatal("no fixtures lacked a code field — testdata may be stale or malformed")
	}
	t.Logf("round-tripped %d real ids (achievement/map) with no external code to cross-check", tested)
}
