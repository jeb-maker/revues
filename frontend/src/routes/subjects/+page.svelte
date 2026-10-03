<script lang="ts">
	import { onMount } from 'svelte';
	import { listSubjects, type SubjectSummary } from '$lib/api/subjects';
	import { formatVisibility } from '$lib/i18n/labels';
	import { inputValue } from '$lib/mb';

	let subjects = $state<SubjectSummary[]>([]);
	let canCreate = $state(false);
	let q = $state('');
	let error = $state('');
	let loading = $state(true);

	async function load(query?: string) {
		loading = true;
		error = '';
		try {
			const res = await listSubjects({ q: query });
			subjects = res.subjects ?? [];
			canCreate = res.can_create;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger les sujets.';
			subjects = [];
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void load();
	});

	async function onSearch(e: Event) {
		e.preventDefault();
		await load(q.trim());
	}
</script>

<svelte:head>
	<title>Sujets — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1>Sujets</h1>
		<p class="lede">Conteneurs de revues — domaines, étiquettes et membres.</p>
		{#if canCreate}
			<p class="actions">
				<mb-button variant="primary" href="/subjects/new">Nouveau sujet</mb-button>
			</p>
		{/if}
	</header>

	<form class="filters" onsubmit={onSearch}>
		<mb-input
			label="Rechercher"
			hide-label
			type="search"
			name="q"
			placeholder="Filtrer par nom…"
			value={q}
			oninput={(e) => (q = inputValue(e))}
		></mb-input>
		<mb-button type="submit" variant="secondary">Filtrer</mb-button>
	</form>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{:else if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if subjects.length === 0}
		<mb-empty-state heading="Aucun sujet">Aucun sujet visible dans l'organisation active.</mb-empty-state>
	{:else}
		<mb-table columns="2fr 1fr auto" sticky-header>
			<mb-table-row slot="head">
				<mb-table-cell>Nom</mb-table-cell>
				<mb-table-cell>Visibilité</mb-table-cell>
				<mb-table-cell>Actions</mb-table-cell>
			</mb-table-row>
			{#each subjects as s (s.id)}
				<mb-table-row>
					<mb-table-cell label="Nom" primary>
						<a href={`/subjects/${s.id}`}>{s.name}</a>
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
						<a href={`/subjects/${s.id}`}>Ouvrir</a>
					</mb-table-cell>
				</mb-table-row>
			{/each}
		</mb-table>
	{/if}
</div>
