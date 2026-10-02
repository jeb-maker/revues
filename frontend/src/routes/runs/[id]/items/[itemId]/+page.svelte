<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { bootstrap } from '$lib/api/auth';
	import {
		getRunItem,
		updateRunItem,
		type RunItemDetail,
		type UpdateRunItemRequest
	} from '$lib/api/runs';

	let csrf = $state('');
	let detail = $state<RunItemDetail | null>(null);
	let error = $state('');
	let loading = $state(true);
	let saving = $state(false);

	let status = $state<'pending' | 'ok' | 'nok' | 'na'>('pending');
	let comment = $state('');
	let assignedTo = $state<string>('');

	function runId(): number {
		return Number($page.params.id);
	}
	function itemId(): number {
		return Number($page.params.itemId);
	}

	function apply(d: RunItemDetail) {
		detail = d;
		status = d.item.status;
		comment = d.item.comment ?? '';
		assignedTo = d.item.assigned_to != null ? String(d.item.assigned_to) : '';
	}

	async function refresh() {
		apply(await getRunItem(runId(), itemId(), csrf));
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
			const body: UpdateRunItemRequest = {};
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
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>{detail?.item.label ?? 'Point'} — Revues</title>
</svelte:head>

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs">
			<a href="/runs">Revues</a>
			{#if detail}
				· <a href={`/runs/${runId()}`}>{detail.run_title ?? 'Revue'}</a>
			{/if}
		</p>
		{#if detail}
			<h1>{detail.item.label}</h1>
			{#if detail.item.section}
				<p class="lede">{detail.item.section}</p>
			{/if}
		{:else}
			<h1>Point</h1>
		{/if}
	</header>

	{#if error}
		<p class="err" role="alert">{error}</p>
	{/if}

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else if detail}
		{#if detail.item.help_text}
			<p class="help">{detail.item.help_text}</p>
		{/if}

		<form class="form" onsubmit={onSave}>
			<label>
				Statut
				<select bind:value={status} disabled={!detail.capabilities.can_update_items}>
					<option value="pending">En attente</option>
					<option value="ok">OK</option>
					<option value="nok">Non validé</option>
					<option value="na">N/A</option>
				</select>
			</label>

			<label>
				Commentaire {#if status === 'nok'}<em>(obligatoire)</em>{/if}
				<textarea
					bind:value={comment}
					rows="4"
					disabled={!detail.capabilities.can_update_items}
					required={status === 'nok'}
				></textarea>
			</label>

			{#if detail.capabilities.can_assign || detail.item.assigned_to}
				<label>
					Assigné
					<select bind:value={assignedTo} disabled={!detail.capabilities.can_assign}>
						<option value="">— Non assigné —</option>
						{#each detail.assignees ?? [] as a}
							<option value={String(a.user_id)}>{a.display_name} (@{a.login})</option>
						{/each}
					</select>
				</label>
			{/if}

			{#if detail.capabilities.can_update_items || detail.capabilities.can_assign}
				<button type="submit" disabled={saving}>{saving ? 'Enregistrement…' : 'Enregistrer'}</button>
			{:else}
				<p class="muted">Revue non éditable.</p>
			{/if}
		</form>

		{#if detail.events?.length}
			<section class="audit">
				<h2>Historique</h2>
				<ul>
					{#each detail.events as ev}
						<li>
							<span class="when">{ev.created_at}</span>
							<span>
								{#if ev.user_login}@{ev.user_login}{/if}
								{ev.old_status ?? '—'} → {ev.new_status}
							</span>
							{#if ev.comment}
								<span class="cmt">{ev.comment}</span>
							{/if}
						</li>
					{/each}
				</ul>
			</section>
		{/if}
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
		max-width: 40rem;
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
		margin: 0 0 1rem;
		color: #94a3b8;
	}
	.help {
		padding: 0.75rem 1rem;
		border-radius: 0.45rem;
		background: rgba(15, 23, 42, 0.5);
		border: 1px solid #334155;
		color: #cbd5e1;
		margin-bottom: 1rem;
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
	.form em {
		color: #fcd34d;
		font-style: normal;
		font-size: 0.85em;
	}
	.form select,
	.form textarea {
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
	.audit {
		margin-top: 2rem;
	}
	.audit h2 {
		margin: 0 0 0.5rem;
		font-size: 0.95rem;
		color: #99f6e4;
	}
	.audit ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.45rem;
	}
	.audit li {
		padding: 0.55rem 0.75rem;
		border-radius: 0.4rem;
		background: rgba(15, 23, 42, 0.4);
		border: 1px solid #334155;
		font-size: 0.88rem;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}
	.when {
		color: #64748b;
		font-size: 0.8rem;
	}
	.cmt {
		color: #94a3b8;
	}
	.err {
		color: #fca5a5;
	}
	.muted {
		color: #94a3b8;
	}
</style>
