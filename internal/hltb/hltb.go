package hltb

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/backlogbasher/backlogbasher/internal/models"
)

//go:embed catalog.json
var embeddedCatalogJSON []byte

type CatalogEntry struct {
	Title         string  `json:"title"`
	MainStory     float64 `json:"main_story"`
	MainExtras    float64 `json:"main_extras"`
	Completionist float64 `json:"completionist"`
	DemandTier    int     `json:"demand_tier"`
	Ecosystem     string  `json:"ecosystem"`
	Genre         string  `json:"genre"`
	ReleaseYear   int     `json:"release_year"`
}

type Service struct {
	client  *http.Client
	catalog []CatalogEntry
}

func NewService() *Service {
	var entries []CatalogEntry
	_ = json.Unmarshal(embeddedCatalogJSON, &entries)
	return &Service{
		client: &http.Client{
			Timeout: 3500 * time.Millisecond,
		},
		catalog: entries,
	}
}

func (s *Service) Catalog() []CatalogEntry {
	return s.catalog
}

type hltbInitResponse struct {
	Token string `json:"token"`
}

type hltbFindResponse struct {
	Data []struct {
		GameName     string  `json:"game_name"`
		CompMain     float64 `json:"comp_main"`
		CompPlus     float64 `json:"comp_plus"`
		Comp100      float64 `json:"comp_100"`
		ReleaseWorld int     `json:"release_world"`
	} `json:"data"`
}

// FetchEstimate queries HowLongToBeat (/api/find/init -> /api/find) and falls back
// gracefully to the local seeded JSON catalog and manual entry.
func (s *Service) FetchEstimate(title string) models.HLTBEstimate {
	clean := strings.TrimSpace(title)
	if clean == "" {
		return models.HLTBEstimate{
			Query:   title,
			Source:  "manual_required",
			Found:   false,
			Message: "Enter a game title to auto-fetch HLTB estimates.",
		}
	}

	// Step 1: Attempt live HLTB session token fetch (/api/find/init) + search POST (/api/find)
	if liveEst, err := s.queryLiveHLTB(clean); err == nil && liveEst.Found {
		// Enrich with local catalog metadata (demand tier / genre) if available
		if local, ok := s.searchLocalCatalog(clean); ok {
			liveEst.SuggestedDemandTier = local.DemandTier
			liveEst.SuggestedEcosystem = local.Ecosystem
			liveEst.Genre = local.Genre
			if liveEst.ReleaseYear == 0 {
				liveEst.ReleaseYear = local.ReleaseYear
			}
		}
		return liveEst
	}

	// Step 2: Graceful fallback to local seeded JSON catalog
	if local, ok := s.searchLocalCatalog(clean); ok {
		return models.HLTBEstimate{
			Query:               clean,
			MatchedTitle:        local.Title,
			MainStory:           local.MainStory,
			MainExtras:          local.MainExtras,
			Completionist:       local.Completionist,
			SuggestedDemandTier: local.DemandTier,
			SuggestedEcosystem:  local.Ecosystem,
			Genre:               local.Genre,
			ReleaseYear:         local.ReleaseYear,
			Source:              "local_catalog",
			Found:               true,
			Message:             fmt.Sprintf("Matched \"%s\" via local seeded HLTB catalog (live HLTB endpoint unreachable or blocked).", local.Title),
		}
	}

	// Step 3: Graceful fallback to manual hour entry
	return models.HLTBEstimate{
		Query:   clean,
		Source:  "manual_required",
		Found:   false,
		Message: fmt.Sprintf("No exact match for \"%s\" in live HLTB or local seeded catalog. Manual hour fields unlocked below.", clean),
	}
}

func (s *Service) queryLiveHLTB(title string) (models.HLTBEstimate, error) {
	initURL := fmt.Sprintf("https://howlongtobeat.com/api/find/init?t=%d", time.Now().UnixMilli())
	req, err := http.NewRequest(http.MethodGet, initURL, nil)
	if err != nil {
		return models.HLTBEstimate{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://howlongtobeat.com/")
	req.Header.Set("Origin", "https://howlongtobeat.com")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return models.HLTBEstimate{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.HLTBEstimate{}, fmt.Errorf("hltb init returned status %d", resp.StatusCode)
	}

	var initData hltbInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&initData); err != nil || initData.Token == "" {
		return models.HLTBEstimate{}, fmt.Errorf("hltb init token parse failed")
	}

	terms := strings.Fields(title)
	payload := map[string]any{
		"searchType":  "games",
		"searchTerms": terms,
		"searchPage":  1,
		"size":        5,
		"searchOptions": map[string]any{
			"games": map[string]any{
				"userId":        0,
				"platform":      "",
				"sortCategory":  "popular",
				"rangeCategory": "main",
				"rangeTime":     map[string]any{"min": nil, "max": nil},
				"gameplay":      map[string]any{"perspective": "", "flow": "", "genre": "", "difficulty": ""},
				"rangeYear":     map[string]any{"min": "", "max": ""},
				"modifier":      "",
			},
			"users":      map[string]any{"sortCategory": "postcount"},
			"lists":      map[string]any{"sortCategory": "follows"},
			"filter":     "",
			"sort":       0,
			"randomizer": 0,
		},
		"useCache": true,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return models.HLTBEstimate{}, err
	}

	findReq, err := http.NewRequest(http.MethodPost, "https://howlongtobeat.com/api/find", bytes.NewReader(bodyBytes))
	if err != nil {
		return models.HLTBEstimate{}, err
	}
	findReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")
	findReq.Header.Set("Referer", "https://howlongtobeat.com/")
	findReq.Header.Set("Origin", "https://howlongtobeat.com")
	findReq.Header.Set("Content-Type", "application/json")
	findReq.Header.Set("x-auth-token", initData.Token)

	findResp, err := s.client.Do(findReq)
	if err != nil {
		return models.HLTBEstimate{}, err
	}
	defer findResp.Body.Close()

	if findResp.StatusCode != http.StatusOK {
		return models.HLTBEstimate{}, fmt.Errorf("hltb find returned status %d", findResp.StatusCode)
	}

	var result hltbFindResponse
	if err := json.NewDecoder(findResp.Body).Decode(&result); err != nil || len(result.Data) == 0 {
		return models.HLTBEstimate{}, fmt.Errorf("hltb find returned no matches")
	}

	top := result.Data[0]
	roundHalf := func(sec float64) float64 {
		hrs := sec / 3600.0
		return math.Round(hrs*2) / 2
	}

	return models.HLTBEstimate{
		Query:         title,
		MatchedTitle:  top.GameName,
		MainStory:     roundHalf(top.CompMain),
		MainExtras:    roundHalf(top.CompPlus),
		Completionist: roundHalf(top.Comp100),
		ReleaseYear:   top.ReleaseWorld,
		Source:        "hltb_live",
		Found:         true,
		Message:       fmt.Sprintf("Fetched live from HowLongToBeat API (%s).", top.GameName),
	}, nil
}

func normalizeTitle(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	replacer := strings.NewReplacer(":", "", "-", " ", "'", "", "’", "", ".", "", "!", "")
	return strings.Join(strings.Fields(replacer.Replace(s)), " ")
}

func (s *Service) searchLocalCatalog(query string) (CatalogEntry, bool) {
	normQuery := normalizeTitle(query)
	if normQuery == "" {
		return CatalogEntry{}, false
	}

	// 1. Exact normalized match
	for _, entry := range s.catalog {
		if normalizeTitle(entry.Title) == normQuery {
			return entry, true
		}
	}

	// 2. Prefix or substring match
	bestIdx := -1
	bestScore := 0
	queryWords := strings.Fields(normQuery)

	for i, entry := range s.catalog {
		normEntry := normalizeTitle(entry.Title)
		score := 0
		if strings.HasPrefix(normEntry, normQuery) || strings.HasPrefix(normQuery, normEntry) {
			score += 50
		} else if strings.Contains(normEntry, normQuery) || strings.Contains(normQuery, normEntry) {
			score += 30
		}

		entryWords := strings.Fields(normEntry)
		matchedWords := 0
		for _, qw := range queryWords {
			if len(qw) < 2 {
				continue
			}
			for _, ew := range entryWords {
				if ew == qw {
					matchedWords++
					score += 15
				} else if len(qw) >= 3 && strings.HasPrefix(ew, qw) {
					matchedWords++
					score += 8
				}
			}
		}

		if matchedWords > 0 && score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	if bestIdx >= 0 && bestScore >= 15 {
		return s.catalog[bestIdx], true
	}

	return CatalogEntry{}, false
}
