<script lang="ts">
	import { goto } from '$app/navigation';
	import { register } from '$lib/api/auth';
	import { refreshSession, resetSession, session } from '$lib/auth/session';
	import { formFieldValue, inputValue } from '$lib/mb';

	let email = $state('');
	let password = $state('');
	let passwordConfirm = $state('');
	let error = $state('');
	let loading = $state(false);

	async function onSubmit(e: Event) {
		e.preventDefault();
		const form = e.currentTarget as HTMLFormElement;
		email = formFieldValue(form, 'email').trim();
		password = formFieldValue(form, 'password');
		passwordConfirm = formFieldValue(form, 'password_confirm');
		error = '';
		loading = true;
		try {
			const res = await register(
				{ email, password, password_confirm: passwordConfirm },
				session().csrf_token
			);
			resetSession();
			await goto(res.redirect || '/', { invalidateAll: true });
		} catch (err) {
			error = err instanceof Error ? err.message : 'Inscription impossible.';
			await refreshSession().catch(() => undefined);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Inscription — Revues</title>
</svelte:head>

<div class="page page--auth">
	<div class="auth-card">
		<header class="page-header">
			<h1>Créer un compte</h1>
		</header>

		{#if error}
			<mb-alert variant="danger">{error}</mb-alert>
		{/if}

		<form class="stack-form" onsubmit={onSubmit}>
			<mb-input
				label="Email"
				type="email"
				name="email"
				autocomplete="email"
				required
				value={email}
				oninput={(e) => (email = inputValue(e))}
			></mb-input>
			<mb-input
				label="Mot de passe"
				hint="12 caractères minimum."
				type="password"
				name="password"
				autocomplete="new-password"
				required
				value={password}
				oninput={(e) => (password = inputValue(e))}
			></mb-input>
			<mb-input
				label="Confirmer le mot de passe"
				type="password"
				name="password_confirm"
				autocomplete="new-password"
				required
				value={passwordConfirm}
				oninput={(e) => (passwordConfirm = inputValue(e))}
			></mb-input>
			<mb-button type="submit" variant="primary" loading={loading}>
				{loading ? 'Création…' : 'Créer mon compte'}
			</mb-button>
		</form>

		<hr class="auth-rule" />
		<p class="auth-footer muted">Déjà un compte ? <a href="/login">Se connecter</a></p>
	</div>
</div>
