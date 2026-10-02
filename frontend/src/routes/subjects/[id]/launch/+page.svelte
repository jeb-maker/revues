<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { bootstrap } from '$lib/api/auth';
	import { getSubject, type SubjectDetail } from '$lib/api/subjects';
	import {
		createRun,
		listSubjectRunTemplates,
		type RunTemplateSummary
	} from '$lib/api/runs';

	let csrf = $state('');
	let subject = $state<SubjectDetail | null>(null);
	let templates = $state<RunTemplateSummary[]>([]);
	let canLaunch = $state(false);
	let templateId = $state('');
	let dueDate = $state('');
	let error = $state('');
	let loading = $state(true);
	let launching = $state(false);

	function subjectId(): number {
		return Number($page.params.id);
	}

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			subject = await getSubject(subjectId(), csrf);
			const res = await listSubjectRunTemplates(subjectId(), csrf);
			templates = res.templates ?? [];
			canLaunch = res.can_launch;
			if (templates.length === 1) {
				templateId = String(templates[0].id);
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger.';
		} finally {
			loading = false;
		}
	});

	async function onLaunch(e: Event) {
		e.preventDefault();
		if (!canLaunch || !templateId) return;
		launching = true;
		error = '';
		try {
			const body: { template_id: number; due_date?: string | null } = {
				template_id: Number(templateId)
			};
			if (dueDate.trim()) {
				body.due_date = dueDate.trim();
			}
			const run = await createRun(subjectId(), body, csrf);
			await goto(`/runs/${run.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Lancement impossible.';
		} finally {
			launching = false;
		}
	}
</script>

<svelte:head>
	<title>Lancer une revue — {subject?.name ?? 'Sujet'}</title>
</svelte:head>

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs">
			<a href="/subjects">Sujets</a>
			{#if subject}
				· <a href={`/subjects/${subject.id}`}>{subject.name}</a>
			{/if}
		</p>
		<h1>Lancer une revue</h1>
		<p class="lede">Snapshot transactionnel du modèle choisi — les points sont figés à ce moment.</p>
	</header>

	{#if error}
		<p class="err" role="alert">{error}</p>
	{/if}

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else if !canLaunch}
		<p class="muted">Vous n’avez pas les droits pour lancer une revue sur ce sujet.</p>
	{:else if templates.length === 0}
		<p class="muted">
			Aucun modèle compatible (domaines). Créez ou étendez un
			<a href="/modeles">modèle</a>.
		</p>
	{:else}
		<form class="form" onsubmit={onLaunch}>
			<label>
				Modèle
				<select bind:value={templateId} required>
					<option value="" disabled>Choisir…</option>
					{#each templates as t}
						<option value={String(t.id)}>
							{t.name} · v{t.latest_version} · {t.item_count} points
						</option>
					{/each}
				</select>
			</label>
			<label>
				Échéance (optionnel)
				<input type="date" bind:value={dueDate} />
			</label>
			<button type="submit" disabled={launching || !templateId}>
				{launching ? 'Lancement…' : 'Lancer'}
			</button>
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
		max-width: 36rem;
		margin: 0 auto;
		padding: 2rem 1.25rem 4rem;
	}
	.brand {
		margin: 0 0 0.35rem;
		font-size: 1.75rem;
		font-weight: 700;
	}
	.brand a {
		color: inherit;
		text-decoration: none;
	}
	.crumbs {
		margin: 0 0 0.5rem;
		font-size: 0.9rem;
		color: #94a3b8;
	}
	.crumbs a {
		color: #5eead4;
	}
	h1 {
		margin: 0 0 0.35rem;
		font-size: 1.25rem;
	}
	.lede {
		margin: 0 0 1.25rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
	}
	.form label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	.form select,
	.form input {
		padding: 0.55rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	.form button {
		align-self: flex-start;
		padding: 0.6rem 1rem;
		border: none;
		border-radius: 0.4rem;
		background: #0f766e;
		color: #ecfdf5;
		font-weight: 600;
		cursor: pointer;
		font: inherit;
	}
	.err {
		color: #fca5a5;
	}
	.muted {
		color: #94a3b8;
	}
	.muted a {
		color: #5eead4;
	}
</style>
