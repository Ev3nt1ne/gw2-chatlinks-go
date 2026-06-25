// Command gw2-chatlinks decodes a Guild Wars 2 chat link ([&...] code) from
// the command line.
//
// Usage:
//
//	gw2-chatlinks "[&...]" [--resolve]
//
// --resolve additionally hits the public GW2 API (no API key needed) to
// translate IDs and build-template palette IDs into real names.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Ev3nt1ne/gw2-chatlinks-go/api"
	"github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"
)

var skillSlotNames = [10]string{
	"heal_terrestrial", "heal_aquatic",
	"util1_terrestrial", "util1_aquatic",
	"util2_terrestrial", "util2_aquatic",
	"util3_terrestrial", "util3_aquatic",
	"elite_terrestrial", "elite_aquatic",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gw2-chatlinks \"[&...]\" [--resolve]")
		os.Exit(1)
	}

	code := os.Args[1]
	resolve := false
	for _, arg := range os.Args[2:] {
		if arg == "--resolve" {
			resolve = true
		}
	}

	if err := run(os.Stdout, code, resolve); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(w io.Writer, code string, resolve bool) error {
	raw, err := chatlinks.DecodeRaw(code)
	if err != nil {
		return err
	}
	t := chatlinks.HeaderType(raw)
	fmt.Fprintln(w, "type:", t)

	switch t {
	case "build_template":
		return printBuildTemplate(w, code, resolve)
	case "skill", "trait", "item", "recipe":
		return printSimpleIDLink(w, code, t, resolve)
	default:
		fmt.Fprintf(w, "raw hex: %x  (decoder for this type not implemented yet)\n", raw)
		return nil
	}
}

func printBuildTemplate(w io.Writer, code string, resolve bool) error {
	bt, err := chatlinks.DecodeBuildTemplate(code)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "profession:", bt.Profession)
	for i, spec := range bt.Specializations {
		fmt.Fprintf(w, "  specialization[%d]: %+v\n", i, spec)
	}

	var client api.Client
	ctx := context.Background()
	for i, paletteID := range bt.SkillPaletteIDs {
		if paletteID == 0 {
			continue
		}
		name := skillSlotNames[i]
		if resolve {
			skillID, ok, err := client.PaletteIDToSkillID(ctx, bt.Profession, paletteID)
			if err != nil {
				return err
			}
			skillName := "?"
			if ok {
				skillName, err = client.ResolveSkillName(ctx, skillID)
				if err != nil {
					return err
				}
			}
			fmt.Fprintf(w, "  %s: %s (palette=%d, skill_id=%d)\n", name, skillName, paletteID, skillID)
		} else {
			fmt.Fprintf(w, "  %s: palette=%d\n", name, paletteID)
		}
	}
	if bt.RangerPets != nil {
		fmt.Fprintf(w, "ranger_pets: %+v\n", *bt.RangerPets)
	}
	if rl := bt.RevenantLegends; rl != nil {
		fmt.Fprintf(w, "revenant_legends: active_terrestrial=%s inactive_terrestrial=%s active_aquatic=%s inactive_aquatic=%s\n",
			legendName(rl.TerrestrialActive), legendName(rl.TerrestrialInactive),
			legendName(rl.AquaticActive), legendName(rl.AquaticInactive))
	}
	for _, weaponID := range bt.WeaponIDs {
		name, ok := chatlinks.WeaponTypes[weaponID]
		if !ok {
			name = fmt.Sprintf("unknown(%d)", weaponID)
		}
		fmt.Fprintf(w, "weapon: %s\n", name)
	}
	for _, skillID := range bt.SkillOverrideIDs {
		if resolve {
			name, err := client.ResolveSkillName(ctx, skillID)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "skill_override: %s (skill_id=%d)\n", name, skillID)
		} else {
			fmt.Fprintf(w, "skill_override: skill_id=%d\n", skillID)
		}
	}
	return nil
}

func legendName(code int) string {
	if code == 0 {
		return "(none)"
	}
	if name, ok := chatlinks.Legends[code]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", code)
}

func printSimpleIDLink(w io.Writer, code, linkType string, resolve bool) error {
	link, err := chatlinks.DecodeSimpleIDLink(code)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "id:", link.ID)
	if !resolve {
		return nil
	}

	var client api.Client
	ctx := context.Background()
	var name string
	switch linkType {
	case "skill":
		name, err = client.ResolveSkillName(ctx, link.ID)
	case "trait":
		name, err = client.ResolveTraitName(ctx, link.ID)
	case "item":
		name, err = client.ResolveItemName(ctx, link.ID)
	default:
		return fmt.Errorf("--resolve not supported for link type %q", linkType)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "name:", name)
	return nil
}
