<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { listRuns, type RunSummary } from '$lib/api/runs';
	import { session } from '$lib/auth/session';
	import { formatRunStatus, runStatusVariant } from '$lib/i18n/labels';
	import { launchRunCTA, runLabels, subjectLabels } from '$lib/i18n/uiLabels';
	import { searchHrefFromListQuery } from '$lib/navigation/listQueryRedirect';

	type RunFilter = '' | 'draft' | 'in_progress' | 'done' | 'overdue';
	const PAGE_SIZE = 25;
	const STATUS_FILTERS: { value: RunFilter; label: string }[] = [
		{ value: '', label: 'Tous' },
		{ value: 'in_progress', label: 'En cours' },
		{ value: 'done', label: 'Terminées' },
		{ value: 'overdue', label: 'En retard' },
		{ value: 'draft', label: 'Brouillons' }
	];

	const boot = session();
	const csrf = boot.csrf_token;
	const run = $derived(runLabels(boot.organization?.ui_run_label));
	const subject = $derived(subjectLabels(boot.organization?.ui_subject_label));

	let runs = $state<RunSummary[]>([]);
	let total = $state(0);
	let status = $state<RunFilter>('');
	let offset = $state(0);
	let error = $state('');
	let loading = $state(true);

	function readURL() {
		const sp = page.url.searchParams;
		const st = (sp.get('status') ?? '') as RunFilter;
		status = STATUS_FILTERS.some((f) => f.value === st) ? st : '';
		const off = Number(sp.get('offset') ?? '0');
		offset = Number.isFinite(off) && off >= 0 ? off : 0;
	}

	async function syncURL() {
		const sp = new URLSearchParams();
		if (status) sp.set('status', status);
		if (offset > 0) sp.set('offset', String(offset));
		const qs = sp.toString();
		await goto(qs ? `/runs?${qs}` : '/runs', { replaceState: true, keepFocus: true, noScroll: true });
	}

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await listRuns({
				csrfToken: csrf,
				status: status || undefined,
				limit: PAGE_SIZE,
				offset
			});
			runs = res.runs ?? [];
			total = res.total;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur de chargement';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		void page.url.search;
		const redirect = searchHrefFromListQuery(page.url.searchParams.get('q'));
		if (redirect) {
			void goto(redirect, { replaceState: true });
			return;
		}
		readURL();
		void load();
	});

	async function onFilter(e: Event) {
		e.preventDefault();
		offset = 0;
		await syncURL();
		await load();
	}

	async function goPage(nextOffset: number) {
		offset = Math.max(0, nextOffset);
		await syncURL();
		await load();
	}

	const pageEnd = $derived(Math.min(offset + PAGE_SIZE, total));
	const hasPrev = $derived(offset > 0);
	const hasNext = $derived(offset + PAGE_SIZE < total);
	const tableColumns = $derived(
		boot.show_subject_column ? '2fr 1.2fr 1fr 0.7fr auto' : '2fr 1fr 0.7fr auto'
	);
</script>

<svelte:head>
	<title>{run.nav} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<div class="page-header__row">
			<h1>{run.nav}</h1>
			<p class="actions page-header__actions">
				<mb-button href="/subjects" variant="primary">
					<svg class="icon icon--fill" viewBox="0 0 24 24" aria-hidden="true"
						><path d="M8 5v14l11-7z" /></svg
					>
					{launchRunCTA(run)}
				</mb-button>
			</p>
		</div>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	<form class="filters" onsubmit={onFilter}>
		<mb-select
			label="Statut"
			hide-label
			value={status}
			onmb-change={(e) => (status = e.detail.value as RunFilter)}
		>
			{#each STATUS_FILTERS as f (f.value)}
				<option value={f.value}>{f.label}</option>
			{/each}
		</mb-select>
		<mb-button type="submit" variant="secondary">Filtrer</mb-button>
		{#if !loading}
			<p class="filters__count muted">
				{total} {total > 1 ? run.plural : run.singular}
				{#if total > PAGE_SIZE}
					· {offset + 1}–{pageEnd}
				{/if}
			</p>
		{/if}
	</form>

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if runs.length === 0}
		<mb-empty-state heading="{run.noneArticle} {run.singular}">
			<a href="/subjects">{launchRunCTA(run)}</a> depuis un {subject.singular.toLowerCase()}.
		</mb-empty-state>
	{:else}
		<mb-table columns={tableColumns} sticky-header>
			<mb-table-row slot="head">
				<mb-table-cell>Titre</mb-table-cell>
				{#if boot.show_subject_column}
					<mb-table-cell>{subject.singular}</mb-table-cell>
				{/if}
				<mb-table-cell>Statut</mb-table-cell>
				<mb-table-cell>Progression</mb-table-cell>
				<mb-table-cell>Actions</mb-table-cell>
			</mb-table-row>
			{#each runs as item (item.id)}
				<mb-table-row>
					<mb-table-cell label="Titre" primary>
						<a href={`/runs/${item.id}`}>{item.title}</a>
					</mb-table-cell>
					{#if boot.show_subject_column}
						<mb-table-cell label={subject.singular}>{item.subject_name}</mb-table-cell>
					{/if}
					<mb-table-cell label="Statut">
						<mb-badge variant={runStatusVariant(item.status)}>{formatRunStatus(item.status)}</mb-badge>
					</mb-table-cell>
					<mb-table-cell label="Progression">
						<span class="pct">{item.progress.percent} %</span>
					</mb-table-cell>
					<mb-table-cell actions>
						<a class="row-action" href={`/runs/${item.id}`} aria-label="Ouvrir" title="Ouvrir">
							<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
								><path d="M5 12h14M13 6l6 6-6 6" /></svg
							>
						</a>
					</mb-table-cell>
				</mb-table-row>
			{/each}
		</mb-table>
		{#if hasPrev || hasNext}
			<p class="actions">
				<mb-button variant="secondary" disabled={!hasPrev} onclick={() => goPage(offset - PAGE_SIZE)}>
					Précédent
				</mb-button>
				<mb-button variant="secondary" disabled={!hasNext} onclick={() => goPage(offset + PAGE_SIZE)}>
					Suivant
				</mb-button>
			</p>
		{/if}
	{/if}
</div>

<style>
	.pct {
		font-variant-numeric: tabular-nums;
	}
</style>
