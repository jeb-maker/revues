<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { login } from '$lib/api/auth';
	import { refreshSession, resetSession, session } from '$lib/auth/session';
	import { loginErrorMessage } from '$lib/auth/messages';
	import { formFieldValue, inputValue } from '$lib/mb';

	let email = $state('');
	let password = $state('');
	let error = $state(loginErrorMessage(page.url.searchParams.get('error') ?? ''));
	let loading = $state(false);

	const githubEnabled = session().github_oauth_enabled;

	async function onSubmit(e: Event) {
		e.preventDefault();
		const form = e.currentTarget as HTMLFormElement;
		// Relire les champs au submit : l'autofill navigateur ne déclenche pas toujours `input`.
		email = formFieldValue(form, 'email').trim();
		password = formFieldValue(form, 'password');
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

<div class="page page--auth">
	<div class="auth-card">
		<header class="page-header">
			<h1>Connexion</h1>
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
			<mb-button type="submit" variant="primary" loading={loading}>
				{loading ? 'Connexion…' : 'Se connecter'}
			</mb-button>
		</form>

		{#if githubEnabled}
			<div class="auth-alt" role="separator">
				<span>ou</span>
			</div>
			<mb-button href="/auth/github/start" variant="secondary">Se connecter avec GitHub</mb-button>
		{/if}

		<hr class="auth-rule" />
		<p class="auth-footer muted">Pas encore de compte ? <a href="/register">Créer un compte</a></p>
	</div>
</div>
