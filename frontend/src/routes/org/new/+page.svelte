<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import { createOrganization } from '$lib/api/orgs';

	let name = $state('');
	let slug = $state('');
	let slugManual = $state(false);
	let csrf = $state('');
	let error = $state('');
	let loading = $state(false);
	let ready = $state(false);

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			if (boot.redirect && boot.redirect !== '/org/new' && boot.redirect !== '/') {
				await goto(boot.redirect);
				return;
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger la page.';
		} finally {
			ready = true;
		}
	});

	function onNameInput(e: Event) {
		name = (e.target as HTMLInputElement).value;
		if (!slugManual) {
			slug = name
				.toLowerCase()
				.replace(/[^a-z0-9]+/g, '-')
				.replace(/^-|-$/g, '');
		}
	}

	function onSlugInput(e: Event) {
		slugManual = true;
		slug = (e.target as HTMLInputElement).value;
	}

	async function onSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await createOrganization({ name, slug: slug || undefined }, csrf);
			await goto(res.redirect || '/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Création impossible.';
			try {
				const boot = await bootstrap();
				csrf = boot.csrf_token;
			} catch {
				/* ignore */
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Nouvelle organisation — Revues</title>
</svelte:head>

<main class="org">
	<p class="brand">Revues</p>
	<h1>Créer votre organisation</h1>
	<p class="lede">Créez votre organisation pour commencer à utiliser Revues.</p>

	{#if !ready}
		<p class="muted">Chargement…</p>
	{:else}
		{#if error}
			<mb-alert variant="danger" role="alert">{error}</mb-alert>
		{/if}

		<form class="form" onsubmit={onSubmit}>
			<label class="field">
				<span>Nom</span>
				<mb-input
					type="text"
					name="name"
					required
					value={name}
					oninput={onNameInput}
				></mb-input>
			</label>
			<label class="field">
				<span>Identifiant (slug)</span>
				<mb-input
					type="text"
					name="slug"
					value={slug}
					placeholder="mon-equipe"
					oninput={onSlugInput}
				></mb-input>
				<span class="hint">Lettres minuscules, chiffres et tirets. Pré-rempli depuis le nom.</span>
			</label>
			<mb-button type="submit" variant="primary" disabled={loading || !csrf || !name.trim()}>
				{loading ? 'Création…' : "Créer l'organisation"}
			</mb-button>
		</form>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background:
			radial-gradient(ellipse 80% 50% at 10% 0%, rgba(15, 118, 110, 0.35), transparent 55%),
			linear-gradient(165deg, #0f172a 0%, #1e293b 50%, #134e4a 100%);
		color: #f8fafc;
	}
	.org {
		max-width: 26rem;
		margin: 0 auto;
		padding: 12vh 1.25rem 3rem;
	}
	.brand {
		margin: 0 0 0.5rem;
		font-size: clamp(2.25rem, 7vw, 3.25rem);
		font-weight: 700;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	h1 {
		margin: 0 0 0.5rem;
		font-size: 1.15rem;
		font-weight: 500;
		color: #99f6e4;
	}
	.lede {
		margin: 0 0 1.5rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.875rem;
		color: #e2e8f0;
	}
	.hint {
		font-size: 0.8rem;
		color: #94a3b8;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
