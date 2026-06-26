package api

import (
	"context"
	"errors"
	"sort"

	"github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"
)

// ResolvedBuildTemplate holds the public-API-resolved names for a build
// template's ID-bearing fields. IDs the API didn't recognize are simply
// absent from these maps (the same convention as the batch Resolve*Names
// methods) — check for missing keys rather than assuming every input ID is
// present. All three maps are non-nil even when empty or when resolution
// failed, so callers can range over them unconditionally.
type ResolvedBuildTemplate struct {
	// PaletteToSkillID maps each non-zero skill palette ID in the build to
	// its public-API skill ID.
	PaletteToSkillID map[int]int
	// SkillNames maps a public-API skill ID to its name, covering both
	// palette-derived skill IDs (via PaletteToSkillID) and the build's
	// SkillOverrideIDs.
	SkillNames map[int]string
	// SpecializationNames maps a specialization ID to its name.
	SpecializationNames map[int]string
}

// ResolveBuildTemplate resolves a decoded build template's IDs to names in as
// few public-API requests as possible: one profession-document fetch (palette
// IDs -> skill IDs), one batched skill-name lookup (palette-derived skill IDs
// merged with the build's skill overrides), and one batched specialization-
// name lookup — at most three requests for a fully-loaded build, versus one
// request per skill slot/override/spec if the single-ID methods are looped.
//
// The three lookups are independent: a failure in one does not abort the
// others, and whatever resolved successfully is still returned alongside a
// joined error (via errors.Join) describing what didn't. This lets callers
// treat resolution as best-effort (decode is the real answer; names are
// enrichment) without losing partial results. IDs the API doesn't recognize
// are absent from the result maps, which is not itself an error.
//
// Weapons, legends, and Ranger pets are not resolved here: weapon and legend
// names come from the static tables in the chatlinks package
// (chatlinks.WeaponTypeName / chatlinks.LegendName, no network), and there is
// no pet-name endpoint wired up. Revenant inactive-utility palette IDs are
// likewise not resolved yet.
func (c *Client) ResolveBuildTemplate(ctx context.Context, bt chatlinks.BuildTemplate) (ResolvedBuildTemplate, error) {
	out := ResolvedBuildTemplate{
		PaletteToSkillID:    map[int]int{},
		SkillNames:          map[int]string{},
		SpecializationNames: map[int]string{},
	}
	var errs []error

	// 1. Skill palette IDs -> skill IDs (one /v2/professions/{p} fetch,
	//    regardless of how many palette slots are set).
	paletteToSkill, err := c.PaletteIDsToSkillIDs(ctx, bt.Profession, bt.SkillPaletteIDs[:])
	if err != nil {
		errs = append(errs, err)
	} else {
		out.PaletteToSkillID = paletteToSkill
	}

	// 2. Skill IDs -> names. Merge palette-derived skill IDs with the build's
	//    skill overrides into one deduplicated batch.
	skillIDSet := make(map[int]bool, len(out.PaletteToSkillID)+len(bt.SkillOverrideIDs))
	for _, skillID := range out.PaletteToSkillID {
		skillIDSet[skillID] = true
	}
	for _, skillID := range bt.SkillOverrideIDs {
		if skillID != 0 {
			skillIDSet[skillID] = true
		}
	}
	if len(skillIDSet) > 0 {
		skillIDs := sortedKeys(skillIDSet)
		names, err := c.ResolveSkillNames(ctx, skillIDs)
		if err != nil {
			errs = append(errs, err)
		} else {
			out.SkillNames = names
		}
	}

	// 3. Specialization IDs -> names.
	specIDSet := make(map[int]bool, len(bt.Specializations))
	for _, spec := range bt.Specializations {
		if spec.SpecializationID != 0 {
			specIDSet[spec.SpecializationID] = true
		}
	}
	if len(specIDSet) > 0 {
		names, err := c.ResolveSpecializationNames(ctx, sortedKeys(specIDSet))
		if err != nil {
			errs = append(errs, err)
		} else {
			out.SpecializationNames = names
		}
	}

	return out, errors.Join(errs...)
}

// sortedKeys returns the keys of set in ascending order, so the request this
// feeds is deterministic (stable ?ids= ordering) rather than depending on Go
// map iteration order.
func sortedKeys(set map[int]bool) []int {
	keys := make([]int, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
