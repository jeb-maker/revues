<script lang="ts">
	import { session } from '$lib/auth/session';
	import { launchRunCTA, runLabels, subjectLabels, templateNavLabel } from '$lib/i18n/uiLabels';

	const boot = session();
	const run = $derived(runLabels(boot.organization?.ui_run_label));
	const subject = $derived(subjectLabels(boot.organization?.ui_subject_label));
	const templatesLabel = $derived(templateNavLabel(boot.organization?.ui_run_label));

	const tiles = $derived([
		{
			href: '/runs',
			title: run.nav,
			desc: `Exécutions en cours et historiques — ${launchRunCTA(run).toLowerCase()} depuis un ${subject.singular.toLowerCase()}.`
		},
		{
			href: '/subjects',
			title: subject.plural,
			desc: `Conteneurs de ${run.plural} : domaines, membres.`
		},
		{
			href: '/modeles',
			title: templatesLabel,
			desc: 'Check-lists versionnées de l’organisation.'
		},
		...(boot.show_my_tasks
			? [
					{
						href: '/mes-taches',
						title: 'Mes tâches',
						desc: `Points de ${run.singular} qui vous sont assignés.`
					}
				]
			: []),
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
				Organisation : <strong>{boot.organization.name}</strong>.
			{/if}
			{launchRunCTA(run)} depuis un {subject.singular.toLowerCase()}, suivez vos points, publiez de
			nouvelles versions de {templatesLabel.toLowerCase()}.
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
