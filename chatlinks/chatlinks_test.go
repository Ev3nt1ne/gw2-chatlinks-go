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
// is exercised here only by the wiki's own worked numeric example (see
// TestSpecializationTraitByte_WikiWorkedExample), not by a real sample with
// a trait actually chosen.
//
// Their weapon arrays, however, ARE real and were cross-checked against
// independently-verified game facts, not just the wiki's abstract spec:
// Thief Dagger+Rifle (Rifle via the Deadeye elite spec), Elementalist Staff,
// Engineer Rifle+Hammer (Hammer via the Scrapper elite spec — confirmed via
// the wiki's own Scrapper article, since "Engineer + Hammer" isn't an
// obviously-correct fact to assume from memory), and Ranger Longbow+
// Greatsword (both core Ranger weapons).
const (
	thiefSample        = "[&DQUAAAAAAAAkDyQPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACLwBVAAA=]"
	elementalistSample = "[&DQYAAAAAAAAnDycPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABWQAA]"
	engineerSample     = "[&DQMAAAAAAAAqDyoPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACVQAzAAA=]"
	rangerPetSample    = "[&DQQAAAAAAAB5AAAAAAAAAAAAAAAAAAAAAAAAADA7FD8AAAAAAAAAAAAAAAACIwAyAAA=]"
)

// Additional real samples covering data the 4 level-2 samples above lack:
// actual trait picks, a 2-legend Revenant build, a fully distinct
// land/water skill loadout, a second Ranger pet build, and a real
// Weaponmaster Training skill-override sample. evokerSample1 and
// evokerSample2 are Evoker, an Elementalist elite spec distinct from
// Catalyst. See VERIFICATION.md for cross-check details.
const (
	weaverTraitSample         = "[&DQYpOxE/OBkAFnQA9RUAAPYAAABPAQAAEhcAAAAAAAAAAAAAAAAAAAAAAAABCQEA]"
	evokerSample1             = "[&DQYaLik7UDt0AHQAywDLAI8AjwAcARwBah1qHQAAAAAAAAAAAAAAAAAAAAACLwBZAAA=]"
	evokerSample2             = "[&DQYfCxEpUDpiHQAATQEAAB0BAACLHQAAah0AAAAAAAAAAAAAAAAAAAAAAAACLwBZAAA=]"
	catalystSample            = "[&DQYfPSkfQyZ0AHUAvgFNAVABkQD4Go8AJgCWAAAAAAAAAAAAAAAAAAAAAAACVgAvAAA=]"
	revenantLegendsRealSample = "[&DQkPPQMZPyrcEdwRBhIGEisSKxLUEdQRyhHKEQUCAwLUESsSBhIGEisS1BECawBaAAA=]"
	untamedPetsSample         = "[&DQQIOyAuSBuhABQblgGWAbYAmgC4ALgAlwHtAB9CFQwAAAAAAAAAAAAAAAACIwAFAAA=]"
	weaponmasterVariantSample = "[&DQQIKR46TilUHXkAmgAAAIMdmgB1HQAAWx0AAEI3ARMAAAAAAAAAAAAAAAACMwAJAQNn9wAAm/YAAOj2AAA=]"
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
	wantWeapons := []int{47, 85} // Dagger, Rifle
	if !equalIntSlices(bt.WeaponIDs, wantWeapons) {
		t.Errorf("WeaponIDs = %v, want %v", bt.WeaponIDs, wantWeapons)
	}
	if len(bt.SkillOverrideIDs) != 0 {
		t.Errorf("SkillOverrideIDs = %v, want empty", bt.SkillOverrideIDs)
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
	wantWeapons := []int{89} // Staff
	if !equalIntSlices(bt.WeaponIDs, wantWeapons) {
		t.Errorf("WeaponIDs = %v, want %v", bt.WeaponIDs, wantWeapons)
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
	// Rifle (core) + Hammer (Scrapper elite spec) — cross-checked against
	// the wiki's Scrapper article, not assumed from memory.
	wantWeapons := []int{85, 51}
	if !equalIntSlices(bt.WeaponIDs, wantWeapons) {
		t.Errorf("WeaponIDs = %v, want %v", bt.WeaponIDs, wantWeapons)
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
	want := RangerPets{TerrestrialPet1: 48, TerrestrialPet2: 59, AquaticPet1: 20, AquaticPet2: 63}
	if bt.RangerPets == nil || *bt.RangerPets != want {
		t.Fatalf("RangerPets = %+v, want %+v", bt.RangerPets, want)
	}
	// Longbow + Greatsword — both core Ranger weapons.
	wantWeapons := []int{35, 50}
	if !equalIntSlices(bt.WeaponIDs, wantWeapons) {
		t.Errorf("WeaponIDs = %v, want %v", bt.WeaponIDs, wantWeapons)
	}
}

// TestDecodeBuildTemplate_RealTraitSelections checks Adept/Master/
// Grandmaster picks against 4 real samples with actual trait choices.
func TestDecodeBuildTemplate_RealTraitSelections(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		profession string
		want       [3]SpecializationChoice
	}{
		{
			name:       "weaver",
			code:       weaverTraitSample,
			profession: "Elementalist",
			want: [3]SpecializationChoice{
				{SpecializationID: 41, Adept: 3, Master: 2, Grandmaster: 3},
				{SpecializationID: 17, Adept: 3, Master: 3, Grandmaster: 3},
				{SpecializationID: 56, Adept: 1, Master: 2, Grandmaster: 1},
			},
		},
		{
			name:       "evoker1",
			code:       evokerSample1,
			profession: "Elementalist",
			want: [3]SpecializationChoice{
				{SpecializationID: 26, Adept: 2, Master: 3, Grandmaster: 2},
				{SpecializationID: 41, Adept: 3, Master: 2, Grandmaster: 3},
				{SpecializationID: 80, Adept: 3, Master: 2, Grandmaster: 3},
			},
		},
		{
			name:       "evoker2",
			code:       evokerSample2,
			profession: "Elementalist",
			want: [3]SpecializationChoice{
				{SpecializationID: 31, Adept: 3, Master: 2, Grandmaster: 0},
				{SpecializationID: 17, Adept: 1, Master: 2, Grandmaster: 2},
				{SpecializationID: 80, Adept: 2, Master: 2, Grandmaster: 3},
			},
		},
		{
			name:       "catalyst",
			code:       catalystSample,
			profession: "Elementalist",
			want: [3]SpecializationChoice{
				{SpecializationID: 31, Adept: 1, Master: 3, Grandmaster: 3},
				{SpecializationID: 41, Adept: 3, Master: 3, Grandmaster: 1},
				{SpecializationID: 67, Adept: 2, Master: 1, Grandmaster: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt, err := DecodeBuildTemplate(tt.code)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if bt.Profession != tt.profession {
				t.Errorf("Profession = %q, want %q", bt.Profession, tt.profession)
			}
			if bt.Specializations != tt.want {
				t.Errorf("Specializations = %+v, want %+v", bt.Specializations, tt.want)
			}
		})
	}
}

// TestDecodeBuildTemplate_RealSampleNoAquaticSkills checks that no
// aquatic skill slot is populated when a real build genuinely sets none.
func TestDecodeBuildTemplate_RealSampleNoAquaticSkills(t *testing.T) {
	bt, err := DecodeBuildTemplate(evokerSample2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, paletteID := range bt.SkillPaletteIDs {
		if i%2 == 1 && paletteID != 0 {
			t.Errorf("SkillPaletteIDs[%d] (an aquatic slot) = %d, want 0", i, paletteID)
		}
	}
}

// TestDecodeBuildTemplate_RealRevenantLegends checks a real Revenant
// build with the active and inactive legend genuinely different from
// each other on both land and water.
func TestDecodeBuildTemplate_RealRevenantLegends(t *testing.T) {
	bt, err := DecodeBuildTemplate(revenantLegendsRealSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.Profession != "Revenant" {
		t.Errorf("Profession = %q, want Revenant", bt.Profession)
	}
	want := RevenantLegends{
		TerrestrialActive:   5, // Legendary Renegade Stance (Kalla)
		TerrestrialInactive: 2, // Legendary Assassin Stance (Shiro)
		AquaticActive:       3, // Legendary Dwarf Stance
		AquaticInactive:     2, // Legendary Assassin Stance (Shiro)
	}
	got := *bt.RevenantLegends
	if got.TerrestrialActive != want.TerrestrialActive ||
		got.TerrestrialInactive != want.TerrestrialInactive ||
		got.AquaticActive != want.AquaticActive ||
		got.AquaticInactive != want.AquaticInactive {
		t.Fatalf("RevenantLegends = %+v, want %+v", got, want)
	}
	wantSpecs := [3]SpecializationChoice{
		{SpecializationID: 15, Adept: 1, Master: 3, Grandmaster: 3},
		{SpecializationID: 3, Adept: 1, Master: 2, Grandmaster: 1},
		{SpecializationID: 63, Adept: 2, Master: 2, Grandmaster: 2},
	}
	if bt.Specializations != wantSpecs {
		t.Errorf("Specializations = %+v, want %+v", bt.Specializations, wantSpecs)
	}
}

// TestDecodeBuildTemplate_RealUntamedPets is a second real Ranger
// sample, checking all 4 pet slots decode to distinct values.
func TestDecodeBuildTemplate_RealUntamedPets(t *testing.T) {
	bt, err := DecodeBuildTemplate(untamedPetsSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.Profession != "Ranger" {
		t.Errorf("Profession = %q, want Ranger", bt.Profession)
	}
	want := RangerPets{TerrestrialPet1: 31, TerrestrialPet2: 66, AquaticPet1: 21, AquaticPet2: 12}
	if bt.RangerPets == nil || *bt.RangerPets != want {
		t.Fatalf("RangerPets = %+v, want %+v", bt.RangerPets, want)
	}
}

// TestDecodeBuildTemplate_RealWeaponmasterVariant is the real-sample
// counterpart to TestDecodeBuildTemplate_SkillOverrides's synthetic one:
// a Ranger build wielding Hammer (not a native Ranger weapon) with 3
// weapon-skill variants selected via Weaponmaster Training.
func TestDecodeBuildTemplate_RealWeaponmasterVariant(t *testing.T) {
	bt, err := DecodeBuildTemplate(weaponmasterVariantSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.Profession != "Ranger" {
		t.Errorf("Profession = %q, want Ranger", bt.Profession)
	}
	wantWeapons := []int{51, 265} // Hammer, Spear
	if !equalIntSlices(bt.WeaponIDs, wantWeapons) {
		t.Errorf("WeaponIDs = %v, want %v", bt.WeaponIDs, wantWeapons)
	}
	// Unleashed Wild Swing/Savage Shock Wave/Thump — Hammer skill variants
	// tagged to specialization 72 (Untamed) by the public API, despite
	// this build's own elite spec being a different one (Galeshot).
	wantOverrides := []int{63335, 63131, 63208}
	if !equalIntSlices(bt.SkillOverrideIDs, wantOverrides) {
		t.Errorf("SkillOverrideIDs = %v, want %v", bt.SkillOverrideIDs, wantOverrides)
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

func TestDecodeBuildTemplate_NoTrailingArrays(t *testing.T) {
	// A pre-SOTO-format build template: exactly minBuildTemplateLen bytes,
	// no weapon/skill-override array at all. Must decode without error and
	// leave both arrays nil, not just empty.
	raw := make([]byte, minBuildTemplateLen)
	raw[0] = 0x0D
	raw[1] = 1 // Guardian
	code := encodeRaw(raw)

	bt, err := DecodeBuildTemplate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.WeaponIDs != nil {
		t.Errorf("WeaponIDs = %v, want nil for a code with no trailing arrays", bt.WeaponIDs)
	}
	if bt.SkillOverrideIDs != nil {
		t.Errorf("SkillOverrideIDs = %v, want nil for a code with no trailing arrays", bt.SkillOverrideIDs)
	}
}

func TestDecodeBuildTemplate_WeaponArrayTruncated(t *testing.T) {
	raw := make([]byte, minBuildTemplateLen+1)
	raw[0] = 0x0D
	raw[1] = 1
	raw[minBuildTemplateLen] = 3 // claims 3 weapons, but no weapon bytes follow
	code := encodeRaw(raw)

	if _, err := DecodeBuildTemplate(code); err == nil {
		t.Error("expected error for a build template declaring more weapons than it has bytes for, got nil")
	}
}

func TestDecodeBuildTemplate_SkillOverrideArrayTruncated(t *testing.T) {
	raw := make([]byte, minBuildTemplateLen+2)
	raw[0] = 0x0D
	raw[1] = 1
	raw[minBuildTemplateLen] = 0   // zero weapons
	raw[minBuildTemplateLen+1] = 2 // claims 2 skill overrides, but no bytes follow
	code := encodeRaw(raw)

	if _, err := DecodeBuildTemplate(code); err == nil {
		t.Error("expected error for a build template declaring more skill overrides than it has bytes for, got nil")
	}
}

// TestSpecializationTraitByte_WikiWorkedExample exercises the wiki's own
// worked numeric example for the trait-tier byte: specialization ID 3
// (0b00000011), trait byte 0b00111001, decoding to Adept=1, Master=2,
// Grandmaster=3. This resolves the ambiguity in the wiki's prose ("2-bit
// values from 0 to 3, in reverse order") with a concrete worked value, in
// lieu of a real sample with a non-zero trait actually chosen.
func TestSpecializationTraitByte_WikiWorkedExample(t *testing.T) {
	raw := make([]byte, minBuildTemplateLen)
	raw[0] = 0x0D
	raw[1] = 1 // Guardian
	raw[2] = 0b00000011
	raw[3] = 0b00111001
	code := encodeRaw(raw)

	bt, err := DecodeBuildTemplate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := bt.Specializations[0]
	want := SpecializationChoice{SpecializationID: 3, Adept: 1, Master: 2, Grandmaster: 3}
	if got != want {
		t.Errorf("Specializations[0] = %+v, want %+v", got, want)
	}
}

func TestDecodeBuildTemplate_RevenantLegends(t *testing.T) {
	raw := make([]byte, minBuildTemplateLen)
	raw[0] = 0x0D
	raw[1] = 9 // Revenant
	const profBytesOffset = 2 + 6 + 20
	raw[profBytesOffset+0] = 1 // active terrestrial: Legendary Dragon Stance
	raw[profBytesOffset+1] = 2 // inactive terrestrial: Legendary Assassin Stance
	raw[profBytesOffset+2] = 3 // active aquatic: Legendary Dwarf Stance
	raw[profBytesOffset+3] = 4 // inactive aquatic: Legendary Demon Stance
	putU16le(raw, profBytesOffset+4, 100)
	putU16le(raw, profBytesOffset+6, 101)
	putU16le(raw, profBytesOffset+8, 102)
	putU16le(raw, profBytesOffset+10, 200)
	putU16le(raw, profBytesOffset+12, 201)
	putU16le(raw, profBytesOffset+14, 202)
	code := encodeRaw(raw)

	bt, err := DecodeBuildTemplate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := RevenantLegends{
		TerrestrialActive:                    1,
		TerrestrialInactive:                  2,
		AquaticActive:                        3,
		AquaticInactive:                      4,
		InactiveTerrestrialUtilityPaletteIDs: [3]int{100, 101, 102},
		InactiveAquaticUtilityPaletteIDs:     [3]int{200, 201, 202},
	}
	if bt.RevenantLegends == nil || *bt.RevenantLegends != want {
		t.Fatalf("RevenantLegends = %+v, want %+v", bt.RevenantLegends, want)
	}
}

func TestDecodeBuildTemplate_SkillOverrides(t *testing.T) {
	raw := make([]byte, minBuildTemplateLen+2)
	raw[0] = 0x0D
	raw[1] = 2                   // Warrior
	raw[minBuildTemplateLen] = 0 // zero weapons
	overrideCountOffset := minBuildTemplateLen + 1
	raw = append(raw, 0, 0, 0, 0, 0, 0, 0, 0) // room for 2 four-byte skill ids
	raw[overrideCountOffset] = 2
	putU32le(raw, overrideCountOffset+1, 12345)
	putU32le(raw, overrideCountOffset+5, 67890)
	code := encodeRaw(raw)

	bt, err := DecodeBuildTemplate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{12345, 67890}
	if !equalIntSlices(bt.SkillOverrideIDs, want) {
		t.Errorf("SkillOverrideIDs = %v, want %v", bt.SkillOverrideIDs, want)
	}
}

// TestEncodeBuildTemplate_RoundTripsRealSamples decodes each real sample
// and re-encodes it, asserting the resulting raw bytes are byte-for-byte
// identical to the original. This is a strong joint check on both
// DecodeBuildTemplate and EncodeBuildTemplate: any mismatch in offsets,
// bit-packing, or array handling between the two would surface here.
func TestEncodeBuildTemplate_RoundTripsRealSamples(t *testing.T) {
	samples := []string{
		thiefSample, elementalistSample, engineerSample, rangerPetSample,
		weaverTraitSample, evokerSample1, evokerSample2, catalystSample,
		revenantLegendsRealSample, untamedPetsSample, weaponmasterVariantSample,
	}
	for _, sample := range samples {
		wantRaw, err := DecodeRaw(sample)
		if err != nil {
			t.Fatalf("DecodeRaw(%q): %v", sample, err)
		}
		bt, err := DecodeBuildTemplate(sample)
		if err != nil {
			t.Fatalf("DecodeBuildTemplate(%q): %v", sample, err)
		}
		reencoded, err := EncodeBuildTemplate(bt)
		if err != nil {
			t.Fatalf("EncodeBuildTemplate: %v", err)
		}
		gotRaw, err := DecodeRaw(reencoded)
		if err != nil {
			t.Fatalf("DecodeRaw(reencoded): %v", err)
		}
		if string(gotRaw) != string(wantRaw) {
			t.Errorf("round-trip mismatch for %q:\n got  %x\n want %x", sample, gotRaw, wantRaw)
		}
	}
}

func TestEncodeBuildTemplate_RevenantLegendsRoundTrip(t *testing.T) {
	bt := BuildTemplate{
		ProfessionID: 9,
		Profession:   "Revenant",
		RevenantLegends: &RevenantLegends{
			TerrestrialActive:                    5,
			TerrestrialInactive:                  6,
			AquaticActive:                        7,
			AquaticInactive:                      8,
			InactiveTerrestrialUtilityPaletteIDs: [3]int{10, 20, 30},
			InactiveAquaticUtilityPaletteIDs:     [3]int{40, 50, 60},
		},
	}
	code, err := EncodeBuildTemplate(bt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := DecodeBuildTemplate(code)
	if err != nil {
		t.Fatalf("unexpected error decoding re-encoded link: %v", err)
	}
	if got.RevenantLegends == nil || *got.RevenantLegends != *bt.RevenantLegends {
		t.Errorf("RevenantLegends = %+v, want %+v", got.RevenantLegends, bt.RevenantLegends)
	}
}

func TestEncodeBuildTemplate_WeaponsAndSkillOverrides(t *testing.T) {
	bt := BuildTemplate{
		ProfessionID:     2, // Warrior
		Profession:       "Warrior",
		WeaponIDs:        []int{90, 51}, // Sword, Hammer
		SkillOverrideIDs: []int{111, 222, 333},
	}
	code, err := EncodeBuildTemplate(bt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := DecodeBuildTemplate(code)
	if err != nil {
		t.Fatalf("unexpected error decoding re-encoded link: %v", err)
	}
	if !equalIntSlices(got.WeaponIDs, bt.WeaponIDs) {
		t.Errorf("WeaponIDs = %v, want %v", got.WeaponIDs, bt.WeaponIDs)
	}
	if !equalIntSlices(got.SkillOverrideIDs, bt.SkillOverrideIDs) {
		t.Errorf("SkillOverrideIDs = %v, want %v", got.SkillOverrideIDs, bt.SkillOverrideIDs)
	}
}

func TestEncodeBuildTemplate_TooManyWeapons(t *testing.T) {
	bt := BuildTemplate{ProfessionID: 1, WeaponIDs: make([]int, maxWeapons+1)}
	if _, err := EncodeBuildTemplate(bt); err == nil {
		t.Error("expected error for more than maxWeapons weapons, got nil")
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
	if link.Quantity != 1 {
		t.Errorf("Quantity = %d, want 1", link.Quantity)
	}
}

func TestDecodeSimpleIDLink_TooShort(t *testing.T) {
	if _, err := DecodeSimpleIDLink("[&Ag==]"); err == nil {
		t.Error("expected error decoding a truncated simple id link, got nil")
	}
}

func TestEncodeSimpleIDLink_RoundTrip(t *testing.T) {
	tests := []SimpleIDLink{
		{LinkType: "skill", ID: 3876},
		{LinkType: "trait", ID: 1234},
		{LinkType: "item", ID: 5678, Quantity: 3},
		{LinkType: "recipe", ID: 42},
	}
	for _, want := range tests {
		code, err := EncodeSimpleIDLink(want)
		if err != nil {
			t.Fatalf("EncodeSimpleIDLink(%+v): %v", want, err)
		}
		got, err := DecodeSimpleIDLink(code)
		if err != nil {
			t.Fatalf("DecodeSimpleIDLink(%q): %v", code, err)
		}
		if got != want {
			t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
		}
	}
}

func TestEncodeSimpleIDLink_ItemDefaultQuantity(t *testing.T) {
	code, err := EncodeSimpleIDLink(SimpleIDLink{LinkType: "item", ID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	link, err := DecodeSimpleIDLink(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.Quantity != 1 {
		t.Errorf("Quantity = %d, want 1 (default)", link.Quantity)
	}
}

func TestEncodeSimpleIDLink_UnknownLinkType(t *testing.T) {
	if _, err := EncodeSimpleIDLink(SimpleIDLink{LinkType: "not-a-real-type", ID: 1}); err == nil {
		t.Error("expected error for unknown link type, got nil")
	}
}

func TestStripAndWrapLink_RoundTrip(t *testing.T) {
	bare := StripLink(thiefSample)
	if WrapLink(bare) != thiefSample {
		t.Errorf("WrapLink(StripLink(code)) = %q, want %q", WrapLink(bare), thiefSample)
	}
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
