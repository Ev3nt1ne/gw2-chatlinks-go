# gw2-chatlinks-go

[![CI](https://github.com/Ev3nt1ne/gw2-chatlinks-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Ev3nt1ne/gw2-chatlinks-go/actions/workflows/ci.yml)

A Go library and CLI for decoding and encoding Guild Wars 2 chat links
(`[&...]` codes) — build templates, skills, traits, items, and more.
General-purpose, not tied to any one project or tool.

Format reference: [GW2 Wiki — Chat link format](https://wiki.guildwars2.com/wiki/Chat_link_format).

## Install

```bash
go get github.com/Ev3nt1ne/gw2-chatlinks-go
```

Or install the CLI directly:

```bash
go install github.com/Ev3nt1ne/gw2-chatlinks-go/cmd/gw2-chatlinks@latest
```

## Usage

CLI:

```bash
gw2-chatlinks "[&DQUAAAAAAAAkDyQPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACLwBVAAA=]" --resolve
```

Library:

```go
import "github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"

bt, err := chatlinks.DecodeBuildTemplate("[&...]")
if err != nil {
    log.Fatal(err)
}
fmt.Println(bt.Profession, bt.SkillPaletteIDs, bt.WeaponIDs)

code, err := chatlinks.EncodeBuildTemplate(bt) // round-trips back to "[&...]"
```

`--resolve` / the `api` package hit the **public** GW2 API (no API key
needed) to translate IDs and build-template "palette IDs" into real names.
Note: `/v2/professions` only returns the `skills_by_palette` field if you
pass an explicit schema version (`?v=...`) — the unversioned default omits
it. `api.Client` handles this; if you call the GW2 API directly elsewhere,
don't forget it.

## What's verified vs. not

This is a Go port of an earlier Python prototype (`gw2-chatlinks-py`, not
published — kept locally as a reference implementation), since extended
with encoding, weapon/skill-override arrays, and Revenant legends. Tested
against 4 real build-template codes (see `chatlinks/chatlinks_test.go`):

- ✅ Header byte (`0x0D` = build template) and profession byte mapping
  (1-9 → Guardian..Revenant) — verified empirically against real codes. An
  earlier AI-summarized search snippet claimed different header values than
  the live wiki page; don't trust search-snippet summaries of this spec,
  fetch the actual wiki page.
- ✅ Skill palette slot order (heal/utility×3/elite, terrestrial+aquatic) and
  palette→real-skill-ID resolution (end-to-end tested: Thief→Skelk Venom,
  Elementalist→Arcane Brilliance, Engineer→A.E.D., all genuine non-passive
  heal skills).
- ✅ Ranger pet ID decoding (profession-specific bytes), now preserving slot
  position (`RangerPets`, a 4-field struct) rather than the original Python
  prototype's lossy "drop empty slots into a flat list" approach, which
  would have shifted later pets into earlier slots on round-trip if any
  slot were actually unset.
- ✅ **Weapon array decoding** — not just structurally per the wiki spec, but
  cross-checked against independently-verified game facts: the Engineer
  sample decodes to Rifle+Hammer, which only makes sense once you confirm
  (via the wiki's own Scrapper article, not assumed from memory) that
  Scrapper grants Hammer access. Thief (Dagger+Rifle, Rifle via Deadeye),
  Ranger (Longbow+Greatsword, both core weapons), and Elementalist (Staff)
  check out the same way.
- ✅ **Specialization/trait-tier bit-packing** (`Adept`/`Master`/`Grandmaster`)
  — all 4 real samples have zero specializations chosen (correct: these are
  level-2 templates, and GW2 doesn't unlock trait lines until ~level 11), so
  this is confirmed against the wiki's own worked numeric example
  (`0b00111001` → Adept=1, Master=2, Grandmaster=3) instead, resolving the
  ambiguity in its prose ("2-bit values from 0 to 3, in reverse order").
  Still not cross-checked against a real exported code with a trait chosen —
  welcome to close that gap with a real sample.
- ⚠️ **Revenant legend bytes and the skill-override array are implemented
  per the wiki spec (including the public API's `/v2/legends` `code` field
  for legend values) but not yet exercised by any real sample** — no
  Revenant or Weaponmaster-Training-variant code was available to test
  against. Covered by structural/round-trip tests only (see
  `TestDecodeBuildTemplate_RevenantLegends`,
  `TestDecodeBuildTemplate_SkillOverrides`). A real sample of either would
  be a welcome contribution.
- ✅ **Encoding** (`EncodeBuildTemplate`, `EncodeSimpleIDLink`) — verified by
  round-tripping all 4 real samples (decode → encode → decode, byte-for-byte
  identical) plus synthetic round-trips for Revenant legends, weapons, and
  skill overrides.

### Skill/trait/item/recipe/achievement/map links — 240 real samples

`chatlinks.DecodeSimpleIDLink`/`EncodeSimpleIDLink` handle skill (`0x06`),
trait (`0x07`), item (`0x02`), recipe (`0x09`), achievement (`0x0E`), and
map/point-of-interest (`0x04`) links — all 6 share the same structural
format (per the wiki spec: header + ID, optionally a quantity byte for
items, always a trailing zero byte). `chatlinks/testdata/realworld_fixtures.json`
holds 240 real samples pulled from the live GW2 API
(`chatlinks/testdata/gather_fixtures.py` regenerates it) and is checked by
`chatlinks/realworld_test.go` on every `go test` — no network access needed
at test time, since the fixture data is a static, version-controlled
snapshot:

- **210 ground-truth pairs** (90 items, 60 skills, 60 recipes) — `/v2/items`,
  `/v2/skills`, and `/v2/recipes` all expose a `chat_link` field directly, so
  these check decode *and* round-trip-encode against an independently
  published code string, not just a self-consistency check.
- **30 self-consistency pairs** (15 achievements, 15 map points of
  interest) — `/v2/achievements` and the continents/floors endpoints don't
  expose `chat_link`, so these only confirm that encoding a real ID and
  decoding the result recovers the same ID.

This run **found and fixed a real bug**: `EncodeSimpleIDLink` was missing
the trailing zero byte that every one of the 210 ground-truth samples
actually has — decode never noticed (it only reads through the ID and
ignores anything after), but the encoder was silently producing
shorter-than-canonical codes. The 4-real-build-template-sample testing
never exercised this, since build templates compute their array lengths
differently. Fixed; both `EncodeSimpleIDLink` test suites pass clean now.

It also surfaced that achievement and map links decode correctly through
the existing generic logic with no extra code — only the CLI's dispatch
table was missing them, now fixed.

## Development

```bash
go build ./...
go vet ./...
go test ./... -race -cover
```

CI runs build, `go vet`, `golangci-lint` (including `gosec`), and tests with
coverage on every push/PR, across a Linux/macOS/Windows build matrix.

## License

MIT — see [LICENSE](LICENSE). Independent of `gw2-mcp`'s AGPL-3.0, so this
library stays maximally reusable by anything, regardless of that consumer's
own license.
