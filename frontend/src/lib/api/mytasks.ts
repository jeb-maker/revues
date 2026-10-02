/** Mes tâches API helpers (OpenAPI-backed). */
import { createApiClient } from './client';
import type { components } from './schema';

export type MyTask = components['schemas']['MyTask'];
export type MyTaskListResponse = components['schemas']['MyTaskListResponse'];

type ApiError = { error?: { code?: string; message?: string } };

function errorMessage(error: unknown, fallback: string): string {
	if (error && typeof error === 'object' && 'error' in error) {
		const body = error as ApiError;
		if (body.error?.message) return body.error.message;
	}
	return fallback;
}

export async function listMyTasks(opts?: {
	status?: 'pending' | 'ok' | 'nok' | 'na';
	q?: string;
	csrfToken?: string;
}): Promise<MyTaskListResponse> {
	const client = createApiClient({ csrfToken: opts?.csrfToken });
	const { data, error, response } = await client.GET('/me/tasks', {
		params: {
			query: {
				...(opts?.status ? { status: opts.status } : {}),
				...(opts?.q ? { q: opts.q } : {})
			}
		}
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Mes tâches: HTTP ${response.status}`));
}
