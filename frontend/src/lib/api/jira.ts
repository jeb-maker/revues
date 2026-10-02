/** Run item Jira link/create helpers. */
import { createApiClient } from './client';
import type { components, paths } from './schema';

export type JiraLink = components['schemas']['JiraLink'];
export type RunItemJira =
	paths['/runs/{runId}/items/{itemId}/jira']['get']['responses']['200']['content']['application/json'];
export type JiraLinkRequest =
	paths['/runs/{runId}/items/{itemId}/jira']['put']['requestBody']['content']['application/json'];
export type JiraCreateRequest = NonNullable<
	paths['/runs/{runId}/items/{itemId}/jira']['post']['requestBody']
>['content']['application/json'];

type ApiError = { error?: { code?: string; message?: string } };

function errorMessage(error: unknown, fallback: string): string {
	if (error && typeof error === 'object' && 'error' in error) {
		const body = error as ApiError;
		if (body.error?.message) return body.error.message;
	}
	return fallback;
}

export async function getRunItemJira(
	runId: number,
	itemId: number,
	csrfToken?: string
): Promise<RunItemJira> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/runs/{runId}/items/{itemId}/jira', {
		params: { path: { runId, itemId } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Jira: HTTP ${response.status}`));
}

export async function putRunItemJiraLink(
	runId: number,
	itemId: number,
	body: JiraLinkRequest,
	csrfToken: string
): Promise<JiraLink> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.PUT('/runs/{runId}/items/{itemId}/jira', {
		params: { path: { runId, itemId } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Lien Jira: HTTP ${response.status}`));
}

export async function postRunItemJiraCreate(
	runId: number,
	itemId: number,
	body: JiraCreateRequest,
	csrfToken: string
): Promise<JiraLink> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/runs/{runId}/items/{itemId}/jira', {
		params: { path: { runId, itemId } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Création Jira: HTTP ${response.status}`));
}
