/** Typed helpers for checklist templates API. */
import { createApiClient } from './client';
import type { paths } from './schema';

export type TemplateSummary =
	paths['/templates']['get']['responses']['200']['content']['application/json']['templates'][number];
export type TemplateDetail =
	paths['/templates/{templateId}']['get']['responses']['200']['content']['application/json'];
export type TemplateWriteRequest =
	paths['/templates']['post']['requestBody']['content']['application/json'];
export type TemplateItemInput = TemplateWriteRequest['items'][number];

type ApiError = {
	error?: { code?: string; message?: string };
};

function messageFromError(error: unknown, fallback: string): string {
	if (error && typeof error === 'object' && 'error' in error) {
		const body = error as ApiError;
		if (body.error?.message) return body.error.message;
	}
	return fallback;
}

export async function listTemplates(csrf: string, q = ''): Promise<TemplateSummary[]> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.GET('/templates', {
		params: { query: q ? { q } : {} }
	});
	if (data) return data.templates ?? [];
	if (response?.status === 401) throw new Error('unauthenticated');
	throw new Error(messageFromError(error, 'Impossible de charger les modèles.'));
}

export async function getTemplate(csrf: string, id: number): Promise<TemplateDetail> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.GET('/templates/{templateId}', {
		params: { path: { templateId: id } }
	});
	if (data) return data;
	if (response?.status === 401) throw new Error('unauthenticated');
	if (response?.status === 404) throw new Error('Modèle introuvable.');
	throw new Error(messageFromError(error, 'Impossible de charger le modèle.'));
}

export async function createTemplate(
	csrf: string,
	body: TemplateWriteRequest
): Promise<TemplateDetail> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error } = await client.POST('/templates', { body });
	if (data) return data;
	throw new Error(messageFromError(error, 'Création impossible.'));
}

export async function saveTemplate(
	csrf: string,
	id: number,
	body: TemplateWriteRequest
): Promise<TemplateDetail> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error } = await client.PUT('/templates/{templateId}', {
		params: { path: { templateId: id } },
		body
	});
	if (data) return data;
	throw new Error(messageFromError(error, 'Enregistrement impossible.'));
}

export async function archiveTemplate(csrf: string, id: number): Promise<void> {
	const client = createApiClient({ csrfToken: csrf });
	const { error, response } = await client.DELETE('/templates/{templateId}', {
		params: { path: { templateId: id } }
	});
	if (response?.status === 204) return;
	throw new Error(messageFromError(error, 'Archivage impossible.'));
}
