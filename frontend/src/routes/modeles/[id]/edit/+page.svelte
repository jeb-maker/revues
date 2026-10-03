<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getTemplate, saveTemplate } from '$lib/api/templates';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';
	import TemplateEditor from '$lib/components/TemplateEditor.svelte';
	import {
		itemsFromDetail,
		toWriteItems,
		type EditorItem
	} from '$lib/components/templateEditorModel';

	const boot = session();
	const csrf = boot.csrf_token;

	let templateId = $state(0);
	let name = $state('');
	let domains = $state('');
	let versionLabel = $state('');
	let items = $state<EditorItem[]>([]);
	let error = $state('');
	let itemsError = $state('');
	let saving = $state(false);
	let ready = $state(false);

	onMount(async () => {
		const id = Number(page.params.id);
		if (!Number.isFinite(id) || id <= 0) {
			error = 'Identifiant invalide.';
			ready = true;
			return;
		}
		templateId = id;
		if (!boot.can_edit) {
			await goto(`/modeles/${id}`);
			return;
		}
		try {
			const detail = await getTemplate(csrf, id);
			if (!detail.can_manage) {
				await goto(`/modeles/${id}`);
				return;
			}
			name = detail.name;
			domains = (detail.domains ?? []).join(', ');
			versionLabel = `v${detail.version.version}`;
			items = itemsFromDetail(detail.items);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
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
			const detail = await saveTemplate(csrf, templateId, {
				name: name.trim(),
				domains: domains
					.split(',')
					.map((d) => d.trim())
					.filter(Boolean),
				items: writeItems
			});
			await goto(`/modeles/${detail.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Modifier {name || 'modèle'} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs">
			<a href="/modeles">Modèles</a> · <a href={`/modeles/${templateId || ''}`}>{name || 'Modèle'}</a> · Modifier
		</p>
		<h1>Modifier le modèle</h1>
		<p class="lede">
			Version actuelle {versionLabel}. L’enregistrement publie une <strong>nouvelle version</strong> ;
			les revues déjà lancées conservent la version figée.
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	{#if !ready}
		<p class="loading">Chargement…</p>
	{:else}
		<form class="stack-form" onsubmit={onSubmit}>
			<mb-input
				label="Nom"
				type="text"
				required
				maxlength="200"
				value={name}
				oninput={(e) => (name = inputValue(e))}
			></mb-input>
			<mb-input
				label="Domaines"
				hint="Séparés par des virgules, ex. infra, ops."
				type="text"
				value={domains}
				oninput={(e) => (domains = inputValue(e))}
			></mb-input>
			<section>
				<h2>Points</h2>
				<TemplateEditor bind:items error={itemsError} />
			</section>
			<p class="actions">
				<mb-button type="submit" variant="primary" disabled={saving}>
					{saving ? 'Enregistrement…' : 'Enregistrer (nouvelle version)'}
				</mb-button>
				<a href={`/modeles/${templateId}`}>Annuler</a>
			</p>
		</form>
	{/if}
</div>
