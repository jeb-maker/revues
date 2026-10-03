<script lang="ts">
	import { goto } from '$app/navigation';
	import { postTemplatesNotionImport } from '$lib/api/notion';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';

	const boot = session();
	const csrf = boot.csrf_token;

	let error = $state('');
	let busy = $state(false);
	let databaseRef = $state('');

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
			error = err instanceof Error ? err.message : 'Échec de l’import Notion.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>Import Notion — Revues</title>
</svelte:head>

<div class="page page--narrow">
	<header class="page-header">
		<p class="crumbs"><a href="/modeles">Modèles</a> · Import Notion</p>
		<h1>Importer depuis Notion</h1>
		<p class="lede">
			Lit la base Notion, applique le mapping par défaut et crée le modèle (version 1).
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	{#if !boot.can_edit}
		<mb-alert variant="warning">Droits éditeur requis pour importer un modèle.</mb-alert>
	{:else}
		<form class="stack-form" onsubmit={onImport}>
			<mb-input
				label="URL ou identifiant de la base Notion"
				type="text"
				required
				autocomplete="off"
				value={databaseRef}
				oninput={(e) => (databaseRef = inputValue(e))}
			></mb-input>
			<mb-button type="submit" variant="primary" disabled={busy}>
				{busy ? 'Import…' : 'Importer'}
			</mb-button>
		</form>
	{/if}
</div>
