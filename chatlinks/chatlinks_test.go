package chatlinks

import (
	"testing"
)

// These 4 codes are the only real-world build template samples available:
// the mandatory level-2 build templates for Thief/Elementalist/Engineer and
// the Ranger pet-disable template, pulled from the Heroes Ascent 3rd-edition
// ruleset doc. All four have zero specializations chosen (correct — GW2
// doesn't unlock trait lines until ~level 11, and these are level-2
// templates), so the Adept/Master/Grandmaster trait-tier bit-packing logic
// is NOT exercised by any of them and remains unverified against real data.
const (
	thiefSample         = "[&DQUAAAAAAAAkDyQPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACLwBVAAA=]"
	elementalistSample  = "[&DQYAAAAAAAAnDycPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABWQAA]"
	engineerSample      = "[&DQMAAAAAAAAqDyoPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACVQAzAAA=]"
	rangerPetSample     = "[&DQQAAAAAAAB5AAAAAAAAAAAAAAAAAAAAAAAAADA7FD8AAAAAAAAAAAAAAAACIwAyAAA=]"
)

func TestDecodeBuildTemplate_Thief(t *testing.T) {
	bt, err := DecodeBuildTemplate(thiefSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.Profession != "Thief" {
		t.Errorf("Profession = %q, want Thief", bt.Profession)
	}
	if bt.SkillPaletteIDs[0] != 3876 || bt.SkillPaletteIDs[1] != 3876 {
		t.Errorf("heal palette ids = %v, want [3876 3876 ...]", bt.SkillPaletteIDs)
	}
	for i, s := range bt.Specializations {
		if s.Adept != 0 || s.Master != 0 || s.Grandmaster != 0 {
			t.Errorf("specialization[%d] = %+v, want all-zero trait tiers", i, s)
		}
	}
}

func TestDecodeBuildTemplate_Elementalist(t *testing.T) {
	bt, err := DecodeBuildTemplate(elementalistSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.Profession != "Elementalist" {
		t.Errorf("Profession = %q, want Elementalist", bt.Profession)
	}
	if bt.SkillPaletteIDs[0] != 3879 || bt.SkillPaletteIDs[1] != 3879 {
		t.Errorf("heal palette ids = %v, want [3879 3879 ...]", bt.SkillPaletteIDs)
	}
}

func TestDecodeBuildTemplate_Engineer(t *testing.T) {
	bt, err := DecodeBuildTemplate(engineerSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.Profession != "Engineer" {
		t.Errorf("Profession = %q, want Engineer", bt.Profession)
	}
	if bt.SkillPaletteIDs[0] != 3882 || bt.SkillPaletteIDs[1] != 3882 {
		t.Errorf("heal palette ids = %v, want [3882 3882 ...]", bt.SkillPaletteIDs)
	}
}

func TestDecodeBuildTemplate_RangerPets(t *testing.T) {
	bt, err := DecodeBuildTemplate(rangerPetSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.Profession != "Ranger" {
		t.Errorf("Profession = %q, want Ranger", bt.Profession)
	}
	want := []int{48, 59, 20, 63}
	if len(bt.RangerPetIDs) != len(want) {
		t.Fatalf("RangerPetIDs = %v, want %v", bt.RangerPetIDs, want)
	}
	for i := range want {
		if bt.RangerPetIDs[i] != want[i] {
			t.Errorf("RangerPetIDs[%d] = %d, want %d", i, bt.RangerPetIDs[i], want[i])
		}
	}
}

func TestDecodeBuildTemplate_WrongHeader(t *testing.T) {
	// An item link (header 0x02: item byte, then a 0-quantity byte, then a
	// 3-byte little-endian id of 0), not a build template.
	itemLink := "[&AgEAAAA=]"
	if _, err := DecodeBuildTemplate(itemLink); err == nil {
		t.Error("expected error decoding a non-build-template link as a build template, got nil")
	}
}

func TestDecodeBuildTemplate_TooShort(t *testing.T) {
	// Valid base64, valid 0x0D header, but far too short to hold the fixed
	// fields a build template requires.
	tooShort := WrapLink("DQUAAAA=")
	if _, err := DecodeBuildTemplate(tooShort); err == nil {
		t.Error("expected error decoding a truncated build template, got nil")
	}
}

func TestDecodeRaw_StripsWrapperAndFixesPadding(t *testing.T) {
	withWrapper, err := DecodeRaw(thiefSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bare := StripLink(thiefSample)
	withoutWrapper, err := DecodeRaw(bare)
	if err != nil {
		t.Fatalf("unexpected error decoding bare payload: %v", err)
	}
	if string(withWrapper) != string(withoutWrapper) {
		t.Error("decoding with and without the [&...] wrapper produced different bytes")
	}
}

func TestDecodeRaw_InvalidBase64(t *testing.T) {
	if _, err := DecodeRaw("[&not-valid-base64!!]"); err == nil {
		t.Error("expected error for invalid base64 payload, got nil")
	}
}

func TestDecodeRaw_Empty(t *testing.T) {
	if _, err := DecodeRaw("[&]"); err == nil {
		t.Error("expected error for empty payload, got nil")
	}
}

func TestHeaderType(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		want string
	}{
		{"build template", []byte{0x0D}, "build_template"},
		{"skill", []byte{0x06}, "skill"},
		{"unknown byte", []byte{0xFF}, "unknown(0xff)"},
		{"empty", []byte{}, "unknown(empty)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HeaderType(tt.raw); got != tt.want {
				t.Errorf("HeaderType(%v) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestDecodeSimpleIDLink(t *testing.T) {
	link, err := DecodeSimpleIDLink("[&AgEAAAA=]") // item link, header 0x02
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.LinkType != "item" {
		t.Errorf("LinkType = %q, want item", link.LinkType)
	}
}

func TestDecodeSimpleIDLink_TooShort(t *testing.T) {
	if _, err := DecodeSimpleIDLink("[&Ag==]"); err == nil {
		t.Error("expected error decoding a truncated simple id link, got nil")
	}
}

func TestStripAndWrapLink_RoundTrip(t *testing.T) {
	bare := StripLink(thiefSample)
	if WrapLink(bare) != thiefSample {
		t.Errorf("WrapLink(StripLink(code)) = %q, want %q", WrapLink(bare), thiefSample)
	}
}
