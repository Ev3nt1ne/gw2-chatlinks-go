// Package api provides optional enrichment for the chatlinks package: it
// resolves numeric IDs (and build-template "palette IDs") to human-readable
// names via the public Guild Wars 2 API. No API key is needed for any
// endpoint used here.
//
// This is kept separate from the chatlinks package so that decoding itself
// never needs network access.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.guildwars2.com/v2"

// schemaVersion is passed as the API's "v" query parameter on every request.
// The default (unversioned) schema omits several fields this package needs
// — notably skills_by_palette on /v2/professions, which only appears once
// an explicit schema version is requested. Verified empirically; "latest"
// works. If you call the GW2 API directly elsewhere, don't forget this.
const schemaVersion = "latest"

// Client wraps an *http.Client for GW2 API calls. The zero value is usable
// and uses http.DefaultClient, a 10s-per-request timeout, and the real
// public API base URL.
type Client struct {
	HTTPClient *http.Client

	// BaseURL overrides the GW2 API base URL. Defaults to baseURL
	// (https://api.guildwars2.com/v2) when empty. Exposed mainly so tests
	// can point at an httptest server instead of the real API.
	BaseURL string
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

func (c *Client) get(ctx context.Context, path string, out any) error {
	u, err := url.Parse(c.baseURL() + path)
	if err != nil {
		return fmt.Errorf("api: invalid path %q: %w", path, err)
	}
	q := u.Query()
	q.Set("v", schemaVersion)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("api: building request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("api: requesting %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api: %s returned HTTP %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("api: decoding response from %s: %w", path, err)
	}
	return nil
}

type namedEntity struct {
	Name string `json:"name"`
}

// ResolveSkillName resolves a public API skill ID to its name.
func (c *Client) ResolveSkillName(ctx context.Context, skillID int) (string, error) {
	var e namedEntity
	if err := c.get(ctx, fmt.Sprintf("/skills/%d", skillID), &e); err != nil {
		return "", err
	}
	return e.Name, nil
}

// ResolveTraitName resolves a public API trait ID to its name.
func (c *Client) ResolveTraitName(ctx context.Context, traitID int) (string, error) {
	var e namedEntity
	if err := c.get(ctx, fmt.Sprintf("/traits/%d", traitID), &e); err != nil {
		return "", err
	}
	return e.Name, nil
}

// ResolveItemName resolves a public API item ID to its name.
func (c *Client) ResolveItemName(ctx context.Context, itemID int) (string, error) {
	var e namedEntity
	if err := c.get(ctx, fmt.Sprintf("/items/%d", itemID), &e); err != nil {
		return "", err
	}
	return e.Name, nil
}

// ResolveSpecializationName resolves a public API specialization ID to its name.
func (c *Client) ResolveSpecializationName(ctx context.Context, specID int) (string, error) {
	var e namedEntity
	if err := c.get(ctx, fmt.Sprintf("/specializations/%d", specID), &e); err != nil {
		return "", err
	}
	return e.Name, nil
}

type professionResponse struct {
	SkillsByPalette [][2]int `json:"skills_by_palette"`
}

// PaletteIDToSkillID translates a build template's palette ID into a public
// API skill ID via /v2/professions/{profession}'s skills_by_palette field.
// Returns ok=false if paletteID is 0 (slot not set) or not found.
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
