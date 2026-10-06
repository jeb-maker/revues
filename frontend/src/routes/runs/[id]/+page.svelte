<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/state';
	import {
		completeRun,
		getRun,
		updateRunItem,
		UpdateRunItemError,
		type RunDetail,
		type RunItem
	} from '$lib/api/runs';
	import { session } from '$lib/auth/session';
	import { formatItemStatus, formatRunStatus, itemStatusVariant } from '$lib/i18n/labels';
	import { runLabels } from '$lib/i18n/uiLabels';
	import { inputValue } from '$lib/mb';

	type ItemStatus = 'pending' | 'ok' | 'nok' | 'na';
	type ItemFilter = 'all' | 'pending' | 'nok' | 'required';

	type ItemDraft = {
		status: ItemStatus;
		comment: string;
		updated_at: string;
		dirty: boolean;
		error: string;
	};

	const ITEM_STATUSES: { value: ItemStatus; label: string }[] = [
		{ value: 'pending', label: 'En attente' },
		{ value: 'ok', label: 'Validé' },
		{ value: 'nok', label: 'Non validé' },
		{ value: 'na', label: 'Non applicable' }
	];

	const ITEM_FILTERS: { value: ItemFilter; label: string }[] = [
		{ value: 'all', label: 'Tous' },
		{ value: 'pending', label: 'En attente' },
		{ value: 'nok', label: 'Non validés' },
		{ value: 'required', label: 'Obligatoires' }
	];

	const boot = session();
	const csrf = boot.csrf_token;
	const runLbl = $derived(runLabels(boot.organization?.ui_run_label));

	let run = $state<RunDetail | null>(null);
	let drafts = $state<Record<number, ItemDraft>>({});
	let error = $state('');
	let loading = $state(true);
	let closing = $state(false);
	let closingNote = $state('');
	let savingId = $state<number | null>(null);
	let itemFilter = $state<ItemFilter>('all');
	let assignDrafts = $state<Record<number, string>>({});

	function runId(): number {
		return Number(page.params.id);
	}

	function formatWhen(iso: string | null | undefined): string {
		if (!iso) return '';
		const d = new Date(iso);
		if (Number.isNaN(d.getTime())) return iso;
		return d.toLocaleString('fr-FR', { dateStyle: 'medium', timeStyle: 'short' });
	}

	function formatDay(iso: string | null | undefined): string {
		if (!iso) return '';
		const d = new Date(iso);
		if (Number.isNaN(d.getTime())) return iso;
		return d.toLocaleDateString('fr-FR', { dateStyle: 'medium' });
	}

	function draftFromItem(item: RunItem): ItemDraft {
		return {
			status: item.status as ItemStatus,
			comment: item.comment ?? '',
			updated_at: item.updated_at,
			dirty: false,
			error: ''
		};
	}

	function syncDrafts(items: RunItem[]) {
		const next = { ...drafts };
		const nextAssign = { ...assignDrafts };
		for (const item of items) {
			const cur = next[item.id];
			if (!cur || !cur.dirty) {
				next[item.id] = draftFromItem(item);
			} else {
				next[item.id] = { ...cur, updated_at: item.updated_at };
			}
			if (savingId !== item.id) {
				nextAssign[item.id] = item.assigned_to != null ? String(item.assigned_to) : '';
			}
		}
		drafts = next;
		assignDrafts = nextAssign;
	}

	function setDraft(id: number, patch: Partial<ItemDraft>) {
		const cur = drafts[id];
		if (!cur) return;
		drafts = { ...drafts, [id]: { ...cur, ...patch } };
	}

	function recomputeProgress(items: RunItem[]): RunDetail['progress'] {
		const total = items.length;
		const done = items.filter((i) => i.status !== 'pending').length;
		return {
			total,
			done,
			percent: total === 0 ? 0 : Math.round((done * 100) / total)
		};
	}

	function pendingRequired(items: RunItem[]): number {
		return items.filter((i) => i.required && i.status === 'pending').length;
	}

	function applyItem(item: RunItem) {
		if (!run) return;
		const items = run.items.map((i) => (i.id === item.id ? item : i));
		run = {
			...run,
			items,
			progress: recomputeProgress(items),
			pending_required_count: pendingRequired(items)
		};
		setDraft(item.id, {
			status: item.status as ItemStatus,
			comment: item.comment ?? '',
			updated_at: item.updated_at,
			dirty: false,
			error: ''
		});
		assignDrafts = {
			...assignDrafts,
			[item.id]: item.assigned_to != null ? String(item.assigned_to) : ''
		};
	}

	function matchesFilter(item: RunItem): boolean {
		switch (itemFilter) {
			case 'pending':
				return (drafts[item.id]?.status ?? item.status) === 'pending';
			case 'nok':
				return (drafts[item.id]?.status ?? item.status) === 'nok';
			case 'required':
				return item.required;
			default:
				return true;
		}
	}

	onMount(async () => {
		try {
			run = await getRun(runId(), csrf);
			closingNote = run.closing_note ?? '';
			syncDrafts(run.items);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Revue introuvable.';
		} finally {
			loading = false;
		}
	});

	async function onComplete(e: Event) {
		e.preventDefault();
		if (!run || !confirm('Clôturer cette revue ?')) return;
		closing = true;
		error = '';
		try {
			run = await completeRun(run.id, { closing_note: closingNote }, csrf);
			syncDrafts(run.items);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Clôture impossible.';
		} finally {
			closing = false;
		}
	}

	async function saveItem(item: RunItem) {
		if (!run) return;
		const d = drafts[item.id] ?? draftFromItem(item);
		const canUpdate = run.capabilities.can_update_items;
		const canAssign = session().show_assign && run.capabilities.can_assign;
		if (!canUpdate && !canAssign) return;

		if (canUpdate && d.status === 'nok' && !d.comment.trim()) {
			setDraft(item.id, {
				error: 'Commentaire obligatoire pour le statut Non validé.',
				dirty: true
			});
			return;
		}

		const assignVal = assignDrafts[item.id] ?? '';
		const prevAssign = item.assigned_to != null ? String(item.assigned_to) : '';
		const assignChanged = canAssign && assignVal !== prevAssign;

		const unchanged =
			d.status === item.status && d.comment.trim() === (item.comment ?? '').trim();
		if (unchanged && !assignChanged && !d.dirty) return;

		savingId = item.id;
		error = '';
		setDraft(item.id, { error: '' });
		try {
			const body: Parameters<typeof updateRunItem>[2] = {
				updated_at: d.updated_at
			};
			if (canUpdate) {
				body.status = d.status;
				body.comment = d.comment;
			}
			if (assignChanged) {
				if (assignVal === '') body.unassign = true;
				else body.assigned_to = Number(assignVal);
			}
			const updated = await updateRunItem(run.id, item.id, body, csrf);
			applyItem(updated.item);
		} catch (err) {
			if (err instanceof UpdateRunItemError && err.status === 409) {
				error = 'Ce point a été modifié ailleurs — rechargement.';
				run = await getRun(run.id, csrf);
				syncDrafts(run.items);
			} else {
				const msg = err instanceof Error ? err.message : 'Mise à jour impossible.';
				setDraft(item.id, { error: msg });
			}
		} finally {
			savingId = null;
		}
	}

	function onStatusChange(item: RunItem, status: ItemStatus) {
		const d = drafts[item.id];
		if (!d || status === d.status) return;
		setDraft(item.id, {
			status,
			dirty: true,
			error:
				status === 'nok' && !d.comment.trim()
					? 'Commentaire obligatoire pour le statut Non validé.'
					: ''
		});
		if (status === 'nok' && !d.comment.trim()) return;
		void saveItem(item);
	}

	async function onStatusKeydown(item: RunItem, e: KeyboardEvent) {
		const keys = ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Home', 'End'];
		if (!keys.includes(e.key)) return;
		e.preventDefault();
		const d = drafts[item.id] ?? draftFromItem(item);
		const idx = ITEM_STATUSES.findIndex((s) => s.value === d.status);
		const cur = idx < 0 ? 0 : idx;
		let next = cur;
		if (e.key === 'Home') next = 0;
		else if (e.key === 'End') next = ITEM_STATUSES.length - 1;
		else if (e.key === 'ArrowUp' || e.key === 'ArrowLeft') {
			next = cur <= 0 ? ITEM_STATUSES.length - 1 : cur - 1;
		} else {
			next = cur >= ITEM_STATUSES.length - 1 ? 0 : cur + 1;
		}
		const status = ITEM_STATUSES[next].value;
		onStatusChange(item, status);
		await tick();
		document.getElementById(`status-${item.id}-${status}`)?.focus();
	}

	function onCommentInput(item: RunItem, value: string) {
		const d = drafts[item.id];
		if (!d) return;
		setDraft(item.id, {
			comment: value,
			dirty: true,
			error:
				d.status === 'nok' && !value.trim()
					? 'Commentaire obligatoire pour le statut Non validé.'
					: ''
		});
	}

	function onCommentCommit(item: RunItem) {
		const d = drafts[item.id];
		if (!d?.dirty) return;
		void saveItem(item);
	}

	function onAssignChange(item: RunItem, value: string) {
		assignDrafts = { ...assignDrafts, [item.id]: value };
		void saveItem(item);
	}

	function groupBySection(items: RunItem[]): { section: string; items: RunItem[] }[] {
		const order: string[] = [];
		const map = new Map<string, RunItem[]>();
		for (const item of items) {
			if (!matchesFilter(item)) continue;
			const key = item.section || 'Général';
			if (!map.has(key)) {
				map.set(key, []);
				order.push(key);
			}
			map.get(key)!.push(item);
		}
		return order.map((section) => ({ section, items: map.get(section)! }));
	}

	const filteredGroups = $derived(run ? groupBySection(run.items) : []);
	const filteredCount = $derived(filteredGroups.reduce((n, g) => n + g.items.length, 0));
	const canEdit = $derived(!!run?.capabilities.can_update_items);
	const showAssign = $derived(
		session().show_assign &&
			!!run &&
			(run.capabilities.can_assign || run.items.some((i) => i.assigned_to != null)) &&
			(run.assignees?.length ?? 0) > 0
	);
	const bodyColumns = $derived(
		showAssign
			? 'minmax(12rem, 1.4fr) 8.5rem minmax(12rem, 1.2fr) 9rem'
			: 'minmax(12rem, 1.4fr) 8.5rem minmax(12rem, 1.2fr)'
	);
</script>

<svelte:head>
	<title>{run?.title ?? runLbl.singular} — Revues</title>
</svelte:head>

<div class="page page--wide run-detail">
	<header class="page-header">
		<p class="crumbs">
			<a href="/runs">{runLbl.nav}</a>
			{#if run}
				· <a href={`/subjects/${run.subject_id}`}>{run.subject_name}</a>
			{/if}
		</p>
		<h1>{run?.template_name ?? runLbl.singular}</h1>
		{#if run}
			<p class="lede">
				{formatDay(run.created_at)}
				· v{run.template_version}
				· <span class="progress-inline">{run.progress.done}/{run.progress.total} · {run.progress.percent} %</span>
				{#if run.status !== 'in_progress'}
					·
					<mb-badge variant={run.status === 'done' ? 'success' : 'info'}
						>{formatRunStatus(run.status)}</mb-badge
					>
				{/if}
				{#if run.due_date}
					· échéance {run.due_date}
				{/if}
			</p>
			<div
				class="progress-quiet"
				role="progressbar"
				aria-valuemin="0"
				aria-valuemax="100"
				aria-valuenow={run.progress.percent}
				aria-label={`Progression ${run.progress.percent} %`}
			>
				<span style={`inline-size: ${run.progress.percent}%`}></span>
			</div>
		{/if}
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if run}
		<div class="card-stack">
			{#if run.status === 'done'}
				<mb-card>
					<h2 slot="header">Attestation de clôture</h2>
					<p>
						Clôturée
						{#if run.completed_by_login}
							par <strong>@{run.completed_by_login}</strong>
						{/if}
						{#if run.completed_at}
							le <strong>{formatWhen(run.completed_at)}</strong>
						{/if}.
					</p>
					{#if run.closing_note}
						<p class="muted">{run.closing_note}</p>
					{/if}
				</mb-card>
			{/if}

			{#if (run.pending_required_count ?? 0) > 0}
				<p class="field-error">{run.pending_required_count} point(s) obligatoire(s) en attente</p>
			{/if}

			{#if run.items.length > 4}
				<div class="item-filters">
					<mb-segmented-control label="Filtrer les points">
						{#each ITEM_FILTERS as f (f.value)}
							<button
								type="button"
								class:is-active={itemFilter === f.value}
								aria-pressed={itemFilter === f.value}
								onclick={() => (itemFilter = f.value)}
							>
								{f.label}
							</button>
						{/each}
					</mb-segmented-control>
				</div>
			{/if}

			{#if filteredCount === 0}
				<p class="muted empty-filter">Aucun point pour ce filtre.</p>
			{:else}
				{#each filteredGroups as group (group.section)}
					{@const showSection = filteredGroups.length > 1 || group.section !== 'Général'}
					<section class="items" aria-label={showSection ? undefined : 'Points'}>
						{#if showSection}
							<h2 class="items__section">{group.section}</h2>
						{/if}
						<ul class="item-list">
							{#each group.items as item (item.id)}
								{@const d = drafts[item.id] ?? draftFromItem(item)}
								{@const commentErrId = `comment-err-${item.id}`}
								<li class="item" class:item--busy={savingId === item.id}>
									<div class="item-title">
										<h3 class="label">
											{item.label}
											{#if item.required}
												<abbr title="Obligatoire" aria-label="obligatoire">*</abbr>
											{/if}
										</h3>
										{#if savingId === item.id}
											<span class="muted saving" aria-live="polite">Enregistrement…</span>
										{/if}
										<a
											class="row-action"
											href={`/runs/${run.id}/items/${item.id}`}
											aria-label={`Détails — ${item.label}`}
											title="Détails"
										>
											<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
												><path d="M5 12h14M13 6l6 6-6 6" /></svg
											>
										</a>
									</div>
									<div class="item-body" style={`--item-body-cols: ${bodyColumns}`}>
										{#if item.help_text?.trim()}
											<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
											<div
												class="item-help-col"
												role="region"
												tabindex="0"
												aria-label={`Explication — ${item.label}`}
											>
												<p class="item-help">{item.help_text.trim()}</p>
											</div>
										{:else}
											<div class="item-help-col">
												<span class="muted">—</span>
											</div>
										{/if}
										<div class="item-status">
											{#if canEdit}
												<div
													class="status-stack"
													role="radiogroup"
													tabindex="-1"
													aria-label={`Statut — ${item.label}`}
													onkeydown={(e) => onStatusKeydown(item, e)}
												>
													{#each ITEM_STATUSES as s (s.value)}
														<button
															type="button"
															id={`status-${item.id}-${s.value}`}
															role="radio"
															class:is-active={d.status === s.value}
															aria-checked={d.status === s.value}
															tabindex={d.status === s.value ? 0 : -1}
															disabled={savingId === item.id}
															onclick={() => onStatusChange(item, s.value)}
														>
															{s.label}
														</button>
													{/each}
												</div>
											{:else}
												<mb-badge variant={itemStatusVariant(item.status)}
													>{formatItemStatus(item.status)}</mb-badge
												>
											{/if}
										</div>
										<div class="item-comment">
											{#if canEdit}
												<textarea
													id={`comment-${item.id}`}
													class="item-comment__field"
													class:item-comment__field--error={!!d.error}
													rows="1"
													aria-label={d.status === 'nok'
														? `Commentaire obligatoire — ${item.label}`
														: `Commentaire — ${item.label}`}
													aria-invalid={d.error ? 'true' : undefined}
													aria-describedby={d.error ? commentErrId : undefined}
													placeholder="Commentaire"
													value={d.comment}
													disabled={savingId === item.id}
													required={d.status === 'nok'}
													oninput={(e) =>
														onCommentInput(
															item,
															(e.currentTarget as HTMLTextAreaElement).value
														)}
													onchange={() => onCommentCommit(item)}
												></textarea>
												{#if d.error}
													<p
														id={commentErrId}
														class="field-error item-comment__error"
														role="alert"
													>
														{d.error}
													</p>
												{/if}
											{:else if item.comment?.trim()}
												<span class="muted comment-ro">{item.comment.trim()}</span>
											{:else}
												<span class="muted">—</span>
											{/if}
										</div>
										{#if showAssign}
											<div class="item-assign">
												<span class="sr-only">Assigné</span>
												{#if run.capabilities.can_assign}
													<mb-select
														label="Assigné à"
														hide-label
														density="compact"
														placeholder="—"
														value={assignDrafts[item.id] ?? ''}
														disabled={savingId === item.id}
														onmb-change={(e) => onAssignChange(item, e.detail.value)}
													>
														<option value="">—</option>
														{#each run.assignees ?? [] as a (a.user_id)}
															<option value={String(a.user_id)}>@{a.login}</option>
														{/each}
													</mb-select>
												{:else if item.assigned_login}
													<span class="muted">@{item.assigned_login}</span>
												{:else}
													<span class="muted">—</span>
												{/if}
											</div>
										{/if}
									</div>
								</li>
							{/each}
						</ul>
					</section>
				{/each}
			{/if}

			{#if run.capabilities.can_complete}
				<form onsubmit={onComplete}>
					<mb-card>
						<h2 slot="header">Clôture</h2>
						{#if (run.pending_required_count ?? 0) > 0}
							<p class="field-error">
								Encore {run.pending_required_count} point(s) obligatoire(s) en attente avant
								clôture.
							</p>
						{:else if run.status === 'in_progress'}
							<p class="field-hint">
								Prêt à clôturer — {run.progress.done}/{run.progress.total} points traités.
							</p>
						{/if}
						<div class="stack-form">
							<mb-textarea
								label="Note de clôture"
								rows="3"
								value={closingNote}
								oninput={(e) => (closingNote = inputValue(e))}
							></mb-textarea>
						</div>
						<div slot="footer">
							<mb-button type="submit" variant="primary" loading={closing}>
								{closing ? 'Clôture…' : `Clôturer ${runLbl.article} ${runLbl.singular}`}
							</mb-button>
						</div>
					</mb-card>
				</form>
			{/if}
		</div>
	{/if}
</div>
