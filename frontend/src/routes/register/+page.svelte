<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap, register, type BootstrapResponse } from '$lib/api/auth';

	let email = $state('');
	let displayName = $state('');
	let password = $state('');
	let passwordConfirm = $state('');
	let csrf = $state('');
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
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger la page.';
		} finally {
			ready = true;
		}
	});

	async function onSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await register(
				{
					email,
					display_name: displayName,
					password,
					password_confirm: passwordConfirm
				},
				csrf
			);
			await goto(res.redirect || '/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Inscription impossible.';
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
	<title>Inscription — Revues</title>
	<link rel="stylesheet" href="/static/vendor/jeb-maker-mb/tokens/tokens-core.css" />
	<link rel="stylesheet" href="/static/vendor/jeb-maker-mb/mb-bridge.css" />
	<script type="module" src="/static/vendor/jeb-maker-mb/mb-boot.js"></script>
</svelte:head>

<main class="auth">
	<p class="brand">Revues</p>
	<h1>Créer un compte</h1>
	<p class="lede">Inscription email + mot de passe (sous réserve d’autorisation).</p>

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
					autocomplete="email"
					required
					value={email}
					oninput={(e: Event) => {
						email = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			<label class="field">
				<span>Nom affiché</span>
				<mb-input
					type="text"
					name="display_name"
					autocomplete="name"
					required
					value={displayName}
					oninput={(e: Event) => {
						displayName = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			<label class="field">
				<span>Mot de passe</span>
				<mb-input
					type="password"
					name="password"
					autocomplete="new-password"
					required
					minlength="8"
					value={password}
					oninput={(e: Event) => {
						password = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			<label class="field">
				<span>Confirmer le mot de passe</span>
				<mb-input
					type="password"
					name="password_confirm"
					autocomplete="new-password"
					required
					minlength="8"
					value={passwordConfirm}
					oninput={(e: Event) => {
						passwordConfirm = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			<mb-button type="submit" variant="primary" disabled={loading || !csrf}>
				{loading ? 'Création…' : 'Créer mon compte'}
			</mb-button>
		</form>

		<p class="footer">
			Déjà un compte ? <a href="/login">Se connecter</a>
		</p>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background:
			radial-gradient(ellipse 80% 50% at 90% 0%, rgba(15, 118, 110, 0.3), transparent 55%),
			linear-gradient(165deg, #0f172a 0%, #1e293b 50%, #134e4a 100%);
		color: #f8fafc;
	}
	.auth {
		max-width: 24rem;
		margin: 0 auto;
		padding: 10vh 1.25rem 3rem;
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
