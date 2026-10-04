<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getSubject, type SubjectDetail } from '$lib/api/subjects';
	import { createRun, listSubjectRunTemplates, type RunTemplateSummary } from '$lib/api/runs';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';

	const csrf = session().csrf_token;

	let subject = $state<SubjectDetail | null>(null);
	let templates = $state<RunTemplateSummary[]>([]);
	let canLaunch = $state(false);
	let templateId = $state('');
	let dueDate = $state('');
	let error = $state('');
	let loading = $state(true);
	let launching = $state(false);

	function subjectId(): number {
		return Number(page.params.id);
	}

	onMount(async () => {
		try {
			subject = await getSubject(subjectId(), csrf);
			const res = await listSubjectRunTemplates(subjectId(), csrf);
			templates = res.templates ?? [];
			canLaunch = res.can_launch;
			const preset = page.url.searchParams.get('template_id');
			if (preset && templates.some((t) => String(t.id) === preset)) {
				templateId = preset;
			} else if (templates.length === 1) {
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

<div class="page">
	<p class="crumbs">
		<a href="/subjects">Sujets</a>
		{#if subject}
			· <a href={`/subjects/${subject.id}`}>{subject.name}</a>
		{/if}
	</p>

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else}
		<form class="page--narrow" onsubmit={onLaunch}>
			<div class="card-stack">
				<mb-card>
					<h1 slot="header">Lancer une revue</h1>
					<div class="stack-form">
						<p class="lede">Les points du modèle choisi sont copiés tels quels au lancement.</p>

						{#if error}
							<mb-alert variant="danger">{error}</mb-alert>
						{/if}

						{#if !canLaunch}
							<p class="muted">Vous n’avez pas les droits pour lancer une revue sur ce sujet.</p>
						{:else if templates.length === 0}
							<p class="muted">
								Aucun modèle compatible (domaines). Créez ou étendez un
								<a href="/modeles">modèle</a>.
							</p>
						{:else}
							<mb-select
								label="Modèle"
								required
								placeholder="Choisir…"
								value={templateId}
								onmb-change={(e) => (templateId = e.detail.value)}
							>
								{#each templates as t (t.id)}
									<option value={String(t.id)}>{t.name} · v{t.latest_version} · {t.item_count} points</option>
								{/each}
							</mb-select>
							<mb-input
								label="Échéance"
								hint="Optionnel — déclenche le rappel J-1."
								type="date"
								value={dueDate}
								oninput={(e) => (dueDate = inputValue(e))}
							></mb-input>
						{/if}
					</div>
					<div slot="footer" class="actions actions--stack">
						{#if canLaunch && templates.length > 0}
							<mb-button type="submit" variant="primary" loading={launching}>
								{launching ? 'Lancement…' : 'Lancer'}
							</mb-button>
						{/if}
						{#if subject}
							<mb-button href={`/subjects/${subject.id}`} variant="secondary">Annuler</mb-button>
						{:else}
							<mb-button href="/subjects" variant="secondary">Annuler</mb-button>
						{/if}
					</div>
				</mb-card>
			</div>
		</form>
	{/if}
</div>
