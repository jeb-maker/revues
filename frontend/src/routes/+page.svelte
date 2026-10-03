<script lang="ts">
	import { session } from '$lib/auth/session';

	const boot = session();

	const tiles = $derived([
		{ href: '/runs', title: 'Revues', desc: 'Exécutions en cours et historiques, progression par snapshot.' },
		{ href: '/subjects', title: 'Sujets', desc: 'Conteneurs de revues : domaines, étiquettes, membres.' },
		{ href: '/modeles', title: 'Modèles', desc: 'Check-lists versionnées de l’organisation.' },
		{ href: '/mes-taches', title: 'Mes tâches', desc: 'Points de revue qui vous sont assignés.' },
		...(boot.can_admin
			? [{ href: '/admin', title: 'Administration', desc: 'Accès, équipes, politiques, intégrations.' }]
			: [])
	]);
</script>

<svelte:head>
	<title>Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1>Bonjour {boot.user?.display_name}</h1>
		<p class="lede">
			{#if boot.organization}
				Organisation active : <strong>{boot.organization.name}</strong>.
			{/if}
			Lancez une revue depuis un sujet, suivez vos points, publiez de nouvelles versions de modèles.
		</p>
	</header>

	<ul class="card-list">
		{#each tiles as tile (tile.href)}
			<li>
				<a href={tile.href}>
					<strong>{tile.title}</strong>
					<span class="desc">{tile.desc}</span>
				</a>
			</li>
		{/each}
	</ul>
</div>
