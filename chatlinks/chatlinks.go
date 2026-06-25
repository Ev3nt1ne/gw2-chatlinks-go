// Package chatlinks decodes and encodes Guild Wars 2 chat links ([&...]
// codes).
//
// Format reference: https://wiki.guildwars2.com/wiki/Chat_link_format
//
// This package started as a Go port of the gw2-chatlinks-py prototype
// written for the Heroes Ascent project. The header byte (0x0D, build
// template), profession byte values, and weapon-array entries were verified
// empirically against real build-template codes pulled from a live ruleset
// document — cross-checked not just against the wiki's documented values,
// but against independently-confirmed game facts (e.g. a decoded Engineer
// weapon array of Rifle+Hammer was cross-checked against the Scrapper elite
// specialization's weapon grant). The trait-tier bit layout was confirmed
// against the wiki's own worked numeric example. Revenant legend bytes and
// the skill-override array are implemented per the wiki spec but, as of
// this writing, not yet exercised by any real sample (see
// chatlinks_test.go for what's covered).
package chatlinks

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// HeaderTypes maps a chat link's first byte to its link type name.
var HeaderTypes = map[byte]string{
	0x01: "coin",
	0x02: "item",
	0x03: "text",
	0x04: "map",
	0x05: "pvp_game",
	0x06: "skill",
	0x07: "trait",
	0x08: "user",
	0x09: "recipe",
	0x0A: "wardrobe",
	0x0B: "outfit",
	0x0C: "wvw_objective",
	0x0D: "build_template",
	0x0E: "achievement",
	0x0F: "wardrobe_template",
	0x10: "travel_template",
}

var linkTypeToHeader = func() map[string]byte {
	m := make(map[string]byte, len(HeaderTypes))
	for b, t := range HeaderTypes {
		m[t] = b
	}
	return m
}()

// Professions maps a build template's profession byte to its name.
var Professions = map[int]string{
	1: "Guardian",
	2: "Warrior",
	3: "Engineer",
	4: "Ranger",
	5: "Thief",
	6: "Elementalist",
	7: "Mesmer",
	8: "Necromancer",
	9: "Revenant",
}

// WeaponTypes maps a build template's weapon-array entry to its weapon
// name, per the wiki's documented "known weapon type IDs" table.
var WeaponTypes = map[int]string{
	5:   "Axe",
	35:  "Longbow",
	47:  "Dagger",
	49:  "Focus",
	50:  "Greatsword",
	51:  "Hammer",
	53:  "Mace",
	54:  "Pistol",
	85:  "Rifle",
	86:  "Scepter",
	87:  "Shield",
	89:  "Staff",
	90:  "Sword",
	102: "Torch",
	103: "Warhorn",
	107: "Shortbow",
	265: "Spear",
}

// Legends maps a Revenant build template's legend byte to its stance name,
// per the wiki's table (itself sourced from the public API's /v2/legends
// `code` field).
var Legends = map[int]string{
	1: "Legendary Dragon Stance",
	2: "Legendary Assassin Stance",
	3: "Legendary Dwarf Stance",
	4: "Legendary Demon Stance",
	5: "Legendary Renegade Stance",
	6: "Legendary Centaur Stance",
	7: "Legendary Alliance Stance",
	8: "Legendary Entity Stance",
}

// minBuildTemplateLen is the byte length required to read every fixed field
// of a build template link: header(1) + profession(1) + specializations(6)
// + skill palette ids(20) + profession-specific bytes(16). Build templates
// created before the 27 June 2023 Secrets of the Obscure update don't carry
// the trailing weapon/skill-override arrays at all, so this is also the
// minimum valid length for a build template link of any vintage.
const minBuildTemplateLen = 1 + 1 + 6 + 20 + 16

// maxWeapons is the documented cap on a build template's weapon array.
const maxWeapons = 8

// StripLink removes the "[&...]" wrapper, if present, returning the bare
// base64 payload.
func StripLink(code string) string {
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "[&") && strings.HasSuffix(code, "]") {
		return code[2 : len(code)-1]
	}
	return code
}

// WrapLink wraps a base64 payload in the "[&...]" chat link syntax.
func WrapLink(b64 string) string {
	return "[&" + b64 + "]"
}

// DecodeRaw strips the [&...] wrapper (if present) and base64-decodes the
// payload. GW2 chat links sometimes omit base64 padding, so it is added
// back as needed.
func DecodeRaw(code string) ([]byte, error) {
	b64 := StripLink(code)
	if pad := (4 - len(b64)%4) % 4; pad != 0 {
		b64 += strings.Repeat("=", pad)
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("chatlinks: invalid base64 payload: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("chatlinks: empty payload")
	}
	return raw, nil
}

// encodeRaw base64-encodes raw and wraps it in "[&...]" syntax.
func encodeRaw(raw []byte) string {
	return WrapLink(base64.StdEncoding.EncodeToString(raw))
}

// HeaderType returns the link type name for a decoded payload's first byte.
func HeaderType(raw []byte) string {
	if len(raw) == 0 {
		return "unknown(empty)"
	}
	if t, ok := HeaderTypes[raw[0]]; ok {
		return t
	}
	return fmt.Sprintf("unknown(0x%02x)", raw[0])
}

func u16le(raw []byte, offset int) int {
	return int(raw[offset]) | int(raw[offset+1])<<8
}

func u24le(raw []byte, offset int) int {
	return int(raw[offset]) | int(raw[offset+1])<<8 | int(raw[offset+2])<<16
}

func u32le(raw []byte, offset int) int {
	return int(raw[offset]) | int(raw[offset+1])<<8 | int(raw[offset+2])<<16 | int(raw[offset+3])<<24
}

func putU16le(buf []byte, offset, value int) {
	buf[offset] = byte(value)
	buf[offset+1] = byte(value >> 8)
}

func putU24le(buf []byte, offset, value int) {
	buf[offset] = byte(value)
	buf[offset+1] = byte(value >> 8)
	buf[offset+2] = byte(value >> 16)
}

func putU32le(buf []byte, offset, value int) {
	buf[offset] = byte(value)
	buf[offset+1] = byte(value >> 8)
	buf[offset+2] = byte(value >> 16)
	buf[offset+3] = byte(value >> 24)
}

// SimpleIDLink represents the common "single ID" link shapes: skill (0x06),
// trait (0x07), item (0x02), recipe (0x09), etc.
type SimpleIDLink struct {
	LinkType string
	ID       int

	// Quantity is only meaningful when LinkType == "item"; it's the stack
	// size encoded immediately before the item ID. Zero/unset for other
	// link types.
	Quantity int
}

// DecodeSimpleIDLink decodes a skill/trait/item/recipe-shaped link. Item
// links carry a quantity byte before the ID; everything else here doesn't.
func DecodeSimpleIDLink(code string) (SimpleIDLink, error) {
	raw, err := DecodeRaw(code)
	if err != nil {
		return SimpleIDLink{}, err
	}
	t := HeaderType(raw)
	offset := 1
	if t == "item" {
		offset = 2
	}
	if len(raw) < offset+3 {
		return SimpleIDLink{}, fmt.Errorf("chatlinks: payload too short for %s link: got %d bytes, need at least %d", t, len(raw), offset+3)
	}
	link := SimpleIDLink{LinkType: t, ID: u24le(raw, offset)}
	if t == "item" {
		link.Quantity = int(raw[1])
	}
	return link, nil
}

// EncodeSimpleIDLink encodes a skill/trait/item/recipe-shaped link.
func EncodeSimpleIDLink(link SimpleIDLink) (string, error) {
	header, ok := linkTypeToHeader[link.LinkType]
	if !ok {
		return "", fmt.Errorf("chatlinks: unknown link type %q", link.LinkType)
	}
	if link.ID < 0 || link.ID > 0xFFFFFF {
		return "", fmt.Errorf("chatlinks: id %d out of range for a 3-byte id", link.ID)
	}

	if link.LinkType == "item" {
		quantity := link.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		if quantity > 0xFF {
			return "", fmt.Errorf("chatlinks: quantity %d out of range for a 1-byte field", quantity)
		}
		raw := make([]byte, 5)
		raw[0] = header
		raw[1] = byte(quantity)
		putU24le(raw, 2, link.ID)
		return encodeRaw(raw), nil
	}

	raw := make([]byte, 4)
	raw[0] = header
	putU24le(raw, 1, link.ID)
	return encodeRaw(raw), nil
}

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
	profession, ok := Professions[professionID]
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
	switch Professions[bt.ProfessionID] {
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
