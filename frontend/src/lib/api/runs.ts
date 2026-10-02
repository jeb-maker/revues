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
	csrfToken?: string;
}): Promise<RunListResponse> {
	const client = createApiClient({ csrfToken: opts?.csrfToken });
	const { data, error, response } = await client.GET('/runs', {
		params: {
			query: {
				...(opts?.status ? { status: opts.status } : {}),
				...(opts?.q ? { q: opts.q } : {})
			}
		}
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Liste revues: HTTP ${response.status}`));
}

export async function listSubjectRuns(
	subjectID: number,
	csrfToken?: string
): Promise<RunListResponse> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/subjects/{subjectID}/runs', {
		params: { path: { subjectID } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Revues sujet: HTTP ${response.status}`));
}

export async function listSubjectRunTemplates(
	subjectID: number,
	csrfToken?: string
): Promise<{ templates: RunTemplateSummary[]; can_launch: boolean }> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/subjects/{subjectID}/run-templates', {
		params: { path: { subjectID } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Modèles: HTTP ${response.status}`));
}

export async function createRun(
	subjectID: number,
	body: CreateRunRequest,
	csrfToken: string
): Promise<RunDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/subjects/{subjectID}/runs', {
		params: { path: { subjectID } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Lancement: HTTP ${response.status}`));
}

export async function getRun(runID: number, csrfToken?: string): Promise<RunDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/runs/{runID}', {
		params: { path: { runID } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Revue: HTTP ${response.status}`));
}

export async function completeRun(
	runID: number,
	body: CompleteRunRequest,
	csrfToken: string
): Promise<RunDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/runs/{runID}/complete', {
		params: { path: { runID } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Clôture: HTTP ${response.status}`));
}

export async function getRunItem(
	runID: number,
	itemID: number,
	csrfToken?: string
): Promise<RunItemDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/runs/{runID}/items/{itemID}', {
		params: { path: { runID, itemID } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Point: HTTP ${response.status}`));
}

export async function updateRunItem(
	runID: number,
	itemID: number,
	body: UpdateRunItemRequest,
	csrfToken: string
): Promise<RunItemDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.PATCH('/runs/{runID}/items/{itemID}', {
		params: { path: { runID, itemID } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Mise à jour: HTTP ${response.status}`));
}
