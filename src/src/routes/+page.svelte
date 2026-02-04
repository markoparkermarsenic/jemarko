<script lang="ts">
    import { AvatarPlaza, AvatarSelection, RSVPForm } from "$lib/components";

    type AppView = "rsvp" | "avatar-selection" | "plaza-only";

    let currentView = $state<AppView>("rsvp");
    let rsvpGuests = $state<string[]>([]);
    let rsvpEmail = $state<string>("");
    let avatarRefreshTrigger = $state(1); // Start at 1 to trigger initial fetch

    // Check if RSVP deadline has passed (August 1st, 2026)
    const rsvpDeadline = new Date("2026-08-01T00:00:00");
    const isRsvpClosed = $derived(new Date() >= rsvpDeadline);

    function handleRSVPComplete(guests: string[], email: string) {
        // This will be called after successful RSVP submission
        // Guest names and email will be set before this is called
        rsvpGuests = guests;
        rsvpEmail = email;
        currentView = "avatar-selection";
    }

    function handleAvatarComplete() {
        // Increment trigger to refresh avatars with newly saved data
        avatarRefreshTrigger++;
        currentView = "plaza-only";
    }

    function goToPlaza() {
        currentView = "plaza-only";
    }

    // If RSVP is closed, show only the plaza
    $effect(() => {
        if (isRsvpClosed) {
            currentView = "plaza-only";
        }
    });
</script>

<div class="page-container">
    <!-- Avatar Plaza persists across all views - never recreated -->
    <AvatarPlaza refreshTrigger={avatarRefreshTrigger} />
    
    <!-- Skip to Plaza button - visible when not in plaza view -->
    {#if currentView !== "plaza-only" && !isRsvpClosed}
        <button class="skip-to-plaza-btn" onclick={goToPlaza}>
            Skip to Guest Plaza →
        </button>
    {/if}
    
    {#if currentView === "rsvp" && !isRsvpClosed}
        <div class="content-overlay">
            <div class="container">
                <RSVPForm oncomplete={handleRSVPComplete} />
            </div>
        </div>
    {:else if currentView === "avatar-selection"}
        <div class="content-overlay">
            <div class="container">
                <AvatarSelection
                    guests={rsvpGuests}
                    email={rsvpEmail}
                    oncomplete={handleAvatarComplete}
                />
            </div>
        </div>
    {:else if currentView === "plaza-only" || isRsvpClosed}
        <div class="content-overlay">
            <div class="plaza-view animate-fadeIn">
                <!-- SVG filter for fuzzy/hand-drawn effect -->
                <svg class="svg-filters" aria-hidden="true">
                    <defs>
                        <filter id="fuzzy-header">
                            <feTurbulence type="fractalNoise" baseFrequency="0.03" numOctaves="3" result="noise" />
                            <feDisplacementMap in="SourceGraphic" in2="noise" scale="2" xChannelSelector="R" yChannelSelector="G" />
                        </filter>
                    </defs>
                </svg>
                <header class="plaza-header fuzzy-border">
                    <img src="/j_and_m.png" alt="Jemarko" class="plaza-logo" />
                    {#if isRsvpClosed}
                        <p class="closed-message">
                            RSVP is now closed. Thank you to all our guests!
                        </p>
                    {:else}
                        <p class="subtitle">Welcome to the guest plaza!</p>
                    {/if}
                </header>
            </div>
        </div>
    {/if}
</div>

<style>
    .svg-filters {
        position: absolute;
        width: 0;
        height: 0;
        overflow: hidden;
    }

    .page-container {
        position: relative;
        min-height: 100vh;
        width: 100%;
        background: var(--color-white);
    }

    .skip-to-plaza-btn {
        position: fixed;
        top: var(--spacing-md);
        right: var(--spacing-md);
        z-index: 100;
        font-family: var(--font-mimko);
        font-size: 0.9rem;
        padding: var(--spacing-sm) var(--spacing-md);
        border: 1px solid var(--color-border);
        background: rgba(255, 255, 255, 0.9);
        color: var(--color-text);
        cursor: pointer;
        transition: all var(--transition-normal);
        border-radius: var(--radius-md);
        backdrop-filter: blur(4px);
    }

    .skip-to-plaza-btn:hover {
        background-color: var(--color-text);
        color: var(--color-white);
        transform: translateX(4px);
    }

    .content-overlay {
        position: relative;
        z-index: 10;
        min-height: 100vh;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: var(--spacing-xl);
        pointer-events: none; /* Allow clicks to pass through to avatars */
    }

    .container {
        width: 100%;
        max-width: 600px;
        pointer-events: auto; /* But form content should be clickable */
    }

    .plaza-view {
        text-align: center;
        pointer-events: none; /* Allow clicks to pass through to avatars */
    }

    .page-header {
        text-align: center;
        margin-bottom: var(--spacing-2xl);
        padding: var(--spacing-xl);
        background: var(--color-white);
        border-radius: var(--radius-lg);
        border: 2px solid var(--color-border);
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--spacing-md);
    }

    .invite-container {
        position: relative;
        text-align: center;
        margin-bottom: var(--spacing-2xl);
        display: flex;
        justify-content: center;
    }

    .invite-image {
        max-width: 100%;
        height: auto;
        display: block;
    }

    .name-input-overlay {
        position: absolute;
        /* Positioned based on coordinates: x1:268, x2:1388, y1:1278, y2:1392 */
        /* Adjusted for actual image proportions */
        left: 16.18%;
        top: 52%;
        width: 67.63%;
        height: 5%;
        padding: 0;
        font-size: 1.2rem;
        font-family: var(--font-mimko);
        border: none;
        background: transparent;
        text-align: center;
        outline: none;
        box-sizing: border-box;
    }

    .name-input-overlay::placeholder {
        color: rgba(0, 0, 0, 0.4);
    }

    .rsvp-buttons {
        position: absolute;
        top: 62%;
        left: 50%;
        transform: translateX(-50%);
        display: flex;
        gap: var(--spacing-lg);
        z-index: 20;
    }

    .rsvp-btn {
        font-family: var(--font-mimko);
        font-size: 1.5rem;
        padding: var(--spacing-sm) var(--spacing-xl);
        border: 2px solid var(--color-border);
        background: var(--color-white);
        color: var(--color-text);
        cursor: pointer;
        transition: all var(--transition-normal);
        border-radius: var(--radius-md);
    }

    .rsvp-btn:hover {
        background-color: var(--color-text);
        color: var(--color-white);
    }

    .yes-btn:hover {
        background-color: #000000;
        color: #ffffff;
    }

    .no-btn:hover {
        background-color: #000000;
        color: #ffffff;
    }

    .plaza-view {
        text-align: center;
    }

    .plaza-header {
        padding: var(--spacing-xl) var(--spacing-2xl);
        border-radius: 20px;
        position: relative;
        z-index: 0; /* Behind avatars */
        border: none;
        background: transparent;
        box-shadow: none;
        opacity: 0.4; /* Reduced opacity */
    }

    .plaza-header h1 {
        font-size: 3rem;
        color: var(--color-text);
        margin-bottom: var(--spacing-sm);
    }

    .plaza-logo {
        max-width: 300px;
        height: auto;
        margin-bottom: var(--spacing-sm);
    }

    .subtitle {
        color: var(--color-text);
        font-size: 1.1rem;
        font-family: var(--font-mimko);
        margin: 0;
    }

    .closed-message {
        color: var(--color-text-light);
        font-size: 1rem;
        font-family: var(--font-mimko);
        margin: 0;
    }

    @media (max-width: 640px) {
        .content-overlay {
            padding: var(--spacing-md);
        }

        .container {
            max-width: 100%;
        }

        .plaza-header {
            padding: var(--spacing-sm) var(--spacing-md);
        }

        .plaza-header h1 {
            font-size: 2.5rem;
        }

        .plaza-logo {
            max-width: 75px;
        }
    }

    @media (max-width: 400px) {
        .page-header h1 {
            font-size: 2.5rem;
        }

        .page-header {
            padding: var(--spacing-lg);
        }

        .plaza-header {
            padding: var(--spacing-sm) var(--spacing-md);
        }

        .plaza-logo {
            max-width: 75px;
        }
    }
</style>
