<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { bootstrap, login, type BootstrapResponse } from '$lib/api/auth';
	import { loginErrorMessage } from '$lib/auth/messages';

	let email = $state('');
	let password = $state('');
	let csrf = $state('');
	let githubEnabled = $state(false);
	let error = $state('');
	let loading = $state(false);
	let ready = $state(false);

	onMount(async () => {
		try {
			const boot: BootstrapResponse = await bootstrap();
			if (boot.authenticated) {
				await goto(boot.redirect || '/');
				return;
			}
			csrf = boot.csrf_token;
			githubEnabled = boot.github_oauth_enabled;
			const q = $page.url.searchParams.get('error');
			if (q) error = loginErrorMessage(q);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger la page de connexion.';
		} finally {
			ready = true;
		}
	});

	async function onSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await login({ email, password }, csrf);
			await goto(res.redirect || '/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Connexion impossible.';
			// Refresh guest CSRF after failed attempt
			try {
				const boot = await bootstrap();
				csrf = boot.csrf_token;
			} catch {
				/* ignore */
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Connexion — Revues</title>
	<link rel="stylesheet" href="/static/vendor/jeb-maker-mb/tokens/tokens-core.css" />
	<link rel="stylesheet" href="/static/vendor/jeb-maker-mb/mb-bridge.css" />
	<script type="module" src="/static/vendor/jeb-maker-mb/mb-boot.js"></script>
</svelte:head>

<main class="auth">
	<p class="brand">Revues</p>
	<h1>Connexion</h1>
	<p class="lede">Accédez à vos check-lists avec une session sécurisée.</p>

	{#if !ready}
		<p class="muted">Chargement…</p>
	{:else}
		{#if error}
			<mb-alert variant="danger" role="alert">{error}</mb-alert>
		{/if}

		<form class="form" onsubmit={onSubmit}>
			<label class="field">
				<span>Email</span>
				<mb-input
					type="email"
					name="email"
					autocomplete="username"
					required
					value={email}
					oninput={(e: Event) => {
						email = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			<label class="field">
				<span>Mot de passe</span>
				<mb-input
					type="password"
					name="password"
					autocomplete="current-password"
					required
					value={password}
					oninput={(e: Event) => {
						password = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			<mb-button type="submit" variant="primary" disabled={loading || !csrf}>
				{loading ? 'Connexion…' : 'Se connecter'}
			</mb-button>
		</form>

		{#if githubEnabled}
			<p class="divider">ou</p>
			<p>
				<a class="oauth" href="/auth/github/start">Continuer avec GitHub</a>
			</p>
		{/if}

		<p class="footer">
			Pas encore de compte ? <a href="/register">Créer un compte</a>
		</p>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background:
			radial-gradient(ellipse 80% 50% at 10% 0%, rgba(15, 118, 110, 0.35), transparent 55%),
			linear-gradient(165deg, #0f172a 0%, #1e293b 50%, #134e4a 100%);
		color: #f8fafc;
	}
	.auth {
		max-width: 24rem;
		margin: 0 auto;
		padding: 12vh 1.25rem 3rem;
	}
	.brand {
		margin: 0 0 0.5rem;
		font-size: clamp(2.25rem, 7vw, 3.25rem);
		font-weight: 700;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	h1 {
		margin: 0 0 0.5rem;
		font-size: 1.15rem;
		font-weight: 500;
		color: #99f6e4;
	}
	.lede {
		margin: 0 0 1.5rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.875rem;
		color: #e2e8f0;
	}
	.divider {
		margin: 1.25rem 0 0.75rem;
		text-align: center;
		color: #94a3b8;
		font-size: 0.85rem;
	}
	.oauth {
		color: #5eead4;
		font-weight: 600;
	}
	.footer {
		margin-top: 1.5rem;
		font-size: 0.9rem;
		color: #94a3b8;
	}
	.footer a {
		color: #5eead4;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
