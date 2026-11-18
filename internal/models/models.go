package models

// HardwarePreset represents a built-in console or handheld preset in the catalog.
type HardwarePreset struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	CPU             string   `json:"cpu"`
	GPU             string   `json:"gpu"`
	RAM             string   `json:"ram"`
	OS              string   `json:"os"`
	PerformanceTier int      `json:"performance_tier"`
	Ecosystems      []string `json:"ecosystems"`
	PhotoURL        string   `json:"photo_url"`
	Notes           string   `json:"notes"`
}

// Device represents an owned hardware rig or console in the Hardware Asset Registry.
type Device struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	PhotoURL        string   `json:"photo_url"`
	CPU             string   `json:"cpu"`
	GPU             string   `json:"gpu"`
	RAM             string   `json:"ram"`
	OS              string   `json:"os"`
	PerformanceTier int      `json:"performance_tier"`
	Ecosystems      []string `json:"ecosystems"`
	IsPreset        bool     `json:"is_preset"`
	PresetID        string   `json:"preset_id,omitempty"`
	CreatedAt       string   `json:"created_at"`
}

// DeviceCompatibility represents the dynamically calculated compatibility badge
// for an owned device that shares a game's PrimaryEcosystem.
type DeviceCompatibility struct {
	DeviceID   int64  `json:"device_id"`
	DeviceName string `json:"device_name"`
	DeviceTier int    `json:"device_tier"`
	GameTier   int    `json:"game_tier"`
	Status     string `json:"status"` // "Optimal", "Playable", "Underpowered"
	Label      string `json:"label"`  // "Optimal", "Playable / Low Settings", "Underpowered"
	Delta      int    `json:"delta"`  // device_tier - game_tier
}

// Game represents a title in the Spec-Aware Cross-Play Library.
type Game struct {
	ID                    int64                 `json:"id"`
	Title                 string                `json:"title"`
	PrimaryEcosystem      string                `json:"primary_ecosystem"`
	Status                string                `json:"status"` // "Backlog", "Playing", "Completed"
	PerformanceDemandTier int                   `json:"performance_demand_tier"`
	HLTBMain              float64               `json:"hltb_main"`
	HLTBExtras            float64               `json:"hltb_extras"`
	HLTBCompletionist     float64               `json:"hltb_completionist"`
	HoursLogged           float64               `json:"hours_logged"`
	PrimaryDeviceID       *int64                `json:"primary_device_id"`
	PrimaryDeviceName     string                `json:"primary_device_name"`
	Genre                 string                `json:"genre"`
	ReleaseYear           int                   `json:"release_year"`
	CoverURL              string                `json:"cover_url,omitempty"`
	Compatibility         []DeviceCompatibility `json:"compatibility"`
	CreatedAt             string                `json:"created_at"`
}

// DeviceUtilization tracks Console Love / Utilization metrics per hardware asset.
type DeviceUtilization struct {
	DeviceID               int64    `json:"device_id"`
	DeviceName             string   `json:"device_name"`
	CPU                    string   `json:"cpu"`
	GPU                    string   `json:"gpu"`
	RAM                    string   `json:"ram"`
	OS                     string   `json:"os"`
	PerformanceTier        int      `json:"performance_tier"`
	Ecosystems             []string `json:"ecosystems"`
	PhotoURL               string   `json:"photo_url"`
	TotalHoursLogged       float64  `json:"total_hours_logged"`
	ActiveBacklogCount     int      `json:"active_backlog_count"`
	AssignedBacklogCount   int      `json:"assigned_backlog_count"`
	CompatibleBacklogCount int      `json:"compatible_backlog_count"`
	PlayingCount           int      `json:"playing_count"`
	CompletedCount         int      `json:"completed_count"`
	UtilizationSharePct    float64  `json:"utilization_share_pct"`
	HealthStatus           string   `json:"health_status"` // "Workhorse", "Balanced", "Underutilized", "Neglected"
}

// ERPStats aggregates fleet-wide utilization and library metrics.
type ERPStats struct {
	TotalDevices          int                 `json:"total_devices"`
	TotalGames            int                 `json:"total_games"`
	BacklogCount          int                 `json:"backlog_count"`
	PlayingCount          int                 `json:"playing_count"`
	CompletedCount        int                 `json:"completed_count"`
	TotalHoursLogged      float64             `json:"total_hours_logged"`
	TotalBacklogHLTBHours float64             `json:"total_backlog_hltb_hours"`
	CompletionRatePct     float64             `json:"completion_rate_pct"`
	DeviceUtilization     []DeviceUtilization `json:"device_utilization"`
}

// RecommendationMatch represents a ranked game suggestion from the Next-Game Recommender.
type RecommendationMatch struct {
	Game                    Game     `json:"game"`
	TargetDeviceID          int64    `json:"target_device_id"`
	TargetDeviceName        string   `json:"target_device_name"`
	TargetDeviceTier        int      `json:"target_device_tier"`
	MatchStatus             string   `json:"match_status"` // "Optimal" or "Playable"
	MatchLabel              string   `json:"match_label"`  // "Optimal" or "Playable / Low Settings"
	EstimatedRemainingHours float64  `json:"estimated_remaining_hours"`
	Score                   int      `json:"score"`
	Reasons                 []string `json:"reasons"`
}

// RecommendResponse wraps the top recommended games and filter metadata.
type RecommendResponse struct {
	DeviceID       int64                 `json:"device_id"`
	TimeBudget     string                `json:"time_budget"`
	MinPerformance string                `json:"min_performance"`
	TotalEligible  int                   `json:"total_eligible"`
	Matches        []RecommendationMatch `json:"matches"`
}

// HLTBEstimate represents time-to-beat data fetched from HLTB or the local fallback catalog.
type HLTBEstimate struct {
	Query                string  `json:"query"`
	MatchedTitle         string  `json:"matched_title"`
	MainStory            float64 `json:"main_story"`
	MainExtras           float64 `json:"main_extras"`
	Completionist        float64 `json:"completionist"`
	SuggestedDemandTier  int     `json:"suggested_demand_tier,omitempty"`
	SuggestedEcosystem   string  `json:"suggested_ecosystem,omitempty"`
	Genre                string  `json:"genre,omitempty"`
	ReleaseYear          int     `json:"release_year,omitempty"`
	CoverURL             string  `json:"cover_url,omitempty"`
	Source               string  `json:"source"` // "hltb_live", "local_catalog", "manual_required"
	Found                bool    `json:"found"`
	Message              string  `json:"message"`
}

// CalculateCompatibility compares a device's PerformanceTier against a game's PerformanceDemandTier.
// Optimal if device_tier >= game_tier
// Playable / Low Settings if device_tier == game_tier - 1
// Underpowered if device_tier < game_tier - 1
func CalculateCompatibility(device Device, gameTier int) DeviceCompatibility {
	delta := device.PerformanceTier - gameTier
	status := "Underpowered"
	label := "Underpowered"
	if delta >= 0 {
		status = "Optimal"
		label = "Optimal"
	} else if delta == -1 {
		status = "Playable"
		label = "Playable / Low Settings"
	}
	return DeviceCompatibility{
		DeviceID:   device.ID,
		DeviceName: device.Name,
		DeviceTier: device.PerformanceTier,
		GameTier:   gameTier,
		Status:     status,
		Label:      label,
		Delta:      delta,
	}
}
