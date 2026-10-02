<script lang="ts">
	import { onMount } from 'svelte';
	import { bootstrap, logout, type BootstrapResponse } from '$lib/api/auth';
	import { getHealth } from '$lib/api';

	let boot = $state<BootstrapResponse | null>(null);
	let error = $state('');
	let health = $state<'…' | 'ok' | 'erreur'>('…');

	onMount(async () => {
		try {
			const [b, h] = await Promise.all([bootstrap(), getHealth().catch(() => null)]);
			boot = b;
			health = h?.status === 'ok' ? 'ok' : 'erreur';
		} catch (e) {
			error = e instanceof Error ? e.message : 'API indisponible';
			health = 'erreur';
		}
	});

	async function onLogout() {
		if (!boot?.csrf_token) return;
		try {
			await logout(boot.csrf_token);
			boot = await bootstrap();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Déconnexion impossible';
		}
	}
</script>

<svelte:head>
	<title>Revues</title>
</svelte:head>

<main class="shell">
	<p class="brand">Revues</p>
	<h1>API + SvelteKit</h1>
	<p class="lede">
		Scaffold rewrite — client OpenAPI + mb · auth API (login / register / OAuth / CSRF).
	</p>
	<p class="status">
		API <code>/api/v1/health</code> :
		{#if health === 'ok'}
			<mb-badge variant="success">ok</mb-badge>
		{:else if health === 'erreur'}
			<mb-badge variant="danger">hors ligne</mb-badge>
		{:else}
			<mb-spinner></mb-spinner>
		{/if}
	</p>

	{#if error}
		<p class="err">{error}</p>
	{:else if !boot}
		<p class="muted">Chargement…</p>
	{:else if boot.authenticated && boot.user}
		<p class="status">
			Connecté en tant que <strong>{boot.user.display_name}</strong> ({boot.user.email})
		</p>
		<p class="actions">
			<button type="button" onclick={onLogout}>Se déconnecter</button>
		</p>
	{:else}
		<p class="actions">
			<a href="/login">Connexion</a>
			·
			<a href="/register">Inscription</a>
		</p>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: var(--mb-font-sans, 'Segoe UI', system-ui, sans-serif);
		background: linear-gradient(
			160deg,
			var(--mb-color-canvas, #0f172a) 0%,
			#1e293b 45%,
			var(--mb-color-accent, #0f766e) 100%
		);
		color: var(--mb-color-fg, #f8fafc);
	}
	.shell {
		max-width: 40rem;
		margin: 0 auto;
		padding: 20vh 1.5rem 4rem;
	}
	.brand {
		margin: 0 0 0.75rem;
		font-size: clamp(2.5rem, 8vw, 4rem);
		font-weight: 700;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	h1 {
		margin: 0 0 0.75rem;
		font-size: 1.25rem;
		font-weight: 500;
		color: var(--mb-color-accent-fg, #99f6e4);
	}
	.lede {
		margin: 0 0 1.25rem;
		max-width: 28rem;
		line-height: 1.5;
		color: var(--mb-color-muted, #cbd5e1);
	}
	.status {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0 0 1rem;
		font-size: 0.95rem;
		color: #e2e8f0;
	}
	.status code {
		font-size: 0.85em;
	}
	.actions {
		margin: 0;
	}
	.actions a,
	.actions button {
		color: #5eead4;
		font-weight: 600;
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		font: inherit;
	}
	.err {
		color: #fca5a5;
	}
	.muted {
		color: #94a3b8;
	}
</style>
