<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { listTemplates, type TemplateSummary } from '$lib/api/templates';
	import { session } from '$lib/auth/session';
	import { templateNavLabel } from '$lib/i18n/uiLabels';
	import { searchHrefFromListQuery } from '$lib/navigation/listQueryRedirect';

	const boot = session();
	const canManage = boot.can_edit;
	const templatesLabel = $derived(templateNavLabel(boot.organization?.ui_run_label));
	const newTemplateCTA = $derived(templatesLabel === 'Listes' ? 'Nouvelle liste' : 'Nouveau modèle');

	let templates = $state<TemplateSummary[]>([]);
	let error = $state('');
	let loading = $state(true);

	async function load() {
		loading = true;
		error = '';
		try {
			templates = await listTemplates(boot.csrf_token);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur de chargement';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		const redirect = searchHrefFromListQuery(page.url.searchParams.get('q'));
		if (redirect) {
			void goto(redirect, { replaceState: true });
			return;
		}
		void load();
	});
</script>

<svelte:head>
	<title>{templatesLabel} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<div class="page-header__row">
			<h1>{templatesLabel}</h1>
			{#if canManage}
				<p class="actions page-header__actions">
					<mb-button variant="primary" href="/modeles/new">
						<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
							><path d="M12 5v14M5 12h14" /></svg
						>
						{newTemplateCTA}
					</mb-button>
				</p>
			{/if}
		</div>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if templates.length === 0}
		<mb-empty-state heading="Aucun modèle">
			{#if canManage}<a href="/modeles/new">Créer le premier modèle</a>.{:else}Aucun modèle publié.{/if}
		</mb-empty-state>
	{:else}
		<mb-table columns="2fr 0.7fr 0.7fr 1.4fr auto" sticky-header>
			<mb-table-row slot="head">
				<mb-table-cell>Nom</mb-table-cell>
				<mb-table-cell>Version</mb-table-cell>
				<mb-table-cell>Points</mb-table-cell>
				<mb-table-cell>Domaines</mb-table-cell>
				<mb-table-cell actions>Actions</mb-table-cell>
			</mb-table-row>
			{#each templates as t (t.id)}
				<mb-table-row>
					<mb-table-cell label="Nom" primary>
						<a href={`/modeles/${t.id}`}>{t.name}</a>
					</mb-table-cell>
					<mb-table-cell label="Version">v{t.latest_version}</mb-table-cell>
					<mb-table-cell label="Points">{t.item_count}</mb-table-cell>
					<mb-table-cell label="Domaines">
						{#if t.domains?.length}
							{t.domains.join(', ')}
						{:else}
							<span class="muted">—</span>
						{/if}
					</mb-table-cell>
					<mb-table-cell actions>
						<span class="row-actions">
							<a class="row-action" href={`/modeles/${t.id}`} aria-label="Ouvrir" title="Ouvrir">
								<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
									><path d="M5 12h14M13 6l6 6-6 6" /></svg
								>
							</a>
							<a
								class="row-action"
								href={`/subjects?template_id=${t.id}`}
								aria-label="Lancer"
								title="Lancer"
							>
								<svg class="icon icon--fill" viewBox="0 0 24 24" aria-hidden="true"
									><path d="M8 5v14l11-7z" /></svg
								>
							</a>
							{#if canManage}
								<a
									class="row-action"
									href={`/modeles/${t.id}/edit`}
									aria-label="Éditer"
									title="Éditer"
								>
									<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
										><path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" /></svg
									>
								</a>
							{/if}
						</span>
					</mb-table-cell>
				</mb-table-row>
			{/each}
		</mb-table>
	{/if}
</div>
