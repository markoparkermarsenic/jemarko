<script lang="ts">
    import { onMount } from "svelte";
    import { getAvatars } from '$lib/api';
    import type { GuestAvatar } from '$lib/api';

    interface Props {
        refreshTrigger?: number;
    }

    let { refreshTrigger = 0 }: Props = $props();

    // Avatar data used only by Svelte for rendering (image, name, showMessage).
    // Position is never stored here — it lives in the plain `pos` map below.
    interface Avatar {
        id: string;
        name: string;
        image: string;
        message?: string;
        showMessage: boolean;
        animDuration: number;
        animDelay: number;
    }

    // Svelte reactive state — drives the template. Never touched in rafTick.
    let avatars = $state<Avatar[]>([]);
    let userInteractionActive = $state(false);

    let containerWidth = 800;
    let containerHeight = 600;
    let container: HTMLDivElement;

    // ── Hot-path data: plain JS, never proxied ────────────────────────────

    // Per-bird position + direction state
    const pos: Record<string, {
        x: number; y: number;
        targetX: number; targetY: number;
        speed: number; direction: "left" | "right";
    }> = {};

    // Flat snapshot used by rafTick — rebuilt from the DOM when stale.
    let rafBirds: Array<{ id: string; el: HTMLDivElement }> = [];

    // Rebuild rafBirds by querying the DOM directly. Called from rafTick's
    // self-heal check whenever the DOM child count doesn't match rafBirds.
    // This is deliberately NOT a $effect — effects in Svelte 5.46 +
    // vite-plugin-svelte@3 have unreliable timing. The RAF loop is the
    // single source of truth: it runs 60fps and self-corrects within one
    // frame if birds are added/removed.
    function rebuildRafBirdsFromDOM() {
        if (!container) return;
        const els = container.querySelectorAll<HTMLDivElement>('.avatar[data-bird-id]');
        const next: Array<{ id: string; el: HTMLDivElement }> = [];
        els.forEach((el) => {
            const id = el.dataset.birdId;
            if (!id) return;
            if (!pos[id]) {
                pos[id] = {
                    x: getRandomPosition(containerWidth),
                    y: getRandomPosition(containerHeight),
                    targetX: getRandomPosition(containerWidth),
                    targetY: getRandomPosition(containerHeight),
                    speed: 0.3 + Math.random() * 0.5,
                    direction: Math.random() > 0.5 ? 'right' : 'left',
                };
            }
            next.push({ id, el });
        });
        rafBirds = next;
    }


    // Guard to prevent concurrent fetchAvatars calls. Without this, if the
    // $effect re-fires before a previous fetch completes, multiple concurrent
    // fetches create duplicate birds with new Date.now() IDs — each spawning
    // pos entries — causing unbounded memory growth.
    let isFetching = false;

    // Function to fetch and merge avatars while preserving existing positions
    async function fetchAvatars() {
        if (isFetching) return;
        isFetching = true;
        try {
            const response = await getAvatars();
            if (response.success && response.avatars.length > 0) {
                // Create a map of existing avatars by name for quick lookup
                const existingAvatarMap = new Map(avatars.map(a => [a.name, a]));
                
                // Track which IDs survive the merge so we can clean up pos
                const survivingIds = new Set<string>();
                
                // Map API avatars, preserving positions for existing ones
                const mergedAvatars = response.avatars.map((guestAvatar: GuestAvatar, index: number) => {
                    const existingAvatar = existingAvatarMap.get(guestAvatar.name);
                    
                    if (existingAvatar) {
                        // Preserve existing avatar's position and movement state
                        survivingIds.add(existingAvatar.id);
                        return {
                            ...existingAvatar,
                            // Update data that might have changed
                            image: `/birds/${guestAvatar.avatar}.png`,
                            message: guestAvatar.message || undefined,
                        };
                    }
                    
                    // New avatar — initialise pos entry and create avatar record
                    const id = `${guestAvatar.name}-${Date.now()}`;
                    survivingIds.add(id);
                    const startX = getRandomPosition(containerWidth);
                    const startY = getRandomPosition(containerHeight);
                    pos[id] = {
                        x: startX, y: startY,
                        targetX: getRandomPosition(containerWidth),
                        targetY: getRandomPosition(containerHeight),
                        speed: 0.3 + Math.random() * 0.5,
                        direction: Math.random() > 0.5 ? "right" : "left",
                    };
                    return {
                        id,
                        name: guestAvatar.name,
                        image: `/birds/${guestAvatar.avatar}.png`,
                        message: guestAvatar.message || undefined,
                        showMessage: false,
                        animDuration: 0.8 + Math.random() * 0.6,
                        animDelay: Math.random() * 0.3,
                    };
                });
                avatars = mergedAvatars;

                // Clean up pos entries for birds that no longer exist
                for (const id in pos) {
                    if (!survivingIds.has(id)) delete pos[id];
                }
            }
        } catch (error) {
            console.error('Error fetching avatars:', error);
        } finally {
            isFetching = false;
        }
    }

    // Re-fetch avatars when refreshTrigger changes.
    // Deferred to an idle callback so the fetch + render of ~100 birds
    // doesn't block button interactivity or hydration. Buttons become
    // active immediately; birds appear during an idle period.
    $effect(() => {
        if (refreshTrigger <= 0) return;
        const run = () => fetchAvatars();
        if (typeof requestIdleCallback !== 'undefined') {
            requestIdleCallback(run);
        } else {
            setTimeout(run, 50);
        }
    });

    function getRandomPosition(max: number, margin: number = 80): number {
        return Math.random() * (max - margin * 2) + margin;
    }

    // ── rafTick — the only code that runs at 60fps ────────────────────────
    // Reads from rafBirds (plain array, not a Svelte proxy) and pos (plain object).
    // Zero allocations in the hot path: no string templates, no map/filter, no
    // array spreads. At 160 birds @ 60fps this runs ~9,600 iterations/sec.
    function rafTick() {
        // Self-heal: if the DOM has a different number of bird elements than
        // rafBirds, rebuild from the DOM. container.children.length is O(1).
        // This catches new birds (after fetchAvatars) and removed birds,
        // without relying on $effect or bind:this timing.
        if (container && container.children.length !== rafBirds.length) {
            rebuildRafBirdsFromDOM();
        }

        const W = containerWidth;
        const H = containerHeight;
        const margin = 80;
        const wRange = W - margin * 2;
        const hRange = H - margin * 2;

        for (let i = 0; i < rafBirds.length; i++) {
            const bird = rafBirds[i];
            let p = pos[bird.id];
            if (!p) {
                p = pos[bird.id] = {
                    x: getRandomPosition(W),
                    y: getRandomPosition(H),
                    targetX: getRandomPosition(W),
                    targetY: getRandomPosition(H),
                    speed: 0.3 + Math.random() * 0.5,
                    direction: Math.random() > 0.5 ? 'right' : 'left',
                };
            }

            const dx = p.targetX - p.x;
            const dy = p.targetY - p.y;
            // Avoid sqrt when close — just pick a new target
            const distSq = dx * dx + dy * dy;

            if (distSq > 1) {
                const invDist = p.speed / Math.sqrt(distSq);
                p.x += dx * invDist;
                p.y += dy * invDist;

                const goingLeft = dx < 0;
                if (goingLeft !== (p.direction === "left")) {
                    p.direction = goingLeft ? "left" : "right";
                    if (goingLeft) {
                        bird.el.classList.add("flip");
                    } else {
                        bird.el.classList.remove("flip");
                    }
                }
            } else {
                p.targetX = Math.random() * wRange + margin;
                p.targetY = Math.random() * hRange + margin;
            }

            // Direct style property writes — faster than template literal + transform
            bird.el.style.left = p.x + "px";
            bird.el.style.top  = p.y + "px";
        }
    }

    const MAX_VISIBLE_MESSAGES = 3;

    // Check if any messages are currently visible and reset interaction flag if none
    function checkAndResetInteraction() {
        const currentlyShowing = avatars.filter(a => a.showMessage).length;
        if (currentlyShowing === 0) {
            userInteractionActive = false;
        }
    }

    function toggleRandomMessage() {
        // Skip random messages if user has interacted
        if (userInteractionActive) {
            checkAndResetInteraction();
            return;
        }

        // Get avatars that have messages and are not currently showing
        const avatarsWithMessages = avatars.filter(a => a.message && !a.showMessage);
        const currentlyShowing = avatars.filter(a => a.showMessage).length;
        
        if (avatarsWithMessages.length === 0 || currentlyShowing >= MAX_VISIBLE_MESSAGES) {
            // Hide a random currently showing message to make room
            const showingAvatars = avatars.filter(a => a.showMessage);
            if (showingAvatars.length > 0) {
                const randomHideIndex = Math.floor(Math.random() * showingAvatars.length);
                showingAvatars[randomHideIndex].showMessage = false;
            }
            return;
        }

        // Pick a random avatar to show message
        const randomIndex = Math.floor(Math.random() * avatarsWithMessages.length);
        const avatarToShow = avatarsWithMessages[randomIndex];
        avatarToShow.showMessage = true;

        // Hide this specific message after 5 seconds
        setTimeout(() => {
            avatarToShow.showMessage = false;
        }, 5000);
    }

    function handleAvatarClick(avatarId: string) {
        const clickedAvatar = avatars.find(a => a.id === avatarId);
        if (!clickedAvatar) return;

        // User has interacted - pause random messages
        userInteractionActive = true;

        // If clicking an avatar that's already showing, hide it
        if (clickedAvatar.showMessage) {
            clickedAvatar.showMessage = false;
            // Check if we should resume random messages
            setTimeout(checkAndResetInteraction, 100);
            return;
        }

        // Count currently visible messages
        const currentlyShowing = avatars.filter(a => a.showMessage).length;

        // If we're at max, hide the oldest/first one
        if (currentlyShowing >= MAX_VISIBLE_MESSAGES) {
            const firstShowing = avatars.find(a => a.showMessage);
            if (firstShowing) {
                firstShowing.showMessage = false;
            }
        }

        // Show the clicked avatar's message
        clickedAvatar.showMessage = true;

        // Hide this specific message after 5 seconds
        setTimeout(() => {
            clickedAvatar.showMessage = false;
            // Check if we should resume random messages
            checkAndResetInteraction();
        }, 5000);
    }

    onMount(() => {
        const measure = () => {
            if (container && container.offsetWidth > 0) {
                containerWidth = container.offsetWidth;
                containerHeight = container.offsetHeight;
            } else if (typeof window !== 'undefined') {
                containerWidth = window.innerWidth;
                containerHeight = window.innerHeight;
            }
        };
        measure();

        // rAF animation loop — smooth 60fps, zero Svelte reactive overhead
        let rafId: number;
        let lastMessageTime = 0;
        const MESSAGE_INTERVAL_MS = 5000;

        function loop(timestamp: number) {
            rafTick();

            // Fire message logic at the same cadence as the old setInterval
            if (timestamp - lastMessageTime >= MESSAGE_INTERVAL_MS) {
                lastMessageTime = timestamp;
                toggleRandomMessage();
            }

            rafId = requestAnimationFrame(loop);
        }

        rafId = requestAnimationFrame(loop);

        const handleResize = () => measure();
        window.addEventListener("resize", handleResize);

        return () => {
            cancelAnimationFrame(rafId);
            window.removeEventListener("resize", handleResize);
        };
    });
</script>

<div class="avatar-plaza" bind:this={container}>
    {#each avatars as avatar (avatar.id)}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
            class="avatar"
            class:has-message={!!avatar.message}
            class:showing-message={avatar.showMessage}
            data-bird-id={avatar.id}
            onclick={(e) => {
                e.stopPropagation();
                handleAvatarClick(avatar.id);
            }}
        >
            <img src={avatar.image} alt={avatar.name} class="avatar-image" />
            <span class="avatar-name">{avatar.name}</span>
            {#if avatar.showMessage && avatar.message}
                <div
                    class="speech-bubble"
                    style="--anim-duration: {avatar.animDuration}s; --anim-delay: {avatar.animDelay}s;"
                >
                    <span class="message">{avatar.message}</span>
                    <div class="bubble-tail"></div>
                </div>
            {/if}
        </div>
    {/each}
</div>

<style>
    .avatar-plaza {
        position: absolute;
        inset: 0;
        overflow: hidden;
        z-index: 0;
        background-color: #ffffff;
        pointer-events: none; /* Container doesn't capture clicks; only individual birds do */
    }

    .avatar {
        position: absolute;
        display: flex;
        flex-direction: column;
        align-items: center;
        transition: transform 0.1s ease;
        z-index: 10;
        font-family: var(--font-body);
        pointer-events: auto; /* Re-enable clicks on individual birds */
    }

    .avatar.has-message {
        cursor: pointer;
    }

    .avatar.showing-message {
        z-index: 999; /* Higher than other avatars when showing message */
    }

    .avatar.has-message:hover .avatar-image {
        transform: scale(1.1);
    }

    .avatar.flip .avatar-image {
        transform: scaleX(-1);
    }

    .avatar.flip.has-message:hover .avatar-image {
        transform: scaleX(-1) scale(1.1);
    }

    .avatar-image {
        width: 60px;
        height: 60px;
        object-fit: contain;
        animation: float 2s ease-in-out infinite;
    }

    .avatar-name {
        font-family: var(--font-display);
        font-size: 0.6rem;
        color: var(--color-text);
        background: var(--color-white);
        padding: 2px 8px;
        border-radius: 10px;
        margin-top: 4px;
        white-space: nowrap;
    }

    .speech-bubble {
        position: absolute;
        bottom: 100%;
        left: 50%;
        transform: translateX(-50%);
        margin-bottom: 12px;
        z-index: 1000; /* Ensure speech bubbles appear above all avatars */
        animation: 
            fadeIn 0.3s ease forwards,
            wobble var(--anim-duration, 1s) ease-in-out var(--anim-delay, 0s) infinite;
    }

    .message {
        display: block;
        background: #ffffff;
        padding: 10px 14px;
        font-size: 0.75rem;
        max-width: 150px;
        text-align: center;
        border: 2px solid #333;
        border-radius: 16px;
    }

    .bubble-tail {
        position: absolute;
        bottom: -8px;
        left: 50%;
        transform: translateX(-50%);
        width: 0;
        height: 0;
        border-left: 8px solid transparent;
        border-right: 8px solid transparent;
        border-top: 10px solid #333;
        animation: tailWobble var(--anim-duration, 1s) ease-in-out var(--anim-delay, 0s) infinite;
    }

    .bubble-tail::before {
        content: '';
        position: absolute;
        top: -12px;
        left: -6px;
        width: 0;
        height: 0;
        border-left: 6px solid transparent;
        border-right: 6px solid transparent;
        border-top: 8px solid white;
    }

    @keyframes float {
        0%,
        100% {
            transform: translateY(0);
        }
        50% {
            transform: translateY(-3px);
        }
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateX(-50%) translateY(10px);
        }
        to {
            opacity: 1;
            transform: translateX(-50%) translateY(0);
        }
    }

    @keyframes wobble {
        0%, 100% {
            transform: translateX(-50%) rotate(0deg);
        }
        25% {
            transform: translateX(-50%) rotate(0.5deg);
        }
        50% {
            transform: translateX(-50%) rotate(-0.3deg);
        }
        75% {
            transform: translateX(-50%) rotate(0.4deg);
        }
    }

    @keyframes tailWobble {
        0%, 100% {
            transform: translateX(-50%) skewX(0deg);
        }
        33% {
            transform: translateX(-50%) skewX(2deg);
        }
        66% {
            transform: translateX(-50%) skewX(-2deg);
        }
    }

</style>
