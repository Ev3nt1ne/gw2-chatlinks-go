package chatlinks_test

import (
	"errors"
	"fmt"

	"github.com/Ev3nt1ne/gw2-chatlinks-go/chatlinks"
)

// Decode a build template (header 0x0D) link into its profession, weapons,
// and skill palette IDs. This is the mandatory level-2 Thief template from the
// Heroes Ascent ruleset (Dagger + Rifle).
func ExampleDecodeBuildTemplate() {
	const code = "[&DQUAAAAAAAAkDyQPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACLwBVAAA=]"

	bt, err := chatlinks.DecodeBuildTemplate(code)
	if err != nil {
		panic(err)
	}

	fmt.Println("profession:", bt.Profession)
	for _, weaponID := range bt.WeaponIDs {
		name, _ := chatlinks.WeaponTypeName(weaponID)
		fmt.Println("weapon:", name)
	}
	fmt.Println("heal palette id:", bt.SkillPaletteIDs[0])

	// EncodeBuildTemplate round-trips back to the same code.
	got, err := chatlinks.EncodeBuildTemplate(bt)
	if err != nil {
		panic(err)
	}
	fmt.Println("round-trips:", got == code)

	// Output:
	// profession: Thief
	// weapon: Dagger
	// weapon: Rifle
	// heal palette id: 3876
	// round-trips: true
}

// Decode a "single ID" link (here, an item link, header 0x02), which carries a
// stack quantity before the item ID.
func ExampleDecodeSimpleIDLink() {
	link, err := chatlinks.DecodeSimpleIDLink("[&AgHwdwAA]")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s id=%d quantity=%d\n", link.LinkType, link.ID, link.Quantity)
	// Output:
	// item id=30704 quantity=1
}

// Sentinel errors let callers classify a failure with errors.Is instead of
// matching message text — e.g. to detect that a code is simply the wrong
// link type for the decoder being tried.
func ExampleErrWrongHeader() {
	// A coin link (header 0x01) is not a build template.
	_, err := chatlinks.DecodeBuildTemplate("[&AQAAAAA=]")
	fmt.Println(errors.Is(err, chatlinks.ErrWrongHeader))
	// Output:
	// true
}
