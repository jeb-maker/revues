<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { bootstrap } from '$lib/api/auth';
	import { archiveTemplate, getTemplate, type TemplateDetail } from '$lib/api/templates';

	let csrf = $state('');
	let detail = $state<TemplateDetail | null>(null);
	let error = $state('');
	let loading = $state(true);

	onMount(async () => {
		const id = Number($page.params.id);
		if (!Number.isFinite(id) || id <= 0) {
			error = 'Identifiant invalide.';
			loading = false;
			return;
		}
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			detail = await getTemplate(csrf, id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible';
		} finally {
			loading = false;
		}
	});

	async function onArchive() {
		if (!detail || !detail.can_manage) return;
		if (!confirm('Archiver ce modèle ?')) return;
		try {
			await archiveTemplate(csrf, detail.id);
			await goto('/modeles');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Archivage impossible';
		}
	}
</script>

<svelte:head>
	<title>{detail?.name ?? 'Modèle'} — Revues</title>
</svelte:head>

<main class="page">
	<p class="brand"><a href="/modeles">← Modèles</a></p>

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else if error}
		<p class="err" role="alert">{error}</p>
	{:else if detail}
		<header>
			<h1>{detail.name}</h1>
			<p class="meta">
				Version v{detail.version.version} · publiée {detail.version.published_at}
				{#if detail.domains?.length}
					· domaines : {detail.domains.join(', ')}
				{/if}
			</p>
			{#if detail.can_manage}
				<p class="actions">
					<a class="primary" href={`/modeles/${detail.id}/edit`}>Modifier</a>
					<button type="button" onclick={onArchive}>Archiver</button>
				</p>
			{/if}
		</header>

		<section>
			<h2>Points ({detail.items.length})</h2>
			{#each detail.items as item, i}
				{#if i === 0 || item.section !== detail.items[i - 1].section}
					{#if item.section}
						<h3>{item.section}</h3>
					{/if}
				{/if}
				<div class="item">
					<span class="pos">{item.position}.</span>
					<span class="label"
						>{item.label}{#if item.required}<span class="req">*</span>{/if}</span
					>
					{#if item.help_text}
						<span class="help">{item.help_text}</span>
					{/if}
				</div>
			{/each}
		</section>
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
	.brand a {
		color: #5eead4;
		text-decoration: none;
	}
	h1 {
		margin: 0.5rem 0 0.35rem;
		font-size: 1.5rem;
	}
	h2 {
		margin: 1.5rem 0 0.75rem;
		font-size: 1.1rem;
	}
	h3 {
		margin: 1rem 0 0.35rem;
		font-size: 0.95rem;
		color: #99f6e4;
	}
	.meta {
		margin: 0 0 1rem;
		color: #94a3b8;
		font-size: 0.9rem;
	}
	.actions {
		display: flex;
		gap: 0.75rem;
		margin: 0 0 1rem;
	}
	.actions a,
	.actions button {
		padding: 0.45rem 0.8rem;
		border-radius: 0.35rem;
		border: 1px solid #334155;
		background: transparent;
		color: #5eead4;
		text-decoration: none;
		font: inherit;
		cursor: pointer;
	}
	.actions a.primary {
		background: #0f766e;
		border-color: #0f766e;
		color: #ecfdf5;
		font-weight: 600;
	}
	.item {
		display: grid;
		grid-template-columns: 2rem 1fr;
		gap: 0.15rem 0.5rem;
		padding: 0.45rem 0;
		border-bottom: 1px solid color-mix(in srgb, #64748b 35%, transparent);
	}
	.pos {
		color: #64748b;
	}
	.req {
		color: #fbbf24;
		margin-left: 0.15rem;
	}
	.help {
		grid-column: 2;
		font-size: 0.85rem;
		color: #94a3b8;
	}
	.err {
		color: #fca5a5;
	}
	.muted {
		color: #94a3b8;
	}
</style>
