<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import {
		deleteAdminJiraSettings,
		getAdminJiraSettings,
		postAdminJiraTest,
		putAdminJiraSettings,
		type JiraSettings
	} from '$lib/api/admin';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';

	const csrf = session().csrf_token;

	let error = $state('');
	let message = $state('');
	let loading = $state(true);
	let saving = $state(false);

	let baseUrl = $state('');
	let email = $state('');
	let apiToken = $state('');
	let projectKey = $state('');
	let issueType = $state('Task');
	let hasApiToken = $state(false);
	let configured = $state(false);

	function apply(s: JiraSettings) {
		baseUrl = s.base_url ?? '';
		email = s.email ?? '';
		projectKey = s.project_key ?? '';
		issueType = s.issue_type || 'Task';
		hasApiToken = s.has_api_token;
		configured = s.configured;
		apiToken = '';
	}

	onMount(async () => {
		try {
			apply(await getAdminJiraSettings(csrf));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Configuration Jira indisponible.';
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
				await putAdminJiraSettings(
					{
						base_url: baseUrl,
						email,
						project_key: projectKey,
						issue_type: issueType,
						...(apiToken ? { api_token: apiToken } : {})
					},
					csrf
				)
			);
			message = 'Configuration Jira enregistrée.';
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
			await postAdminJiraTest(csrf);
			message = 'Connexion Jira réussie.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Test impossible.';
		}
	}

	async function onClear() {
		if (!confirm('Effacer la configuration Jira ?')) return;
		error = '';
		message = '';
		try {
			await deleteAdminJiraSettings(csrf);
			apply({
				configured: false,
				base_url: '',
				email: '',
				project_key: '',
				issue_type: 'Task',
				has_api_token: false
			});
			message = 'Configuration Jira effacée.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Suppression impossible.';
		}
	}
</script>

<svelte:head>
	<title>Jira — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs">
			<a href="/admin">Administration</a> · <a href="/admin/integrations">Intégrations</a> · Jira
		</p>
		<h1>Jira Cloud</h1>
		<p class="lede">
			Le jeton est enregistré de façon sécurisée et n’est jamais réaffiché. Jira Server / Data
			Center n’est pas pris en charge.
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

	<AdminNav section="integrations">
		{#if loading}
			<p class="loading">Chargement…</p>
		{:else}
			<form class="stack-form" onsubmit={onSave}>
				<mb-input
					label="URL de l’instance"
					type="url"
					required
					autocomplete="off"
					placeholder="https://votre-domaine.atlassian.net"
					value={baseUrl}
					oninput={(e) => (baseUrl = inputValue(e))}
				></mb-input>
				<mb-input
					label="Email Atlassian"
					type="email"
					required
					autocomplete="username"
					value={email}
					oninput={(e) => (email = inputValue(e))}
				></mb-input>
				<mb-input
					label="Jeton Atlassian"
					type="password"
					autocomplete="new-password"
					hint={hasApiToken
						? 'Un jeton est enregistré ; laissez vide pour le conserver.'
						: 'Jeton Atlassian (compte → Sécurité).'}
					value={apiToken}
					oninput={(e) => (apiToken = inputValue(e))}
				></mb-input>
				<mb-input
					label="Clé de projet par défaut"
					hint="Requise pour créer des tickets, ex. REV."
					type="text"
					autocomplete="off"
					value={projectKey}
					oninput={(e) => (projectKey = inputValue(e))}
				></mb-input>
				<mb-input
					label="Type de ticket par défaut"
					type="text"
					autocomplete="off"
					placeholder="Task"
					value={issueType}
					oninput={(e) => (issueType = inputValue(e))}
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
				<h2>Test de connexion</h2>
				<p class="muted">Vérifie que Revues peut se connecter à votre instance Jira.</p>
				<mb-button type="button" variant="secondary" disabled={!configured} onclick={onTest}>
					Tester la connexion
				</mb-button>
			</section>
		{/if}
	</AdminNav>
</div>
