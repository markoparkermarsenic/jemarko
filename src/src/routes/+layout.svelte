<script lang="ts">
	import "../app.css";
	import favicon from "$lib/assets/favicon.ico";

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Jemarko Wedding</title>
	<meta
		name="description"
		content="Join us for our special day - RSVP and leave your mark in the avatar plaza!"
	/>
</svelte:head>

<main class="app">
	{@render children()}
</main>

<style>
	.app {
		min-height: 100vh;
		position: relative;
		overflow: hidden;
	}
</style>
