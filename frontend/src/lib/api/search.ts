/** Global search API helpers (OpenAPI-backed). */
import { createApiClient } from './client';
import type { components } from './schema';

export type SearchResult = components['schemas']['SearchResult'];
export type SearchResponse = components['schemas']['SearchResponse'];
export type SearchResultKind = components['schemas']['SearchResultKind'];

type ApiError = { error?: { code?: string; message?: string } };

function errorMessage(error: unknown, fallback: string): string {
	if (error && typeof error === 'object' && 'error' in error) {
		const body = error as ApiError;
		if (body.error?.message) return body.error.message;
	}
	return fallback;
}

export async function globalSearch(opts: {
	q: string;
	limit?: number;
	csrfToken?: string;
}): Promise<SearchResponse> {
	const client = createApiClient({ csrfToken: opts.csrfToken });
	const { data, error, response } = await client.GET('/search', {
		params: {
			query: {
				q: opts.q,
				...(opts.limit != null ? { limit: opts.limit } : {})
			}
		}
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Recherche: HTTP ${response.status}`));
}
