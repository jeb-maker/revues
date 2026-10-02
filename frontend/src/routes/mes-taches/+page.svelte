<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap, type BootstrapResponse } from '$lib/api/auth';
	import { listMyTasks, type MyTask } from '$lib/api/mytasks';

	let boot = $state<BootstrapResponse | null>(null);
	let tasks = $state<MyTask[]>([]);
	let q = $state('');
	let status = $state<'' | 'pending' | 'ok' | 'nok' | 'na'>('');
	let error = $state('');
	let loading = $state(true);

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await listMyTasks({
				csrfToken: boot?.csrf_token,
				q: q.trim() || undefined,
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

	onMount(async () => {
		try {
			boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			await load();
		} catch (e) {
			error = e instanceof Error ? e.message : 'API indisponible';
			loading = false;
		}
	});

	async function onFilter(e: Event) {
		e.preventDefault();
		await load();
	}

	function statusLabel(s: string): string {
		switch (s) {
			case 'pending':
				return 'À faire';
			case 'ok':
				return 'OK';
			case 'nok':
				return 'NOK';
			case 'na':
				return 'N/A';
			default:
				return s;
		}
	}

	function statusVariant(s: string): 'neutral' | 'success' | 'danger' | 'warning' {
		switch (s) {
			case 'ok':
				return 'success';
			case 'nok':
				return 'danger';
			case 'na':
				return 'warning';
			default:
				return 'neutral';
		}
	}
</script>

<svelte:head>
	<title>Mes tâches — Revues</title>
</svelte:head>

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<nav class="nav">
			<a href="/subjects">Sujets</a>
			<a href="/runs">Revues</a>
			<a href="/mes-taches" aria-current="page">Mes tâches</a>
			{#if boot?.user}
				<span class="who">{boot.user.display_name}</span>
			{/if}
		</nav>
	</header>

	<section class="hero">
		<h1>Mes tâches</h1>
		<p class="lede">Points de revue qui vous sont assignés — filtrez et ouvrez le détail.</p>
	</section>

	<form class="filters" onsubmit={onFilter}>
		<label class="field grow">
			<span class="sr">Rechercher</span>
			<mb-input
				type="search"
				name="q"
				placeholder="Sujet, modèle, libellé…"
				value={q}
				oninput={(e: Event) => {
					q = (e.target as HTMLInputElement).value;
				}}
			></mb-input>
		</label>
		<label class="field">
			<span class="sr">Statut</span>
			<select bind:value={status} aria-label="Statut">
				<option value="">Tous</option>
				<option value="pending">À faire</option>
				<option value="ok">OK</option>
				<option value="nok">NOK</option>
				<option value="na">N/A</option>
			</select>
		</label>
		<mb-button type="submit" variant="secondary">Filtrer</mb-button>
	</form>

	{#if error}
		<mb-alert variant="danger" role="alert">{error}</mb-alert>
	{:else if loading}
		<p class="muted"><mb-spinner></mb-spinner> Chargement…</p>
	{:else if tasks.length === 0}
		<mb-empty-state label="Aucune tâche">
			Aucun point assigné avec ces filtres.
		</mb-empty-state>
	{:else}
		<p class="meta">{tasks.length} tâche{tasks.length > 1 ? 's' : ''}</p>
		<mb-table columns="2fr 1.2fr 1fr 0.8fr 0.9fr" sticky-header>
			<mb-table-row slot="head">
				<mb-table-cell>Point</mb-table-cell>
				<mb-table-cell>Sujet</mb-table-cell>
				<mb-table-cell>Revue</mb-table-cell>
				<mb-table-cell>Statut</mb-table-cell>
				<mb-table-cell></mb-table-cell>
			</mb-table-row>
			{#each tasks as task (task.id)}
				<mb-table-row>
					<mb-table-cell>
						<a class="name" href={`/runs/${task.run_id}/items/${task.id}`}>{task.label}</a>
						{#if task.section}
							<span class="desc">{task.section}</span>
						{/if}
					</mb-table-cell>
					<mb-table-cell>
						<a class="sub" href={`/subjects/${task.subject_id}`}>{task.subject_name}</a>
					</mb-table-cell>
					<mb-table-cell>
						<a class="sub" href={`/runs/${task.run_id}`}>{task.run_title}</a>
					</mb-table-cell>
					<mb-table-cell>
						<mb-badge variant={statusVariant(task.status)}>{statusLabel(task.status)}</mb-badge>
					</mb-table-cell>
					<mb-table-cell actions>
						<a href={`/runs/${task.run_id}/items/${task.id}`}>Ouvrir</a>
					</mb-table-cell>
				</mb-table-row>
			{/each}
		</mb-table>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background:
			radial-gradient(ellipse 70% 45% at 90% 0%, rgba(15, 118, 110, 0.28), transparent 50%),
			linear-gradient(165deg, #0f172a 0%, #1e293b 55%, #0f766e 120%);
		color: #f8fafc;
	}
	.page {
		max-width: 56rem;
		margin: 0 auto;
		padding: 1.25rem 1.25rem 3rem;
	}
	.top {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 1rem;
		margin-bottom: 2rem;
		flex-wrap: wrap;
	}
	.brand {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 700;
		letter-spacing: -0.03em;
	}
	.brand a {
		color: inherit;
		text-decoration: none;
	}
	.nav {
		display: flex;
		gap: 1rem;
		align-items: center;
		font-size: 0.9rem;
		flex-wrap: wrap;
	}
	.nav a {
		color: #5eead4;
		font-weight: 600;
		text-decoration: none;
	}
	.nav a[aria-current='page'] {
		text-decoration: underline;
		text-underline-offset: 0.2em;
	}
	.who {
		color: #94a3b8;
	}
	.hero h1 {
		margin: 0 0 0.35rem;
		font-size: clamp(1.75rem, 4vw, 2.25rem);
		letter-spacing: -0.03em;
	}
	.lede {
		margin: 0 0 1.25rem;
		color: #cbd5e1;
		max-width: 32rem;
	}
	.filters {
		display: flex;
		gap: 0.75rem;
		align-items: end;
		margin-bottom: 1.25rem;
		flex-wrap: wrap;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		min-width: 8rem;
	}
	.field.grow {
		flex: 1;
		min-width: 12rem;
	}
	.field select {
		padding: 0.5rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	.sr {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
	}
	.meta {
		margin: 0 0 0.5rem;
		color: #94a3b8;
		font-size: 0.9rem;
	}
	.name {
		color: #5eead4;
		font-weight: 600;
		text-decoration: none;
	}
	.sub {
		color: #cbd5e1;
		text-decoration: none;
		font-size: 0.9rem;
	}
	.sub:hover,
	.name:hover {
		text-decoration: underline;
	}
	.desc {
		display: block;
		color: #94a3b8;
		font-size: 0.85rem;
		margin-top: 0.15rem;
	}
	.muted {
		color: #94a3b8;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	mb-alert,
	mb-empty-state,
	mb-table {
		display: block;
		margin-top: 0.5rem;
	}
</style>
