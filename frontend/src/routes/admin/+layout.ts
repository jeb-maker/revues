import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

/** Garde client : les membres non admin n'entrent pas dans /admin/*. */
export const load: LayoutLoad = async ({ parent }) => {
	const { boot } = await parent();
	if (!boot.can_admin) {
		redirect(302, '/runs');
	}
	return {};
};
