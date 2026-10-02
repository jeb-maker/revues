<script lang="ts">
	import { onMount } from 'svelte';
	import { bootstrap, logout, type BootstrapResponse } from '$lib/api/auth';

	let boot = $state<BootstrapResponse | null>(null);
	let error = $state('');

	onMount(async () => {
		try {
			boot = await bootstrap();
		} catch (e) {
			error = e instanceof Error ? e.message : 'API indisponible';
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
		Scaffold rewrite — auth API (WP-004) : login / register / OAuth / CSRF.
	</p>

	{#if error}
		<p class="err">{error}</p>
	{:else if !boot}
		<p class="muted">Chargement…</p>
	{:else if boot.authenticated && boot.user}
		<p class="status">Connecté en tant que <strong>{boot.user.display_name}</strong> ({boot.user.email})</p>
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
		font-family: 'Segoe UI', system-ui, sans-serif;
		background: linear-gradient(160deg, #0f172a 0%, #1e293b 45%, #0f766e 100%);
		color: #f8fafc;
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
		color: #99f6e4;
	}
	.lede {
		margin: 0 0 1.25rem;
		max-width: 28rem;
		line-height: 1.5;
		color: #cbd5e1;
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
	.status {
		color: #e2e8f0;
	}
	.err {
		color: #fca5a5;
	}
	.muted {
		color: #94a3b8;
	}
</style>
