<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		acceptOrganizationInvitation,
		listOrganizations,
		selectActiveOrganization,
		type OrganizationListResponse
	} from '$lib/api/orgs';
	import { resetSession, session } from '$lib/auth/session';

	let data = $state<OrganizationListResponse | null>(null);
	let selectedId = $state<number | null>(null);
	let error = $state('');
	let loading = $state(false);
	let ready = $state(false);
	let acceptingId = $state<number | null>(null);

	async function activate(organizationId: number) {
		const res = await selectActiveOrganization({ organization_id: organizationId }, session().csrf_token);
		resetSession();
		await goto(res.redirect || '/', { invalidateAll: true });
	}

	onMount(async () => {
		try {
			data = await listOrganizations();
			const invitations = data.invitations?.length ?? 0;
			if (data.organizations.length === 0 && invitations === 0) {
				await goto('/org/new');
				return;
			}
			if (data.organizations.length === 1 && invitations === 0) {
				await activate(data.organizations[0].id);
				return;
			}
			selectedId = data.default_organization_id ?? data.active_organization_id ?? null;
			if (selectedId == null && data.organizations[0]) {
				selectedId = data.organizations[0].id;
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger les organisations.';
		} finally {
			ready = true;
		}
	});

	async function onSubmit(e: Event) {
		e.preventDefault();
		if (selectedId == null) {
			error = 'Choisissez une organisation.';
			return;
		}
		error = '';
		loading = true;
		try {
			await activate(selectedId);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Sélection impossible.';
		} finally {
			loading = false;
		}
	}

	async function onAccept(invitationId: number) {
		error = '';
		acceptingId = invitationId;
		try {
			const res = await acceptOrganizationInvitation(invitationId, session().csrf_token);
			resetSession();
			await goto(res.redirect || '/', { invalidateAll: true });
		} catch (err) {
			error = err instanceof Error ? err.message : "Acceptation d'invitation impossible.";
			data = await listOrganizations().catch(() => data);
		} finally {
			acceptingId = null;
		}
	}
</script>

<svelte:head>
	<title>Choisir une organisation — Revues</title>
</svelte:head>

<div class="page page--narrow">
	<header class="page-header">
		<h1>Sélectionner une organisation</h1>
		<p class="lede">Sélectionnez l'organisation avec laquelle vous souhaitez travailler.</p>
	</header>

	{#if !ready}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else}
		{#if error}
			<mb-alert variant="danger">{error}</mb-alert>
		{/if}

		{#if data?.invitations && data.invitations.length > 0}
			<section class="section" aria-labelledby="invitations">
				<h2 id="invitations">Invitations</h2>
				<ul class="row-list">
					{#each data.invitations as inv (inv.id)}
						<li>
							<span>{inv.organization_name}</span>
							<mb-button
								type="button"
								variant="secondary"
								size="sm"
								disabled={acceptingId !== null}
								onclick={() => onAccept(inv.id)}
							>
								{acceptingId === inv.id ? 'Acceptation…' : 'Accepter'}
							</mb-button>
						</li>
					{/each}
				</ul>
			</section>
		{/if}

		{#if data && data.organizations.length > 0}
			<form class="stack-form" onsubmit={onSubmit}>
				<mb-radio-group
					label="Organisations"
					name="organization_id"
					value={selectedId == null ? '' : String(selectedId)}
					onmb-change={(e) => (selectedId = Number(e.detail.value))}
				>
					{#each data.organizations as org (org.id)}
						<mb-radio value={String(org.id)} label={`${org.name} (${org.slug})`}></mb-radio>
					{/each}
				</mb-radio-group>
				<mb-button type="submit" variant="primary" disabled={loading || selectedId == null}>
					{loading ? 'Continuation…' : 'Continuer'}
				</mb-button>
			</form>
		{/if}
	{/if}
</div>
