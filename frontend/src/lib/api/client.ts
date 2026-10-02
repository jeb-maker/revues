/**
 * Typed OpenAPI client for `/api/v1`.
 * Types: `./schema.d.ts` (generated). Regenerate with `make frontend-api`.
 */
import createClient from 'openapi-fetch';
import type { paths } from './schema';

/** Same-origin API base (Vite proxies `/api` → Go `:8080` in dev). */
export const API_V1_BASE = '/api/v1';

/**
 * Shared fetch client: cookies (session) + optional CSRF on mutations.
 * Pass `csrf` via `client.use(...)` middleware or per-call headers.
 */
export function createApiClient(opts?: { csrfToken?: string }) {
	const client = createClient<paths>({
		baseUrl: API_V1_BASE,
		credentials: 'include'
	});

	if (opts?.csrfToken) {
		client.use({
			onRequest({ request }) {
				const method = request.method.toUpperCase();
				if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
					request.headers.set('X-CSRF-Token', opts.csrfToken!);
				}
				return request;
			}
		});
	}

	return client;
}

/** Default client without CSRF (safe for GET bootstrap / health). */
export const api = createApiClient();

/** Convenience: GET /api/v1/health */
export async function getHealth(): Promise<{ status: 'ok' }> {
	const { data, error } = await api.GET('/health');
	if (data) return data;
	const message =
		error && typeof error === 'object' && 'message' in error
			? String((error as { message?: unknown }).message)
			: 'health request failed';
	throw new Error(message);
}

export type { paths };
export type HealthResponse = paths['/health']['get']['responses']['200']['content']['application/json'];
