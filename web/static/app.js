(() => {
  const state = {
    stats: null,
    devices: [],
    presets: [],
    games: [],
    recommend: {
      deviceId: 0,
      timeBudget: 'any',
      minPerf: 'optimal',
      data: null,
    },
    filters: {
      search: '',
      ecosystem: 'all',
      status: 'all',
      deviceId: 0,
    },
    activeTab: 'all',
  };

  const SEGMENT_COLORS = [
    '#f59e0b',
    '#10b981',
    '#38bdf8',
    '#fb7185',
    '#a3e635',
    '#60a5fa',
    '#f97316',
  ];

  const TIER_LABELS = {
    1: 'T1 • Indie/Retro',
    2: 'T2 • Hybrid/Light 3D',
    3: 'T3 • Handheld APU',
    4: 'T4 • High AAA',
    5: 'T5 • Heavy AAA',
  };

  function escapeHTML(str) {
    if (str === null || str === undefined) return '';
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function showToast(msg) {
    const container = document.getElementById('toast-container');
    if (!container) return;
    const el = document.createElement('div');
    el.className = 'toast';
    el.textContent = msg;
    container.appendChild(el);
    setTimeout(() => {
      el.remove();
    }, 3400);
  }

  async function apiJSON(url, options = {}) {
    const res = await fetch(url, options);
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new Error(data.error || `Request failed (${res.status})`);
    }
    return data;
  }

  async function loadAllData() {
    try {
      const [stats, devices, presets, games] = await Promise.all([
        apiJSON('/api/stats'),
        apiJSON('/api/devices'),
        apiJSON('/api/presets'),
        apiJSON('/api/games'),
      ]);
      state.stats = stats;
      state.devices = devices;
      state.presets = presets;
      state.games = games;

      populateDeviceDropdowns();
      renderKPIStrip();
      renderCommandBrief();
      renderUtilizationSection();
      renderLibrarySection();
      renderPresetsAndHardware();
      await loadRecommendations();
    } catch (err) {
      showToast(`Sync error: ${err.message}`);
    }
  }

  async function loadCatalogDatalist() {
    try {
      const catalog = await apiJSON('/api/hltb/catalog');
      const dl = document.getElementById('hltb-catalog-datalist');
      if (!dl) return;
      dl.innerHTML = catalog
        .map((c) => `<option value="${escapeHTML(c.title)}">${escapeHTML(c.ecosystem)} • ${c.main_story}h Main</option>`)
        .join('');
    } catch (_) {
      // Non-fatal
    }
  }

  function populateDeviceDropdowns() {
    const recSelect = document.getElementById('rec-device-select');
    const libDevSelect = document.getElementById('lib-device-filter');
    const gameDevSelect = document.getElementById('game-device-select');

    if (recSelect) {
      const prev = recSelect.value || '0';
      recSelect.innerHTML =
        `<option value="0">Any Owned Rig (Auto-Match Best Hardware)</option>` +
        state.devices
          .map(
            (d) =>
              `<option value="${d.id}">${escapeHTML(d.name)} (Tier ${d.performance_tier} • ${escapeHTML(
                d.ecosystems.join('/')
              )})</option>`
          )
          .join('');
      recSelect.value = state.devices.some((d) => String(d.id) === String(prev)) ? prev : '0';
      state.recommend.deviceId = Number(recSelect.value) || 0;
    }

    if (libDevSelect) {
      const prev = libDevSelect.value || '0';
      libDevSelect.innerHTML =
        `<option value="0">All Hardware Rigs</option>` +
        state.devices.map((d) => `<option value="${d.id}">${escapeHTML(d.name)} (T${d.performance_tier})</option>`).join('');
      libDevSelect.value = state.devices.some((d) => String(d.id) === String(prev)) ? prev : '0';
    }

    if (gameDevSelect) {
      const prev = gameDevSelect.value || '';
      gameDevSelect.innerHTML =
        `<option value="">— Unassigned —</option>` +
        state.devices
          .map((d) => `<option value="${d.id}">${escapeHTML(d.name)} (Tier ${d.performance_tier})</option>`)
          .join('');
      gameDevSelect.value = prev;
    }
  }

  function renderKPIStrip() {
    if (!state.stats) return;
    const s = state.stats;
    document.getElementById('kpi-total-hours').textContent = `${s.total_hours_logged.toFixed(1)}h`;
    document.getElementById('kpi-playing-count').textContent = String(s.playing_count);
    document.getElementById('kpi-backlog-hours').textContent = `${s.total_backlog_hltb_hours.toFixed(1)}h`;
    document.getElementById('kpi-completion-rate').textContent = `${s.completion_rate_pct.toFixed(0)}%`;
    document.getElementById('nav-game-count').textContent = String(s.total_games);
    document.getElementById('nav-device-count').textContent = String(s.total_devices);
  }

  // --- SECTION 1: Console Love / Utilization Summary ---
  function renderCommandBrief() {
    const grid = document.getElementById('command-brief-grid');
    if (!grid) return;

    const playing = state.games.filter((g) => g.status === 'Playing');
    const quickWins = state.games
      .filter((g) => g.status === 'Backlog' && g.hltb_main > 0 && g.hltb_main <= 5)
      .sort((a, b) => a.hltb_main - b.hltb_main)
      .slice(0, 4);

    const utils = (state.stats && state.stats.device_utilization) || [];
    const neglected = utils.filter((u) => u.health_status === 'Neglected');

    const playingHTML = playing.length
      ? playing
          .slice(0, 4)
          .map((g) => {
            const pct = g.hltb_main > 0 ? Math.min(100, Math.round((g.hours_logged / g.hltb_main) * 100)) : 0;
            const eco = g.primary_ecosystem || g.ecosystem || 'Steam';
            const thumb = g.cover_url
              ? `<img src="${escapeHTML(g.cover_url)}" alt="${escapeHTML(g.title)}" style="width:28px;height:40px;object-fit:cover;border-radius:4px;flex-shrink:0;background:#0d121a;" loading="lazy" />`
              : '';
            return `<div style="display:flex;justify-content:space-between;align-items:center;padding:6px 0;border-bottom:1px solid rgba(255,255,255,0.06);gap:10px;">
              <div style="display:flex;align-items:center;gap:10px;min-width:0;">
                ${thumb}
                <div style="min-width:0;">
                  <div style="font-weight:600;color:#e2e8f0;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${escapeHTML(g.title)}</div>
                  <div style="font-size:0.78rem;color:#94a3b8;">${g.hours_logged.toFixed(1)}h / ${g.hltb_main > 0 ? g.hltb_main.toFixed(1) + 'h' : '?'} (${pct}%)</div>
                </div>
              </div>
              <span class="eco-badge eco-${escapeHTML(eco)}">${escapeHTML(eco)}</span>
            </div>`;
          })
          .join('')
      : `<div style="color:#64748b;font-size:0.85rem;">No active playthroughs. Pick a game from Quick Wins or Recommender!</div>`;

    const quickWinsHTML = quickWins.length
      ? quickWins
          .map((g) => {
            const thumb = g.cover_url
              ? `<img src="${escapeHTML(g.cover_url)}" alt="${escapeHTML(g.title)}" style="width:26px;height:36px;object-fit:cover;border-radius:4px;flex-shrink:0;background:#0d121a;" loading="lazy" />`
              : '';
            return `<div style="display:flex;justify-content:space-between;align-items:center;padding:6px 0;border-bottom:1px solid rgba(255,255,255,0.06);gap:10px;">
              <div style="display:flex;align-items:center;gap:8px;min-width:0;">
                ${thumb}
                <span style="font-weight:500;color:#e2e8f0;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${escapeHTML(g.title)}</span>
              </div>
              <span style="font-size:0.8rem;color:#38bdf8;font-family:monospace;flex-shrink:0;">~${g.hltb_main.toFixed(1)}h</span>
            </div>`;
          })
          .join('')
      : `<div style="color:#64748b;font-size:0.85rem;">No sub-5h backlog games found.</div>`;

    const neglectedHTML = neglected.length
      ? neglected
          .slice(0, 4)
          .map((u) => {
            const thumb = u.photo_url
              ? `<img src="${escapeHTML(u.photo_url)}" alt="${escapeHTML(u.device_name)}" style="width:36px;height:26px;object-fit:contain;border-radius:4px;background:#0d121a;padding:2px;flex-shrink:0;" loading="lazy" />`
              : '';
            return `<div style="display:flex;justify-content:space-between;align-items:center;padding:6px 0;border-bottom:1px solid rgba(255,255,255,0.06);gap:10px;">
              <div style="display:flex;align-items:center;gap:8px;min-width:0;">
                ${thumb}
                <span style="font-weight:500;color:#fbbf24;">${escapeHTML(u.device_name)}</span>
              </div>
              <span style="font-size:0.78rem;color:#94a3b8;">${(u.total_hours_logged || 0).toFixed(1)}h logged</span>
            </div>`;
          })
          .join('')
      : `<div style="color:#34d399;font-size:0.85rem;">All rigs are actively getting love!</div>`;

    grid.innerHTML = `
      <div class="util-card">
        <div class="util-card-header"><span class="util-rig-name">🎮 Active Playthroughs (${playing.length})</span></div>
        <div style="margin-top:8px;">${playingHTML}</div>
      </div>
      <div class="util-card">
        <div class="util-card-header"><span class="util-rig-name">⚡ Quick Wins (&lt;5h Backlog)</span></div>
        <div style="margin-top:8px;">${quickWinsHTML}</div>
      </div>
      <div class="util-card">
        <div class="util-card-header"><span class="util-rig-name">⚠ Neglected Hardware (${neglected.length})</span></div>
        <div style="margin-top:8px;">${neglectedHTML}</div>
      </div>
    `;
  }

  function renderUtilizationSection() {
    const s = state.stats;
    if (!s) return;
    const utils = s.device_utilization || [];
    document.getElementById('fleet-bar-summary-text').textContent = `${s.total_hours_logged.toFixed(
      1
    )}h logged across ${utils.length} hardware rigs • ${s.backlog_count} games in backlog`;

    const barEl = document.getElementById('fleet-stacked-bar');
    const legendEl = document.getElementById('fleet-legend');

    if (utils.length === 0) {
      barEl.innerHTML = '';
      legendEl.innerHTML = '<span>No hardware rigs registered yet.</span>';
      document.getElementById('utilization-grid').innerHTML =
        '<div class="empty-state">Register your first console or handheld rig to track Console Love utilization.</div>';
      return;
    }

    barEl.innerHTML = utils
      .map((u, idx) => {
        const color = SEGMENT_COLORS[idx % SEGMENT_COLORS.length];
        const widthPct = Math.max(u.utilization_share_pct, 3);
        return `<div class="fleet-segment" style="width:${widthPct}%; background:${color};" title="${escapeHTML(
          u.device_name
        )}: ${u.total_hours_logged.toFixed(1)}h (${u.utilization_share_pct}%)"></div>`;
      })
      .join('');

    legendEl.innerHTML = utils
      .map((u, idx) => {
        const color = SEGMENT_COLORS[idx % SEGMENT_COLORS.length];
        return `<div class="fleet-legend-item">
          <span class="legend-dot" style="background:${color};"></span>
          <span>${escapeHTML(u.device_name)}: <strong>${u.total_hours_logged.toFixed(1)}h</strong> (${u.utilization_share_pct}%)</span>
        </div>`;
      })
      .join('');

    const gridEl = document.getElementById('utilization-grid');
    gridEl.innerHTML = utils
      .map((u, idx) => {
        const color = SEGMENT_COLORS[idx % SEGMENT_COLORS.length];
        const isNeglected = u.health_status === 'Neglected';
        const healthLabel = isNeglected ? '⚠ Neglected Rig' : u.health_status;
        const ecoPills = (u.ecosystems || [])
          .map((e) => `<span class="eco-badge eco-${escapeHTML(e)}">${escapeHTML(e)}</span>`)
          .join('');
        const photoThumb = u.photo_url
          ? `<img src="${escapeHTML(u.photo_url)}" alt="${escapeHTML(u.device_name)}" style="width:56px;height:42px;object-fit:contain;border-radius:6px;background:#0d121a;padding:3px;border:1px solid rgba(255,255,255,0.08);flex-shrink:0;" loading="lazy" />`
          : '';

        return `
          <article class="util-card ${isNeglected ? 'neglected' : ''}">
            <div class="util-card-top">
              <div style="display:flex;align-items:center;gap:10px;">
                ${photoThumb}
                <div>
                  <div class="util-device-name">${escapeHTML(u.device_name)}</div>
                  <div class="util-device-sub">${escapeHTML(TIER_LABELS[u.performance_tier] || `Tier ${u.performance_tier}`)}</div>
                </div>
              </div>
              <span class="health-badge health-${escapeHTML(u.health_status)}">${escapeHTML(healthLabel)}</span>
            </div>

            <div class="util-metrics-row">
              <div class="util-metric">
                <span class="util-metric-label">Hours Logged</span>
                <span class="util-metric-value">${u.total_hours_logged.toFixed(1)}h</span>
              </div>
              <div class="util-metric">
                <span class="util-metric-label">Active Backlog</span>
                <span class="util-metric-value">${u.active_backlog_count} <small style="font-size:0.7rem;color:var(--text-muted)">(${u.compatible_backlog_count} compat)</small></span>
              </div>
              <div class="util-metric">
                <span class="util-metric-label">Now Playing</span>
                <span class="util-metric-value">${u.playing_count}</span>
              </div>
            </div>

            <div class="util-progress-wrap">
              <div class="util-progress-labels">
                <span>Fleet Playtime Share</span>
                <span>${u.utilization_share_pct.toFixed(1)}%</span>
              </div>
              <div class="util-progress-track">
                <div class="util-progress-fill" style="width:${Math.min(100, Math.max(4, u.utilization_share_pct))}%; background:${color};"></div>
              </div>
            </div>

            <div class="util-card-footer">
              <div class="rig-ecosystems">${ecoPills}</div>
              <button type="button" class="btn btn-sm" data-pick-device="${u.device_id}">
                Recommend for Rig →
              </button>
            </div>
          </article>
        `;
      })
      .join('');

    gridEl.querySelectorAll('[data-pick-device]').forEach((btn) => {
      btn.addEventListener('click', async () => {
        const devId = Number(btn.getAttribute('data-pick-device'));
        state.recommend.deviceId = devId;
        const recSelect = document.getElementById('rec-device-select');
        if (recSelect) recSelect.value = String(devId);
        await loadRecommendations();
        document.getElementById('section-recommender')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      });
    });
  }

  // --- SECTION 2: Next-Game Recommender ---
  async function loadRecommendations() {
    try {
      const params = new URLSearchParams({
        device_id: String(state.recommend.deviceId || 0),
        time_budget: state.recommend.timeBudget || 'any',
        min_perf: state.recommend.minPerf || 'optimal',
      });
      const data = await apiJSON(`/api/recommend?${params.toString()}`);
      state.recommend.data = data;
      renderRecommendations();
    } catch (err) {
      showToast(`Recommender error: ${err.message}`);
    }
  }

  function renderRecommendations() {
    const data = state.recommend.data;
    const grid = document.getElementById('recommend-results-grid');
    const badge = document.getElementById('rec-eligible-badge');
    if (!data || !grid) return;

    if (badge) {
      badge.textContent = `${data.total_eligible} ELIGIBLE IN BACKLOG • TOP ${data.matches.length} SHOWN`;
    }

    if (!data.matches || data.matches.length === 0) {
      grid.innerHTML = `
        <div class="empty-state" style="grid-column: 1 / -1;">
          No Backlog games match the selected Device in Hand, Time Budget, and Performance filter. Try broadening Time Budget to "Any" or switching Minimum Performance to "Include Playable".
        </div>
      `;
      return;
    }

    grid.innerHTML = data.matches
      .map((m, idx) => {
        const g = m.game;
        const reasonsHTML = (m.reasons || []).map((r) => `<li>${escapeHTML(r)}</li>`).join('');
        const coverImg = g.cover_url
          ? `<img src="${escapeHTML(g.cover_url)}" alt="${escapeHTML(g.title)}" style="width:58px;height:86px;object-fit:cover;border-radius:6px;border:1px solid rgba(255,255,255,0.12);flex-shrink:0;background:#0d121a;" loading="lazy" />`
          : '';
        return `
          <article class="rec-card ${idx === 0 ? 'rank-1' : ''}">
            <div class="rec-rank-row">
              <span class="rec-rank-badge">#${idx + 1} RECOMMENDED DISPATCH</span>
              <span class="compat-badge compat-${escapeHTML(m.match_status)}">
                ● ${escapeHTML(m.match_label)}
              </span>
            </div>

            <div style="display:flex;align-items:center;gap:12px;">
              ${coverImg}
              <div>
                <h3 class="rec-title">${escapeHTML(g.title)}</h3>
                <div class="game-sub-meta" style="margin-top:3px;">
                  Target Rig: <strong>${escapeHTML(m.target_device_name)}</strong> (Tier ${m.target_device_tier})
                </div>
              </div>
            </div>

            <div class="rec-meta-row">
              <span class="eco-badge eco-${escapeHTML(g.primary_ecosystem)}">${escapeHTML(g.primary_ecosystem)}</span>
              <span class="tier-badge">Demand Tier ${g.performance_demand_tier}</span>
              <span class="tier-badge" style="color:var(--accent-cyan);border-color:rgba(56,189,248,0.4);">
                Main: ${g.hltb_main}h • +Extra: ${g.hltb_extras}h
              </span>
            </div>

            <ul class="rec-reasons">${reasonsHTML}</ul>

            <div class="rec-footer">
              <div class="game-sub-meta">
                Est. Remaining: <strong style="color:var(--text-primary);">${m.estimated_remaining_hours}h</strong>
              </div>
              <button
                type="button"
                class="btn btn-emerald"
                data-start-game="${g.id}"
                data-target-device="${m.target_device_id}"
              >
                ▶ Start Playing on ${escapeHTML(m.target_device_name)}
              </button>
            </div>
          </article>
        `;
      })
      .join('');

    grid.querySelectorAll('[data-start-game]').forEach((btn) => {
      btn.addEventListener('click', async () => {
        const gameId = Number(btn.getAttribute('data-start-game'));
        const devId = Number(btn.getAttribute('data-target-device'));
        await startPlayingGame(gameId, devId);
      });
    });
  }

  async function startPlayingGame(gameId, deviceId) {
    try {
      const updated = await apiJSON(`/api/games/${gameId}/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ device_id: deviceId || null }),
      });
      showToast(`Started playing "${updated.title}" on ${updated.primary_device_name || 'assigned rig'}!`);
      await loadAllData();
    } catch (err) {
      showToast(`Failed to start game: ${err.message}`);
    }
  }

  // --- SECTION 3: Spec-Aware Cross-Play Library ---
  function renderLibrarySection() {
    const tbody = document.getElementById('games-table-body');
    if (!tbody) return;

    const search = state.filters.search.toLowerCase().trim();
    const eco = state.filters.ecosystem;
    const status = state.filters.status;
    const devId = Number(state.filters.deviceId) || 0;

    const filtered = state.games.filter((g) => {
      if (eco !== 'all' && g.primary_ecosystem.toLowerCase() !== eco.toLowerCase()) return false;
      if (status !== 'all' && g.status.toLowerCase() !== status.toLowerCase()) return false;
      if (search && !g.title.toLowerCase().includes(search) && !(g.genre || '').toLowerCase().includes(search)) {
        return false;
      }
      if (devId > 0) {
        const isPrimary = g.primary_device_id === devId;
        const isCompat = (g.compatibility || []).some((c) => c.device_id === devId);
        if (!isPrimary && !isCompat) return false;
      }
      return true;
    });

    if (filtered.length === 0) {
      tbody.innerHTML = `
        <tr>
          <td colspan="7">
            <div class="empty-state">No games match the active library filters.</div>
          </td>
        </tr>
      `;
      return;
    }

    tbody.innerHTML = filtered
      .map((g) => {
        const compatBadges =
          (g.compatibility || []).length > 0
            ? g.compatibility
                .map((c) => {
                  const dot = c.status === 'Optimal' ? '●' : c.status === 'Playable' ? '◐' : '✕';
                  return `<span class="compat-badge compat-${escapeHTML(c.status)}" title="Device Tier ${
                    c.device_tier
                  } vs Game Demand Tier ${c.game_tier}">
                    ${dot} ${escapeHTML(c.device_name)} (T${c.device_tier}): ${escapeHTML(c.label)}
                  </span>`;
                })
                .join('')
            : `<span class="game-sub-meta">No owned rig configured for ${escapeHTML(g.primary_ecosystem)} ecosystem</span>`;

        const subInfo = [g.genre, g.release_year ? String(g.release_year) : ''].filter(Boolean).join(' • ');
        const coverThumb = g.cover_url
          ? `<img src="${escapeHTML(g.cover_url)}" alt="${escapeHTML(g.title)}" style="width:44px;height:64px;object-fit:cover;border-radius:5px;border:1px solid rgba(255,255,255,0.12);flex-shrink:0;background:#0d121a;" loading="lazy" />`
          : `<div style="width:44px;height:64px;border-radius:5px;background:#131b26;border:1px dashed rgba(255,255,255,0.12);display:flex;align-items:center;justify-content:center;font-size:0.65rem;color:#64748b;flex-shrink:0;">NO ART</div>`;

        return `
          <tr>
            <td>
              <div style="display:flex;align-items:center;gap:12px;">
                ${coverThumb}
                <div class="game-title-cell">
                  <span class="game-title-text">${escapeHTML(g.title)}</span>
                  <div style="display:flex;align-items:center;gap:6px;margin-top:2px;">
                    <span class="eco-badge eco-${escapeHTML(g.primary_ecosystem)}">${escapeHTML(g.primary_ecosystem)}</span>
                    ${subInfo ? `<span class="game-sub-meta">${escapeHTML(subInfo)}</span>` : ''}
                  </div>
                </div>
              </div>
            </td>

            <td>
              <select class="status-select status-${escapeHTML(g.status)}" data-status-game="${g.id}" aria-label="Status for ${escapeHTML(g.title)}">
                <option value="Backlog" ${g.status === 'Backlog' ? 'selected' : ''}>Backlog</option>
                <option value="Playing" ${g.status === 'Playing' ? 'selected' : ''}>Playing</option>
                <option value="Completed" ${g.status === 'Completed' ? 'selected' : ''}>Completed</option>
              </select>
            </td>

            <td>
              <span class="tier-badge">${escapeHTML(TIER_LABELS[g.performance_demand_tier] || `Tier ${g.performance_demand_tier}`)}</span>
            </td>

            <td>
              <div class="compat-stack">${compatBadges}</div>
            </td>

            <td>
              <div class="hltb-grid-mini">
                <div class="hltb-pill"><span>Main</span><span>${g.hltb_main}h</span></div>
                <div class="hltb-pill"><span>+Extra</span><span>${g.hltb_extras}h</span></div>
                <div class="hltb-pill"><span>100%</span><span>${g.hltb_completionist}h</span></div>
              </div>
            </td>

            <td>
              <div class="hours-control">
                <strong>${g.hours_logged.toFixed(1)}h</strong>
                <button type="button" class="btn btn-sm" data-add-hours="${g.id}" data-delta="1" title="Log +1 hour">+1h</button>
              </div>
              <div class="game-sub-meta" style="margin-top:4px;">
                Rig: <strong style="color:var(--text-secondary);">${escapeHTML(g.primary_device_name || 'Unassigned')}</strong>
              </div>
            </td>

            <td>
              <div style="display:flex;gap:6px;flex-wrap:wrap;">
                ${
                  g.status === 'Backlog'
                    ? `<button type="button" class="btn btn-sm btn-emerald" data-quick-start="${g.id}">▶ Start</button>`
                    : ''
                }
                <button type="button" class="btn btn-sm" data-edit-game="${g.id}">Edit</button>
                <button type="button" class="btn btn-sm btn-danger" data-delete-game="${g.id}">Del</button>
              </div>
            </td>
          </tr>
        `;
      })
      .join('');

    // Wire inline status updates
    tbody.querySelectorAll('[data-status-game]').forEach((sel) => {
      sel.addEventListener('change', async () => {
        const id = Number(sel.getAttribute('data-status-game'));
        try {
          await apiJSON(`/api/games/${id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: sel.value }),
          });
          showToast(`Updated status to ${sel.value}`);
          await loadAllData();
        } catch (err) {
          showToast(err.message);
        }
      });
    });

    // Wire +1h logging
    tbody.querySelectorAll('[data-add-hours]').forEach((btn) => {
      btn.addEventListener('click', async () => {
        const id = Number(btn.getAttribute('data-add-hours'));
        const delta = Number(btn.getAttribute('data-delta')) || 1;
        const game = state.games.find((x) => x.id === id);
        if (!game) return;
        try {
          await apiJSON(`/api/games/${id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ hours_logged: Number((game.hours_logged + delta).toFixed(1)) }),
          });
          showToast(`Logged +${delta}h on ${game.title}`);
          await loadAllData();
        } catch (err) {
          showToast(err.message);
        }
      });
    });

    // Wire quick start
    tbody.querySelectorAll('[data-quick-start]').forEach((btn) => {
      btn.addEventListener('click', async () => {
        const id = Number(btn.getAttribute('data-quick-start'));
        const game = state.games.find((x) => x.id === id);
        await startPlayingGame(id, game?.primary_device_id || null);
      });
    });

    // Wire edit game
    tbody.querySelectorAll('[data-edit-game]').forEach((btn) => {
      btn.addEventListener('click', () => {
        const id = Number(btn.getAttribute('data-edit-game'));
        const game = state.games.find((x) => x.id === id);
        if (game) openGameModal(game);
      });
    });

    // Wire delete game
    tbody.querySelectorAll('[data-delete-game]').forEach((btn) => {
      btn.addEventListener('click', async () => {
        const id = Number(btn.getAttribute('data-delete-game'));
        const game = state.games.find((x) => x.id === id);
        if (!game) return;
        try {
          await apiJSON(`/api/games/${id}`, { method: 'DELETE' });
          showToast(`Removed "${game.title}" from library`);
          await loadAllData();
        } catch (err) {
          showToast(err.message);
        }
      });
    });
  }

  // --- SECTION 4: Hardware Asset Registry (Presets + Owned Rigs) ---
  function renderPresetsAndHardware() {
    // Populate preset prefill dropdown inside modal
    const presetSelect = document.getElementById('device-preset-prefill');
    if (presetSelect) {
      presetSelect.innerHTML =
        `<option value="">— Custom Build / Handheld (Enter Specs Manually) —</option>` +
        state.presets
          .map(
            (p) =>
              `<option value="${escapeHTML(p.id)}">${escapeHTML(p.name)} (Tier ${p.performance_tier} • ${escapeHTML(
                p.ecosystems.join('/')
              )})</option>`
          )
          .join('');
    }

    // Render Built-in Preset Catalog cards
    const presetGrid = document.getElementById('preset-catalog-grid');
    if (presetGrid) {
      presetGrid.innerHTML = state.presets
        .map((p) => {
          const alreadyOwned = state.devices.some(
            (d) => d.preset_id === p.id || d.name.toLowerCase() === p.name.toLowerCase()
          );
          return `
            <div class="preset-mini-card">
              <img src="${escapeHTML(p.photo_url)}" alt="${escapeHTML(p.name)}" style="width:100%;height:96px;object-fit:contain;padding:6px;border-radius:6px;margin-bottom:8px;background:#0d121a;border:1px solid rgba(255,255,255,0.06);" loading="lazy" />
              <div class="preset-mini-top">
                <span class="preset-mini-name">${escapeHTML(p.name)}</span>
                <span class="tier-badge">T${p.performance_tier}</span>
              </div>
              <div class="preset-mini-spec" title="${escapeHTML(p.cpu)}">${escapeHTML(p.cpu)}</div>
              <div class="preset-mini-spec" title="${escapeHTML(p.gpu)}">${escapeHTML(p.gpu)}</div>
              <div class="rig-ecosystems" style="margin-top:2px;">
                ${p.ecosystems.map((e) => `<span class="eco-badge eco-${escapeHTML(e)}">${escapeHTML(e)}</span>`).join('')}
              </div>
              <div class="preset-mini-actions">
                <button type="button" class="btn btn-sm btn-primary" data-quick-add-preset="${escapeHTML(p.id)}">
                  ${alreadyOwned ? '+ Add Another' : '+ Quick Deploy'}
                </button>
                <button type="button" class="btn btn-sm" data-customize-preset="${escapeHTML(p.id)}">
                  Customize
                </button>
              </div>
            </div>
          `;
        })
        .join('');

      presetGrid.querySelectorAll('[data-quick-add-preset]').forEach((btn) => {
        btn.addEventListener('click', async () => {
          const presetId = btn.getAttribute('data-quick-add-preset');
          const p = state.presets.find((x) => x.id === presetId);
          if (!p) return;
          try {
            const created = await apiJSON('/api/devices', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                name: p.name,
                photo_url: p.photo_url,
                cpu: p.cpu,
                gpu: p.gpu,
                ram: p.ram,
                os: p.os,
                performance_tier: p.performance_tier,
                ecosystems: p.ecosystems,
                is_preset: true,
                preset_id: p.id,
              }),
            });
            showToast(`Deployed preset "${created.name}" (Tier ${created.performance_tier}) to Hardware Registry!`);
            await loadAllData();
          } catch (err) {
            showToast(err.message);
          }
        });
      });

      presetGrid.querySelectorAll('[data-customize-preset]').forEach((btn) => {
        btn.addEventListener('click', () => {
          const presetId = btn.getAttribute('data-customize-preset');
          const p = state.presets.find((x) => x.id === presetId);
          if (p) {
            openDeviceModal(null, p);
          }
        });
      });
    }

    // Render Owned Hardware Rigs Grid
    const hwGrid = document.getElementById('hardware-rigs-grid');
    if (!hwGrid) return;

    if (state.devices.length === 0) {
      hwGrid.innerHTML = `<div class="empty-state">No hardware rigs registered yet. Deploy a preset above or click "+ Add Custom Build / Handheld".</div>`;
      return;
    }

    hwGrid.innerHTML = state.devices
      .map((d) => {
        const photoSrc = d.photo_url || '';
        const ecoBadges = (d.ecosystems || [])
          .map((e) => `<span class="eco-badge eco-${escapeHTML(e)}">${escapeHTML(e)}</span>`)
          .join('');
        return `
          <article class="rig-card">
            <div class="rig-photo-frame" style="background:#0d121a;">
              <img src="${escapeHTML(photoSrc)}" alt="${escapeHTML(d.name)}" style="object-fit:contain;padding:8px;" loading="lazy" />
              <span class="rig-tier-overlay">${escapeHTML(TIER_LABELS[d.performance_tier] || `Tier ${d.performance_tier}`)}</span>
            </div>
            <div class="rig-body">
              <div class="rig-title-row">
                <h3 class="rig-name">${escapeHTML(d.name)}</h3>
                <span class="instance-tag">${d.is_preset ? 'PRESET CONSOLE' : 'CUSTOM RIG'}</span>
              </div>

              <div class="spec-sheet">
                <span class="spec-key">CPU</span>
                <span class="spec-val" title="${escapeHTML(d.cpu)}">${escapeHTML(d.cpu || '—')}</span>
                <span class="spec-key">GPU</span>
                <span class="spec-val" title="${escapeHTML(d.gpu)}">${escapeHTML(d.gpu || '—')}</span>
                <span class="spec-key">RAM</span>
                <span class="spec-val" title="${escapeHTML(d.ram)}">${escapeHTML(d.ram || '—')}</span>
                <span class="spec-key">OS</span>
                <span class="spec-val" title="${escapeHTML(d.os)}">${escapeHTML(d.os || '—')}</span>
              </div>

              <div class="rig-ecosystems">${ecoBadges}</div>

              <div class="rig-footer">
                <button type="button" class="btn btn-sm" data-edit-device="${d.id}">Edit Specs / Photo</button>
                <button type="button" class="btn btn-sm btn-danger" data-delete-device="${d.id}">Remove</button>
              </div>
            </div>
          </article>
        `;
      })
      .join('');

    hwGrid.querySelectorAll('[data-edit-device]').forEach((btn) => {
      btn.addEventListener('click', () => {
        const id = Number(btn.getAttribute('data-edit-device'));
        const dev = state.devices.find((x) => x.id === id);
        if (dev) openDeviceModal(dev, null);
      });
    });

    hwGrid.querySelectorAll('[data-delete-device]').forEach((btn) => {
      btn.addEventListener('click', async () => {
        const id = Number(btn.getAttribute('data-delete-device'));
        const dev = state.devices.find((x) => x.id === id);
        if (!dev) return;
        try {
          await apiJSON(`/api/devices/${id}`, { method: 'DELETE' });
          showToast(`Removed hardware asset "${dev.name}"`);
          await loadAllData();
        } catch (err) {
          showToast(err.message);
        }
      });
    });
  }

  // --- Modal Management: Game Modal & Auto-Fetch HLTB + SteamGridDB ---
  function openGameModal(game = null) {
    const modal = document.getElementById('modal-game');
    const titleEl = document.getElementById('modal-game-title');
    const banner = document.getElementById('hltb-status-banner');
    if (banner) {
      banner.className = 'hltb-status-banner';
      banner.textContent = '';
    }

    document.getElementById('game-id-input').value = game ? String(game.id) : '';
    titleEl.textContent = game ? `Edit Game — ${game.title}` : 'Add Game to Spec-Aware Library';

    document.getElementById('game-title-input').value = game ? game.title : '';
    const coverInput = document.getElementById('game-cover-url-input');
    if (coverInput) coverInput.value = game ? game.cover_url || '' : '';
    document.getElementById('game-ecosystem-select').value = game ? game.primary_ecosystem : 'Steam';
    document.getElementById('game-status-select').value = game ? game.status : 'Backlog';
    document.getElementById('game-tier-select').value = game ? String(game.performance_demand_tier) : '3';
    document.getElementById('game-device-select').value =
      game && game.primary_device_id ? String(game.primary_device_id) : '';
    document.getElementById('game-hltb-main').value = game ? String(game.hltb_main) : '12';
    document.getElementById('game-hltb-extras').value = game ? String(game.hltb_extras) : '22';
    document.getElementById('game-hltb-comp').value = game ? String(game.hltb_completionist) : '45';
    document.getElementById('game-hours-logged').value = game ? String(game.hours_logged) : '0';
    document.getElementById('game-genre-input').value = game ? game.genre || '' : '';
    document.getElementById('game-year-input').value = game && game.release_year ? String(game.release_year) : '2024';

    modal.classList.add('open');
  }

  function closeGameModal() {
    document.getElementById('modal-game')?.classList.remove('open');
  }

  async function handleAutoFetchHLTB() {
    const titleInput = document.getElementById('game-title-input');
    const coverInput = document.getElementById('game-cover-url-input');
    const banner = document.getElementById('hltb-status-banner');
    const btn = document.getElementById('btn-auto-fetch-hltb');
    const query = titleInput.value.trim();

    if (!query) {
      banner.className = 'hltb-status-banner visible warn';
      banner.textContent = 'Please enter a game title first before clicking Auto-Fetch.';
      return;
    }

    btn.disabled = true;
    btn.textContent = '⏳ Querying HLTB + Cover...';
    banner.className = 'hltb-status-banner visible';
    banner.textContent = 'Fetching HLTB hours & cover art via SteamGridDB / remote catalog...';

    try {
      const est = await apiJSON(`/api/hltb?title=${encodeURIComponent(query)}`);
      if (est.cover_url && coverInput) {
        coverInput.value = est.cover_url;
      }
      if (est.found) {
        if (est.matched_title) {
          titleInput.value = est.matched_title;
        }
        document.getElementById('game-hltb-main').value = String(est.main_story);
        document.getElementById('game-hltb-extras').value = String(est.main_extras);
        document.getElementById('game-hltb-comp').value = String(est.completionist);
        if (est.suggested_demand_tier) {
          document.getElementById('game-tier-select').value = String(est.suggested_demand_tier);
        }
        if (est.suggested_ecosystem) {
          document.getElementById('game-ecosystem-select').value = est.suggested_ecosystem;
        }
        if (est.genre) {
          document.getElementById('game-genre-input').value = est.genre;
        }
        if (est.release_year) {
          document.getElementById('game-year-input').value = String(est.release_year);
        }

        banner.className = 'hltb-status-banner visible success';
        banner.textContent = `✓ [${est.source.toUpperCase()}] ${est.message} (Main: ${est.main_story}h • +Extras: ${est.main_extras}h • 100%: ${est.completionist}h)`;
      } else {
        banner.className = 'hltb-status-banner visible warn';
        banner.textContent = `⚠ ${est.message}`;
        document.getElementById('game-hltb-main').focus();
      }
    } catch (err) {
      banner.className = 'hltb-status-banner visible warn';
      banner.textContent = `HLTB lookup failed (${err.message}). Manual hour entry is active below.`;
    } finally {
      btn.disabled = false;
      btn.textContent = '⚡ Auto-Fetch HLTB + Cover';
    }
  }

  async function handleFetchSGDBCoverForGame() {
    const titleInput = document.getElementById('game-title-input');
    const coverInput = document.getElementById('game-cover-url-input');
    const banner = document.getElementById('hltb-status-banner');
    const btn = document.getElementById('btn-fetch-sgdb-cover');
    const query = titleInput.value.trim();

    if (!query) {
      banner.className = 'hltb-status-banner visible warn';
      banner.textContent = 'Enter a game title first to fetch cover art from SteamGridDB.';
      return;
    }

    btn.disabled = true;
    btn.textContent = '⏳ Fetching Cover...';
    try {
      const res = await apiJSON(`/api/steamgriddb?title=${encodeURIComponent(query)}`);
      if (res.found && res.cover_url) {
        coverInput.value = res.cover_url;
        banner.className = 'hltb-status-banner visible success';
        banner.textContent = `✓ Sourced remote cover art via ${res.source.toUpperCase()}`;
      } else {
        banner.className = 'hltb-status-banner visible warn';
        banner.textContent = res.has_api_key
          ? `No cover found on SteamGridDB for "${query}".`
          : `No cover found. Configure a SteamGridDB API key in the header to query SteamGridDB v2 grids directly!`;
      }
    } catch (err) {
      banner.className = 'hltb-status-banner visible warn';
      banner.textContent = `Cover lookup error: ${err.message}`;
    } finally {
      btn.disabled = false;
      btn.textContent = '🖼️ Fetch SteamGridDB Cover';
    }
  }

  // --- Modal Management: SteamGridDB Integration Modal ---
  async function openSGDBModal() {
    const modal = document.getElementById('modal-sgdb');
    const banner = document.getElementById('sgdb-status-banner');
    modal?.classList.add('open');
    try {
      const status = await apiJSON('/api/steamgriddb');
      if (banner) {
        banner.className = `hltb-status-banner visible ${status.has_api_key ? 'success' : 'warn'}`;
        banner.textContent = status.has_api_key
          ? '✓ SteamGridDB v2 API Key is ACTIVE. Click below to sync vertical grid covers across your library.'
          : 'ℹ No SteamGridDB API Key saved yet (currently using remote Steam/Wikipedia CDN URLs). Paste your SteamGridDB API key below to source grids directly from SteamGridDB!';
      }
    } catch (err) {
      if (banner) {
        banner.className = 'hltb-status-banner visible warn';
        banner.textContent = `Status check failed: ${err.message}`;
      }
    }
  }

  function closeSGDBModal() {
    document.getElementById('modal-sgdb')?.classList.remove('open');
  }

  // --- Modal Management: Hardware Rig Modal ---
  function openDeviceModal(device = null, preset = null) {
    const modal = document.getElementById('modal-device');
    const titleEl = document.getElementById('modal-device-title');
    const fileInput = document.getElementById('device-photo-file');
    if (fileInput) fileInput.value = '';

    document.getElementById('device-id-input').value = device ? String(device.id) : '';
    titleEl.textContent = device
      ? `Edit Hardware Rig — ${device.name}`
      : preset
      ? `Customize Preset — ${preset.name}`
      : 'Register Hardware Asset (Preset or Custom Rig)';

    const src = device || preset;
    document.getElementById('device-preset-prefill').value = preset ? preset.id : '';
    document.getElementById('device-name-input').value = src ? src.name : '';
    document.getElementById('device-tier-select').value = src ? String(src.performance_tier) : '3';
    document.getElementById('device-cpu-input').value = src ? src.cpu : '';
    document.getElementById('device-gpu-input').value = src ? src.gpu : '';
    document.getElementById('device-ram-input').value = src ? src.ram : '';
    document.getElementById('device-os-input').value = src ? src.os : '';
    document.getElementById('device-photo-url-input').value = src ? src.photo_url || '' : '';
    document.getElementById('device-preset-id-input').value = preset ? preset.id : device ? device.preset_id || '' : '';

    const activeEcos = src && src.ecosystems ? src.ecosystems : ['Steam', 'Emulation'];
    document.querySelectorAll('#device-ecosystems-group input[name="eco_check"]').forEach((cb) => {
      cb.checked = activeEcos.some((e) => e.toLowerCase() === cb.value.toLowerCase());
    });

    modal.classList.add('open');
  }

  function closeDeviceModal() {
    document.getElementById('modal-device')?.classList.remove('open');
  }

  function applyPresetToDeviceForm(presetId) {
    const p = state.presets.find((x) => x.id === presetId);
    if (!p) return;
    document.getElementById('device-name-input').value = p.name;
    document.getElementById('device-tier-select').value = String(p.performance_tier);
    document.getElementById('device-cpu-input').value = p.cpu;
    document.getElementById('device-gpu-input').value = p.gpu;
    document.getElementById('device-ram-input').value = p.ram;
    document.getElementById('device-os-input').value = p.os;
    document.getElementById('device-photo-url-input').value = p.photo_url;
    document.getElementById('device-preset-id-input').value = p.id;

    document.querySelectorAll('#device-ecosystems-group input[name="eco_check"]').forEach((cb) => {
      cb.checked = p.ecosystems.some((e) => e.toLowerCase() === cb.value.toLowerCase());
    });
  }

  // --- Section Tab Switching ---
  function switchTab(section) {
    state.activeTab = section;
    document.querySelectorAll('.nav-tab').forEach((btn) => {
      btn.classList.toggle('active', btn.getAttribute('data-section') === section);
    });

    const secBrief = document.getElementById('section-command-brief');
    const secUtil = document.getElementById('section-utilization');
    const secRec = document.getElementById('section-recommender');
    const secLib = document.getElementById('section-library');
    const secHw = document.getElementById('section-hardware');

    if (section === 'all') {
      if (secBrief) secBrief.style.display = '';
      secUtil.style.display = '';
      secRec.style.display = '';
      secLib.style.display = '';
      secHw.style.display = '';
    } else {
      if (secBrief) secBrief.style.display = 'none';
      secUtil.style.display = section === 'recommender' ? '' : 'none';
      secRec.style.display = section === 'recommender' ? '' : 'none';
      secLib.style.display = section === 'library' ? '' : 'none';
      secHw.style.display = section === 'hardware' ? '' : 'none';
    }
  }

  // --- Event Listeners Setup ---
  function initEvents() {
    // Top tabs
    document.querySelectorAll('.nav-tab').forEach((btn) => {
      btn.addEventListener('click', () => {
        switchTab(btn.getAttribute('data-section'));
      });
    });

    // Open Modals
    document.getElementById('btn-open-add-game')?.addEventListener('click', () => openGameModal(null));
    document.getElementById('btn-section-add-game')?.addEventListener('click', () => openGameModal(null));
    document.getElementById('btn-close-game-modal')?.addEventListener('click', closeGameModal);
    document.getElementById('btn-cancel-game-modal')?.addEventListener('click', closeGameModal);

    document.getElementById('btn-open-add-device')?.addEventListener('click', () => openDeviceModal(null, null));
    document.getElementById('btn-section-add-device')?.addEventListener('click', () => openDeviceModal(null, null));
    document.getElementById('btn-close-device-modal')?.addEventListener('click', closeDeviceModal);
    document.getElementById('btn-cancel-device-modal')?.addEventListener('click', closeDeviceModal);

    // SteamGridDB Modal & Sync
    document.getElementById('btn-open-sgdb-modal')?.addEventListener('click', openSGDBModal);
    document.getElementById('btn-library-sync-sgdb')?.addEventListener('click', openSGDBModal);
    document.getElementById('btn-close-sgdb-modal')?.addEventListener('click', closeSGDBModal);
    document.getElementById('btn-cancel-sgdb-modal')?.addEventListener('click', closeSGDBModal);

    document.getElementById('form-sgdb')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      const apiKey = document.getElementById('sgdb-api-key-input')?.value.trim() || '';
      const forceAll = Boolean(document.getElementById('sgdb-force-all-check')?.checked);
      const btn = document.getElementById('btn-save-sync-sgdb');
      const banner = document.getElementById('sgdb-status-banner');
      if (btn) {
        btn.disabled = true;
        btn.textContent = '⏳ Syncing Covers with SteamGridDB...';
      }
      try {
        const res = await apiJSON('/api/steamgriddb', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ api_key: apiKey, sync_games: true, force_all: forceAll }),
        });
        showToast(`SteamGridDB sync complete! Updated ${res.updated_games} game cover(s).`);
        if (banner) {
          banner.className = 'hltb-status-banner visible success';
          banner.textContent = `✓ Synced! Updated ${res.updated_games} game cover(s). SteamGridDB Key Active: ${res.has_api_key ? 'Yes' : 'No'}`;
        }
        await loadAllData();
      } catch (err) {
        showToast(`SteamGridDB error: ${err.message}`);
      } finally {
        if (btn) {
          btn.disabled = false;
          btn.textContent = '⚡ Save Key & Sync Library Covers';
        }
      }
    });

    // Auto-Fetch HLTB & SteamGridDB Cover buttons
    document.getElementById('btn-auto-fetch-hltb')?.addEventListener('click', handleAutoFetchHLTB);
    document.getElementById('btn-fetch-sgdb-cover')?.addEventListener('click', handleFetchSGDBCoverForGame);

    // Preset prefill dropdown in device modal
    document.getElementById('device-preset-prefill')?.addEventListener('change', (e) => {
      if (e.target.value) {
        applyPresetToDeviceForm(e.target.value);
      }
    });

    // Save Game Form
    document.getElementById('form-game')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      const idVal = document.getElementById('game-id-input').value;
      const devVal = document.getElementById('game-device-select').value;
      const payload = {
        title: document.getElementById('game-title-input').value.trim(),
        cover_url: document.getElementById('game-cover-url-input')?.value.trim() || '',
        primary_ecosystem: document.getElementById('game-ecosystem-select').value,
        status: document.getElementById('game-status-select').value,
        performance_demand_tier: Number(document.getElementById('game-tier-select').value) || 3,
        primary_device_id: devVal ? Number(devVal) : null,
        hltb_main: Number(document.getElementById('game-hltb-main').value) || 0,
        hltb_extras: Number(document.getElementById('game-hltb-extras').value) || 0,
        hltb_completionist: Number(document.getElementById('game-hltb-comp').value) || 0,
        hours_logged: Number(document.getElementById('game-hours-logged').value) || 0,
        genre: document.getElementById('game-genre-input').value.trim(),
        release_year: Number(document.getElementById('game-year-input').value) || 0,
      };

      try {
        if (idVal) {
          await apiJSON(`/api/games/${idVal}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
          });
          showToast(`Updated "${payload.title}"`);
        } else {
          await apiJSON('/api/games', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
          });
          showToast(`Added "${payload.title}" to library!`);
        }
        closeGameModal();
        await loadAllData();
      } catch (err) {
        showToast(`Save error: ${err.message}`);
      }
    });

    // Save Device Form (uses multipart/form-data to support local photo upload to $UPLOAD_DIR)
    document.getElementById('form-device')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      const idVal = document.getElementById('device-id-input').value;
      const checkedEcos = Array.from(
        document.querySelectorAll('#device-ecosystems-group input[name="eco_check"]:checked')
      ).map((cb) => cb.value);

      if (checkedEcos.length === 0) {
        showToast('Select at least one supported gaming ecosystem.');
        return;
      }

      const fd = new FormData();
      fd.append('name', document.getElementById('device-name-input').value.trim());
      fd.append('performance_tier', document.getElementById('device-tier-select').value);
      fd.append('cpu', document.getElementById('device-cpu-input').value.trim());
      fd.append('gpu', document.getElementById('device-gpu-input').value.trim());
      fd.append('ram', document.getElementById('device-ram-input').value.trim());
      fd.append('os', document.getElementById('device-os-input').value.trim());
      fd.append('photo_url', document.getElementById('device-photo-url-input').value);
      fd.append('preset_id', document.getElementById('device-preset-id-input').value);
      fd.append('is_preset', document.getElementById('device-preset-id-input').value ? 'true' : 'false');
      fd.append('ecosystems_json', JSON.stringify(checkedEcos));

      const fileInput = document.getElementById('device-photo-file');
      if (fileInput && fileInput.files && fileInput.files[0]) {
        fd.append('photo', fileInput.files[0]);
      }

      try {
        const url = idVal ? `/api/devices/${idVal}` : '/api/devices';
        const method = idVal ? 'PUT' : 'POST';
        const saved = await apiJSON(url, { method, body: fd });
        showToast(`Saved hardware rig "${saved.name}" (Tier ${saved.performance_tier})!`);
        closeDeviceModal();
        await loadAllData();
      } catch (err) {
        showToast(`Hardware save error: ${err.message}`);
      }
    });

    // Recommender controls
    document.getElementById('rec-device-select')?.addEventListener('change', (e) => {
      state.recommend.deviceId = Number(e.target.value) || 0;
      loadRecommendations();
    });

    document.querySelectorAll('#rec-time-group .seg-btn').forEach((btn) => {
      btn.addEventListener('click', () => {
        document.querySelectorAll('#rec-time-group .seg-btn').forEach((b) => b.classList.remove('active'));
        btn.classList.add('active');
        state.recommend.timeBudget = btn.getAttribute('data-budget');
        loadRecommendations();
      });
    });

    document.querySelectorAll('#rec-perf-group .seg-btn').forEach((btn) => {
      btn.addEventListener('click', () => {
        document.querySelectorAll('#rec-perf-group .seg-btn').forEach((b) => b.classList.remove('active'));
        btn.classList.add('active');
        state.recommend.minPerf = btn.getAttribute('data-perf');
        loadRecommendations();
      });
    });

    // Library filters
    document.getElementById('lib-search-input')?.addEventListener('input', (e) => {
      state.filters.search = e.target.value;
      renderLibrarySection();
    });

    document.getElementById('lib-ecosystem-filter')?.addEventListener('change', (e) => {
      state.filters.ecosystem = e.target.value;
      renderLibrarySection();
    });

    document.getElementById('lib-status-filter')?.addEventListener('change', (e) => {
      state.filters.status = e.target.value;
      renderLibrarySection();
    });

    document.getElementById('lib-device-filter')?.addEventListener('change', (e) => {
      state.filters.deviceId = Number(e.target.value) || 0;
      renderLibrarySection();
    });
  }

  document.addEventListener('DOMContentLoaded', () => {
    initEvents();
    loadAllData();
    loadCatalogDatalist();
  });
})();
