<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { listRuns, type RunSummary } from '$lib/api/runs';
	import { session } from '$lib/auth/session';
	import { formatRunStatus, runStatusVariant } from '$lib/i18n/labels';
	import { launchRunCTA, runLabels, subjectLabels } from '$lib/i18n/uiLabels';
	import { searchHrefFromListQuery } from '$lib/navigation/listQueryRedirect';

	type RunFilter = '' | 'draft' | 'in_progress' | 'done' | 'overdue';
	type SortKey = 'titre' | 'sujet' | 'date' | 'statut' | 'progression';
	type SortDirection = 'asc' | 'desc';

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
	let sortKey = $state<SortKey>('date');
	let sortDirection = $state<SortDirection>('desc');

	function parseStatus(sp: URLSearchParams): RunFilter {
		const st = (sp.get('status') ?? '') as RunFilter;
		return STATUS_FILTERS.some((f) => f.value === st) ? st : '';
	}

	function parseOffset(sp: URLSearchParams): number {
		const off = Number(sp.get('offset') ?? '0');
		return Number.isFinite(off) && off >= 0 ? off : 0;
	}

	function runsHref(nextStatus: RunFilter, nextOffset: number): string {
		const sp = new URLSearchParams();
		if (nextStatus) sp.set('status', nextStatus);
		if (nextOffset > 0) sp.set('offset', String(nextOffset));
		const qs = sp.toString();
		return qs ? `/runs?${qs}` : '/runs';
	}

	async function loadWith(statusFilter: RunFilter, pageOffset: number) {
		loading = true;
		error = '';
		try {
			const res = await listRuns({
				csrfToken: csrf,
				status: statusFilter || undefined,
				limit: PAGE_SIZE,
				offset: pageOffset
			});
			runs = res.runs ?? [];
			total = res.total;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur de chargement';
		} finally {
			loading = false;
		}
	}

	// URL is the source of truth. untrack(load) so reading status/offset inside
	// listRuns does not re-subscribe this effect (that was resetting the select).
	$effect(() => {
		void page.url.search;
		const redirect = searchHrefFromListQuery(page.url.searchParams.get('q'));
		if (redirect) {
			void goto(redirect, { replaceState: true });
			return;
		}
		const sp = page.url.searchParams;
		const nextStatus = parseStatus(sp);
		const nextOffset = parseOffset(sp);
		status = nextStatus;
		offset = nextOffset;
		untrack(() => {
			void loadWith(nextStatus, nextOffset);
		});
	});

	async function onStatusChange(e: CustomEvent<{ value: string }>) {
		const next = e.detail.value as RunFilter;
		status = STATUS_FILTERS.some((f) => f.value === next) ? next : '';
		offset = 0;
		await goto(runsHref(status, 0), { replaceState: true, keepFocus: true, noScroll: true });
	}

	async function goPage(nextOffset: number) {
		const off = Math.max(0, nextOffset);
		offset = off;
		await goto(runsHref(status, off), { replaceState: true, keepFocus: true, noScroll: true });
	}

	function formatCreatedDate(iso: string): string {
		const t = Date.parse(iso);
		if (Number.isNaN(t)) return iso || '—';
		return new Date(t).toLocaleDateString('fr-FR');
	}

	function onSort(e: CustomEvent<{ key: string; direction: string }>) {
		const key = e.detail.key as SortKey;
		const direction = e.detail.direction === 'asc' ? 'asc' : 'desc';
		if (key !== 'titre' && key !== 'sujet' && key !== 'date' && key !== 'statut' && key !== 'progression') {
			return;
		}
		sortKey = key;
		sortDirection = direction;
	}

	const displayed = $derived.by(() => {
		const list = [...runs];
		const dir = sortDirection === 'asc' ? 1 : -1;
		list.sort((a, b) => {
			switch (sortKey) {
				case 'titre':
					return a.title.localeCompare(b.title, 'fr') * dir;
				case 'sujet':
					return a.subject_name.localeCompare(b.subject_name, 'fr') * dir;
				case 'statut':
					return a.status.localeCompare(b.status) * dir;
				case 'progression':
					return (a.progress.percent - b.progress.percent) * dir;
				case 'date':
				default:
					return a.created_at.localeCompare(b.created_at) * dir;
			}
		});
		return list;
	});

	const pageEnd = $derived(Math.min(offset + PAGE_SIZE, total));
	const hasPrev = $derived(offset > 0);
	const hasNext = $derived(offset + PAGE_SIZE < total);
	const tableColumns = $derived(
		boot.show_subject_column
			? '2fr 1.1fr 0.8fr 0.9fr 0.7fr 2.75rem'
			: '2fr 0.8fr 0.9fr 0.7fr 2.75rem'
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

	<div class="filters">
		<mb-select
			label="Statut"
			hide-label
			value={status}
			onmb-change={onStatusChange}
		>
			{#each STATUS_FILTERS as f (f.value)}
				<option value={f.value}>{f.label}</option>
			{/each}
		</mb-select>
		{#if !loading}
			<p class="filters__count muted">
				{total} {total > 1 ? run.plural : run.singular}
				{#if total > PAGE_SIZE}
					· {offset + 1}–{pageEnd}
				{/if}
			</p>
		{/if}
	</div>

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if runs.length === 0}
		<mb-empty-state heading="{run.noneArticle} {run.singular}">
			<a href="/subjects">{launchRunCTA(run)}</a> depuis un {subject.singular.toLowerCase()}.
		</mb-empty-state>
	{:else}
		<mb-table
			columns={tableColumns}
			sort-key={sortKey}
			sort-direction={sortDirection}
			sort-label="Trier par {name}"
			onmb-sort={onSort}
		>
			<mb-table-row slot="head">
				<mb-table-cell sort-key="titre">Titre</mb-table-cell>
				{#if boot.show_subject_column}
					<mb-table-cell sort-key="sujet">{subject.singular}</mb-table-cell>
				{/if}
				<mb-table-cell sort-key="date" align="center">Date</mb-table-cell>
				<mb-table-cell sort-key="statut" align="center">Statut</mb-table-cell>
				<mb-table-cell sort-key="progression" align="center">Progression</mb-table-cell>
				<mb-table-cell actions><span class="sr-only">Actions</span></mb-table-cell>
			</mb-table-row>
			{#each displayed as item (item.id)}
				<mb-table-row>
					<mb-table-cell label="Titre" primary sort-value={item.title}>
						<a href={`/runs/${item.id}`}>{item.title}</a>
					</mb-table-cell>
					{#if boot.show_subject_column}
						<mb-table-cell label={subject.singular} sort-value={item.subject_name}
							>{item.subject_name}</mb-table-cell
						>
					{/if}
					<mb-table-cell label="Date" align="center" sort-value={item.created_at}>
						<span class="date">{formatCreatedDate(item.created_at)}</span>
					</mb-table-cell>
					<mb-table-cell label="Statut" align="center" sort-value={item.status}>
						<mb-badge variant={runStatusVariant(item.status)}>{formatRunStatus(item.status)}</mb-badge>
					</mb-table-cell>
					<mb-table-cell
						label="Progression"
						align="center"
						sort-value={String(item.progress.percent).padStart(3, '0')}
					>
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
	.pct,
	.date {
		font-variant-numeric: tabular-nums;
	}
</style>
