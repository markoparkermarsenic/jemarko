<script lang="ts">
    import { AvatarPlaza, AvatarSelection } from "$lib/components";

    type AppView = "rsvp" | "avatar-selection" | "plaza-only";

    let currentView = $state<AppView>("rsvp");
    let rsvpGuests = $state<string[]>([]);
    let guestName = $state("");

    // Check if RSVP deadline has passed (August 1st, 2026)
    const rsvpDeadline = new Date("2026-08-01T00:00:00");
    const isRsvpClosed = $derived(new Date() >= rsvpDeadline);

    function handleRSVPComplete() {
        // For demo, use some sample guest names
        // In production, this would come from the RSVP form
        rsvpGuests = ["John Smith", "Jane Smith"];
        currentView = "avatar-selection";
    }

    function handleAvatarComplete() {
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
    <!-- Avatar Plaza is always visible in the background -->
    <AvatarPlaza />

    <!-- Content overlay -->
    <div class="content-overlay">
        {#if currentView === "rsvp" && !isRsvpClosed}
            <div class="content-wrapper animate-fadeIn">
                <div class="invite-container">
                    <img
                        src="/invite-with-rsvp-box.png"
                        alt="Wedding Invitation - RSVP"
                        class="invite-image"
                    />
                    <input
                        type="text"
                        bind:value={guestName}
                        placeholder="Enter your name"
                        class="name-input-overlay"
                    />
                    <div class="rsvp-buttons">
                        <button
                            class="rsvp-btn yes-btn"
                            on:click={handleRSVPComplete}
                        >
                            Yes
                        </button>
                        <button class="rsvp-btn no-btn" on:click={() => {}}>
                            No
                        </button>
                    </div>
                </div>
            </div>
        {:else if currentView === "avatar-selection"}
            <div class="content-wrapper animate-fadeIn">
                <AvatarSelection
                    guests={rsvpGuests}
                    oncomplete={handleAvatarComplete}
                />
            </div>
        {:else if currentView === "plaza-only" || isRsvpClosed}
            <div class="plaza-view animate-fadeIn">
                <header class="plaza-header">
                    <h1>Jemarko</h1>
                    {#if isRsvpClosed}
                        <p class="closed-message">
                            RSVP is now closed. Thank you to all our guests!
                        </p>
                    {:else}
                        <p class="subtitle">Welcome to the guest plaza!</p>
                    {/if}
                </header>
            </div>
        {/if}
    </div>
</div>

<style>
    .page-container {
        position: relative;
        min-height: 100vh;
        width: 100%;
        background: var(--color-white);
    }

    .content-overlay {
        position: relative;
        z-index: 10;
        min-height: 100vh;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: var(--spacing-lg);
    }

    .content-wrapper {
        width: 100%;
        max-width: 500px;
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
        background: var(--color-white);
        padding: var(--spacing-xl) var(--spacing-2xl);
        border-radius: var(--radius-lg);
        border: 2px solid var(--color-border);
    }

    .plaza-header h1 {
        font-size: 3rem;
        color: var(--color-text);
        margin-bottom: var(--spacing-sm);
    }

    .closed-message {
        color: var(--color-text-light);
        font-size: 1rem;
        font-family: var(--font-mimko);
    }

    @media (max-width: 640px) {
        .content-overlay {
            padding: var(--spacing-md);
        }

        .page-header h1 {
            font-size: 3rem;
        }

        .subtitle {
            font-size: 1.25rem;
        }

        .plaza-header h1 {
            font-size: 2.5rem;
        }
    }

    @media (max-width: 400px) {
        .page-header h1 {
            font-size: 2.5rem;
        }

        .page-header {
            padding: var(--spacing-lg);
        }
    }
</style>
