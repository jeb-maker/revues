/**
 * Session SPA partagée.
 *
 * `bootstrap()` est appelé UNE fois (cache module) par le `load` du layout
 * racine, qui expose le résultat dans `page.data.boot`. Les pages lisent la
 * session via `session()` ; après login/logout, `resetSession()` vide le cache
 * et la navigation suivante (`goto(..., { invalidateAll: true })`) relance le load.
 */
import { invalidateAll } from '$app/navigation';
import { page } from '$app/state';
import { bootstrap, type BootstrapResponse } from '$lib/api/auth';
import { listOrganizations } from '$lib/api/orgs';

export type ActiveOrganization = { id: number; name: string; role: string };

export type Session = BootstrapResponse & {
	/** Organisation active sur la session (null si aucune ou non authentifié). */
	organization: ActiveOrganization | null;
	/** Nombre d'organisations de l'utilisateur (sélecteur affiché si > 1). */
	organization_count: number;
	/** Admin global ou owner/admin de l'organisation active — le serveur tranche. */
	can_admin: boolean;
	/** Rôle global `editor` ou `admin` (modèles, création de sujet). */
	can_edit: boolean;
};

const ORG_GATES = new Set(['/org/new', '/org/select']);

let cached: Promise<Session> | null = null;

/** Charge (ou réutilise) la session ; une seule requête bootstrap par chargement d'app. */
export function loadSession(): Promise<Session> {
	cached ??= fetchSession().catch((e: unknown) => {
		cached = null;
		throw e;
	});
	return cached;
}

/** Oublie la session en cache (login, logout, changement d'organisation). */
export function resetSession(): void {
	cached = null;
}

/** Recharge la session sur place (ex. jeton CSRF invité après un échec de login). */
export async function refreshSession(): Promise<void> {
	resetSession();
	await invalidateAll();
}

/** Session courante exposée par le layout racine. */
export function session(): Session {
	return page.data.boot;
}

/** Session « hors ligne » utilisée quand l'API ne répond pas. */
export function offlineSession(): Session {
	return {
		authenticated: false,
		csrf_token: '',
		github_oauth_enabled: false,
		organization: null,
		organization_count: 0,
		can_admin: false,
		can_edit: false
	};
}

async function fetchSession(): Promise<Session> {
	const boot = await bootstrap();
	const role = boot.user?.role ?? '';
	let organization: ActiveOrganization | null = null;
	let organizationCount = 0;

	if (boot.authenticated && boot.user && !ORG_GATES.has(boot.redirect ?? '/')) {
		try {
			const orgs = await listOrganizations();
			organizationCount = orgs.organizations.length;
			const active = orgs.organizations.find((o) => o.id === orgs.active_organization_id);
			if (active) organization = { id: active.id, name: active.name, role: active.role };
		} catch {
			/* en-tête dégradé : pas de nom d'organisation */
		}
	}

	const orgRole = organization?.role ?? '';
	return {
		...boot,
		organization,
		organization_count: organizationCount,
		can_admin: role === 'admin' || orgRole === 'owner' || orgRole === 'admin',
		can_edit: role === 'admin' || role === 'editor'
	};
}
