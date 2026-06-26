package chatlinks

import "fmt"

// simpleIDLinkTypes is the set of link types DecodeSimpleIDLink /
// EncodeSimpleIDLink actually support: those shaped as "header (+ optional
// 1-byte quantity) + 3-byte id". Other known headers (coin, text, pvp_game,
// user, build/wardrobe templates, etc.) have entirely different layouts and
// must not be (mis)handled here.
var simpleIDLinkTypes = map[string]bool{
	"map":         true, // 0x04
	"skill":       true, // 0x06
	"trait":       true, // 0x07
	"item":        true, // 0x02 (carries a quantity byte before the id)
	"recipe":      true, // 0x09
	"achievement": true, // 0x0E
}

// SimpleIDLink represents the common "single ID" link shapes: skill (0x06),
// trait (0x07), item (0x02), recipe (0x09), etc.
type SimpleIDLink struct {
	LinkType string `json:"link_type"`
	ID       int    `json:"id"`

	// Quantity is only meaningful when LinkType == "item"; it's the stack
	// size encoded immediately before the item ID. Zero/unset for other
	// link types.
	Quantity int `json:"quantity,omitempty"`
}

// DecodeSimpleIDLink decodes a map/skill/trait/item/recipe/achievement link
// (the "header + 3-byte id" shapes; see simpleIDLinkTypes). Item links carry
// a quantity byte before the ID; everything else here doesn't. Any other link
// type (coin, text, build template, ...) returns an error wrapping
// ErrWrongHeader rather than being silently mis-decoded.
func DecodeSimpleIDLink(code string) (SimpleIDLink, error) {
	raw, err := DecodeRaw(code)
	if err != nil {
		return SimpleIDLink{}, err
	}
	t := HeaderType(raw)
	if !simpleIDLinkTypes[t] {
		return SimpleIDLink{}, fmt.Errorf("%w: %s is not an id-shaped link type", ErrWrongHeader, t)
	}
	offset := 1
	if t == "item" {
		offset = 2
	}
	if len(raw) < offset+3 {
		return SimpleIDLink{}, fmt.Errorf("%w: payload too short for %s link: got %d bytes, need at least %d", ErrTruncated, t, len(raw), offset+3)
	}
	link := SimpleIDLink{LinkType: t, ID: u24le(raw, offset)}
	if t == "item" {
		link.Quantity = int(raw[1])
	}
	return link, nil
}

// EncodeSimpleIDLink encodes a map/skill/trait/item/recipe/achievement link
// (the "header + 3-byte id" shapes; see simpleIDLinkTypes).
func EncodeSimpleIDLink(link SimpleIDLink) (string, error) {
	if !simpleIDLinkTypes[link.LinkType] {
		return "", fmt.Errorf("%w: %q is not an id-shaped link type", ErrUnknownLinkType, link.LinkType)
	}
	header := linkTypeToHeader[link.LinkType]
	if link.ID < 0 || link.ID > 0xFFFFFF {
		return "", fmt.Errorf("%w: id %d out of range for a 3-byte id", ErrValueOutOfRange, link.ID)
	}

	if link.LinkType == "item" {
		quantity := link.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		if quantity > 0xFF {
			return "", fmt.Errorf("%w: quantity %d out of range for a 1-byte field", ErrValueOutOfRange, quantity)
		}
		// header + quantity + 3-byte id + 1 trailing zero byte. The
		// trailing byte is real, not assumed: every item chat_link the
		// live GW2 API returns ends in one (see chatlinks/testdata).
		raw := make([]byte, 6)
		raw[0] = header
		raw[1] = byte(quantity)
		putU24le(raw, 2, link.ID)
		return encodeRaw(raw), nil
	}

	// header + 3-byte id + 1 trailing zero byte — same real-data basis as
	// the item case above.
	raw := make([]byte, 5)
	raw[0] = header
	putU24le(raw, 1, link.ID)
	return encodeRaw(raw), nil
}
