<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import {
		completeRun,
		getRun,
		updateRunItem,
		UpdateRunItemError,
		type RunDetail,
		type RunItem
	} from '$lib/api/runs';
	import { session } from '$lib/auth/session';
	import { formatItemStatus, formatRunStatus, itemStatusVariant } from '$lib/i18n/labels';
	import { runLabels } from '$lib/i18n/uiLabels';
	import { inputValue } from '$lib/mb';

	type ItemStatus = 'pending' | 'ok' | 'nok' | 'na';
	const ITEM_STATUSES: { value: ItemStatus; label: string }[] = [
		{ value: 'pending', label: 'En attente' },
		{ value: 'ok', label: 'Validé' },
		{ value: 'nok', label: 'Non validé' },
		{ value: 'na', label: 'Non applicable' }
	];

	const boot = session();
	const csrf = boot.csrf_token;
	const runLbl = $derived(runLabels(boot.organization?.ui_run_label));

	let run = $state<RunDetail | null>(null);
	let error = $state('');
	let loading = $state(true);
	let closing = $state(false);
	let closingNote = $state('');
	let savingId = $state<number | null>(null);

	function runId(): number {
		return Number(page.params.id);
	}

	function formatWhen(iso: string | null | undefined): string {
		if (!iso) return '';
		const d = new Date(iso);
		if (Number.isNaN(d.getTime())) return iso;
		return d.toLocaleString('fr-FR', { dateStyle: 'medium', timeStyle: 'short' });
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

	async function onStatusChange(item: RunItem, status: ItemStatus) {
		if (!run || !run.capabilities.can_update_items) return;
		savingId = item.id;
		error = '';
		try {
			const updated = await updateRunItem(
				run.id,
				item.id,
				{ status, updated_at: item.updated_at },
				csrf
			);
			run = await getRun(run.id, csrf);
			void updated;
		} catch (err) {
			if (err instanceof UpdateRunItemError && err.status === 409) {
				error = 'Ce point a été modifié ailleurs — rechargement.';
				run = await getRun(run.id, csrf);
			} else {
				error = err instanceof Error ? err.message : 'Mise à jour impossible.';
			}
		} finally {
			savingId = null;
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
	<title>{run?.title ?? runLbl.singular} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs">
			<a href="/runs">{runLbl.nav}</a>
			{#if run}
				· <a href={`/subjects/${run.subject_id}`}>{run.subject_name}</a>
			{/if}
		</p>
		<h1>{run?.template_name ?? runLbl.singular}</h1>
		{#if run}
			<p class="lede">
				v{run.template_version}
				{#if run.status !== 'in_progress'}
					·
					<mb-badge variant={run.status === 'done' ? 'success' : 'info'}
						>{formatRunStatus(run.status)}</mb-badge
					>
				{/if}
				{#if run.due_date}
					· échéance {run.due_date}
				{/if}
			</p>
		{/if}
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if run}
		<div class="card-stack">
			{#if run.status === 'done'}
				<mb-card>
					<h2 slot="header">Attestation de clôture</h2>
					<p>
						Clôturée
						{#if run.completed_by_login}
							par <strong>@{run.completed_by_login}</strong>
						{/if}
						{#if run.completed_at}
							le <strong>{formatWhen(run.completed_at)}</strong>
						{/if}.
					</p>
					{#if run.closing_note}
						<p class="muted">{run.closing_note}</p>
					{/if}
				</mb-card>
			{/if}

			<mb-card>
				<h2 slot="header">Points</h2>
				<p class="progress-row">
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

				{#each groupBySection(run.items) as group (group.section)}
					<section class="items">
						<h3>{group.section}</h3>
						<ul class="item-grid">
							{#each group.items as item (item.id)}
								<li class="item-row">
									<span class="label">
										{item.label}
										{#if item.required}<abbr title="Obligatoire">*</abbr>{/if}
									</span>
									{#if run.capabilities.can_update_items}
										<mb-select
											label="Statut"
											hide-label
											value={item.status}
											disabled={savingId === item.id}
											onmb-change={(e) => onStatusChange(item, e.detail.value as ItemStatus)}
										>
											{#each ITEM_STATUSES as s (s.value)}
												<option value={s.value}>{s.label}</option>
											{/each}
										</mb-select>
									{:else}
										<mb-badge variant={itemStatusVariant(item.status)}
											>{formatItemStatus(item.status)}</mb-badge
										>
									{/if}
									{#if item.assigned_login}
										<span class="muted">@{item.assigned_login}</span>
									{/if}
									<a class="details" href={`/runs/${run.id}/items/${item.id}`}>Détails</a>
								</li>
							{/each}
						</ul>
					</section>
				{/each}
			</mb-card>

			{#if run.capabilities.can_complete}
				<form onsubmit={onComplete}>
					<mb-card>
						<h2 slot="header">Clôture</h2>
						<div class="stack-form">
							<mb-textarea
								label="Note de clôture"
								rows="3"
								value={closingNote}
								oninput={(e) => (closingNote = inputValue(e))}
							></mb-textarea>
						</div>
						<div slot="footer">
							<mb-button type="submit" variant="primary" disabled={closing}>
								{closing ? 'Clôture…' : `Clôturer ${runLbl.article} ${runLbl.singular}`}
							</mb-button>
						</div>
					</mb-card>
				</form>
			{/if}
		</div>
	{/if}
</div>

<style>
	.progress-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin: 0 0 var(--mb-space-2);
	}
	.progress-row strong {
		font-size: var(--mb-font-size-lg);
	}
	.items {
		margin-top: var(--mb-space-4);
	}
	.items h3 {
		margin: 0 0 var(--mb-space-2);
		font-size: var(--mb-font-size-md);
	}
	.item-grid {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.item-row {
		display: flex;
		flex-wrap: wrap;
		gap: var(--mb-space-2);
		align-items: center;
		padding: var(--mb-space-2) 0;
		border-bottom: 1px solid var(--mb-color-border);
	}
	.label {
		flex: 1;
		min-width: 10rem;
	}
	.details {
		font-size: var(--mb-font-size-sm);
		color: var(--mb-color-muted);
	}
	abbr {
		color: var(--mb-color-danger);
		text-decoration: none;
	}
</style>
