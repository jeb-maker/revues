<script lang="ts">
	import { onMount } from 'svelte';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { requireAdminSession } from '$lib/admin/session';

	let ready = $state(false);
	let error = $state('');

	onMount(async () => {
		try {
			const gate = await requireAdminSession();
			if (!gate) return;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chargement impossible.';
		} finally {
			ready = true;
		}
	});
</script>

<svelte:head>
	<title>Administration — Revues</title>
</svelte:head>

<main class="admin">
	<p class="brand">Revues</p>
	<h1>Administration</h1>
	<p class="lede">Gérez la whitelist, les rôles, les équipes et les politiques de l'organisation active.</p>

	{#if !ready}
		<p class="muted">Chargement…</p>
	{:else if error}
		<mb-alert variant="danger" role="alert">{error}</mb-alert>
	{:else}
		<AdminNav section="hub">
			<ul class="cards">
				<li><a href="/admin/users">Emails autorisés</a> — whitelist login + rôles reader/editor</li>
				<li><a href="/admin/members">Membres</a> — rôles org owner / admin / member</li>
				<li><a href="/admin/teams">Équipes</a> — composition des équipes</li>
				<li><a href="/admin/settings/policies">Politiques</a> — délégation aux leads</li>
			</ul>
		</AdminNav>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background:
			radial-gradient(ellipse 80% 50% at 10% 0%, rgba(15, 118, 110, 0.28), transparent 55%),
			linear-gradient(165deg, #0f172a 0%, #1e293b 55%, #134e4a 100%);
		color: #f8fafc;
	}
	.admin {
		max-width: 42rem;
		margin: 0 auto;
		padding: 8vh 1.25rem 3rem;
	}
	.brand {
		margin: 0 0 0.35rem;
		font-size: clamp(2rem, 6vw, 2.75rem);
		font-weight: 700;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	h1 {
		margin: 0 0 0.45rem;
		font-size: 1.15rem;
		font-weight: 500;
		color: #99f6e4;
	}
	.lede {
		margin: 0 0 1.5rem;
		color: #cbd5e1;
		line-height: 1.45;
	}
	.cards {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
	}
	.cards a {
		color: #5eead4;
		font-weight: 600;
	}
	.muted {
		color: #94a3b8;
	}
	mb-alert {
		display: block;
		margin-bottom: 1rem;
	}
</style>
