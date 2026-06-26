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
// the skill-override (Weaponmaster Training) array are now also verified
// against real samples — see VERIFICATION.md and chatlinks_test.go for the
// full coverage breakdown.
//
// Decoders return errors wrapping the package's sentinel error values
// (ErrInvalidPayload, ErrWrongHeader, ErrTruncated) and encoders wrap
// ErrUnknownLinkType / ErrValueOutOfRange, so callers can classify failures
// with errors.Is rather than matching message text.
package chatlinks
