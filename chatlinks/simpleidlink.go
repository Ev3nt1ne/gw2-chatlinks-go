package chatlinks

import "fmt"

// SimpleIDLink represents the common "single ID" link shapes: skill (0x06),
// trait (0x07), item (0x02), recipe (0x09), etc.
type SimpleIDLink struct {
	LinkType string
	ID       int

	// Quantity is only meaningful when LinkType == "item"; it's the stack
	// size encoded immediately before the item ID. Zero/unset for other
	// link types.
	Quantity int
}

// DecodeSimpleIDLink decodes a skill/trait/item/recipe-shaped link. Item
// links carry a quantity byte before the ID; everything else here doesn't.
func DecodeSimpleIDLink(code string) (SimpleIDLink, error) {
	raw, err := DecodeRaw(code)
	if err != nil {
		return SimpleIDLink{}, err
	}
	t := HeaderType(raw)
	offset := 1
	if t == "item" {
		offset = 2
	}
	if len(raw) < offset+3 {
		return SimpleIDLink{}, fmt.Errorf("chatlinks: payload too short for %s link: got %d bytes, need at least %d", t, len(raw), offset+3)
	}
	link := SimpleIDLink{LinkType: t, ID: u24le(raw, offset)}
	if t == "item" {
		link.Quantity = int(raw[1])
	}
	return link, nil
}

// EncodeSimpleIDLink encodes a skill/trait/item/recipe-shaped link.
func EncodeSimpleIDLink(link SimpleIDLink) (string, error) {
	header, ok := linkTypeToHeader[link.LinkType]
	if !ok {
		return "", fmt.Errorf("chatlinks: unknown link type %q", link.LinkType)
	}
	if link.ID < 0 || link.ID > 0xFFFFFF {
		return "", fmt.Errorf("chatlinks: id %d out of range for a 3-byte id", link.ID)
	}

	if link.LinkType == "item" {
		quantity := link.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		if quantity > 0xFF {
			return "", fmt.Errorf("chatlinks: quantity %d out of range for a 1-byte field", quantity)
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
