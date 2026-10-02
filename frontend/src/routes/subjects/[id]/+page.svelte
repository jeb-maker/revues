<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { bootstrap } from '$lib/api/auth';
	import {
		addSubjectMember,
		archiveSubject,
		getSubject,
		removeSubjectMember,
		updateSubject,
		type SubjectDetail
	} from '$lib/api/subjects';

	let csrf = $state('');
	let subject = $state<SubjectDetail | null>(null);
	let error = $state('');
	let loading = $state(true);
	let editing = $state(false);
	let saving = $state(false);

	let name = $state('');
	let description = $state('');
	let domains = $state('');
	let tags = $state('');
	let visibility = $state<'normal' | 'private'>('normal');

	let memberEmail = $state('');
	let memberRole = $state<'lead' | 'contributor' | 'viewer'>('viewer');

	function subjectId(): number {
		return Number($page.params.id);
	}

	function joinCSV(items: string[] | undefined): string {
		return (items ?? []).join(', ');
	}

	function splitCSV(raw: string): string[] {
		return raw
			.split(',')
			.map((s) => s.trim())
			.filter(Boolean);
	}

	function applySubject(s: SubjectDetail) {
		subject = s;
		name = s.name;
		description = s.description ?? '';
		domains = joinCSV(s.domains);
		tags = joinCSV(s.tags);
		visibility = s.visibility;
	}

	async function refresh() {
		const s = await getSubject(subjectId(), csrf);
		applySubject(s);
	}

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Sujet introuvable.';
		} finally {
			loading = false;
		}
	});

	async function onSave(e: Event) {
		e.preventDefault();
		if (!subject) return;
		saving = true;
		error = '';
		try {
			const body: Parameters<typeof updateSubject>[1] = {
				name: name.trim(),
				description: description.trim(),
				domains: splitCSV(domains),
				tags: splitCSV(tags)
			};
			if (subject.capabilities.can_set_visibility) {
				body.visibility = visibility;
			}
			const updated = await updateSubject(subject.id, body, csrf);
			applySubject(updated);
			editing = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
		} finally {
			saving = false;
		}
	}

	async function onArchive() {
		if (!subject || !confirm('Archiver ce sujet ?')) return;
		error = '';
		try {
			await archiveSubject(subject.id, csrf);
			await goto('/subjects');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Archivage impossible.';
		}
	}

	async function onAddMember(e: Event) {
		e.preventDefault();
		if (!subject) return;
		error = '';
		try {
			await addSubjectMember(
				subject.id,
				{ email: memberEmail.trim(), role: memberRole },
				csrf
			);
			memberEmail = '';
			await refresh();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Ajout membre impossible.';
		}
	}

	async function onRemoveMember(userId: number) {
		if (!subject || !confirm('Retirer ce membre ?')) return;
		error = '';
		try {
			await removeSubjectMember(subject.id, userId, csrf);
			await refresh();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Retrait impossible.';
		}
	}
</script>

<svelte:head>
	<title>{subject?.name ?? 'Sujet'} — Revues</title>
</svelte:head>

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<nav><a href="/subjects">← Sujets</a></nav>
	</header>

	{#if loading}
		<p class="muted"><mb-spinner></mb-spinner> Chargement…</p>
	{:else if !subject}
		<mb-alert variant="danger" role="alert">{error || 'Sujet introuvable.'}</mb-alert>
	{:else}
		{#if error}
			<mb-alert variant="danger" role="alert">{error}</mb-alert>
		{/if}

		<section class="hero">
			<h1>{subject.name}</h1>
			{#if subject.description}
				<p class="lede">{subject.description}</p>
			{/if}
			<p class="meta">
				{#if subject.visibility === 'private'}
					<mb-badge variant="warning">privé</mb-badge>
				{:else}
					<mb-badge>normal</mb-badge>
				{/if}
				{#if subject.access.role}
					<span class="role">rôle {subject.access.role}</span>
				{/if}
			</p>
			<p class="actions">
				{#if subject.capabilities.can_launch}
					<a class="launch" href={`/subjects/${subject.id}/launch`}>Lancer une revue</a>
				{/if}
				<a class="launch secondary" href="/runs">Voir les revues</a>
				{#if subject.capabilities.can_manage}
					<mb-button
						variant="secondary"
						onclick={() => {
							editing = !editing;
						}}
					>
						{editing ? 'Annuler' : 'Modifier'}
					</mb-button>
					<mb-button variant="danger" onclick={onArchive}>Archiver</mb-button>
				{/if}
			</p>
		</section>

		{#if editing}
			<form class="form" onsubmit={onSave}>
				<label class="field">
					<span>Nom</span>
					<mb-input
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
						value={description}
						oninput={(e: Event) => {
							description = (e.target as HTMLTextAreaElement).value;
						}}
					></mb-textarea>
				</label>
				<label class="field">
					<span>Domaines</span>
					<mb-input
						value={domains}
						oninput={(e: Event) => {
							domains = (e.target as HTMLInputElement).value;
						}}
					></mb-input>
				</label>
				<label class="field">
					<span>Étiquettes</span>
					<mb-input
						value={tags}
						oninput={(e: Event) => {
							tags = (e.target as HTMLInputElement).value;
						}}
					></mb-input>
				</label>
				{#if subject.capabilities.can_set_visibility}
					<label class="field">
						<span>Visibilité</span>
						<select bind:value={visibility}>
							<option value="normal">Normal</option>
							<option value="private">Privé</option>
						</select>
					</label>
				{/if}
				<mb-button type="submit" variant="primary" disabled={saving}>
					{saving ? 'Enregistrement…' : 'Enregistrer'}
				</mb-button>
			</form>
		{:else}
			<section class="block">
				<h2>Domaines</h2>
				{#if subject.domains.length === 0}
					<p class="muted">Aucun domaine.</p>
				{:else}
					<p class="tags">
						{#each subject.domains as d}
							<mb-tag>{d}</mb-tag>
						{/each}
					</p>
				{/if}
			</section>
			<section class="block">
				<h2>Étiquettes</h2>
				{#if subject.tags.length === 0}
					<p class="muted">Aucune étiquette.</p>
				{:else}
					<p class="tags">
						{#each subject.tags as t}
							<mb-tag>{t}</mb-tag>
						{/each}
					</p>
				{/if}
			</section>
		{/if}

		<section class="block">
			<h2>Membres directs</h2>
			{#if subject.members.length === 0}
				<p class="muted">Aucun membre direct (accès via org / équipes).</p>
			{:else}
				<ul class="members">
					{#each subject.members as m (m.user_id)}
						<li>
							<span class="mn">{m.display_name}</span>
							<span class="me">{m.email}</span>
							<mb-badge>{m.role}</mb-badge>
							{#if subject.capabilities.can_manage_members}
								<button
									type="button"
									class="linkish"
									onclick={() => onRemoveMember(m.user_id)}>Retirer</button
								>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}

			{#if subject.capabilities.can_manage_members}
				<form class="member-form" onsubmit={onAddMember}>
					<label class="field">
						<span>Email</span>
						<mb-input
							type="email"
							required
							value={memberEmail}
							oninput={(e: Event) => {
								memberEmail = (e.target as HTMLInputElement).value;
							}}
						></mb-input>
					</label>
					<label class="field">
						<span>Rôle</span>
						<select bind:value={memberRole}>
							<option value="viewer">Lecteur</option>
							<option value="contributor">Contributeur</option>
							<option value="lead">Responsable</option>
						</select>
					</label>
					<mb-button type="submit" variant="primary">Ajouter</mb-button>
				</form>
			{/if}
		</section>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background:
			radial-gradient(ellipse 60% 40% at 0% 0%, rgba(15, 118, 110, 0.3), transparent 55%),
			linear-gradient(165deg, #0f172a 0%, #1e293b 55%, #134e4a 100%);
		color: #f8fafc;
	}
	.page {
		max-width: 40rem;
		margin: 0 auto;
		padding: 1.25rem 1.25rem 3rem;
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
	.hero h1 {
		margin: 0 0 0.35rem;
		font-size: clamp(1.6rem, 4vw, 2rem);
		letter-spacing: -0.03em;
	}
	.lede {
		margin: 0 0 0.75rem;
		color: #cbd5e1;
	}
	.meta {
		display: flex;
		gap: 0.75rem;
		align-items: center;
		margin: 0 0 1rem;
	}
	.role {
		color: #94a3b8;
		font-size: 0.9rem;
	}
	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		align-items: center;
		margin: 0.75rem 0 1.5rem;
	}
	.launch {
		display: inline-flex;
		align-items: center;
		padding: 0.45rem 0.85rem;
		border-radius: 0.4rem;
		background: #0f766e;
		color: #ecfdf5;
		font-weight: 600;
		text-decoration: none;
		font-size: 0.9rem;
	}
	.launch.secondary {
		background: transparent;
		border: 1px solid #334155;
		color: #99f6e4;
	}
	.block {
		margin: 1.5rem 0;
	}
	.block h2 {
		margin: 0 0 0.5rem;
		font-size: 1rem;
		color: #99f6e4;
		font-weight: 600;
	}
	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
		margin: 0;
	}
	.form,
	.member-form {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
		margin: 1rem 0;
	}
	.member-form {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid #334155;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
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
	.members {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.65rem;
	}
	.members li {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		align-items: center;
	}
	.mn {
		font-weight: 600;
	}
	.me {
		color: #94a3b8;
		font-size: 0.85rem;
	}
	.linkish {
		background: none;
		border: none;
		color: #fca5a5;
		cursor: pointer;
		font: inherit;
		padding: 0;
	}
	.muted {
		color: #94a3b8;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
