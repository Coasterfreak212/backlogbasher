package db

import (
	"strings"

	"github.com/backlogbasher/backlogbasher/internal/models"
)

// DefaultGameCoverURLs returns a built-in dictionary of cover art URLs for catalog and seeded games.
func DefaultGameCoverURLs() map[string]string {
	return map[string]string{
		"Cyberpunk 2077":                            "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1091500/library_600x900.jpg",
		"Cyberpunk 2077: Phantom Liberty":           "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2138330/header.jpg",
		"Elden Ring":                                "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1245620/library_600x900.jpg",
		"Elden Ring: Shadow of the Erdtree":         "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2778580/header.jpg",
		"Titanfall 2":                               "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1237970/library_600x900.jpg",
		"A Short Hike":                              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1055540/library_600x900.jpg",
		"Hades II":                                  "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1145350/library_600x900.jpg",
		"Hades":                                     "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1145360/library_600x900.jpg",
		"Balatro":                                   "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2379780/library_600x900.jpg",
		"Chrono Trigger":                            "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/613830/library_600x900.jpg",
		"Halo Infinite":                             "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1240440/library_600x900.jpg",
		"Starfield":                                 "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1716740/library_600x900.jpg",
		"Metroid Dread":                             "https://upload.wikimedia.org/wikipedia/en/f/f7/Metroid_Dread_Banner.png",
		"The Legend of Zelda: Tears of the Kingdom": "https://upload.wikimedia.org/wikipedia/en/f/fb/The_Legend_of_Zelda_Tears_of_the_Kingdom_cover.jpg",
		"The Legend of Zelda: Breath of the Wild":   "https://upload.wikimedia.org/wikipedia/en/c/c6/The_Legend_of_Zelda_Breath_of_the_Wild.jpg",
		"Super Mario Bros. Wonder":                  "https://upload.wikimedia.org/wikipedia/en/a/a3/Mariowonder.png",
		"Super Mario Odyssey":                       "https://upload.wikimedia.org/wikipedia/en/8/8d/Super_Mario_Odyssey.jpg",
		"Xenoblade Chronicles 3":                    "https://upload.wikimedia.org/wikipedia/en/4/4a/Xenoblade_Chronicles_2.jpg",
		"Fire Emblem: Three Houses":                 "https://upload.wikimedia.org/wikipedia/en/1/16/Fire_Emblem_Three_Houses.jpg",
		"Black Myth: Wukong":                        "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2358720/library_600x900.jpg",
		"Alan Wake 2":                               "https://upload.wikimedia.org/wikipedia/en/e/ed/Alan_Wake_2_box_art.jpg",
		"Baldur's Gate 3":                           "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1086940/library_600x900.jpg",
		"Dave the Diver":                            "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1868140/library_600x900.jpg",
		"Hollow Knight":                             "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/367520/library_600x900.jpg",
		"Celeste":                                   "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/504230/library_600x900.jpg",
		"Slay the Spire":                            "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/646570/library_600x900.jpg",
		"Vampire Survivors":                         "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1794680/library_600x900.jpg",
		"Dead Cells":                                "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/588650/library_600x900.jpg",
		"Persona 5 Royal":                           "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1687950/library_600x900.jpg",
		"Metaphor: ReFantazio":                      "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2679460/library_600x900.jpg",
		"Red Dead Redemption 2":                     "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1174180/library_600x900.jpg",
		"The Witcher 3: Wild Hunt":                  "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/292030/library_600x900.jpg",
		"God of War Ragnarök":                       "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2322010/library_600x900.jpg",
		"Helldivers 2":                              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/553850/library_600x900.jpg",
		"Space Marine 2":                            "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2183900/library_600x900.jpg",
		"Forza Horizon 5":                           "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1551360/library_600x900.jpg",
		"Hi-Fi RUSH":                                "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1817230/library_600x900.jpg",
		"Indiana Jones and the Great Circle":        "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2677660/library_600x900.jpg",
		"Gears 5":                                   "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1097840/library_600x900.jpg",
		"Final Fantasy VII":                         "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/39140/library_600x900.jpg",
		"Disco Elysium":                             "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/632470/library_600x900.jpg",
		"Outer Wilds":                               "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/753640/library_600x900.jpg",
		"Subnautica":                                "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/264710/library_600x900.jpg",
		"Stardew Valley":                            "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/413150/library_600x900.jpg",
		"Armored Core VI: Fires of Rubicon":         "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1888160/library_600x900.jpg",
		"Sekiro: Shadows Die Twice":                 "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/814380/library_600x900.jpg",
		"Return of the Obra Dinn":                   "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/653530/library_600x900.jpg",
		"Stray":                                     "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1332010/library_600x900.jpg",
		"Portal 2":                                  "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/620/library_600x900.jpg",
		"Inside":                                    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/304430/library_600x900.jpg",
		"Mouthwashing":                              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2475490/library_600x900.jpg",
		"TowerFall Ascension":                       "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/251470/library_600x900.jpg",
		"Super Metroid":                             "https://upload.wikimedia.org/wikipedia/en/e/e4/Smetroidbox.jpg",
		"Castlevania: Symphony of the Night":        "https://upload.wikimedia.org/wikipedia/en/c/cf/Castlevania_SOTN_PAL.jpg",
		"Shadow of the Colossus":                    "https://upload.wikimedia.org/wikipedia/en/f/f8/Shadow_of_the_Colossus_%282005%29_cover.jpg",
		"God of War II":                             "https://upload.wikimedia.org/wikipedia/en/2/25/God_of_War_II_cover.jpg",
		"Demon's Souls (PS3)":                       "https://upload.wikimedia.org/wikipedia/en/9/91/Demon%27s_Souls_Cover.jpg",
	}
}

func normalizeCoverKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	r := strings.NewReplacer(":", "", "-", " ", "'", "", "’", "", ".", "", "!", "", "ö", "o", "(", "", ")", "")
	return strings.Join(strings.Fields(r.Replace(s)), " ")
}

// LookupDefaultCoverURL finds a cover art URL by case- and punctuation-insensitive title match.
func LookupDefaultCoverURL(title string) (string, bool) {
	norm := normalizeCoverKey(title)
	if norm == "" {
		return "", false
	}
	covers := DefaultGameCoverURLs()
	for k, v := range covers {
		if normalizeCoverKey(k) == norm {
			return v, true
		}
	}
	for k, v := range covers {
		nk := normalizeCoverKey(k)
		if strings.HasPrefix(nk, norm) || strings.HasPrefix(norm, nk) {
			return v, true
		}
	}
	return "", false
}

// LookupDefaultHardwarePhotoURL returns a Wikimedia Commons hardware photo URL for any preset or known device.
func LookupDefaultHardwarePhotoURL(name, presetID string) string {
	pid := strings.ToLower(strings.TrimSpace(presetID))
	cleanName := strings.ToLower(strings.TrimSpace(name))
	for _, p := range BuiltinPresets() {
		if (pid != "" && strings.ToLower(p.ID) == pid) || (cleanName != "" && strings.ToLower(p.Name) == cleanName) {
			return p.PhotoURL
		}
	}
	switch {
	case strings.Contains(cleanName, "gpd"):
		return "https://upload.wikimedia.org/wikipedia/commons/8/8f/GPD_Win_2.jpg"
	case strings.Contains(cleanName, "bazzite") || strings.Contains(cleanName, "pc"):
		return "https://upload.wikimedia.org/wikipedia/commons/3/31/Gaming_PC-Setup_-_Astaroth-_The_Completed_System.jpg"
	case strings.Contains(cleanName, "switch"):
		return "https://upload.wikimedia.org/wikipedia/commons/7/76/Nintendo-Switch-Console-Docked-wJoyConRB.jpg"
	case strings.Contains(cleanName, "series x"):
		return "https://upload.wikimedia.org/wikipedia/commons/b/b0/Xbox_Series_X_and_S_with_controllers_(transparent_background).png"
	case strings.Contains(cleanName, "deck"):
		return "https://upload.wikimedia.org/wikipedia/commons/5/5d/Steam_Deck_(front).png"
	case strings.Contains(cleanName, "ps5") || strings.Contains(cleanName, "playstation 5"):
		return "https://upload.wikimedia.org/wikipedia/commons/0/03/PS5DigitalEdition.png"
	default:
		return "https://upload.wikimedia.org/wikipedia/commons/3/31/Gaming_PC-Setup_-_Astaroth-_The_Completed_System.jpg"
	}
}

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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/7/76/Nintendo-Switch-Console-Docked-wJoyConRB.jpg",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/b/b0/Xbox_Series_X_and_S_with_controllers_(transparent_background).png",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/0/03/PS5DigitalEdition.png",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/5/5d/Steam_Deck_(front).png",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/a/af/OUYA-Console-set-h.jpg",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/e/e2/NVIDIA_SHIELD_TV_main_console.jpg",
			Notes:           "Compact living-room Android TV microconsole built for lightweight Android titles and retro cores.",
		},
		// 2nd Generation
		{
			ID:              "atari-2600",
			Name:            "Atari 2600",
			Category:        "2nd-Gen Home Console",
			CPU:             "MOS Technology 6507 @ 1.19 MHz",
			GPU:             "Television Interface Adaptor (TIA)",
			RAM:             "128 Bytes",
			OS:              "None (Direct Cartridge ROM)",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/0/02/Atari-2600-Wood-4Sw-Set.png",
			Notes:           "1977 pioneer of cartridge-based home gaming.",
		},
		// 3rd Generation
		{
			ID:              "nes",
			Name:            "Nintendo Entertainment System (NES)",
			Category:        "3rd-Gen Home Console",
			CPU:             "Ricoh 2A03 (6502) @ 1.79 MHz",
			GPU:             "Ricoh 2C02 PPU (256x240)",
			RAM:             "2 KB WRAM + 2 KB VRAM",
			OS:              "None (Direct Cartridge ROM)",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/b/b2/NES-Console-Set.png",
			Notes:           "8-bit icon that revitalized the home console market.",
		},
		{
			ID:              "sega-master-system",
			Name:            "Sega Master System",
			Category:        "3rd-Gen Home Console",
			CPU:             "Zilog Z80A @ 3.58 MHz",
			GPU:             "Sega VDP (TMS9918 derivative)",
			RAM:             "8 KB RAM + 16 KB VRAM",
			OS:              "Built-in BIOS",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/b/ba/Sega-Master-System-Console-FL.png",
			Notes:           "8-bit Sega home console known for vibrant color palettes.",
		},
		// 4th Generation
		{
			ID:              "snes",
			Name:            "Super Nintendo (SNES)",
			Category:        "4th-Gen Home Console",
			CPU:             "Ricoh 5A22 (65C816) @ 3.58 MHz",
			GPU:             "S-PPU1 / S-PPU2 (Mode 7)",
			RAM:             "128 KB WRAM + 64 KB VRAM",
			OS:              "None (Direct Cartridge ROM)",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/3/36/SNES-Mod1-Console-Set.png",
			Notes:           "16-bit powerhouse with Mode 7 scaling and legendary RPG/platformer library.",
		},
		{
			ID:              "sega-genesis",
			Name:            "Sega Genesis / Mega Drive",
			Category:        "4th-Gen Home Console",
			CPU:             "Motorola 68000 @ 7.67 MHz + Z80 @ 3.58 MHz",
			GPU:             "Yamaha YM7101 VDP",
			RAM:             "64 KB RAM + 64 KB VRAM",
			OS:              "None (Direct Cartridge ROM)",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/2/26/Sega-Genesis-Mod1-Set.png",
			Notes:           "16-bit arcade-at-home console with FM synthesis audio.",
		},
		{
			ID:              "game-boy",
			Name:            "Game Boy / Game Boy Color",
			Category:        "4th/5th-Gen Handheld",
			CPU:             "Sharp LR35902 (Z80-like) @ 4.19–8.38 MHz",
			GPU:             "Integrated LCD Controller (160x144)",
			RAM:             "8–32 KB WRAM + 8–16 KB VRAM",
			OS:              "Internal Boot ROM",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/7/7c/Game-Boy-FL.png",
			Notes:           "Iconic pocket handhelds with massive cartridge libraries.",
		},
		// 5th Generation
		{
			ID:              "playstation-1",
			Name:            "PlayStation (PS1)",
			Category:        "5th-Gen Home Console",
			CPU:             "MIPS R3000A @ 33.87 MHz",
			GPU:             "Sony GPU + Geometry Transformation Engine",
			RAM:             "2 MB RAM + 1 MB VRAM",
			OS:              "PlayStation BIOS",
			PerformanceTier: 1,
			Ecosystems:      []string{"PlayStation", "Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/9/95/PSX-Console-wController.png",
			Notes:           "CD-ROM 3D era pioneer.",
		},
		{
			ID:              "nintendo-64",
			Name:            "Nintendo 64",
			Category:        "5th-Gen Home Console",
			CPU:             "NEC VR4300 (64-bit MIPS) @ 93.75 MHz",
			GPU:             "SGI Reality Coprocessor (RCP) @ 62.5 MHz",
			RAM:             "4 MB RDRAM (8 MB w/ Expansion Pak)",
			OS:              "None (Direct Cartridge ROM)",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/0/02/N64-Console-Set.png",
			Notes:           "64-bit 3D console with 4-player local couch multiplayer.",
		},
		{
			ID:              "sega-saturn",
			Name:            "Sega Saturn",
			Category:        "5th-Gen Home Console",
			CPU:             "Dual Hitachi SH-2 @ 28.6 MHz",
			GPU:             "VDP1 + VDP2 Dual Video Processors",
			RAM:             "2 MB RAM + 1.5 MB VRAM",
			OS:              "Saturn BIOS",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/c/cb/Sega-Saturn-Console-Set-Mk2.png",
			Notes:           "2D arcade port powerhouse and cult-classic 32-bit console.",
		},
		// 6th Generation
		{
			ID:              "playstation-2",
			Name:            "PlayStation 2",
			Category:        "6th-Gen Home Console",
			CPU:             "Emotion Engine (MIPS R5900) @ 294.9 MHz",
			GPU:             "Graphics Synthesizer @ 147.5 MHz",
			RAM:             "32 MB RDRAM + 4 MB eDRAM",
			OS:              "PS2 OS / BIOS",
			PerformanceTier: 1,
			Ecosystems:      []string{"PlayStation", "Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/3/39/PS2-Versions.png",
			Notes:           "Best-selling home console of all time with DVD playback.",
		},
		{
			ID:              "gamecube",
			Name:            "Nintendo GameCube",
			Category:        "6th-Gen Home Console",
			CPU:             "IBM PowerPC Gekko @ 486 MHz",
			GPU:             "ATI Flipper @ 162 MHz",
			RAM:             "24 MB 1T-SRAM + 16 MB ARAM",
			OS:              "Dolphin OS",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/d/d1/GameCube-Set.jpg",
			Notes:           "Compact mini-DVD console with WaveBird & 4-player classics.",
		},
		{
			ID:              "xbox-original",
			Name:            "Xbox (Original)",
			Category:        "6th-Gen Home Console",
			CPU:             "Intel Pentium III (Coppermine) @ 733 MHz",
			GPU:             "NVIDIA NV2A (GeForce 3-class) @ 233 MHz",
			RAM:             "64 MB DDR SDRAM",
			OS:              "Xbox OS (Windows 2000 Kernel)",
			PerformanceTier: 1,
			Ecosystems:      []string{"Xbox", "Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/c/c2/Xbox-Console-Set.png",
			Notes:           "Built-in HDD, Ethernet, and Xbox Live pioneer.",
		},
		{
			ID:              "dreamcast",
			Name:            "Sega Dreamcast",
			Category:        "6th-Gen Home Console",
			CPU:             "Hitachi SH-4 @ 200 MHz",
			GPU:             "NEC PowerVR2 CLX2 @ 100 MHz",
			RAM:             "16 MB RAM + 8 MB VRAM",
			OS:              "Katana / Windows CE",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/8/81/Dreamcast-Console-Set.jpg",
			Notes:           "Ahead-of-its-time online-ready arcade home console.",
		},
		{
			ID:              "game-boy-advance",
			Name:            "Game Boy Advance",
			Category:        "6th-Gen Handheld",
			CPU:             "ARM7TDMI @ 16.78 MHz",
			GPU:             "Custom 2D PPU (240x160)",
			RAM:             "32 KB IWRAM + 256 KB EWRAM + 96 KB VRAM",
			OS:              "GBA BIOS",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/2/2d/Nintendo-Game-Boy-Advance-Black-FL.jpg",
			Notes:           "32-bit pocketable SNES-class handheld.",
		},
		// 7th Generation
		{
			ID:              "xbox-360",
			Name:            "Xbox 360",
			Category:        "7th-Gen Home Console",
			CPU:             "IBM Xenon Tri-Core PowerPC @ 3.2 GHz",
			GPU:             "ATI Xenos @ 500 MHz",
			RAM:             "512 MB GDDR3",
			OS:              "Xbox 360 OS",
			PerformanceTier: 2,
			Ecosystems:      []string{"Xbox"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/0/03/Xbox-360-Consoles-Infobox.png",
			Notes:           "HD era defining console with achievements and Xbox Live Arcade.",
		},
		{
			ID:              "playstation-3",
			Name:            "PlayStation 3",
			Category:        "7th-Gen Home Console",
			CPU:             "Cell Broadband Engine (PPE + 7 SPEs) @ 3.2 GHz",
			GPU:             "NVIDIA RSX Reality Synthesizer @ 550 MHz",
			RAM:             "256 MB XDR + 256 MB GDDR3",
			OS:              "PS3 XMB System Software",
			PerformanceTier: 2,
			Ecosystems:      []string{"PlayStation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/3/30/PS3-Consoles-Set.jpg",
			Notes:           "Blu-ray equipped 7th-gen console with deep exclusives.",
		},
		{
			ID:              "nintendo-wii",
			Name:            "Nintendo Wii",
			Category:        "7th-Gen Home Console",
			CPU:             "IBM PowerPC Broadway @ 729 MHz",
			GPU:             "ATI Hollywood @ 243 MHz",
			RAM:             "88 MB (24 MB 1T-SRAM + 64 MB GDDR3)",
			OS:              "Wii System Menu",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/f/f3/Wii-Console.png",
			Notes:           "Motion-control phenomenon with native GameCube backward compatibility.",
		},
		{
			ID:              "nintendo-ds",
			Name:            "Nintendo DS",
			Category:        "7th-Gen Handheld",
			CPU:             "ARM946E-S @ 67 MHz + ARM7TDMI @ 33 MHz",
			GPU:             "Dual-Screen 2D/3D Engine",
			RAM:             "4 MB RAM",
			OS:              "Nintendo DS Firmware",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/a/a0/Nintendo-DS-Lite-Black-Open.jpg",
			Notes:           "Dual-screen touchscreen handheld with massive library.",
		},
		{
			ID:              "psp",
			Name:            "PlayStation Portable (PSP)",
			Category:        "7th-Gen Handheld",
			CPU:             "MIPS R4000 (Allegrex) @ 333 MHz",
			GPU:             "PSP Media Engine / 166 MHz GPU",
			RAM:             "32–64 MB RAM",
			OS:              "PSP XMB System Software",
			PerformanceTier: 1,
			Ecosystems:      []string{"PlayStation", "Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/4/46/Psp-1000.jpg",
			Notes:           "Widescreen UMD handheld with strong PS1/retro emulation.",
		},
		// 8th Generation
		{
			ID:              "playstation-4",
			Name:            "PlayStation 4 / PS4 Pro",
			Category:        "8th-Gen Home Console",
			CPU:             "AMD Jaguar 8-Core @ 1.6–2.1 GHz",
			GPU:             "AMD GCN (1.84–4.20 TFLOPS)",
			RAM:             "8 GB GDDR5",
			OS:              "Orbis OS",
			PerformanceTier: 3,
			Ecosystems:      []string{"PlayStation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/8/8c/PS4-Console-wDS4.png",
			Notes:           "8th-gen PlayStation workhorse with huge library.",
		},
		{
			ID:              "xbox-one",
			Name:            "Xbox One / Xbox One X",
			Category:        "8th-Gen Home Console",
			CPU:             "AMD Jaguar 8-Core @ 1.75–2.3 GHz",
			GPU:             "AMD GCN (1.31–6.0 TFLOPS)",
			RAM:             "8–12 GB DDR3/GDDR5",
			OS:              "Xbox OS",
			PerformanceTier: 3,
			Ecosystems:      []string{"Xbox"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/0/05/Microsoft-Xbox-One-Console-Set-wKinect.jpg",
			Notes:           "8th-gen Xbox with 4K One X enhancement support.",
		},
		{
			ID:              "wii-u",
			Name:            "Nintendo Wii U",
			Category:        "8th-Gen Home Console",
			CPU:             "IBM PowerPC Espresso Tri-Core @ 1.24 GHz",
			GPU:             "AMD Radeon Latte @ 550 MHz",
			RAM:             "2 GB DDR3",
			OS:              "Wii U OS",
			PerformanceTier: 2,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/4/4a/Wii_U_Console_and_Gamepad.png",
			Notes:           "HD Nintendo console with GamePad off-TV play.",
		},
		{
			ID:              "nintendo-3ds",
			Name:            "Nintendo 3DS",
			Category:        "8th-Gen Handheld",
			CPU:             "ARM11 MPCore (Dual/Quad) @ 268–804 MHz",
			GPU:             "DMP PICA200 @ 268 MHz",
			RAM:             "128–256 MB FCRAM",
			OS:              "Nintendo 3DS OS",
			PerformanceTier: 1,
			Ecosystems:      []string{"Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/0/0a/Nintendo-3DS-AquaOpen.png",
			Notes:           "Glasses-free 3D dual-screen handheld.",
		},
		{
			ID:              "ps-vita",
			Name:            "PlayStation Vita",
			Category:        "8th-Gen Handheld",
			CPU:             "Quad-Core ARM Cortex-A9 MPCore",
			GPU:             "PowerVR SGX543MP4+",
			RAM:             "512 MB RAM + 128 MB VRAM",
			OS:              "LiveArea OS",
			PerformanceTier: 1,
			Ecosystems:      []string{"PlayStation", "Retro", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/e/e0/PlayStation-Vita-1101-FL.png",
			Notes:           "OLED/LCD handheld beloved for JRPGs, indies, and homebrew.",
		},
		// 9th Generation & Modern Handhelds
		{
			ID:              "xbox-series-s",
			Name:            "Xbox Series S",
			Category:        "9th-Gen Home Console",
			CPU:             "Custom AMD Zen 2 (8C/16T @ 3.6 GHz)",
			GPU:             "Custom AMD RDNA 2 (20 CUs @ 1.565 GHz, 4 TFLOPS)",
			RAM:             "10 GB GDDR6",
			OS:              "Xbox System Software",
			PerformanceTier: 4,
			Ecosystems:      []string{"Xbox"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/5/55/Xbox_Series_S_with_controller_(transparent_background).png",
			Notes:           "Compact all-digital 1440p 9th-gen Xbox.",
		},
		{
			ID:              "rog-ally-x",
			Name:            "ASUS ROG Ally X",
			Category:        "PC Gaming Handheld",
			CPU:             "AMD Ryzen Z1 Extreme (8C/16T Zen 4)",
			GPU:             "AMD Radeon 780M (12 RDNA 3 CUs)",
			RAM:             "24 GB LPDDR5X-7500",
			OS:              "Windows 11 / Bazzite",
			PerformanceTier: 3,
			Ecosystems:      []string{"Steam", "Xbox", "Emulation"},
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/c/ca/ROG_ALLY_-_11.jpg",
			Notes:           "High-capacity 80Wh Windows/Linux PC handheld.",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/8/8f/GPD_Win_2.jpg",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/3/31/Gaming_PC-Setup_-_Astaroth-_The_Completed_System.jpg",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/b/b0/Xbox_Series_X_and_S_with_controllers_(transparent_background).png",
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
			PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/7/76/Nintendo-Switch-Console-Docked-wJoyConRB.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1091500/library_600x900.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1245620/library_600x900.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1237970/library_600x900.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1055540/library_600x900.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1145350/library_600x900.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2379780/library_600x900.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/613830/library_600x900.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1240440/library_600x900.jpg",
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
			CoverURL:              "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1716740/library_600x900.jpg",
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
			CoverURL:              "https://upload.wikimedia.org/wikipedia/en/f/f7/Metroid_Dread_Banner.png",
		},
	}

	for _, gm := range seededGames {
		if _, err := s.CreateGame(gm); err != nil {
			return err
		}
	}

	return nil
}
