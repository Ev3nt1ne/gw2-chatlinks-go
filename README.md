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

## What's covered

- Skill (`0x06`), trait (`0x07`), item (`0x02`), recipe (`0x09`),
  achievement (`0x0E`), and map/point-of-interest (`0x04`) links via
  `DecodeSimpleIDLink`/`EncodeSimpleIDLink`.
- Build templates (`0x0D`) via `DecodeBuildTemplate`/`EncodeBuildTemplate`:
  profession, specializations/trait tiers, skill palette IDs, weapon
  arrays, Weaponmaster Training skill overrides, Ranger pets, and Revenant
  legends.

See [VERIFICATION.md](VERIFICATION.md) for exactly what's been checked
against real game data versus implemented from the spec alone, and how to
keep it current as the game patches.

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
