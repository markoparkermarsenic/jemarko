import adapter from '@sveltejs/adapter-vercel';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),

	kit: {
		adapter: adapter(),
		prerender: {
			// Crawl from the homepage and prerender all reachable pages that
			// have `export const prerender = true`. The /info page is static
			// and gets prerendered to a static HTML file, which @vercel/static-build
			// can serve directly (without needing an SSR serverless function).
			entries: ['*']
		}
	}
};

export default config;
