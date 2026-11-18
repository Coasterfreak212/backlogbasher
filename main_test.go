package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/backlogbasher/backlogbasher/internal/api"
	"github.com/backlogbasher/backlogbasher/internal/db"
	"github.com/backlogbasher/backlogbasher/internal/models"
	"github.com/backlogbasher/backlogbasher/web"
)

func setupTestServer(t *testing.T) (*api.Server, *db.Store, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_consoleerp.db")
	uploadDir := filepath.Join(tmpDir, "uploads")

	store, err := db.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	staticFS, err := web.StaticFS()
	if err != nil {
		t.Fatalf("StaticFS error: %v", err)
	}

	srv, err := api.NewServer(store, uploadDir, staticFS)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}
	return srv, store, uploadDir
}

func TestPresetsSeedAndCompatibility(t *testing.T) {
	srv, _, _ := setupTestServer(t)

	// 1. Verify /api/presets has the expanded generational console presets
	req := httptest.NewRequest(http.MethodGet, "/api/presets", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/presets status = %d", rec.Code)
	}
	var presets []models.HardwarePreset
	if err := json.Unmarshal(rec.Body.Bytes(), &presets); err != nil {
		t.Fatalf("unmarshal presets: %v", err)
	}
	if len(presets) < 30 {
		t.Fatalf("expected at least 30 presets, got %d", len(presets))
	}
	foundAtari := false
	for _, p := range presets {
		if p.ID == "atari-2600" {
			foundAtari = true
			break
		}
	}
	if !foundAtari {
		t.Fatalf("expected atari-2600 in presets")
	}

	// 2. Verify /api/devices has the 4 seeded rigs
	req = httptest.NewRequest(http.MethodGet, "/api/devices", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var devices []models.Device
	if err := json.Unmarshal(rec.Body.Bytes(), &devices); err != nil {
		t.Fatalf("unmarshal devices: %v", err)
	}
	if len(devices) != 4 {
		t.Fatalf("expected 4 seeded devices, got %d", len(devices))
	}

	// 3. Verify /api/games and Smart Cross-Play Compatibility Engine
	req = httptest.NewRequest(http.MethodGet, "/api/games", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var games []models.Game
	if err := json.Unmarshal(rec.Body.Bytes(), &games); err != nil {
		t.Fatalf("unmarshal games: %v", err)
	}
	if len(games) < 8 {
		t.Fatalf("expected at least 8 seeded games, got %d", len(games))
	}

	// Check Elden Ring (Tier 4, Steam) -> Bazzite (Tier 5) = Optimal, GPD Win 4 (Tier 3) = Playable / Low Settings
	for _, g := range games {
		if g.Title == "Elden Ring" {
			foundOptimal := false
			foundPlayable := false
			for _, c := range g.Compatibility {
				if c.DeviceName == "Bazzite Living Room PC" && c.Status == "Optimal" {
					foundOptimal = true
				}
				if c.DeviceName == "GPD Win 4" && c.Status == "Playable" && c.Label == "Playable / Low Settings" {
					foundPlayable = true
				}
			}
			if !foundOptimal || !foundPlayable {
				t.Fatalf("Elden Ring compatibility mismatch: %+v", g.Compatibility)
			}
		}
		// Check Cyberpunk 2077 (Tier 5, Steam) -> GPD Win 4 (Tier 3) = Underpowered
		if g.Title == "Cyberpunk 2077" {
			foundUnderpowered := false
			for _, c := range g.Compatibility {
				if c.DeviceName == "GPD Win 4" && c.Status == "Underpowered" {
					foundUnderpowered = true
				}
			}
			if !foundUnderpowered {
				t.Fatalf("Cyberpunk 2077 expected Underpowered on GPD Win 4, got: %+v", g.Compatibility)
			}
		}
	}
}

func TestRecommendStatsHLTBAndUpload(t *testing.T) {
	srv, _, _ := setupTestServer(t)

	// 1. Test /api/recommend
	req := httptest.NewRequest(http.MethodGet, "/api/recommend?time_budget=5-15h&min_perf=optimal", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/recommend status = %d", rec.Code)
	}
	var recResp models.RecommendResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &recResp); err != nil {
		t.Fatalf("unmarshal recommend: %v", err)
	}
	if len(recResp.Matches) == 0 || len(recResp.Matches) > 3 {
		t.Fatalf("expected 1..3 recommendation matches, got %d", len(recResp.Matches))
	}

	// 2. Test /api/stats
	req = httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var stats models.ERPStats
	if err := json.Unmarshal(rec.Body.Bytes(), &stats); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	if stats.TotalDevices != 4 || stats.TotalHoursLogged <= 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	// 3. Test /api/hltb fallback
	req = httptest.NewRequest(http.MethodGet, "/api/hltb?title=Hollow+Knight", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var hltbEst models.HLTBEstimate
	if err := json.Unmarshal(rec.Body.Bytes(), &hltbEst); err != nil {
		t.Fatalf("unmarshal hltb: %v", err)
	}
	if !hltbEst.Found || hltbEst.MainStory <= 0 {
		t.Fatalf("expected Hollow Knight HLTB estimate to be found, got: %+v", hltbEst)
	}

	// 4. Test multipart photo upload on POST /api/devices
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("name", "ROG Ally X")
	_ = mw.WriteField("cpu", "AMD Ryzen Z1 Extreme")
	_ = mw.WriteField("gpu", "Radeon 780M")
	_ = mw.WriteField("ram", "24GB LPDDR5X")
	_ = mw.WriteField("os", "Bazzite")
	_ = mw.WriteField("performance_tier", "3")
	_ = mw.WriteField("ecosystems_json", `["Steam","Emulation"]`)
	fw, err := mw.CreateFormFile("photo", "allyx.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = fw.Write([]byte("fake-png-bytes"))
	_ = mw.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/devices", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/devices multipart status = %d: %s", rec.Code, rec.Body.String())
	}
	var createdDev models.Device
	if err := json.Unmarshal(rec.Body.Bytes(), &createdDev); err != nil {
		t.Fatalf("unmarshal created device: %v", err)
	}
	if createdDev.Name != "ROG Ally X" || createdDev.PhotoURL == "" {
		t.Fatalf("unexpected created device: %+v", createdDev)
	}
}
