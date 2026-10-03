<script lang="ts">
	import { onMount } from 'svelte';
	import { listRuns, type RunSummary } from '$lib/api/runs';
	import { session } from '$lib/auth/session';
	import { formatRunStatus, runStatusVariant } from '$lib/i18n/labels';
	import { runLabels } from '$lib/i18n/uiLabels';
	import { inputValue } from '$lib/mb';

	type RunFilter = '' | 'draft' | 'in_progress' | 'done' | 'overdue';
	const STATUS_FILTERS: { value: RunFilter; label: string }[] = [
		{ value: '', label: 'Tous les statuts' },
		{ value: 'in_progress', label: 'En cours' },
		{ value: 'done', label: 'Terminées' },
		{ value: 'overdue', label: 'En retard' },
		{ value: 'draft', label: 'Brouillons' }
	];

	const boot = session();
	const csrf = boot.csrf_token;
	const run = $derived(runLabels(boot.organization?.ui_run_label));

	let runs = $state<RunSummary[]>([]);
	let total = $state(0);
	let q = $state('');
	let status = $state<RunFilter>('');
	let error = $state('');
	let loading = $state(true);

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await listRuns({
				csrfToken: csrf,
				q: q.trim() || undefined,
				status: status || undefined
			});
			runs = res.runs ?? [];
			total = res.total;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur de chargement';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void load();
	});

	async function onFilter(e: Event) {
		e.preventDefault();
		await load();
	}
</script>

<svelte:head>
	<title>{run.nav} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1>{run.nav}</h1>
		<p class="lede">Exécutions en cours et historiques — progression par snapshot.</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	<form class="filters" onsubmit={onFilter}>
		<mb-input
			label="Rechercher"
			hide-label
			type="search"
			placeholder="Rechercher…"
			value={q}
			oninput={(e) => (q = inputValue(e))}
		></mb-input>
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
	</form>

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if runs.length === 0}
		<mb-empty-state heading="Aucune revue">
			Lancez-en une depuis un <a href="/subjects">sujet</a>.
		</mb-empty-state>
	{:else}
		<p class="muted">{total} revue{total > 1 ? 's' : ''}</p>
		<ul class="card-list">
			{#each runs as run (run.id)}
				<li>
					<a href={`/runs/${run.id}`}>
						<strong>{run.title}</strong>
						<span class="row">
							<span class="muted">{run.subject_name}</span>
							<mb-badge variant={runStatusVariant(run.status)}>{formatRunStatus(run.status)}</mb-badge>
							<span class="pct">{run.progress.percent} %</span>
						</span>
						<mb-progress
							percent={run.progress.percent}
							aria-label={`Progression ${run.progress.percent} %`}
						></mb-progress>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.row {
		margin-top: var(--mb-space-1);
		font-size: var(--mb-font-size-sm);
	}
	.pct {
		margin-left: auto;
		font-variant-numeric: tabular-nums;
	}
	mb-progress {
		margin: var(--mb-space-2) 0 0;
	}
</style>
