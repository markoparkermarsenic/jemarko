<script lang="ts">
    import { onMount } from "svelte";

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
    }

    // Demo avatars - in production these would come from the database
    let avatars = $state<Avatar[]>([
        {
            id: "1",
            name: "Guest 1",
            image: "/cat-1.png",
            message: "So excited for the wedding!",
            x: 100,
            y: 200,
            targetX: 300,
            targetY: 250,
            speed: 0.5,
            showMessage: false,
            direction: "right",
        },
        {
            id: "2",
            name: "Guest 2",
            image: "/cat-2.png",
            message: "Congratulations! 🎉",
            x: 400,
            y: 150,
            targetX: 200,
            targetY: 300,
            speed: 0.7,
            showMessage: false,
            direction: "left",
        },
        {
            id: "3",
            name: "Guest 3",
            image: "/cat-3.png",
            message: "Love you both!",
            x: 250,
            y: 350,
            targetX: 450,
            targetY: 180,
            speed: 0.4,
            showMessage: false,
            direction: "right",
        },
    ]);

    let containerWidth = $state(0);
    let containerHeight = $state(0);
    let container: HTMLDivElement;

    function getRandomPosition(max: number, margin: number = 80): number {
        return Math.random() * (max - margin * 2) + margin;
    }

    function updateAvatarTargets() {
        avatars = avatars.map((avatar) => {
            // If close to target, set new target
            const dx = avatar.targetX - avatar.x;
            const dy = avatar.targetY - avatar.y;
            const distance = Math.sqrt(dx * dx + dy * dy);

            if (distance < 10) {
                return {
                    ...avatar,
                    targetX: getRandomPosition(containerWidth),
                    targetY: getRandomPosition(containerHeight),
                };
            }
            return avatar;
        });
    }

    function moveAvatars() {
        avatars = avatars.map((avatar) => {
            const dx = avatar.targetX - avatar.x;
            const dy = avatar.targetY - avatar.y;
            const distance = Math.sqrt(dx * dx + dy * dy);

            if (distance > 1) {
                const newX = avatar.x + (dx / distance) * avatar.speed;
                const newY = avatar.y + (dy / distance) * avatar.speed;
                return {
                    ...avatar,
                    x: newX,
                    y: newY,
                    direction: dx > 0 ? "right" : "left",
                };
            }
            return avatar;
        });
    }

    function toggleRandomMessage() {
        const randomIndex = Math.floor(Math.random() * avatars.length);
        avatars = avatars.map((avatar, index) => ({
            ...avatar,
            showMessage: index === randomIndex ? !avatar.showMessage : false,
        }));

        // Hide message after 3 seconds
        setTimeout(() => {
            avatars = avatars.map((avatar) => ({
                ...avatar,
                showMessage: false,
            }));
        }, 3000);
    }

    onMount(() => {
        if (container) {
            containerWidth = container.offsetWidth;
            containerHeight = container.offsetHeight;

            // Initialize positions based on container size
            avatars = avatars.map((avatar) => ({
                ...avatar,
                x: getRandomPosition(containerWidth),
                y: getRandomPosition(containerHeight),
                targetX: getRandomPosition(containerWidth),
                targetY: getRandomPosition(containerHeight),
            }));
        }

        // Animation loop
        const moveInterval = setInterval(() => {
            moveAvatars();
            updateAvatarTargets();
        }, 50);

        // Show messages periodically
        const messageInterval = setInterval(toggleRandomMessage, 5000);

        // Handle resize
        const handleResize = () => {
            if (container) {
                containerWidth = container.offsetWidth;
                containerHeight = container.offsetHeight;
            }
        };
        window.addEventListener("resize", handleResize);

        return () => {
            clearInterval(moveInterval);
            clearInterval(messageInterval);
            window.removeEventListener("resize", handleResize);
        };
    });
</script>

<div class="avatar-plaza" bind:this={container}>
    {#each avatars as avatar (avatar.id)}
        <div
            class="avatar"
            class:flip={avatar.direction === "left"}
            style="left: {avatar.x}px; top: {avatar.y}px;"
        >
            {#if avatar.showMessage && avatar.message}
                <div class="speech-bubble">
                    <img src="/speech-bubble.png" alt="" class="bubble-bg" />
                    <span class="message">{avatar.message}</span>
                </div>
            {/if}
            <img src={avatar.image} alt={avatar.name} class="avatar-image" />
            <span class="avatar-name">{avatar.name}</span>
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
    }

    .avatar {
        position: absolute;
        display: flex;
        flex-direction: column;
        align-items: center;
        transition: transform 0.1s ease;
        z-index: 10;
        font-family: var(--font-mimko);
    }

    .avatar.flip .avatar-image {
        transform: scaleX(-1);
    }

    .avatar-image {
        width: 60px;
        height: 60px;
        object-fit: contain;
        animation: float 2s ease-in-out infinite;
    }

    .avatar-name {
        font-family: var(--font-display);
        font-size: 0.75rem;
        color: var(--color-text);
        background: var(--color-white);
        padding: 2px 8px;
        border-radius: 10px;
        margin-top: 4px;
        white-space: nowrap;
        border: 1px solid var(--color-border);
    }

    .speech-bubble {
        position: absolute;
        bottom: 100%;
        left: 50%;
        transform: translateX(-50%);
        margin-bottom: 8px;
        animation: fadeIn 0.3s ease;
    }

    .bubble-bg {
        width: 120px;
        height: auto;
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        z-index: -1;
        opacity: 0.8;
    }

    .message {
        display: block;
        background: white;
        padding: 8px 12px;
        border-radius: 12px;
        font-size: 0.75rem;
        max-width: 140px;
        text-align: center;
        border: 2px solid var(--color-border);
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
</style>
