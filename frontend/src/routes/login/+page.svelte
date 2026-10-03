<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { login } from '$lib/api/auth';
	import { refreshSession, resetSession, session } from '$lib/auth/session';
	import { loginErrorMessage } from '$lib/auth/messages';
	import { inputValue } from '$lib/mb';

	let email = $state('');
	let password = $state('');
	let error = $state(loginErrorMessage(page.url.searchParams.get('error') ?? ''));
	let loading = $state(false);

	const githubEnabled = session().github_oauth_enabled;

	async function onSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await login({ email, password }, session().csrf_token);
			resetSession();
			await goto(res.redirect || '/runs', { invalidateAll: true });
		} catch (err) {
			error = err instanceof Error ? err.message : 'Connexion impossible.';
			await refreshSession().catch(() => undefined);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Connexion — Revues</title>
</svelte:head>

<div class="page page--narrow">
	<header class="page-header">
		<h1>Connexion</h1>
		<p class="lede">Accédez à vos check-lists avec une session sécurisée.</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	<form class="stack-form" onsubmit={onSubmit}>
		<mb-input
			label="Email"
			type="email"
			name="email"
			autocomplete="username"
			required
			value={email}
			oninput={(e) => (email = inputValue(e))}
		></mb-input>
		<mb-input
			label="Mot de passe"
			type="password"
			name="password"
			autocomplete="current-password"
			required
			value={password}
			oninput={(e) => (password = inputValue(e))}
		></mb-input>
		<mb-button type="submit" variant="primary" disabled={loading}>
			{loading ? 'Connexion…' : 'Se connecter'}
		</mb-button>
	</form>

	{#if githubEnabled}
		<p class="muted">ou</p>
		<p><a href="/auth/github/start">Continuer avec GitHub</a></p>
	{/if}

	<p class="muted">Pas encore de compte ? <a href="/register">Créer un compte</a></p>
</div>
