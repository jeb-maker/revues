<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import { createTemplate } from '$lib/api/templates';
	import TemplateEditor from '$lib/components/TemplateEditor.svelte';
	import {
		emptyEditorItem,
		toWriteItems,
		type EditorItem
	} from '$lib/components/templateEditorModel';

	let csrf = $state('');
	let name = $state('');
	let domains = $state('');
	let items = $state<EditorItem[]>([emptyEditorItem()]);
	let error = $state('');
	let itemsError = $state('');
	let saving = $state(false);
	let ready = $state(false);

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			const role = boot.user?.role;
			if (role !== 'admin' && role !== 'editor') {
				await goto('/modeles');
				return;
			}
			csrf = boot.csrf_token;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible';
		} finally {
			ready = true;
		}
	});

	async function onSubmit(e: Event) {
		e.preventDefault();
		error = '';
		itemsError = '';
		const writeItems = toWriteItems(items);
		if (!name.trim()) {
			error = 'Le nom est obligatoire.';
			return;
		}
		if (!writeItems.length) {
			itemsError = 'Ajoutez au moins un point au modèle.';
			return;
		}
		saving = true;
		try {
			const detail = await createTemplate(csrf, {
				name: name.trim(),
				domains: domains
					.split(',')
					.map((d) => d.trim())
					.filter(Boolean),
				items: writeItems
			});
			await goto(`/modeles/${detail.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Création impossible.';
			try {
				const boot = await bootstrap();
				csrf = boot.csrf_token;
			} catch {
				/* ignore */
			}
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Nouveau modèle — Revues</title>
</svelte:head>

<main class="page">
	<header>
		<p class="brand"><a href="/modeles">← Modèles</a></p>
		<h1>Nouveau modèle</h1>
		<p class="lede">La création publie immédiatement la version 1 (snapshot immuable).</p>
	</header>

	{#if !ready}
		<p class="muted">Chargement…</p>
	{:else}
		{#if error}
			<p class="err" role="alert">{error}</p>
		{/if}
		<form class="form" onsubmit={onSubmit}>
			<label class="field">
				<span>Nom</span>
				<input type="text" bind:value={name} required maxlength="200" />
			</label>
			<label class="field">
				<span>Domaines (virgules)</span>
				<input type="text" bind:value={domains} placeholder="ex. infra, ops" />
			</label>
			<section>
				<h2>Points</h2>
				<TemplateEditor bind:items error={itemsError} />
			</section>
			<div class="actions">
				<button type="submit" disabled={saving}>{saving ? 'Création…' : 'Créer'}</button>
				<a href="/modeles">Annuler</a>
			</div>
		</form>
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
		max-width: 48rem;
		margin: 0 auto;
		padding: 2rem 1.25rem 4rem;
	}
	.brand {
		margin: 0 0 0.5rem;
	}
	.brand a {
		color: #5eead4;
		text-decoration: none;
	}
	h1 {
		margin: 0 0 0.35rem;
		font-size: 1.35rem;
	}
	h2 {
		margin: 1rem 0 0.5rem;
		font-size: 1.05rem;
	}
	.lede {
		margin: 0 0 1.25rem;
		color: #cbd5e1;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		margin-bottom: 0.85rem;
	}
	.field input {
		padding: 0.5rem 0.65rem;
		border-radius: 0.35rem;
		border: 1px solid #334155;
		background: #0b1220;
		color: #f8fafc;
		font: inherit;
	}
	.actions {
		display: flex;
		gap: 0.75rem;
		align-items: center;
		margin-top: 1.25rem;
	}
	.actions button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.35rem;
		background: #0f766e;
		color: #ecfdf5;
		font: inherit;
		font-weight: 600;
		cursor: pointer;
	}
	.actions button:disabled {
		opacity: 0.6;
	}
	.actions a {
		color: #94a3b8;
	}
	.err {
		color: #fca5a5;
	}
	.muted {
		color: #94a3b8;
	}
</style>
