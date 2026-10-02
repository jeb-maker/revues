/** Shared bootstrap gate for /admin/* pages. */
import { goto } from '$app/navigation';
import { bootstrap } from '$lib/api/auth';

export type AdminGate = {
	csrf: string;
	email: string;
};

export async function requireAdminSession(): Promise<AdminGate | null> {
	const boot = await bootstrap();
	if (!boot.authenticated || !boot.user) {
		await goto('/login');
		return null;
	}
	if (boot.redirect === '/org/new' || boot.redirect === '/org/select') {
		await goto(boot.redirect);
		return null;
	}
	return { csrf: boot.csrf_token, email: boot.user.email };
}

export async function refreshCsrf(): Promise<string> {
	const boot = await bootstrap();
	return boot.csrf_token;
}
