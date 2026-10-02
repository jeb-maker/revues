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
	<title>Webhooks — Revues</title>
</svelte:head>

{#if ready}
	<main class="admin-page">
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs"><a href="/admin/integrations">Intégrations</a> · Webhooks</p>
		<h1>Webhooks sortants</h1>
		<p class="lede">
			Configuration, signature HMAC et drain arriveront avec WP-022. Le hub indique déjà si des
			URLs sont configurées.
		</p>
		<p><a class="back" href="/admin/integrations">← Retour au hub</a></p>
	</main>
{/if}
