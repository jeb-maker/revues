<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { requireAdminSession, refreshCsrf } from '$lib/admin/session';
	import { getLeadPolicies, updateLeadPolicies, type LeadPolicies } from '$lib/api/admin';

	let csrf = $state('');
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
			const gate = await requireAdminSession();
			if (!gate) return;
			csrf = gate.csrf;
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
			policies = await updateLeadPolicies(policies, csrf);
			csrf = await refreshCsrf();
			message = 'Politiques mises à jour.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
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
	<title>Politiques — Revues</title>
</svelte:head>

<main class="admin">
	<p class="brand">Revues</p>
	<h1>Politiques de délégation</h1>
	<p class="lede">
		Contrôlez ce que les responsables de sujet peuvent faire sans passer par un admin org.
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

		<AdminNav section="policies">
			<form class="form" onsubmit={onSave}>
				<label class="check">
					<input type="checkbox" bind:checked={policies.leads_may_assign_teams} />
					Les responsables peuvent affecter des équipes à leur sujet
				</label>
				<label class="check">
					<input type="checkbox" bind:checked={policies.leads_may_invite_members} />
					Les responsables peuvent inviter des membres de l'organisation
				</label>
				<label class="check">
					<input type="checkbox" bind:checked={policies.leads_may_invite_externals} />
					Les responsables peuvent inviter des personnes hors organisation
				</label>
				<p class="hint">Les admins org et l'admin global ne sont jamais limités par ces options.</p>
				<mb-button type="submit" variant="primary" disabled={loading || !csrf}>
					{loading ? 'Enregistrement…' : 'Enregistrer'}
				</mb-button>
			</form>
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
	.lede {
		margin: 0 0 1.25rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
	}
	.check {
		display: flex;
		align-items: flex-start;
		gap: 0.6rem;
		line-height: 1.4;
		cursor: pointer;
	}
	.hint {
		margin: 0;
		font-size: 0.85rem;
		color: #94a3b8;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
