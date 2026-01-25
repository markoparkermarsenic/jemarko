<script lang="ts">
    import { verifyName, submitRSVP } from '$lib/api';
    import type { RSVPRequest } from '$lib/api';

    type RSVPStep = "name" | "email" | "attending" | "guests" | "complete";

    interface Guest {
        id: string;
        name: string;
        attending: boolean;
    }

    let step = $state<RSVPStep>("name");
    let nameInput = $state("");
    let emailInput = $state("");
    let dietaryRequirements = $state("");
    let isAttending = $state(true);
    let isVerified = $state(false);
    let isLoading = $state(false);
    let errorMessage = $state("");
    let guests = $state<Guest[]>([]);
    let verifiedName = $state(""); // Store the verified name

    async function handleVerifyName() {
        isLoading = true;
        errorMessage = "";

        try {
            const response = await verifyName(nameInput.trim());
            
            if (response.success) {
                isVerified = true;
                verifiedName = nameInput.trim();
                step = "email";
            } else {
                errorMessage = response.message || "Name not found on the guest list. Please check the spelling or contact us.";
            }
        } catch (error) {
            errorMessage = "Unable to verify name. Please try again.";
            console.error("Verify name error:", error);
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

        step = "attending";
        isLoading = false;
    }

    function selectAttending(attending: boolean) {
        isAttending = attending;
        if (attending) {
            // If attending, ask for guest selection
            // Initialize with the verified name as the first guest
            guests = [
                {
                    id: 'guest-0',
                    name: verifiedName,
                    attending: true
                }
            ];
            step = "guests";
        } else {
            // If not attending, skip directly to submission
            handleSubmitRSVP();
        }
    }

    function addGuest() {
        const newGuestName = prompt("Enter guest name:");
        if (newGuestName && newGuestName.trim()) {
            guests = [
                ...guests,
                {
                    id: `guest-${guests.length}`,
                    name: newGuestName.trim(),
                    attending: true
                }
            ];
        }
    }

    function removeGuest(guestId: string) {
        // Don't allow removing the first guest (the person who initiated RSVP)
        if (guestId === 'guest-0') {
            return;
        }
        guests = guests.filter((g) => g.id !== guestId);
    }

    function toggleGuest(guestId: string) {
        guests = guests.map((g) =>
            g.id === guestId ? { ...g, attending: !g.attending } : g,
        );
    }

    async function handleSubmitRSVP() {
        isLoading = true;
        errorMessage = "";

        try {
            const attendingGuestNames = guests
                .filter((g) => g.attending)
                .map((g) => g.name);

            const rsvpData: RSVPRequest = {
                name: verifiedName,
                email: emailInput.trim(),
                isAttending,
                attendingGuests: isAttending ? attendingGuestNames : [],
                diet: isAttending && dietaryRequirements.trim() ? dietaryRequirements.trim() : undefined,
            };

            const response = await submitRSVP(rsvpData);

            if (response.success) {
                step = "complete";
            } else {
                errorMessage = response.message || "Failed to submit RSVP. Please try again.";
            }
        } catch (error: any) {
            errorMessage = error.message || "Unable to submit RSVP. Please try again.";
            console.error("Submit RSVP error:", error);
        }

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
                    onkeypress={(e) => handleKeyPress(e, handleVerifyName)}
                    disabled={isLoading}
                />
            </div>
            {#if errorMessage}
                <p class="error">{errorMessage}</p>
            {/if}
            <button
                class="btn btn-primary"
                onclick={handleVerifyName}
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
    {:else if step === "attending"}
        <div class="form-step">
            <p class="instruction">Will you be attending?</p>
            <div class="attending-choice">
                <button
                    class="choice-btn"
                    onclick={() => selectAttending(true)}
                    disabled={isLoading}
                >
                    <img src="/confirm.png" alt="Yes" class="choice-icon" />
                    <span>Yes, I'll be there!</span>
                </button>
                <button
                    class="choice-btn"
                    onclick={() => selectAttending(false)}
                    disabled={isLoading}
                >
                    <img src="/cancel.png" alt="No" class="choice-icon" />
                    <span>Sorry, can't make it</span>
                </button>
            </div>
            {#if errorMessage}
                <p class="error">{errorMessage}</p>
            {/if}
            <button
                class="btn btn-outline"
                onclick={() => (step = "email")}
                disabled={isLoading}
            >
                Back
            </button>
        </div>
    {:else if step === "guests"}
        <div class="form-step">
            <p class="instruction">Who will be attending?</p>
            <div class="guests-list">
                {#each guests as guest (guest.id)}
                    <div class="guest-item" class:attending={guest.attending}>
                        <label class="guest-checkbox">
                            <input
                                type="checkbox"
                                checked={guest.attending}
                                onchange={() => toggleGuest(guest.id)}
                            />
                            <span class="guest-name">{guest.name}</span>
                        </label>
                        {#if guest.id !== 'guest-0'}
                            <button
                                class="btn-remove"
                                onclick={() => removeGuest(guest.id)}
                                title="Remove guest"
                            >
                                ×
                            </button>
                        {/if}
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
                    </div>
                {/each}
            </div>
            <button class="btn btn-outline" onclick={addGuest} type="button">
                + Add Another Guest
            </button>

            <div class="input-group">
                <label for="diet">Dietary Requirements (Optional)</label>
                <textarea
                    id="diet"
                    bind:value={dietaryRequirements}
                    placeholder="Any allergies or dietary restrictions?"
                    rows="3"
                    disabled={isLoading}
                ></textarea>
                <p class="helper-text">Let us know about any dietary needs for your party</p>
            </div>

            {#if errorMessage}
                <p class="error">{errorMessage}</p>
            {/if}
            <div class="button-group">
                <button
                    class="btn btn-outline"
                    onclick={() => (step = "attending")}
                    disabled={isLoading}
                >
                    Back
                </button>
                <button
                    class="btn btn-primary"
                    onclick={handleSubmitRSVP}
                    disabled={isLoading || guests.filter(g => g.attending).length === 0}
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
            class:completed={step === "attending" || step === "guests" || step === "complete"}
        ></span>
        <span
            class="dot"
            class:active={step === "attending"}
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

    .input-group textarea {
        width: 100%;
        padding: var(--spacing-sm);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-sm);
        font-family: inherit;
        font-size: 1rem;
        resize: vertical;
    }

    .helper-text {
        font-size: 0.75rem;
        color: var(--color-text-light);
        margin-top: -4px;
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

    .attending-choice {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-md);
    }

    .choice-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: var(--spacing-md);
        padding: var(--spacing-lg);
        background: var(--color-white);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: all var(--transition-fast);
        font-size: 1.1rem;
        font-weight: 500;
    }

    .choice-btn:hover {
        border-color: var(--color-text);
        background: var(--color-background-alt);
    }

    .choice-icon {
        width: 32px;
        height: 32px;
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
        transition: all var(--transition-fast);
        border: 2px solid var(--color-border-light);
    }

    .guest-item.attending {
        border-color: var(--color-border);
        background: var(--color-background-alt);
    }

    .guest-checkbox {
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        flex: 1;
        cursor: pointer;
    }

    .guest-checkbox input[type="checkbox"] {
        width: 20px;
        height: 20px;
        cursor: pointer;
        accent-color: var(--color-text);
    }

    .guest-name {
        font-weight: 500;
    }

    .btn-remove {
        background: none;
        border: none;
        font-size: 1.5rem;
        color: var(--color-text-light);
        cursor: pointer;
        padding: 0;
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: color var(--transition-fast);
    }

    .btn-remove:hover {
        color: var(--color-text);
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
