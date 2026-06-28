<script lang="ts">
    import { onMount } from "svelte";
    import { getAvatars } from '$lib/api';
    import type { GuestAvatar } from '$lib/api';

    interface Props {
        refreshTrigger?: number;
    }

    let { refreshTrigger = 0 }: Props = $props();

    interface Avatar {
        id: string;
        name: string;
        image: string;
        message?: string;
        x: number;
        y: number;
        targetX: number;
        targetY: number;
        speed: number;
        showMessage: boolean;
        direction: "left" | "right";
        animDuration: number;
        animDelay: number;
        borderSeed: number;
        // DOM element ref — used to move the bird without touching Svelte state
        el?: HTMLDivElement;
    }

    // avatars drives the Svelte template (for mounting/unmounting birds).
    // Position updates bypass reactivity entirely — see rafLoop below.
    let avatars = $state<Avatar[]>([]);
    let userInteractionActive = $state(false);

    let containerWidth = 800;
    let containerHeight = 600;
    let container: HTMLDivElement;

    // ── Hot-path data ─────────────────────────────────────────────────────
    // These plain arrays/objects are read 60× per second inside rafTick.
    // They MUST NOT be Svelte proxies — any proxy read inside a tight loop
    // costs ~3× vs a plain property access due to the proxy trap overhead.

    // Per-bird position state — plain object, never proxied
    const pos: Record<string, {
        x: number; y: number;
        targetX: number; targetY: number;
        speed: number; direction: "left" | "right";
    }> = {};

    // Flat array of { id, el } — rebuilt only when birds are added/removed
    // rafTick reads this instead of the $state proxy array
    let rafBirds: Array<{ id: string; el: HTMLDivElement }> = [];


    // Function to fetch and merge avatars while preserving existing positions
    async function fetchAvatars() {
        try {
            const response = await getAvatars();
            if (response.success && response.avatars.length > 0) {
                // Create a map of existing avatars by name for quick lookup
                const existingAvatarMap = new Map(avatars.map(a => [a.name, a]));
                
                // Map API avatars, preserving positions for existing ones
                const mergedAvatars = response.avatars.map((guestAvatar: GuestAvatar, index: number) => {
                    const existingAvatar = existingAvatarMap.get(guestAvatar.name);
                    
                    if (existingAvatar) {
                        // Preserve existing avatar's position and movement state
                        return {
                            ...existingAvatar,
                            // Update data that might have changed
                            image: `/birds/${guestAvatar.avatar}.png`,
                            message: guestAvatar.message || undefined,
                        };
                    }
                    
                    // New avatar — initialise pos entry and create avatar record
                    const id = `${guestAvatar.name}-${Date.now()}`;
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
                        // x/y on the avatar object are only used for initial Svelte render;
                        // the rAF loop takes over from there via pos[id]
                        x: startX, y: startY,
                        targetX: startX, targetY: startY,
                        speed: 0.3 + Math.random() * 0.5,
                        showMessage: false,
                        direction: Math.random() > 0.5 ? "right" : "left" as "left" | "right",
                        animDuration: 0.8 + Math.random() * 0.6,
                        animDelay: Math.random() * 0.3,
                        borderSeed: Math.random(),
                    };
                });
                avatars = mergedAvatars;
            }
        } catch (error) {
            console.error('Error fetching avatars:', error);
        }
    }

    // Re-fetch avatars when refreshTrigger changes
    $effect(() => {
        if (refreshTrigger > 0) {
            fetchAvatars();
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
        const W = containerWidth;
        const H = containerHeight;
        const margin = 80;
        const wRange = W - margin * 2;
        const hRange = H - margin * 2;

        for (let i = 0; i < rafBirds.length; i++) {
            const bird = rafBirds[i];
            const p = pos[bird.id];
            if (!p) continue;

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

    // Rebuild the plain rafBirds array whenever the Svelte avatars list changes.
    // This runs at most once per fetch (not per frame).
    // We use $effect so it fires after the DOM has settled (bind:this populated).
    $effect(() => {
        // Read avatars.length to subscribe; then snapshot into a plain array
        // using a microtask so all bind:this refs are guaranteed to be set.
        const _ = avatars.length;
        void 0; // silence unused warning
        queueMicrotask(() => {
            rafBirds = avatars
                .filter(a => a.el != null)
                .map(a => ({ id: a.id, el: a.el as HTMLDivElement }));
        });
    });

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
                const avatarToHide = showingAvatars[randomHideIndex];
                avatars = avatars.map((avatar) => ({
                    ...avatar,
                    showMessage: avatar.id === avatarToHide.id ? false : avatar.showMessage,
                }));
            }
            return;
        }

        // Pick a random avatar to show message
        const randomIndex = Math.floor(Math.random() * avatarsWithMessages.length);
        const avatarToShow = avatarsWithMessages[randomIndex];
        
        avatars = avatars.map((avatar) => ({
            ...avatar,
            showMessage: avatar.id === avatarToShow.id ? true : avatar.showMessage,
        }));

        // Hide this specific message after 5 seconds
        setTimeout(() => {
            avatars = avatars.map((avatar) => ({
                ...avatar,
                showMessage: avatar.id === avatarToShow.id ? false : avatar.showMessage,
            }));
        }, 5000);
    }

    function handleAvatarClick(avatarId: string) {
        const clickedAvatar = avatars.find(a => a.id === avatarId);
        if (!clickedAvatar) return;

        // User has interacted - pause random messages
        userInteractionActive = true;

        // If clicking an avatar that's already showing, hide it
        if (clickedAvatar.showMessage) {
            avatars = avatars.map((avatar) => ({
                ...avatar,
                showMessage: avatar.id === avatarId ? false : avatar.showMessage,
            }));
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
                avatars = avatars.map((avatar) => ({
                    ...avatar,
                    showMessage: avatar.id === firstShowing.id ? false : avatar.showMessage,
                }));
            }
        }

        // Show the clicked avatar's message
        avatars = avatars.map((avatar) => ({
            ...avatar,
            showMessage: avatar.id === avatarId ? true : avatar.showMessage,
        }));

        // Hide this specific message after 5 seconds
        setTimeout(() => {
            avatars = avatars.map((avatar) => ({
                ...avatar,
                showMessage: avatar.id === avatarId ? false : avatar.showMessage,
            }));
            // Check if we should resume random messages
            checkAndResetInteraction();
        }, 5000);
    }

    onMount(() => {
        if (container) {
            containerWidth = container.offsetWidth;
            containerHeight = container.offsetHeight;
        }

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

        const handleResize = () => {
            if (container) {
                containerWidth = container.offsetWidth;
                containerHeight = container.offsetHeight;
            }
        };
        window.addEventListener("resize", handleResize);

        return () => {
            cancelAnimationFrame(rafId);
            window.removeEventListener("resize", handleResize);
        };
    });
</script>

<!-- SVG filter for fuzzy/hand-drawn effect -->
<svg class="svg-filters" aria-hidden="true">
    <defs>
        <filter id="fuzzy-border">
            <feTurbulence type="fractalNoise" baseFrequency="0.04" numOctaves="3" result="noise" />
            <feDisplacementMap in="SourceGraphic" in2="noise" scale="2" xChannelSelector="R" yChannelSelector="G" />
        </filter>
    </defs>
</svg>

<div class="avatar-plaza" bind:this={container}>
    {#each avatars as avatar (avatar.id)}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
            class="avatar"
            class:flip={avatar.direction === "left"}
            class:has-message={!!avatar.message}
            class:showing-message={avatar.showMessage}
            style="left:0;top:0;will-change:left,top;"
            bind:this={avatar.el}
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
                    style="--anim-duration: {avatar.animDuration}s; --anim-delay: {avatar.animDelay}s; --border-seed: {avatar.borderSeed};"
                >
                    <span class="message">{avatar.message}</span>
                    <div class="bubble-tail"></div>
                </div>
            {/if}
        </div>
    {/each}
</div>

<style>
    .svg-filters {
        position: absolute;
        width: 0;
        height: 0;
        overflow: hidden;
    }

    .avatar-plaza {
        position: absolute;
        inset: 0;
        overflow: hidden;
        z-index: 0;
        background-color: #ffffff;
    }

    .avatar {
        position: absolute;
        display: flex;
        flex-direction: column;
        align-items: center;
        transition: transform 0.1s ease;
        z-index: 10;
        font-family: var(--font-body);
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
        background: #ffffff; /* Solid white background */
        padding: 10px 14px;
        font-size: 0.75rem;
        max-width: 150px;
        text-align: center;
        position: relative;
        /* Fuzzy hand-drawn border effect */
        border: 2px solid #333;
        border-radius: 16px;
        filter: url(#fuzzy-border);
        animation: borderDraw var(--anim-duration, 1s) ease-out forwards;
        /* Ensure solid background */
        background-clip: padding-box;
        -webkit-background-clip: padding-box;
    }
    
    /* Add a white backdrop layer underneath the message */
    .message::before {
        content: '';
        position: absolute;
        inset: 0;
        background: #ffffff;
        border-radius: 16px;
        z-index: -1;
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
        filter: url(#fuzzy-border);
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

    @keyframes borderDraw {
        0% {
            clip-path: polygon(0 0, 0 0, 0 100%, 0 100%);
        }
        100% {
            clip-path: polygon(0 0, 100% 0, 100% 100%, 0 100%);
        }
    }
</style>
