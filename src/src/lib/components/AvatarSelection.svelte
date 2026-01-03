<script lang="ts">
    interface AvatarOption {
        id: string;
        image: string;
        name: string;
    }

    interface GuestAvatar {
        guestName: string;
        selectedAvatar: string | null;
        message: string;
    }

    let {
        guests = ["Guest"],
        oncomplete,
    }: { guests?: string[]; oncomplete?: () => void } = $props();

    // Available avatar options - using the cat assets
    const avatarOptions: AvatarOption[] = [
        { id: "cat-1", image: "/cat-1.png", name: "Whiskers" },
        { id: "cat-2", image: "/cat-2.png", name: "Mittens" },
        { id: "cat-3", image: "/cat-3.png", name: "Shadow" },
    ];

    let currentGuestIndex = $state(0);
    let guestAvatars = $state<GuestAvatar[]>(
        guests.map((name) => ({
            guestName: name,
            selectedAvatar: null,
            message: "",
        })),
    );
    let isSubmitting = $state(false);

    function selectAvatar(avatarId: string) {
        guestAvatars = guestAvatars.map((ga, index) =>
            index === currentGuestIndex
                ? { ...ga, selectedAvatar: avatarId }
                : ga,
        );
    }

    function updateMessage(event: Event) {
        const target = event.target as HTMLTextAreaElement;
        const message = target.value.slice(0, 140); // Enforce 140 char limit
        guestAvatars = guestAvatars.map((ga, index) =>
            index === currentGuestIndex ? { ...ga, message } : ga,
        );
    }

    function nextGuest() {
        if (currentGuestIndex < guests.length - 1) {
            currentGuestIndex++;
        }
    }

    function prevGuest() {
        if (currentGuestIndex > 0) {
            currentGuestIndex--;
        }
    }

    async function submitAvatars() {
        isSubmitting = true;

        // Simulate API call
        await new Promise((resolve) => setTimeout(resolve, 1000));

        console.log("Avatars submitted:", guestAvatars);

        isSubmitting = false;
        oncomplete?.();
    }

    $effect(() => {
        // Update guestAvatars when guests prop changes
        if (guests.length !== guestAvatars.length) {
            guestAvatars = guests.map((name) => ({
                guestName: name,
                selectedAvatar: null,
                message: "",
            }));
        }
    });

    let currentGuest = $derived(guestAvatars[currentGuestIndex]);
    let canSubmit = $derived(
        guestAvatars.every((ga) => ga.selectedAvatar !== null),
    );
    let messageLength = $derived(currentGuest?.message?.length || 0);
</script>

<div class="avatar-selection card animate-fadeIn">
    <div class="selection-header">
        <h2>Choose Your Avatar</h2>
        {#if guests.length > 1}
            <p class="guest-indicator">
                Selecting for: <strong>{currentGuest?.guestName}</strong>
                <span class="guest-count"
                    >({currentGuestIndex + 1} of {guests.length})</span
                >
            </p>
        {/if}
    </div>

    <div class="avatar-grid">
        {#each avatarOptions as avatar (avatar.id)}
            <button
                class="avatar-option"
                class:selected={currentGuest?.selectedAvatar === avatar.id}
                onclick={() => selectAvatar(avatar.id)}
            >
                <img src={avatar.image} alt={avatar.name} class="avatar-img" />
                <span class="avatar-name">{avatar.name}</span>
                {#if currentGuest?.selectedAvatar === avatar.id}
                    <img
                        src="/confirm.png"
                        alt="Selected"
                        class="selected-badge"
                    />
                {/if}
            </button>
        {/each}
    </div>

    <div class="message-section">
        <label for="message">
            Leave a message (optional)
            <span class="char-count" class:warning={messageLength > 120}>
                {messageLength}/140
            </span>
        </label>
        <textarea
            id="message"
            placeholder="Congratulations! We're so happy for you! 💕"
            value={currentGuest?.message || ""}
            oninput={updateMessage}
            maxlength="140"
            rows="3"
        ></textarea>
    </div>

    {#if guests.length > 1}
        <div class="navigation">
            <button
                class="btn btn-outline"
                onclick={prevGuest}
                disabled={currentGuestIndex === 0}
            >
                Previous
            </button>

            <div class="guest-dots">
                {#each guests as _, index}
                    <span
                        class="dot"
                        class:active={index === currentGuestIndex}
                        class:completed={guestAvatars[index]?.selectedAvatar !==
                            null}
                    ></span>
                {/each}
            </div>

            {#if currentGuestIndex < guests.length - 1}
                <button
                    class="btn btn-primary"
                    onclick={nextGuest}
                    disabled={!currentGuest?.selectedAvatar}
                >
                    Next Guest
                    <img src="/next-arrow.png" alt="" class="btn-icon-img" />
                </button>
            {:else}
                <button
                    class="btn btn-primary"
                    onclick={submitAvatars}
                    disabled={!canSubmit || isSubmitting}
                >
                    {#if isSubmitting}
                        Saving...
                    {:else}
                        Complete
                        <img src="/confirm.png" alt="" class="btn-icon-img" />
                    {/if}
                </button>
            {/if}
        </div>
    {:else}
        <button
            class="btn btn-primary submit-single"
            onclick={submitAvatars}
            disabled={!canSubmit || isSubmitting}
        >
            {#if isSubmitting}
                Saving...
            {:else}
                Join the Plaza!
                <img src="/confirm.png" alt="" class="btn-icon-img" />
            {/if}
        </button>
    {/if}
</div>

<style>
    .avatar-selection {
        max-width: 500px;
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
        margin-bottom: var(--spacing-sm);
    }

    .guest-indicator {
        color: var(--color-text-light);
        font-size: 0.95rem;
    }

    .guest-count {
        color: var(--color-text-light);
        font-size: 0.85rem;
    }

    .avatar-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: var(--spacing-md);
        margin-bottom: var(--spacing-xl);
    }

    .avatar-option {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: var(--spacing-md);
        background: var(--color-white);
        border: 2px solid var(--color-border-light);
        border-radius: var(--radius-lg);
        cursor: pointer;
        transition: all var(--transition-normal);
        position: relative;
    }

    .avatar-option:hover {
        transform: translateY(-4px);
        border-color: var(--color-border);
    }

    .avatar-option.selected {
        border-color: var(--color-border);
        background: var(--color-background-alt);
    }

    .avatar-img {
        width: 80px;
        height: 80px;
        object-fit: contain;
        margin-bottom: var(--spacing-sm);
        transition: transform var(--transition-normal);
    }

    .avatar-option:hover .avatar-img {
        transform: scale(1.1);
    }

    .avatar-name {
        font-family: var(--font-display);
        font-size: 1rem;
        color: var(--color-text);
    }

    .selected-badge {
        position: absolute;
        top: -8px;
        right: -8px;
        width: 28px;
        height: 28px;
        background: var(--color-white);
        border-radius: 50%;
        padding: 4px;
        border: 2px solid var(--color-border);
    }

    .message-section {
        margin-bottom: var(--spacing-xl);
        font-family: var(--font-mimko);
    }

    .message-section label {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: var(--spacing-xs);
        font-family: var(--font-mimko);
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
        font-family: var(--font-mimko);
        resize: none;
        min-height: 80px;
    }

    .navigation {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: var(--spacing-md);
    }

    .guest-dots {
        display: flex;
        gap: var(--spacing-xs);
    }

    .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: var(--color-border-light);
        border: 1px solid var(--color-border);
        transition: all var(--transition-normal);
    }

    .dot.active {
        background: var(--color-text);
        transform: scale(1.3);
    }

    .dot.completed {
        background: var(--color-text);
    }

    .btn-icon-img {
        width: 16px;
        height: 16px;
        object-fit: contain;
    }

    .submit-single {
        width: 100%;
    }

    @media (max-width: 480px) {
        .avatar-selection {
            padding: var(--spacing-lg);
        }

        .avatar-grid {
            grid-template-columns: repeat(3, 1fr);
            gap: var(--spacing-sm);
        }

        .avatar-img {
            width: 60px;
            height: 60px;
        }

        .navigation {
            flex-wrap: wrap;
            justify-content: center;
        }

        .guest-dots {
            order: -1;
            width: 100%;
            justify-content: center;
            margin-bottom: var(--spacing-md);
        }
    }
</style>
