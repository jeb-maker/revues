<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { requireAdminSession, refreshCsrf } from '$lib/admin/session';
	import { createAdminTeam, listAdminTeams, type AdminTeamListResponse } from '$lib/api/admin';

	let csrf = $state('');
	let teams = $state<AdminTeamListResponse['teams']>([]);
	let name = $state('');
	let slug = $state('');
	let description = $state('');
	let error = $state('');
	let loading = $state(false);
	let ready = $state(false);

	onMount(async () => {
		try {
			const gate = await requireAdminSession();
			if (!gate) return;
			csrf = gate.csrf;
			const data = await listAdminTeams();
			teams = data.teams ?? [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
		} finally {
			ready = true;
		}
	});

	async function onCreate(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const body: { name: string; slug?: string; description?: string } = { name };
			if (slug.trim()) body.slug = slug.trim();
			if (description.trim()) body.description = description.trim();
			const team = await createAdminTeam(body, csrf);
			await goto(`/admin/teams/${team.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Création impossible.';
			try {
				csrf = await refreshCsrf();
			} catch {
				/* ignore */
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Équipes — Revues</title>
</svelte:head>

<main class="admin">
	<p class="brand">Revues</p>
	<h1>Équipes</h1>
	<p class="lede">Créez des équipes et gérez leurs membres (membres de l'organisation uniquement).</p>

	{#if !ready}
		<p class="muted">Chargement…</p>
	{:else}
		{#if error}
			<mb-alert variant="danger" role="alert">{error}</mb-alert>
		{/if}

		<AdminNav section="teams">
			<form class="form" onsubmit={onCreate}>
				<h2>Créer une équipe</h2>
				<label>
					Nom
					<input type="text" bind:value={name} required />
				</label>
				<label>
					Slug
					<input type="text" bind:value={slug} placeholder="Optionnel" />
				</label>
				<label>
					Description
					<textarea bind:value={description} rows="2"></textarea>
				</label>
				<mb-button type="submit" variant="primary" disabled={loading || !csrf}>
					{loading ? 'Création…' : 'Créer'}
				</mb-button>
			</form>

			<section>
				<h2>Équipes ({teams.length})</h2>
				{#if teams.length === 0}
					<p class="muted">Aucune équipe pour cette organisation.</p>
				{:else}
					<ul>
						{#each teams as team (team.id)}
							<li>
								<a href={`/admin/teams/${team.id}`}>{team.name}</a>
								<span class="muted"> · {team.slug} · {team.member_count} membre(s)</span>
							</li>
						{/each}
					</ul>
				{/if}
			</section>
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
		margin-bottom: 1.75rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	input,
	textarea {
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
		gap: 0.65rem;
	}
	a {
		color: #5eead4;
		font-weight: 600;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
