<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import {
		deleteAdminNotionSettings,
		getAdminNotionSettings,
		postAdminNotionTest,
		putAdminNotionSettings,
		type NotionSettings
	} from '$lib/api/notion';

	let csrf = $state('');
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
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			apply(await getAdminNotionSettings(csrf));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Notion indisponible.';
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

<main class="admin-page">
	<header>
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs">
			<a href="/">Accueil</a> · <a href="/admin/integrations">Intégrations</a> · Notion
		</p>
		<h1>Notion</h1>
		<p class="lede">Jeton chiffré — jamais renvoyé. Base par défaut requise pour l'export.</p>
	</header>

	{#if error}<p class="err" role="alert">{error}</p>{/if}
	{#if message}<p class="ok" role="status">{message}</p>{/if}

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else}
		<form class="stack" onsubmit={onSave}>
			<label
				>Jeton<input
					type="password"
					bind:value={apiToken}
					placeholder={hasApiToken ? '•••• (vide = conserver)' : 'secret_… / ntn_…'}
					autocomplete="new-password"
				/></label
			>
			<label>Workspace<input bind:value={workspaceName} autocomplete="off" /></label>
			<label
				>Base par défaut<input
					bind:value={defaultDatabaseId}
					placeholder="32 hex / UUID"
					autocomplete="off"
				/></label
			>
			<p class="muted">
				{exportReady
					? 'Export prêt.'
					: configured
						? 'Renseignez une base pour l’export.'
						: 'Non configuré.'}
			</p>
			<div class="row">
				<button type="submit" disabled={saving}>{saving ? '…' : 'Enregistrer'}</button>
				{#if configured}
					<button type="button" class="ghost" onclick={onClear}>Effacer</button>
				{/if}
			</div>
		</form>
		<section class="stack sep">
			<h2>Test</h2>
			<button type="button" disabled={!configured} onclick={onTest}>Tester users/me</button>
		</section>
	{/if}
</main>

<style>
	.stack {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.sep {
		margin-top: 1.5rem;
		padding-top: 1rem;
		border-top: 1px solid #334155;
	}
	.sep h2 {
		margin: 0;
		font-size: 1rem;
		color: #99f6e4;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	input {
		padding: 0.5rem 0.6rem;
		border-radius: 0.35rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	.row {
		display: flex;
		gap: 0.65rem;
		flex-wrap: wrap;
	}
	button {
		padding: 0.55rem 0.9rem;
		border: none;
		border-radius: 0.35rem;
		background: #0f766e;
		color: #ecfdf5;
		font: inherit;
		font-weight: 600;
		cursor: pointer;
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
