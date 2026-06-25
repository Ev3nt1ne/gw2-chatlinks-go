# Verification

What's been checked against real game data, what's implemented from the
[wiki spec](https://wiki.guildwars2.com/wiki/Chat_link_format) alone and
not yet sample-verified, and where the test fixtures come from.

## Build templates (`0x0D`)

Originated as a Go port of an earlier Python prototype
(`gw2-chatlinks-py`, not published — kept locally as a reference
implementation), since extended with encoding, weapon/skill-override
arrays, and Revenant legends. Verified against 4 real build-template
codes (see `chatlinks/chatlinks_test.go`):

- **Header byte and profession byte mapping** (`0x0D` = build template,
  1-9 → Guardian..Revenant) — verified empirically against real codes.
- **Skill palette slot order** (heal/utility×3/elite, terrestrial +
  aquatic) and palette→real-skill-ID resolution — end-to-end tested:
  Thief→Skelk Venom, Elementalist→Arcane Brilliance, Engineer→A.E.D., all
  genuine non-passive heal skills.
- **Ranger pet decoding** preserves slot position (`RangerPets`, a
  4-field struct) rather than compacting non-zero slots into a flat list,
  which would shift later pets into earlier slots on round-trip if any
  slot were actually unset.
- **Weapon array decoding** — cross-checked against independently
  verified game facts, not just the spec: the Engineer sample decodes to
  Rifle+Hammer, consistent with the wiki's own Scrapper article (Scrapper
  grants Hammer access). Thief (Dagger+Rifle, Rifle via Deadeye), Ranger
  (Longbow+Greatsword, both core weapons), and Elementalist (Staff) check
  out the same way.
- **Specialization/trait-tier bit-packing** (`Adept`/`Master`/
  `Grandmaster`) — all 4 real samples have zero specializations chosen
  (expected: these are level-2 templates, and GW2 doesn't unlock trait
  lines until ~level 11), so the bit layout is instead confirmed against
  the wiki's own worked numeric example (`0b00111001` → Adept=1,
  Master=2, Grandmaster=3), resolving the ambiguity in its prose ("2-bit
  values from 0 to 3, in reverse order").
- **Encoding** (`EncodeBuildTemplate`, `EncodeSimpleIDLink`) — verified by
  round-tripping all 4 real samples (decode → encode → decode,
  byte-for-byte identical) plus synthetic round-trips for Revenant
  legends, weapons, and skill overrides.

**Gaps, not yet covered by any real sample:**

- A real code with a trait actually chosen (all 4 samples are
  level-2/no-traits).
- Revenant legend bytes and the skill-override array (Weaponmaster
  Training) — implemented per the wiki spec, including the public API's
  `/v2/legends` `code` field for legend values, but only exercised by
  structural/round-trip tests (`TestDecodeBuildTemplate_RevenantLegends`,
  `TestDecodeBuildTemplate_SkillOverrides`), not a real exported code.
- 4+ weapons across two weapon sets, a 4-pet Ranger build, and an
  underwater-loadout variant.

A real-sample wishlist for closing these gaps, with exact export
instructions per item, lives outside this repo (HeroAscent workspace:
`docs/GW2_CHATLINKS_GO_BUILD_WISHLIST.md`).

## Skill/trait/item/recipe/achievement/map links

`DecodeSimpleIDLink`/`EncodeSimpleIDLink` handle skill (`0x06`), trait
(`0x07`), item (`0x02`), recipe (`0x09`), achievement (`0x0E`), and
map/point-of-interest (`0x04`) links — all 6 share the same structural
format per the wiki spec: header + ID, optionally a quantity byte for
items, always a trailing zero byte.

`chatlinks/testdata/realworld_fixtures.json` holds 416 real samples
pulled from the live GW2 API (`scripts/gather_fixtures.py` regenerates
it) and is checked by `chatlinks/realworld_test.go` on every `go test` —
no network access needed at test time, since the fixture data is a
static, version-controlled snapshot. Samples are chosen by *category*,
not raw random count, so rare subtypes (e.g. dyes, which are a small
fraction of the item catalog) are reliably represented rather than left
to chance:

- **360 ground-truth pairs** — `/v2/items`, `/v2/skills`, and
  `/v2/recipes` all expose a `chat_link` field directly, so these check
  decode *and* round-trip-encode against an independently published code
  string, not just a self-consistency check.
  - 126 items across 17 categories (`Armor`, `Back`, `Bag`, `Consumable`,
    `Container`, `CraftingMaterial`, `Dye`, `Gathering`, `Gizmo`,
    `JadeTechModule`, `MiniPet`, `Relic`, `Tool`, `Trinket`, `Trophy`,
    `UpgradeComponent`, `Weapon`) — dye unlocks are carved out of
    `Consumable` as their own category rather than left to chance.
  - 80 skills spread across all 9 professions plus profession-agnostic
    skills.
  - 154 recipes across 52 crafting-type categories (a smaller per-category
    count than items/skills, since recipe `type` is metadata diversity —
    it doesn't change how the chat link decodes — not code-path
    diversity).
- **56 self-consistency pairs** — `/v2/achievements` and the
  continents/floors endpoints don't expose `chat_link`, so these only
  confirm that encoding a real ID and decoding the result recovers the
  same ID.
  - 24 achievements split across Permanent/Repeatable/other.
  - 32 map points of interest across all 4 known point types
    (landmark/vista/waypoint/unlock), pulled from 5 different
    continent/floor combinations rather than just continent 1 floor 1.

### Bugs this testing found

- `EncodeSimpleIDLink` was missing the trailing zero byte that every
  ground-truth sample actually has — decode never noticed (it only reads
  through the ID and ignores anything after), but the encoder was
  silently producing shorter-than-canonical codes. The 4 real
  build-template samples never exercised this, since build templates
  compute their array lengths differently. Fixed; both
  `EncodeSimpleIDLink` test suites pass clean now.
- Achievement and map links already decoded correctly through the
  existing generic logic with no extra code — only the CLI's dispatch
  table was missing them. Fixed.

## Regenerating the fixture set

```bash
python3 scripts/gather_fixtures.py
```

Hits the live GW2 API directly (no key needed) and overwrites
`chatlinks/testdata/realworld_fixtures.json`. Not run automatically by
`go test` or CI — re-run it manually and review the diff if you want to
grow or rebalance the sample set, or refresh it after the game adds new
item/skill/recipe categories.
