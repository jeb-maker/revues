/** Run item attachments API helpers. */
import { API_V1_BASE, createApiClient } from './client';
import type { components } from './schema';

export type Attachment = components['schemas']['Attachment'];

type ApiError = { error?: { code?: string; message?: string } };

function errorMessage(error: unknown, fallback: string): string {
	if (error && typeof error === 'object' && 'error' in error) {
		const body = error as ApiError;
		if (body.error?.message) return body.error.message;
	}
	return fallback;
}

export async function getRunItemAttachment(
	runId: number,
	itemId: number,
	csrfToken?: string
): Promise<Attachment> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET(
		'/runs/{runId}/items/{itemId}/attachments',
		{ params: { path: { runId, itemId } } }
	);
	if (data) return data;
	throw new Error(errorMessage(error, `Pièce jointe: HTTP ${response.status}`));
}

/** Multipart upload — uses FormData + CSRF header (not JSON body). */
export async function uploadRunItemAttachment(
	runId: number,
	itemId: number,
	file: File,
	csrfToken: string
): Promise<Attachment> {
	const fd = new FormData();
	fd.append('file', file);
	const res = await fetch(
		`${API_V1_BASE}/runs/${runId}/items/${itemId}/attachments`,
		{
			method: 'POST',
			credentials: 'include',
			headers: { 'X-CSRF-Token': csrfToken },
			body: fd
		}
	);
	const body = await res.json().catch(() => null);
	if (!res.ok) {
		throw new Error(errorMessage(body, `Upload: HTTP ${res.status}`));
	}
	return body as Attachment;
}

export function attachmentDownloadURL(
	runId: number,
	itemId: number,
	attachmentId: number
): string {
	return `${API_V1_BASE}/runs/${runId}/items/${itemId}/attachments/${attachmentId}`;
}
