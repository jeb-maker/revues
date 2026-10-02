/** Subjects API helpers (OpenAPI-backed). */
import { createApiClient } from './client';
import type { components } from './schema';

export type SubjectSummary = components['schemas']['SubjectSummary'];
export type SubjectDetail = components['schemas']['SubjectDetail'];
export type SubjectMember = components['schemas']['SubjectMember'];
export type SubjectListResponse = components['schemas']['SubjectListResponse'];
export type SubjectWriteRequest = components['schemas']['SubjectWriteRequest'];
export type AddSubjectMemberRequest = components['schemas']['AddSubjectMemberRequest'];

type ApiError = { error?: { code?: string; message?: string } };

function errorMessage(error: unknown, fallback: string): string {
	if (error && typeof error === 'object' && 'error' in error) {
		const body = error as ApiError;
		if (body.error?.message) return body.error.message;
	}
	return fallback;
}

export async function listSubjects(opts?: {
	q?: string;
	csrfToken?: string;
}): Promise<SubjectListResponse> {
	const client = createApiClient({ csrfToken: opts?.csrfToken });
	const { data, error, response } = await client.GET('/subjects', {
		params: { query: opts?.q ? { q: opts.q } : {} }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Liste sujets: HTTP ${response.status}`));
}

export async function getSubject(id: number, csrfToken?: string): Promise<SubjectDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/subjects/{subjectId}', {
		params: { path: { subjectId: id } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Sujet: HTTP ${response.status}`));
}

export async function createSubject(
	body: SubjectWriteRequest,
	csrfToken: string
): Promise<SubjectDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/subjects', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Création: HTTP ${response.status}`));
}

export async function updateSubject(
	id: number,
	body: SubjectWriteRequest,
	csrfToken: string
): Promise<SubjectDetail> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.PATCH('/subjects/{subjectId}', {
		params: { path: { subjectId: id } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Mise à jour: HTTP ${response.status}`));
}

export async function archiveSubject(id: number, csrfToken: string): Promise<void> {
	const client = createApiClient({ csrfToken });
	const { error, response } = await client.POST('/subjects/{subjectId}/archive', {
		params: { path: { subjectId: id } }
	});
	if (response.status === 204) return;
	throw new Error(errorMessage(error, `Archivage: HTTP ${response.status}`));
}

export async function addSubjectMember(
	subjectId: number,
	body: AddSubjectMemberRequest,
	csrfToken: string
): Promise<SubjectMember> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.POST('/subjects/{subjectId}/members', {
		params: { path: { subjectId } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Ajout membre: HTTP ${response.status}`));
}

export async function removeSubjectMember(
	subjectId: number,
	userId: number,
	csrfToken: string
): Promise<void> {
	const client = createApiClient({ csrfToken });
	const { error, response } = await client.DELETE('/subjects/{subjectId}/members/{userId}', {
		params: { path: { subjectId, userId } }
	});
	if (response.status === 204) return;
	throw new Error(errorMessage(error, `Retrait membre: HTTP ${response.status}`));
}
