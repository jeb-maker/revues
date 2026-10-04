<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import {
		deleteAdminSMTPSettings,
		getAdminSMTPSettings,
		postAdminSMTPTest,
		putAdminSMTPSettings,
		type SMTPSettings
	} from '$lib/api/admin';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';

	const boot = session();
	const csrf = boot.csrf_token;

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
	let testRecipient = $state(boot.user?.email ?? '');

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
			apply(await getAdminSMTPSettings(csrf));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Configuration SMTP indisponible.';
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
		if (!confirm('Effacer la configuration SMTP ?')) return;
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

<div class="page">
	<header class="page-header">
		<p class="crumbs"><a href="/admin">Administration</a> · SMTP</p>
		<h1>Relais SMTP</h1>
		<p class="lede">Le mot de passe est enregistré de façon sécurisée et n’est jamais réaffiché.</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

	<AdminNav section="smtp">
		{#if loading}
			<p class="loading">Chargement…</p>
		{:else}
			<form class="stack-form" onsubmit={onSave}>
				<mb-input
					label="Hôte"
					type="text"
					required
					autocomplete="off"
					value={host}
					oninput={(e) => (host = inputValue(e))}
				></mb-input>
				<mb-input
					label="Port"
					type="number"
					min="1"
					max="65535"
					required
					value={String(port)}
					oninput={(e) => (port = Number(inputValue(e)) || 0)}
				></mb-input>
				<mb-checkbox
					label="Connexion TLS"
					checked={tls}
					onmb-change={(e) => (tls = !!e.detail.checked)}
				></mb-checkbox>
				<mb-input
					label="Identifiant"
					type="text"
					autocomplete="username"
					value={username}
					oninput={(e) => (username = inputValue(e))}
				></mb-input>
				<mb-input
					label="Mot de passe"
					type="password"
					autocomplete="new-password"
					hint={hasPassword ? 'Un mot de passe est enregistré ; laissez vide pour le conserver.' : undefined}
					value={password}
					oninput={(e) => (password = inputValue(e))}
				></mb-input>
				<mb-input
					label="Expéditeur"
					type="email"
					required
					value={from}
					oninput={(e) => (from = inputValue(e))}
				></mb-input>
				<p class="actions">
					<mb-button type="submit" variant="primary" loading={saving}>
						{saving ? 'Enregistrement…' : 'Enregistrer'}
					</mb-button>
					{#if configured}
						<mb-button type="button" variant="danger" onclick={onClear}>Effacer</mb-button>
					{/if}
				</p>
			</form>

			<section class="section">
				<h2>Email de test</h2>
				<div class="stack-form">
					<mb-input
						label="Destinataire"
						type="email"
						value={testRecipient}
						oninput={(e) => (testRecipient = inputValue(e))}
					></mb-input>
					<mb-button type="button" variant="secondary" disabled={!configured} onclick={onTest}>
						Envoyer un test
					</mb-button>
				</div>
			</section>
		{/if}
	</AdminNav>
</div>
