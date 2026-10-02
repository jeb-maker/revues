/** Notion integration API helpers. */
import { createApiClient } from './client';
import type { paths } from './schema';

type ApiErrorBody = {
	error?: { code?: string; message?: string };
};

function errorMessage(payload: unknown, fallback: string): string {
	if (payload && typeof payload === 'object' && 'error' in payload) {
		const err = (payload as ApiErrorBody).error;
		if (err?.message) return err.message;
	}
	return fallback;
}

export type NotionSettings =
	paths['/admin/integrations/notion']['get']['responses']['200']['content']['application/json'];
export type NotionSettingsUpdate =
	paths['/admin/integrations/notion']['put']['requestBody']['content']['application/json'];
export type NotionTestResponse =
	paths['/admin/integrations/notion/test']['post']['responses']['200']['content']['application/json'];
export type NotionImportRequest =
	paths['/templates/notion-import']['post']['requestBody']['content']['application/json'];
export type NotionImportResponse =
	paths['/templates/notion-import']['post']['responses']['200']['content']['application/json'];
export type NotionExportResponse =
	paths['/runs/{runId}/notion-export']['post']['responses']['200']['content']['application/json'];

export async function getAdminNotionSettings(csrfToken?: string): Promise<NotionSettings> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/admin/integrations/notion');
	if (data) return data;
	throw new Error(errorMessage(error, `Notion: HTTP ${response.status}`));
}

export async function putAdminNotionSettings(
	body: NotionSettingsUpdate,
	csrfToken: string
): Promise<NotionSettings> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.PUT('/admin/integrations/notion', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Enregistrement Notion: HTTP ${response.status}`));
}

export async function deleteAdminNotionSettings(csrfToken: string): Promise<void> {
	const client = createApiClient({ csrfToken });
	const { error, response } = await client.DELETE('/admin/integrations/notion');
	if (response.ok) return;
	throw new Error(errorMessage(error, `Suppression Notion: HTTP ${response.status}`));
}

export async function postAdminNotionTest(csrfToken: string): Promise<NotionTestResponse> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/admin/integrations/notion/test');
	if (data) return data;
	throw new Error(errorMessage(error, `Test Notion: HTTP ${response.status}`));
}

export async function postTemplatesNotionImport(
	body: NotionImportRequest,
	csrfToken: string
): Promise<NotionImportResponse> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/templates/notion-import', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Import Notion: HTTP ${response.status}`));
}

export async function postRunNotionExport(
	runId: number,
	csrfToken: string
): Promise<NotionExportResponse> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/runs/{runId}/notion-export', {
		params: { path: { runId } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Export Notion: HTTP ${response.status}`));
}
