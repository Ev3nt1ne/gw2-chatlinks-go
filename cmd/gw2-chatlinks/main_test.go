package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Ev3nt1ne/gw2-chatlinks-go/api"
	"github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"
)

const thiefSample = "[&DQUAAAAAAAAkDyQPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACLwBVAAA=]"
const rangerPetSample = "[&DQQAAAAAAAB5AAAAAAAAAAAAAAAAAAAAAAAAADA7FD8AAAAAAAAAAAAAAAACIwAyAAA=]"

func TestRun_BuildTemplate_NoResolve(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, thiefSample, options{}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "type: build_template") {
		t.Errorf("output missing type line: %s", out)
	}
	if !strings.Contains(out, "profession: Thief") {
		t.Errorf("output missing profession line: %s", out)
	}
	if !strings.Contains(out, "palette=3876") {
		t.Errorf("output missing heal palette: %s", out)
	}
}

func TestRun_ItemLink_NoResolve(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, "[&AgEAAAA=]", options{}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "type: item") {
		t.Errorf("output missing type line: %s", out)
	}
	if !strings.Contains(out, "id: 0") {
		t.Errorf("output missing id line: %s", out)
	}
}

func TestRun_AchievementAndMapLinks_NoResolve(t *testing.T) {
	// Real-world testing (chatlinks/testdata) confirmed achievement (0x0E)
	// and map (0x04) links are structurally identical to skill/item/recipe
	// links and already decode correctly via DecodeSimpleIDLink — this
	// just confirms the CLI actually routes to it instead of falling into
	// the generic "not implemented" branch.
	achievementCode, err := chatlinks.EncodeSimpleIDLink(chatlinks.SimpleIDLink{LinkType: "achievement", ID: 1})
	if err != nil {
		t.Fatalf("unexpected error building test fixture: %v", err)
	}
	mapCode, err := chatlinks.EncodeSimpleIDLink(chatlinks.SimpleIDLink{LinkType: "map", ID: 99})
	if err != nil {
		t.Fatalf("unexpected error building test fixture: %v", err)
	}

	for _, tt := range []struct{ code, wantType string }{
		{achievementCode, "achievement"},
		{mapCode, "map"},
	} {
		var buf bytes.Buffer
		if err := run(&buf, tt.code, options{}, &api.Client{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "type: "+tt.wantType) {
			t.Errorf("output missing type line for %s: %s", tt.wantType, out)
		}
		if strings.Contains(out, "not implemented yet") {
			t.Errorf("%s link fell through to the unimplemented-type branch: %s", tt.wantType, out)
		}
	}
}

func TestRun_UnimplementedType(t *testing.T) {
	var buf bytes.Buffer
	// header 0x01 = coin, no decoder implemented.
	if err := run(&buf, "[&AQAAAAA=]", options{}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "decoder for this type not implemented yet") {
		t.Errorf("expected fallback message, got: %s", buf.String())
	}
}

func TestRun_InvalidCode(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, "[&not-valid!!]", options{}, &api.Client{}); err == nil {
		t.Error("expected error for invalid chat link, got nil")
	}
}

func TestRun_ResolveUnsupportedForRecipe(t *testing.T) {
	// Recipe header (0x09) with a 3-byte id; --resolve has no recipe-name
	// endpoint wired up, so this should surface an error rather than hang
	// or silently no-op.
	var buf bytes.Buffer
	err := run(&buf, "[&CQEAAAA=]", options{resolve: true}, &api.Client{})
	if err == nil {
		t.Error("expected error for --resolve on a recipe link, got nil")
	}
}

func TestRun_RangerPetsAndWeapons_NoResolve(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, rangerPetSample, options{}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ranger_pets:") {
		t.Errorf("output missing ranger_pets line: %s", out)
	}
	if !strings.Contains(out, "weapon: Longbow") || !strings.Contains(out, "weapon: Greatsword") {
		t.Errorf("output missing weapon lines: %s", out)
	}
}

func TestRun_RevenantLegends_NoResolve(t *testing.T) {
	code, err := chatlinks.EncodeBuildTemplate(chatlinks.BuildTemplate{
		ProfessionID: 9, // Revenant
		RevenantLegends: &chatlinks.RevenantLegends{
			TerrestrialActive:   1,
			TerrestrialInactive: 2,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error building test fixture: %v", err)
	}

	var buf bytes.Buffer
	if err := run(&buf, code, options{}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "active_terrestrial=Legendary Dragon Stance") {
		t.Errorf("output missing resolved legend name: %s", out)
	}
	if !strings.Contains(out, "inactive_aquatic=(none)") {
		t.Errorf("output missing (none) for an unset legend slot: %s", out)
	}
}

func TestRun_RevenantLegends_UnknownCode(t *testing.T) {
	code, err := chatlinks.EncodeBuildTemplate(chatlinks.BuildTemplate{
		ProfessionID:    9,
		RevenantLegends: &chatlinks.RevenantLegends{TerrestrialActive: 99},
	})
	if err != nil {
		t.Fatalf("unexpected error building test fixture: %v", err)
	}

	var buf bytes.Buffer
	if err := run(&buf, code, options{}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "active_terrestrial=unknown(99)") {
		t.Errorf("output missing unknown-legend fallback: %s", buf.String())
	}
}

func TestRun_SkillOverride_NoResolve(t *testing.T) {
	code, err := chatlinks.EncodeBuildTemplate(chatlinks.BuildTemplate{
		ProfessionID:     2, // Warrior
		SkillOverrideIDs: []int{12345},
	})
	if err != nil {
		t.Fatalf("unexpected error building test fixture: %v", err)
	}

	var buf bytes.Buffer
	if err := run(&buf, code, options{}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "skill_override: skill_id=12345") {
		t.Errorf("output missing skill_override line: %s", buf.String())
	}
}
