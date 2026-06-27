<script lang="ts">
    import "../../app.css";
    import favicon from "$lib/assets/favicon.ico";

    let { children } = $props();
</script>

<svelte:head>
    <link rel="icon" href={favicon} />
    <title>Admin — Jemarko</title>
</svelte:head>

{@render children()}

<style>
    :global(html, body) {
        overflow: auto;
        height: auto;
        background: #f5f5f5;
    }
</style>
