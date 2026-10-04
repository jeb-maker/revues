<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { getLeadPolicies, updateLeadPolicies, type LeadPolicies } from '$lib/api/admin';
	import { session } from '$lib/auth/session';

	const csrf = session().csrf_token;

	let policies = $state<LeadPolicies>({
		leads_may_assign_teams: true,
		leads_may_invite_members: true,
		leads_may_invite_externals: false
	});
	let error = $state('');
	let message = $state('');
	let loading = $state(false);
	let ready = $state(false);

	onMount(async () => {
		try {
			policies = await getLeadPolicies();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
		} finally {
			ready = true;
		}
	});

	async function onSave(e: Event) {
		e.preventDefault();
		error = '';
		message = '';
		loading = true;
		try {
			// Conserver leads_may_assign_teams tel que chargé (UI masquée, icebox équipes).
			policies = await updateLeadPolicies(policies, csrf);
			message = 'Politiques mises à jour.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Politiques — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs"><a href="/admin">Administration</a> · Politiques</p>
		<h1>Politiques de délégation</h1>
		<p class="lede">
			Contrôlez ce que les référents de sujet peuvent faire sans passer par un administrateur.
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

	<AdminNav section="policies">
		{#if !ready}
			<p class="loading">Chargement…</p>
		{:else}
			<form class="stack-form" onsubmit={onSave}>
				<mb-checkbox
					label="Les référents peuvent affecter des personnes membres de l’organisation"
					checked={policies.leads_may_invite_members}
					onmb-change={(e) => (policies.leads_may_invite_members = !!e.detail.checked)}
				></mb-checkbox>
				<mb-checkbox
					label="Les référents peuvent affecter des personnes hors organisation (compte existant)"
					checked={policies.leads_may_invite_externals}
					onmb-change={(e) => (policies.leads_may_invite_externals = !!e.detail.checked)}
				></mb-checkbox>
				<p class="field-hint">
					Les administrateurs de l’organisation et l’administrateur global ne sont jamais limités
					par ces options.
				</p>
				<mb-button type="submit" variant="primary" loading={loading}>
					{loading ? 'Enregistrement…' : 'Enregistrer'}
				</mb-button>
			</form>
		{/if}
	</AdminNav>
</div>
