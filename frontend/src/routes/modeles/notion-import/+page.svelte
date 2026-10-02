<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import { postTemplatesNotionImport } from '$lib/api/notion';

	let csrf = $state('');
	let error = $state('');
	let loading = $state(true);
	let busy = $state(false);
	let databaseRef = $state('');

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			if (boot.user?.role !== 'admin' && boot.user?.role !== 'editor') {
				error = 'Droits éditeur requis.';
				return;
			}
			csrf = boot.csrf_token;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Session indisponible.';
		} finally {
			loading = false;
		}
	});

	async function onImport(e: Event) {
		e.preventDefault();
		busy = true;
		error = '';
		try {
			const fetched = await postTemplatesNotionImport(
				{ action: 'fetch', database_ref: databaseRef },
				csrf
			);
			const created = await postTemplatesNotionImport(
				{
					action: 'import',
					database_id: fetched.database_id ?? '',
					template_name: fetched.template_name ?? '',
					mapping: fetched.mapping ?? {}
				},
				csrf
			);
			if (created.template?.id) await goto(`/modeles/${created.template.id}`);
			else error = 'Import terminé sans modèle.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Échec Notion.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>Import Notion — Revues</title>
</svelte:head>

<main class="admin-page">
	<header>
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs"><a href="/modeles">Modèles</a> · Import Notion</p>
		<h1>Importer depuis Notion</h1>
		<p class="lede">Lit la base, applique le mapping par défaut, crée le modèle (API wizard complète).</p>
	</header>
	{#if error}<p class="err" role="alert">{error}</p>{/if}
	{#if loading}
		<p class="muted">Chargement…</p>
	{:else}
		<form class="stack" onsubmit={onImport}>
			<label>URL ou ID base<input bind:value={databaseRef} required autocomplete="off" /></label>
			<button type="submit" disabled={busy}>{busy ? 'Import…' : 'Importer'}</button>
		</form>
	{/if}
</main>

<style>
	.stack {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	input {
		padding: 0.5rem 0.6rem;
		border-radius: 0.35rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	button {
		align-self: flex-start;
		padding: 0.55rem 0.9rem;
		border: none;
		border-radius: 0.35rem;
		background: #0f766e;
		color: #ecfdf5;
		font: inherit;
		font-weight: 600;
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.5;
	}
</style>
