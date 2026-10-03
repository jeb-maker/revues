<script lang="ts">
	import { goto } from '$app/navigation';
	import { createOrganization } from '$lib/api/orgs';
	import { resetSession, session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';

	let name = $state('');
	let slug = $state('');
	let slugManual = $state(false);
	let error = $state('');
	let loading = $state(false);

	function onNameInput(e: Event) {
		name = inputValue(e);
		if (!slugManual) {
			slug = name
				.toLowerCase()
				.replace(/[^a-z0-9]+/g, '-')
				.replace(/^-|-$/g, '');
		}
	}

	function onSlugInput(e: Event) {
		slugManual = true;
		slug = inputValue(e);
	}

	async function onSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await createOrganization({ name, slug: slug || undefined }, session().csrf_token);
			resetSession();
			await goto(res.redirect || '/', { invalidateAll: true });
		} catch (err) {
			error = err instanceof Error ? err.message : 'Création impossible.';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Nouvelle organisation — Revues</title>
</svelte:head>

<div class="page page--narrow">
	<header class="page-header">
		<h1>Créer votre organisation</h1>
		<p class="lede">Créez votre organisation pour commencer à utiliser Revues.</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	<form class="stack-form" onsubmit={onSubmit}>
		<mb-input label="Nom" type="text" name="name" required value={name} oninput={onNameInput}
		></mb-input>
		<mb-input
			label="Identifiant (slug)"
			hint="Lettres minuscules, chiffres et tirets. Pré-rempli depuis le nom."
			type="text"
			name="slug"
			value={slug}
			placeholder="mon-equipe"
			oninput={onSlugInput}
		></mb-input>
		<mb-button type="submit" variant="primary" disabled={loading || !name.trim()}>
			{loading ? 'Création…' : "Créer l'organisation"}
		</mb-button>
	</form>
</div>
