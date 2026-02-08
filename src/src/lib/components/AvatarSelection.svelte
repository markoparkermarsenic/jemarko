<script lang="ts">
    import { saveAvatars } from '$lib/api';
    import type { AvatarSelection as AvatarSelectionType } from '$lib/api';

    interface AvatarOption {
        id: string;
        image: string;
        name: string;
    }

    interface GuestAvatar {
        guestName: string;
        selectedAvatar: string | null;
        message: string;
        isDropdownOpen: boolean;
    }

    // Define the type for completed avatar data
    interface CompletedAvatar {
        name: string;
        avatar: string;
        message: string;
    }

    let {
        guests = ["Guest"],
        email = "",
        oncomplete,
    }: { guests?: string[]; email?: string; oncomplete?: (avatars: CompletedAvatar[]) => void } = $props();

    // Available avatar options - using the bird assets
    const avatarOptions: AvatarOption[] = [
        { id: "albatross", image: "/birds/albatross.png", name: "Albatross" },
        { id: "bluetit", image: "/birds/bluetit.png", name: "Blue Tit" },
        { id: "eagle", image: "/birds/eagle.png", name: "Eagle" },
        { id: "goose", image: "/birds/goose.png", name: "Goose" },
        { id: "hummingbird", image: "/birds/hummingbird.png", name: "Hummingbird" },
        { id: "owl", image: "/birds/owl.png", name: "Owl" },
        { id: "pigeon", image: "/birds/pigeon.png", name: "Pigeon" },
        { id: "raven", image: "/birds/raven.png", name: "Raven" },
        { id: "robin", image: "/birds/robin.png", name: "Robin" },
        { id: "swallow", image: "/birds/swallow.png", name: "Swallow" },
        { id: "swan", image: "/birds/swan.png", name: "Swan" },
    ];

    let guestAvatars = $state<GuestAvatar[]>(
        guests.map((name) => ({
            guestName: name,
            selectedAvatar: null,
            message: "",
            isDropdownOpen: false,
        })),
    );
    let isSubmitting = $state(false);

    function toggleDropdown(guestIndex: number) {
        guestAvatars = guestAvatars.map((ga, index) =>
            index === guestIndex
                ? { ...ga, isDropdownOpen: !ga.isDropdownOpen }
                : { ...ga, isDropdownOpen: false }
        );
    }

    function selectAvatar(guestIndex: number, avatarId: string) {
        guestAvatars = guestAvatars.map((ga, index) =>
            index === guestIndex
                ? { ...ga, selectedAvatar: avatarId, isDropdownOpen: false }
                : ga,
        );
    }

    function updateMessage(guestIndex: number, event: Event) {
        const target = event.target as HTMLTextAreaElement;
        const message = target.value.slice(0, 140); // Enforce 140 char limit
        guestAvatars = guestAvatars.map((ga, index) =>
            index === guestIndex ? { ...ga, message } : ga,
        );
    }

    let errorMessage = $state("");

    async function submitAvatars() {
        errorMessage = "";
        isSubmitting = true;

        try {
            // Prepare the avatar selections for the API
            const avatars: AvatarSelectionType[] = guestAvatars.map((ga) => ({
                guestName: ga.guestName,
                avatar: ga.selectedAvatar || "",
                message: ga.message,
            }));

            // Call the API to save avatar selections
            const response = await saveAvatars({
                email,
                avatars,
            });

            if (response.success) {
                console.log("Avatars saved successfully:", avatars);
                oncomplete?.();
            } else {
                errorMessage = response.message || "Failed to save avatar selections";
            }
        } catch (error: any) {
            errorMessage = error.message || "Unable to save avatar selections. Please try again.";
            console.error("Save avatars error:", error);
        } finally {
            isSubmitting = false;
        }
    }

    $effect(() => {
        // Update guestAvatars when guests prop changes
        if (guests.length !== guestAvatars.length) {
            guestAvatars = guests.map((name) => ({
                guestName: name,
                selectedAvatar: null,
                message: "",
                isDropdownOpen: false,
            }));
        }
    });

    function getSelectedAvatarOption(avatarId: string | null) {
        return avatarOptions.find((opt) => opt.id === avatarId);
    }

    let canSubmit = $derived(
        guestAvatars.every((ga) => ga.selectedAvatar !== null),
    );
</script>

<div class="avatar-selection card animate-fadeIn">
    <div class="selection-header">
        <h2>Choose Your Avatars</h2>
        <p class="subtitle">Select a bird and leave a message! this will be visible to everyone!!</p>
    </div>

    <div class="guests-list">
        {#each guestAvatars as guestAvatar, guestIndex (guestAvatar.guestName)}
            <div class="guest-section">
                <div class="guest-name-row">
                    <h3 class="guest-name">{guestAvatar.guestName}</h3>
                </div>

                <div class="avatar-dropdown-container">
                    <button
                        type="button"
                        class="dropdown-trigger"
                        class:has-selection={guestAvatar.selectedAvatar !== null}
                        onclick={() => toggleDropdown(guestIndex)}
                    >
                        {#if guestAvatar.selectedAvatar}
                            {@const selectedOption = getSelectedAvatarOption(guestAvatar.selectedAvatar)}
                            {#if selectedOption}
                                <img
                                    src={selectedOption.image}
                                    alt={selectedOption.name}
                                    class="dropdown-preview-img"
                                />
                                <span>{selectedOption.name}</span>
                            {/if}
                        {:else}
                            <span class="placeholder">Select an avatar...</span>
                        {/if}
                        <span class="dropdown-arrow" class:open={guestAvatar.isDropdownOpen}>▼</span>
                    </button>

                    {#if guestAvatar.isDropdownOpen}
                        <div class="dropdown-menu">
                            {#each avatarOptions as avatar (avatar.id)}
                                <button
                                    type="button"
                                    class="dropdown-option"
                                    class:selected={guestAvatar.selectedAvatar === avatar.id}
                                    onclick={() => selectAvatar(guestIndex, avatar.id)}
                                >
                                    <img
                                        src={avatar.image}
                                        alt={avatar.name}
                                        class="dropdown-option-img"
                                    />
                                    <span>{avatar.name}</span>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>

                <div class="message-section">
                    <label for="message-{guestIndex}">
                        Message (optional)
                        <span class="char-count" class:warning={guestAvatar.message.length > 120}>
                            {guestAvatar.message.length}/140
                        </span>
                    </label>
                    <textarea
                        id="message-{guestIndex}"
                        placeholder="Leave a message for the happy couple 💕"
                        value={guestAvatar.message}
                        oninput={(e) => updateMessage(guestIndex, e)}
                        maxlength="140"
                        rows="2"
                    ></textarea>
                </div>
            </div>
        {/each}
    </div>

    {#if errorMessage}
        <p class="error">{errorMessage}</p>
    {/if}

    <button
        class="btn btn-primary submit-btn"
        onclick={submitAvatars}
        disabled={!canSubmit || isSubmitting}
    >
        {#if isSubmitting}
            Saving...
        {:else}
            Join the Plaza!
        {/if}
    </button>
</div>

<style>
    .avatar-selection {
        max-width: 600px;
        width: 100%;
        margin: 0 auto;
        position: relative;
        z-index: 100;
        background: var(--color-white);
        border: 2px solid var(--color-border);
    }

    .selection-header {
        text-align: center;
        margin-bottom: var(--spacing-xl);
    }

    .selection-header h2 {
        font-size: 2rem;
        color: var(--color-text);
        margin-bottom: var(--spacing-xs);
    }

    .subtitle {
        color: var(--color-text-light);
        font-size: 0.95rem;
        margin: 0;
    }

    .guests-list {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-xl);
        margin-bottom: var(--spacing-xl);
    }

    .guest-section {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-md);
        padding: var(--spacing-lg);
        background: var(--color-background-alt);
        border: 2px solid var(--color-border-light);
        border-radius: var(--radius-md);
    }

    .guest-name-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .guest-name {
        font-family: var(--font-display);
        font-size: 1.3rem;
        color: var(--color-text);
        margin: 0;
    }

    .selection-badge {
        width: 24px;
        height: 24px;
        object-fit: contain;
    }

    .avatar-dropdown-container {
        position: relative;
    }

    .dropdown-trigger {
        width: 100%;
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        padding: var(--spacing-md);
        background: var(--color-white);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: all var(--transition-fast);
        font-family: var(--font-mimko);
        font-size: 1rem;
        text-align: left;
    }

    .dropdown-trigger:hover {
        border-color: var(--color-text);
    }

    .dropdown-trigger.has-selection {
        border-color: var(--color-text);
    }

    .dropdown-preview-img {
        width: 32px;
        height: 32px;
        object-fit: contain;
    }

    .placeholder {
        color: var(--color-text-light);
        flex: 1;
    }

    .dropdown-trigger span:not(.dropdown-arrow):not(.placeholder) {
        flex: 1;
    }

    .dropdown-arrow {
        margin-left: auto;
        font-size: 0.75rem;
        transition: transform var(--transition-fast);
    }

    .dropdown-arrow.open {
        transform: rotate(180deg);
    }

    .dropdown-menu {
        position: absolute;
        top: calc(100% + 4px);
        left: 0;
        right: 0;
        background: var(--color-white);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        max-height: 300px;
        overflow-y: auto;
        z-index: 200;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }

    .dropdown-option {
        width: 100%;
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        padding: var(--spacing-sm) var(--spacing-md);
        background: var(--color-white);
        border: none;
        border-bottom: 1px solid var(--color-border-light);
        cursor: pointer;
        transition: background var(--transition-fast);
        font-family: var(--font-mimko);
        font-size: 0.95rem;
        text-align: left;
    }

    .dropdown-option:last-child {
        border-bottom: none;
    }

    .dropdown-option:hover {
        background: var(--color-background-alt);
    }

    .dropdown-option.selected {
        background: var(--color-background-alt);
    }

    .dropdown-option-img {
        width: 32px;
        height: 32px;
        object-fit: contain;
    }

    .dropdown-option span {
        flex: 1;
    }

    .option-check {
        width: 18px;
        height: 18px;
        object-fit: contain;
    }

    .message-section {
        font-family: var(--font-mimko);
    }

    .message-section label {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: var(--spacing-xs);
        font-family: var(--font-mimko);
        font-size: 0.9rem;
    }

    .char-count {
        font-size: 0.75rem;
        color: var(--color-text-light);
    }

    .char-count.warning {
        color: var(--color-text);
        font-weight: 600;
    }

    textarea {
        width: 100%;
        padding: var(--spacing-sm) var(--spacing-md);
        border: 2px solid var(--color-border);
        border-radius: var(--radius-md);
        font-family: var(--font-mimko);
        font-size: 0.95rem;
        resize: none;
    }

    .submit-btn {
        width: 100%;
        font-size: 1.1rem;
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
        margin-bottom: var(--spacing-md);
    }

    .btn-icon-img {
        width: 18px;
        height: 18px;
        object-fit: contain;
    }

    @media (max-width: 480px) {
        .avatar-selection {
            padding: var(--spacing-lg);
        }

        .guest-section {
            padding: var(--spacing-md);
        }

        .guest-name {
            font-size: 1.1rem;
        }
    }
</style>
