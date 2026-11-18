package db

import (
	"github.com/backlogbasher/backlogbasher/internal/models"
)

// BuiltinPresets returns the built-in preset database of standard consoles and handhelds.
func BuiltinPresets() []models.HardwarePreset {
	return []models.HardwarePreset{
		{
			ID:              "nintendo-switch",
			Name:            "Nintendo Switch",
			Category:        "Hybrid Handheld / Console",
			CPU:             "NVIDIA Custom Tegra X1 (4x ARM Cortex-A57 @ 1.02 GHz)",
			GPU:             "NVIDIA GM20B Maxwell (256 CUDA Cores @ 307–768 MHz)",
			RAM:             "4 GB LPDDR4 (1600 MHz)",
			OS:              "Nintendo Switch Horizon OS",
			PerformanceTier: 2,
			Ecosystems:      []string{"Switch"},
			PhotoURL:        "/assets/rig-switch.svg",
			Notes:           "720p handheld / 1080p docked hybrid console tailored for Nintendo first-party & indie titles.",
		},
		{
			ID:              "xbox-series-x",
			Name:            "Xbox Series X",
			Category:        "9th-Gen Home Console",
			CPU:             "Custom AMD Zen 2 (8C/16T @ 3.8 GHz)",
			GPU:             "Custom AMD RDNA 2 (52 CUs @ 1.825 GHz, 12.15 TFLOPS)",
			RAM:             "16 GB GDDR6 (320-bit bus, 560 GB/s)",
			OS:              "Xbox System Software (NT 10 Core)",
			PerformanceTier: 5,
			Ecosystems:      []string{"Xbox"},
			PhotoURL:        "/assets/rig-xbox-series-x.svg",
			Notes:           "4K/60–120Hz flagship Microsoft console with Quick Resume & 4-generation backward compatibility.",
		},
		{
			ID:              "playstation-5",
			Name:            "PlayStation 5",
			Category:        "9th-Gen Home Console",
			CPU:             "Custom AMD Zen 2 (8C/16T @ 3.5 GHz Variable)",
			GPU:             "Custom AMD RDNA 2 (36 CUs @ 2.23 GHz, 10.28 TFLOPS)",
			RAM:             "16 GB GDDR6 (256-bit bus, 448 GB/s)",
			OS:              "PlayStation 5 System Software (Orbis OS)",
			PerformanceTier: 5,
			Ecosystems:      []string{"PlayStation"},
			PhotoURL:        "/assets/rig-ps5.svg",
			Notes:           "Ultra-high-speed NVMe I/O console with DualSense haptics and 4K Tempest 3D AudioTech.",
		},
		{
			ID:              "steam-deck-oled",
			Name:            "Steam Deck OLED",
			Category:        "PC Gaming Handheld",
			CPU:             "AMD Sephiroth Custom APU (6nm Zen 2 4C/8T @ 2.4–3.5 GHz)",
			GPU:             "8x AMD RDNA 2 CUs @ 1.6 GHz (1.6 TFLOPS)",
			RAM:             "16 GB LPDDR5 (6400 MT/s Quad-Channel)",
			OS:              "SteamOS 3.6 (Arch Linux / Proton)",
			PerformanceTier: 3,
			Ecosystems:      []string{"Steam", "Emulation"},
			PhotoURL:        "/assets/rig-steam-deck.svg",
			Notes:           "7.4-inch 90Hz HDR OLED handheld PC with 50Wh battery and deep Proton compatibility.",
		},
		{
			ID:              "ouya",
			Name:            "OUYA",
			Category:        "Android Microconsole",
			CPU:             "NVIDIA Tegra 3 T33 Quad-Core ARM Cortex-A9 @ 1.7 GHz",
			GPU:             "NVIDIA ULP GeForce (12 Cores @ 520 MHz)",
			RAM:             "1 GB DDR3-1600",
			OS:              "OUYA OS (Android 4.1 Jelly Bean / Community Lineage)",
			PerformanceTier: 1,
			Ecosystems:      []string{"Android", "Emulation"},
			PhotoURL:        "/assets/rig-ouya.svg",
			Notes:           "Iconic 2013 Kickstarter cube microconsole suited for 8/16-bit retro emulation & indie Android APKs.",
		},
		{
			ID:              "razer-forge-tv",
			Name:            "Razer Forge TV",
			Category:        "Android Microconsole",
			CPU:             "Qualcomm Snapdragon 805 Quad-Core Krait 450 @ 2.5 GHz",
			GPU:             "Qualcomm Adreno 420 (600 MHz)",
			RAM:             "2 GB Dual-Channel LPDDR3",
			OS:              "Android TV 5.0 Lollipop",
			PerformanceTier: 1,
			Ecosystems:      []string{"Android", "Emulation"},
			PhotoURL:        "/assets/rig-razer-forge.svg",
			Notes:           "Compact living-room Android TV microconsole built for lightweight Android titles and retro cores.",
		},
	}
}

// SeedIfEmpty populates the SQLite database on first boot when no devices exist.
func (s *Store) SeedIfEmpty() error {
	var deviceCount int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM devices").Scan(&deviceCount); err != nil {
		return err
	}
	if deviceCount > 0 {
		return nil
	}

	seededDevices := []models.Device{
		{
			Name:            "GPD Win 4",
			PhotoURL:        "/assets/rig-gpd-win4.svg",
			CPU:             "AMD Ryzen 7 7840U (8C/16T Zen 4 @ 5.1 GHz)",
			GPU:             "AMD Radeon 780M (12 RDNA 3 CUs @ 2.7 GHz)",
			RAM:             "32 GB LPDDR5x-7500",
			OS:              "Bazzite Deck Edition (Linux 6.11 / Proton)",
			PerformanceTier: 3,
			Ecosystems:      []string{"Steam", "Emulation"},
			IsPreset:        false,
		},
		{
			Name:            "Bazzite Living Room PC",
			PhotoURL:        "/assets/rig-bazzite-pc.svg",
			CPU:             "AMD Ryzen 7 7800X3D (8C/16T 3D V-Cache @ 5.0 GHz)",
			GPU:             "AMD Radeon RX 7900 XTX (24 GB GDDR6)",
			RAM:             "32 GB DDR5-6000 CL30",
			OS:              "Bazzite 41 HTPC (Gamescope Session)",
			PerformanceTier: 5,
			Ecosystems:      []string{"Steam", "Emulation"},
			IsPreset:        false,
		},
		{
			Name:            "Xbox Series X",
			PhotoURL:        "/assets/rig-xbox-series-x.svg",
			CPU:             "Custom AMD Zen 2 (8C/16T @ 3.8 GHz)",
			GPU:             "Custom AMD RDNA 2 (52 CUs @ 1.825 GHz, 12.15 TFLOPS)",
			RAM:             "16 GB GDDR6 (320-bit bus)",
			OS:              "Xbox System Software (NT 10 Core)",
			PerformanceTier: 5,
			Ecosystems:      []string{"Xbox"},
			IsPreset:        true,
			PresetID:        "xbox-series-x",
		},
		{
			Name:            "Nintendo Switch",
			PhotoURL:        "/assets/rig-switch.svg",
			CPU:             "NVIDIA Custom Tegra X1 (4x ARM Cortex-A57 @ 1.02 GHz)",
			GPU:             "NVIDIA GM20B Maxwell (256 CUDA Cores)",
			RAM:             "4 GB LPDDR4",
			OS:              "Nintendo Switch Horizon OS",
			PerformanceTier: 2,
			Ecosystems:      []string{"Switch"},
			IsPreset:        true,
			PresetID:        "nintendo-switch",
		},
	}

	deviceIDByName := make(map[string]int64)
	for _, dev := range seededDevices {
		created, err := s.CreateDevice(dev)
		if err != nil {
			return err
		}
		deviceIDByName[created.Name] = created.ID
	}

	idPtr := func(name string) *int64 {
		id, ok := deviceIDByName[name]
		if !ok {
			return nil
		}
		return &id
	}

	seededGames := []models.Game{
		{
			Title:                 "Cyberpunk 2077",
			PrimaryEcosystem:      "Steam",
			Status:                "Backlog",
			PerformanceDemandTier: 5,
			HLTBMain:              25.5,
			HLTBExtras:            62.0,
			HLTBCompletionist:     105.0,
			HoursLogged:           3.5,
			PrimaryDeviceID:       idPtr("Bazzite Living Room PC"),
			Genre:                 "Action RPG",
			ReleaseYear:           2020,
		},
		{
			Title:                 "Elden Ring",
			PrimaryEcosystem:      "Steam",
			Status:                "Backlog",
			PerformanceDemandTier: 4,
			HLTBMain:              59.5,
			HLTBExtras:            101.0,
			HLTBCompletionist:     134.0,
			HoursLogged:           12.0,
			PrimaryDeviceID:       idPtr("Bazzite Living Room PC"),
			Genre:                 "Action RPG",
			ReleaseYear:           2022,
		},
		{
			Title:                 "Titanfall 2",
			PrimaryEcosystem:      "Steam",
			Status:                "Backlog",
			PerformanceDemandTier: 3,
			HLTBMain:              6.0,
			HLTBExtras:            8.0,
			HLTBCompletionist:     15.5,
			HoursLogged:           0.0,
			PrimaryDeviceID:       idPtr("GPD Win 4"),
			Genre:                 "First-Person Shooter",
			ReleaseYear:           2016,
		},
		{
			Title:                 "A Short Hike",
			PrimaryEcosystem:      "Steam",
			Status:                "Backlog",
			PerformanceDemandTier: 1,
			HLTBMain:              1.5,
			HLTBExtras:            2.5,
			HLTBCompletionist:     4.5,
			HoursLogged:           0.0,
			PrimaryDeviceID:       idPtr("GPD Win 4"),
			Genre:                 "Cozy Exploration",
			ReleaseYear:           2019,
		},
		{
			Title:                 "Hades II",
			PrimaryEcosystem:      "Steam",
			Status:                "Playing",
			PerformanceDemandTier: 2,
			HLTBMain:              19.5,
			HLTBExtras:            38.0,
			HLTBCompletionist:     68.0,
			HoursLogged:           18.0,
			PrimaryDeviceID:       idPtr("GPD Win 4"),
			Genre:                 "Action Roguelike",
			ReleaseYear:           2024,
		},
		{
			Title:                 "Balatro",
			PrimaryEcosystem:      "Steam",
			Status:                "Completed",
			PerformanceDemandTier: 1,
			HLTBMain:              7.5,
			HLTBExtras:            36.0,
			HLTBCompletionist:     195.0,
			HoursLogged:           34.0,
			PrimaryDeviceID:       idPtr("GPD Win 4"),
			Genre:                 "Roguelike Deckbuilder",
			ReleaseYear:           2024,
		},
		{
			Title:                 "Chrono Trigger",
			PrimaryEcosystem:      "Emulation",
			Status:                "Backlog",
			PerformanceDemandTier: 1,
			HLTBMain:              23.0,
			HLTBExtras:            26.5,
			HLTBCompletionist:     42.0,
			HoursLogged:           0.0,
			PrimaryDeviceID:       idPtr("GPD Win 4"),
			Genre:                 "JRPG",
			ReleaseYear:           1995,
		},
		{
			Title:                 "Halo Infinite",
			PrimaryEcosystem:      "Xbox",
			Status:                "Backlog",
			PerformanceDemandTier: 4,
			HLTBMain:              11.0,
			HLTBExtras:            18.5,
			HLTBCompletionist:     28.0,
			HoursLogged:           2.0,
			PrimaryDeviceID:       idPtr("Xbox Series X"),
			Genre:                 "First-Person Shooter",
			ReleaseYear:           2021,
		},
		{
			Title:                 "Starfield",
			PrimaryEcosystem:      "Xbox",
			Status:                "Completed",
			PerformanceDemandTier: 5,
			HLTBMain:              23.0,
			HLTBExtras:            66.0,
			HLTBCompletionist:     148.0,
			HoursLogged:           48.5,
			PrimaryDeviceID:       idPtr("Xbox Series X"),
			Genre:                 "Sci-Fi RPG",
			ReleaseYear:           2023,
		},
		{
			Title:                 "Metroid Dread",
			PrimaryEcosystem:      "Switch",
			Status:                "Backlog",
			PerformanceDemandTier: 2,
			HLTBMain:              9.0,
			HLTBExtras:            11.5,
			HLTBCompletionist:     13.5,
			HoursLogged:           1.5,
			PrimaryDeviceID:       idPtr("Nintendo Switch"),
			Genre:                 "Metroidvania",
			ReleaseYear:           2021,
		},
	}

	for _, gm := range seededGames {
		if _, err := s.CreateGame(gm); err != nil {
			return err
		}
	}

	return nil
}
