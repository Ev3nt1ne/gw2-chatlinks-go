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

Prebuilt binaries for Linux/macOS/Windows (amd64 + arm64) are attached to
each [tagged release](https://github.com/Ev3nt1ne/gw2-chatlinks-go/releases)
for non-Go users.

## Usage

CLI — decode offline (no network):

```console
$ gw2-chatlinks "[&DQUAAAAAAAAkDyQPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACLwBVAAA=]"
type: build_template
profession: Thief
  specialization[0]: {SpecializationID:0 Adept:0 Master:0 Grandmaster:0}
  ...
weapon: Dagger
weapon: Rifle
```

The code can also be piped on stdin, and `--json` emits a machine-parseable
form:

```bash
echo "[&AgHwdwAA]" | gw2-chatlinks --json
gw2-chatlinks --resolve "[&...]"   # add real names via the public GW2 API
gw2-chatlinks --version
```

Flags (long `--flag` or short `-flag`, before or after the code):

| flag        | effect                                                          |
|-------------|----------------------------------------------------------------|
| `--resolve` | resolve IDs / palette IDs to names via the public GW2 API (network) |
| `--json`    | emit the decoded result as JSON instead of text                |
| `--version` | print the version and exit                                     |

Library:

```go
import (
    "errors"
    "github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"
)

bt, err := chatlinks.DecodeBuildTemplate("[&...]")
if errors.Is(err, chatlinks.ErrWrongHeader) {
    // not a build template — try a different decoder
} else if err != nil {
    log.Fatal(err)
}
fmt.Println(bt.Profession, bt.SkillPaletteIDs, bt.WeaponIDs)

code, err := chatlinks.EncodeBuildTemplate(bt) // round-trips back to "[&...]"
```

Decoders/encoders wrap sentinel errors (`ErrInvalidPayload`, `ErrWrongHeader`,
`ErrTruncated`, `ErrUnknownLinkType`, `ErrValueOutOfRange`) so failures can be
classified with `errors.Is` instead of string-matching. Runnable examples are
on [pkg.go.dev](https://pkg.go.dev/github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks).

`--resolve` / the `api` package hit the **public** GW2 API (no API key
needed) to translate IDs and build-template "palette IDs" into real names.
Note: `/v2/professions` only returns the `skills_by_palette` field if you
pass an explicit schema version (`?v=...`) — the unversioned default omits
it. `api.Client` handles this; if you call the GW2 API directly elsewhere,
don't forget it. `api.Client` always sends a descriptive default
User-Agent; set `Client.UserAgent` if you're embedding this client in your
own application and want outbound requests attributed to your own identity
instead (see the field's doc comment for why an "if-empty" check in your
own `http.RoundTripper` won't work for this).

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

## Rate limits & batching

The GW2 API rate-limits per IP (shared across all software using that IP,
not just this library) and the limit itself is not fixed — see
[API:Best_practices](https://wiki.guildwars2.com/wiki/API:Best_practices).
`api.Client` follows ArenaNet's documented ID-batching recommendation:

```go
var client api.Client
names, err := client.ResolveSkillNames(ctx, []int{12345, 23456, 34567})
// one (or, above 200 ids, a few chunked) request instead of one per id

paletteToSkill, err := client.PaletteIDsToSkillIDs(ctx, "Thief", bt.SkillPaletteIDs[:])
// fetches /v2/professions/Thief once, regardless of how many palette ids
```

To resolve a whole decoded build template at once, `ResolveBuildTemplate`
does the orchestration for you — all of a build's skills (palette-derived and
overrides) plus specializations in **at most three requests**, returning maps
of resolved names (unrecognized IDs are simply absent):

```go
resolved, err := client.ResolveBuildTemplate(ctx, bt) // bt from chatlinks.DecodeBuildTemplate
// resolved.PaletteToSkillID, resolved.SkillNames, resolved.SpecializationNames
```

The three lookups are independent and best-effort: if one fails the others
still populate, and the error is returned alongside the partial result.

A 429 response surfaces as a `*api.RateLimitError` (wrapping
`api.ErrRateLimited`, classifiable via `errors.Is`), carrying `RetryAfter`
and the live `Limit` value when the server sends them. This package never
retries automatically on a 429 — that's a deliberate choice so a hidden
retry loop can't surprise a caller with unexpected latency; back off using
`RetryAfter` yourself if you need that. The CLI's `--resolve` already uses
the batch methods (a build template resolves in ~2 requests total, not one
per skill slot).

## Deferred

Intentionally out of scope for now — doesn't affect decode/encode
correctness, lives in the optional `api` enrichment path:

- **`api` pins the GW2 schema version to `latest`.** Convenient, but a
  future ArenaNet schema change could shift field shapes underneath callers.
  A date-stamped pinned schema would be more defensive; the tradeoff is
  documented inline in `api/api.go`.

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
