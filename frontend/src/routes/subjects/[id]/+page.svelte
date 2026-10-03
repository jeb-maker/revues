<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		addSubjectMember,
		archiveSubject,
		getSubject,
		removeSubjectMember,
		updateSubject,
		type SubjectDetail
	} from '$lib/api/subjects';
	import { session } from '$lib/auth/session';
	import { formatRole, formatVisibility, roleOptions } from '$lib/i18n/labels';
	import { inputValue } from '$lib/mb';

	type SubjectRole = 'lead' | 'contributor' | 'viewer';
	const SUBJECT_ROLES = roleOptions<SubjectRole>(['viewer', 'contributor', 'lead']);

	const csrf = session().csrf_token;

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
	let memberRole = $state<SubjectRole>('viewer');

	function subjectId(): number {
		return Number(page.params.id);
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
		applySubject(await getSubject(subjectId(), csrf));
	}

	onMount(async () => {
		try {
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
			applySubject(await updateSubject(subject.id, body, csrf));
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
			await addSubjectMember(subject.id, { email: memberEmail.trim(), role: memberRole }, csrf);
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

<div class="page">
	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if !subject}
		<mb-alert variant="danger">{error || 'Sujet introuvable.'}</mb-alert>
	{:else}
		<header class="page-header">
			<p class="crumbs"><a href="/subjects">Sujets</a> · {subject.name}</p>
			<h1>{subject.name}</h1>
			{#if subject.description}
				<p class="lede">{subject.description}</p>
			{/if}
			<p class="actions">
				<mb-badge variant={subject.visibility === 'private' ? 'warning' : 'neutral'}>
					{formatVisibility(subject.visibility)}
				</mb-badge>
				{#if subject.access.role}
					<span class="muted">Votre rôle : {formatRole(subject.access.role)}</span>
				{/if}
			</p>
			<p class="actions">
				{#if subject.capabilities.can_launch}
					<mb-button variant="primary" href={`/subjects/${subject.id}/launch`}>Lancer une revue</mb-button>
				{/if}
				<mb-button variant="secondary" href="/runs">Voir les revues</mb-button>
				{#if subject.capabilities.can_manage}
					<mb-button variant="ghost" onclick={() => (editing = !editing)}>
						{editing ? 'Annuler' : 'Modifier'}
					</mb-button>
					<mb-button variant="danger" onclick={onArchive}>Archiver</mb-button>
				{/if}
			</p>
		</header>

		{#if error}
			<mb-alert variant="danger">{error}</mb-alert>
		{/if}

		{#if editing}
			<form class="stack-form" onsubmit={onSave}>
				<mb-input label="Nom" required value={name} oninput={(e) => (name = inputValue(e))}
				></mb-input>
				<mb-textarea
					label="Description"
					value={description}
					oninput={(e) => (description = inputValue(e))}
				></mb-textarea>
				<mb-input
					label="Domaines"
					hint="Séparés par des virgules."
					value={domains}
					oninput={(e) => (domains = inputValue(e))}
				></mb-input>
				<mb-input
					label="Étiquettes"
					hint="Séparées par des virgules."
					value={tags}
					oninput={(e) => (tags = inputValue(e))}
				></mb-input>
				{#if subject.capabilities.can_set_visibility}
					<mb-select
						label="Visibilité"
						required
						value={visibility}
						onmb-change={(e) => (visibility = e.detail.value as 'normal' | 'private')}
					>
						<option value="normal">{formatVisibility('normal')}</option>
						<option value="private">{formatVisibility('private')}</option>
					</mb-select>
				{/if}
				<mb-button type="submit" variant="secondary" disabled={saving}>
					{saving ? 'Enregistrement…' : 'Enregistrer'}
				</mb-button>
			</form>
		{:else}
			<section class="section" aria-labelledby="domaines">
				<h2 id="domaines">Domaines</h2>
				{#if subject.domains.length === 0}
					<p class="muted">Aucun domaine.</p>
				{:else}
					<p class="tags">
						{#each subject.domains as d (d)}
							<mb-tag>{d}</mb-tag>
						{/each}
					</p>
				{/if}
			</section>
			<section class="section" aria-labelledby="etiquettes">
				<h2 id="etiquettes">Étiquettes</h2>
				{#if subject.tags.length === 0}
					<p class="muted">Aucune étiquette.</p>
				{:else}
					<p class="tags">
						{#each subject.tags as t (t)}
							<mb-tag>{t}</mb-tag>
						{/each}
					</p>
				{/if}
			</section>
		{/if}

		<section class="section" aria-labelledby="membres">
			<h2 id="membres">Membres directs</h2>
			{#if subject.members.length === 0}
				<p class="muted">Aucun membre direct (accès via organisation / équipes).</p>
			{:else}
				<ul class="row-list">
					{#each subject.members as m (m.user_id)}
						<li>
							<strong>{m.display_name}</strong>
							<span class="muted">{m.email}</span>
							<mb-badge>{formatRole(m.role)}</mb-badge>
							{#if subject.capabilities.can_manage_members}
								<mb-button variant="ghost" size="sm" onclick={() => onRemoveMember(m.user_id)}>
									Retirer
								</mb-button>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}

			{#if subject.capabilities.can_manage_members}
				<form class="stack-form" onsubmit={onAddMember}>
					<mb-input
						label="Email du membre"
						type="email"
						required
						value={memberEmail}
						oninput={(e) => (memberEmail = inputValue(e))}
					></mb-input>
					<mb-select
						label="Rôle"
						required
						value={memberRole}
						onmb-change={(e) => (memberRole = e.detail.value as SubjectRole)}
					>
						{#each SUBJECT_ROLES as r (r.value)}
							<option value={r.value}>{r.label}</option>
						{/each}
					</mb-select>
					<mb-button type="submit" variant="secondary">Ajouter</mb-button>
				</form>
			{/if}
		</section>
	{/if}
</div>
