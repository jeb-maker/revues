<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import { listAdminIntegrations, type IntegrationSummary } from '$lib/api/admin';

	let csrf = $state('');
	let items = $state<IntegrationSummary[]>([]);
	let error = $state('');
	let loading = $state(true);

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			const overview = await listAdminIntegrations(csrf);
			items = overview.items ?? [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Hub intégrations indisponible.';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intégrations — Revues</title>
</svelte:head>

<main class="admin-page">
	<header>
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs"><a href="/">Accueil</a> · Admin</p>
		<h1>Intégrations</h1>
		<p class="lede">État des connecteurs de l'organisation active.</p>
	</header>

	{#if error}
		<p class="err" role="alert">{error}</p>
	{/if}

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else}
		<ul class="list">
			{#each items as item}
				<li>
					<div class="row">
						<div>
							<p class="name">{item.name}</p>
							<p class="desc">{item.description}</p>
						</div>
						<span class="badge" class:on={item.enabled}>{item.enabled ? 'Actif' : 'Inactif'}</span>
					</div>
					<a class="cfg" href={item.config_path}>Configurer</a>
				</li>
			{/each}
		</ul>
	{/if}
</main>


