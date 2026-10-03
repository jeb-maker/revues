<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { completeRun, getRun, type RunDetail, type RunItem } from '$lib/api/runs';
	import { postRunNotionExport } from '$lib/api/notion';
	import { session } from '$lib/auth/session';
	import { formatItemStatus, formatRunStatus, itemStatusVariant } from '$lib/i18n/labels';
	import { inputValue } from '$lib/mb';

	const csrf = session().csrf_token;

	let run = $state<RunDetail | null>(null);
	let error = $state('');
	let message = $state('');
	let loading = $state(true);
	let closing = $state(false);
	let exporting = $state(false);
	let closingNote = $state('');

	function runId(): number {
		return Number(page.params.id);
	}

	onMount(async () => {
		try {
			run = await getRun(runId(), csrf);
			closingNote = run.closing_note ?? '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Revue introuvable.';
		} finally {
			loading = false;
		}
	});

	async function onComplete(e: Event) {
		e.preventDefault();
		if (!run || !confirm('Clôturer cette revue ?')) return;
		closing = true;
		error = '';
		try {
			run = await completeRun(run.id, { closing_note: closingNote }, csrf);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Clôture impossible.';
		} finally {
			closing = false;
		}
	}

	async function onExportNotion() {
		if (!run) return;
		exporting = true;
		error = '';
		message = '';
		try {
			const res = await postRunNotionExport(run.id, csrf);
			run = await getRun(run.id, csrf);
			message = `Exportée vers Notion : ${res.notion_url}`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Export Notion impossible.';
		} finally {
			exporting = false;
		}
	}

	function groupBySection(items: RunItem[]): { section: string; items: RunItem[] }[] {
		const order: string[] = [];
		const map = new Map<string, RunItem[]>();
		for (const item of items) {
			const key = item.section || 'Général';
			if (!map.has(key)) {
				map.set(key, []);
				order.push(key);
			}
			map.get(key)!.push(item);
		}
		return order.map((section) => ({ section, items: map.get(section)! }));
	}
</script>

<svelte:head>
	<title>{run?.title ?? 'Revue'} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs">
			<a href="/runs">Revues</a>
			{#if run}
				· <a href={`/subjects/${run.subject_id}`}>{run.subject_name}</a>
			{/if}
		</p>
		<h1>{run?.title ?? 'Revue'}</h1>
		{#if run}
			<p class="lede">
				{run.template_name} · v{run.template_version} ·
				<mb-badge variant={run.status === 'done' ? 'success' : 'info'}>{formatRunStatus(run.status)}</mb-badge>
			</p>
		{/if}
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if run}
		<section class="progress" aria-label="Progression">
			<p class="row">
				<strong>{run.progress.percent} %</strong>
				<span class="muted">{run.progress.done}/{run.progress.total} traités</span>
			</p>
			<mb-progress
				percent={run.progress.percent}
				aria-label={`Progression ${run.progress.percent} %`}
			></mb-progress>
			{#if (run.pending_required_count ?? 0) > 0}
				<p class="field-error">{run.pending_required_count} point(s) obligatoire(s) en attente</p>
			{/if}
		</section>

		{#each groupBySection(run.items) as group (group.section)}
			<section class="items">
				<h2>{group.section}</h2>
				<ul class="card-list">
					{#each group.items as item (item.id)}
						<li>
							<a href={`/runs/${run.id}/items/${item.id}`} class="row">
								<span class="label">
									{item.label}
									{#if item.required}<abbr title="Obligatoire">*</abbr>{/if}
								</span>
								<mb-badge variant={itemStatusVariant(item.status)}>{formatItemStatus(item.status)}</mb-badge>
								{#if item.assigned_login}
									<span class="muted">@{item.assigned_login}</span>
								{/if}
							</a>
						</li>
					{/each}
				</ul>
			</section>
		{/each}

		{#if run.capabilities.can_complete}
			<form class="stack-form section" onsubmit={onComplete}>
				<h2>Clôture</h2>
				<mb-textarea
					label="Note de clôture"
					rows="3"
					value={closingNote}
					oninput={(e) => (closingNote = inputValue(e))}
				></mb-textarea>
				<mb-button type="submit" variant="primary" disabled={closing}>
					{closing ? 'Clôture…' : 'Clôturer la revue'}
				</mb-button>
			</form>
		{:else if run.status === 'done' && run.closing_note}
			<p class="callout">Note de clôture : {run.closing_note}</p>
		{/if}

		{#if run.status === 'done'}
			<section class="section" aria-labelledby="notion">
				<h2 id="notion">Notion</h2>
				{#if run.notion_url}
					<p><a href={run.notion_url} target="_blank" rel="noopener noreferrer">Voir sur Notion</a></p>
				{:else if run.capabilities.can_export_notion}
					<mb-button variant="primary" disabled={exporting} onclick={onExportNotion}>
						{exporting ? 'Export…' : 'Exporter vers Notion'}
					</mb-button>
				{:else}
					<p class="muted">Aucun export Notion.</p>
				{/if}
			</section>
		{/if}
	{/if}
</div>

<style>
	.progress {
		margin-bottom: var(--mb-space-5);
	}
	.progress .row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin: 0 0 var(--mb-space-2);
	}
	.progress strong {
		font-size: var(--mb-font-size-lg);
	}
	.items {
		margin-bottom: var(--mb-space-5);
	}
	.items h2 {
		font-size: var(--mb-font-size-md);
	}
	.label {
		flex: 1;
		min-width: 10rem;
	}
	abbr {
		color: var(--mb-color-danger);
		text-decoration: none;
	}
</style>
