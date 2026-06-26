// Package api provides optional enrichment for the chatlinks package: it
// resolves numeric IDs (and build-template "palette IDs") to human-readable
// names via the public Guild Wars 2 API. No API key is needed for any
// endpoint used here.
//
// This is kept separate from the chatlinks package so that decoding itself
// never needs network access.
//
// # GW2 API etiquette
//
// The API rate-limits per IP — not per key, not per app — so all software
// sharing an outbound IP shares one budget. The limit itself is not fixed:
// it has been observed to differ from what's commonly documented (see
// https://wiki.guildwars2.com/wiki/API:Best_practices), so prefer reading
// RateLimitError.Limit for the live value over hardcoding one.
//
// This package never retries automatically on a 429 — a hidden retry loop
// could surprise a caller with unexpected latency — so callers that want to
// back off should use RateLimitError.RetryAfter/Limit themselves. Callers
// resolving more than one ID should prefer the batch methods
// (ResolveSkillNames, ResolveTraitNames, ResolveItemNames,
// ResolveSpecializationNames, PaletteIDsToSkillIDs) over looping the
// single-ID methods — they follow ArenaNet's documented ID-batching
// recommendation (up to 200 ids per request, chunked automatically here)
// instead of issuing one request per ID.
package api
