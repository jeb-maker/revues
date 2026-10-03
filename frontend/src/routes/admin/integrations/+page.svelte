<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { listAdminIntegrations, type IntegrationSummary } from '$lib/api/admin';
	import { session } from '$lib/auth/session';

	const csrf = session().csrf_token;

	let items = $state<IntegrationSummary[]>([]);
	let error = $state('');
	let loading = $state(true);

	onMount(async () => {
		try {
			const overview = await listAdminIntegrations(csrf);
			items = overview.items ?? [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Intégrations indisponibles.';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intégrations — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs"><a href="/admin">Administration</a> · Intégrations</p>
		<h1>Intégrations</h1>
		<p class="lede">État des connecteurs de l’organisation active.</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	<AdminNav section="integrations">
		{#if loading}
			<p class="loading">Chargement…</p>
		{:else if items.length === 0}
			<p class="muted">Aucun connecteur disponible.</p>
		{:else}
			<ul class="card-list">
				{#each items as item (item.config_path)}
					<li class="card">
						<div class="row">
							<strong class="grow">{item.name}</strong>
							<mb-badge variant={item.enabled ? 'success' : 'neutral'}>
								{item.enabled ? 'Actif' : 'Inactif'}
							</mb-badge>
						</div>
						<p class="desc">{item.description}</p>
						<mb-button href={item.config_path} variant="secondary" size="sm">Configurer</mb-button>
					</li>
				{/each}
			</ul>
		{/if}
	</AdminNav>
</div>

<style>
	.grow {
		flex: 1;
	}
	.card .desc {
		margin: var(--mb-space-1) 0 var(--mb-space-3);
	}
</style>
