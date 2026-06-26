package chatlinks

import (
	"errors"
	"strings"
	"testing"
)

func TestDecodeRaw_ErrorKinds(t *testing.T) {
	tests := []struct {
		name string
		code string
		want error
	}{
		{"invalid base64", "[&not valid base64!!]", ErrInvalidPayload},
		{"empty payload", "[&]", ErrInvalidPayload},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeRaw(tt.code)
			if !errors.Is(err, tt.want) {
				t.Errorf("DecodeRaw(%q) error = %v, want errors.Is %v", tt.code, err, tt.want)
			}
		})
	}
}

func TestDecodeBuildTemplate_ErrorKinds(t *testing.T) {
	// A valid coin link (header 0x01) is the wrong header for a build template.
	coin := "[&AQAAAAA=]"
	if _, err := DecodeBuildTemplate(coin); !errors.Is(err, ErrWrongHeader) {
		t.Errorf("DecodeBuildTemplate(coin) error = %v, want errors.Is ErrWrongHeader", err)
	}

	// Header 0x0D but far too few bytes to be a build template.
	short := WrapLink("DQA=")
	if _, err := DecodeBuildTemplate(short); !errors.Is(err, ErrTruncated) {
		t.Errorf("DecodeBuildTemplate(short) error = %v, want errors.Is ErrTruncated", err)
	}
}

func TestDecodeSimpleIDLink_RejectsNonIDHeaders(t *testing.T) {
	// coin (0x01), text (0x03), and build_template (0x0D) are all known
	// headers but not "header + 3-byte id" shaped, so DecodeSimpleIDLink must
	// refuse them rather than return a meaningless ID.
	for _, code := range []string{
		"[&AQAAAAA=]", // coin
		"[&AwAAAAA=]", // text
		"[&DQUAAAAAAAAkDyQPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACLwBVAAA=]", // build template
	} {
		if _, err := DecodeSimpleIDLink(code); !errors.Is(err, ErrWrongHeader) {
			t.Errorf("DecodeSimpleIDLink(%q) error = %v, want errors.Is ErrWrongHeader", code, err)
		}
	}
}

func TestEncodeSimpleIDLink_ErrorKinds(t *testing.T) {
	if _, err := EncodeSimpleIDLink(SimpleIDLink{LinkType: "coin", ID: 1}); !errors.Is(err, ErrUnknownLinkType) {
		t.Errorf("encoding a coin link: error = %v, want errors.Is ErrUnknownLinkType", err)
	}
	if _, err := EncodeSimpleIDLink(SimpleIDLink{LinkType: "skill", ID: 1 << 25}); !errors.Is(err, ErrValueOutOfRange) {
		t.Errorf("encoding an oversized id: error = %v, want errors.Is ErrValueOutOfRange", err)
	}
}

func TestEncodeBuildTemplate_RangeChecks(t *testing.T) {
	tests := []struct {
		name string
		bt   BuildTemplate
	}{
		{"profession id > 255", BuildTemplate{ProfessionID: 256}},
		{"specialization id > 255", BuildTemplate{Specializations: [3]SpecializationChoice{{SpecializationID: 300}}}},
		{"palette id > 65535", BuildTemplate{SkillPaletteIDs: [10]int{0x1_0000}}},
		{"weapon id > 65535", BuildTemplate{WeaponIDs: []int{0x1_0000}}},
		{"skill override id > uint32", BuildTemplate{SkillOverrideIDs: []int{0x1_0000_0000}}},
		{"too many weapons", BuildTemplate{WeaponIDs: make([]int, maxWeapons+1)}},
		{"ranger pet > 255", BuildTemplate{ProfessionID: 4, RangerPets: &RangerPets{TerrestrialPet1: 256}}},
		{"revenant legend > 255", BuildTemplate{ProfessionID: 9, RevenantLegends: &RevenantLegends{TerrestrialActive: 256}}},
		{"revenant utility palette > 65535", BuildTemplate{ProfessionID: 9, RevenantLegends: &RevenantLegends{InactiveTerrestrialUtilityPaletteIDs: [3]int{0x1_0000}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncodeBuildTemplate(tt.bt)
			if !errors.Is(err, ErrValueOutOfRange) {
				t.Errorf("EncodeBuildTemplate(%s) error = %v, want errors.Is ErrValueOutOfRange", tt.name, err)
			}
		})
	}
}

func TestEncodeBuildTemplate_IgnoresStructForWrongProfession(t *testing.T) {
	// An out-of-range RangerPets on a non-Ranger build must not be rejected,
	// because EncodeBuildTemplate never reads it for that profession.
	bt := BuildTemplate{ProfessionID: 2 /* Warrior */, RangerPets: &RangerPets{TerrestrialPet1: 9999}}
	if _, err := EncodeBuildTemplate(bt); err != nil {
		t.Errorf("EncodeBuildTemplate ignored-struct case: unexpected error %v", err)
	}
}

func TestErrorMessagesStillDescriptive(t *testing.T) {
	// The sentinel is the stable contract, but the wrapped detail must remain
	// for humans. Spot-check one.
	_, err := DecodeBuildTemplate("[&AQAAAAA=]")
	if err == nil || !strings.Contains(err.Error(), "header 0x01") {
		t.Errorf("error lost its descriptive detail: %v", err)
	}
}
