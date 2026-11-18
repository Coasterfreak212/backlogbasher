package steamgriddb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	apiKey string
	http   *http.Client
}

func NewClient() *Client {
	return &Client{
		apiKey: strings.TrimSpace(os.Getenv("STEAMGRIDDB_API_KEY")),
		http:   &http.Client{Timeout: 4 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != ""
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

// FetchCoverURL searches SteamGridDB for the given game title and returns a grid cover URL.
// Returns "" if STEAMGRIDDB_API_KEY is not set or no grid is found.
func (c *Client) FetchCoverURL(title string) string {
	if !c.Enabled() || strings.TrimSpace(title) == "" {
		return ""
	}

	searchURL := fmt.Sprintf("https://www.steamgriddb.com/api/v2/search/autocomplete/%s", url.PathEscape(strings.TrimSpace(title)))
	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
	gReq.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	if gr.Data[0].Thumb != "" {
		return gr.Data[0].Thumb
	}
	return gr.Data[0].URL
}
