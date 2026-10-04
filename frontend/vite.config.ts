import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

/**
 * Dev proxy → Go API (`go run ./cmd/revues` on :8080).
 * Same-origin paths used by the SPA and mb vendor:
 *   /api/*      → OpenAPI JSON
 *   /auth/*     → OAuth start/callback
 *   /healthz    → infra probe
 *   /static/*   → embedded vendor (mb tokens + CE)
 * See docs/FRONTEND.md.
 */
export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		proxy: {
			'/api': 'http://127.0.0.1:8080',
			'/auth': 'http://127.0.0.1:8080',
			'/healthz': 'http://127.0.0.1:8080',
			'/static': 'http://127.0.0.1:8080'
		},
		// check.sh écrit frontend/build/ — sans ignore, Vite HMR recharge ces HTML
		// et peut casser le client SPA (« Internal Error »).
		watch: {
			ignored: ['**/build/**', '**/.svelte-kit/output/**']
		}
	}
});
