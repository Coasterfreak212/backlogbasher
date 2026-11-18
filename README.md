# BacklogBasher — Self-Hosted Home Gaming ERP

**BacklogBasher** is a single-tenant, self-hosted **Home Gaming ERP** web application designed to run as a single compiled Go binary or Docker container. Built with Go (`net/http` + `embed.FS` HTML/JS/CSS frontend) and local **SQLite** persistence (`WAL` mode), it treats your living-room consoles, handheld PCs, and retro systems like an enterprise hardware fleet—complete with spec-aware cross-play compatibility calculation, real-time HowLongToBeat (HLTB) search & autosuggest, SteamGridDB v2 cover art integration, Wikimedia Commons console hardware photography, hardware neglect detection, and an intelligent backlog dispatch engine.

> **Zero Auth / Single-Owner Architecture**: BacklogBasher intentionally omits login screens, registration flows, and multi-user overhead. Drop the binary or container on your homelab, NAS, CachyOS desktop, or Bazzite HTPC and manage your gaming fleet immediately.

---

## Core Features

### 1. Dedicated Multi-Page Layout & Executive Command Overview
BacklogBasher organizes your gaming telemetry into four dedicated views with zero section duplication:
- **Command Overview (`Command Overview`)**:
  - **6 Executive KPI Cards**: *Fleet Playtime Logged*, *Remaining Backlog Debt* (+ average hours/game), *Library Completion Rate*, *Optimal 60+ FPS Hardware Coverage %*, *Lead Workhorse Rig*, and *Neglected Rigs Alert*.
  - **3 Analytics Breakdown Panels**: *Library Ecosystem Breakdown* bar chart, *Hardware Demand Tier (T1–T5) Profile* bar chart, and a **#1 Priority Backlog Dispatch Spotlight** card with cover art and a one-click **▶ Start Playing** button.
  - **3 Actionable Dispatch Lists**: *Active Playthroughs*, *Quick Wins (<5h)*, and *Neglected Hardware* with thumbnails and quick-jump navigation buttons.
  - **Hardware Fleet Utilization ("Console Love")**: Visual stacked bar and per-rig health cards showing playtime share, active backlog, and neglect warnings (`<10%` share).
- **Next-Game Recommender (`Next-Game Recommender`)**: Dedicated interactive matcher filtering `Backlog` titles by **Device in Hand**, **Time Budget** (`<5h`, `5–15h`, `15–30h`, `Any`), and **Minimum Performance** (`Optimal Only` vs. `Include Playable`), returning the Top 3 ranked dispatches.
- **Cross-Play Library (`Cross-Play Library`)**: Searchable, filterable library table with vertical game cover art, live **Library Summary Bar** (filtered count, status breakdown, and filtered HLTB Main hours), inline status toggles, `+1h` session logging, and per-device compatibility badges.
- **Hardware Asset Registry (`Hardware Asset Registry`)**: Displays your **Registered & Owned Hardware Rigs** first, followed by the searchable and collapsible **31-Console Generational Preset Catalog** and Fleet Utilization telemetry.

### 2. 31-Console Generational Preset Catalog & Custom Rigs
- **Built-In Hardware Preset Database (`/api/presets`)**: Instant one-click deployment or form pre-filling for **31 iconic consoles, handhelds, and microconsoles** spanning 1977 to modern handheld PCs—each pre-configured with authentic **Wikimedia Commons / Wikipedia** hardware photography, CPU/GPU/RAM/OS specs, Performance Tier (`1–5`), and supported ecosystems:
  - **Nintendo**: NES, SNES, Nintendo 64, GameCube, Wii, Wii U, Nintendo Switch, Game Boy / Color, Game Boy Advance, Nintendo DS, Nintendo 3DS
  - **PlayStation**: PlayStation (PS1), PlayStation 2, PlayStation 3, PlayStation 4 Pro, PlayStation 5, PSP, PS Vita
  - **Xbox**: Xbox (Original), Xbox 360, Xbox One X, Xbox Series X, Xbox Series S
  - **Sega & Retro**: Atari 2600, Sega Genesis / Mega Drive, Sega Saturn, Sega Dreamcast
  - **Handheld PCs & Microconsoles**: Steam Deck OLED, ASUS ROG Ally X, OUYA, Razer Forge TV
- **Custom Builds & Handhelds**: Register bespoke rigs (e.g., *GPD Win 4*, *Bazzite Living Room PC*, *Custom Desktop*) with custom specs and either a remote image URL or a **Local Photo Upload** (saved directly to `$UPLOAD_DIR` and served from `/uploads/`).

### 3. Real-Time HLTB Search Autosuggest + Auto-Fetch
- **Live Typeahead Autosuggest (`/api/hltb/search?q=...`)**: When adding or editing a game, typing into the **Game Title** field automatically scours HowLongToBeat (live API + embedded 50+ game catalog fallback) and displays a rich dropdown of matching titles with cover art thumbnails, release year, ecosystem, and `Main` / `+Extra` / `100%` hour badges.
- **One-Click Auto-Fill**: Clicking any suggestion in the dropdown (or pressing **⚡ Auto-Fetch HLTB + Cover**) immediately populates:
  - Canonical **Game Title** & **Cover Art URL**
  - **HLTB Main Story**, **Main + Extras**, and **Completionist** hours
  - Suggested **Performance Demand Tier (1–5)**, **Primary Ecosystem**, **Genre**, and **Release Year**
- **Graceful Offline Fallback**: If HowLongToBeat's live search endpoint is unreachable or rate-limited, BacklogBasher seamlessly falls back to its embedded `internal/hltb/catalog.json` dataset and unlocks manual hour entry.

### 4. SteamGridDB v2 Game Cover Art Integration
- **No Local Bloat**: Game cover art is sourced dynamically via **SteamGridDB v2** (`https://www.steamgriddb.com/api/v2`) and official Steam / Wikipedia CDN vertical box art URLs—never custom placeholder SVGs.
- **SteamGridDB Sync Modal**: Click **🖼️ SteamGridDB Covers** in the header (or pass `STEAMGRIDDB_API_KEY` in `.env`) to save your SteamGridDB API key in SQLite and batch-sync vertical `600x900` grid covers across your entire library with one click.

### 5. Smart Spec-Aware Cross-Play Compatibility Engine
- For any game in an ecosystem shared across multiple owned devices (such as `Steam` or `Emulation`), BacklogBasher compares the game's **Performance Demand Tier (`1–5`)** against each compatible device's **Performance Tier (`1–5`)**:
  - **Optimal 60+ FPS** (`device_tier >= game_tier`)
  - **Playable / 30–45 FPS Low-Med** (`device_tier == game_tier - 1`)
  - **Underpowered** (`device_tier < game_tier - 1`)

---

## First-Boot Database Seeding & Automatic Asset Backfill

When `DB_PATH` is empty on first boot, BacklogBasher automatically seeds:
- **4 Hardware Rigs** with Wikimedia Commons photography:
  1. **GPD Win 4** (Tier 3 • `Steam`, `Emulation`)
  2. **Bazzite Living Room PC** (Tier 5 • `Steam`, `Emulation`)
  3. **Xbox Series X** (Tier 5 • `Xbox`)
  4. **Nintendo Switch** (Tier 2 • `Switch`)
- **10 Realistic Games** across varying demand tiers (1–5), ecosystems, statuses (`Backlog`, `Playing`, `Completed`), and HLTB lengths (*Cyberpunk 2077*, *Elden Ring*, *Titanfall 2*, *A Short Hike*, *Hades II*, *Balatro*, *Chrono Trigger*, *Halo Infinite*, *Starfield*, *Metroid Dread*) with official vertical cover art.
- **Non-Destructive Startup Backfill**: On every server start, BacklogBasher automatically upgrades any blank or legacy placeholder image URLs on existing database rows to real Wikimedia Commons console photos and remote game cover art without touching your logged hours or custom uploads.

---

## Quick Start

### 1. First-Time Installation (Cloning the Repository)

If you are installing BacklogBasher for the first time, open your terminal and run:

```bash
git clone https://github.com/Coasterfreak212/backlogbasher.git
cd backlogbasher
cp .env.example .env
docker compose up -d --build
```

Open `http://localhost:8080` in your browser. Your SQLite database and uploaded hardware photos persist safely in `./data`.

---

### 2. How to Update to the Latest Version (`git pull` Guide for Beginners)

Whenever new features, UI improvements, or bug fixes are pushed to GitHub, you can update your local copy in a few seconds using `git pull`. Follow these step-by-step instructions:

1. **Open your terminal** and navigate into your `backlogbasher` folder:
   ```bash
   cd ~/backlogbasher
   ```
   *(If you cloned the folder somewhere else, replace `~/backlogbasher` with the path where you downloaded it.)*

2. **Check which branch you are on and pull the latest changes from GitHub**:
   ```bash
   git checkout main
   git pull origin main
   ```
   - **What this does**: `git checkout main` makes sure you are on the main release branch, and `git pull origin main` downloads and merges all the latest code commits from GitHub onto your machine.
   - **Troubleshooting local edits**: If `git pull` ever warns you that local files would be overwritten (for example, if a code file was modified locally), you can safely reset the code files while keeping your `./data` database untouched by running:
     ```bash
     git fetch origin main
     git reset --hard origin/main
     ```

3. **Rebuild and restart the Docker container** so the new changes take effect:
   ```bash
   docker compose up -d --build
   ```

4. **Refresh your browser**:
   - Visit `http://localhost:8080` and press **`Ctrl + Shift + R`** (or **`Cmd + Shift + R`** on macOS) to do a hard refresh so your browser loads the updated JavaScript and CSS.

---

### Option 2: Running as a Single Go Binary (Without Docker)

Requirements: Go 1.22+ and `libsqlite3-dev` + `gcc` (for Cgo SQLite3 compilation).

```bash
git pull origin main
CGO_ENABLED=1 go build -o backlogbasher .
./backlogbasher
```

---

## Environment Variables

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | HTTP port the web server binds to |
| `DB_PATH` | `./data/consoleerp.db` | File path for the SQLite database |
| `UPLOAD_DIR` | `./data/uploads` | Directory where uploaded rig photos are saved |
| `STEAMGRIDDB_API_KEY` | *(optional)* | SteamGridDB v2 API Bearer key for fetching vertical game covers (can also be configured in the web UI) |

---

## REST JSON API Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/devices` | List all owned hardware rigs in the registry |
| `POST` | `/api/devices` | Create a hardware rig (`application/json` or `multipart/form-data` with `photo`) |
| `PUT` / `DELETE` | `/api/devices/{id}` | Update or remove a hardware rig |
| `GET` | `/api/presets` | List all 31 built-in generational console & handheld presets |
| `GET` | `/api/games` | List all library games with computed `compatibility` badges per owned device |
| `POST` | `/api/games` | Add a new game to the library (auto-resolves cover art if omitted) |
| `PUT` / `PATCH` / `DELETE` | `/api/games/{id}` | Update game fields (status, hours, HLTB, cover URL, device) or delete game |
| `POST` | `/api/games/{id}/start` | One-click transition game to `Playing` on a target `device_id` |
| `GET` / `POST` | `/api/recommend` | Query Top 3 backlog recommendations (`device_id`, `time_budget`, `min_perf`) |
| `GET` | `/api/stats` | Fleet KPI summary and per-rig Console Love / Utilization metrics |
| `GET` / `POST` | `/api/hltb?title=...` | Auto-fetch single best HLTB time estimate + cover art |
| `GET` | `/api/hltb/search?q=...` | Real-time HLTB autosuggest search returning up to 8 ranked matches with hours & covers |
| `GET` | `/api/hltb/catalog` | List the embedded fallback HLTB game catalog |
| `GET` / `POST` | `/api/steamgriddb` | Check SteamGridDB status, lookup cover art by `?title=...`, or save API key & batch-sync library covers |

---

## v2 Roadmap: Automated Steam & RetroArch Playtime Sync

Planned enhancements for **BacklogBasher v2.0**:

1. **Automated Steam Web API Playtime & Library Sync**
   - Background cron worker polling `IPlayerService/GetOwnedGames/v0001` using an owner-configured `STEAM_API_KEY` and `STEAM_ID64`.
   - Automatic delta reconciliation of `playtime_forever` and `playtime_2weeks` directly into `hours_logged`, plus automatic status promotion from `Backlog` → `Playing` when new session activity is detected.
   - Steam Deck / ProtonDB verification tier ingestion (`Verified`, `Playable`, `Unsupported`) combined with local hardware tier heuristics.
2. **RetroArch & EmuDeck Activity Log Ingestion**
   - WebDAV / Syncthing / local directory watcher for RetroArch runtime logs (`content_history.lpl` and `.lrtl` core runtime logs) across handhelds (*GPD Win 4*, *Steam Deck OLED*, *OUYA*).
   - Automatic aggregation of per-core emulation hours (SNES9x, SwanStation, PCSX2, RPCS3) mapped back to the specific device node where the session took place.
3. **Bazzite / SteamOS Decky Loader Companion Plugin**
   - Lightweight telemetry webhook plugin for Bazzite HTPCs and handhelds that reports active TDP wattage, battery draw, and session duration to BacklogBasher in real time.
4. **GOG Galaxy / Xbox Live / Exophase Webhook Connectors**
   - Achievement completion triggers (`100%` unlock detection) to automatically mark titles as `Completed` and archive final completionist timestamps.
