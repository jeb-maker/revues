<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { requireAdminSession, refreshCsrf } from '$lib/admin/session';
	import {
		addAdminTeamMember,
		getAdminTeam,
		removeAdminTeamMember,
		type AdminTeamDetailResponse
	} from '$lib/api/admin';

	let csrf = $state('');
	let detail = $state<AdminTeamDetailResponse | null>(null);
	let selectedUserId = $state<number | null>(null);
	let error = $state('');
	let message = $state('');
	let ready = $state(false);
	let loading = $state(false);

	const teamId = $derived(Number($page.params.id));

	async function load(id: number) {
		detail = await getAdminTeam(id);
		selectedUserId = detail.candidates[0]?.user_id ?? null;
	}

	onMount(async () => {
		try {
			const gate = await requireAdminSession();
			if (!gate) return;
			csrf = gate.csrf;
			if (!Number.isFinite(teamId) || teamId <= 0) {
				error = 'Équipe invalide.';
				return;
			}
			await load(teamId);
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
			await addAdminTeamMember(teamId, selectedUserId, csrf);
			csrf = await refreshCsrf();
			await load(teamId);
			message = 'Membre ajouté.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Ajout impossible.';
			try {
				csrf = await refreshCsrf();
			} catch {
				/* ignore */
			}
		} finally {
			loading = false;
		}
	}

	async function onRemove(userId: number) {
		error = '';
		message = '';
		try {
			await removeAdminTeamMember(teamId, userId, csrf);
			csrf = await refreshCsrf();
			await load(teamId);
			message = 'Membre retiré.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Retrait impossible.';
			try {
				csrf = await refreshCsrf();
			} catch {
				/* ignore */
			}
		}
	}
</script>

<svelte:head>
	<title>{detail?.team.name ?? 'Équipe'} — Revues</title>
</svelte:head>

<main class="admin">
	<p class="brand">Revues</p>
	<h1>{detail?.team.name ?? 'Équipe'}</h1>
	<p class="lede">
		{#if detail}
			<code>{detail.team.slug}</code>
			{#if detail.team.description}
				— {detail.team.description}
			{/if}
		{:else}
			Détail de l'équipe.
		{/if}
	</p>

	{#if !ready}
		<p class="muted">Chargement…</p>
	{:else}
		{#if error}
			<mb-alert variant="danger" role="alert">{error}</mb-alert>
		{/if}
		{#if message}
			<mb-alert variant="success" role="status">{message}</mb-alert>
		{/if}

		<AdminNav section="teams">
			<p><a href="/admin/teams">← Toutes les équipes</a></p>

			{#if detail}
				<section>
					<h2>Membres ({detail.members.length})</h2>
					{#if detail.members.length === 0}
						<p class="muted">Aucun membre.</p>
					{:else}
						<ul>
							{#each detail.members as m (m.user_id)}
								<li>
									<span>{m.display_name || m.login} <span class="muted">({m.email})</span></span>
									<mb-button type="button" variant="ghost" onclick={() => onRemove(m.user_id)}>
										Retirer
									</mb-button>
								</li>
							{/each}
						</ul>
					{/if}
				</section>

				<form class="form" onsubmit={onAdd}>
					<h2>Ajouter un membre</h2>
					{#if detail.candidates.length === 0}
						<p class="muted">Tous les membres de l'organisation sont déjà dans l'équipe.</p>
					{:else}
						<label>
							Membre
							<select
								value={selectedUserId ?? ''}
								onchange={(e) => {
									const v = (e.currentTarget as HTMLSelectElement).value;
									selectedUserId = v ? Number(v) : null;
								}}
								required
							>
								{#each detail.candidates as c (c.user_id)}
									<option value={c.user_id}>{c.display_name || c.login} ({c.email})</option>
								{/each}
							</select>
						</label>
						<mb-button type="submit" variant="primary" disabled={loading || !csrf}>
							{loading ? 'Ajout…' : 'Ajouter'}
						</mb-button>
					{/if}
				</form>
			{/if}
		</AdminNav>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background:
			radial-gradient(ellipse 80% 50% at 10% 0%, rgba(15, 118, 110, 0.28), transparent 55%),
			linear-gradient(165deg, #0f172a 0%, #1e293b 55%, #134e4a 100%);
		color: #f8fafc;
	}
	.admin {
		max-width: 42rem;
		margin: 0 auto;
		padding: 6vh 1.25rem 3rem;
	}
	.brand {
		margin: 0 0 0.35rem;
		font-size: clamp(2rem, 6vw, 2.75rem);
		font-weight: 700;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	h1 {
		margin: 0 0 0.45rem;
		font-size: 1.15rem;
		font-weight: 500;
		color: #99f6e4;
	}
	h2 {
		margin: 0 0 0.75rem;
		font-size: 1rem;
	}
	.lede {
		margin: 0 0 1.25rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
		margin-top: 1.5rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	select {
		padding: 0.55rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid rgba(148, 163, 184, 0.45);
		background: rgba(15, 23, 42, 0.55);
		color: #f8fafc;
	}
	ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.55rem;
	}
	li {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
	}
	a {
		color: #5eead4;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
	code {
		color: #99f6e4;
	}
</style>
