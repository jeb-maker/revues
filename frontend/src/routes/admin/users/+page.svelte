<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import {
		createAllowedEmail,
		deleteAllowedEmail,
		listAllowedEmails,
		type AllowedEmailListResponse
	} from '$lib/api/admin';
	import { session } from '$lib/auth/session';
	import { formatRole, roleOptions } from '$lib/i18n/labels';
	import { inputValue } from '$lib/mb';

	type LoginRole = 'reader' | 'editor';
	const ROLES = roleOptions<LoginRole>(['reader', 'editor']);

	const csrf = session().csrf_token;

	let emails = $state<AllowedEmailListResponse['emails']>([]);
	let email = $state('');
	let role = $state<LoginRole>('reader');
	let error = $state('');
	let message = $state('');
	let loading = $state(false);
	let ready = $state(false);

	async function reload() {
		const data = await listAllowedEmails();
		emails = data.emails ?? [];
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

	async function onAdd(e: Event) {
		e.preventDefault();
		error = '';
		message = '';
		loading = true;
		try {
			await createAllowedEmail({ email, role }, csrf);
			await reload();
			email = '';
			role = 'reader';
			message = 'Email enregistré.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
		} finally {
			loading = false;
		}
	}

	async function onRemove(target: string) {
		if (!confirm(`Retirer ${target} de la liste autorisée ?`)) return;
		error = '';
		message = '';
		try {
			await deleteAllowedEmail(target, csrf);
			await reload();
			message = 'Email retiré.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Retrait impossible.';
		}
	}
</script>

<svelte:head>
	<title>Emails autorisés — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs"><a href="/admin">Administration</a> · Emails autorisés</p>
		<h1>Emails autorisés</h1>
		<p class="lede">
			Liste blanche de connexion (lecteur / éditeur). L’administrateur global se configure via
			<code>REVUES_BOOTSTRAP_ADMIN_EMAIL</code>.
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

	<AdminNav section="users">
		<form class="stack-form" onsubmit={onAdd}>
			<h2>Ajouter ou mettre à jour</h2>
			<mb-input
				label="Email"
				type="email"
				required
				autocomplete="off"
				value={email}
				oninput={(e) => (email = inputValue(e))}
			></mb-input>
			<mb-select
				label="Rôle"
				name="role"
				required
				value={role}
				onmb-change={(e) => (role = e.detail.value as LoginRole)}
			>
				{#each ROLES as opt (opt.value)}
					<option value={opt.value}>{opt.label}</option>
				{/each}
			</mb-select>
			<mb-button type="submit" variant="primary" disabled={loading}>
				{loading ? 'Enregistrement…' : 'Enregistrer'}
			</mb-button>
		</form>

		<section class="section">
			<h2>Liste ({emails.length})</h2>
			{#if !ready}
				<p class="loading">Chargement…</p>
			{:else if emails.length === 0}
				<p class="muted">Aucun email autorisé pour cette organisation.</p>
			{:else}
				<div class="table-scroll">
					<table>
						<thead>
							<tr>
								<th scope="col">Email</th>
								<th scope="col">Rôle</th>
								<th scope="col"><span class="sr-only">Actions</span></th>
							</tr>
						</thead>
						<tbody>
							{#each emails as row (row.email)}
								<tr>
									<td>{row.email}</td>
									<td>{formatRole(row.role)}</td>
									<td>
										<mb-button
											type="button"
											variant="ghost"
											size="sm"
											onclick={() => onRemove(String(row.email))}
										>
											Retirer
										</mb-button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</section>
	</AdminNav>
</div>
