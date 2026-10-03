<script lang="ts">
	import AdminNav from '$lib/components/AdminNav.svelte';

	const sections = [
		{ href: '/admin/users', title: 'Emails autorisés', desc: 'Whitelist de connexion et rôles lecteur / éditeur.' },
		{ href: '/admin/members', title: 'Membres', desc: 'Rôles dans l’organisation : propriétaire, administrateur, membre.' },
		{ href: '/admin/teams', title: 'Équipes', desc: 'Composition des équipes.' },
		{ href: '/admin/settings/policies', title: 'Politiques', desc: 'Délégation aux référents de sujet.' },
		{ href: '/admin/integrations', title: 'Intégrations', desc: 'Jira, Notion, webhooks sortants.' },
		{ href: '/admin/settings/smtp', title: 'SMTP', desc: 'Relais d’envoi des notifications.' }
	];
</script>

<svelte:head>
	<title>Administration — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1>Administration</h1>
		<p class="lede">Whitelist, rôles, équipes, politiques et intégrations de l’organisation active.</p>
	</header>

	<AdminNav section="hub">
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
	</AdminNav>
</div>
