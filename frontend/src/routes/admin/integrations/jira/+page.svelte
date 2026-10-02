<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import {
		deleteAdminJiraSettings,
		getAdminJiraSettings,
		postAdminJiraTest,
		putAdminJiraSettings,
		type JiraSettings
	} from '$lib/api/admin';

	let csrf = $state('');
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
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			apply(await getAdminJiraSettings(csrf));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Jira indisponible.';
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

<main class="admin-page">
	<header>
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs">
			<a href="/">Accueil</a> · <a href="/admin/integrations">Intégrations</a> · Jira
		</p>
		<h1>Jira Cloud</h1>
		<p class="lede">
			Credentials chiffrés — le jeton API n'est jamais renvoyé par l'API. Server/DC hors scope.
		</p>
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
				URL de l'instance
				<input
					type="url"
					bind:value={baseUrl}
					required
					autocomplete="off"
					placeholder="https://votre-domaine.atlassian.net"
				/>
			</label>
			<label>
				Email Atlassian
				<input type="email" bind:value={email} required autocomplete="username" />
			</label>
			<label>
				Jeton API
				<input
					type="password"
					bind:value={apiToken}
					placeholder={hasApiToken ? '•••••••• (laisser vide pour conserver)' : 'Jeton API Atlassian'}
					autocomplete="new-password"
				/>
			</label>
			<label>
				Clé projet par défaut
				<input
					bind:value={projectKey}
					autocomplete="off"
					placeholder="REV (requis pour créer des tickets)"
				/>
			</label>
			<label>
				Type d'issue par défaut
				<input bind:value={issueType} autocomplete="off" placeholder="Task" />
			</label>
			<div class="actions">
				<button type="submit" disabled={saving}>{saving ? 'Enregistrement…' : 'Enregistrer'}</button>
				{#if configured}
					<button type="button" class="ghost" onclick={onClear}>Effacer</button>
				{/if}
			</div>
		</form>

		<section class="test">
			<h2>Test de connexion</h2>
			<p class="muted">Vérifie les identifiants enregistrés via l'API REST Jira (<code>/myself</code>).</p>
			<button type="button" disabled={!configured} onclick={onTest}>Tester la connexion</button>
		</section>
	{/if}
</main>
