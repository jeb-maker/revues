<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import {
		deleteAdminSMTPSettings,
		getAdminSMTPSettings,
		postAdminSMTPTest,
		putAdminSMTPSettings,
		type SMTPSettings
	} from '$lib/api/admin';

	let csrf = $state('');
	let error = $state('');
	let message = $state('');
	let loading = $state(true);
	let saving = $state(false);

	let host = $state('');
	let port = $state(587);
	let tls = $state(true);
	let username = $state('');
	let password = $state('');
	let from = $state('');
	let hasPassword = $state(false);
	let configured = $state(false);
	let testRecipient = $state('');

	function apply(s: SMTPSettings) {
		host = s.host ?? '';
		port = s.port || 587;
		tls = s.tls;
		username = s.username ?? '';
		from = s.from ?? '';
		hasPassword = s.has_password;
		configured = s.configured;
		password = '';
	}

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			testRecipient = boot.user?.email ?? '';
			apply(await getAdminSMTPSettings(csrf));
		} catch (e) {
			error = e instanceof Error ? e.message : 'SMTP indisponible.';
		} finally {
			loading = false;
		}
	});

	async function onSave(e: Event) {
		e.preventDefault();
		saving = true;
		error = '';
		message = '';
		try {
			apply(
				await putAdminSMTPSettings(
					{
						host,
						port,
						tls,
						username,
						from,
						...(password ? { password } : { password: '' })
					},
					csrf
				)
			);
			message = 'Configuration SMTP enregistrée.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
		} finally {
			saving = false;
		}
	}

	async function onTest() {
		error = '';
		message = '';
		try {
			await postAdminSMTPTest(csrf, testRecipient || undefined);
			message = 'Email de test envoyé.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Test impossible.';
		}
	}

	async function onClear() {
		error = '';
		message = '';
		try {
			await deleteAdminSMTPSettings(csrf);
			apply({
				configured: false,
				enabled: false,
				host: '',
				port: 587,
				tls: true,
				username: '',
				from: '',
				has_password: false
			});
			message = 'Configuration SMTP effacée.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Suppression impossible.';
		}
	}
</script>

<svelte:head>
	<title>SMTP — Revues</title>
</svelte:head>

<main class="admin-page">
	<header>
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs">
			<a href="/">Accueil</a> · <a href="/admin/integrations">Intégrations</a> · SMTP
		</p>
		<h1>Relais SMTP</h1>
		<p class="lede">Configuration chiffrée — le mot de passe n'est jamais renvoyé par l'API.</p>
	</header>

	{#if error}
		<p class="err" role="alert">{error}</p>
	{/if}
	{#if message}
		<p class="ok" role="status">{message}</p>
	{/if}

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else}
		<form class="form" onsubmit={onSave}>
			<label>
				Hôte
				<input bind:value={host} required autocomplete="off" />
			</label>
			<label>
				Port
				<input type="number" min="1" max="65535" bind:value={port} required />
			</label>
			<label class="check">
				<input type="checkbox" bind:checked={tls} />
				TLS
			</label>
			<label>
				Identifiant
				<input bind:value={username} autocomplete="username" />
			</label>
			<label>
				Mot de passe
				<input
					type="password"
					bind:value={password}
					placeholder={hasPassword ? '•••••••• (laisser vide pour conserver)' : ''}
					autocomplete="new-password"
				/>
			</label>
			<label>
				Expéditeur
				<input type="email" bind:value={from} required />
			</label>
			<div class="actions">
				<button type="submit" disabled={saving}>{saving ? 'Enregistrement…' : 'Enregistrer'}</button>
				{#if configured}
					<button type="button" class="ghost" onclick={onClear}>Effacer</button>
				{/if}
			</div>
		</form>

		<section class="test">
			<h2>Email de test</h2>
			<label>
				Destinataire
				<input type="email" bind:value={testRecipient} />
			</label>
			<button type="button" disabled={!configured} onclick={onTest}>Envoyer un test</button>
		</section>
	{/if}
</main>

<style>
	.form,
	.test {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
	}
	.test {
		margin-top: 2rem;
		padding-top: 1.25rem;
		border-top: 1px solid #334155;
	}
	.test h2 {
		margin: 0;
		font-size: 1rem;
		color: #99f6e4;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	label.check {
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;
	}
	input:not([type='checkbox']) {
		padding: 0.55rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	.actions {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
	}
	button {
		align-self: flex-start;
		padding: 0.6rem 1rem;
		border: none;
		border-radius: 0.4rem;
		background: #0f766e;
		color: #ecfdf5;
		font-weight: 600;
		cursor: pointer;
		font: inherit;
	}
	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	button.ghost {
		background: transparent;
		border: 1px solid #475569;
		color: #cbd5e1;
	}
</style>
