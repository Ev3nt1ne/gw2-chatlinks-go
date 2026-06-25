package api

import (
	"context"
	"net/http"
	"net/http/httptest"
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
