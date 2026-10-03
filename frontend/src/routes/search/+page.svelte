<script lang="ts">
	import { page } from '$app/state';
	import { globalSearch, type SearchResult, type SearchResultKind } from '$lib/api/search';
	import { session } from '$lib/auth/session';
	import { runLabels, subjectLabels, templateNavLabel } from '$lib/i18n/uiLabels';

	const boot = session();
	const subject = $derived(subjectLabels(boot.organization?.ui_subject_label));
	const run = $derived(runLabels(boot.organization?.ui_run_label));
	const templatesLabel = $derived(templateNavLabel(boot.organization?.ui_run_label));

	const q = $derived((page.url.searchParams.get('q') ?? '').trim());

	let results = $state<SearchResult[]>([]);
	let error = $state('');
	let loading = $state(false);

	const kindOrder: SearchResultKind[] = ['subject', 'run', 'template', 'task'];

	function kindLabel(kind: SearchResultKind): string {
		switch (kind) {
			case 'subject':
				return subject.plural;
			case 'run':
				return run.nav;
			case 'template':
				return templatesLabel;
			case 'task':
				return 'Mes tâches';
			default:
				return kind;
		}
	}

	const groups = $derived.by(() => {
		const map = new Map<SearchResultKind, SearchResult[]>();
		for (const kind of kindOrder) map.set(kind, []);
		for (const item of results) {
			const list = map.get(item.kind) ?? [];
			list.push(item);
			map.set(item.kind, list);
		}
		return kindOrder
			.map((kind) => ({ kind, label: kindLabel(kind), items: map.get(kind) ?? [] }))
			.filter((g) => g.items.length > 0);
	});

	$effect(() => {
		const query = q;
		if (!query) {
			results = [];
			error = '';
			loading = false;
			return;
		}
		let cancelled = false;
		loading = true;
		error = '';
		void globalSearch({ q: query, csrfToken: boot.csrf_token })
			.then((res) => {
				if (cancelled) return;
				results = res.results ?? [];
			})
			.catch((e) => {
				if (cancelled) return;
				error = e instanceof Error ? e.message : 'Recherche impossible.';
				results = [];
			})
			.finally(() => {
				if (!cancelled) loading = false;
			});
		return () => {
			cancelled = true;
		};
	});
</script>

<svelte:head>
	<title>Recherche — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1>Recherche</h1>
		{#if q}
			<p class="lede">Résultats pour « {q} ».</p>
		{:else}
			<p class="lede">Saisissez un terme dans la barre du haut, puis validez.</p>
		{/if}
	</header>

	{#if !q}
		<mb-empty-state heading="Aucune requête">
			Utilisez le champ Recherche du bandeau pour lancer une recherche.
		</mb-empty-state>
	{:else if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{:else if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Recherche…</p>
	{:else if groups.length === 0}
		<mb-empty-state heading="Aucun résultat">Aucun élément ne correspond à « {q} ».</mb-empty-state>
	{:else}
		{#each groups as group (group.kind)}
			<section class="search-group" aria-labelledby={`search-${group.kind}`}>
				<h2 id={`search-${group.kind}`}>{group.label}</h2>
				<ul class="search-results">
					{#each group.items as item (item.kind + '-' + item.id)}
						<li>
							<a href={item.href}>
								<span class="search-results__title">{item.title}</span>
								{#if item.subtitle}
									<span class="muted">{item.subtitle}</span>
								{/if}
							</a>
						</li>
					{/each}
				</ul>
			</section>
		{/each}
	{/if}
</div>

<style>
	.search-group {
		margin-bottom: var(--mb-space-5);
	}

	.search-group h2 {
		margin-bottom: var(--mb-space-3);
	}

	.search-results {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.search-results li + li {
		margin-top: var(--mb-space-2);
	}

	.search-results a {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		text-decoration: none;
		color: inherit;
	}

	.search-results a:hover .search-results__title {
		color: var(--mb-color-accent);
	}

	.search-results__title {
		font-weight: 600;
		color: var(--mb-color-accent);
	}
</style>
