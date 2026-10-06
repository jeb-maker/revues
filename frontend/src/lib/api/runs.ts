/** Runs / items API helpers (OpenAPI-backed). */
import { createApiClient } from './client';
import type { components } from './schema';

export type RunSummary = components['schemas']['RunSummary'];
export type RunDetail = components['schemas']['RunDetail'];
export type RunItem = components['schemas']['RunItem'];
export type RunItemDetail = components['schemas']['RunItemDetail'];
export type RunListResponse = components['schemas']['RunListResponse'];
export type RunTemplateSummary = components['schemas']['RunTemplateSummary'];
export type CreateRunRequest = components['schemas']['CreateRunRequest'];
export type UpdateRunItemRequest = components['schemas']['UpdateRunItemRequest'];
export type CompleteRunRequest = components['schemas']['CompleteRunRequest'];
export type ConfluenceLink = components['schemas']['ConfluenceLink'];

type ApiError = { error?: { code?: string; message?: string } };

function errorMessage(error: unknown, fallback: string): string {
	if (error && typeof error === 'object' && 'error' in error) {
		const body = error as ApiError;
		if (body.error?.message) return body.error.message;
	}
	return fallback;
}

export async function listRuns(opts?: {
	status?: 'draft' | 'in_progress' | 'done' | 'overdue';
	q?: string;
	limit?: number;
	offset?: number;
	csrfToken?: string;
}): Promise<RunListResponse> {
	const client = createApiClient({ csrfToken: opts?.csrfToken });
	const { data, error, response } = await client.GET('/runs', {
		params: {
			query: {
				...(opts?.status ? { status: opts.status } : {}),
				...(opts?.q ? { q: opts.q } : {}),
				...(opts?.limit != null ? { limit: opts.limit } : {}),
				...(opts?.offset != null ? { offset: opts.offset } : {})
			}
		}
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Liste revues: HTTP ${response.status}`));
}

export async function listSubjectRuns(
	subjectId: number,
	csrfToken?: string
): Promise<RunListResponse> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/subjects/{subjectId}/runs', {
		params: { path: { subjectId } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Revues sujet: HTTP ${response.status}`));
}

export async function listSubjectRunTemplates(
	subjectId: number,
	csrfToken?: string
): Promise<{ templates: RunTemplateSummary[]; can_launch: boolean }> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/subjects/{subjectId}/run-templates', {
		params: { path: { subjectId } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Modèles: HTTP ${response.status}`));
}

export async function createRun(
	subjectId: number,
	body: CreateRunRequest,
	csrfToken: string
): Promise<RunDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/subjects/{subjectId}/runs', {
		params: { path: { subjectId } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Lancement: HTTP ${response.status}`));
}

export async function getRun(runId: number, csrfToken?: string): Promise<RunDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/runs/{runId}', {
		params: { path: { runId } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Revue: HTTP ${response.status}`));
}

export async function completeRun(
	runId: number,
	body: CompleteRunRequest,
	csrfToken: string
): Promise<RunDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/runs/{runId}/complete', {
		params: { path: { runId } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Clôture: HTTP ${response.status}`));
}

export async function publishRunConfluence(
	runId: number,
	csrfToken: string
): Promise<ConfluenceLink> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/runs/{runId}/confluence', {
		params: { path: { runId } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Confluence: HTTP ${response.status}`));
}

export async function getRunItem(
	runId: number,
	itemId: number,
	csrfToken?: string
): Promise<RunItemDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/runs/{runId}/items/{itemId}', {
		params: { path: { runId, itemId } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Point: HTTP ${response.status}`));
}

/** Thrown by `updateRunItem`; `status` lets callers react to 409 (verrou optimiste `updated_at`). */
export class UpdateRunItemError extends Error {
	status: number;
	constructor(message: string, status: number) {
		super(message);
		this.name = 'UpdateRunItemError';
		this.status = status;
	}
}

/**
 * PATCH a run item. Pass `body.updated_at` (the `RunItem.updated_at` the user saw) to get a
 * 409 instead of silently overwriting a concurrent change.
 */
export async function updateRunItem(
	runId: number,
	itemId: number,
	body: UpdateRunItemRequest,
	csrfToken: string
): Promise<RunItemDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.PATCH('/runs/{runId}/items/{itemId}', {
		params: { path: { runId, itemId } },
		body
	});
	if (data) return data;
	throw new UpdateRunItemError(
		errorMessage(error, `Mise à jour: HTTP ${response.status}`),
		response.status
	);
}
