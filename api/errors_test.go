package api

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestGet_RateLimited(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.Header().Set("X-Rate-Limit-Limit", "600")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := client.ResolveSkillName(context.Background(), 1)
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("error = %v, want errors.Is ErrRateLimited", err)
	}
	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("error = %v, want errors.As *RateLimitError to succeed", err)
	}
	if rlErr.RetryAfter.Seconds() != 30 {
		t.Errorf("RetryAfter = %v, want 30s", rlErr.RetryAfter)
	}
	if rlErr.Limit != 600 {
		t.Errorf("Limit = %d, want 600", rlErr.Limit)
	}
}

func TestGet_RateLimited_NoHeaders(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := client.ResolveSkillName(context.Background(), 1)
	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("error = %v, want errors.As *RateLimitError to succeed", err)
	}
	if rlErr.RetryAfter != 0 {
		t.Errorf("RetryAfter = %v, want 0 (unknown) when header absent", rlErr.RetryAfter)
	}
	if rlErr.Limit != 0 {
		t.Errorf("Limit = %d, want 0 (unknown) when header absent", rlErr.Limit)
	}
}

func TestGet_RateLimited_UnparseableRetryAfter(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Retry-After can also be an HTTP-date per spec; this library only
		// supports the integer-seconds form GW2 actually sends, and must
		// degrade to "unknown" rather than erroring on anything else.
		w.Header().Set("Retry-After", "Fri, 26 Jun 2026 17:00:00 GMT")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := client.ResolveSkillName(context.Background(), 1)
	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("error = %v, want errors.As *RateLimitError to succeed", err)
	}
	if rlErr.RetryAfter != 0 {
		t.Errorf("RetryAfter = %v, want 0 for an unparseable (non-integer-seconds) header", rlErr.RetryAfter)
	}
}

func TestRateLimitError_ErrorMessage(t *testing.T) {
	withRetry := &RateLimitError{Path: "/skills/1", RetryAfter: 30 * time.Second}
	if got := withRetry.Error(); got == "" {
		t.Error("Error() returned empty string")
	}
	withoutRetry := &RateLimitError{Path: "/skills/1"}
	if got := withoutRetry.Error(); got == "" {
		t.Error("Error() returned empty string")
	}
}
