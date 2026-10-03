<script lang="ts">
	import { page } from '$app/state';

	const title = $derived(page.status === 404 ? 'Page introuvable' : 'Une erreur est survenue');
</script>

<svelte:head>
	<title>{title} — Revues</title>
</svelte:head>

<div class="page page--narrow">
	<header class="page-header">
		<h1>{title}</h1>
		<p class="lede">
			{#if page.status === 404}
				Cette adresse ne correspond à aucune page de Revues.
			{:else}
				{page.error?.message ?? 'Réessayez dans un instant.'}
			{/if}
		</p>
	</header>
	<p class="actions"><a href="/">Retour à l'accueil</a></p>
</div>
