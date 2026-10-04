<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { createTemplate } from '$lib/api/templates';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';
	import TemplateEditor from '$lib/components/TemplateEditor.svelte';
	import { emptyEditorItem, toWriteItems, type EditorItem } from '$lib/components/templateEditorModel';

	const boot = session();

	let name = $state('');
	let domains = $state('');
	let items = $state<EditorItem[]>([emptyEditorItem()]);
	let error = $state('');
	let itemsError = $state('');
	let saving = $state(false);

	onMount(async () => {
		if (!boot.can_edit) {
			await goto('/modeles');
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
			const detail = await createTemplate(boot.csrf_token, {
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
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Nouveau modèle — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs"><a href="/modeles">Modèles</a> · Nouveau</p>
		<h1>Nouveau modèle</h1>
		<p class="lede">Le modèle est utilisable dès sa création.</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	<form onsubmit={onSubmit}>
		<div class="card-stack">
			<mb-card>
				<h2 slot="header">Identité</h2>
				<div class="stack-form">
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
				</div>
			</mb-card>

			<mb-card>
				<h2 slot="header">Points</h2>
				<TemplateEditor bind:items error={itemsError} />
				<div slot="footer" class="actions">
					<mb-button type="submit" variant="primary" loading={saving}>
						{saving ? 'Création…' : 'Créer'}
					</mb-button>
					<a href="/modeles">Annuler</a>
				</div>
			</mb-card>
		</div>
	</form>
</div>
