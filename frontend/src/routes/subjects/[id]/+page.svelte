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
	import { listOrgDirectory, type OrgDirectoryResponse } from '$lib/api/orgs';
	import { listSubjectRuns, type RunSummary } from '$lib/api/runs';
	import { session } from '$lib/auth/session';
	import { formatRole, formatRunStatus, formatVisibility, roleOptions, runStatusVariant } from '$lib/i18n/labels';
	import { launchRunCTA, runLabels, subjectLabels } from '$lib/i18n/uiLabels';
	import { inputValue } from '$lib/mb';

	type SubjectRole = 'lead' | 'contributor' | 'viewer';
	const SUBJECT_ROLES = roleOptions<SubjectRole>(['viewer', 'contributor', 'lead']);

	const boot = session();
	const csrf = boot.csrf_token;
	const subjectLbl = $derived(subjectLabels(boot.organization?.ui_subject_label));
	const runLbl = $derived(runLabels(boot.organization?.ui_run_label));

	let subject = $state<SubjectDetail | null>(null);
	let runs = $state<RunSummary[]>([]);
	let orgPeople = $state<OrgDirectoryResponse['members']>([]);
	let error = $state('');
	let loading = $state(true);
	let editing = $state(false);
	let saving = $state(false);

	let name = $state('');
	let description = $state('');
	let domains = $state('');
	let visibility = $state<'normal' | 'private'>('normal');

	let memberPick = $state('');
	let memberEmail = $state('');
	let memberRole = $state<SubjectRole>('contributor');
	let showEmailFallback = $state(false);

	const candidatePeople = $derived.by(() => {
		const taken = new Set((subject?.members ?? []).map((m) => m.user_id));
		return orgPeople.filter((p) => !taken.has(p.user_id));
	});

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
		visibility = s.visibility;
	}

	async function refresh() {
		applySubject(await getSubject(subjectId(), csrf));
		try {
			const res = await listSubjectRuns(subjectId(), csrf);
			runs = res.runs ?? [];
		} catch {
			runs = [];
		}
	}

	onMount(async () => {
		try {
			await refresh();
			try {
				const dir = await listOrgDirectory();
				orgPeople = dir.members ?? [];
			} catch {
				orgPeople = [];
			}
		} catch (e) {
			error = e instanceof Error ? e.message : `${subjectLbl.singular} introuvable.`;
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
				domains: splitCSV(domains)
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
		const email =
			memberPick.trim() ||
			(showEmailFallback ? memberEmail.trim() : '');
		if (!email) {
			error = 'Choisissez une personne ou saisissez un email.';
			return;
		}
		try {
			await addSubjectMember(subject.id, { email, role: memberRole }, csrf);
			memberPick = '';
			memberEmail = '';
			showEmailFallback = false;
			await refresh();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Affectation impossible.';
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
			<p class="crumbs"><a href="/subjects">{subjectLbl.plural}</a> · {subject.name}</p>
			<h1>{subject.name}</h1>
			{#if subject.description && !editing}
				<p class="lede">{subject.description}</p>
			{/if}
			<p class="meta-row">
				<mb-badge variant={subject.visibility === 'private' ? 'warning' : 'neutral'}>
					{formatVisibility(subject.visibility)}
				</mb-badge>
				{#if subject.access.role}
					<span class="muted">Votre rôle : {formatRole(subject.access.role)}</span>
				{/if}
			</p>
		</header>

		{#if error}
			<mb-alert variant="danger">{error}</mb-alert>
		{/if}

		<div class="card-stack">
			<mb-card>
				<h2 slot="header">{runLbl.nav}</h2>
				{#if runs.length === 0}
					<p class="muted">{runLbl.noneArticle} {runLbl.singular} pour ce {subjectLbl.singular.toLowerCase()}.</p>
				{:else}
					<ul class="row-list">
						{#each runs as r (r.id)}
							<li>
								<a href={`/runs/${r.id}`}><strong>{r.title}</strong></a>
								<mb-badge variant={runStatusVariant(r.status)}>{formatRunStatus(r.status)}</mb-badge>
								<span class="muted">{r.progress.percent} %</span>
							</li>
						{/each}
					</ul>
				{/if}
				{#if subject.capabilities.can_launch || subject.capabilities.can_manage}
					<div slot="footer" class="actions">
						{#if subject.capabilities.can_launch && !editing}
							<mb-button variant="primary" href={`/subjects/${subject.id}/launch`}
								>{launchRunCTA(runLbl)}</mb-button
							>
						{/if}
						{#if subject.capabilities.can_manage}
							<mb-button variant="secondary" onclick={() => (editing = !editing)}>
								{editing ? 'Annuler' : 'Modifier'}
							</mb-button>
						{/if}
					</div>
				{/if}
			</mb-card>

			{#if editing}
				<form onsubmit={onSave}>
					<mb-card>
						<h2 slot="header">Modifier</h2>
						<div class="stack-form">
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
						</div>
						<div slot="footer">
							<mb-button type="submit" variant="primary" loading={saving}>
								{saving ? 'Enregistrement…' : 'Enregistrer'}
							</mb-button>
						</div>
					</mb-card>
				</form>
			{:else}
				<mb-card>
					<h2 slot="header">Domaines</h2>
					{#if subject.domains.length === 0}
						<p class="muted">Aucun domaine — tous les modèles sont compatibles.</p>
					{:else}
						<p class="tags">
							{#each subject.domains as d (d)}
								<mb-tag>{d}</mb-tag>
							{/each}
						</p>
					{/if}
				</mb-card>
			{/if}

			<mb-card>
				<h2 slot="header">Personnes</h2>
				<p class="muted">
					Personnes affectées à ce {subjectLbl.singular.toLowerCase()}.
				</p>
				{#if subject.members.length === 0}
					<p class="muted">
						Aucune personne affectée. Affectez un membre de l’organisation{#if boot.can_admin},
							ou invitez-le d’abord dans l’org (<a href="/admin/members">Admin · Membres</a>){/if}.
					</p>
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
						{#if candidatePeople.length > 0}
							<mb-select
								label="Personne (membres de l’organisation)"
								value={memberPick}
								onmb-change={(e) => {
									memberPick = e.detail.value;
									if (memberPick) showEmailFallback = false;
								}}
							>
								<option value="">Choisir…</option>
								{#each candidatePeople as p (p.user_id)}
									<option value={p.email}>{p.display_name || p.email} · {p.email}</option>
								{/each}
							</mb-select>
						{:else}
							<p class="field-hint">
								Tous les membres org sont déjà affectés, ou l’annuaire est vide.
							</p>
						{/if}
						<mb-select
							label="Rôle sur le {subjectLbl.singular.toLowerCase()}"
							required
							value={memberRole}
							onmb-change={(e) => (memberRole = e.detail.value as SubjectRole)}
						>
							{#each SUBJECT_ROLES as r (r.value)}
								<option value={r.value}>{r.label}</option>
							{/each}
						</mb-select>
						{#if showEmailFallback}
							<mb-input
								label="Email (compte existant hors liste)"
								type="email"
								value={memberEmail}
								oninput={(e) => (memberEmail = inputValue(e))}
							></mb-input>
						{:else}
							<p class="actions">
								<mb-button
									type="button"
									variant="ghost"
									size="sm"
									onclick={() => {
										showEmailFallback = true;
										memberPick = '';
									}}
								>
									Ajouter par email
								</mb-button>
							</p>
						{/if}
						<mb-button type="submit" variant="secondary">Affecter</mb-button>
					</form>
				{/if}
			</mb-card>

			{#if subject.capabilities.can_manage}
				<mb-card>
					<h2 slot="header">Archiver</h2>
					<p class="muted">Le {subjectLbl.singular.toLowerCase()} ne sera plus disponible pour de nouvelles revues.</p>
					<div slot="footer">
						<mb-button variant="danger" onclick={onArchive}>Archiver</mb-button>
					</div>
				</mb-card>
			{/if}
		</div>
	{/if}
</div>

<style>
	h3 {
		margin: var(--mb-space-3) 0 var(--mb-space-2);
		font-size: var(--mb-font-size-md);
	}
	section:first-of-type h3 {
		margin-top: 0;
	}
</style>
