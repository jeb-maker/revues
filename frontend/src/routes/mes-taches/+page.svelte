<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { listMyTasks, type MyTask } from '$lib/api/mytasks';
	import { session } from '$lib/auth/session';
	import { formatItemStatus, itemStatusVariant, type ItemStatus } from '$lib/i18n/labels';
	import { searchHrefFromListQuery } from '$lib/navigation/listQueryRedirect';

	type StatusFilter = '' | ItemStatus;
	const STATUS_FILTERS: { value: StatusFilter; label: string }[] = [
		{ value: '', label: 'Tous les statuts' },
		{ value: 'pending', label: formatItemStatus('pending') },
		{ value: 'ok', label: formatItemStatus('ok') },
		{ value: 'nok', label: formatItemStatus('nok') },
		{ value: 'na', label: formatItemStatus('na') }
	];

	const csrf = session().csrf_token;

	let tasks = $state<MyTask[]>([]);
	let status = $state<StatusFilter>('');
	let error = $state('');
	let loading = $state(true);

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await listMyTasks({
				csrfToken: csrf,
				status: status || undefined
			});
			tasks = res.tasks ?? [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger les tâches.';
			tasks = [];
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

	async function onStatusChange(e: CustomEvent<{ value: string }>) {
		status = e.detail.value as StatusFilter;
		await load();
	}
</script>

<svelte:head>
	<title>Mes tâches — Revues</title>
</svelte:head>

<div class="page page--wide">
	<header class="page-header">
		<div class="page-header__row">
			<h1>Mes tâches</h1>
		</div>
	</header>

	<div class="filters">
		<mb-select label="Statut" hide-label value={status} onmb-change={onStatusChange}>
			{#each STATUS_FILTERS as f (f.value)}
				<option value={f.value}>{f.label}</option>
			{/each}
		</mb-select>
		{#if !loading}
			<p class="filters__count muted">{tasks.length} tâche{tasks.length > 1 ? 's' : ''}</p>
		{/if}
	</div>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{:else if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if tasks.length === 0}
		<mb-empty-state heading="Aucune tâche">Aucun point assigné avec ces filtres.</mb-empty-state>
	{:else}
		<mb-table columns="2fr 1.2fr 1fr 0.8fr auto">
			<mb-table-row slot="head">
				<mb-table-cell>Point</mb-table-cell>
				<mb-table-cell>Sujet</mb-table-cell>
				<mb-table-cell>Revue</mb-table-cell>
				<mb-table-cell>Statut</mb-table-cell>
				<mb-table-cell actions>Actions</mb-table-cell>
			</mb-table-row>
			{#each tasks as task (task.id)}
				<mb-table-row>
					<mb-table-cell label="Point" primary>
						<a href={`/runs/${task.run_id}/items/${task.id}`}>{task.label}</a>
						{#if task.section}
							<span class="desc">{task.section}</span>
						{/if}
					</mb-table-cell>
					<mb-table-cell label="Sujet">
						<a href={`/subjects/${task.subject_id}`}>{task.subject_name}</a>
					</mb-table-cell>
					<mb-table-cell label="Revue">
						<a href={`/runs/${task.run_id}`}>{task.run_title}</a>
					</mb-table-cell>
					<mb-table-cell label="Statut">
						<mb-badge variant={itemStatusVariant(task.status)}>{formatItemStatus(task.status)}</mb-badge>
					</mb-table-cell>
					<mb-table-cell actions>
						<a
							class="row-action"
							href={`/runs/${task.run_id}/items/${task.id}`}
							aria-label="Ouvrir"
							title="Ouvrir"
						>
							<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
								><path d="M5 12h14M13 6l6 6-6 6" /></svg
							>
						</a>
					</mb-table-cell>
				</mb-table-row>
			{/each}
		</mb-table>
	{/if}
</div>
