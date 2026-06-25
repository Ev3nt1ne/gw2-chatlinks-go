# gw2-chatlinks-go

[![CI](https://github.com/Ev3nt1ne/gw2-chatlinks-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Ev3nt1ne/gw2-chatlinks-go/actions/workflows/ci.yml)

A Go library and CLI for decoding Guild Wars 2 chat links (`[&...]` codes) —
build templates, skills, traits, items, and more. General-purpose, not tied
to any one project or tool.

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
fmt.Println(bt.Profession, bt.SkillPaletteIDs)
```

`--resolve` / the `api` package hit the **public** GW2 API (no API key
needed) to translate IDs and build-template "palette IDs" into real names.
Note: `/v2/professions` only returns the `skills_by_palette` field if you
pass an explicit schema version (`?v=...`) — the unversioned default omits
it. `api.Client` handles this; if you call the GW2 API directly elsewhere,
don't forget it.

## What's verified vs. not

This is a Go port of an earlier Python prototype (`gw2-chatlinks-py`, not
published — kept locally as a reference implementation). Tested against 4
real build-template codes (see `chatlinks/chatlinks_test.go`):

- ✅ Header byte (`0x0D` = build template) and profession byte mapping
  (1-9 → Guardian..Revenant) — verified empirically against real codes. An
  earlier AI-summarized search snippet claimed different header values than
  the live wiki page; don't trust search-snippet summaries of this spec,
  fetch the actual wiki page.
- ✅ Skill palette slot order (heal/utility×3/elite, terrestrial+aquatic) and
  palette→real-skill-ID resolution (end-to-end tested: Thief→Skelk Venom,
  Elementalist→Arcane Brilliance, Engineer→A.E.D., all genuine non-passive
  heal skills).
- ✅ Ranger pet ID decoding (profession-specific bytes).
- ❌ **Specialization/trait-tier bit-packing (`Adept`/`Master`/`Grandmaster`
  fields) is NOT verified** — all 4 real samples have zero specializations
  chosen (correct: these are level-2 templates, and GW2 doesn't unlock trait
  lines until ~level 11), so that code path has never been exercised against
  real data. Don't trust it without testing against a real code that has at
  least one trait chosen, cross-checked against
  [gw2skills.net's editor](https://en.gw2skills.net/editor/).
- ❌ Weapon array and skill-override sections (documented in the wiki spec)
  are not implemented.
- ❌ Revenant legend bytes are not implemented (same profession-specific-bytes
  region as Ranger pets, different layout).
- ❌ **Encoding is not implemented yet.** Only decoding is currently supported,
  despite the repo name's "chatlinks" generality — encode support is planned
  but not started, and won't be added without real samples to verify
  round-trip correctness against (see the verification gaps above).

`chatlinks.DecodeSimpleIDLink` handles skill (`0x06`), trait (`0x07`), and
item (`0x02`) links structurally per the wiki spec — not independently
verified against real samples the way build templates were. Cross-check
important results.

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
