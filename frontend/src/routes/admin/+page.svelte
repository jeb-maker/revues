<script lang="ts">
	const sections = [
		{ href: '/admin/members', title: 'Membres', desc: 'Inviter et gérer les rôles : propriétaire, administrateur, membre.' },
		{ href: '/admin/settings/policies', title: 'Politiques', desc: 'Délégation aux référents de projet.' },
		{ href: '/admin/integrations', title: 'Intégrations', desc: 'Jira et webhooks sortants.' },
		{ href: '/admin/settings/smtp', title: 'SMTP', desc: 'Relais d’envoi des notifications.' }
	];
</script>

<svelte:head>
	<title>Administration — Revues</title>
</svelte:head>

	<header class="page-header">
		<h1>Administration</h1>
		<p class="lede">Membres, invitations, politiques et intégrations de cette organisation.</p>
	</header>

	<ul class="card-list">
		{#each sections as s (s.href)}
			<li>
				<a href={s.href}>
					<strong>{s.title}</strong>
					<span class="desc">{s.desc}</span>
				</a>
			</li>
		{/each}
	</ul>
