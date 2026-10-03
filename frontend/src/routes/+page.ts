import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

/** Accueil connecté = hub revues (décisions produit). */
export const load: PageLoad = async ({ parent }) => {
	const { boot } = await parent();
	if (boot.authenticated) {
		redirect(302, boot.redirect && boot.redirect !== '/' ? boot.redirect : '/runs');
	}
	redirect(302, '/login');
};
