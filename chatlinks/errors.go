package chatlinks

import "errors"

// Sentinel errors returned (wrapped) by this package. Callers can classify a
// failure with errors.Is without substring-matching the message text, e.g. to
// fall back to a different decoder or to tell user error from corrupt data:
//
//	bt, err := chatlinks.DecodeBuildTemplate(code)
//	switch {
//	case errors.Is(err, chatlinks.ErrWrongHeader):
//	    // not a build template — try another decoder
//	case errors.Is(err, chatlinks.ErrTruncated):
//	    // the code is malformed/cut off
//	}
//
// The wrapped messages still carry the specific detail; only the kind is
// promised as a stable contract.
var (
	// ErrInvalidPayload means the "[&...]" payload could not be
	// base64-decoded, or decoded to zero bytes.
	ErrInvalidPayload = errors.New("chatlinks: invalid payload")

	// ErrWrongHeader means the decoded payload's leading header byte does
	// not match the link type the called decoder handles (e.g. a coin link
	// passed to DecodeBuildTemplate, or any non-ID-shaped link passed to
	// DecodeSimpleIDLink).
	ErrWrongHeader = errors.New("chatlinks: wrong header type")

	// ErrTruncated means the payload ended before all the bytes a link of
	// its declared type/shape requires were present.
	ErrTruncated = errors.New("chatlinks: truncated payload")

	// ErrUnknownLinkType means an encoder was asked for a link-type name it
	// has no mapping for.
	ErrUnknownLinkType = errors.New("chatlinks: unknown link type")

	// ErrValueOutOfRange means a field value passed to an encoder does not
	// fit the fixed-width field it must be written into, which would have
	// silently produced a corrupt (valid-looking but wrong) chat link.
	ErrValueOutOfRange = errors.New("chatlinks: value out of range")
)
