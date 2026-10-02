/** Organization API helpers (typed paths from generated schema). */
import { createApiClient } from './client';
import type { paths } from './schema';

export type OrganizationListResponse =
	paths['/orgs']['get']['responses']['200']['content']['application/json'];
export type OrganizationActionResponse =
	paths['/orgs']['post']['responses']['201']['content']['application/json'];
export type CreateOrganizationRequest =
	paths['/orgs']['post']['requestBody']['content']['application/json'];
export type SelectOrganizationRequest =
	paths['/orgs/active']['post']['requestBody']['content']['application/json'];

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

export async function listOrganizations(): Promise<OrganizationListResponse> {
	const client = createApiClient();
	const { data, error, response } = await client.GET('/orgs');
	if (data) return data;
	throw new Error(errorMessage(error, `Liste organisations: ${response.status}`));
}

export async function createOrganization(
	body: CreateOrganizationRequest,
	csrf: string
): Promise<OrganizationActionResponse> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.POST('/orgs', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Création organisation: ${response.status}`));
}

export async function selectActiveOrganization(
	body: SelectOrganizationRequest,
	csrf: string
): Promise<OrganizationActionResponse> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.POST('/orgs/active', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Sélection organisation: ${response.status}`));
}

export async function acceptOrganizationInvitation(
	invitationID: number,
	csrf: string
): Promise<OrganizationActionResponse> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.POST('/orgs/invitations/{invitationID}/accept', {
		params: { path: { invitationID } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Acceptation invitation: ${response.status}`));
}
