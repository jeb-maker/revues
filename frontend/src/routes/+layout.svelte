<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { ensureMb } from '$lib/mb';

	let { children }: { children: Snippet } = $props();

	onMount(() => {
		void ensureMb();
	});
</script>

{@render children()}
