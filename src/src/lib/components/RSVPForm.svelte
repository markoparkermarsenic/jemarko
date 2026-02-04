<script lang="ts">
    import { verifyName, submitRSVP } from '$lib/api';
    import type { RSVPRequest, FamilyMember } from '$lib/api';

    type RSVPStep = "initial" | "family-selection" | "complete";

    interface GuestSelection {
        id: string;
        name: string;
        isAttending: boolean;
    }

    let step = $state<RSVPStep>("initial");
    let nameInput = $state("");
    let emailInput = $state("");
    let dietaryRequirements = $state("");
    let familyMembers = $state<GuestSelection[]>([]);
    let isLoading = $state(false);
    let errorMessage = $state("");

    async function handleVerifyAndContinue() {
        errorMessage = "";

        // Validation
        if (!nameInput.trim()) {
            errorMessage = "Please enter your name.";
            return;
        }

        if (!emailInput.trim()) {
            errorMessage = "Please enter your email.";
            return;
        }

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(emailInput)) {
            errorMessage = "Please enter a valid email address.";
            return;
        }

        isLoading = true;

        try {
            const response = await verifyName(nameInput.trim(), emailInput.trim());
            
            if (response.success && response.familyMembers) {
                // Convert family members to guest selections with all initially marked as attending
                familyMembers = response.familyMembers.map(member => ({
                    id: member.id,
                    name: member.name,
                    isAttending: true
                }));
                
                step = "family-selection";
            } else {
                errorMessage = response.message || "Name not found on the guest list. Please check the spelling or contact us.";
            }
        } catch (error: any) {
            errorMessage = error.message || "Unable to verify name. Please try again.";
            console.error("Verify name error:", error);
        }

        isLoading = false;
    }

    function toggleGuestAttending(guestId: string) {
        familyMembers = familyMembers.map(member =>
            member.id === guestId 
                ? { ...member, isAttending: !member.isAttending }
                : member
        );
    }

    async function handleSubmitRSVP() {
        errorMessage = "";

        const attendingGuests = familyMembers
            .filter(member => member.isAttending)
            .map(member => member.name);

        if (attendingGuests.length === 0) {
            errorMessage = "Please select at least one person or mark everyone as not attending.";
            return;
        }

        isLoading = true;

        try {
            const rsvpData: RSVPRequest = {
                name: nameInput.trim(),
                email: emailInput.trim(),
                isAttending: attendingGuests.length > 0,
                attendingGuests: attendingGuests,
                diet: dietaryRequirements.trim() || undefined,
            };

            const response = await submitRSVP(rsvpData);

            if (response.success) {
                step = "complete";
                // Pass the attending guests and email to the parent
                oncomplete?.(attendingGuests, emailInput.trim());
            } else {
                errorMessage = response.message || "Failed to submit RSVP. Please try again.";
            }
        } catch (error: any) {
            errorMessage = error.message || "Unable to submit RSVP. Please try again.";
            console.error("Submit RSVP error:", error);
        }

        isLoading = false;
    }

    function handleBack() {
        step = "initial";
        familyMembers = [];
        errorMessage = "";
    }

    // Event to notify parent that RSVP is complete
    let { oncomplete }: { oncomplete?: (guests: string[], email: string) => void } = $props();
</script>

{#if step === "initial"}
    <div class="rsvp-form card animate-fadeIn">
        <div class="form-header">
            <h1>RSVP</h1>
            <p class="subtitle">Please enter your details to get started</p>
        </div>

        <div class="form-content">
            <div class="input-group">
                <label for="name">Your Name</label>
                <input
                    type="text"
                    id="name"
                    bind:value={nameInput}
                    placeholder="Enter your full name as on invitation"
                    disabled={isLoading}
                />
            </div>

            <div class="input-group">
                <label for="email">Email Address</label>
                <input
                    type="email"
                    id="email"
                    bind:value={emailInput}
                    placeholder="your@email.com"
                    disabled={isLoading}
                />
            </div>

            {#if errorMessage}
                <p class="error">{errorMessage}</p>
            {/if}

            <button
                class="btn btn-primary submit-btn"
                onclick={handleVerifyAndContinue}
                disabled={isLoading}
            >
                {#if isLoading}
                    Verifying...
                {:else}
                    <span>Continue</span>
                {/if}
            </button>
        </div>
    </div>
{:else if step === "family-selection"}
    <div class="rsvp-form card animate-fadeIn">
        <div class="form-header">
            <h2>RSVP</h2>
            <p class="subtitle">Please let us know who will be attending</p>
        </div>

        <div class="form-content">
            <div class="guests-section">
                <label class="section-label">Select who will be attending:</label>
                <div class="family-list">
                    {#each familyMembers as member (member.id)}
                        <div class="family-member-row">
                            <span class="member-name">{member.name}</span>
                            <div class="yes-no-buttons">
                                <button
                                    class="yes-no-btn yes-btn"
                                    class:selected={member.isAttending}
                                    onclick={() => {
                                        familyMembers = familyMembers.map(m =>
                                            m.id === member.id ? { ...m, isAttending: true } : m
                                        );
                                    }}
                                    type="button"
                                    disabled={isLoading}
                                >
                                    Yes
                                </button>
                                <button
                                    class="yes-no-btn no-btn"
                                    class:selected={!member.isAttending}
                                    onclick={() => {
                                        familyMembers = familyMembers.map(m =>
                                            m.id === member.id ? { ...m, isAttending: false } : m
                                        );
                                    }}
                                    type="button"
                                    disabled={isLoading}
                                >
                                    No
                                </button>
                            </div>
                        </div>
                    {/each}
                </div>
            </div>

            {#if familyMembers.some(m => m.isAttending)}
                <div class="input-group">
                    <label for="diet">Dietary Requirements (Optional)</label>
                    <textarea
                        id="diet"
                        bind:value={dietaryRequirements}
                        placeholder="Any allergies or dietary restrictions for your party?"
                        rows="3"
                        disabled={isLoading}
                    ></textarea>
                </div>
            {/if}

            {#if errorMessage}
                <p class="error">{errorMessage}</p>
            {/if}

            <div class="button-group">
                <button
                    class="btn btn-outline"
                    onclick={handleBack}
                    disabled={isLoading}
                >
                    Back
                </button>
                <button
                    class="btn btn-primary"
                    onclick={handleSubmitRSVP}
                    disabled={isLoading}
                >
                    {#if isLoading}
                        Submitting...
                    {:else}
                        <span>Submit RSVP</span>
                    {/if}
                </button>
            </div>
        </div>
    </div>
{:else if step === "complete"}
    <div class="rsvp-form card complete animate-fadeIn">
        <div class="success-content">
            <h3>Thank You!</h3>
            <p>Your RSVP has been submitted successfully.</p>
            <p class="sub-text">
                A confirmation email will be sent to {emailInput}
            </p>
            <button class="btn btn-primary" onclick={() => {
                const attendingGuests = familyMembers
                    .filter(member => member.isAttending)
                    .map(member => member.name);
                oncomplete?.(attendingGuests, emailInput.trim());
            }}>
                <span>Choose Your Avatar</span>
                <img src="/next-arrow.png" alt="" class="btn-icon-img" />
            </button>
        </div>
    </div>
{/if}

<style>
    .rsvp-form {
        max-width: 500px;
        width: 100%;
        margin: 0 auto;
    }

    .form-header {
        text-align: center;
        margin-bottom: var(--spacing-xl);
        position: relative;
    }

    .form-header h1 {
        font-size: 3rem;
        color: var(--color-text);
        margin-bottom: var(--spacing-sm);
        font-family: var(--font-diplomata);
    }

    .form-header h2 {
        font-size: 2.5rem;
        color: var(--color-text);
        margin-bottom: var(--spacing-sm);
    }

    .subtitle {
        color: var(--color-text-light);
        font-size: 1rem;
        margin: 0;
    }

    .border-decoration {
        position: absolute;
        top: -30px;
        left: 50%;
        transform: translateX(-50%);
        width: 80px;
        height: auto;
        opacity: 0.8;
    }

    .form-content {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-lg);
    }

    .input-group {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-xs);
    }

    .input-group textarea {
        width: 100%;
        padding: var(--spacing-sm) var(--spacing-md);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        font-family: inherit;
        font-size: 0.5rem;
        resize: vertical;
    }

    .section-label {
        font-size: 1rem;
        font-weight: 600;
        color: var(--color-text);
        margin-bottom: var(--spacing-sm);
    }

    .guests-section {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-sm);
    }

    .family-list {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-sm);
    }

    .family-member-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--spacing-md) var(--spacing-lg);
        background: var(--color-white);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        font-family: var(--font-mimko);
        font-size: 1.1rem;
    }

    .member-name {
        font-weight: 500;
        flex: 1;
    }

    .yes-no-buttons {
        display: flex;
        gap: var(--spacing-sm);
    }

    .yes-no-btn {
        padding: var(--spacing-xs) var(--spacing-md);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        background: var(--color-white);
        font-family: var(--font-mimko);
        font-size: 0.9rem;
        cursor: pointer;
        transition: all var(--transition-fast);
        min-width: 50px;
    }

    .yes-no-btn:hover {
        border-color: var(--color-text);
    }

    .yes-no-btn.selected {
        border-color: var(--color-text);
        background: var(--color-text);
        color: var(--color-white);
    }

    .yes-btn.selected {
        background: #333;
    }

    .no-btn.selected {
        background: #666;
    }

    .error {
        color: var(--color-text);
        font-size: 0.875rem;
        text-align: center;
        font-weight: 600;
        background: var(--color-background-alt);
        padding: var(--spacing-sm);
        border-radius: var(--radius-sm);
        border: 2px solid var(--color-border);
    }

    .button-group {
        display: flex;
        gap: var(--spacing-md);
        justify-content: center;
        margin-top: var(--spacing-md);
    }

    .submit-btn {
        margin-top: var(--spacing-md);
        font-size: 1.1rem;
        padding: var(--spacing-md) var(--spacing-xl);
    }

    .btn-icon-img {
        width: 18px;
        height: 18px;
        object-fit: contain;
    }

    .complete {
        text-align: center;
    }

    .success-content {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-xl) 0;
    }

    .success-icon img {
        width: 64px;
        height: 64px;
    }

    .success-content h3 {
        font-size: 2rem;
        color: var(--color-text);
        margin: 0;
    }

    .success-content p {
        margin: 0;
    }

    .sub-text {
        font-size: 0.875rem;
        color: var(--color-text-light);
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
