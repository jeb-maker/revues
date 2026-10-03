<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import {
		deleteAdminNotionSettings,
		getAdminNotionSettings,
		postAdminNotionTest,
		putAdminNotionSettings,
		type NotionSettings
	} from '$lib/api/notion';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';

	const csrf = session().csrf_token;

	let error = $state('');
	let message = $state('');
	let loading = $state(true);
	let saving = $state(false);

	let apiToken = $state('');
	let workspaceName = $state('');
	let defaultDatabaseId = $state('');
	let hasApiToken = $state(false);
	let configured = $state(false);
	let exportReady = $state(false);

	const statusText = $derived(
		exportReady
			? 'Export prêt.'
			: configured
				? 'Renseignez une base par défaut pour activer l’export.'
				: 'Non configuré.'
	);

	function apply(s: NotionSettings) {
		workspaceName = s.workspace_name ?? '';
		defaultDatabaseId = s.default_database_id ?? '';
		hasApiToken = s.has_api_token;
		configured = s.configured;
		exportReady = s.export_ready;
		apiToken = '';
	}

	onMount(async () => {
		try {
			apply(await getAdminNotionSettings(csrf));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Configuration Notion indisponible.';
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
				await putAdminNotionSettings(
					{
						workspace_name: workspaceName,
						default_database_id: defaultDatabaseId,
						...(apiToken ? { api_token: apiToken } : { api_token: '' })
					},
					csrf
				)
			);
			message = 'Configuration Notion enregistrée.';
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
			const res = await postAdminNotionTest(csrf);
			message = res.message || 'Connexion Notion réussie.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Test impossible.';
		}
	}

	async function onClear() {
		if (!confirm('Effacer la configuration Notion ?')) return;
		error = '';
		message = '';
		try {
			await deleteAdminNotionSettings(csrf);
			apply({
				configured: false,
				export_ready: false,
				workspace_name: '',
				default_database_id: '',
				has_api_token: false
			});
			message = 'Configuration Notion effacée.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Suppression impossible.';
		}
	}
</script>

<svelte:head>
	<title>Notion — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="crumbs">
			<a href="/admin">Administration</a> · <a href="/admin/integrations">Intégrations</a> · Notion
		</p>
		<h1>Notion</h1>
		<p class="lede">
			Jeton chiffré — jamais renvoyé par l’API. Une base par défaut est requise pour l’export.
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
					label="Jeton d’intégration"
					type="password"
					autocomplete="new-password"
					placeholder="secret_… / ntn_…"
					hint={hasApiToken ? 'Un jeton est enregistré ; laissez vide pour le conserver.' : undefined}
					value={apiToken}
					oninput={(e) => (apiToken = inputValue(e))}
				></mb-input>
				<mb-input
					label="Espace de travail"
					type="text"
					autocomplete="off"
					value={workspaceName}
					oninput={(e) => (workspaceName = inputValue(e))}
				></mb-input>
				<mb-input
					label="Base par défaut"
					hint="Identifiant 32 hex ou UUID."
					type="text"
					autocomplete="off"
					value={defaultDatabaseId}
					oninput={(e) => (defaultDatabaseId = inputValue(e))}
				></mb-input>
				<p class="muted">{statusText}</p>
				<p class="actions">
					<mb-button type="submit" variant="primary" disabled={saving}>
						{saving ? 'Enregistrement…' : 'Enregistrer'}
					</mb-button>
					{#if configured}
						<mb-button type="button" variant="danger" onclick={onClear}>Effacer</mb-button>
					{/if}
				</p>
			</form>

			<section class="section">
				<h2>Test de connexion</h2>
				<p class="muted">Appelle <code>users/me</code> avec le jeton enregistré.</p>
				<mb-button type="button" variant="secondary" disabled={!configured} onclick={onTest}>
					Tester la connexion
				</mb-button>
			</section>
		{/if}
	</AdminNav>
</div>
