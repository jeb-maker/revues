// SPA : Go sert le build statique ; ce load tourne uniquement côté navigateur.
import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';
import { loadSession, offlineSession, type Session } from '$lib/auth/session';
import { ensureMb } from '$lib/mb';

export const ssr = false;
export const prerender = true;

/** Routes accessibles sans session. */
const PUBLIC_PATHS = new Set(['/login', '/register']);
/** Pages d'entrée imposées par le serveur tant qu'aucune organisation n'est active. */
const ORG_GATES = new Set(['/org/new', '/org/select']);

type Data = { boot: Session; apiError?: string };

export const load: LayoutLoad = async ({ url }): Promise<Data> => {
	let boot: Session;
	try {
		[boot] = await Promise.all([loadSession(), ensureMb().catch(() => undefined)]);
	} catch (e) {
		return {
			boot: offlineSession(),
			apiError: e instanceof Error ? e.message : 'API indisponible.'
		};
	}

	const path = url.pathname.replace(/\/+$/, '') || '/';
	const isPublic = PUBLIC_PATHS.has(path) || path.startsWith('/auth/');

	if (!boot.authenticated) {
		if (!isPublic) redirect(302, '/login');
		return { boot };
	}

	const target = boot.redirect || '/';
	if (PUBLIC_PATHS.has(path)) redirect(302, target);
	if (ORG_GATES.has(target) && !path.startsWith('/org/')) redirect(302, target);
	return { boot };
};
