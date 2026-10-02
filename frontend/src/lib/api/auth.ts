/** Thin fetch helpers for auth until OpenAPI TS client (WP-005). */

export type User = {
	id: number;
	login: string;
	email: string;
	display_name: string;
	avatar_url?: string;
	role: string;
};

export type BootstrapResponse = {
	authenticated: boolean;
	csrf_token: string;
	github_oauth_enabled: boolean;
	user?: User;
	redirect?: string;
};

export type AuthSuccessResponse = {
	user: User;
	csrf_token: string;
	redirect: string;
};

type ApiError = {
	error?: { code?: string; message?: string };
};

async function parseError(res: Response): Promise<string> {
	try {
		const data = (await res.json()) as ApiError;
		if (data?.error?.message) return data.error.message;
	} catch {
		/* ignore */
	}
	return res.statusText || 'Erreur réseau';
}

export async function bootstrap(): Promise<BootstrapResponse> {
	const res = await fetch('/api/v1/bootstrap', { credentials: 'include' });
	if (!res.ok) throw new Error(await parseError(res));
	return (await res.json()) as BootstrapResponse;
}

export async function login(
	body: { email: string; password: string },
	csrf: string
): Promise<AuthSuccessResponse> {
	const res = await fetch('/api/v1/auth/login', {
		method: 'POST',
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			'X-CSRF-Token': csrf
		},
		body: JSON.stringify(body)
	});
	if (!res.ok) throw new Error(await parseError(res));
	return (await res.json()) as AuthSuccessResponse;
}

export async function register(
	body: {
		email: string;
		display_name: string;
		password: string;
		password_confirm: string;
	},
	csrf: string
): Promise<AuthSuccessResponse> {
	const res = await fetch('/api/v1/auth/register', {
		method: 'POST',
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			'X-CSRF-Token': csrf
		},
		body: JSON.stringify(body)
	});
	if (!res.ok) throw new Error(await parseError(res));
	return (await res.json()) as AuthSuccessResponse;
}

export async function logout(csrf: string): Promise<void> {
	const res = await fetch('/api/v1/auth/logout', {
		method: 'POST',
		credentials: 'include',
		headers: { 'X-CSRF-Token': csrf }
	});
	if (!res.ok && res.status !== 204) throw new Error(await parseError(res));
}
