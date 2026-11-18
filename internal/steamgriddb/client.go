package steamgriddb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	mu     sync.RWMutex
	apiKey string
	http   *http.Client
}

func NewClient() *Client {
	return &Client{
		apiKey: strings.TrimSpace(os.Getenv("STEAMGRIDDB_API_KEY")),
		http:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) SetAPIKey(key string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiKey = strings.TrimSpace(key)
}

func (c *Client) HasAPIKey() bool {
	if c == nil {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.apiKey != ""
}

func (c *Client) getAPIKey() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.apiKey
}

func (c *Client) Enabled() bool {
	return c != nil
}

type searchResp struct {
	Success bool `json:"success"`
	Data    []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"data"`
}

type gridsResp struct {
	Success bool `json:"success"`
	Data    []struct {
		URL   string `json:"url"`
		Thumb string `json:"thumb"`
	} `json:"data"`
}

type steamStoreSearchResp struct {
	Items []struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		TinyImage string `json:"tiny_image"`
	} `json:"items"`
}

type wikiSummaryResp struct {
	Thumbnail struct {
		Source string `json:"source"`
	} `json:"thumbnail"`
}

// FetchCoverURL queries SteamGridDB v2 first (when an API key is configured),
// and falls back to public Steam Store Search & Wikipedia cover art APIs.
func (c *Client) FetchCoverURL(title string) string {
	clean := strings.TrimSpace(title)
	if c == nil || clean == "" {
		return ""
	}

	if u := c.FetchFromSteamGridDB(clean); u != "" {
		return u
	}
	if u := c.fetchFromSteamStore(clean); u != "" {
		return u
	}
	if u := c.fetchFromWikipedia(clean); u != "" {
		return u
	}
	return ""
}

// FetchFromSteamGridDB queries SteamGridDB v2 exclusively using the configured API key.
func (c *Client) FetchFromSteamGridDB(title string) string {
	clean := strings.TrimSpace(title)
	key := c.getAPIKey()
	if key == "" || clean == "" {
		return ""
	}

	searchURL := fmt.Sprintf("https://www.steamgriddb.com/api/v2/search/autocomplete/%s", url.PathEscape(clean))
	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := c.http.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var sr searchResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil || !sr.Success || len(sr.Data) == 0 {
		return ""
	}

	gameID := sr.Data[0].ID
	gridsURL := fmt.Sprintf("https://www.steamgriddb.com/api/v2/grids/game/%d?dimensions=600x900,460x215,920x430", gameID)
	gReq, err := http.NewRequest(http.MethodGet, gridsURL, nil)
	if err != nil {
		return ""
	}
	gReq.Header.Set("Authorization", "Bearer "+key)

	gResp, err := c.http.Do(gReq)
	if err != nil {
		return ""
	}
	defer gResp.Body.Close()
	if gResp.StatusCode != http.StatusOK {
		return ""
	}

	var gr gridsResp
	if err := json.NewDecoder(gResp.Body).Decode(&gr); err != nil || !gr.Success || len(gr.Data) == 0 {
		return ""
	}

	if gr.Data[0].URL != "" {
		return gr.Data[0].URL
	}
	return gr.Data[0].Thumb
}

func (c *Client) fetchFromSteamStore(title string) string {
	u := fmt.Sprintf("https://store.steampowered.com/api/storesearch/?term=%s&l=english&cc=US", url.QueryEscape(title))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "ConsoleLoveERP/2.4")

	resp, err := c.http.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var sr steamStoreSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil || len(sr.Items) == 0 {
		return ""
	}

	appID := sr.Items[0].ID
	if appID > 0 {
		return fmt.Sprintf("https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/%d/library_600x900.jpg", appID)
	}
	return sr.Items[0].TinyImage
}

func (c *Client) fetchFromWikipedia(title string) string {
	slug := strings.ReplaceAll(title, " ", "_")
	u := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", url.PathEscape(slug))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "ConsoleLoveERP/2.4")

	resp, err := c.http.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var wr wikiSummaryResp
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		return ""
	}
	return wr.Thumbnail.Source
}
