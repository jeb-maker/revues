<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import {
		getRunItem,
		updateRunItem,
		UpdateRunItemError,
		type RunItemDetail,
		type UpdateRunItemRequest
	} from '$lib/api/runs';
	import {
		attachmentDownloadURL,
		uploadRunItemAttachment,
		type Attachment
	} from '$lib/api/attachments';
	import {
		getRunItemJira,
		postRunItemJiraCreate,
		putRunItemJiraLink,
		type JiraLink,
		type RunItemJira
	} from '$lib/api/jira';
	import { session } from '$lib/auth/session';
	import { formatItemStatus, type ItemStatus } from '$lib/i18n/labels';
	import { inputValue } from '$lib/mb';

	const ITEM_STATUSES: ItemStatus[] = ['pending', 'ok', 'nok', 'na'];
	const csrf = session().csrf_token;

	let detail = $state<RunItemDetail | null>(null);
	let error = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let uploading = $state(false);
	let uploadError = $state('');

	let status = $state<ItemStatus>('pending');
	let comment = $state('');
	let assignedTo = $state<string>('');
	let attachment = $state<Attachment | null>(null);

	let jira = $state<RunItemJira | null>(null);
	let jiraLink = $state<JiraLink | null>(null);
	let jiraIssue = $state('');
	let jiraTitle = $state('');
	let jiraDescription = $state('');
	let jiraBusy = $state(false);
	let jiraError = $state('');
	let jiraMessage = $state('');

	function runId(): number {
		return Number(page.params.id);
	}
	function itemId(): number {
		return Number(page.params.itemId);
	}

	function apply(d: RunItemDetail) {
		detail = d;
		status = d.item.status;
		comment = d.item.comment ?? '';
		assignedTo = d.item.assigned_to != null ? String(d.item.assigned_to) : '';
		attachment = d.attachment ?? null;
		jiraLink = d.jira_link ?? null;
	}

	async function refreshJira() {
		const state = await getRunItemJira(runId(), itemId(), csrf);
		jira = state;
		jiraLink = state.link ?? null;
		if (state.default_title) jiraTitle = state.default_title;
		if (state.default_description) jiraDescription = state.default_description;
	}

	onMount(async () => {
		try {
			apply(await getRunItem(runId(), itemId(), csrf));
			await refreshJira();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Point introuvable.';
		} finally {
			loading = false;
		}
	});

	async function onSave(e: Event) {
		e.preventDefault();
		if (!detail) return;
		saving = true;
		error = '';
		try {
			// Verrou optimiste : on renvoie l'updated_at affiché ; périmé → 409.
			const body: UpdateRunItemRequest = { updated_at: detail.item.updated_at };
			if (detail.capabilities.can_update_items) {
				body.status = status;
				body.comment = comment;
			}
			if (detail.capabilities.can_assign) {
				if (assignedTo === '') {
					body.unassign = true;
				} else {
					body.assigned_to = Number(assignedTo);
				}
			}
			apply(await updateRunItem(runId(), itemId(), body, csrf));
			await refreshJira();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
			if (err instanceof UpdateRunItemError && err.status === 409) {
				// Conflit : recharger le point pour afficher l'état courant (le message reste visible).
				try {
					apply(await getRunItem(runId(), itemId(), csrf));
				} catch {
					/* le message de conflit reste affiché */
				}
			}
		} finally {
			saving = false;
		}
	}

	async function onUpload(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file || !detail?.capabilities.can_update_items) return;
		uploading = true;
		uploadError = '';
		try {
			attachment = await uploadRunItemAttachment(runId(), itemId(), file, csrf);
		} catch (err) {
			uploadError = err instanceof Error ? err.message : 'Upload impossible.';
		} finally {
			uploading = false;
			input.value = '';
		}
	}

	async function onLinkJira(e: Event) {
		e.preventDefault();
		jiraBusy = true;
		jiraError = '';
		jiraMessage = '';
		try {
			jiraLink = await putRunItemJiraLink(runId(), itemId(), { issue: jiraIssue }, csrf);
			jiraIssue = '';
			jiraMessage = 'Lien Jira enregistré.';
			await refreshJira();
		} catch (err) {
			jiraError = err instanceof Error ? err.message : 'Liaison impossible.';
		} finally {
			jiraBusy = false;
		}
	}

	async function onCreateJira(e: Event) {
		e.preventDefault();
		jiraBusy = true;
		jiraError = '';
		jiraMessage = '';
		try {
			jiraLink = await postRunItemJiraCreate(
				runId(),
				itemId(),
				{ title: jiraTitle, description: jiraDescription },
				csrf
			);
			jiraMessage = 'Ticket Jira créé.';
			await refreshJira();
		} catch (err) {
			jiraError = err instanceof Error ? err.message : 'Création impossible.';
		} finally {
			jiraBusy = false;
		}
	}
</script>

<svelte:head>
	<title>{detail?.item.label ?? 'Point'} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs">
			<a href="/runs">Revues</a>
			{#if detail}
				· <a href={`/runs/${runId()}`}>{detail.run_title ?? 'Revue'}</a>
			{/if}
		</p>
		<h1>{detail?.item.label ?? 'Point'}</h1>
		{#if detail?.item.section}
			<p class="lede">{detail.item.section}</p>
		{/if}
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if detail}
		{#if detail.item.help_text}
			<p class="callout">{detail.item.help_text}</p>
		{/if}

		<div class="card-stack">
			<form onsubmit={onSave}>
				<mb-card>
					<h2 slot="header">Saisie</h2>
					<div class="stack-form">
						<mb-select
							label="Statut"
							required
							value={status}
							disabled={!detail.capabilities.can_update_items}
							onmb-change={(e) => (status = e.detail.value as ItemStatus)}
						>
							{#each ITEM_STATUSES as s (s)}
								<option value={s}>{formatItemStatus(s)}</option>
							{/each}
						</mb-select>

						<mb-textarea
							label={status === 'nok' ? 'Commentaire (obligatoire)' : 'Commentaire'}
							rows="4"
							value={comment}
							disabled={!detail.capabilities.can_update_items}
							required={status === 'nok'}
							oninput={(e) => (comment = inputValue(e))}
						></mb-textarea>

						{#if detail.capabilities.can_assign || detail.item.assigned_to}
							<mb-select
								label="Assigné à"
								placeholder="— Non assigné —"
								value={assignedTo}
								disabled={!detail.capabilities.can_assign}
								onmb-change={(e) => (assignedTo = e.detail.value)}
							>
								{#each detail.assignees ?? [] as a (a.user_id)}
									<option value={String(a.user_id)}>{a.display_name} (@{a.login})</option>
								{/each}
							</mb-select>
						{/if}
					</div>
					{#if detail.capabilities.can_update_items || detail.capabilities.can_assign}
						<div slot="footer">
							<mb-button type="submit" variant="primary" loading={saving}>
								{saving ? 'Enregistrement…' : 'Enregistrer'}
							</mb-button>
						</div>
					{:else}
						<p slot="footer" class="muted">Revue non éditable.</p>
					{/if}
				</mb-card>
			</form>

			<mb-card>
				<h2 slot="header">Pièce jointe</h2>
				{#if attachment}
					<p class="attachment">
						{#if attachment.is_image}
							<img
								class="preview"
								src={attachmentDownloadURL(runId(), itemId(), attachment.id)}
								alt={attachment.filename}
							/>
						{/if}
						<a href={attachmentDownloadURL(runId(), itemId(), attachment.id)} download={attachment.filename}>
							{attachment.filename}
						</a>
						<span class="muted">({Math.round(attachment.size_bytes / 1024)} Ko)</span>
					</p>
				{:else}
					<p class="muted">Aucune pièce jointe.</p>
				{/if}
				{#if detail.capabilities.can_update_items}
					<label class="file">
						<span>{uploading ? 'Envoi…' : 'Ajouter ou remplacer'}</span>
						<input
							type="file"
							accept=".jpg,.jpeg,.png,.webp,.pdf,image/*,application/pdf"
							onchange={onUpload}
							disabled={uploading}
						/>
						<span class="field-hint">JPEG, PNG, WebP ou PDF · 5 Mo max.</span>
					</label>
				{/if}
				{#if uploadError}
					<mb-alert variant="danger">{uploadError}</mb-alert>
				{/if}
			</mb-card>

			<mb-card>
				<h2 slot="header">Issue Jira</h2>
				{#if jiraLink}
					<p>
						Liée à :
						<a href={jiraLink.external_url} target="_blank" rel="noopener noreferrer">{jiraLink.external_key}</a>
					</p>
				{:else}
					<p class="muted">Aucune issue Jira liée.</p>
				{/if}

				{#if jiraError}
					<mb-alert variant="danger">{jiraError}</mb-alert>
				{/if}
				{#if jiraMessage}
					<mb-alert variant="success">{jiraMessage}</mb-alert>
				{/if}

				{#if jira?.can_link && jira.configured}
					<form class="stack-form" onsubmit={onLinkJira}>
						<mb-input
							label="Clé ou URL Jira"
							hint="Ex. PROJ-123 ou https://…/browse/PROJ-123"
							required
							autocomplete="off"
							value={jiraIssue}
							oninput={(e) => (jiraIssue = inputValue(e))}
						></mb-input>
						<mb-button type="submit" variant="secondary" loading={jiraBusy}>
							{jiraLink ? 'Mettre à jour le lien' : "Lier l'issue"}
						</mb-button>
					</form>

					{#if jira.can_create && !jiraLink}
						<form class="stack-form" onsubmit={onCreateJira}>
							<mb-input
								label="Titre du ticket"
								required
								value={jiraTitle}
								oninput={(e) => (jiraTitle = inputValue(e))}
							></mb-input>
							<mb-textarea
								label="Description"
								rows="5"
								required
								value={jiraDescription}
								oninput={(e) => (jiraDescription = inputValue(e))}
							></mb-textarea>
							<mb-button type="submit" variant="secondary" loading={jiraBusy}>Créer le ticket Jira</mb-button>
						</form>
					{/if}
				{:else if jira?.can_link && !jira.configured}
					<p class="muted">Jira n'est pas configuré — contactez un administrateur.</p>
				{/if}
			</mb-card>

			{#if detail.events?.length}
				<mb-card>
					<h2 slot="header">Historique</h2>
					<ul class="row-list">
						{#each detail.events as ev (ev.id)}
							<li class="event">
								<span class="muted">{ev.created_at}</span>
								<span>
									{#if ev.user_login}@{ev.user_login}{/if}
									{ev.old_status ? formatItemStatus(ev.old_status) : '—'} → {formatItemStatus(ev.new_status)}
								</span>
								{#if ev.comment}
									<span class="muted">{ev.comment}</span>
								{/if}
							</li>
						{/each}
					</ul>
				</mb-card>
			{/if}
		</div>
	{/if}
</div>

<style>
	.attachment {
		display: flex;
		flex-direction: column;
		gap: var(--mb-space-2);
		align-items: flex-start;
	}
	.preview {
		max-width: 100%;
		max-height: 220px;
		border: 1px solid var(--mb-color-border);
		border-radius: var(--mb-radius-sm);
	}
	.file {
		display: flex;
		flex-direction: column;
		gap: var(--mb-space-1);
		font-size: var(--mb-font-size-sm);
	}
	.event {
		flex-direction: column;
		align-items: flex-start;
		gap: var(--mb-space-1);
	}
	.row-list {
		margin: 0;
	}
</style>
