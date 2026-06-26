package api

import (
	"errors"
	"fmt"
	"time"
)

// ErrRateLimited is wrapped by every RateLimitError. Callers can classify a
// failure with errors.Is without inspecting the concrete type:
//
//	_, err := client.ResolveSkillName(ctx, id)
//	if errors.Is(err, api.ErrRateLimited) {
//	    // back off (see RateLimitError.RetryAfter) and try again later
//	}
var ErrRateLimited = errors.New("api: rate limited (HTTP 429)")

// RateLimitError is returned when the GW2 API responds with HTTP 429. This
// package deliberately does not retry automatically — a hidden retry loop
// could surprise a caller with unexpected latency — so it's the caller's
// choice whether/how to back off. RetryAfter and Limit are best-effort: GW2
// rate-limits per IP, and the limit itself is not fixed (it has been
// observed to differ from what's commonly documented), so read these values
// live rather than assuming either is always present.
type RateLimitError struct {
	// Path is the request path that was rate limited.
	Path string

	// RetryAfter is parsed from the response's Retry-After header (sent by
	// GW2 as an integer number of seconds). Zero means the header was
	// absent or not a parseable integer — treat as "unknown", not "zero
	// wait", before retrying.
	RetryAfter time.Duration

	// Limit is the most recently observed X-Rate-Limit-Limit response
	// header value. Zero means the header was absent. GW2 does not send a
	// corresponding "remaining" header (confirmed empirically), so this is
	// the only rate-limit signal available to surface here.
	Limit int
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("api: %s rate limited (HTTP 429), retry after %s", e.Path, e.RetryAfter)
	}
	return fmt.Sprintf("api: %s rate limited (HTTP 429)", e.Path)
}

func (e *RateLimitError) Unwrap() error {
	return ErrRateLimited
}
