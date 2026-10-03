<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import {
		listOrganizationMembers,
		updateOrganizationMemberRole,
		type OrganizationMemberListResponse
	} from '$lib/api/admin';
	import { session } from '$lib/auth/session';
	import { roleOptions } from '$lib/i18n/labels';

	type OrgRole = 'owner' | 'admin' | 'member';
	const ORG_ROLES = roleOptions<OrgRole>(['owner', 'admin', 'member']);

	const csrf = session().csrf_token;

	let members = $state<OrganizationMemberListResponse['members']>([]);
	let error = $state('');
	let message = $state('');
	let ready = $state(false);
	let pendingId = $state<number | null>(null);

	async function reload() {
		const data = await listOrganizationMembers();
		members = data.members ?? [];
	}

	onMount(async () => {
		try {
			await reload();
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
			message = 'Rôle mis à jour.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Mise à jour impossible.';
			try {
				await reload();
			} catch {
				/* l'erreur initiale reste affichée */
			}
		} finally {
			pendingId = null;
		}
	}
</script>

<svelte:head>
	<title>Membres — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs"><a href="/admin">Administration</a> · Membres</p>
		<h1>Membres de l’organisation</h1>
		<p class="lede">
			Rôles : propriétaire, administrateur, membre. Le dernier propriétaire ne peut pas être
			rétrogradé.
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

	<AdminNav section="members">
		{#if !ready}
			<p class="loading">Chargement…</p>
		{:else if members.length === 0}
			<p class="muted">Aucun membre.</p>
		{:else}
			<div class="table-scroll">
				<table>
					<thead>
						<tr>
							<th scope="col">Membre</th>
							<th scope="col">Email</th>
							<th scope="col">Rôle</th>
						</tr>
					</thead>
					<tbody>
						{#each members as m (m.user_id)}
							<tr>
								<td>{m.display_name || m.login}</td>
								<td>{m.email}</td>
								<td>
									<mb-select
										label={`Rôle de ${m.display_name || m.login}`}
										hide-label
										name={`role-${m.user_id}`}
										required
										value={m.role}
										disabled={pendingId === m.user_id}
										onmb-change={(e) => onRoleChange(m.user_id, e.detail.value as OrgRole)}
									>
										{#each ORG_ROLES as opt (opt.value)}
											<option value={opt.value}>{opt.label}</option>
										{/each}
									</mb-select>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</AdminNav>
</div>
