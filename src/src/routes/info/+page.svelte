<script lang="ts">
    // No dynamic data needed — purely informational
</script>

<svelte:head>
    <title>Info — Jemarko Wedding</title>
    <meta name="description" content="All the info you need for Jemima & Marko's wedding — location, theme, food, gifts and more." />
</svelte:head>

<div class="info-page">
    <a href="/" class="back-btn" data-sveltekit-reload>← Back</a>

    <header class="info-header">
        <img src="/j_and_m.png" alt="Jemarko" class="logo" />
        <h1>Wedding Info</h1>
    </header>

    <main class="info-content">

        <!-- LOCATION -->
        <section class="info-section">
            <h2>📍 Location</h2>

            <div class="sub-section">
                <h3>Ceremony: 16:00</h3>
                <p class="venue-name">Wandsworth Town Hall</p>
                <p class="note">
                    ⚠️ You need a <strong>specific invitation to the ceremony</strong> — this will be noted on
                    the invitation you were sent. If you're unsure whether you're invited to the ceremony, please reach out to us!
                </p>
            </div>

            <div class="sub-section">
                <h3>Reception: 18:00 - 23:00</h3>
                <p class="venue-name">Sands End Arts &amp; Community Centre</p>
                <p class="venue-address">Peterborough Rd, London SW6 3EZ</p>
                <p class="note">Enter through Southpark on Clancarty Road.</p>
                <a
                    class="map-link"
                    href="https://maps.google.com/?q=Sands+End+Arts+%26+Community+Centre,+Peterborough+Rd,+London+SW6+3EZ"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    View on Google Maps →
                </a>
            </div>
        </section>

        <!-- THEME -->
        <section class="info-section">
            <h2>🦜 Theme</h2>
            <p class="theme-title">ORNITHOLOGY</p>
            <p>
                Inspired by the Charlie Parker tune of the same name and our shared love of birds,
                feel free to interpret this as <em>loosely</em> or as <em>specifically</em> as you like!
            </p>
        </section>

        <!-- FOOD & DRINK -->
        <section class="info-section">
            <h2>🍽️ Food &amp; Drink</h2>
            <p>
                There will be a <strong>buffet style selection</strong> with options for coeliacs,
                vegetarians and vegans alike! If you have any worries or questions about food,
                just reach out to one of us any time.
            </p>
            <p>We will have a selection of <strong>soft and alcoholic drinks</strong> available!</p>
        </section>

        <!-- GIFTS -->
        <section class="info-section">
            <h2>🎁 Gifts</h2>
            <p>
                We have <strong>no expectation of gifts</strong> — we are just excited to spend the day
                with you all! But if you do feel compelled to give us something, we would really appreciate
                gifts of any size towards our <strong>honeymoon in Japan</strong>!
            </p>
            <p>You can send directly to our account:</p>
            <a
                class="gift-link"
                href="https://monzo.me/jemimaswan?h=qG89fK"
                target="_blank"
                rel="noopener noreferrer"
            >
                monzo.me/jemimaswan →
            </a>
            <p class="note">Or give us cash on the day — both are equally wonderful! 🎌</p>
        </section>

        <!-- NOTE FOR DRIVERS -->
        <section class="info-section">
            <h2>🚗 Note for Drivers</h2>
            <p>
                If you will be driving to the reception in your own car, please keep in mind that the route
                to the venue passes through a <strong>restricted zone</strong> where a code is required to avoid a fine.
            </p>
            <div class="driver-box">
                <p>Install the <strong>RingGo app</strong> and use parking code:</p>
                <p class="parking-code">17471</p>
                <p class="parking-note">
                    It will appear as <em>"PG Hitchcock &amp; King (CLEAN acc)"</em>.
                    This code gives you access to the LTN in which the venue exists.
                </p>
                <p class="warning-note">
                    ⚠️ This code grants <strong>access to the zone only</strong> — it does not allow you to park.
                    You will need to pay for parking separately.
                </p>
            </div>
        </section>

    </main>

    <footer class="info-footer">
        <p>Any questions? Just reach out — we can't wait to celebrate with you! 🐦</p>
        <a href="/" class="back-btn-bottom">← Back to home</a>
    </footer>
</div>

<style>
    .info-page {
        min-height: 100vh;
        background: var(--color-white);
        padding: var(--spacing-xl) var(--spacing-md);
        max-width: 720px;
        margin: 0 auto;
    }

    .back-btn {
        display: inline-block;
        font-family: var(--font-mimko);
        font-size: 0.95rem;
        padding: var(--spacing-sm) var(--spacing-md);
        border: 1px solid var(--color-border);
        border-radius: var(--radius-md);
        background: rgba(255, 255, 255, 0.9);
        color: var(--color-text);
        text-decoration: none;
        transition: all var(--transition-normal);
        margin-bottom: var(--spacing-xl);
    }

    .back-btn:hover { background: var(--color-text); color: var(--color-white); opacity: 1; }

    .info-header {
        text-align: center;
        margin-bottom: var(--spacing-2xl);
        padding-bottom: var(--spacing-xl);
        border-bottom: 2px solid var(--color-border);
    }

    .logo { width: 140px; height: auto; margin-bottom: var(--spacing-md); }

    .info-header h1 {
        font-family: var(--font-diplomata);
        font-size: 2.5rem;
        color: var(--color-text);
        margin: 0;
    }

    .info-content { display: flex; flex-direction: column; gap: var(--spacing-2xl); }

    .info-section {
        border-bottom: 1px solid var(--color-border-light);
        padding-bottom: var(--spacing-2xl);
    }

    .info-section:last-child { border-bottom: none; }

    .info-section h2 {
        font-family: var(--font-diplomata);
        font-size: 1.6rem;
        color: var(--color-text);
        margin-bottom: var(--spacing-lg);
        letter-spacing: 0.02em;
    }

    .info-section p {
        font-family: var(--font-mimko);
        font-size: 1rem;
        line-height: 1.7;
        color: var(--color-text);
        margin-bottom: var(--spacing-md);
    }

    .info-section p:last-child { margin-bottom: 0; }

    .sub-section {
        margin-bottom: var(--spacing-xl);
        padding: var(--spacing-lg);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
    }

    .sub-section:last-child { margin-bottom: 0; }

    .sub-section h3 {
        font-family: var(--font-mimko);
        font-size: 0.8rem;
        text-transform: uppercase;
        letter-spacing: 0.1em;
        color: var(--color-text-light);
        margin-bottom: var(--spacing-sm);
    }

    .venue-name { font-size: 1.2rem; font-weight: 700; margin-bottom: var(--spacing-xs) !important; }
    .venue-address { font-size: 0.95rem; color: var(--color-text-light); margin-bottom: var(--spacing-sm) !important; }
    .note { font-size: 0.9rem !important; color: var(--color-text-light); font-style: italic; }

    .map-link {
        display: inline-block;
        font-family: var(--font-mimko);
        font-size: 0.95rem;
        color: var(--color-text);
        text-decoration: underline;
        text-underline-offset: 3px;
        margin-top: var(--spacing-sm);
        transition: opacity var(--transition-fast);
    }

    .map-link:hover { opacity: 0.6; }

    .theme-title {
        font-family: var(--font-diplomata);
        font-size: 2rem !important;
        letter-spacing: 0.1em;
        margin-bottom: var(--spacing-md) !important;
    }

    .gift-link {
        display: block;
        font-family: var(--font-mimko);
        font-size: 1.1rem;
        font-weight: 700;
        text-align: center;
        padding: var(--spacing-md);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        margin: var(--spacing-md) 0;
        text-decoration: none;
        color: var(--color-text);
        transition: all var(--transition-normal);
    }

    .gift-link:hover { background: var(--color-text); color: var(--color-white); opacity: 1; }

    .driver-box {
        background: var(--color-background-alt);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        padding: var(--spacing-lg);
        margin-top: var(--spacing-md);
    }

    .driver-box p { margin-bottom: var(--spacing-sm) !important; }

    .parking-code {
        font-family: var(--font-diplomata);
        font-size: 2.5rem !important;
        text-align: center;
        letter-spacing: 0.15em;
        margin: var(--spacing-sm) 0 !important;
    }

    .parking-note { font-size: 0.9rem !important; color: var(--color-text-light); }

    .warning-note {
        font-size: 0.9rem !important;
        border-top: 1px solid var(--color-border-light);
        padding-top: var(--spacing-sm);
        margin-top: var(--spacing-sm) !important;
    }

    .info-footer {
        text-align: center;
        margin-top: var(--spacing-2xl);
        padding-top: var(--spacing-xl);
        border-top: 2px solid var(--color-border);
    }

    .info-footer p { font-family: var(--font-mimko); font-size: 1rem; margin-bottom: var(--spacing-lg); }

    .back-btn-bottom {
        display: inline-block;
        font-family: var(--font-mimko);
        font-size: 0.95rem;
        padding: var(--spacing-sm) var(--spacing-lg);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        background: var(--color-white);
        color: var(--color-text);
        text-decoration: none;
        transition: all var(--transition-normal);
    }

    .back-btn-bottom:hover { background: var(--color-text); color: var(--color-white); opacity: 1; }

    @media (max-width: 640px) {
        .info-page { padding: var(--spacing-md); }
        .info-header h1 { font-size: 2rem; }
        .info-section h2 { font-size: 1.3rem; }
        .theme-title { font-size: 1.6rem !important; }
        .parking-code { font-size: 2rem !important; }
        .logo { width: 100px; }
    }
</style>

