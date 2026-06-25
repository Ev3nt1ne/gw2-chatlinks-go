package chatlinks

import "fmt"

// minBuildTemplateLen is the byte length required to read every fixed field
// of a build template link: header(1) + profession(1) + specializations(6)
// + skill palette ids(20) + profession-specific bytes(16). Build templates
// created before the 27 June 2023 Secrets of the Obscure update don't carry
// the trailing weapon/skill-override arrays at all, so this is also the
// minimum valid length for a build template link of any vintage.
const minBuildTemplateLen = 1 + 1 + 6 + 20 + 16

// maxWeapons is the documented cap on a build template's weapon array.
const maxWeapons = 8

// SpecializationChoice is one of a build template's 3 specialization slots.
// Each trait tier (Adept/Master/Grandmaster) is 0 (none chosen) or 1-3
// (which of the 3 trait options in that tier was picked).
type SpecializationChoice struct {
	SpecializationID int
	Adept            int
	Master           int
	Grandmaster      int
}

// RangerPets holds a Ranger build template's 4 pet slots. 0 means no pet
// set for that slot. Field order matches the wiki's documented byte order.
type RangerPets struct {
	TerrestrialPet1 int
	TerrestrialPet2 int
	AquaticPet1     int
	AquaticPet2     int
}

// RevenantLegends holds a Revenant build template's legend selection and
// the display order of its inactive legend's utility skills. Legend values
// are codes 1-8 (see the Legends map), 0 meaning unset.
type RevenantLegends struct {
	// TerrestrialActive is the active legend, terrestrial (above weapon
	// skill 2). TerrestrialInactive is the inactive legend, terrestrial
	// (above weapon skill 1). Aquatic* are the same pairing underwater.
	TerrestrialActive   int
	TerrestrialInactive int
	AquaticActive       int
	AquaticInactive     int

	// InactiveTerrestrialUtilityPaletteIDs / InactiveAquaticUtilityPaletteIDs
	// record the display order of the *inactive* legend's 3 utility skills
	// (palette IDs, resolved the same way as BuildTemplate.SkillPaletteIDs
	// — see the api package). Revenant utility skills are fixed per legend,
	// not individually chosen, but their on-screen order can be customized.
	InactiveTerrestrialUtilityPaletteIDs [3]int
	InactiveAquaticUtilityPaletteIDs     [3]int
}

// BuildTemplate is a decoded build template (header 0x0D) link.
type BuildTemplate struct {
	ProfessionID    int
	Profession      string
	Specializations [3]SpecializationChoice

	// SkillPaletteIDs holds "palette IDs" (NOT public API skill IDs), in
	// order: heal, utility x3, elite — terrestrial then aquatic for each,
	// i.e. [heal_t, heal_a, util1_t, util1_a, util2_t, util2_a, util3_t,
	// util3_a, elite_t, elite_a]. Resolve via the public API's
	// /v2/professions skills_by_palette field (see the api package).
	SkillPaletteIDs [10]int

	// ProfessionBytes are the raw 16 profession-specific bytes, preserved
	// verbatim for professions with no dedicated decoded view below (and as
	// the byte-for-byte source for round-tripping unknown professions).
	// When RangerPets or RevenantLegends is set, EncodeBuildTemplate uses
	// that structured field instead of ProfessionBytes for the
	// corresponding profession.
	ProfessionBytes [16]byte

	// RangerPets is only populated when Profession == "Ranger".
	RangerPets *RangerPets

	// RevenantLegends is only populated when Profession == "Revenant".
	RevenantLegends *RevenantLegends

	// WeaponIDs are the build's terrestrial weapons (see WeaponTypes), 0-8
	// entries. Added by the 27 June 2023 Secrets of the Obscure update;
	// nil for build templates created before that, which don't carry this
	// section at all.
	WeaponIDs []int

	// SkillOverrideIDs are public-API skill IDs for weapon-skill variants
	// selected via Weaponmaster Training, in slot order. Added alongside
	// Weaponmaster Training; nil if no overrides are set or the build
	// template predates the weapon array (and therefore this section too).
	SkillOverrideIDs []int

	RawHex string
}

// DecodeBuildTemplate decodes a build template (header 0x0D) chat link.
func DecodeBuildTemplate(code string) (BuildTemplate, error) {
	raw, err := DecodeRaw(code)
	if err != nil {
		return BuildTemplate{}, err
	}
	if raw[0] != 0x0D {
		return BuildTemplate{}, fmt.Errorf("chatlinks: not a build template link (header 0x%02x)", raw[0])
	}
	if len(raw) < minBuildTemplateLen {
		return BuildTemplate{}, fmt.Errorf("chatlinks: build template payload too short: got %d bytes, need at least %d", len(raw), minBuildTemplateLen)
	}

	professionID := int(raw[1])
	profession, ok := professions[professionID]
	if !ok {
		profession = fmt.Sprintf("unknown(%d)", professionID)
	}

	var specs [3]SpecializationChoice
	const specOffset = 2
	for i := 0; i < 3; i++ {
		specID := int(raw[specOffset+i*2])
		traitByte := raw[specOffset+i*2+1]
		// Bit layout confirmed against the wiki's own worked numeric
		// example (byte 0b00111001 -> Adept=1, Master=2, Grandmaster=3):
		// bits 0-1 = Adept, bits 2-3 = Master, bits 4-5 = Grandmaster, top
		// 2 bits unused. The wiki's "in reverse order" phrasing refers to
		// the human (MSB-first) reading order encountering
		// Grandmaster/Master/Adept, not a reversed value mapping.
		specs[i] = SpecializationChoice{
			SpecializationID: specID,
			Adept:            int(traitByte & 0b11),
			Master:           int((traitByte >> 2) & 0b11),
			Grandmaster:      int((traitByte >> 4) & 0b11),
		}
	}

	const skillsOffset = specOffset + 6 // 3 specs x 2 bytes
	var skillPaletteIDs [10]int
	for i := 0; i < 10; i++ {
		skillPaletteIDs[i] = u16le(raw, skillsOffset+i*2)
	}

	const profBytesOffset = skillsOffset + 20 // 10 skills x 2 bytes
	var professionBytes [16]byte
	copy(professionBytes[:], raw[profBytesOffset:profBytesOffset+16])

	var rangerPets *RangerPets
	var revenantLegends *RevenantLegends
	switch profession {
	case "Ranger":
		rangerPets = &RangerPets{
			TerrestrialPet1: int(professionBytes[0]),
			TerrestrialPet2: int(professionBytes[1]),
			AquaticPet1:     int(professionBytes[2]),
			AquaticPet2:     int(professionBytes[3]),
		}
	case "Revenant":
		rl := RevenantLegends{
			TerrestrialActive:   int(professionBytes[0]),
			TerrestrialInactive: int(professionBytes[1]),
			AquaticActive:       int(professionBytes[2]),
			AquaticInactive:     int(professionBytes[3]),
		}
		for i := 0; i < 3; i++ {
			rl.InactiveTerrestrialUtilityPaletteIDs[i] = u16le(professionBytes[:], 4+i*2)
			rl.InactiveAquaticUtilityPaletteIDs[i] = u16le(professionBytes[:], 10+i*2)
		}
		revenantLegends = &rl
	}

	arrayOffset := profBytesOffset + 16
	var weaponIDs []int
	var skillOverrideIDs []int
	if len(raw) > arrayOffset {
		weaponCount := int(raw[arrayOffset])
		weaponsEnd := arrayOffset + 1 + weaponCount*2
		if len(raw) < weaponsEnd {
			return BuildTemplate{}, fmt.Errorf("chatlinks: weapon array truncated: declared %d weapons but payload too short", weaponCount)
		}
		for i := 0; i < weaponCount; i++ {
			weaponIDs = append(weaponIDs, u16le(raw, arrayOffset+1+i*2))
		}

		if len(raw) > weaponsEnd {
			overrideCount := int(raw[weaponsEnd])
			overridesEnd := weaponsEnd + 1 + overrideCount*4
			if len(raw) < overridesEnd {
				return BuildTemplate{}, fmt.Errorf("chatlinks: skill override array truncated: declared %d overrides but payload too short", overrideCount)
			}
			for i := 0; i < overrideCount; i++ {
				skillOverrideIDs = append(skillOverrideIDs, u32le(raw, weaponsEnd+1+i*4))
			}
		}
	}

	return BuildTemplate{
		ProfessionID:     professionID,
		Profession:       profession,
		Specializations:  specs,
		SkillPaletteIDs:  skillPaletteIDs,
		ProfessionBytes:  professionBytes,
		RangerPets:       rangerPets,
		RevenantLegends:  revenantLegends,
		WeaponIDs:        weaponIDs,
		SkillOverrideIDs: skillOverrideIDs,
		RawHex:           fmt.Sprintf("%x", raw),
	}, nil
}

// EncodeBuildTemplate encodes a build template (header 0x0D) chat link.
// RangerPets / RevenantLegends, when set, take precedence over
// ProfessionBytes for the corresponding profession; ProfessionBytes is
// otherwise copied through verbatim (so unknown professions and any region
// this package doesn't structurally model round-trip byte-for-byte).
func EncodeBuildTemplate(bt BuildTemplate) (string, error) {
	if len(bt.WeaponIDs) > maxWeapons {
		return "", fmt.Errorf("chatlinks: too many weapons: %d (max %d)", len(bt.WeaponIDs), maxWeapons)
	}
	if len(bt.SkillOverrideIDs) > 0xFF {
		return "", fmt.Errorf("chatlinks: too many skill overrides: %d (max 255)", len(bt.SkillOverrideIDs))
	}

	const specOffset = 2
	const skillsOffset = specOffset + 6
	const profBytesOffset = skillsOffset + 20
	arrayOffset := profBytesOffset + 16
	weaponsEnd := arrayOffset + 1 + len(bt.WeaponIDs)*2
	size := weaponsEnd + 1 + len(bt.SkillOverrideIDs)*4

	raw := make([]byte, size)
	raw[0] = 0x0D
	raw[1] = byte(bt.ProfessionID)

	for i, spec := range bt.Specializations {
		raw[specOffset+i*2] = byte(spec.SpecializationID)
		raw[specOffset+i*2+1] = byte(spec.Adept&0b11) | byte(spec.Master&0b11)<<2 | byte(spec.Grandmaster&0b11)<<4
	}

	for i, paletteID := range bt.SkillPaletteIDs {
		putU16le(raw, skillsOffset+i*2, paletteID)
	}

	profBytes := bt.ProfessionBytes
	// Derive from ProfessionID (the canonical byte-level field) rather
	// than the Profession display-name string, so encoding is correct
	// even if a caller builds a BuildTemplate by hand and only sets
	// ProfessionID.
	switch professions[bt.ProfessionID] {
	case "Ranger":
		if bt.RangerPets != nil {
			profBytes[0] = byte(bt.RangerPets.TerrestrialPet1)
			profBytes[1] = byte(bt.RangerPets.TerrestrialPet2)
			profBytes[2] = byte(bt.RangerPets.AquaticPet1)
			profBytes[3] = byte(bt.RangerPets.AquaticPet2)
		}
	case "Revenant":
		if rl := bt.RevenantLegends; rl != nil {
			profBytes[0] = byte(rl.TerrestrialActive)
			profBytes[1] = byte(rl.TerrestrialInactive)
			profBytes[2] = byte(rl.AquaticActive)
			profBytes[3] = byte(rl.AquaticInactive)
			for i, id := range rl.InactiveTerrestrialUtilityPaletteIDs {
				putU16le(profBytes[:], 4+i*2, id)
			}
			for i, id := range rl.InactiveAquaticUtilityPaletteIDs {
				putU16le(profBytes[:], 10+i*2, id)
			}
		}
	}
	copy(raw[profBytesOffset:profBytesOffset+16], profBytes[:])

	raw[arrayOffset] = byte(len(bt.WeaponIDs))
	for i, weaponID := range bt.WeaponIDs {
		putU16le(raw, arrayOffset+1+i*2, weaponID)
	}

	raw[weaponsEnd] = byte(len(bt.SkillOverrideIDs))
	for i, skillID := range bt.SkillOverrideIDs {
		putU32le(raw, weaponsEnd+1+i*4, skillID)
	}

	return encodeRaw(raw), nil
}
