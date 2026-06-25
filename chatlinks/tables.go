package chatlinks

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
