<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import {
		createAdminInvitation,
		deleteAdminInvitation,
		listAdminInvitations,
		listOrganizationMembers,
		updateOrganizationMemberRole,
		type AdminInvitationListResponse,
		type OrganizationMemberListResponse
	} from '$lib/api/admin';
	import { session } from '$lib/auth/session';
	import { roleOptions } from '$lib/i18n/labels';
	import { inputValue } from '$lib/mb';

	type OrgRole = 'owner' | 'admin' | 'member';
	const ORG_ROLES = roleOptions<OrgRole>(['owner', 'admin', 'member']);
	const INVITE_ROLES = roleOptions<OrgRole>(['member', 'admin', 'owner']);

	const csrf = session().csrf_token;

	let members = $state<OrganizationMemberListResponse['members']>([]);
	let invitations = $state<AdminInvitationListResponse['invitations']>([]);
	let error = $state('');
	let message = $state('');
	let ready = $state(false);
	let pendingId = $state<number | null>(null);
	let inviteEmail = $state('');
	let inviteRole = $state<OrgRole>('member');
	let inviting = $state(false);
	let revokingId = $state<number | null>(null);

	async function reload() {
		const [mem, inv] = await Promise.all([listOrganizationMembers(), listAdminInvitations()]);
		members = mem.members ?? [];
		invitations = inv.invitations ?? [];
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

	async function onInvite(e: Event) {
		e.preventDefault();
		error = '';
		message = '';
		inviting = true;
		try {
			const created = await createAdminInvitation(
				{ email: inviteEmail.trim(), org_role: inviteRole },
				csrf
			);
			inviteEmail = '';
			inviteRole = 'member';
			await reload();
			message = created.email_queued
				? 'Invitation enregistrée — un email a été mis en file.'
				: 'Invitation enregistrée — l’invité devra se connecter avec cet email (SMTP non configuré ou enqueue impossible).';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Invitation impossible.';
		} finally {
			inviting = false;
		}
	}

	async function onRevoke(id: number) {
		if (!confirm('Révoquer cette invitation ?')) return;
		error = '';
		message = '';
		revokingId = id;
		try {
			await deleteAdminInvitation(id, csrf);
			invitations = invitations.filter((i) => i.id !== id);
			message = 'Invitation révoquée.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Révocation impossible.';
		} finally {
			revokingId = null;
		}
	}

	function roleLabel(role: string): string {
		return ORG_ROLES.find((r) => r.value === role)?.label ?? role;
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
			rétrogradé. Invitez une personne par email pour qu’elle rejoigne l’organisation.
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
		{:else}
			<section class="stack-form" aria-labelledby="invite-heading">
				<h2 id="invite-heading">Inviter par email</h2>
				<form class="filters" onsubmit={onInvite}>
					<mb-input
						label="Email"
						type="email"
						name="email"
						required
						placeholder="personne@exemple.com"
						value={inviteEmail}
						oninput={(e) => (inviteEmail = inputValue(e))}
					></mb-input>
					<mb-select
						label="Rôle"
						value={inviteRole}
						onmb-change={(e) => (inviteRole = e.detail.value as OrgRole)}
					>
						{#each INVITE_ROLES as opt (opt.value)}
							<option value={opt.value}>{opt.label}</option>
						{/each}
					</mb-select>
					<mb-button type="submit" variant="primary" disabled={inviting}>Inviter</mb-button>
				</form>
			</section>

			<section aria-labelledby="pending-heading">
				<h2 id="pending-heading">Invitations en attente</h2>
				{#if invitations.length === 0}
					<p class="muted">Aucune invitation en attente.</p>
				{:else}
					<div class="table-scroll">
						<table>
							<thead>
								<tr>
									<th scope="col">Email</th>
									<th scope="col">Rôle</th>
									<th scope="col">Créée</th>
									<th scope="col">Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each invitations as inv (inv.id)}
									<tr>
										<td>{inv.email}</td>
										<td>{roleLabel(inv.org_role)}</td>
										<td>{inv.created_at}</td>
										<td>
											<mb-button
												variant="ghost"
												size="sm"
												disabled={revokingId === inv.id}
												onclick={() => onRevoke(inv.id)}
											>
												Révoquer
											</mb-button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</section>

			<section aria-labelledby="members-heading">
				<h2 id="members-heading">Membres</h2>
				{#if members.length === 0}
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
			</section>
		{/if}
	</AdminNav>
</div>
