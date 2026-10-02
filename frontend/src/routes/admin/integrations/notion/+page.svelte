<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';

	let ready = $state(false);
	onMount(async () => {
		const boot = await bootstrap();
		if (!boot.authenticated) {
			await goto('/login');
			return;
		}
		ready = true;
	});
</script>

<svelte:head>
	<title>Notion — Revues</title>
</svelte:head>

{#if ready}
	<main class="admin-page">
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs"><a href="/admin/integrations">Intégrations</a> · Notion</p>
		<h1>Notion</h1>
		<p class="lede">
			Import / export Notion arriveront avec WP-021. L'état de configuration est déjà visible
			dans le hub intégrations.
		</p>
		<p><a class="back" href="/admin/integrations">← Retour au hub</a></p>
	</main>
{/if}
