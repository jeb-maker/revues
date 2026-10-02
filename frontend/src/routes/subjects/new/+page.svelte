<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import { createSubject } from '$lib/api/subjects';

	let csrf = $state('');
	let name = $state('');
	let description = $state('');
	let domains = $state('');
	let tags = $state('');
	let visibility = $state<'normal' | 'private'>('normal');
	let canSetVisibility = $state(false);
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
			// Editors may create; visibility only if org admin — probed via empty create capability on list.
			const { listSubjects } = await import('$lib/api/subjects');
			const list = await listSubjects({ csrfToken: csrf });
			if (!list.can_create) {
				error = 'Droits insuffisants pour créer un sujet.';
			}
			canSetVisibility = boot.user?.role === 'admin';
			ready = true;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
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
			const created = await createSubject(body, csrf);
			await goto(`/subjects/${created.id}`);
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
	<title>Nouveau sujet — Revues</title>
</svelte:head>

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<nav><a href="/subjects">← Sujets</a></nav>
	</header>

	<h1>Nouveau sujet</h1>
	<p class="lede">Nom, domaines de matching et étiquettes descriptives.</p>

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
					name="name"
					required
					value={name}
					oninput={(e: Event) => {
						name = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			<label class="field">
				<span>Description</span>
				<mb-textarea
					name="description"
					value={description}
					oninput={(e: Event) => {
						description = (e.target as HTMLTextAreaElement).value;
					}}
				></mb-textarea>
			</label>
			<label class="field">
				<span>Domaines (virgules)</span>
				<mb-input
					name="domains"
					placeholder="frontend, auth"
					value={domains}
					oninput={(e: Event) => {
						domains = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			<label class="field">
				<span>Étiquettes (virgules)</span>
				<mb-input
					name="tags"
					placeholder="prio, beta"
					value={tags}
					oninput={(e: Event) => {
						tags = (e.target as HTMLInputElement).value;
					}}
				></mb-input>
			</label>
			{#if canSetVisibility}
				<label class="field">
					<span>Visibilité</span>
					<select bind:value={visibility}>
						<option value="normal">Normal</option>
						<option value="private">Privé</option>
					</select>
				</label>
			{/if}
			<mb-button type="submit" variant="primary" disabled={loading || !csrf || !name.trim()}>
				{loading ? 'Création…' : 'Créer'}
			</mb-button>
		</form>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background: linear-gradient(165deg, #0f172a 0%, #1e293b 50%, #134e4a 100%);
		color: #f8fafc;
	}
	.page {
		max-width: 28rem;
		margin: 0 auto;
		padding: 1.5rem 1.25rem 3rem;
	}
	.top {
		display: flex;
		justify-content: space-between;
		margin-bottom: 1.5rem;
	}
	.brand {
		margin: 0;
		font-size: 1.35rem;
		font-weight: 700;
	}
	.brand a,
	nav a {
		color: #5eead4;
		text-decoration: none;
		font-weight: 600;
	}
	h1 {
		margin: 0 0 0.35rem;
		font-size: 1.5rem;
	}
	.lede {
		margin: 0 0 1.25rem;
		color: #cbd5e1;
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
	select {
		padding: 0.5rem;
		border-radius: 0.35rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: #f8fafc;
		font: inherit;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
