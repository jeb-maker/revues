<script lang="ts">
	import { onMount } from 'svelte';
	import {
		deleteAdminConfluenceSettings,
		getAdminConfluenceSettings,
		postAdminConfluenceTest,
		putAdminConfluenceSettings
	} from '$lib/api/confluence';
	import { session } from '$lib/auth/session';

	const csrf = session().csrf_token;

	let error = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let baseUrl = $state('');
	let email = $state('');
	let apiToken = $state('');
	let spaceKey = $state('');
	let parentPageId = $state('');
	let hasApiToken = $state(false);
	let configured = $state(false);

	onMount(async () => {
		try {
			const s = await getAdminConfluenceSettings(csrf);
			baseUrl = s.base_url ?? '';
			email = s.email ?? '';
			spaceKey = s.space_key ?? '';
			parentPageId = s.parent_page_id ?? '';
			hasApiToken = s.has_api_token;
			configured = s.configured;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Indisponible.';
		} finally {
			loading = false;
		}
	});

	async function onSave(e: Event) {
		e.preventDefault();
		saving = true;
		error = '';
		try {
			const s = await putAdminConfluenceSettings(
				{
					base_url: baseUrl,
					email,
					space_key: spaceKey,
					parent_page_id: parentPageId,
					...(apiToken ? { api_token: apiToken } : {})
				},
				csrf
			);
			hasApiToken = s.has_api_token;
			configured = s.configured;
			apiToken = '';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Échec.';
		} finally {
			saving = false;
		}
	}

	async function onTest() {
		error = '';
		try {
			await postAdminConfluenceTest(csrf);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Échec.';
		}
	}

	async function onClear() {
		if (!confirm('Effacer ?')) return;
		error = '';
		try {
			await deleteAdminConfluenceSettings(csrf);
			baseUrl = email = spaceKey = parentPageId = apiToken = '';
			hasApiToken = configured = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Échec.';
		}
	}
</script>

<svelte:head><title>Confluence — Revues</title></svelte:head>

<header class="page-header">
	<p class="crumbs">
		<a href="/admin">Administration</a> · <a href="/admin/integrations">Intégrations</a> · Confluence
	</p>
	<h1>Confluence Cloud</h1>
	<p class="lede">Jeton chiffré. Archive wiki d’une revue clôturée.</p>
</header>

{#if error}<mb-alert variant="danger">{error}</mb-alert>{/if}

{#if loading}
	<p class="loading">Chargement…</p>
{:else}
	<form class="stack-form" onsubmit={onSave}>
		<label>URL <input type="url" required autocomplete="off" bind:value={baseUrl} /></label>
		<label>Email <input type="email" required autocomplete="username" bind:value={email} /></label>
		<label
			>Jeton{hasApiToken ? ' (vide = garder)' : ''}
			<input type="password" autocomplete="new-password" bind:value={apiToken} />
		</label>
		<label>Espace <input type="text" required autocomplete="off" bind:value={spaceKey} /></label>
		<label>Page parente <input type="text" autocomplete="off" bind:value={parentPageId} /></label>
		<p class="actions">
			<mb-button type="submit" variant="primary" loading={saving}>Enregistrer</mb-button>
			{#if configured}
				<mb-button type="button" variant="danger" onclick={onClear}>Effacer</mb-button>
			{/if}
			<mb-button type="button" variant="secondary" disabled={!configured} onclick={onTest}
				>Tester</mb-button
			>
		</p>
	</form>
{/if}
