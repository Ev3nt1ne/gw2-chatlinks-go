# Verification

What's been checked against real game data, what's implemented from the
[wiki spec](https://wiki.guildwars2.com/wiki/Chat_link_format) alone and
not yet sample-verified, and where the test fixtures come from.

## Build templates (`0x0D`)

Originated as a Go port of an earlier Python prototype
(`gw2-chatlinks-py`, not published — kept locally as a reference
implementation), since extended with encoding, weapon/skill-override
arrays, and Revenant legends. Verified against 4 real level-2
build-template codes plus 6 further real samples covering trait choices,
a distinct land/water skill loadout, Revenant legends, and a second
Ranger pet build (see `chatlinks/chatlinks_test.go`):

- **Header byte and profession byte mapping** (`0x0D` = build template,
  1-9 → Guardian..Revenant) — verified empirically against real codes.
- **Skill palette slot order** (heal/utility×3/elite, terrestrial +
  aquatic) and palette→real-skill-ID resolution — end-to-end tested:
  Thief→Skelk Venom, Elementalist→Arcane Brilliance, Engineer→A.E.D., all
  genuine non-passive heal skills.
- **Ranger pet decoding** preserves slot position (`RangerPets`, a
  4-field struct) rather than compacting non-zero slots into a flat list,
  which would shift later pets into earlier slots on round-trip if any
  slot were actually unset. Checked against 2 independent real samples
  (`rangerPetSample`, `untamedPetsSample`), both with all 4 slots
  distinct and nonzero.
- **Weapon array decoding** — cross-checked against independently
  verified game facts, not just the spec: the Engineer sample decodes to
  Rifle+Hammer, consistent with the wiki's own Scrapper article (Scrapper
  grants Hammer access). Thief (Dagger+Rifle, Rifle via Deadeye), Ranger
  (Longbow+Greatsword, both core weapons), and Elementalist (Staff) check
  out the same way.
- **Specialization/trait-tier bit-packing** (`Adept`/`Master`/
  `Grandmaster`) — the 4 level-2 samples have zero specializations chosen
  (expected: GW2 doesn't unlock trait lines until ~level 11), so the bit
  layout was originally confirmed only against the wiki's own worked
  numeric example (`0b00111001` → Adept=1, Master=2, Grandmaster=3). Now
  also verified against 4 real samples with actual trait picks
  (`weaverTraitSample`, `evokerSample1`, `evokerSample2`,
  `catalystSample`) — every specialization slot's Adept/Master/Grandmaster
  values were independently cross-checked against the plain-text
  description given alongside each exported code and matched exactly.
  `evokerSample1`/`evokerSample2` are Evoker, a *different* Elementalist
  elite spec from Catalyst (Visions of Eternity, not End of Dragons —
  confirmed via the wiki rather than assumed, since it postdates this
  package's original reference knowledge); its familiar pick
  (Fox/Otter/Hare/Toad) was checked and is **not** stored in the build
  template at all — both samples' `ProfessionBytes` decode all-zero and
  byte-identical despite picking different familiars, the same situation
  as Catalyst's Jade Sphere element choice.
- **Land/water skill loadout** — `catalystSample` has a genuinely
  different skill on every single terrestrial/aquatic pair (heal, all 3
  utilities, elite), confirmed by decoding it: e.g. heal_terrestrial =
  Signet of Restoration vs. heal_aquatic = Ether Renewal, elite_terrestrial
  = Glyph of Elementals vs. elite_aquatic = Tornado. Closes the
  underwater-loadout gap below without needing a separate sample.
- **Revenant legends** — `revenantLegendsRealSample` is a real build with
  the active and inactive legend genuinely different from each other on
  *both* land and water (terrestrial active=Renegade/inactive=Assassin,
  aquatic active=Dwarf/inactive=Assassin), matching the human's
  description for the legend bytes and both other specialization tiers
  exactly. One discrepancy: the human's note says "3 3 3" for the
  elite-spec tier, but the link decodes to 2-2-2 there, consistently
  across repeated decodes — most likely a one-line transcription slip in
  the note (everything else in the same sample matches), not a decoder
  bug; the test asserts the real decoded value.
- **Skill-override array (Weaponmaster Training)** —
  `weaponmasterVariantSample` is a real Ranger build wielding Hammer (not
  a native Ranger weapon) with 3 skill-override entries. Verified against
  the live API, not just the sample's own internal consistency: all 3
  resolve to real skills named "Unleashed Wild Swing/Savage Shock
  Wave/Thump" and are tagged `specialization: 72` (Untamed) by the API
  itself, while this build's own elite spec (decoded from the same
  sample) is 78 (Galeshot) — a different spec. That mismatch is the
  override mechanic working as intended: a Hammer skill variant borrowed
  from a spec the build isn't using.
- **Weapon array only stores terrestrial weapons** — confirmed against
  the wiki's [Chat link format](https://wiki.guildwars2.com/wiki/Chat_link_format)
  page, which states this explicitly for the build-template weapon
  section: "The first byte indicates the number of terrestrial weapons
  stored in the code... Aquatic weapons are not stored." This is a
  limitation of the game's own chat-link format, not a gap in this
  package — `WeaponIDs` correctly reflects everything the format
  contains. Confirmed empirically too: 4 real samples whose human-given
  descriptions mentioned an aquatic weapon (trident, spear, harpoon gun)
  all decode to exactly the terrestrial weapons and nothing else, with
  the array's own length byte (checked at the raw-byte level) agreeing —
  not a decoder truncation. The same wiki passage also confirms a build
  with two of the same weapon type in one set (e.g. dual daggers) stores
  that type only once, not twice — also matches every real sample.
- **Encoding** (`EncodeBuildTemplate`, `EncodeSimpleIDLink`) — verified by
  round-tripping all 11 real samples (decode → encode → decode,
  byte-for-byte identical) plus synthetic round-trips for Revenant
  legends, weapons, and skill overrides.

**Gaps, not yet covered by any real sample:**

- 4+ weapons across two weapon sets. The array is verified correct above
  and is capped at 8 entries, but since it only stores terrestrial
  weapons (see above), reaching more than 2 entries needs distinct
  weapon types across *both land weapon sets* — an aquatic weapon won't
  help, no matter how it's equipped.

Equipment templates (armor/runes/trinkets) were considered for inclusion
but are out of scope: confirmed against the wiki's [Chat link
format](https://wiki.guildwars2.com/wiki/Chat_link_format) page that
`0x0D` (build template) is the *only* template-related chat-link header
that exists — there's no way to export equipment-template data as a
chat link at all, regardless of tooling.

A real-sample wishlist for closing the remaining gaps, with exact export
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
