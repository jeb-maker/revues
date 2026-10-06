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
	import '$lib/styles/subject-detail.css';

	type SubjectRole = 'lead' | 'contributor' | 'viewer';
	type SortKey = 'titre' | 'date' | 'statut' | 'progression';
	type SortDirection = 'asc' | 'desc';
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
	let sortKey = $state<SortKey>('date');
	let sortDirection = $state<SortDirection>('desc');

	let name = $state('');
	let description = $state('');
	let domains = $state('');
	let visibility = $state<'normal' | 'private'>('normal');

	let memberPick = $state('');
	let memberEmail = $state('');
	let memberRole = $state<SubjectRole>('contributor');
	let showEmailFallback = $state(false);
	let memberBusyId = $state<number | 'new' | null>(null);

	const candidatePeople = $derived.by(() => {
		const taken = new Set((subject?.members ?? []).map((m) => m.user_id));
		return orgPeople.filter((p) => !taken.has(p.user_id));
	});

	const selectedCandidate = $derived(
		candidatePeople.find((p) => p.email === memberPick) ?? null
	);

	/** Domaines visibles hors édition : multi-sujet, ou déjà renseignés. */
	const showDomainsRead = $derived(
		!!subject && (boot.show_subject_column || subject.domains.length > 0)
	);

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

	function formatCreatedDate(iso: string): string {
		const t = Date.parse(iso);
		if (Number.isNaN(t)) return iso || '—';
		return new Date(t).toLocaleDateString('fr-FR');
	}

	function onRunsSort(e: CustomEvent<{ key: string; direction: string }>) {
		const key = e.detail.key as SortKey;
		const direction = e.detail.direction === 'asc' ? 'asc' : 'desc';
		if (key !== 'titre' && key !== 'date' && key !== 'statut' && key !== 'progression') {
			return;
		}
		sortKey = key;
		sortDirection = direction;
	}

	const displayedRuns = $derived.by(() => {
		const list = [...runs];
		const dir = sortDirection === 'asc' ? 1 : -1;
		list.sort((a, b) => {
			switch (sortKey) {
				case 'titre':
					return a.title.localeCompare(b.title, 'fr') * dir;
				case 'statut':
					return a.status.localeCompare(b.status) * dir;
				case 'progression':
					return (a.progress.percent - b.progress.percent) * dir;
				case 'date':
				default:
					return a.created_at.localeCompare(b.created_at) * dir;
			}
		});
		return list;
	});

	function applySubject(s: SubjectDetail) {
		subject = s;
		name = s.name;
		description = s.description ?? '';
		domains = joinCSV(s.domains);
		visibility = s.visibility;
	}

	function startEdit() {
		if (!subject) return;
		applySubject(subject);
		editing = true;
		error = '';
	}

	function cancelEdit() {
		if (subject) applySubject(subject);
		editing = false;
		error = '';
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
			if (boot.show_collab) {
				try {
					const dir = await listOrgDirectory();
					orgPeople = dir.members ?? [];
				} catch {
					orgPeople = [];
				}
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
		if (!subject || !confirm(`Archiver ce ${subjectLbl.singular.toLowerCase()} ?`)) return;
		error = '';
		try {
			await archiveSubject(subject.id, csrf);
			await goto('/subjects');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Archivage impossible.';
		}
	}

	async function onAddMember() {
		if (!subject) return;
		error = '';
		const email = (showEmailFallback ? memberEmail : memberPick).trim();
		if (!email) {
			error = 'Choisissez une personne ou saisissez un email.';
			return;
		}
		memberBusyId = 'new';
		try {
			await addSubjectMember(subject.id, { email, role: memberRole }, csrf);
			memberPick = '';
			memberEmail = '';
			showEmailFallback = false;
			memberRole = 'contributor';
			await refresh();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Ajout impossible.';
		} finally {
			memberBusyId = null;
		}
	}

	async function onMemberRoleChange(userId: number, email: string, role: SubjectRole) {
		if (!subject) return;
		const current = subject.members.find((m) => m.user_id === userId);
		if (!current || current.role === role) return;
		error = '';
		memberBusyId = userId;
		try {
			// POST upsert : même endpoint que l'ajout, met à jour le rôle.
			await addSubjectMember(subject.id, { email, role }, csrf);
			await refresh();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Changement de rôle impossible.';
			await refresh();
		} finally {
			memberBusyId = null;
		}
	}

	async function onRemoveMember(userId: number) {
		if (!subject || !confirm('Retirer cette personne ?')) return;
		error = '';
		memberBusyId = userId;
		try {
			await removeSubjectMember(subject.id, userId, csrf);
			await refresh();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Retrait impossible.';
		} finally {
			memberBusyId = null;
		}
	}
</script>

<svelte:head>
	<title>{subject?.name ?? subjectLbl.singular} — Revues</title>
</svelte:head>

<div class="page subject-detail">
	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if !subject}
		<p class="crumbs"><a href="/subjects">{subjectLbl.plural}</a></p>
		<mb-alert variant="danger">{error || `${subjectLbl.singular} introuvable.`}</mb-alert>
	{:else}
		<p class="crumbs">
			<a href="/subjects">{subjectLbl.plural}</a> · {subject.name}
		</p>

		<header class="page-header">
			<div class="page-header__row">
				<h1>{subject.name}</h1>
				{#if subject.capabilities.can_manage && !editing}
					<div class="page-header__actions">
						<mb-button variant="secondary" onclick={startEdit}>Modifier</mb-button>
					</div>
				{/if}
			</div>
			{#if subject.description && !editing}
				<p class="lede">{subject.description}</p>
			{/if}
			<p class="meta-row">
				{#if subject.visibility === 'private'}
					<mb-badge variant="warning">{formatVisibility('private')}</mb-badge>
				{/if}
				{#if subject.access.role}
					<span class="muted">Votre rôle : {formatRole(subject.access.role)}</span>
				{/if}
				{#if !editing && showDomainsRead && subject.domains.length > 0}
					<span class="meta-domains" aria-label="Domaines">
						{#each subject.domains as d (d)}
							<mb-tag>{d}</mb-tag>
						{/each}
					</span>
				{/if}
			</p>
		</header>

		{#if error}
			<mb-alert variant="danger">{error}</mb-alert>
		{/if}

		<div class="stack">
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
							<details class="advanced">
								<summary>Options avancées</summary>
								<div class="stack-form advanced__body">
									<mb-input
										label="Domaines"
										hint="Séparés par des virgules. Associent les modèles compatibles (intersection)."
										value={domains}
										oninput={(e) => (domains = inputValue(e))}
									></mb-input>
								</div>
							</details>
						</div>
						<div slot="footer" class="actions">
							<mb-button type="submit" variant="primary" loading={saving}>
								{saving ? 'Enregistrement…' : 'Enregistrer'}
							</mb-button>
							<mb-button type="button" variant="secondary" onclick={cancelEdit} disabled={saving}>
								Annuler
							</mb-button>
						</div>
					</mb-card>
				</form>
			{/if}

			<section class="block" aria-labelledby="subject-runs">
				<div class="block__head">
					<h2 id="subject-runs">
						{runLbl.nav}
						{#if runs.length > 0}
							<span class="muted">· {runs.length}</span>
						{/if}
					</h2>
					{#if subject.capabilities.can_launch && !editing}
						<div class="actions">
							<mb-button variant="primary" href={`/subjects/${subject.id}/launch`}
								>{launchRunCTA(runLbl)}</mb-button
							>
						</div>
					{/if}
				</div>
				{#if runs.length === 0}
					<p class="muted">
						{runLbl.noneArticle}
						{runLbl.singular} pour ce {subjectLbl.singular.toLowerCase()}.
					</p>
				{:else}
					<mb-table
						columns="2fr 0.8fr 0.9fr 0.7fr 2.75rem"
						sort-key={sortKey}
						sort-direction={sortDirection}
						sort-label="Trier par {name}"
						onmb-sort={onRunsSort}
					>
						<mb-table-row slot="head">
							<mb-table-cell sort-key="titre"><span class="th-label">Titre</span></mb-table-cell>
							<mb-table-cell sort-key="date" align="center"
								><span class="th-label">Date</span></mb-table-cell
							>
							<mb-table-cell sort-key="statut" align="center"
								><span class="th-label">Statut</span></mb-table-cell
							>
							<mb-table-cell sort-key="progression" align="center"
								><span class="th-label">Progression</span></mb-table-cell
							>
							<mb-table-cell actions><span class="sr-only">Actions</span></mb-table-cell>
						</mb-table-row>
						{#each displayedRuns as r (r.id)}
							<mb-table-row>
								<mb-table-cell label="Titre" primary sort-value={r.title}>
									<a href={`/runs/${r.id}`}>{r.title}</a>
								</mb-table-cell>
								<mb-table-cell label="Date" align="center" sort-value={r.created_at}>
									<span class="date">{formatCreatedDate(r.created_at)}</span>
								</mb-table-cell>
								<mb-table-cell label="Statut" align="center" sort-value={r.status}>
									<mb-badge variant={runStatusVariant(r.status)}
										>{formatRunStatus(r.status)}</mb-badge
									>
								</mb-table-cell>
								<mb-table-cell
									label="Progression"
									align="center"
									sort-value={String(r.progress.percent).padStart(3, '0')}
								>
									<span class="pct">{r.progress.percent} %</span>
								</mb-table-cell>
								<mb-table-cell actions>
									<a
										class="row-action"
										href={`/runs/${r.id}`}
										aria-label="Ouvrir"
										title="Ouvrir"
									>
										<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
											><path d="M5 12h14M13 6l6 6-6 6" /></svg
										>
									</a>
								</mb-table-cell>
							</mb-table-row>
						{/each}
					</mb-table>
				{/if}
			</section>

			{#if !editing && showDomainsRead && subject.domains.length === 0 && boot.show_subject_column}
				<details class="advanced advanced--panel">
					<summary>Domaines</summary>
					<p class="muted">Aucun domaine — tous les modèles sont compatibles.</p>
				</details>
			{/if}

			{#if boot.show_collab}
				<section class="block" aria-labelledby="subject-people">
					<h2 id="subject-people">Personnes</h2>
					{#if subject.members.length === 0 && !subject.capabilities.can_manage_members}
						<p class="muted">Aucune personne sur ce {subjectLbl.singular.toLowerCase()}.</p>
					{:else}
						<mb-table
							columns={subject.capabilities.can_manage_members
								? 'minmax(12rem, 1.6fr) 9rem 6.5rem'
								: 'minmax(12rem, 1.6fr) 9rem'}
						>
							<mb-table-row slot="head">
								<mb-table-cell><span class="th-label">Personne</span></mb-table-cell>
								<mb-table-cell><span class="th-label">Rôle</span></mb-table-cell>
								{#if subject.capabilities.can_manage_members}
									<mb-table-cell></mb-table-cell>
								{/if}
							</mb-table-row>
							{#each subject.members as m (m.user_id)}
								<mb-table-row>
									<mb-table-cell label="Personne" primary>
										<span class="person">
											<strong>{m.display_name}</strong>
											<span class="muted">{m.email}</span>
										</span>
									</mb-table-cell>
									<mb-table-cell label="Rôle">
										{#if subject.capabilities.can_manage_members}
											<mb-select
												label="Rôle"
												hide-label
												density="compact"
												value={m.role}
												disabled={memberBusyId === m.user_id}
												onmb-change={(e) =>
													onMemberRoleChange(
														m.user_id,
														m.email,
														e.detail.value as SubjectRole
													)}
											>
												{#each SUBJECT_ROLES as r (r.value)}
													<option value={r.value}>{r.label}</option>
												{/each}
											</mb-select>
										{:else}
											<mb-badge>{formatRole(m.role)}</mb-badge>
										{/if}
									</mb-table-cell>
									{#if subject.capabilities.can_manage_members}
										<mb-table-cell actions>
											<mb-button
												variant="ghost"
												size="sm"
												disabled={memberBusyId === m.user_id}
												onclick={() => onRemoveMember(m.user_id)}
											>
												Retirer
											</mb-button>
										</mb-table-cell>
									{/if}
								</mb-table-row>
							{/each}
							{#if subject.capabilities.can_manage_members}
								<mb-table-row>
									<mb-table-cell label="Personne" primary>
										{#if showEmailFallback}
											<mb-input
												label="Email"
												hide-label
												density="compact"
												type="email"
												placeholder="email@exemple.com"
												value={memberEmail}
												disabled={memberBusyId === 'new'}
												oninput={(e) => (memberEmail = inputValue(e))}
											></mb-input>
										{:else if candidatePeople.length > 0}
											<mb-select
												label="Personne"
												hide-label
												density="compact"
												placeholder="Choisir une personne…"
												value={memberPick}
												disabled={memberBusyId === 'new'}
												onmb-change={(e) => (memberPick = e.detail.value)}
											>
												<option value="">Choisir une personne…</option>
												{#each candidatePeople as p (p.user_id)}
													<option value={p.email}
														>{p.display_name || p.email}</option
													>
												{/each}
											</mb-select>
											{#if selectedCandidate}
												<span class="muted person-email">{selectedCandidate.email}</span>
											{/if}
										{:else}
											<p class="field-hint">
												Plus personne à ajouter dans l’org.{#if boot.can_admin}
													<a href="/admin/members">Inviter</a>{/if}
											</p>
										{/if}
										{#if !showEmailFallback && (candidatePeople.length > 0 || boot.can_admin)}
											<button
												type="button"
												class="linkish"
												disabled={memberBusyId === 'new'}
												onclick={() => {
													showEmailFallback = true;
													memberPick = '';
												}}
											>
												Saisir un email
											</button>
										{:else if showEmailFallback}
											<button
												type="button"
												class="linkish"
												disabled={memberBusyId === 'new'}
												onclick={() => {
													showEmailFallback = false;
													memberEmail = '';
												}}
											>
												Choisir dans la liste
											</button>
										{/if}
									</mb-table-cell>
									<mb-table-cell label="Rôle">
										<mb-select
											label="Rôle"
											hide-label
											density="compact"
											value={memberRole}
											disabled={memberBusyId === 'new'}
											onmb-change={(e) =>
												(memberRole = e.detail.value as SubjectRole)}
										>
											{#each SUBJECT_ROLES as r (r.value)}
												<option value={r.value}>{r.label}</option>
											{/each}
										</mb-select>
									</mb-table-cell>
									<mb-table-cell actions>
										<mb-button
											variant="secondary"
											size="sm"
											loading={memberBusyId === 'new'}
											disabled={memberBusyId === 'new' ||
												(!memberPick && !memberEmail.trim())}
											onclick={onAddMember}
										>
											Ajouter
										</mb-button>
									</mb-table-cell>
								</mb-table-row>
							{/if}
						</mb-table>
					{/if}
				</section>
			{/if}

			{#if subject.capabilities.can_manage && !editing}
				<mb-card>
					<h2 slot="header">Archiver</h2>
					<p class="muted">
						Le {subjectLbl.singular.toLowerCase()} ne sera plus disponible pour de nouvelles
						{runLbl.plural.toLowerCase()}.
					</p>
					<div slot="footer">
						<mb-button variant="danger" onclick={onArchive}>Archiver</mb-button>
					</div>
				</mb-card>
			{/if}
		</div>
	{/if}
</div>
