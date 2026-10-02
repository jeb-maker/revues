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
	<title>Jira — Revues</title>
</svelte:head>

{#if ready}
	<main class="admin-page">
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs"><a href="/admin/integrations">Intégrations</a> · Jira</p>
		<h1>Jira Cloud</h1>
		<p class="lede">
			Configuration et liaison de tickets arriveront avec WP-020. Le hub affiche déjà l'état
			configuré / non configuré.
		</p>
		<p><a class="back" href="/admin/integrations">← Retour au hub</a></p>
	</main>
{/if}
