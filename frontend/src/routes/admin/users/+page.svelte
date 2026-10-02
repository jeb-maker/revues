<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { requireAdminSession, refreshCsrf } from '$lib/admin/session';
	import {
		createAllowedEmail,
		deleteAllowedEmail,
		listAllowedEmails,
		type AllowedEmailListResponse
	} from '$lib/api/admin';

	let csrf = $state('');
	let emails = $state<AllowedEmailListResponse['emails']>([]);
	let email = $state('');
	let role = $state<'reader' | 'editor'>('reader');
	let error = $state('');
	let message = $state('');
	let loading = $state(false);
	let ready = $state(false);

	onMount(async () => {
		try {
			const gate = await requireAdminSession();
			if (!gate) return;
			csrf = gate.csrf;
			const data = await listAllowedEmails();
			emails = data.emails ?? [];
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
			csrf = await refreshCsrf();
			const data = await listAllowedEmails();
			emails = data.emails ?? [];
			email = '';
			role = 'reader';
			message = 'Email enregistré.';
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

	async function onRemove(target: string) {
		if (!confirm('Retirer cet email de la liste autorisée ?')) return;
		error = '';
		message = '';
		try {
			await deleteAllowedEmail(target, csrf);
			csrf = await refreshCsrf();
			const data = await listAllowedEmails();
			emails = data.emails ?? [];
			message = 'Email retiré.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Retrait impossible.';
			try {
				csrf = await refreshCsrf();
			} catch {
				/* ignore */
			}
		}
	}
</script>

<svelte:head>
	<title>Emails autorisés — Revues</title>
</svelte:head>

<main class="admin">
	<p class="brand">Revues</p>
	<h1>Emails autorisés</h1>
	<p class="lede">
		Whitelist login (lecteur / éditeur). L'admin global se configure via
		<code>REVUES_BOOTSTRAP_ADMIN_EMAIL</code>.
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

		<AdminNav section="users">
			<form class="form" onsubmit={onAdd}>
				<h2>Ajouter ou mettre à jour</h2>
				<label>
					Email
					<input type="email" bind:value={email} required />
				</label>
				<label>
					Rôle
					<select bind:value={role} required>
						<option value="reader">Lecteur</option>
						<option value="editor">Éditeur</option>
					</select>
				</label>
				<mb-button type="submit" variant="primary" disabled={loading || !csrf}>
					{loading ? 'Enregistrement…' : 'Enregistrer'}
				</mb-button>
			</form>

			<section>
				<h2>Liste ({emails.length})</h2>
				{#if emails.length === 0}
					<p class="muted">Aucun email autorisé pour cette organisation.</p>
				{:else}
					<table>
						<thead>
							<tr>
								<th>Email</th>
								<th>Rôle</th>
								<th></th>
							</tr>
						</thead>
						<tbody>
							{#each emails as row (row.email)}
								<tr>
									<td>{row.email}</td>
									<td>{row.role}</td>
									<td>
										<mb-button
											type="button"
											variant="ghost"
											onclick={() => onRemove(String(row.email))}
										>
											Retirer
										</mb-button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
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
	select {
		padding: 0.55rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid rgba(148, 163, 184, 0.45);
		background: rgba(15, 23, 42, 0.55);
		color: #f8fafc;
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
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
	code {
		font-size: 0.85em;
		color: #99f6e4;
	}
</style>
