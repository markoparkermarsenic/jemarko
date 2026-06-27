<script lang="ts">
    // ── Types ──────────────────────────────────────────────────────────────
    interface GuestMember {
        name: string;
        rsvpStatus: 'attending' | 'not_attending' | 'no_response';
        verified: boolean;
    }

    interface GuestGroup {
        address: string;
        members: GuestMember[];
    }

    interface DietaryEntry {
        name: string;
        email: string;
        diet: string;
        verified: boolean;
    }

    interface UnverifiedRSVP {
        name: string;
        email: string;
        isAttending: boolean;
        attendingGuests: string[];
        diet: string;
        submittedAt: string;
    }

    interface Stats {
        totalInvited: number;
        totalRSVPd: number;
        attending: number;
        notAttending: number;
        noResponse: number;
        withDietary: number;
        unverifiedCount: number;
    }

    interface DashboardData {
        success: boolean;
        stats: Stats;
        guestGroups: GuestGroup[];
        dietaryRequirements: DietaryEntry[];
        unverifiedRSVPs: UnverifiedRSVP[];
    }

    // ── State ──────────────────────────────────────────────────────────────
    type View = 'login' | 'dashboard';

    let view = $state<View>('login');
    let token = $state('');

    // Login form
    let username = $state('');
    let password = $state('');
    let loginError = $state('');
    let loginLoading = $state(false);

    // Dashboard
    let dashboard = $state<DashboardData | null>(null);
    let dashboardLoading = $state(false);
    let dashboardError = $state('');

    // Active section tab
    let activeSection = $state<'rsvp' | 'dietary' | 'unverified'>('rsvp');

    // Search / filter
    let rsvpFilter = $state<'all' | 'attending' | 'not_attending' | 'no_response'>('all');
    let searchQuery = $state('');

    // ── Derived ────────────────────────────────────────────────────────────
    let filteredGroups = $derived.by(() => {
        if (!dashboard) return [];
        const q = searchQuery.toLowerCase().trim();
        return dashboard.guestGroups
            .map(group => ({
                ...group,
                members: group.members.filter(m => {
                    const matchesFilter = rsvpFilter === 'all' || m.rsvpStatus === rsvpFilter;
                    const matchesSearch = !q || m.name.toLowerCase().includes(q) || group.address.toLowerCase().includes(q);
                    return matchesFilter && matchesSearch;
                })
            }))
            .filter(group => group.members.length > 0);
    });

    // ── Helpers ────────────────────────────────────────────────────────────
    function formatDate(iso: string): string {
        if (!iso) return '—';
        try {
            return new Date(iso).toLocaleDateString('en-GB', {
                day: 'numeric', month: 'short', year: 'numeric',
                hour: '2-digit', minute: '2-digit'
            });
        } catch {
            return iso;
        }
    }

    function statusLabel(status: GuestMember['rsvpStatus']): string {
        if (status === 'attending') return '✅ Attending';
        if (status === 'not_attending') return '❌ Not Attending';
        return '⏳ No Response';
    }

    // ── API calls ──────────────────────────────────────────────────────────
    async function handleLogin() {
        loginError = '';
        if (!username.trim() || !password.trim()) {
            loginError = 'Please enter your username and password.';
            return;
        }
        loginLoading = true;
        try {
            const res = await fetch('/api/admin-login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username: username.trim(), password })
            });
            const data = await res.json();
            if (!res.ok || !data.success) {
                loginError = data.message || 'Invalid username or password.';
                loginLoading = false;
                return;
            }
            token = data.token;
            await loadDashboard();
            view = 'dashboard';
        } catch {
            loginError = 'Network error — please try again.';
        }
        loginLoading = false;
    }

    async function loadDashboard() {
        dashboardLoading = true;
        dashboardError = '';
        try {
            const res = await fetch('/api/admin-dashboard', {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (res.status === 401) {
                view = 'login';
                loginError = 'Session expired — please log in again.';
                dashboardLoading = false;
                return;
            }
            const data: DashboardData = await res.json();
            if (!data.success) {
                dashboardError = 'Failed to load dashboard data.';
            } else {
                dashboard = data;
            }
        } catch {
            dashboardError = 'Network error — could not load dashboard.';
        }
        dashboardLoading = false;
    }

    function handleLogout() {
        token = '';
        dashboard = null;
        username = '';
        password = '';
        loginError = '';
        view = 'login';
    }
</script>

<svelte:head>
    <title>Admin — Jemarko</title>
</svelte:head>

{#if view === 'login'}
<div class="login-page">
    <div class="login-card">
        <img src="/j_and_m.png" alt="Jemarko" class="login-logo" />
        <h1 class="login-title">Admin Dashboard</h1>
        <p class="login-subtitle">Sign in to view RSVP data</p>
        <form class="login-form" onsubmit={(e) => { e.preventDefault(); handleLogin(); }}>
            <div class="field">
                <label for="username">Username</label>
                <input id="username" type="text" bind:value={username}
                    placeholder="jemarko" autocomplete="username" disabled={loginLoading} />
            </div>
            <div class="field">
                <label for="password">Password</label>
                <input id="password" type="password" bind:value={password}
                    placeholder="••••••••" autocomplete="current-password" disabled={loginLoading} />
            </div>
            {#if loginError}
                <p class="error-msg">{loginError}</p>
            {/if}
            <button type="submit" class="login-btn" disabled={loginLoading}>
                {loginLoading ? 'Signing in…' : 'Sign In'}
            </button>
        </form>
    </div>
</div>

{:else}
<div class="dashboard">
    <header class="topbar">
        <div class="topbar-left">
            <img src="/j_and_m.png" alt="Jemarko" class="topbar-logo" />
            <h1 class="topbar-title">Admin Dashboard</h1>
        </div>
        <div class="topbar-right">
            <button class="refresh-btn" onclick={loadDashboard} disabled={dashboardLoading}>
                {dashboardLoading ? '↻ Loading…' : '↻ Refresh'}
            </button>
            <button class="logout-btn" onclick={handleLogout}>Sign Out</button>
        </div>
    </header>

    {#if dashboardLoading && !dashboard}
        <div class="loading-state"><p>Loading dashboard data…</p></div>
    {:else if dashboardError}
        <div class="error-banner">{dashboardError}</div>
    {:else if dashboard}
        <section class="stats-bar">
            <div class="stat-card">
                <span class="stat-number">{dashboard.stats.totalInvited}</span>
                <span class="stat-label">Invited</span>
            </div>
            <div class="stat-card stat-highlight">
                <span class="stat-number">{dashboard.stats.totalRSVPd}</span>
                <span class="stat-label">RSVPd</span>
            </div>
            <div class="stat-card stat-attending">
                <span class="stat-number">{dashboard.stats.attending}</span>
                <span class="stat-label">Attending</span>
            </div>
            <div class="stat-card stat-declined">
                <span class="stat-number">{dashboard.stats.notAttending}</span>
                <span class="stat-label">Not Attending</span>
            </div>
            <div class="stat-card stat-pending">
                <span class="stat-number">{dashboard.stats.noResponse}</span>
                <span class="stat-label">No Response</span>
            </div>
            <div class="stat-card stat-dietary">
                <span class="stat-number">{dashboard.stats.withDietary}</span>
                <span class="stat-label">Dietary Needs</span>
            </div>
            {#if dashboard.stats.unverifiedCount > 0}
            <div class="stat-card stat-unverified">
                <span class="stat-number">{dashboard.stats.unverifiedCount}</span>
                <span class="stat-label">Unverified</span>
            </div>
            {/if}
        </section>

        <nav class="section-tabs">
            <button class="tab-btn {activeSection === 'rsvp' ? 'active' : ''}"
                onclick={() => activeSection = 'rsvp'}>RSVP Breakdown</button>
            <button class="tab-btn {activeSection === 'dietary' ? 'active' : ''}"
                onclick={() => activeSection = 'dietary'}>
                Dietary Requirements
                {#if dashboard.dietaryRequirements.length > 0}
                    <span class="badge">{dashboard.dietaryRequirements.length}</span>
                {/if}
            </button>
            {#if dashboard.unverifiedRSVPs.length > 0}
            <button class="tab-btn {activeSection === 'unverified' ? 'active' : ''}"
                onclick={() => activeSection = 'unverified'}>
                Unverified RSVPs
                <span class="badge badge-warn">{dashboard.unverifiedRSVPs.length}</span>
            </button>
            {/if}
        </nav>


        {#if activeSection === 'rsvp'}
        <section class="content-section">
            <div class="filters-row">
                <input class="search-input" type="text"
                    placeholder="Search by name or address…" bind:value={searchQuery} />
                <div class="filter-pills">
                    {#each (['all', 'attending', 'not_attending', 'no_response'] as const) as f}
                        <button class="pill {rsvpFilter === f ? 'pill-active' : ''}"
                            onclick={() => rsvpFilter = f}>
                            {f === 'all' ? 'All' : f === 'attending' ? '✅ Attending' : f === 'not_attending' ? '❌ Not Attending' : '⏳ No Response'}
                        </button>
                    {/each}
                </div>
            </div>
            {#if filteredGroups.length === 0}
                <p class="empty-msg">No guests match your filter.</p>
            {:else}
                <div class="groups-list">
                    {#each filteredGroups as group}
                        <div class="group-card">
                            {#if group.address}
                                <div class="group-address">📍 {group.address}</div>
                            {:else}
                                <div class="group-address group-address-individual">Individual guest</div>
                            {/if}
                            <table class="members-table">
                                <thead>
                                    <tr><th>Name</th><th>Status</th><th>Verified</th></tr>
                                </thead>
                                <tbody>
                                    {#each group.members as member}
                                        <tr class="member-row member-row--{member.rsvpStatus}">
                                            <td class="member-name">{member.name}</td>
                                            <td>
                                                <span class="status-badge status--{member.rsvpStatus}">
                                                    {statusLabel(member.rsvpStatus)}
                                                </span>
                                            </td>
                                            <td>
                                                {#if member.verified}
                                                    <span class="verified-badge">✓ Verified</span>
                                                {:else if member.rsvpStatus !== 'no_response'}
                                                    <span class="unverified-badge">⚠ Pending</span>
                                                {:else}
                                                    <span class="na-text">—</span>
                                                {/if}
                                            </td>
                                        </tr>
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                    {/each}
                </div>
            {/if}
        </section>

        {:else if activeSection === 'dietary'}
        <section class="content-section">
            {#if dashboard.dietaryRequirements.length === 0}
                <p class="empty-msg">No dietary requirements have been submitted yet.</p>
            {:else}
                <p class="section-intro">
                    {dashboard.dietaryRequirements.length} guest{dashboard.dietaryRequirements.length !== 1 ? 's have' : ' has'} submitted dietary requirements.
                </p>
                <table class="dietary-table">
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Email</th>
                            <th>Dietary Requirements</th>
                            <th>Verified</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each dashboard.dietaryRequirements as entry}
                            <tr>
                                <td class="dietary-name">{entry.name}</td>
                                <td class="dietary-email">{entry.email}</td>
                                <td class="dietary-req">{entry.diet}</td>
                                <td>
                                    {#if entry.verified}
                                        <span class="verified-badge">✓ Verified</span>
                                    {:else}
                                        <span class="unverified-badge">⚠ Pending</span>
                                    {/if}
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            {/if}
        </section>

        {:else if activeSection === 'unverified'}
        <section class="content-section">
            {#if dashboard.unverifiedRSVPs.length === 0}
                <p class="empty-msg">No unverified RSVPs.</p>
            {:else}
                <p class="section-intro warn-intro">
                    ⚠️ These RSVPs were submitted by people not on the invite list.
                </p>
                <div class="unverified-list">
                    {#each dashboard.unverifiedRSVPs as rsvp}
                        <div class="unverified-card">
                            <div class="unverified-header">
                                <div>
                                    <span class="unverified-name">{rsvp.name}</span>
                                    <span class="unverified-email">{rsvp.email}</span>
                                </div>
                                <span class="unverified-date">{formatDate(rsvp.submittedAt)}</span>
                            </div>
                            <div class="unverified-body">
                                <div class="unverified-row">
                                    <span class="u-label">Status:</span>
                                    <span>{rsvp.isAttending ? '✅ Attending' : '❌ Not Attending'}</span>
                                </div>
                                {#if rsvp.attendingGuests && rsvp.attendingGuests.length > 0}
                                    <div class="unverified-row">
                                        <span class="u-label">Guests:</span>
                                        <span>{rsvp.attendingGuests.join(', ')}</span>
                                    </div>
                                {/if}
                                {#if rsvp.diet}
                                    <div class="unverified-row">
                                        <span class="u-label">Dietary:</span>
                                        <span>{rsvp.diet}</span>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </section>
        {/if}

    {/if}
</div>
{/if}


<style>
    /* ── Global resets for this page ──────────────────────────────────── */
    :global(body) { background: #f5f5f5; }

    /* ══════════════════════════════════════════════════════════════════ */
    /*  LOGIN PAGE                                                        */
    /* ══════════════════════════════════════════════════════════════════ */
    .login-page {
        min-height: 100vh;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: var(--spacing-lg);
        background: var(--color-background-alt);
    }

    .login-card {
        background: var(--color-white);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-lg);
        padding: var(--spacing-2xl);
        width: 100%;
        max-width: 400px;
        text-align: center;
        box-shadow: var(--shadow-lg);
    }

    .login-logo {
        width: 120px;
        height: auto;
        margin-bottom: var(--spacing-lg);
    }

    .login-title {
        font-family: var(--font-display);
        font-size: 1.8rem;
        margin-bottom: var(--spacing-xs);
    }

    .login-subtitle {
        font-family: var(--font-body);
        color: var(--color-text-light);
        font-size: 0.9rem;
        margin-bottom: var(--spacing-xl);
    }

    .login-form {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-md);
        text-align: left;
    }

    .field { display: flex; flex-direction: column; gap: var(--spacing-xs); }

    .field label {
        font-family: var(--font-body);
        font-size: 0.85rem;
        font-weight: 600;
        color: var(--color-text);
    }

    .field input {
        font-family: var(--font-body);
        font-size: 1rem;
        padding: var(--spacing-sm) var(--spacing-md);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        background: var(--color-white);
        color: var(--color-text);
        width: 100%;
    }

    .field input:focus {
        outline: none;
        box-shadow: 0 0 0 3px rgba(0,0,0,0.1);
    }

    .field input:disabled { opacity: 0.6; cursor: not-allowed; }

    .error-msg {
        font-family: var(--font-body);
        font-size: 0.85rem;
        color: #c00;
        background: #fff0f0;
        border: 1px solid #fcc;
        border-radius: var(--radius-sm);
        padding: var(--spacing-sm) var(--spacing-md);
        margin: 0;
        text-align: center;
    }

    .login-btn {
        font-family: var(--font-mimko);
        font-size: 1.1rem;
        padding: var(--spacing-md);
        background: var(--color-text);
        color: var(--color-white);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: opacity var(--transition-fast);
        margin-top: var(--spacing-sm);
    }

    .login-btn:hover:not(:disabled) { opacity: 0.8; }
    .login-btn:disabled { opacity: 0.5; cursor: not-allowed; }

    /* ══════════════════════════════════════════════════════════════════ */
    /*  DASHBOARD SHELL                                                   */
    /* ══════════════════════════════════════════════════════════════════ */
    .dashboard {
        min-height: 100vh;
        background: var(--color-background-alt);
        font-family: var(--font-body);
    }

    /* Top bar */
    .topbar {
        background: var(--color-white);
        border-bottom: 2px solid var(--color-border);
        padding: var(--spacing-md) var(--spacing-xl);
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: var(--spacing-md);
        position: sticky;
        top: 0;
        z-index: 100;
    }

    .topbar-left {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
    }

    .topbar-logo { width: 50px; height: auto; }

    .topbar-title {
        font-family: var(--font-display);
        font-size: 1.3rem;
        margin: 0;
    }

    .topbar-right {
        display: flex;
        gap: var(--spacing-sm);
        align-items: center;
    }

    .refresh-btn, .logout-btn {
        font-family: var(--font-body);
        font-size: 0.85rem;
        padding: var(--spacing-xs) var(--spacing-md);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        background: var(--color-white);
        cursor: pointer;
        transition: all var(--transition-fast);
    }

    .refresh-btn:hover:not(:disabled), .logout-btn:hover {
        background: var(--color-text);
        color: var(--color-white);
    }

    .refresh-btn:disabled { opacity: 0.5; cursor: not-allowed; }

    /* Loading / error states */
    .loading-state {
        display: flex;
        justify-content: center;
        padding: var(--spacing-3xl);
        color: var(--color-text-light);
        font-family: var(--font-body);
    }

    .error-banner {
        margin: var(--spacing-xl);
        padding: var(--spacing-md) var(--spacing-lg);
        background: #fff0f0;
        border: 2px solid #f00;
        border-radius: var(--radius-md);
        color: #c00;
        font-family: var(--font-body);
    }

    /* Stats */
    .stats-bar {
        display: flex;
        flex-wrap: wrap;
        gap: var(--spacing-md);
        padding: var(--spacing-xl);
    }
    .stat-card {
        flex: 1 1 120px;
        background: var(--color-white);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        padding: var(--spacing-lg);
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--spacing-xs);
        min-width: 100px;
    }
    .stat-number { font-family: var(--font-display); font-size: 2.2rem; font-weight: 700; line-height: 1; }
    .stat-label  { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--color-text-light); text-align: center; }
    .stat-highlight  { border-width: 3px; }
    .stat-attending  { background: #f0fff4; border-color: #2d6a4f; }
    .stat-attending .stat-number  { color: #2d6a4f; }
    .stat-declined   { background: #fff5f5; border-color: #9b2226; }
    .stat-declined .stat-number   { color: #9b2226; }
    .stat-pending    { background: #fffbeb; border-color: #b45309; }
    .stat-pending .stat-number    { color: #b45309; }
    .stat-dietary    { background: #eff6ff; border-color: #1d4ed8; }
    .stat-dietary .stat-number    { color: #1d4ed8; }
    .stat-unverified { background: #fff7ed; border-color: #c2410c; }
    .stat-unverified .stat-number { color: #c2410c; }

    /* Tabs */
    .section-tabs {
        display: flex;
        gap: 0;
        padding: 0 var(--spacing-xl);
        border-bottom: 2px solid var(--color-border);
        background: var(--color-white);
    }
    .tab-btn {
        font-family: var(--font-body);
        font-size: 0.9rem;
        padding: var(--spacing-md) var(--spacing-lg);
        border: none;
        border-bottom: 3px solid transparent;
        background: transparent;
        cursor: pointer;
        color: var(--color-text-light);
        transition: all var(--transition-fast);
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        margin-bottom: -2px;
    }
    .tab-btn:hover { color: var(--color-text); }
    .tab-btn.active { color: var(--color-text); border-bottom-color: var(--color-border); font-weight: 600; }
    .badge { background: var(--color-text); color: var(--color-white); font-size: 0.7rem; font-weight: 700; padding: 2px 7px; border-radius: var(--radius-full); min-width: 20px; text-align: center; }
    .badge-warn { background: #c2410c; }

    /* Content section */
    .content-section { padding: var(--spacing-xl); max-width: 1100px; margin: 0 auto; }
    .section-intro { color: var(--color-text-light); margin-bottom: var(--spacing-lg); }
    .warn-intro { color: #92400e; background: #fffbeb; border: 1px solid #fbbf24; border-radius: var(--radius-md); padding: var(--spacing-sm) var(--spacing-md); }
    .empty-msg { color: var(--color-text-light); text-align: center; padding: var(--spacing-3xl); border: 2px dashed var(--color-border-light); border-radius: var(--radius-md); background: var(--color-white); }

    /* Filters */
    .filters-row { display: flex; flex-wrap: wrap; gap: var(--spacing-md); align-items: center; margin-bottom: var(--spacing-lg); }
    .search-input { flex: 1 1 240px; font-size: 0.9rem; padding: var(--spacing-sm) var(--spacing-md); border: 2px solid var(--color-border); border-radius: var(--radius-md); background: var(--color-white); }
    .search-input:focus { outline: none; box-shadow: 0 0 0 3px rgba(0,0,0,0.1); }
    .filter-pills { display: flex; flex-wrap: wrap; gap: var(--spacing-sm); }
    .pill { font-size: 0.8rem; padding: 4px 12px; border: 2px solid var(--color-border); border-radius: var(--radius-full); background: var(--color-white); cursor: pointer; transition: all var(--transition-fast); white-space: nowrap; }
    .pill:hover { background: var(--color-background-alt); }
    .pill-active { background: var(--color-text); color: var(--color-white); }

    /* Group cards */
    .groups-list { display: flex; flex-direction: column; gap: var(--spacing-lg); }
    .group-card { background: var(--color-white); border: 2px solid var(--color-border); border-radius: var(--radius-md); overflow: hidden; }
    .group-address { padding: var(--spacing-sm) var(--spacing-lg); background: var(--color-background-alt); font-size: 0.85rem; font-weight: 600; border-bottom: 1px solid var(--color-border-light); }
    .group-address-individual { color: var(--color-text-light); font-style: italic; }

    /* Members table */
    .members-table { width: 100%; border-collapse: collapse; }
    .members-table th { text-align: left; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--color-text-light); padding: var(--spacing-sm) var(--spacing-lg); background: var(--color-white); border-bottom: 1px solid var(--color-border-light); }
    .members-table td { padding: var(--spacing-md) var(--spacing-lg); border-bottom: 1px solid var(--color-background-alt); font-size: 0.95rem; }
    .member-row:last-child td { border-bottom: none; }
    .member-name { font-weight: 600; }

    /* Badges */
    .status-badge { display: inline-block; font-size: 0.8rem; padding: 3px 10px; border-radius: var(--radius-full); font-weight: 600; white-space: nowrap; }
    .status--attending     { background: #d1fae5; color: #065f46; }
    .status--not_attending { background: #fee2e2; color: #991b1b; }
    .status--no_response   { background: #fef3c7; color: #92400e; }
    .verified-badge   { font-size: 0.8rem; color: #065f46; background: #d1fae5; padding: 2px 8px; border-radius: var(--radius-full); white-space: nowrap; }
    .unverified-badge { font-size: 0.8rem; color: #92400e; background: #fef3c7; padding: 2px 8px; border-radius: var(--radius-full); white-space: nowrap; }
    .na-text { color: var(--color-text-light); }

    /* Dietary table */
    .dietary-table { width: 100%; border-collapse: collapse; background: var(--color-white); border: 2px solid var(--color-border); border-radius: var(--radius-md); overflow: hidden; }
    .dietary-table th { text-align: left; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--color-text-light); padding: var(--spacing-sm) var(--spacing-lg); background: var(--color-background-alt); border-bottom: 2px solid var(--color-border-light); }
    .dietary-table td { padding: var(--spacing-md) var(--spacing-lg); border-bottom: 1px solid var(--color-background-alt); font-size: 0.95rem; vertical-align: top; }
    .dietary-table tr:last-child td { border-bottom: none; }
    .dietary-name  { font-weight: 600; }
    .dietary-email { color: var(--color-text-light); font-size: 0.85rem; }
    .dietary-req   { font-style: italic; max-width: 400px; }

    /* Unverified cards */
    .unverified-list { display: flex; flex-direction: column; gap: var(--spacing-md); }
    .unverified-card { background: var(--color-white); border: 2px solid #fbbf24; border-radius: var(--radius-md); overflow: hidden; }
    .unverified-header { display: flex; justify-content: space-between; align-items: flex-start; padding: var(--spacing-md) var(--spacing-lg); background: #fffbeb; border-bottom: 1px solid #fde68a; gap: var(--spacing-md); flex-wrap: wrap; }
    .unverified-name  { font-weight: 700; font-size: 1rem; display: block; }
    .unverified-email { font-size: 0.85rem; color: var(--color-text-light); display: block; }
    .unverified-date  { font-size: 0.8rem; color: var(--color-text-light); white-space: nowrap; }
    .unverified-body  { padding: var(--spacing-md) var(--spacing-lg); display: flex; flex-direction: column; gap: var(--spacing-sm); }
    .unverified-row   { display: flex; gap: var(--spacing-md); font-size: 0.9rem; }
    .u-label { font-weight: 600; min-width: 70px; color: var(--color-text-light); }

    /* Responsive */
    @media (max-width: 640px) {
        .topbar { padding: var(--spacing-sm) var(--spacing-md); }
        .topbar-title { font-size: 1rem; }
        .stats-bar { padding: var(--spacing-md); gap: var(--spacing-sm); }
        .stat-card { padding: var(--spacing-md); flex: 1 1 80px; }
        .stat-number { font-size: 1.6rem; }
        .section-tabs { padding: 0 var(--spacing-md); overflow-x: auto; }
        .tab-btn { padding: var(--spacing-sm) var(--spacing-md); font-size: 0.8rem; }
        .content-section { padding: var(--spacing-md); }
        .members-table th, .members-table td { padding: var(--spacing-sm) var(--spacing-md); }
        .dietary-table th, .dietary-table td { padding: var(--spacing-sm) var(--spacing-md); }
        .dietary-email { display: none; }
        .filter-pills { gap: 4px; }
    }
</style>

