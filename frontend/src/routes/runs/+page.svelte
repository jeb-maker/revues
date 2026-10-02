<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import { listRuns, type RunSummary } from '$lib/api/runs';

	let csrf = $state('');
	let runs = $state<RunSummary[]>([]);
	let total = $state(0);
	let q = $state('');
	let status = $state<'' | 'draft' | 'in_progress' | 'done' | 'overdue'>('');
	let error = $state('');
	let loading = $state(true);

	async function load() {
		loading = true;
		error = '';
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
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

	function statusLabel(s: string): string {
		switch (s) {
			case 'in_progress':
				return 'En cours';
			case 'done':
				return 'Terminée';
			case 'draft':
				return 'Brouillon';
			default:
				return s;
		}
	}
</script>

<svelte:head>
	<title>Revues — Revues</title>
</svelte:head>

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<h1>Revues</h1>
		<p class="lede">Exécutions en cours et historiques — progression par snapshot.</p>
	</header>

	{#if error}
		<p class="err" role="alert">{error}</p>
	{/if}

	<form class="toolbar" onsubmit={onFilter}>
		<input type="search" bind:value={q} placeholder="Rechercher…" aria-label="Recherche" />
		<select bind:value={status} aria-label="Statut">
			<option value="">Tous</option>
			<option value="in_progress">En cours</option>
			<option value="done">Terminées</option>
			<option value="overdue">En retard</option>
			<option value="draft">Brouillons</option>
		</select>
		<button type="submit">Filtrer</button>
		<a href="/subjects">Sujets</a>
	</form>

	{#if loading}
		<p class="muted"><mb-spinner></mb-spinner> Chargement…</p>
	{:else if runs.length === 0}
		<p class="muted">Aucune revue. Lancez-en une depuis un <a href="/subjects">sujet</a>.</p>
	{:else}
		<p class="meta">{total} revue{total > 1 ? 's' : ''}</p>
		<ul class="list">
			{#each runs as run (run.id)}
				<li>
					<a href={`/runs/${run.id}`}>
						<strong>{run.title}</strong>
						<span class="row">
							<span>{run.subject_name}</span>
							<mb-badge variant={run.status === 'done' ? 'success' : 'neutral'}
								>{statusLabel(run.status)}</mb-badge
							>
							<span class="pct">{run.progress.percent}%</span>
						</span>
						<div class="bar" aria-hidden="true">
							<span style={`width:${run.progress.percent}%`}></span>
						</div>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: var(--mb-font-sans, 'Segoe UI', system-ui, sans-serif);
		background: linear-gradient(160deg, #0f172a 0%, #1e293b 50%, #0f766e 100%);
		color: #f8fafc;
	}
	.page {
		max-width: 44rem;
		margin: 0 auto;
		padding: 2rem 1.25rem 4rem;
	}
	.brand {
		margin: 0 0 0.5rem;
		font-size: 1.75rem;
		font-weight: 700;
		letter-spacing: -0.03em;
	}
	.brand a {
		color: inherit;
		text-decoration: none;
	}
	h1 {
		margin: 0 0 0.5rem;
		font-size: 1.15rem;
		font-weight: 500;
		color: #99f6e4;
	}
	.lede {
		margin: 0 0 1.25rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	.toolbar {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1rem;
		align-items: center;
	}
	.toolbar input,
	.toolbar select {
		flex: 1;
		min-width: 8rem;
		padding: 0.5rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
	}
	.toolbar button,
	.toolbar a {
		padding: 0.5rem 0.75rem;
		border-radius: 0.4rem;
		border: none;
		background: #134e4a;
		color: #ccfbf1;
		font-weight: 600;
		text-decoration: none;
		cursor: pointer;
		font: inherit;
	}
	.list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.65rem;
	}
	.list a {
		display: block;
		padding: 0.85rem 1rem;
		border-radius: 0.5rem;
		background: rgba(15, 23, 42, 0.55);
		border: 1px solid #334155;
		color: inherit;
		text-decoration: none;
	}
	.list a:hover {
		border-color: #2dd4bf;
	}
	.row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.65rem;
		align-items: center;
		margin-top: 0.35rem;
		font-size: 0.9rem;
		color: #94a3b8;
	}
	.pct {
		margin-left: auto;
		font-variant-numeric: tabular-nums;
		color: #5eead4;
	}
	.bar {
		margin-top: 0.55rem;
		height: 0.35rem;
		border-radius: 999px;
		background: #1e293b;
		overflow: hidden;
	}
	.bar span {
		display: block;
		height: 100%;
		background: linear-gradient(90deg, #0d9488, #5eead4);
	}
	.err {
		color: #fca5a5;
	}
	.muted {
		color: #94a3b8;
	}
	.meta {
		margin: 0 0 0.75rem;
		color: #94a3b8;
		font-size: 0.9rem;
	}
</style>
