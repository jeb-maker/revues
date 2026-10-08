<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { deleteMeAtlassian, getMeAtlassian, type MeAtlassian } from '$lib/api/auth';
	import { session } from '$lib/auth/session';

	const csrf = session().csrf_token;

	let status = $state<MeAtlassian | null>(null);
	let error = $state('');
	let message = $state('');
	let loading = $state(true);
	let busy = $state(false);

	onMount(async () => {
		const connected = page.url.searchParams.get('connected');
		const err = page.url.searchParams.get('error');
		if (connected === '1') {
			message = 'Compte Atlassian connecté.';
		}
		if (err) {
			error = err;
		}
		try {
			status = await getMeAtlassian(csrf);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Statut Atlassian indisponible.';
		} finally {
			loading = false;
		}
	});

	async function onDisconnect() {
		busy = true;
		error = '';
		message = '';
		try {
			await deleteMeAtlassian(csrf);
			status = await getMeAtlassian(csrf);
			message = 'Compte Atlassian déconnecté.';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Déconnexion impossible.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>Atlassian — Revues</title>
</svelte:head>

<div class="page page--narrow">
	<header class="page-header">
		<p class="crumbs"><a href="/runs">Revues</a> · Compte</p>
		<h1>Atlassian</h1>
		<p class="lede">
			Connectez votre compte Atlassian pour créer et lier des issues Jira en votre nom.
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if status}
		<mb-card>
			<h2 slot="header">Connexion Jira (OAuth)</h2>
			{#if !status.enabled}
				<p class="muted">
					OAuth Atlassian n'est pas configuré sur ce serveur. Contactez un administrateur.
				</p>
			{:else if status.connected}
				<p>
					Connecté
					{#if status.account_email}
						en tant que <strong>{status.account_email}</strong>
					{/if}
					{#if status.site_url}
						· <a href={status.site_url} target="_blank" rel="noopener noreferrer">{status.site_url}</a>
					{/if}
				</p>
				<div slot="footer">
					<mb-button variant="danger" loading={busy} onclick={onDisconnect}>
						Déconnecter Atlassian
					</mb-button>
				</div>
			{:else}
				<p class="muted">Aucun compte Atlassian lié à votre profil Revues.</p>
				<div slot="footer">
					<mb-button variant="primary" href="/auth/atlassian/start">
						Connecter Atlassian
					</mb-button>
				</div>
			{/if}
		</mb-card>
	{/if}
</div>
