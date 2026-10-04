import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

/** Mes tâches : gate session `show_my_tasks` (FEATURE_ASSIGN_TASKS + ≥ 2 membres). */
export const load: LayoutLoad = async ({ parent }) => {
	const { boot } = await parent();
	if (!boot.show_my_tasks) {
		redirect(302, '/runs');
	}
	return {};
};
