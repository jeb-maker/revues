<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { listSubjects, type SubjectSummary } from '$lib/api/subjects';
	import { session } from '$lib/auth/session';
	import { formatVisibility } from '$lib/i18n/labels';
	import { subjectLabels } from '$lib/i18n/uiLabels';
	import { searchHrefFromListQuery } from '$lib/navigation/listQueryRedirect';

	const boot = session();
	const subject = $derived(subjectLabels(boot.organization?.ui_subject_label));
	const templateId = $derived(page.url.searchParams.get('template_id'));

	let subjects = $state<SubjectSummary[]>([]);
	let canCreate = $state(false);
	let error = $state('');
	let loading = $state(true);

	function subjectHref(id: number): string {
		if (templateId) return `/subjects/${id}/launch?template_id=${encodeURIComponent(templateId)}`;
		return `/subjects/${id}`;
	}

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await listSubjects();
			subjects = res.subjects ?? [];
			canCreate = res.can_create;
		} catch (e) {
			error = e instanceof Error ? e.message : `Impossible de charger les ${subject.plural.toLowerCase()}.`;
			subjects = [];
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
	<title>{subject.plural} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<div class="page-header__row">
			<h1>{subject.plural}</h1>
			{#if canCreate && !templateId}
				<p class="actions page-header__actions">
					<mb-button variant="primary" href="/subjects/new">
						<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
							><path d="M12 5v14M5 12h14" /></svg
						>
						Nouveau {subject.singular.toLowerCase()}
					</mb-button>
				</p>
			{/if}
		</div>
		{#if templateId}
			<p class="lede">Choisissez un {subject.singular.toLowerCase()} pour lancer avec ce modèle.</p>
		{/if}
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{:else if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if subjects.length === 0}
		<mb-empty-state heading="Aucun {subject.singular.toLowerCase()}"
			>Aucun {subject.singular.toLowerCase()} visible dans cette organisation.</mb-empty-state
		>
	{:else}
		<mb-table columns="2fr 1fr auto">
			<mb-table-row slot="head">
				<mb-table-cell>Nom</mb-table-cell>
				<mb-table-cell>Visibilité</mb-table-cell>
				<mb-table-cell actions>Actions</mb-table-cell>
			</mb-table-row>
			{#each subjects as s (s.id)}
				<mb-table-row>
					<mb-table-cell label="Nom" primary>
						<a href={subjectHref(s.id)}>{s.name}</a>
						{#if s.description}
							<span class="desc">{s.description}</span>
						{/if}
					</mb-table-cell>
					<mb-table-cell label="Visibilité">
						<mb-badge variant={s.visibility === 'private' ? 'warning' : 'neutral'}>
							{formatVisibility(s.visibility)}
						</mb-badge>
					</mb-table-cell>
					<mb-table-cell actions>
						<a
							class="row-action"
							href={subjectHref(s.id)}
							aria-label={templateId ? 'Lancer' : 'Ouvrir'}
							title={templateId ? 'Lancer' : 'Ouvrir'}
						>
							{#if templateId}
								<svg class="icon icon--fill" viewBox="0 0 24 24" aria-hidden="true"
									><path d="M8 5v14l11-7z" /></svg
								>
							{:else}
								<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
									><path d="M5 12h14M13 6l6 6-6 6" /></svg
								>
							{/if}
						</a>
					</mb-table-cell>
				</mb-table-row>
			{/each}
		</mb-table>
	{/if}
</div>
