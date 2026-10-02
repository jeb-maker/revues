<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap, type BootstrapResponse } from '$lib/api/auth';
	import { listSubjects, type SubjectSummary } from '$lib/api/subjects';

	let boot = $state<BootstrapResponse | null>(null);
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

	async function onSearch(e: Event) {
		e.preventDefault();
		await load(q.trim());
	}
</script>

<svelte:head>
	<title>Sujets — Revues</title>
</svelte:head>

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<nav class="nav">
			<a href="/subjects" aria-current="page">Sujets</a>
			{#if boot?.user}
				<span class="who">{boot.user.display_name}</span>
			{/if}
		</nav>
	</header>

	<section class="hero">
		<h1>Sujets</h1>
		<p class="lede">Conteneurs de revues — domaines, étiquettes et membres.</p>
		{#if canCreate}
			<p class="actions">
				<a href="/subjects/new"><mb-button variant="primary">Nouveau sujet</mb-button></a>
			</p>
		{/if}
	</section>

	<form class="filters" onsubmit={onSearch}>
		<label class="field">
			<span class="sr">Rechercher</span>
			<mb-input
				type="search"
				name="q"
				placeholder="Filtrer par nom…"
				value={q}
				oninput={(e: Event) => {
					q = (e.target as HTMLInputElement).value;
				}}
			></mb-input>
		</label>
		<mb-button type="submit" variant="secondary">Filtrer</mb-button>
	</form>

	{#if error}
		<mb-alert variant="danger" role="alert">{error}</mb-alert>
	{:else if loading}
		<p class="muted"><mb-spinner></mb-spinner> Chargement…</p>
	{:else if subjects.length === 0}
		<mb-empty-state label="Aucun sujet">
			Aucun sujet visible dans l'organisation active.
		</mb-empty-state>
	{:else}
		<mb-table columns="2fr 1fr 1fr" sticky-header>
			<mb-table-row slot="head">
				<mb-table-cell>Nom</mb-table-cell>
				<mb-table-cell>Visibilité</mb-table-cell>
				<mb-table-cell></mb-table-cell>
			</mb-table-row>
			{#each subjects as s (s.id)}
				<mb-table-row>
					<mb-table-cell>
						<a class="name" href={`/subjects/${s.id}`}>{s.name}</a>
						{#if s.description}
							<span class="desc">{s.description}</span>
						{/if}
					</mb-table-cell>
					<mb-table-cell>
						{#if s.visibility === 'private'}
							<mb-badge variant="warning">privé</mb-badge>
						{:else}
							<mb-badge>normal</mb-badge>
						{/if}
					</mb-table-cell>
					<mb-table-cell actions>
						<a href={`/subjects/${s.id}`}>Ouvrir</a>
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
		max-width: 52rem;
		margin: 0 auto;
		padding: 1.25rem 1.25rem 3rem;
	}
	.top {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 1rem;
		margin-bottom: 2rem;
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
	}
	.nav a {
		color: #5eead4;
		font-weight: 600;
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
		margin: 0 0 1rem;
		color: #cbd5e1;
		max-width: 28rem;
	}
	.actions {
		margin: 0 0 1.25rem;
	}
	.actions a {
		text-decoration: none;
	}
	.filters {
		display: flex;
		gap: 0.75rem;
		align-items: end;
		margin-bottom: 1.25rem;
	}
	.field {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.sr {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
	}
	.name {
		color: #5eead4;
		font-weight: 600;
		text-decoration: none;
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
