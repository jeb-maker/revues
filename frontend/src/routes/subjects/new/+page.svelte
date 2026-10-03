<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { createSubject, listSubjects } from '$lib/api/subjects';
	import { session } from '$lib/auth/session';
	import { formatVisibility } from '$lib/i18n/labels';
	import { inputValue } from '$lib/mb';

	const boot = session();
	// Seul l'admin global peut fixer la visibilité à la création (le serveur tranche).
	const canSetVisibility = boot.user?.role === 'admin';

	let name = $state('');
	let description = $state('');
	let domains = $state('');
	let tags = $state('');
	let visibility = $state<'normal' | 'private'>('normal');
	let error = $state('');
	let loading = $state(false);
	let ready = $state(false);

	onMount(async () => {
		try {
			const list = await listSubjects();
			if (!list.can_create) {
				error = 'Droits insuffisants pour créer un sujet.';
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
		} finally {
			ready = true;
		}
	});

	function splitCSV(raw: string): string[] {
		return raw
			.split(',')
			.map((s) => s.trim())
			.filter(Boolean);
	}

	async function onSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const body: Parameters<typeof createSubject>[0] = {
				name: name.trim(),
				description: description.trim(),
				domains: splitCSV(domains),
				tags: splitCSV(tags)
			};
			if (canSetVisibility) {
				body.visibility = visibility;
			}
			const created = await createSubject(body, boot.csrf_token);
			await goto(`/subjects/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Création impossible.';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Nouveau sujet — Revues</title>
</svelte:head>

<div class="page page--narrow">
	<header class="page-header">
		<p class="crumbs"><a href="/subjects">Sujets</a> · Nouveau</p>
		<h1>Nouveau sujet</h1>
		<p class="lede">Nom, domaines de matching et étiquettes descriptives.</p>
	</header>

	{#if !ready}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else}
		{#if error}
			<mb-alert variant="danger">{error}</mb-alert>
		{/if}

		<form class="stack-form" onsubmit={onSubmit}>
			<mb-input label="Nom" name="name" required value={name} oninput={(e) => (name = inputValue(e))}
			></mb-input>
			<mb-textarea
				label="Description"
				name="description"
				value={description}
				oninput={(e) => (description = inputValue(e))}
			></mb-textarea>
			<mb-input
				label="Domaines"
				hint="Séparés par des virgules, ex. frontend, auth. Servent au matching des modèles."
				name="domains"
				value={domains}
				oninput={(e) => (domains = inputValue(e))}
			></mb-input>
			<mb-input
				label="Étiquettes"
				hint="Séparées par des virgules, ex. prio, beta."
				name="tags"
				value={tags}
				oninput={(e) => (tags = inputValue(e))}
			></mb-input>
			{#if canSetVisibility}
				<mb-select
					label="Visibilité"
					name="visibility"
					value={visibility}
					onmb-change={(e) => (visibility = e.detail.value as 'normal' | 'private')}
				>
					<option value="normal">{formatVisibility('normal')}</option>
					<option value="private">{formatVisibility('private')}</option>
				</mb-select>
			{/if}
			<mb-button type="submit" variant="primary" disabled={loading || !name.trim()}>
				{loading ? 'Création…' : 'Créer'}
			</mb-button>
		</form>
	{/if}
</div>
