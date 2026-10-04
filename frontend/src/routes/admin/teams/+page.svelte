<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { createAdminTeam, listAdminTeams, type AdminTeamListResponse } from '$lib/api/admin';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';

	const csrf = session().csrf_token;

	let teams = $state<AdminTeamListResponse['teams']>([]);
	let name = $state('');
	let slug = $state('');
	let description = $state('');
	let error = $state('');
	let loading = $state(false);
	let ready = $state(false);

	onMount(async () => {
		try {
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
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Équipes — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs"><a href="/admin">Administration</a> · Équipes</p>
		<h1>Équipes</h1>
		<p class="lede">
			Créez des équipes et gérez leurs membres (membres de l’organisation uniquement).
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	<AdminNav section="hub">
		<form class="stack-form" onsubmit={onCreate}>
			<h2>Créer une équipe</h2>
			<mb-input
				label="Nom"
				type="text"
				required
				value={name}
				oninput={(e) => (name = inputValue(e))}
			></mb-input>
			<mb-input
				label="Identifiant"
				hint="Optionnel — généré depuis le nom si vide."
				type="text"
				value={slug}
				oninput={(e) => (slug = inputValue(e))}
			></mb-input>
			<mb-textarea
				label="Description"
				rows="2"
				value={description}
				oninput={(e) => (description = inputValue(e))}
			></mb-textarea>
			<mb-button type="submit" variant="primary" loading={loading}>
				{loading ? 'Création…' : 'Créer'}
			</mb-button>
		</form>

		<section class="section">
			<h2>Équipes ({teams.length})</h2>
			{#if !ready}
				<p class="loading">Chargement…</p>
			{:else if teams.length === 0}
				<p class="muted">Aucune équipe pour cette organisation.</p>
			{:else}
				<ul class="card-list">
					{#each teams as team (team.id)}
						<li>
							<a href={`/admin/teams/${team.id}`}>
								<strong>{team.name}</strong>
								<span class="desc">{team.slug} · {team.member_count} membre(s)</span>
							</a>
						</li>
					{/each}
				</ul>
			{/if}
		</section>
	</AdminNav>
</div>
