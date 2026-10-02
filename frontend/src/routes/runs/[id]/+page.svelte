<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { bootstrap } from '$lib/api/auth';
	import { completeRun, getRun, type RunDetail, type RunItem } from '$lib/api/runs';

	let csrf = $state('');
	let run = $state<RunDetail | null>(null);
	let error = $state('');
	let loading = $state(true);
	let closing = $state(false);
	let closingNote = $state('');

	function runId(): number {
		return Number($page.params.id);
	}

	async function refresh() {
		run = await getRun(runId(), csrf);
		closingNote = run.closing_note ?? '';
	}

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			await refresh();
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

	function statusVariant(s: string): string {
		switch (s) {
			case 'ok':
				return 'success';
			case 'nok':
				return 'danger';
			case 'na':
				return 'neutral';
			default:
				return 'warning';
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

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs">
			<a href="/runs">Revues</a>
			{#if run}
				· <a href={`/subjects/${run.subject_id}`}>{run.subject_name}</a>
			{/if}
		</p>
		{#if run}
			<h1>{run.title}</h1>
			<p class="lede">
				{run.template_name} · v{run.template_version} ·
				{#if run.status === 'done'}
					terminée
				{:else if run.status === 'in_progress'}
					en cours
				{:else}
					{run.status}
				{/if}
			</p>
		{:else}
			<h1>Revue</h1>
		{/if}
	</header>

	{#if error}
		<p class="err" role="alert">{error}</p>
	{/if}

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else if run}
		<section class="progress" aria-label="Progression">
			<div class="pct">
				<strong>{run.progress.percent}%</strong>
				<span>{run.progress.done}/{run.progress.total} traités</span>
			</div>
			<div class="bar"><span style={`width:${run.progress.percent}%`}></span></div>
			{#if (run.pending_required_count ?? 0) > 0}
				<p class="warn">{run.pending_required_count} point(s) obligatoire(s) en attente</p>
			{/if}
		</section>

		{#each groupBySection(run.items) as group}
			<section class="section">
				<h2>{group.section}</h2>
				<ul>
					{#each group.items as item (item.id)}
						<li>
							<a href={`/runs/${run.id}/items/${item.id}`}>
								<span class="label">
									{item.label}
									{#if item.required}<em>*</em>{/if}
								</span>
								<mb-badge variant={statusVariant(item.status)}>{item.status}</mb-badge>
								{#if item.assigned_login}
									<span class="assignee">@{item.assigned_login}</span>
								{/if}
							</a>
						</li>
					{/each}
				</ul>
			</section>
		{/each}

		{#if run.capabilities.can_complete}
			<form class="complete" onsubmit={onComplete}>
				<label>
					Note de clôture
					<textarea bind:value={closingNote} rows="3"></textarea>
				</label>
				<button type="submit" disabled={closing}>
					{closing ? 'Clôture…' : 'Clôturer la revue'}
				</button>
			</form>
		{:else if run.status === 'done' && run.closing_note}
			<p class="note">Note : {run.closing_note}</p>
		{/if}
	{/if}
</main>

<style>
	
	
	
	
	.progress {
		margin-bottom: 1.5rem;
		padding: 1rem;
		border-radius: 0.5rem;
		background: rgba(15, 23, 42, 0.55);
		border: 1px solid #334155;
	}
	.pct {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 0.5rem;
	}
	.pct strong {
		font-size: 1.5rem;
		color: #5eead4;
	}
	.bar {
		height: 0.45rem;
		border-radius: 999px;
		background: #1e293b;
		overflow: hidden;
	}
	.bar span {
		display: block;
		height: 100%;
		background: linear-gradient(90deg, #0d9488, #5eead4);
	}
	.warn {
		margin: 0.65rem 0 0;
		color: #fcd34d;
		font-size: 0.9rem;
	}
	.section {
		margin-bottom: 1.25rem;
	}
	.section h2 {
		margin: 0 0 0.5rem;
		font-size: 0.95rem;
		color: #99f6e4;
		font-weight: 600;
	}
	.section ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}
	.section a {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		align-items: center;
		padding: 0.65rem 0.85rem;
		border-radius: 0.45rem;
		background: rgba(15, 23, 42, 0.45);
		border: 1px solid #334155;
		color: inherit;
		text-decoration: none;
	}
	.section a:hover {
		border-color: #2dd4bf;
	}
	.label {
		flex: 1;
		min-width: 10rem;
	}
	.label em {
		color: #fcd34d;
		font-style: normal;
	}
	.assignee {
		font-size: 0.85rem;
		color: #94a3b8;
	}
	.complete {
		margin-top: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.complete label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	.complete textarea {
		padding: 0.6rem;
		border-radius: 0.4rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	.complete button {
		align-self: flex-start;
		padding: 0.6rem 1rem;
		border: none;
		border-radius: 0.4rem;
		background: #0f766e;
		color: #ecfdf5;
		font-weight: 600;
		cursor: pointer;
		font: inherit;
	}
	.complete button:disabled {
		opacity: 0.6;
		cursor: wait;
	}
	.note {
		margin-top: 1rem;
		color: #cbd5e1;
	}
	
</style>
