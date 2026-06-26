// Command gw2-chatlinks decodes a Guild Wars 2 chat link ([&...] code).
//
// Usage:
//
//	gw2-chatlinks [flags] "[&...]"
//	echo "[&...]" | gw2-chatlinks [flags]
//
// The code may be given as the single positional argument or, if omitted (or
// given as "-"), read from stdin. Flags:
//
//	--resolve   resolve IDs and build-template palette IDs to names via the
//	            public GW2 API (no API key needed; this is the only path that
//	            makes network calls).
//	--json      emit the decoded result as JSON instead of text.
//	--version   print the version and exit.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Ev3nt1ne/gw2-chatlinks-go/api"
	"github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"
)

// version is the CLI version, overridden at release build time via
// -ldflags "-X main.version=v1.2.3" (goreleaser does this automatically).
var version = "dev"

var skillSlotNames = [10]string{
	"heal_terrestrial", "heal_aquatic",
	"util1_terrestrial", "util1_aquatic",
	"util2_terrestrial", "util2_aquatic",
	"util3_terrestrial", "util3_aquatic",
	"elite_terrestrial", "elite_aquatic",
}

type options struct {
	resolve bool
	asJSON  bool
}

func main() {
	opts, code, showVersion, err := parseArgs(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		// flag has already printed the parse error and usage to stderr.
		os.Exit(2)
	}
	if showVersion {
		fmt.Println("gw2-chatlinks", version)
		return
	}

	code, err = resolveCode(code, os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if err := run(os.Stdout, code, opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// parseArgs parses CLI arguments into options plus the positional code. It
// returns showVersion=true when --version was requested. Help (-h/--help)
// surfaces as flag.ErrHelp; unknown flags and extra positionals are errors.
func parseArgs(args []string, stderr io.Writer) (opts options, code string, showVersion bool, err error) {
	fs := flag.NewFlagSet("gw2-chatlinks", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprint(stderr, "gw2-chatlinks — decode Guild Wars 2 chat links ([&...] codes)\n\n"+
			"Usage:\n"+
			"  gw2-chatlinks [flags] \"[&...]\"\n"+
			"  echo \"[&...]\" | gw2-chatlinks [flags]\n\n"+
			"If no code is given (or it is \"-\"), the link is read from stdin.\n"+
			"Flags may appear before or after the code.\n\n"+
			"Flags (long form --flag also accepted):\n")
		fs.PrintDefaults()
	}
	fs.BoolVar(&opts.resolve, "resolve", false, "resolve IDs/palette IDs to names via the public GW2 API (network)")
	fs.BoolVar(&opts.asJSON, "json", false, "emit the decoded result as JSON")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")

	// flag.Parse stops at the first non-flag argument, so to accept flags on
	// either side of the positional code (e.g. both "--resolve [&..]" and
	// "[&..] --resolve") we parse, collect one positional, and parse again
	// from what's left until nothing remains.
	var positionals []string
	rest := args
	for {
		if err = fs.Parse(rest); err != nil {
			return options{}, "", false, err
		}
		rest = fs.Args()
		if len(rest) == 0 {
			break
		}
		positionals = append(positionals, rest[0])
		rest = rest[1:]
	}
	if len(positionals) > 1 {
		fmt.Fprintf(stderr, "error: expected at most one chat link, got %d\n", len(positionals))
		return options{}, "", false, fmt.Errorf("unexpected extra arguments: %v", positionals[1:])
	}
	if len(positionals) == 1 {
		code = positionals[0]
	}
	return opts, code, showVersion, nil
}

// resolveCode returns the chat link to decode: the positional code if given
// (and not "-"), otherwise the trimmed contents of stdin.
func resolveCode(code string, stdin io.Reader) (string, error) {
	if code != "" && code != "-" {
		return code, nil
	}
	b, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("reading chat link from stdin: %w", err)
	}
	code = strings.TrimSpace(string(b))
	if code == "" {
		return "", errors.New("no chat link provided (pass one as an argument or on stdin)")
	}
	return code, nil
}

func run(w io.Writer, code string, opts options) error {
	raw, err := chatlinks.DecodeRaw(code)
	if err != nil {
		return err
	}
	t := chatlinks.HeaderType(raw)

	switch t {
	case "build_template":
		return printBuildTemplate(w, code, opts)
	case "skill", "trait", "item", "recipe", "achievement", "map":
		return printSimpleIDLink(w, code, t, opts)
	default:
		if opts.asJSON {
			return writeJSON(w, map[string]any{
				"type":    t,
				"raw_hex": fmt.Sprintf("%x", raw),
				"note":    "decoder for this type not implemented yet",
			})
		}
		fmt.Fprintln(w, "type:", t)
		fmt.Fprintf(w, "raw hex: %x  (decoder for this type not implemented yet)\n", raw)
		return nil
	}
}

// skillResult is one decoded skill-palette slot. SkillID/Name are populated
// only with --resolve.
type skillResult struct {
	Slot      string `json:"slot"`
	PaletteID int    `json:"palette_id"`
	SkillID   int    `json:"skill_id,omitempty"`
	Name      string `json:"name,omitempty"`
}

type weaponResult struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type overrideResult struct {
	SkillID int    `json:"skill_id"`
	Name    string `json:"name,omitempty"`
}

type legendsResult struct {
	ActiveTerrestrial   string `json:"active_terrestrial"`
	InactiveTerrestrial string `json:"inactive_terrestrial"`
	ActiveAquatic       string `json:"active_aquatic"`
	InactiveAquatic     string `json:"inactive_aquatic"`
}

type buildTemplateResult struct {
	Type            string                            `json:"type"`
	Profession      string                            `json:"profession"`
	ProfessionID    int                               `json:"profession_id"`
	Specializations [3]chatlinks.SpecializationChoice `json:"specializations"`
	Skills          []skillResult                     `json:"skills"`
	RangerPets      *chatlinks.RangerPets             `json:"ranger_pets,omitempty"`
	RevenantLegends *legendsResult                    `json:"revenant_legends,omitempty"`
	Weapons         []weaponResult                    `json:"weapons,omitempty"`
	SkillOverrides  []overrideResult                  `json:"skill_overrides,omitempty"`
}

func printBuildTemplate(w io.Writer, code string, opts options) error {
	bt, err := chatlinks.DecodeBuildTemplate(code)
	if err != nil {
		return err
	}

	res := buildTemplateResult{
		Type:            "build_template",
		Profession:      bt.Profession,
		ProfessionID:    bt.ProfessionID,
		Specializations: bt.Specializations,
		RangerPets:      bt.RangerPets,
	}

	var client api.Client
	ctx := context.Background()

	for i, paletteID := range bt.SkillPaletteIDs {
		if paletteID == 0 {
			continue
		}
		s := skillResult{Slot: skillSlotNames[i], PaletteID: paletteID}
		if opts.resolve {
			skillID, ok, err := client.PaletteIDToSkillID(ctx, bt.Profession, paletteID)
			if err != nil {
				return err
			}
			s.SkillID = skillID
			s.Name = "?"
			if ok {
				s.Name, err = client.ResolveSkillName(ctx, skillID)
				if err != nil {
					return err
				}
			}
		}
		res.Skills = append(res.Skills, s)
	}

	if rl := bt.RevenantLegends; rl != nil {
		res.RevenantLegends = &legendsResult{
			ActiveTerrestrial:   legendName(rl.TerrestrialActive),
			InactiveTerrestrial: legendName(rl.TerrestrialInactive),
			ActiveAquatic:       legendName(rl.AquaticActive),
			InactiveAquatic:     legendName(rl.AquaticInactive),
		}
	}

	for _, weaponID := range bt.WeaponIDs {
		name, ok := chatlinks.WeaponTypeName(weaponID)
		if !ok {
			name = fmt.Sprintf("unknown(%d)", weaponID)
		}
		res.Weapons = append(res.Weapons, weaponResult{ID: weaponID, Name: name})
	}

	for _, skillID := range bt.SkillOverrideIDs {
		o := overrideResult{SkillID: skillID}
		if opts.resolve {
			o.Name, err = client.ResolveSkillName(ctx, skillID)
			if err != nil {
				return err
			}
		}
		res.SkillOverrides = append(res.SkillOverrides, o)
	}

	if opts.asJSON {
		return writeJSON(w, res)
	}
	return res.writeText(w)
}

func (res buildTemplateResult) writeText(w io.Writer) error {
	fmt.Fprintln(w, "type:", res.Type)
	fmt.Fprintln(w, "profession:", res.Profession)
	for i, spec := range res.Specializations {
		fmt.Fprintf(w, "  specialization[%d]: %+v\n", i, spec)
	}
	for _, s := range res.Skills {
		if s.Name != "" {
			fmt.Fprintf(w, "  %s: %s (palette=%d, skill_id=%d)\n", s.Slot, s.Name, s.PaletteID, s.SkillID)
		} else {
			fmt.Fprintf(w, "  %s: palette=%d\n", s.Slot, s.PaletteID)
		}
	}
	if res.RangerPets != nil {
		fmt.Fprintf(w, "ranger_pets: %+v\n", *res.RangerPets)
	}
	if rl := res.RevenantLegends; rl != nil {
		fmt.Fprintf(w, "revenant_legends: active_terrestrial=%s inactive_terrestrial=%s active_aquatic=%s inactive_aquatic=%s\n",
			rl.ActiveTerrestrial, rl.InactiveTerrestrial, rl.ActiveAquatic, rl.InactiveAquatic)
	}
	for _, weapon := range res.Weapons {
		fmt.Fprintf(w, "weapon: %s\n", weapon.Name)
	}
	for _, o := range res.SkillOverrides {
		if o.Name != "" {
			fmt.Fprintf(w, "skill_override: %s (skill_id=%d)\n", o.Name, o.SkillID)
		} else {
			fmt.Fprintf(w, "skill_override: skill_id=%d\n", o.SkillID)
		}
	}
	return nil
}

func legendName(code int) string {
	if code == 0 {
		return "(none)"
	}
	if name, ok := chatlinks.LegendName(code); ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", code)
}

type simpleLinkResult struct {
	Type     string `json:"type"`
	ID       int    `json:"id"`
	Quantity int    `json:"quantity,omitempty"`
	Name     string `json:"name,omitempty"`
}

func printSimpleIDLink(w io.Writer, code, linkType string, opts options) error {
	link, err := chatlinks.DecodeSimpleIDLink(code)
	if err != nil {
		return err
	}

	res := simpleLinkResult{Type: linkType, ID: link.ID, Quantity: link.Quantity}

	if opts.resolve {
		var client api.Client
		ctx := context.Background()
		switch linkType {
		case "skill":
			res.Name, err = client.ResolveSkillName(ctx, link.ID)
		case "trait":
			res.Name, err = client.ResolveTraitName(ctx, link.ID)
		case "item":
			res.Name, err = client.ResolveItemName(ctx, link.ID)
		default:
			return fmt.Errorf("--resolve not supported for link type %q", linkType)
		}
		if err != nil {
			return err
		}
	}

	if opts.asJSON {
		return writeJSON(w, res)
	}
	fmt.Fprintln(w, "type:", res.Type)
	fmt.Fprintln(w, "id:", res.ID)
	if res.Name != "" {
		fmt.Fprintln(w, "name:", res.Name)
	}
	return nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
