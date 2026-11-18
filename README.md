# BacklogBasher — Self-Hosted Home Gaming ERP

**BacklogBasher** is a single-tenant, self-hosted **Home Gaming ERP** web application designed to run as a single compiled Go binary or Docker container. Built with Go (`net/http` + `embed.FS` HTML/JS/CSS frontend) and local **SQLite** persistence (`WAL` mode), it treats your living-room consoles, handheld PCs, and retro microconsoles like an enterprise hardware fleet—complete with spec-aware cross-play compatibility calculation, HowLongToBeat (HLTB) time telemetry, hardware neglect detection, and an intelligent backlog dispatch engine.

> **Zero Auth / Single-Owner Architecture**: BacklogBasher intentionally omits login screens, registration flows, and multi-user overhead. Drop the binary on your homelab, NAS, or Bazzite box and manage your gaming fleet immediately.

---

## Core Features

### 1. Hardware Asset Registry (Preset Catalog + Custom Rigs)
- **Built-In Hardware Preset Database (`/api/presets`)**: Instant one-click deployment or form pre-filling for standard consoles and handhelds:
  - **Nintendo Switch** (Tier 2 • `Switch`)
  - **Xbox Series X** (Tier 5 • `Xbox`)
  - **PlayStation 5** (Tier 5 • `PlayStation`)
  - **Steam Deck OLED** (Tier 3 • `Steam`, `Emulation`)
  - **OUYA** (Tier 1 • `Android`, `Emulation`)
  - **Razer Forge TV** (Tier 1 • `Android`, `Emulation`)
- **Custom Builds & Handhelds**: Register bespoke rigs (e.g., *GPD Win 4*, *Bazzite Living Room PC*, *ROG Ally X*) with custom fields:
  - **Name** & **Local Photo Upload** (saved directly to `$UPLOAD_DIR` and served from `/uploads/`)
  - **CPU**, **GPU**, **RAM**, **OS**
  - **Performance Tier** (`1` = Indie/Retro to `5` = Heavy AAA)
  - **Supported Ecosystems** (`Steam`, `Xbox`, `Switch`, `Emulation`, `Android`, `PlayStation`)

### 2. Spec-Aware Cross-Play & Library Management
- Track games with **Title**, **Primary Ecosystem**, **Status** (`Backlog`, `Playing`, `Completed`), **Performance Demand Tier** (`1–5`), **HLTB Estimates** (`Main Story`, `Main + Extras`, `Completionist`), **Hours Logged**, and **Primary Device Played On**.
- **Smart Cross-Play Compatibility Engine**: For any game in an ecosystem shared across multiple owned devices (such as `Steam` or `Emulation`), BacklogBasher dynamically computes and renders per-device compatibility badges by comparing the game's **Performance Demand Tier** against each device's **Performance Tier**:
  - **Optimal** (`device_tier >= game_tier`): Native target framerate and resolution headroom.
  - **Playable / Low Settings** (`device_tier == game_tier - 1`): Runs with FSR/XeSS upscaling or reduced presets.
  - **Underpowered** (`device_tier < game_tier - 1`): Below playable spec threshold on that specific rig.

### 3. HLTB Time Estimates (Live API + Local Seeded Fallback)
- Includes an **"⚡ Auto-Fetch HLTB"** button inside the Add/Edit Game modal.
- **How it works**:
  1. Queries HowLongToBeat's live search pipeline by fetching a session token from `GET https://howlongtobeat.com/api/find/init` and posting the search query to `POST https://howlongtobeat.com/api/find` with the `x-auth-token` header.
  2. **Graceful Fallback**: If HLTB blocks automated requests, changes its token format, or the instance is running offline, BacklogBasher automatically falls back to an embedded local seeded JSON catalog (`internal/hltb/catalog.json` with 50+ popular titles) and unlocks manual hour entry.

### 4. Console Love / Utilization Bar & Next-Game Recommender
- **Console Love / Utilization Summary**: Top-level visual breakdown displaying fleet-wide playtime share, total hours logged, and active backlog count per console/rig so you can spot **Neglected** hardware at a glance.
- **Next-Game Recommender**: Filter `Backlog` titles by:
  - **Device in Hand** (Specific owned rig or *Any Owned Rig*)
  - **Available Time Budget** (`<5h`, `5–15h`, `15–30h`, `Any`)
  - **Minimum Performance Match** (`Optimal Only` vs. `Include Playable`)
  Returns the **Top 3 Matches** ranked by hardware tier sweet-spot, neglected-console bonus, and remaining hours, complete with a one-click **"▶ Start Playing"** button.

---

## First-Boot Database Seeding

When `DB_PATH` is empty on first boot, BacklogBasher automatically seeds:
- **4 Hardware Rigs**:
  1. **GPD Win 4** (Tier 3 • `Steam`, `Emulation`)
  2. **Bazzite Living Room PC** (Tier 5 • `Steam`, `Emulation`)
  3. **Xbox Series X** (Tier 5 • `Xbox`)
  4. **Nintendo Switch** (Tier 2 • `Switch`)
- **10 Realistic Games** across varying demand tiers (1–5), ecosystems, statuses (`Backlog`, `Playing`, `Completed`), and HLTB lengths (*Cyberpunk 2077*, *Elden Ring*, *Titanfall 2*, *A Short Hike*, *Hades II*, *Balatro*, *Chrono Trigger*, *Halo Infinite*, *Starfield*, *Metroid Dread*).

---

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
cp .env.example .env
docker compose up -d --build
```

Open `http://localhost:8080` in your browser. SQLite data and uploaded hardware photos persist in `./data`.

### Option 2: Single Go Binary

Requirements: Go 1.22+ and `libsqlite3-dev` + `gcc` (for Cgo SQLite3 compilation).

```bash
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

---

## REST JSON API Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/devices` | List all owned hardware rigs in the registry |
| `POST` | `/api/devices` | Create a hardware rig (`application/json` or `multipart/form-data` with `photo`) |
| `PUT` / `DELETE` | `/api/devices/{id}` | Update or remove a hardware rig |
| `GET` | `/api/presets` | List the 6 built-in console/handheld presets |
| `GET` | `/api/games` | List all library games with computed `compatibility` badges per owned device |
| `POST` | `/api/games` | Add a new game to the library |
| `PUT` / `PATCH` / `DELETE` | `/api/games/{id}` | Update game fields (status, hours, HLTB, device) or delete game |
| `POST` | `/api/games/{id}/start` | One-click transition game to `Playing` on a target `device_id` |
| `GET` / `POST` | `/api/recommend` | Query Top 3 backlog recommendations (`device_id`, `time_budget`, `min_perf`) |
| `GET` | `/api/stats` | Fleet KPI summary and per-rig Console Love / Utilization metrics |
| `GET` / `POST` | `/api/hltb?title=...` | Auto-fetch HLTB time estimates with live API + local catalog fallback |

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
