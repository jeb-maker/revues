<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { archiveTemplate, getTemplate, type TemplateDetail } from '$lib/api/templates';
	import { session } from '$lib/auth/session';

	const csrf = session().csrf_token;

	let detail = $state<TemplateDetail | null>(null);
	let error = $state('');
	let loading = $state(true);

	onMount(async () => {
		const id = Number(page.params.id);
		if (!Number.isFinite(id) || id <= 0) {
			error = 'Identifiant invalide.';
			loading = false;
			return;
		}
		try {
			detail = await getTemplate(csrf, id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
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
			error = e instanceof Error ? e.message : 'Archivage impossible.';
		}
	}
</script>

<svelte:head>
	<title>{detail?.name ?? 'Modèle'} — Revues</title>
</svelte:head>

<div class="page">
	<p class="crumbs"><a href="/modeles">Modèles</a> · {detail?.name ?? 'Modèle'}</p>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	{#if loading}
		<p class="loading">Chargement…</p>
	{:else if detail}
		<header class="page-header">
			<h1>{detail.name}</h1>
			<p class="muted">
				Version v{detail.version.version} · publiée le {detail.version.published_at}
				{#if detail.domains?.length}
					· domaines : {detail.domains.join(', ')}
				{/if}
			</p>
			{#if detail.can_manage}
				<p class="actions">
					<mb-button href={`/modeles/${detail.id}/edit`} variant="primary">Modifier</mb-button>
					<mb-button variant="danger" onclick={onArchive}>Archiver</mb-button>
				</p>
			{/if}
		</header>

		<section class="section">
			<h2>Points ({detail.items.length})</h2>
			<ol class="items">
				{#each detail.items as item, i (item.position)}
					{#if (i === 0 || item.section !== detail.items[i - 1].section) && item.section}
						<li class="items__section" aria-hidden="true"><h3>{item.section}</h3></li>
					{/if}
					<li class="item">
						<span class="item__pos">{item.position}.</span>
						<span class="item__label">
							{item.label}{#if item.required}<span class="item__req" title="Obligatoire">*</span>{/if}
						</span>
						{#if item.help_text}
							<span class="item__help">{item.help_text}</span>
						{/if}
					</li>
				{/each}
			</ol>
		</section>
	{/if}
</div>

<style>
	.items {
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.items__section h3 {
		margin: var(--mb-space-4) 0 var(--mb-space-1);
	}
	.item {
		display: grid;
		grid-template-columns: 2.25rem 1fr;
		gap: var(--mb-space-1) var(--mb-space-2);
		padding: var(--mb-space-2) 0;
		border-bottom: 1px solid var(--mb-color-border);
	}
	.item__pos {
		color: var(--mb-color-muted);
	}
	.item__req {
		margin-left: 0.15rem;
		color: var(--mb-color-danger);
	}
	.item__help {
		grid-column: 2;
		font-size: var(--mb-font-size-sm);
		color: var(--mb-color-muted);
	}
</style>
