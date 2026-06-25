package chatlinks

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// StripLink removes the "[&...]" wrapper, if present, returning the bare
// base64 payload.
func StripLink(code string) string {
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "[&") && strings.HasSuffix(code, "]") {
		return code[2 : len(code)-1]
	}
	return code
}

// WrapLink wraps a base64 payload in the "[&...]" chat link syntax.
func WrapLink(b64 string) string {
	return "[&" + b64 + "]"
}

// DecodeRaw strips the [&...] wrapper (if present) and base64-decodes the
// payload. GW2 chat links sometimes omit base64 padding, so it is added
// back as needed.
func DecodeRaw(code string) ([]byte, error) {
	b64 := StripLink(code)
	if pad := (4 - len(b64)%4) % 4; pad != 0 {
		b64 += strings.Repeat("=", pad)
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("chatlinks: invalid base64 payload: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("chatlinks: empty payload")
	}
	return raw, nil
}

// encodeRaw base64-encodes raw and wraps it in "[&...]" syntax.
func encodeRaw(raw []byte) string {
	return WrapLink(base64.StdEncoding.EncodeToString(raw))
}

// HeaderType returns the link type name for a decoded payload's first byte.
func HeaderType(raw []byte) string {
	if len(raw) == 0 {
		return "unknown(empty)"
	}
	if t, ok := headerTypes[raw[0]]; ok {
		return t
	}
	return fmt.Sprintf("unknown(0x%02x)", raw[0])
}

func u16le(raw []byte, offset int) int {
	return int(raw[offset]) | int(raw[offset+1])<<8
}

func u24le(raw []byte, offset int) int {
	return int(raw[offset]) | int(raw[offset+1])<<8 | int(raw[offset+2])<<16
}

func u32le(raw []byte, offset int) int {
	return int(raw[offset]) | int(raw[offset+1])<<8 | int(raw[offset+2])<<16 | int(raw[offset+3])<<24
}

func putU16le(buf []byte, offset, value int) {
	buf[offset] = byte(value)
	buf[offset+1] = byte(value >> 8)
}

func putU24le(buf []byte, offset, value int) {
	buf[offset] = byte(value)
	buf[offset+1] = byte(value >> 8)
	buf[offset+2] = byte(value >> 16)
}

func putU32le(buf []byte, offset, value int) {
	buf[offset] = byte(value)
	buf[offset+1] = byte(value >> 8)
	buf[offset+2] = byte(value >> 16)
	buf[offset+3] = byte(value >> 24)
}
