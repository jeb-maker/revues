<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import {
		acceptOrganizationInvitation,
		listOrganizations,
		selectActiveOrganization,
		type OrganizationListResponse
	} from '$lib/api/orgs';

	let csrf = $state('');
	let data = $state<OrganizationListResponse | null>(null);
	let selectedId = $state<number | null>(null);
	let error = $state('');
	let loading = $state(false);
	let ready = $state(false);
	let acceptingId = $state<number | null>(null);

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			data = await listOrganizations();
			if (data.organizations.length === 0 && (data.invitations?.length ?? 0) === 0) {
				await goto('/org/new');
				return;
			}
			if (data.organizations.length === 1 && (data.invitations?.length ?? 0) === 0) {
				const only = data.organizations[0];
				const res = await selectActiveOrganization({ organization_id: only.id }, csrf);
				await goto(res.redirect || '/');
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
			const res = await selectActiveOrganization({ organization_id: selectedId }, csrf);
			await goto(res.redirect || '/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Sélection impossible.';
			try {
				const boot = await bootstrap();
				csrf = boot.csrf_token;
			} catch {
				/* ignore */
			}
		} finally {
			loading = false;
		}
	}

	async function onAccept(invitationId: number) {
		error = '';
		acceptingId = invitationId;
		try {
			const res = await acceptOrganizationInvitation(invitationId, csrf);
			await goto(res.redirect || '/');
		} catch (err) {
			error = err instanceof Error ? err.message : "Acceptation d'invitation impossible.";
			try {
				const boot = await bootstrap();
				csrf = boot.csrf_token;
				data = await listOrganizations();
			} catch {
				/* ignore */
			}
		} finally {
			acceptingId = null;
		}
	}
</script>

<svelte:head>
	<title>Choisir une organisation — Revues</title>
</svelte:head>

<main class="org">
	<p class="brand">Revues</p>
	<h1>Sélectionner une organisation</h1>
	<p class="lede">Sélectionnez l'organisation avec laquelle vous souhaitez travailler.</p>

	{#if !ready}
		<p class="muted">Chargement…</p>
	{:else}
		{#if error}
			<mb-alert variant="danger" role="alert">{error}</mb-alert>
		{/if}

		{#if data?.invitations && data.invitations.length > 0}
			<section class="invites" aria-label="Invitations en attente">
				<h2>Invitations</h2>
				<ul>
					{#each data.invitations as inv (inv.id)}
						<li>
							<span>{inv.organization_name}</span>
							<mb-button
								type="button"
								variant="secondary"
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
			<form class="form" onsubmit={onSubmit}>
				<fieldset class="orgs">
					<legend>Organisations</legend>
					{#each data.organizations as org (org.id)}
						<label class="choice">
							<input
								type="radio"
								name="organization_id"
								value={org.id}
								checked={selectedId === org.id}
								onchange={() => {
									selectedId = org.id;
								}}
								required
							/>
							<span>{org.name} <span class="slug">({org.slug})</span></span>
						</label>
					{/each}
				</fieldset>
				<mb-button type="submit" variant="primary" disabled={loading || !csrf || selectedId == null}>
					{loading ? 'Continuation…' : 'Continuer'}
				</mb-button>
			</form>
		{/if}
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background:
			radial-gradient(ellipse 80% 50% at 90% 10%, rgba(15, 118, 110, 0.3), transparent 50%),
			linear-gradient(165deg, #0f172a 0%, #1e293b 50%, #134e4a 100%);
		color: #f8fafc;
	}
	.org {
		max-width: 28rem;
		margin: 0 auto;
		padding: 12vh 1.25rem 3rem;
	}
	.brand {
		margin: 0 0 0.5rem;
		font-size: clamp(2.25rem, 7vw, 3.25rem);
		font-weight: 700;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	h1 {
		margin: 0 0 0.5rem;
		font-size: 1.15rem;
		font-weight: 500;
		color: #99f6e4;
	}
	h2 {
		margin: 0 0 0.75rem;
		font-size: 0.95rem;
		font-weight: 600;
		color: #e2e8f0;
	}
	.lede {
		margin: 0 0 1.5rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}
	.orgs {
		border: 1px solid rgba(148, 163, 184, 0.35);
		border-radius: 0.5rem;
		padding: 0.75rem 1rem 1rem;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.65rem;
	}
	.orgs legend {
		padding: 0 0.35rem;
		font-size: 0.85rem;
		color: #cbd5e1;
	}
	.choice {
		display: flex;
		align-items: flex-start;
		gap: 0.6rem;
		cursor: pointer;
		line-height: 1.35;
	}
	.slug {
		color: #94a3b8;
		font-size: 0.9em;
	}
	.invites {
		margin-bottom: 1.5rem;
	}
	.invites ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.invites li {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.65rem 0.75rem;
		border: 1px solid rgba(148, 163, 184, 0.35);
		border-radius: 0.5rem;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
