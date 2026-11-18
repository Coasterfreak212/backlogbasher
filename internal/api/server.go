package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/backlogbasher/backlogbasher/internal/db"
	"github.com/backlogbasher/backlogbasher/internal/hltb"
	"github.com/backlogbasher/backlogbasher/internal/models"
	"github.com/backlogbasher/backlogbasher/internal/steamgriddb"
)

type Server struct {
	store     *db.Store
	hltbSvc   *hltb.Service
	sgdb      *steamgriddb.Client
	uploadDir string
	staticFS  fs.FS
	mux       *http.ServeMux
}

func NewServer(store *db.Store, uploadDir string, staticFS fs.FS) (*Server, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir %s: %w", uploadDir, err)
	}

	s := &Server{
		store:     store,
		hltbSvc:   hltb.NewService(),
		sgdb:      steamgriddb.NewClient(),
		uploadDir: uploadDir,
		staticFS:  staticFS,
		mux:       http.NewServeMux(),
	}
	if !s.sgdb.HasAPIKey() {
		if savedKey := store.GetSetting("steamgriddb_api_key"); savedKey != "" {
			s.sgdb.SetAPIKey(savedKey)
		}
	}
	s.routes()
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// Core REST JSON API endpoints
	s.mux.HandleFunc("/api/presets", s.handlePresets)
	s.mux.HandleFunc("/api/devices", s.handleDevices)
	s.mux.HandleFunc("/api/devices/", s.handleDeviceByID)
	s.mux.HandleFunc("/api/games", s.handleGames)
	s.mux.HandleFunc("/api/games/", s.handleGameByID)
	s.mux.HandleFunc("/api/recommend", s.handleRecommend)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/hltb", s.handleHLTB)
	s.mux.HandleFunc("/api/hltb/search", s.handleHLTBSearch)
	s.mux.HandleFunc("/api/hltb/catalog", s.handleHLTBCatalog)
	s.mux.HandleFunc("/api/steamgriddb", s.handleSteamGridDB)
	s.mux.HandleFunc("/api/upload", s.handleUpload)

	// Serve uploaded hardware photos from $UPLOAD_DIR
	uploadFileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.uploadDir)))
	s.mux.Handle("/uploads/", uploadFileServer)

	// Serve embedded frontend static files
	s.mux.Handle("/", http.FileServer(http.FS(s.staticFS)))
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// --- /api/presets ---

func (s *Server) handlePresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, db.BuiltinPresets())
}

// --- /api/devices & /api/devices/{id} ---

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		devices, err := s.store.ListDevices()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, devices)

	case http.MethodPost:
		dev, err := s.parseDeviceRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		created, err := s.store.CreateDevice(dev)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleDeviceByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/devices/")
	id, err := strconv.ParseInt(strings.Trim(idStr, "/"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		dev, err := s.store.GetDevice(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeJSON(w, http.StatusOK, dev)

	case http.MethodPut, http.MethodPatch:
		existing, err := s.store.GetDevice(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		dev, err := s.parseDeviceRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		dev.ID = id
		if dev.PhotoURL == "" {
			dev.PhotoURL = existing.PhotoURL
		}
		updated, err := s.store.UpdateDevice(dev)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, updated)

	case http.MethodDelete:
		if err := s.store.DeleteDevice(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": id})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) parseDeviceRequest(r *http.Request) (models.Device, error) {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			return models.Device{}, fmt.Errorf("parse multipart form: %w", err)
		}
		tier, _ := strconv.Atoi(r.FormValue("performance_tier"))
		if tier == 0 {
			tier = 3
		}

		var ecosystems []string
		if ecoJSON := r.FormValue("ecosystems_json"); ecoJSON != "" {
			_ = json.Unmarshal([]byte(ecoJSON), &ecosystems)
		}
		if len(ecosystems) == 0 {
			if formEcos := r.MultipartForm.Value["ecosystems"]; len(formEcos) > 0 {
				for _, item := range formEcos {
					for _, part := range strings.Split(item, ",") {
						if p := strings.TrimSpace(part); p != "" {
							ecosystems = append(ecosystems, p)
						}
					}
				}
			}
		}

		photoURL := strings.TrimSpace(r.FormValue("photo_url"))
		file, header, err := r.FormFile("photo")
		if err == nil && file != nil {
			defer file.Close()
			savedURL, saveErr := s.saveUploadedFile(file, header)
			if saveErr != nil {
				return models.Device{}, saveErr
			}
			photoURL = savedURL
		}

		return models.Device{
			Name:            r.FormValue("name"),
			PhotoURL:        photoURL,
			CPU:             r.FormValue("cpu"),
			GPU:             r.FormValue("gpu"),
			RAM:             r.FormValue("ram"),
			OS:              r.FormValue("os"),
			PerformanceTier: tier,
			Ecosystems:      ecosystems,
			IsPreset:        r.FormValue("is_preset") == "true" || r.FormValue("is_preset") == "1",
			PresetID:        r.FormValue("preset_id"),
		}, nil
	}

	var dev models.Device
	if err := json.NewDecoder(r.Body).Decode(&dev); err != nil {
		return models.Device{}, fmt.Errorf("invalid JSON payload: %w", err)
	}
	return dev, nil
}

// --- /api/upload ---

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart upload")
		return
	}
	file, header, err := r.FormFile("photo")
	if err != nil {
		file, header, err = r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "missing 'photo' or 'file' form field")
			return
		}
	}
	defer file.Close()

	savedURL, err := s.saveUploadedFile(file, header)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{
		"url":      savedURL,
		"filename": filepath.Base(savedURL),
	})
}

func (s *Server) saveUploadedFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg":
	default:
		ext = ".png"
	}

	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	filename := fmt.Sprintf("rig_%d_%s%s", time.Now().Unix(), hex.EncodeToString(randBytes), ext)
	dstPath := filepath.Join(s.uploadDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create upload file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, io.LimitReader(file, 16<<20)); err != nil {
		return "", fmt.Errorf("save upload file: %w", err)
	}
	return "/uploads/" + filename, nil
}

// --- /api/games & /api/games/{id} ---

func (s *Server) handleGames(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		games, err := s.store.ListGames()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Optional query filters
		ecoFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("ecosystem")))
		statusFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
		searchFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
		devIDFilter, _ := strconv.ParseInt(r.URL.Query().Get("device_id"), 10, 64)

		if ecoFilter != "" || statusFilter != "" || searchFilter != "" || devIDFilter > 0 {
			var filtered []models.Game
			for _, g := range games {
				if ecoFilter != "" && ecoFilter != "all" && strings.ToLower(g.PrimaryEcosystem) != ecoFilter {
					continue
				}
				if statusFilter != "" && statusFilter != "all" && strings.ToLower(g.Status) != statusFilter {
					continue
				}
				if searchFilter != "" && !strings.Contains(strings.ToLower(g.Title), searchFilter) &&
					!strings.Contains(strings.ToLower(g.Genre), searchFilter) {
					continue
				}
				if devIDFilter > 0 {
					matched := g.PrimaryDeviceID != nil && *g.PrimaryDeviceID == devIDFilter
					if !matched {
						for _, c := range g.Compatibility {
							if c.DeviceID == devIDFilter {
								matched = true
								break
							}
						}
					}
					if !matched {
						continue
					}
				}
				filtered = append(filtered, g)
			}
			if filtered == nil {
				filtered = []models.Game{}
			}
			games = filtered
		}

		writeJSON(w, http.StatusOK, games)

	case http.MethodPost:
		var g models.Game
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error())
			return
		}
		if strings.TrimSpace(g.CoverURL) == "" {
			if sgdbURL := s.sgdb.FetchFromSteamGridDB(g.Title); sgdbURL != "" {
				g.CoverURL = sgdbURL
			} else if defURL, ok := db.LookupDefaultCoverURL(g.Title); ok {
				g.CoverURL = defURL
			} else if remoteURL := s.sgdb.FetchCoverURL(g.Title); remoteURL != "" {
				g.CoverURL = remoteURL
			}
		}
		created, err := s.store.CreateGame(g)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleGameByID(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/games/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "missing game id")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid game id")
		return
	}

	// Sub-action: POST /api/games/{id}/start
	if len(parts) == 2 && parts[1] == "start" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			DeviceID *int64 `json:"device_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		updated, err := s.store.StartPlayingGame(id, body.DeviceID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, updated)
		return
	}

	switch r.Method {
	case http.MethodGet:
		game, err := s.store.GetGame(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		writeJSON(w, http.StatusOK, game)

	case http.MethodPut, http.MethodPatch:
		existing, err := s.store.GetGame(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		// Merge fields onto existing game
		if v, ok := payload["title"].(string); ok {
			existing.Title = v
		}
		if v, ok := payload["primary_ecosystem"].(string); ok {
			existing.PrimaryEcosystem = v
		}
		if v, ok := payload["status"].(string); ok {
			existing.Status = v
		}
		if v, ok := toInt(payload["performance_demand_tier"]); ok {
			existing.PerformanceDemandTier = v
		}
		if v, ok := toFloat(payload["hltb_main"]); ok {
			existing.HLTBMain = v
		}
		if v, ok := toFloat(payload["hltb_extras"]); ok {
			existing.HLTBExtras = v
		}
		if v, ok := toFloat(payload["hltb_completionist"]); ok {
			existing.HLTBCompletionist = v
		}
		if v, ok := toFloat(payload["hours_logged"]); ok {
			existing.HoursLogged = v
		}
		if rawDev, exists := payload["primary_device_id"]; exists {
			if rawDev == nil {
				existing.PrimaryDeviceID = nil
			} else if devID, ok := toInt64(rawDev); ok {
				if devID <= 0 {
					existing.PrimaryDeviceID = nil
				} else {
					existing.PrimaryDeviceID = &devID
				}
			}
		}
		if v, ok := payload["genre"].(string); ok {
			existing.Genre = v
		}
		if v, ok := toInt(payload["release_year"]); ok {
			existing.ReleaseYear = v
		}
		if v, ok := payload["cover_url"].(string); ok {
			existing.CoverURL = strings.TrimSpace(v)
		}
		if existing.CoverURL == "" {
			if sgdbURL := s.sgdb.FetchFromSteamGridDB(existing.Title); sgdbURL != "" {
				existing.CoverURL = sgdbURL
			} else if defURL, ok := db.LookupDefaultCoverURL(existing.Title); ok {
				existing.CoverURL = defURL
			} else if remoteURL := s.sgdb.FetchCoverURL(existing.Title); remoteURL != "" {
				existing.CoverURL = remoteURL
			}
		}

		updated, err := s.store.UpdateGame(existing)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, updated)

	case http.MethodDelete:
		if err := s.store.DeleteGame(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": id})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// --- /api/recommend ---

func (s *Server) handleRecommend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	deviceID, _ := strconv.ParseInt(r.URL.Query().Get("device_id"), 10, 64)
	timeBudget := r.URL.Query().Get("time_budget")
	minPerf := r.URL.Query().Get("min_perf")

	if r.Method == http.MethodPost {
		var body struct {
			DeviceID       int64  `json:"device_id"`
			TimeBudget     string `json:"time_budget"`
			MinPerformance string `json:"min_perf"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if body.DeviceID > 0 {
				deviceID = body.DeviceID
			}
			if body.TimeBudget != "" {
				timeBudget = body.TimeBudget
			}
			if body.MinPerformance != "" {
				minPerf = body.MinPerformance
			}
		}
	}

	if timeBudget == "" {
		timeBudget = "any"
	}
	if minPerf == "" {
		minPerf = "optimal"
	}

	resp, err := s.store.RecommendGames(deviceID, timeBudget, minPerf)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// --- /api/stats ---

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	stats, err := s.store.GetStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// --- /api/hltb & /api/hltb/catalog ---

func (s *Server) handleHLTB(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if r.Method == http.MethodPost && title == "" {
		var body struct {
			Title string `json:"title"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		title = body.Title
	}
	est := s.hltbSvc.FetchEstimate(title)
	lookupTitle := est.MatchedTitle
	if lookupTitle == "" {
		lookupTitle = title
	}
	if sgdbURL := s.sgdb.FetchFromSteamGridDB(lookupTitle); sgdbURL != "" {
		est.CoverURL = sgdbURL
	} else if est.CoverURL == "" {
		if defURL, ok := db.LookupDefaultCoverURL(lookupTitle); ok {
			est.CoverURL = defURL
		} else {
			est.CoverURL = s.sgdb.FetchCoverURL(lookupTitle)
		}
	}
	writeJSON(w, http.StatusOK, est)
}

func (s *Server) handleHLTBSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("title"))
	}
	matches := s.hltbSvc.SearchEstimates(query, 8)
	for i := range matches {
		title := matches[i].MatchedTitle
		if title == "" {
			continue
		}
		if sgdbURL := s.sgdb.FetchFromSteamGridDB(title); sgdbURL != "" {
			matches[i].CoverURL = sgdbURL
		} else if defURL, ok := db.LookupDefaultCoverURL(title); ok {
			matches[i].CoverURL = defURL
		} else if matches[i].CoverURL == "" && i < 3 {
			matches[i].CoverURL = s.sgdb.FetchCoverURL(title)
		}
	}
	writeJSON(w, http.StatusOK, matches)
}

func (s *Server) handleHLTBCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, s.hltbSvc.Catalog())
}

// --- /api/steamgriddb ---

func (s *Server) handleSteamGridDB(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		qTitle := strings.TrimSpace(r.URL.Query().Get("title"))
		if qTitle != "" {
			cover := s.sgdb.FetchFromSteamGridDB(qTitle)
			source := "steamgriddb"
			if cover == "" {
				if defURL, ok := db.LookupDefaultCoverURL(qTitle); ok {
					cover = defURL
					source = "catalog_cdn"
				} else {
					cover = s.sgdb.FetchCoverURL(qTitle)
					source = "steam_wikipedia_fallback"
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"title":           qTitle,
				"cover_url":       cover,
				"source":          source,
				"has_api_key":     s.sgdb.HasAPIKey(),
				"found":           cover != "",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"has_api_key": s.sgdb.HasAPIKey(),
			"provider":    "SteamGridDB v2 + Steam/Wikipedia Fallback",
		})

	case http.MethodPost:
		var body struct {
			APIKey    string `json:"api_key"`
			SyncGames bool   `json:"sync_games"`
			ForceAll  bool   `json:"force_all"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		if strings.TrimSpace(body.APIKey) != "" {
			s.sgdb.SetAPIKey(body.APIKey)
			_ = s.store.SetSetting("steamgriddb_api_key", body.APIKey)
		}

		updatedCount := 0
		if body.SyncGames {
			games, err := s.store.ListGames()
			if err == nil {
				for _, g := range games {
					isSGDB := strings.Contains(g.CoverURL, "steamgriddb.com")
					if !body.ForceAll && g.CoverURL != "" && (!s.sgdb.HasAPIKey() || isSGDB) {
						continue
					}
					newCover := s.sgdb.FetchFromSteamGridDB(g.Title)
					if newCover == "" && g.CoverURL == "" {
						if defURL, ok := db.LookupDefaultCoverURL(g.Title); ok {
							newCover = defURL
						} else {
							newCover = s.sgdb.FetchCoverURL(g.Title)
						}
					}
					if newCover != "" && newCover != g.CoverURL {
						g.CoverURL = newCover
						if _, err := s.store.UpdateGame(g); err == nil {
							updatedCount++
						}
					}
				}
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"has_api_key":   s.sgdb.HasAPIKey(),
			"updated_games": updatedCount,
		})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// --- Helpers ---

func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func toInt(v any) (int, bool) {
	f, ok := toFloat(v)
	return int(f), ok
}

func toInt64(v any) (int64, bool) {
	f, ok := toFloat(v)
	return int64(f), ok
}
