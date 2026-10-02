<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { requireAdminSession, refreshCsrf } from '$lib/admin/session';
	import {
		listOrganizationMembers,
		updateOrganizationMemberRole,
		type OrganizationMemberListResponse
	} from '$lib/api/admin';

	type OrgRole = 'owner' | 'admin' | 'member';

	let csrf = $state('');
	let members = $state<OrganizationMemberListResponse['members']>([]);
	let error = $state('');
	let message = $state('');
	let ready = $state(false);
	let pendingId = $state<number | null>(null);

	onMount(async () => {
		try {
			const gate = await requireAdminSession();
			if (!gate) return;
			csrf = gate.csrf;
			const data = await listOrganizationMembers();
			members = data.members ?? [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
		} finally {
			ready = true;
		}
	});

	async function onRoleChange(userId: number, role: OrgRole) {
		error = '';
		message = '';
		pendingId = userId;
		try {
			const updated = await updateOrganizationMemberRole(userId, { role }, csrf);
			members = members.map((m) => (m.user_id === userId ? updated : m));
			csrf = await refreshCsrf();
			message = 'Rôle mis à jour.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Mise à jour impossible.';
			try {
				csrf = await refreshCsrf();
				const data = await listOrganizationMembers();
				members = data.members ?? [];
			} catch {
				/* ignore */
			}
		} finally {
			pendingId = null;
		}
	}
</script>

<svelte:head>
	<title>Membres — Revues</title>
</svelte:head>

<main class="admin">
	<p class="brand">Revues</p>
	<h1>Membres de l'organisation</h1>
	<p class="lede">Rôles org : owner, admin, member. Le dernier owner ne peut pas être rétrogradé.</p>

	{#if !ready}
		<p class="muted">Chargement…</p>
	{:else}
		{#if error}
			<mb-alert variant="danger" role="alert">{error}</mb-alert>
		{/if}
		{#if message}
			<mb-alert variant="success" role="status">{message}</mb-alert>
		{/if}

		<AdminNav section="members">
			{#if members.length === 0}
				<p class="muted">Aucun membre.</p>
			{:else}
				<table>
					<thead>
						<tr>
							<th>Membre</th>
							<th>Email</th>
							<th>Rôle</th>
						</tr>
					</thead>
					<tbody>
						{#each members as m (m.user_id)}
							<tr>
								<td>{m.display_name || m.login}</td>
								<td>{m.email}</td>
								<td>
									<select
										value={m.role}
										disabled={pendingId === m.user_id}
										onchange={(e) =>
											onRoleChange(m.user_id, (e.currentTarget as HTMLSelectElement).value as OrgRole)}
									>
										<option value="owner">owner</option>
										<option value="admin">admin</option>
										<option value="member">member</option>
									</select>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
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
	.lede {
		margin: 0 0 1.25rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.92rem;
	}
	th,
	td {
		text-align: left;
		padding: 0.55rem 0.35rem;
		border-bottom: 1px solid rgba(148, 163, 184, 0.25);
	}
	select {
		padding: 0.35rem 0.45rem;
		border-radius: 0.35rem;
		border: 1px solid rgba(148, 163, 184, 0.45);
		background: rgba(15, 23, 42, 0.55);
		color: #f8fafc;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
