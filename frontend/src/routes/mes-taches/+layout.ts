import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

/** Mes tâches : visible seulement si ≥ 2 membres org. */
export const load: LayoutLoad = async ({ parent }) => {
	const { boot } = await parent();
	if (!boot.show_my_tasks) {
		redirect(302, '/runs');
	}
	return {};
};
