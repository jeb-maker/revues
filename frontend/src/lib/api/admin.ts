/** Admin org API helpers (typed paths from generated schema). */
import { createApiClient } from './client';
import type { paths } from './schema';

export type AllowedEmailListResponse =
	paths['/admin/allowed-emails']['get']['responses']['200']['content']['application/json'];
export type AllowedEmailWriteRequest =
	paths['/admin/allowed-emails']['post']['requestBody']['content']['application/json'];
export type OrganizationMemberListResponse =
	paths['/admin/members']['get']['responses']['200']['content']['application/json'];
export type UpdateOrganizationMemberRoleRequest =
	paths['/admin/members/{userId}']['patch']['requestBody']['content']['application/json'];
export type AdminTeamListResponse =
	paths['/admin/teams']['get']['responses']['200']['content']['application/json'];
export type CreateAdminTeamRequest =
	paths['/admin/teams']['post']['requestBody']['content']['application/json'];
export type AdminTeamDetailResponse =
	paths['/admin/teams/{teamId}']['get']['responses']['200']['content']['application/json'];
export type LeadPolicies =
	paths['/admin/settings/policies']['get']['responses']['200']['content']['application/json'];

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

export async function listAllowedEmails(): Promise<AllowedEmailListResponse> {
	const client = createApiClient();
	const { data, error, response } = await client.GET('/admin/allowed-emails');
	if (data) return data;
	throw new Error(errorMessage(error, `Emails autorisés: ${response.status}`));
}

export async function createAllowedEmail(
	body: AllowedEmailWriteRequest,
	csrf: string
): Promise<AllowedEmailListResponse['emails'][number]> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.POST('/admin/allowed-emails', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Ajout email: ${response.status}`));
}

export async function deleteAllowedEmail(email: string, csrf: string): Promise<void> {
	const client = createApiClient({ csrfToken: csrf });
	const { error, response } = await client.DELETE('/admin/allowed-emails/{email}', {
		params: { path: { email } }
	});
	if (response.status === 204) return;
	throw new Error(errorMessage(error, `Retrait email: ${response.status}`));
}

export async function listOrganizationMembers(): Promise<OrganizationMemberListResponse> {
	const client = createApiClient();
	const { data, error, response } = await client.GET('/admin/members');
	if (data) return data;
	throw new Error(errorMessage(error, `Membres: ${response.status}`));
}

export async function updateOrganizationMemberRole(
	userId: number,
	body: UpdateOrganizationMemberRoleRequest,
	csrf: string
): Promise<OrganizationMemberListResponse['members'][number]> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.PATCH('/admin/members/{userId}', {
		params: { path: { userId } },
		body
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Mise à jour rôle: ${response.status}`));
}

export async function listAdminTeams(): Promise<AdminTeamListResponse> {
	const client = createApiClient();
	const { data, error, response } = await client.GET('/admin/teams');
	if (data) return data;
	throw new Error(errorMessage(error, `Équipes: ${response.status}`));
}

export async function createAdminTeam(
	body: CreateAdminTeamRequest,
	csrf: string
): Promise<AdminTeamListResponse['teams'][number]> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.POST('/admin/teams', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Création équipe: ${response.status}`));
}

export async function getAdminTeam(teamId: number): Promise<AdminTeamDetailResponse> {
	const client = createApiClient();
	const { data, error, response } = await client.GET('/admin/teams/{teamId}', {
		params: { path: { teamId } }
	});
	if (data) return data;
	throw new Error(errorMessage(error, `Équipe: ${response.status}`));
}

export async function addAdminTeamMember(
	teamId: number,
	userId: number,
	csrf: string
): Promise<void> {
	const client = createApiClient({ csrfToken: csrf });
	const { error, response } = await client.POST('/admin/teams/{teamId}/members', {
		params: { path: { teamId } },
		body: { user_id: userId }
	});
	if (response.status === 204) return;
	throw new Error(errorMessage(error, `Ajout membre équipe: ${response.status}`));
}

export async function removeAdminTeamMember(
	teamId: number,
	userId: number,
	csrf: string
): Promise<void> {
	const client = createApiClient({ csrfToken: csrf });
	const { error, response } = await client.DELETE('/admin/teams/{teamId}/members/{userId}', {
		params: { path: { teamId, userId } }
	});
	if (response.status === 204) return;
	throw new Error(errorMessage(error, `Retrait membre équipe: ${response.status}`));
}

export async function getLeadPolicies(): Promise<LeadPolicies> {
	const client = createApiClient();
	const { data, error, response } = await client.GET('/admin/settings/policies');
	if (data) return data;
	throw new Error(errorMessage(error, `Politiques: ${response.status}`));
}

export async function updateLeadPolicies(body: LeadPolicies, csrf: string): Promise<LeadPolicies> {
	const client = createApiClient({ csrfToken: csrf });
	const { data, error, response } = await client.PUT('/admin/settings/policies', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Enregistrement politiques: ${response.status}`));
}

export type SMTPSettings =
	paths['/admin/settings/smtp']['get']['responses']['200']['content']['application/json'];
export type SMTPSettingsUpdate =
	paths['/admin/settings/smtp']['put']['requestBody']['content']['application/json'];
export type IntegrationsOverview =
	paths['/admin/integrations']['get']['responses']['200']['content']['application/json'];
export type IntegrationSummary = IntegrationsOverview['items'][number];

export async function getAdminSMTPSettings(csrfToken?: string): Promise<SMTPSettings> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/admin/settings/smtp');
	if (data) return data;
	throw new Error(errorMessage(error, `SMTP: HTTP ${response.status}`));
}

export async function putAdminSMTPSettings(
	body: SMTPSettingsUpdate,
	csrfToken: string
): Promise<SMTPSettings> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.PUT('/admin/settings/smtp', { body });
	if (data) return data;
	throw new Error(errorMessage(error, `Enregistrement SMTP: HTTP ${response.status}`));
}

export async function deleteAdminSMTPSettings(csrfToken: string): Promise<void> {
	const client = createApiClient({ csrfToken });
	const { error, response } = await client.DELETE('/admin/settings/smtp');
	if (response.ok) return;
	throw new Error(errorMessage(error, `Suppression SMTP: HTTP ${response.status}`));
}

export async function postAdminSMTPTest(csrfToken: string, recipient?: string): Promise<void> {
	const client = createApiClient({ csrfToken });
	const { error, response } = await client.POST('/admin/settings/smtp/test', {
		body: recipient ? { recipient } : {}
	});
	if (response.status === 204) return;
	throw new Error(errorMessage(error, `Test SMTP: HTTP ${response.status}`));
}

export async function listAdminIntegrations(csrfToken?: string): Promise<IntegrationsOverview> {
	const client = createApiClient({ csrfToken });
	const { data, error, response } = await client.GET('/admin/integrations');
	if (data) return data;
	throw new Error(errorMessage(error, `Intégrations: HTTP ${response.status}`));
}
