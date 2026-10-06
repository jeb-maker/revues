/** Confluence Cloud admin config (OpenAPI-backed). */
import { createApiClient } from './client';
import type { paths } from './schema';

export type ConfluenceSettings =
	paths['/admin/integrations/confluence']['get']['responses']['200']['content']['application/json'];
export type ConfluenceSettingsUpdate =
	paths['/admin/integrations/confluence']['put']['requestBody']['content']['application/json'];

type ApiError = { error?: { code?: string; message?: string } };

function errMsg(error: unknown, fallback: string): string {
	if (error && typeof error === 'object' && 'error' in error) {
		const body = error as ApiError;
		if (body.error?.message) return body.error.message;
	}
	return fallback;
}

export async function getAdminConfluenceSettings(csrfToken?: string): Promise<ConfluenceSettings> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/admin/integrations/confluence');
	if (data) return data;
	throw new Error(errMsg(error, `Confluence: HTTP ${response.status}`));
}

export async function putAdminConfluenceSettings(
	body: ConfluenceSettingsUpdate,
	csrfToken: string
): Promise<ConfluenceSettings> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.PUT('/admin/integrations/confluence', { body });
	if (data) return data;
	throw new Error(errMsg(error, `Enregistrement: HTTP ${response.status}`));
}

export async function deleteAdminConfluenceSettings(csrfToken: string): Promise<void> {
	const client = createApiClient({ csrfToken });
	const { error, response } = await client.DELETE('/admin/integrations/confluence');
	if (response.ok) return;
	throw new Error(errMsg(error, `Suppression: HTTP ${response.status}`));
}

export async function postAdminConfluenceTest(csrfToken: string): Promise<void> {
	const client = createApiClient({ csrfToken });
	const { error, response } = await client.POST('/admin/integrations/confluence/test');
	if (response.status === 204) return;
	throw new Error(errMsg(error, `Test: HTTP ${response.status}`));
}
