<script lang="ts">
    type RSVPStep = "name" | "email" | "guests" | "complete";

    interface Guest {
        id: string;
        name: string;
        attending: boolean;
    }

    let step = $state<RSVPStep>("name");
    let nameInput = $state("");
    let emailInput = $state("");
    let isVerified = $state(false);
    let isLoading = $state(false);
    let errorMessage = $state("");
    let guests = $state<Guest[]>([]);

    // Demo guest list - in production this would come from the backend
    const guestList = [
        {
            id: "1",
            name: "John Smith",
            allowedGuests: ["John Smith", "Jane Smith"],
        },
        {
            id: "2",
            name: "Jane Smith",
            allowedGuests: ["John Smith", "Jane Smith"],
        },
        { id: "3", name: "Bob Johnson", allowedGuests: ["Bob Johnson"] },
        {
            id: "4",
            name: "Alice Williams",
            allowedGuests: ["Alice Williams", "Tom Williams"],
        },
    ];

    async function verifyName() {
        isLoading = true;
        errorMessage = "";

        // Simulate API call
        await new Promise((resolve) => setTimeout(resolve, 800));

        const foundGuest = guestList.find(
            (g) => g.name.toLowerCase() === nameInput.toLowerCase(),
        );

        if (foundGuest) {
            isVerified = true;
            guests = foundGuest.allowedGuests.map((name, index) => ({
                id: `guest-${index}`,
                name,
                attending: true,
            }));
            step = "email";
        } else {
            errorMessage =
                "Name not found on the guest list. Please check the spelling or contact us.";
        }

        isLoading = false;
    }

    async function submitEmail() {
        isLoading = true;
        errorMessage = "";

        // Basic email validation
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(emailInput)) {
            errorMessage = "Please enter a valid email address.";
            isLoading = false;
            return;
        }

        await new Promise((resolve) => setTimeout(resolve, 500));
        step = "guests";
        isLoading = false;
    }

    function toggleGuest(guestId: string) {
        guests = guests.map((g) =>
            g.id === guestId ? { ...g, attending: !g.attending } : g,
        );
    }

    async function submitRSVP() {
        isLoading = true;
        errorMessage = "";

        // Simulate API call
        await new Promise((resolve) => setTimeout(resolve, 1000));

        // In production, this would send data to the backend
        console.log("RSVP submitted:", {
            email: emailInput,
            guests: guests.filter((g) => g.attending),
        });

        step = "complete";
        isLoading = false;
    }

    function handleKeyPress(event: KeyboardEvent, action: () => void) {
        if (event.key === "Enter") {
            action();
        }
    }

    // Event to notify parent that RSVP is complete
    let { oncomplete }: { oncomplete?: () => void } = $props();
</script>

<div class="rsvp-form card animate-fadeIn">
    <div class="form-header">
        <img src="/border.png" alt="" class="border-decoration" />
        <h2>RSVP</h2>
    </div>

    {#if step === "name"}
        <div class="form-step">
            <p class="instruction">
                Please enter your name as it appears on the invitation
            </p>
            <div class="input-group">
                <label for="name">Your Name</label>
                <input
                    type="text"
                    id="name"
                    bind:value={nameInput}
                    placeholder="Enter your full name"
                    onkeypress={(e) => handleKeyPress(e, verifyName)}
                    disabled={isLoading}
                />
            </div>
            {#if errorMessage}
                <p class="error">{errorMessage}</p>
            {/if}
            <button
                class="btn btn-primary"
                onclick={verifyName}
                disabled={isLoading || !nameInput.trim()}
            >
                {#if isLoading}
                    Checking...
                {:else}
                    <span>Continue</span>
                    <img src="/arrow-r-1.png" alt="" class="btn-icon-img" />
                {/if}
            </button>
        </div>
    {:else if step === "email"}
        <div class="form-step">
            <p class="instruction">
                Great! We found you on the guest list. Please enter your email
                for confirmation.
            </p>
            <div class="input-group">
                <label for="email">Email Address</label>
                <input
                    type="email"
                    id="email"
                    bind:value={emailInput}
                    placeholder="your@email.com"
                    onkeypress={(e) => handleKeyPress(e, submitEmail)}
                    disabled={isLoading}
                />
            </div>
            {#if errorMessage}
                <p class="error">{errorMessage}</p>
            {/if}
            <div class="button-group">
                <button
                    class="btn btn-outline"
                    onclick={() => (step = "name")}
                    disabled={isLoading}
                >
                    Back
                </button>
                <button
                    class="btn btn-primary"
                    onclick={submitEmail}
                    disabled={isLoading || !emailInput.trim()}
                >
                    {#if isLoading}
                        Verifying...
                    {:else}
                        <span>Continue</span>
                        <img src="/arrow-r-1.png" alt="" class="btn-icon-img" />
                    {/if}
                </button>
            </div>
        </div>
    {:else if step === "guests"}
        <div class="form-step">
            <p class="instruction">Who will be attending?</p>
            <div class="guests-list">
                {#each guests as guest (guest.id)}
                    <label class="guest-item" class:attending={guest.attending}>
                        <input
                            type="checkbox"
                            checked={guest.attending}
                            onchange={() => toggleGuest(guest.id)}
                        />
                        <span class="guest-name">{guest.name}</span>
                        <span class="status">
                            {#if guest.attending}
                                <img
                                    src="/confirm.png"
                                    alt="Attending"
                                    class="status-icon"
                                />
                            {:else}
                                <img
                                    src="/cancel.png"
                                    alt="Not attending"
                                    class="status-icon"
                                />
                            {/if}
                        </span>
                    </label>
                {/each}
            </div>
            {#if errorMessage}
                <p class="error">{errorMessage}</p>
            {/if}
            <div class="button-group">
                <button
                    class="btn btn-outline"
                    onclick={() => (step = "email")}
                    disabled={isLoading}
                >
                    Back
                </button>
                <button
                    class="btn btn-primary"
                    onclick={submitRSVP}
                    disabled={isLoading}
                >
                    {#if isLoading}
                        Submitting...
                    {:else}
                        <span>Submit RSVP</span>
                        <img src="/confirm.png" alt="" class="btn-icon-img" />
                    {/if}
                </button>
            </div>
        </div>
    {:else if step === "complete"}
        <div class="form-step complete">
            <div class="success-icon">
                <img src="/confirm.png" alt="Success" />
            </div>
            <h3>Thank You!</h3>
            <p>Your RSVP has been submitted successfully.</p>
            <p class="sub-text">
                A confirmation email will be sent to {emailInput}
            </p>
            <button class="btn btn-primary" onclick={() => oncomplete?.()}>
                <span>Choose Your Avatar</span>
                <img src="/next-arrow.png" alt="" class="btn-icon-img" />
            </button>
        </div>
    {/if}

    <div class="progress-indicator">
        <span
            class="dot"
            class:active={step === "name"}
            class:completed={step !== "name"}
        ></span>
        <span
            class="dot"
            class:active={step === "email"}
            class:completed={step === "guests" || step === "complete"}
        ></span>
        <span
            class="dot"
            class:active={step === "guests"}
            class:completed={step === "complete"}
        ></span>
        <span class="dot" class:active={step === "complete"}></span>
    </div>
</div>

<style>
    .rsvp-form {
        max-width: 420px;
        width: 100%;
        margin: 0 auto;
        position: relative;
        z-index: 100;
        background: var(--color-white);
        border: 2px solid var(--color-border);
    }

    .form-header {
        text-align: center;
        margin-bottom: var(--spacing-xl);
        position: relative;
    }

    .form-header h2 {
        font-size: 2.5rem;
        color: var(--color-text);
    }

    .border-decoration {
        position: absolute;
        top: -20px;
        left: 50%;
        transform: translateX(-50%);
        width: 80px;
        height: auto;
        opacity: 0.8;
    }

    .form-step {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-md);
    }

    .instruction {
        text-align: center;
        color: var(--color-text-light);
        margin-bottom: var(--spacing-sm);
    }

    .input-group {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-xs);
    }

    .error {
        color: var(--color-text);
        font-size: 0.875rem;
        text-align: center;
        font-weight: 600;
    }

    .button-group {
        display: flex;
        gap: var(--spacing-md);
        justify-content: center;
        margin-top: var(--spacing-md);
    }

    .btn-icon-img {
        width: 16px;
        height: 16px;
        object-fit: contain;
    }

    .guests-list {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-sm);
    }

    .guest-item {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-md);
        background: var(--color-white);
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: all var(--transition-fast);
        border: 2px solid var(--color-border-light);
    }

    .guest-item:hover {
        border-color: var(--color-border);
    }

    .guest-item.attending {
        border-color: var(--color-border);
        background: var(--color-background-alt);
    }

    .guest-item input[type="checkbox"] {
        width: 20px;
        height: 20px;
        cursor: pointer;
        accent-color: var(--color-text);
    }

    .guest-name {
        flex: 1;
        font-weight: 500;
    }

    .status-icon {
        width: 24px;
        height: 24px;
        object-fit: contain;
    }

    .complete {
        text-align: center;
        padding: var(--spacing-xl) 0;
    }

    .success-icon img {
        width: 64px;
        height: 64px;
        margin-bottom: var(--spacing-md);
    }

    .complete h3 {
        font-size: 2rem;
        color: var(--color-text);
        margin-bottom: var(--spacing-sm);
    }

    .sub-text {
        font-size: 0.875rem;
        color: var(--color-text-light);
    }

    .progress-indicator {
        display: flex;
        justify-content: center;
        gap: var(--spacing-sm);
        margin-top: var(--spacing-xl);
    }

    .dot {
        width: 10px;
        height: 10px;
        border-radius: 50%;
        background: var(--color-border-light);
        border: 1px solid var(--color-border);
        transition: all var(--transition-normal);
    }

    .dot.active {
        background: var(--color-text);
        transform: scale(1.2);
    }

    .dot.completed {
        background: var(--color-text);
    }

    @media (max-width: 480px) {
        .rsvp-form {
            padding: var(--spacing-lg);
        }

        .button-group {
            flex-direction: column;
        }

        .btn {
            width: 100%;
        }
    }
</style>
