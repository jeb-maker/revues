import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

/** Lecteurs : pas d'accès catalogue modèles (décisions). */
export const load: LayoutLoad = async ({ parent }) => {
	const { boot } = await parent();
	if (!boot.show_modeles) {
		redirect(302, '/runs');
	}
	return {};
};
