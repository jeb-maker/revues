<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import AdminNav, { adminNavSection } from '$lib/components/AdminNav.svelte';

	let { children }: { children: Snippet } = $props();

	const section = $derived(adminNavSection(page.url.pathname));
</script>

<div class="page page--admin">
	<AdminNav {section} />
	{@render children()}
</div>
