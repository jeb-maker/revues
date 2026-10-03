<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import {
		addAdminTeamMember,
		getAdminTeam,
		removeAdminTeamMember,
		type AdminTeamDetailResponse
	} from '$lib/api/admin';
	import { session } from '$lib/auth/session';

	const csrf = session().csrf_token;

	let detail = $state<AdminTeamDetailResponse | null>(null);
	let selectedUserId = $state<number | null>(null);
	let error = $state('');
	let message = $state('');
	let ready = $state(false);
	let loading = $state(false);

	function teamId(): number {
		return Number(page.params.id);
	}

	async function load() {
		detail = await getAdminTeam(teamId());
		selectedUserId = detail.candidates[0]?.user_id ?? null;
	}

	onMount(async () => {
		const id = teamId();
		if (!Number.isFinite(id) || id <= 0) {
			error = 'Équipe invalide.';
			ready = true;
			return;
		}
		try {
			await load();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
		} finally {
			ready = true;
		}
	});

	async function onAdd(e: Event) {
		e.preventDefault();
		if (selectedUserId == null) {
			error = 'Choisissez un membre.';
			return;
		}
		error = '';
		message = '';
		loading = true;
		try {
			await addAdminTeamMember(teamId(), selectedUserId, csrf);
			await load();
			message = 'Membre ajouté.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Ajout impossible.';
		} finally {
			loading = false;
		}
	}

	async function onRemove(userId: number, name: string) {
		if (!confirm(`Retirer ${name} de l’équipe ?`)) return;
		error = '';
		message = '';
		try {
			await removeAdminTeamMember(teamId(), userId, csrf);
			await load();
			message = 'Membre retiré.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Retrait impossible.';
		}
	}
</script>

<svelte:head>
	<title>{detail?.team.name ?? 'Équipe'} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs">
			<a href="/admin">Administration</a> · <a href="/admin/teams">Équipes</a> ·
			{detail?.team.name ?? 'Équipe'}
		</p>
		<h1>{detail?.team.name ?? 'Équipe'}</h1>
		{#if detail}
			<p class="lede">
				<code>{detail.team.slug}</code>
				{#if detail.team.description}
					— {detail.team.description}
				{/if}
			</p>
		{/if}
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

	<AdminNav section="teams">
		{#if !ready}
			<p class="loading">Chargement…</p>
		{:else if detail}
			<section>
				<h2>Membres ({detail.members.length})</h2>
				{#if detail.members.length === 0}
					<p class="muted">Aucun membre.</p>
				{:else}
					<ul class="row-list">
						{#each detail.members as m (m.user_id)}
							<li>
								<span>{m.display_name || m.login} <span class="muted">({m.email})</span></span>
								<mb-button
									type="button"
									variant="ghost"
									size="sm"
									onclick={() => onRemove(m.user_id, m.display_name || m.login)}
								>
									Retirer
								</mb-button>
							</li>
						{/each}
					</ul>
				{/if}
			</section>

			<form class="stack-form section" onsubmit={onAdd}>
				<h2>Ajouter un membre</h2>
				{#if detail.candidates.length === 0}
					<p class="muted">Tous les membres de l’organisation sont déjà dans l’équipe.</p>
				{:else}
					<mb-select
						label="Membre"
						name="user_id"
						required
						value={selectedUserId == null ? '' : String(selectedUserId)}
						onmb-change={(e) => (selectedUserId = e.detail.value ? Number(e.detail.value) : null)}
					>
						{#each detail.candidates as c (c.user_id)}
							<option value={String(c.user_id)}>{c.display_name || c.login} ({c.email})</option>
						{/each}
					</mb-select>
					<mb-button type="submit" variant="primary" disabled={loading}>
						{loading ? 'Ajout…' : 'Ajouter'}
					</mb-button>
				{/if}
			</form>
		{/if}
	</AdminNav>
</div>
