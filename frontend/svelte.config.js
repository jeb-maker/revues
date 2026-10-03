import adapter from '@sveltejs/adapter-static';

/**
 * `mb-button` rend un vrai `<button>` dans son shadow DOM : les règles a11y
 * « click sans clavier / élément statique interactif » sont des faux positifs
 * sur l'hôte. Filtre central plutôt que des `svelte-ignore` dispersés.
 */
const MB_BUTTON_FALSE_POSITIVES = new Set([
	'a11y_click_events_have_key_events',
	'a11y_no_static_element_interactions'
]);

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		warningFilter: (warning) =>
			!(MB_BUTTON_FALSE_POSITIVES.has(warning.code) && warning.message.includes('<mb-button>'))
	},
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',
			precompress: false,
			strict: true
		}),
		prerender: {
			handleUnseenRoutes: 'ignore',
			// SPA (ssr=false) : les coquilles prerendues n'ont aucun lien à suivre, et les
			// assets mb référencés dans app.html sont servis par Go (/static), pas par le prerender.
			crawl: false
		}
	}
};

export default config;
