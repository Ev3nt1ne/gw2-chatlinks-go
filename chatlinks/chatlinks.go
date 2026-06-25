// Package chatlinks decodes (and partially encodes) Guild Wars 2 chat links
// ([&...] codes).
//
// Format reference: https://wiki.guildwars2.com/wiki/Chat_link_format
//
// This package is a Go port of the gw2-chatlinks-py prototype written for
// the Heroes Ascent project. The header byte (0x0D, build template) and the
// profession byte values were verified empirically against real
// build-template codes pulled from a live ruleset document before trusting
// the wiki's documented values for the rest. Anything not covered by a real
// sample (see chatlinks_test.go) should be treated as "per the wiki,
// unverified" — cross-check against https://en.gw2skills.net/editor/ before
// relying on it.
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

// minBuildTemplateLen is the byte length required to read every fixed field
// of a build template link: header(1) + profession(1) + specializations(6)
// + skill palette ids(20) + profession-specific bytes(16).
const minBuildTemplateLen = 1 + 1 + 6 + 20 + 16

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

func u24le(raw []byte, offset int) int {
	return int(raw[offset]) | int(raw[offset+1])<<8 | int(raw[offset+2])<<16
}

func u16le(raw []byte, offset int) int {
	return int(raw[offset]) | int(raw[offset+1])<<8
}

// SimpleIDLink represents the common "single ID" link shapes: skill (0x06),
// trait (0x07), item (0x02, ignoring its quantity byte), recipe (0x09), etc.
type SimpleIDLink struct {
	LinkType string
	ID       int
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
	return SimpleIDLink{LinkType: t, ID: u24le(raw, offset)}, nil
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

// BuildTemplate is a decoded build template (header 0x0D) link.
type BuildTemplate struct {
	ProfessionID int
	Profession   string
	Specializations [3]SpecializationChoice

	// SkillPaletteIDs holds "palette IDs" (NOT public API skill IDs), in
	// order: heal, utility x3, elite — terrestrial then aquatic for each,
	// i.e. [heal_t, heal_a, util1_t, util1_a, util2_t, util2_a, util3_t,
	// util3_a, elite_t, elite_a]. Resolve via the public API's
	// /v2/professions skills_by_palette field (see the api package).
	SkillPaletteIDs [10]int

	// ProfessionBytes are the raw 16 profession-specific bytes. Ranger pet
	// IDs and Revenant legend IDs live in this region with different
	// layouts; only Ranger is decoded here (see RangerPetIDs).
	ProfessionBytes [16]byte

	// RangerPetIDs is only populated when Profession == "Ranger".
	RangerPetIDs []int

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
		// "2-bit values from 0 to 3, in reverse order" per the wiki - i.e.
		// the byte holds 3 tiers x 2 bits each. Bit layout confirmed
		// against gw2skills.net for the build-template samples in
		// chatlinks_test.go, but none of those samples have a non-zero
		// trait tier — this path is structurally implemented but NOT
		// verified against real data with a trait actually chosen.
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

	var rangerPetIDs []int
	if profession == "Ranger" {
		for i := 0; i < 4; i++ {
			if professionBytes[i] != 0 {
				rangerPetIDs = append(rangerPetIDs, int(professionBytes[i]))
			}
		}
	}

	return BuildTemplate{
		ProfessionID:     professionID,
		Profession:       profession,
		Specializations:  specs,
		SkillPaletteIDs:  skillPaletteIDs,
		ProfessionBytes:  professionBytes,
		RangerPetIDs:     rangerPetIDs,
		RawHex:           fmt.Sprintf("%x", raw),
	}, nil
}
