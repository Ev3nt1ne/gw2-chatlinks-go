package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://api.guildwars2.com/v2"

// schemaVersion is passed as the API's "v" query parameter on every request.
// The default (unversioned) schema omits several fields this package needs
// — notably skills_by_palette on /v2/professions, which only appears once
// an explicit schema version is requested. Verified empirically; "latest"
// works. If you call the GW2 API directly elsewhere, don't forget this.
const schemaVersion = "latest"

// defaultUserAgent identifies this library's outbound requests when the
// caller hasn't set Client.UserAgent.
const defaultUserAgent = "gw2-chatlinks-go (+https://github.com/Ev3nt1ne/gw2-chatlinks-go)"

// Client wraps an *http.Client for GW2 API calls. The zero value is usable
// and uses http.DefaultClient, a 10s-per-request timeout, and the real
// public API base URL.
type Client struct {
	HTTPClient *http.Client

	// BaseURL overrides the GW2 API base URL. Defaults to baseURL
	// (https://api.guildwars2.com/v2) when empty. Exposed mainly so tests
	// can point at an httptest server instead of the real API.
	BaseURL string

	// UserAgent overrides the User-Agent header sent on every request.
	// Defaults to a string identifying this library when empty.
	//
	// If you're embedding this client inside your own application and want
	// outbound traffic attributed to your own identity instead, set this
	// field explicitly rather than relying on an "if empty" check in a
	// custom HTTPClient.Transport: this client always sends a non-empty
	// User-Agent, so a wrapping RoundTripper's header-presence check will
	// never see it empty.
	UserAgent string
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return baseURL
}

func (c *Client) userAgent() string {
	if c.UserAgent != "" {
		return c.UserAgent
	}
	return defaultUserAgent
}

// do builds and sends a GET request to path, returning the raw response.
// The caller must close resp.Body and call cancel (in that order, via
// defer) once done reading it.
func (c *Client) do(ctx context.Context, path string) (resp *http.Response, cancel context.CancelFunc, err error) {
	u, err := url.Parse(c.baseURL() + path)
	if err != nil {
		return nil, nil, fmt.Errorf("api: invalid path %q: %w", path, err)
	}
	q := u.Query()
	q.Set("v", schemaVersion)
	u.RawQuery = q.Encode()

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("api: building request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent())

	resp, err = c.httpClient().Do(req)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("api: requesting %s: %w", path, err)
	}
	return resp, cancel, nil
}

// rateLimitErrorFromResponse builds a *RateLimitError from a 429 response.
func rateLimitErrorFromResponse(path string, resp *http.Response) error {
	return &RateLimitError{
		Path:       path,
		RetryAfter: parseRetryAfterHeader(resp.Header.Get("Retry-After")),
		Limit:      parseRateLimitHeader(resp.Header.Get("X-Rate-Limit-Limit")),
	}
}

// parseRetryAfterHeader parses GW2's Retry-After header, sent as an integer
// number of seconds. Returns 0 ("unknown") if absent or not a parseable
// non-negative integer -- never an error, since this is best-effort.
func parseRetryAfterHeader(v string) time.Duration {
	seconds, err := strconv.Atoi(v)
	if v == "" || err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// parseRateLimitHeader parses GW2's X-Rate-Limit-Limit header. Returns 0
// ("unknown") if absent or not a parseable non-negative integer.
func parseRateLimitHeader(v string) int {
	limit, err := strconv.Atoi(v)
	if v == "" || err != nil || limit < 0 {
		return 0
	}
	return limit
}

func statusError(path string, status int) error {
	return fmt.Errorf("api: %s returned HTTP %d", path, status)
}

// get issues a GET to a single-resource path (e.g. /skills/123) and decodes
// the JSON response into out. A 429 returns a *RateLimitError; any other
// non-200 returns a generic error.
func (c *Client) get(ctx context.Context, path string, out any) error {
	resp, cancel, err := c.do(ctx, path)
	if err != nil {
		return err
	}
	defer cancel()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return rateLimitErrorFromResponse(path, resp)
	}
	if resp.StatusCode != http.StatusOK {
		return statusError(path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("api: decoding response from %s: %w", path, err)
	}
	return nil
}

// getBatch issues a GET to a "?ids=..." batch endpoint and decodes the JSON
// response into out (a pointer to a slice).
//
// Verified empirically against the live API (2026-06-26): a batch with at
// least one recognized id returns 200 (all recognized) or 206 Partial
// Content (some unrecognized -- body holds only the recognized entries,
// with a Warning header naming the unrecognized ones); a batch where every
// id is unrecognized returns 404 with a non-array body
// ({"text":"all ids provided are invalid"}). All three are valid outcomes
// for a batch call, not failures -- the 404 case is treated as "zero
// results" without inspecting its body text, since getBatch is only ever
// called with ids=-shaped paths (a single-resource 404, e.g. /skills/123,
// goes through get instead and keeps meaning "not found").
func (c *Client) getBatch(ctx context.Context, path string, out any) error {
	resp, cancel, err := c.do(ctx, path)
	if err != nil {
		return err
	}
	defer cancel()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return rateLimitErrorFromResponse(path, resp)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil // no ids in this batch were recognized; out stays zero-valued
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return statusError(path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("api: decoding response from %s: %w", path, err)
	}
	return nil
}

type namedEntity struct {
	Name string `json:"name"`
}

// ResolveSkillName resolves a public API skill ID to its name. Prefer
// ResolveSkillNames when resolving more than one ID -- it batches requests
// per ArenaNet's documented recommendation instead of one request per ID.
func (c *Client) ResolveSkillName(ctx context.Context, skillID int) (string, error) {
	var e namedEntity
	if err := c.get(ctx, fmt.Sprintf("/skills/%d", skillID), &e); err != nil {
		return "", err
	}
	return e.Name, nil
}

// ResolveTraitName resolves a public API trait ID to its name. Prefer
// ResolveTraitNames when resolving more than one ID.
func (c *Client) ResolveTraitName(ctx context.Context, traitID int) (string, error) {
	var e namedEntity
	if err := c.get(ctx, fmt.Sprintf("/traits/%d", traitID), &e); err != nil {
		return "", err
	}
	return e.Name, nil
}

// ResolveItemName resolves a public API item ID to its name. Prefer
// ResolveItemNames when resolving more than one ID.
func (c *Client) ResolveItemName(ctx context.Context, itemID int) (string, error) {
	var e namedEntity
	if err := c.get(ctx, fmt.Sprintf("/items/%d", itemID), &e); err != nil {
		return "", err
	}
	return e.Name, nil
}

// ResolveSpecializationName resolves a public API specialization ID to its
// name. Prefer ResolveSpecializationNames when resolving more than one ID.
func (c *Client) ResolveSpecializationName(ctx context.Context, specID int) (string, error) {
	var e namedEntity
	if err := c.get(ctx, fmt.Sprintf("/specializations/%d", specID), &e); err != nil {
		return "", err
	}
	return e.Name, nil
}

type namedEntityWithID struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// maxBatchIDs is ArenaNet's documented per-request cap on a "?ids=..."
// batch request (see https://wiki.guildwars2.com/wiki/API:Best_practices).
// Batches larger than this are split into sequential sub-requests and
// merged, so callers never need to worry about the cap themselves.
const maxBatchIDs = 200

// chunkInts splits ids into groups of at most size each, preserving order.
// Returns nil for empty input.
func chunkInts(ids []int, size int) [][]int {
	var chunks [][]int
	for len(ids) > 0 {
		n := size
		if n > len(ids) {
			n = len(ids)
		}
		chunks = append(chunks, ids[:n])
		ids = ids[n:]
	}
	return chunks
}

func joinInts(ids []int) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.Itoa(id)
	}
	return strings.Join(strs, ",")
}

// resolveNamesBatch resolves ids against collection (e.g. "skills") via
// ArenaNet's documented ID-batching convention (?ids=1,2,3), chunking at
// maxBatchIDs and merging the results. IDs the API doesn't recognize are
// simply absent from the returned map -- not an error (see getBatch).
// Empty ids returns an empty map with no HTTP call.
func (c *Client) resolveNamesBatch(ctx context.Context, collection string, ids []int) (map[int]string, error) {
	result := make(map[int]string, len(ids))
	for _, chunk := range chunkInts(ids, maxBatchIDs) {
		var entities []namedEntityWithID
		path := fmt.Sprintf("/%s?ids=%s", collection, joinInts(chunk))
		if err := c.getBatch(ctx, path, &entities); err != nil {
			return nil, err
		}
		for _, e := range entities {
			result[e.ID] = e.Name
		}
	}
	return result, nil
}

// ResolveSkillNames resolves multiple public API skill IDs to their names in
// as few requests as possible. IDs the API doesn't recognize are absent from
// the result map -- check for missing keys rather than assuming every input
// ID round-trips to an entry.
func (c *Client) ResolveSkillNames(ctx context.Context, skillIDs []int) (map[int]string, error) {
	return c.resolveNamesBatch(ctx, "skills", skillIDs)
}

// ResolveTraitNames is the batch form of ResolveTraitName — see
// ResolveSkillNames for the missing-ID and batching behavior.
func (c *Client) ResolveTraitNames(ctx context.Context, traitIDs []int) (map[int]string, error) {
	return c.resolveNamesBatch(ctx, "traits", traitIDs)
}

// ResolveItemNames is the batch form of ResolveItemName — see
// ResolveSkillNames for the missing-ID and batching behavior.
func (c *Client) ResolveItemNames(ctx context.Context, itemIDs []int) (map[int]string, error) {
	return c.resolveNamesBatch(ctx, "items", itemIDs)
}

// ResolveSpecializationNames is the batch form of ResolveSpecializationName
// — see ResolveSkillNames for the missing-ID and batching behavior.
func (c *Client) ResolveSpecializationNames(ctx context.Context, specIDs []int) (map[int]string, error) {
	return c.resolveNamesBatch(ctx, "specializations", specIDs)
}

type professionResponse struct {
	SkillsByPalette [][2]int `json:"skills_by_palette"`
}

// PaletteIDToSkillID translates a build template's palette ID into a public
// API skill ID via /v2/professions/{profession}'s skills_by_palette field.
// Returns ok=false if paletteID is 0 (slot not set) or not found.
//
// Resolving more than one palette ID for the same profession? Prefer
// PaletteIDsToSkillIDs — this method re-fetches the whole profession
// document on every call, which is wasteful when called in a loop.
func (c *Client) PaletteIDToSkillID(ctx context.Context, profession string, paletteID int) (skillID int, ok bool, err error) {
	if paletteID == 0 {
		return 0, false, nil
	}
	var resp professionResponse
	if err := c.get(ctx, fmt.Sprintf("/professions/%s", profession), &resp); err != nil {
		return 0, false, err
	}
	for _, pair := range resp.SkillsByPalette {
		if pair[0] == paletteID {
			return pair[1], true, nil
		}
	}
	return 0, false, nil
}

// PaletteIDsToSkillIDs translates multiple build-template palette IDs for
// the same profession into public API skill IDs, fetching
// /v2/professions/{profession} exactly once regardless of how many palette
// IDs are requested — unlike calling PaletteIDToSkillID in a loop, which
// re-fetches that document on every call. Palette IDs of 0 (unset slots) or
// not found in the profession's skills_by_palette are simply absent from
// the result map. Returns an empty map with no HTTP call if every input ID
// is 0.
func (c *Client) PaletteIDsToSkillIDs(ctx context.Context, profession string, paletteIDs []int) (map[int]int, error) {
	needed := make(map[int]bool, len(paletteIDs))
	for _, id := range paletteIDs {
		if id != 0 {
			needed[id] = true
		}
	}
	result := make(map[int]int, len(needed))
	if len(needed) == 0 {
		return result, nil
	}

	var resp professionResponse
	if err := c.get(ctx, fmt.Sprintf("/professions/%s", profession), &resp); err != nil {
		return nil, err
	}
	for _, pair := range resp.SkillsByPalette {
		if needed[pair[0]] {
			result[pair[0]] = pair[1]
		}
	}
	return result, nil
}
