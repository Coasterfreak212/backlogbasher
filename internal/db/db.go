package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/backlogbasher/backlogbasher/internal/models"
	_ "github.com/backlogbasher/backlogbasher/internal/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}
	conn.SetMaxOpenConns(1)

	store := &Store{db: conn}
	if err := store.migrate(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}
	if err := store.SeedIfEmpty(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("seed db: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	stmts := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			photo_url TEXT NOT NULL DEFAULT '',
			cpu TEXT NOT NULL DEFAULT '',
			gpu TEXT NOT NULL DEFAULT '',
			ram TEXT NOT NULL DEFAULT '',
			os TEXT NOT NULL DEFAULT '',
			performance_tier INTEGER NOT NULL DEFAULT 3 CHECK (performance_tier BETWEEN 1 AND 5),
			ecosystems_json TEXT NOT NULL DEFAULT '[]',
			is_preset INTEGER NOT NULL DEFAULT 0,
			preset_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			primary_ecosystem TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'Backlog' CHECK (status IN ('Backlog', 'Playing', 'Completed')),
			performance_demand_tier INTEGER NOT NULL DEFAULT 3 CHECK (performance_demand_tier BETWEEN 1 AND 5),
			hltb_main REAL NOT NULL DEFAULT 0,
			hltb_extras REAL NOT NULL DEFAULT 0,
			hltb_completionist REAL NOT NULL DEFAULT 0,
			hours_logged REAL NOT NULL DEFAULT 0,
			primary_device_id INTEGER NULL REFERENCES devices(id) ON DELETE SET NULL,
			genre TEXT NOT NULL DEFAULT '',
			release_year INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("exec migration (%s): %w", q, err)
		}
	}
	return nil
}

// --- Device CRUD ---

func (s *Store) ListDevices() ([]models.Device, error) {
	rows, err := s.db.Query(`
		SELECT id, name, photo_url, cpu, gpu, ram, os, performance_tier, ecosystems_json, is_preset, preset_id, created_at
		FROM devices
		ORDER BY performance_tier DESC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		var ecoJSON string
		var isPreset int
		if err := rows.Scan(
			&d.ID, &d.Name, &d.PhotoURL, &d.CPU, &d.GPU, &d.RAM, &d.OS,
			&d.PerformanceTier, &ecoJSON, &isPreset, &d.PresetID, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		d.IsPreset = isPreset != 0
		_ = json.Unmarshal([]byte(ecoJSON), &d.Ecosystems)
		if d.Ecosystems == nil {
			d.Ecosystems = []string{}
		}
		devices = append(devices, d)
	}
	if devices == nil {
		devices = []models.Device{}
	}
	return devices, rows.Err()
}

func (s *Store) GetDevice(id int64) (models.Device, error) {
	var d models.Device
	var ecoJSON string
	var isPreset int
	err := s.db.QueryRow(`
		SELECT id, name, photo_url, cpu, gpu, ram, os, performance_tier, ecosystems_json, is_preset, preset_id, created_at
		FROM devices WHERE id = ?
	`, id).Scan(
		&d.ID, &d.Name, &d.PhotoURL, &d.CPU, &d.GPU, &d.RAM, &d.OS,
		&d.PerformanceTier, &ecoJSON, &isPreset, &d.PresetID, &d.CreatedAt,
	)
	if err != nil {
		return models.Device{}, err
	}
	d.IsPreset = isPreset != 0
	_ = json.Unmarshal([]byte(ecoJSON), &d.Ecosystems)
	if d.Ecosystems == nil {
		d.Ecosystems = []string{}
	}
	return d, nil
}

func clampTier(t int) int {
	if t < 1 {
		return 1
	}
	if t > 5 {
		return 5
	}
	return t
}

func (s *Store) CreateDevice(d models.Device) (models.Device, error) {
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		return models.Device{}, errors.New("device name is required")
	}
	d.PerformanceTier = clampTier(d.PerformanceTier)
	if len(d.Ecosystems) == 0 {
		d.Ecosystems = []string{"Steam"}
	}
	if d.PhotoURL == "" {
		d.PhotoURL = "/assets/rig-custom.svg"
	}
	ecoBytes, _ := json.Marshal(d.Ecosystems)
	isPreset := 0
	if d.IsPreset {
		isPreset = 1
	}

	res, err := s.db.Exec(`
		INSERT INTO devices (name, photo_url, cpu, gpu, ram, os, performance_tier, ecosystems_json, is_preset, preset_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.Name, d.PhotoURL, d.CPU, d.GPU, d.RAM, d.OS, d.PerformanceTier, string(ecoBytes), isPreset, d.PresetID)
	if err != nil {
		return models.Device{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Device{}, err
	}
	return s.GetDevice(id)
}

func (s *Store) UpdateDevice(d models.Device) (models.Device, error) {
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		return models.Device{}, errors.New("device name is required")
	}
	d.PerformanceTier = clampTier(d.PerformanceTier)
	if len(d.Ecosystems) == 0 {
		d.Ecosystems = []string{"Steam"}
	}
	ecoBytes, _ := json.Marshal(d.Ecosystems)
	isPreset := 0
	if d.IsPreset {
		isPreset = 1
	}

	_, err := s.db.Exec(`
		UPDATE devices
		SET name = ?, photo_url = ?, cpu = ?, gpu = ?, ram = ?, os = ?, performance_tier = ?, ecosystems_json = ?, is_preset = ?, preset_id = ?
		WHERE id = ?
	`, d.Name, d.PhotoURL, d.CPU, d.GPU, d.RAM, d.OS, d.PerformanceTier, string(ecoBytes), isPreset, d.PresetID, d.ID)
	if err != nil {
		return models.Device{}, err
	}
	return s.GetDevice(d.ID)
}

func (s *Store) DeleteDevice(id int64) error {
	_, err := s.db.Exec(`UPDATE games SET primary_device_id = NULL WHERE primary_device_id = ?`, id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM devices WHERE id = ?`, id)
	return err
}

// --- Game CRUD & Smart Cross-Play Compatibility Engine ---

func deviceSupportsEcosystem(dev models.Device, ecosystem string) bool {
	target := strings.ToLower(strings.TrimSpace(ecosystem))
	for _, eco := range dev.Ecosystems {
		if strings.ToLower(strings.TrimSpace(eco)) == target {
			return true
		}
	}
	return false
}

func computeGameCompatibility(game models.Game, devices []models.Device) []models.DeviceCompatibility {
	var compat []models.DeviceCompatibility
	for _, dev := range devices {
		if deviceSupportsEcosystem(dev, game.PrimaryEcosystem) {
			compat = append(compat, models.CalculateCompatibility(dev, game.PerformanceDemandTier))
		}
	}
	if compat == nil {
		compat = []models.DeviceCompatibility{}
	}
	return compat
}

func (s *Store) ListGames() ([]models.Game, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return nil, err
	}
	devNameByID := make(map[int64]string, len(devices))
	for _, d := range devices {
		devNameByID[d.ID] = d.Name
	}

	rows, err := s.db.Query(`
		SELECT id, title, primary_ecosystem, status, performance_demand_tier,
		       hltb_main, hltb_extras, hltb_completionist, hours_logged,
		       primary_device_id, genre, release_year, created_at
		FROM games
		ORDER BY
			CASE status
				WHEN 'Playing' THEN 1
				WHEN 'Backlog' THEN 2
				WHEN 'Completed' THEN 3
				ELSE 4
			END,
			id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var g models.Game
		var devID sql.NullInt64
		if err := rows.Scan(
			&g.ID, &g.Title, &g.PrimaryEcosystem, &g.Status, &g.PerformanceDemandTier,
			&g.HLTBMain, &g.HLTBExtras, &g.HLTBCompletionist, &g.HoursLogged,
			&devID, &g.Genre, &g.ReleaseYear, &g.CreatedAt,
		); err != nil {
			return nil, err
		}
		if devID.Valid {
			idVal := devID.Int64
			g.PrimaryDeviceID = &idVal
			g.PrimaryDeviceName = devNameByID[idVal]
		}
		g.Compatibility = computeGameCompatibility(g, devices)
		games = append(games, g)
	}
	if games == nil {
		games = []models.Game{}
	}
	return games, rows.Err()
}

func (s *Store) GetGame(id int64) (models.Game, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return models.Game{}, err
	}
	devNameByID := make(map[int64]string, len(devices))
	for _, d := range devices {
		devNameByID[d.ID] = d.Name
	}

	var g models.Game
	var devID sql.NullInt64
	err = s.db.QueryRow(`
		SELECT id, title, primary_ecosystem, status, performance_demand_tier,
		       hltb_main, hltb_extras, hltb_completionist, hours_logged,
		       primary_device_id, genre, release_year, created_at
		FROM games WHERE id = ?
	`, id).Scan(
		&g.ID, &g.Title, &g.PrimaryEcosystem, &g.Status, &g.PerformanceDemandTier,
		&g.HLTBMain, &g.HLTBExtras, &g.HLTBCompletionist, &g.HoursLogged,
		&devID, &g.Genre, &g.ReleaseYear, &g.CreatedAt,
	)
	if err != nil {
		return models.Game{}, err
	}
	if devID.Valid {
		idVal := devID.Int64
		g.PrimaryDeviceID = &idVal
		g.PrimaryDeviceName = devNameByID[idVal]
	}
	g.Compatibility = computeGameCompatibility(g, devices)
	return g, nil
}

func normalizeStatus(st string) string {
	switch strings.ToLower(strings.TrimSpace(st)) {
	case "playing":
		return "Playing"
	case "completed":
		return "Completed"
	default:
		return "Backlog"
	}
}

func (s *Store) CreateGame(g models.Game) (models.Game, error) {
	g.Title = strings.TrimSpace(g.Title)
	if g.Title == "" {
		return models.Game{}, errors.New("game title is required")
	}
	if strings.TrimSpace(g.PrimaryEcosystem) == "" {
		g.PrimaryEcosystem = "Steam"
	}
	g.Status = normalizeStatus(g.Status)
	g.PerformanceDemandTier = clampTier(g.PerformanceDemandTier)

	var devID any = nil
	if g.PrimaryDeviceID != nil && *g.PrimaryDeviceID > 0 {
		devID = *g.PrimaryDeviceID
	}

	res, err := s.db.Exec(`
		INSERT INTO games (
			title, primary_ecosystem, status, performance_demand_tier,
			hltb_main, hltb_extras, hltb_completionist, hours_logged,
			primary_device_id, genre, release_year
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, g.Title, g.PrimaryEcosystem, g.Status, g.PerformanceDemandTier,
		g.HLTBMain, g.HLTBExtras, g.HLTBCompletionist, g.HoursLogged,
		devID, g.Genre, g.ReleaseYear,
	)
	if err != nil {
		return models.Game{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Game{}, err
	}
	return s.GetGame(id)
}

func (s *Store) UpdateGame(g models.Game) (models.Game, error) {
	g.Title = strings.TrimSpace(g.Title)
	if g.Title == "" {
		return models.Game{}, errors.New("game title is required")
	}
	if strings.TrimSpace(g.PrimaryEcosystem) == "" {
		g.PrimaryEcosystem = "Steam"
	}
	g.Status = normalizeStatus(g.Status)
	g.PerformanceDemandTier = clampTier(g.PerformanceDemandTier)

	var devID any = nil
	if g.PrimaryDeviceID != nil && *g.PrimaryDeviceID > 0 {
		devID = *g.PrimaryDeviceID
	}

	_, err := s.db.Exec(`
		UPDATE games
		SET title = ?, primary_ecosystem = ?, status = ?, performance_demand_tier = ?,
		    hltb_main = ?, hltb_extras = ?, hltb_completionist = ?, hours_logged = ?,
		    primary_device_id = ?, genre = ?, release_year = ?
		WHERE id = ?
	`, g.Title, g.PrimaryEcosystem, g.Status, g.PerformanceDemandTier,
		g.HLTBMain, g.HLTBExtras, g.HLTBCompletionist, g.HoursLogged,
		devID, g.Genre, g.ReleaseYear, g.ID,
	)
	if err != nil {
		return models.Game{}, err
	}
	return s.GetGame(g.ID)
}

func (s *Store) DeleteGame(id int64) error {
	_, err := s.db.Exec(`DELETE FROM games WHERE id = ?`, id)
	return err
}

func (s *Store) StartPlayingGame(gameID int64, deviceID *int64) (models.Game, error) {
	game, err := s.GetGame(gameID)
	if err != nil {
		return models.Game{}, err
	}
	game.Status = "Playing"
	if deviceID != nil && *deviceID > 0 {
		game.PrimaryDeviceID = deviceID
	}
	return s.UpdateGame(game)
}

// --- Console Love / Utilization Summary (/api/stats) ---

func (s *Store) GetStats() (models.ERPStats, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return models.ERPStats{}, err
	}
	games, err := s.ListGames()
	if err != nil {
		return models.ERPStats{}, err
	}

	stats := models.ERPStats{
		TotalDevices: len(devices),
		TotalGames:   len(games),
	}

	devUtilMap := make(map[int64]*models.DeviceUtilization, len(devices))
	for _, d := range devices {
		devUtilMap[d.ID] = &models.DeviceUtilization{
			DeviceID:        d.ID,
			DeviceName:      d.Name,
			CPU:             d.CPU,
			GPU:             d.GPU,
			RAM:             d.RAM,
			OS:              d.OS,
			PerformanceTier: d.PerformanceTier,
			Ecosystems:      d.Ecosystems,
			PhotoURL:        d.PhotoURL,
		}
	}

	for _, g := range games {
		stats.TotalHoursLogged += g.HoursLogged
		switch g.Status {
		case "Backlog":
			stats.BacklogCount++
			rem := g.HLTBMain - g.HoursLogged
			if rem < 0 {
				rem = 0
			}
			stats.TotalBacklogHLTBHours += rem
		case "Playing":
			stats.PlayingCount++
			rem := g.HLTBMain - g.HoursLogged
			if rem > 0 {
				stats.TotalBacklogHLTBHours += rem
			}
		case "Completed":
			stats.CompletedCount++
		}

		if g.PrimaryDeviceID != nil {
			if u, ok := devUtilMap[*g.PrimaryDeviceID]; ok {
				u.TotalHoursLogged += g.HoursLogged
				switch g.Status {
				case "Backlog":
					u.AssignedBacklogCount++
				case "Playing":
					u.PlayingCount++
				case "Completed":
					u.CompletedCount++
				}
			}
		}

		// Count compatible backlog games per device
		if g.Status == "Backlog" {
			for _, comp := range g.Compatibility {
				if comp.Status == "Optimal" || comp.Status == "Playable" {
					if u, ok := devUtilMap[comp.DeviceID]; ok {
						u.CompatibleBacklogCount++
					}
				}
			}
		}
	}

	if stats.TotalGames > 0 {
		stats.CompletionRatePct = math.Round((float64(stats.CompletedCount)/float64(stats.TotalGames))*1000) / 10
	}
	stats.TotalHoursLogged = math.Round(stats.TotalHoursLogged*10) / 10
	stats.TotalBacklogHLTBHours = math.Round(stats.TotalBacklogHLTBHours*10) / 10

	var utilList []models.DeviceUtilization
	for _, d := range devices {
		u := devUtilMap[d.ID]
		u.TotalHoursLogged = math.Round(u.TotalHoursLogged*10) / 10
		if stats.TotalHoursLogged > 0 {
			u.UtilizationSharePct = math.Round((u.TotalHoursLogged/stats.TotalHoursLogged)*1000) / 10
		}
		// ActiveBacklogCount represents backlog games assigned to this rig (or compatible if none explicitly assigned)
		if u.AssignedBacklogCount > 0 {
			u.ActiveBacklogCount = u.AssignedBacklogCount
		} else {
			u.ActiveBacklogCount = u.CompatibleBacklogCount
		}

		// Determine Console Love / HealthStatus
		if u.TotalHoursLogged <= 5.0 && (u.AssignedBacklogCount > 0 || u.CompatibleBacklogCount > 0) {
			u.HealthStatus = "Neglected"
		} else if u.UtilizationSharePct >= 30.0 || u.PlayingCount >= 2 {
			u.HealthStatus = "Workhorse"
		} else if u.TotalHoursLogged < 15.0 {
			u.HealthStatus = "Underutilized"
		} else {
			u.HealthStatus = "Balanced"
		}
		utilList = append(utilList, *u)
	}

	stats.DeviceUtilization = utilList
	return stats, nil
}

// --- Next-Game Recommender Engine (/api/recommend) ---

func matchesTimeBudget(hours float64, budget string) bool {
	switch strings.ToLower(strings.TrimSpace(budget)) {
	case "<5h", "under_5", "lt5":
		return hours > 0 && hours < 5.0
	case "5-15h", "5–15h", "5_15":
		return hours >= 5.0 && hours <= 15.0
	case "15-30h", "15–30h", "15_30":
		return hours > 15.0 && hours <= 30.0
	default:
		return true // "any" or empty
	}
}

func (s *Store) RecommendGames(deviceID int64, timeBudget string, minPerf string) (models.RecommendResponse, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return models.RecommendResponse{}, err
	}
	games, err := s.ListGames()
	if err != nil {
		return models.RecommendResponse{}, err
	}
	stats, err := s.GetStats()
	if err != nil {
		return models.RecommendResponse{}, err
	}

	devByID := make(map[int64]models.Device, len(devices))
	for _, d := range devices {
		devByID[d.ID] = d
	}
	utilByID := make(map[int64]models.DeviceUtilization, len(stats.DeviceUtilization))
	for _, u := range stats.DeviceUtilization {
		utilByID[u.DeviceID] = u
	}

	includePlayable := strings.ToLower(strings.TrimSpace(minPerf)) == "playable" ||
		strings.Contains(strings.ToLower(minPerf), "playable")

	var candidates []models.RecommendationMatch

	for _, g := range games {
		if g.Status != "Backlog" {
			continue
		}

		// Calculate remaining hours (or Main Story hours if not started)
		remHours := g.HLTBMain - g.HoursLogged
		if remHours <= 0 {
			remHours = g.HLTBMain
		}
		remHours = math.Round(remHours*10) / 10

		if !matchesTimeBudget(remHours, timeBudget) && !matchesTimeBudget(g.HLTBMain, timeBudget) {
			continue
		}

		// Determine candidate devices for this game
		var candidateDevs []models.Device
		if deviceID > 0 {
			if d, ok := devByID[deviceID]; ok {
				candidateDevs = append(candidateDevs, d)
			}
		} else {
			candidateDevs = devices
		}

		// Pick the best matching device among candidateDevs
		var bestMatch *models.RecommendationMatch
		for _, dev := range candidateDevs {
			if !deviceSupportsEcosystem(dev, g.PrimaryEcosystem) {
				continue
			}
			comp := models.CalculateCompatibility(dev, g.PerformanceDemandTier)
			if comp.Status == "Underpowered" {
				continue
			}
			if comp.Status == "Playable" && !includePlayable {
				continue
			}

			score := 0
			var reasons []string

			if comp.Status == "Optimal" {
				score += 50
				if comp.Delta == 0 {
					score += 20 // Exact hardware tier sweet spot
					reasons = append(reasons, fmt.Sprintf("Exact Tier %d hardware match on %s", dev.PerformanceTier, dev.Name))
				} else {
					score += 12
					reasons = append(reasons, fmt.Sprintf("Optimal Tier %d headroom (+%d tier) on %s", dev.PerformanceTier, comp.Delta, dev.Name))
				}
			} else {
				score += 25
				reasons = append(reasons, fmt.Sprintf("Playable / Low Settings (Tier %d rig vs Tier %d demand)", dev.PerformanceTier, g.PerformanceDemandTier))
			}

			// Time budget reason
			if remHours > 0 {
				if g.HoursLogged > 0 {
					score += 15
					reasons = append(reasons, fmt.Sprintf("%.1fh remaining (%.1fh already logged of %.1fh Main Story)", remHours, g.HoursLogged, g.HLTBMain))
				} else {
					reasons = append(reasons, fmt.Sprintf("%.1fh HLTB Main Story estimate", g.HLTBMain))
				}
			}

			// Neglected hardware bonus
			if u, ok := utilByID[dev.ID]; ok {
				if u.HealthStatus == "Neglected" {
					score += 25
					reasons = append(reasons, fmt.Sprintf("Gives love to neglected hardware (%s: %.1fh total)", dev.Name, u.TotalHoursLogged))
				} else if u.HealthStatus == "Underutilized" {
					score += 10
				}
			}

			// Primary assigned device bonus
			if g.PrimaryDeviceID != nil && *g.PrimaryDeviceID == dev.ID {
				score += 10
			}

			match := models.RecommendationMatch{
				Game:                    g,
				TargetDeviceID:          dev.ID,
				TargetDeviceName:        dev.Name,
				TargetDeviceTier:        dev.PerformanceTier,
				MatchStatus:             comp.Status,
				MatchLabel:              comp.Label,
				EstimatedRemainingHours: remHours,
				Score:                   score,
				Reasons:                 reasons,
			}

			if bestMatch == nil || match.Score > bestMatch.Score {
				mCopy := match
				bestMatch = &mCopy
			}
		}

		if bestMatch != nil {
			candidates = append(candidates, *bestMatch)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].EstimatedRemainingHours < candidates[j].EstimatedRemainingHours
	})

	totalEligible := len(candidates)
	if len(candidates) > 3 {
		candidates = candidates[:3]
	}
	if candidates == nil {
		candidates = []models.RecommendationMatch{}
	}

	return models.RecommendResponse{
		DeviceID:       deviceID,
		TimeBudget:     timeBudget,
		MinPerformance: minPerf,
		TotalEligible:  totalEligible,
		Matches:        candidates,
	}, nil
}
