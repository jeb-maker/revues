<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { globalSearch, type SearchResult, type SearchResultKind } from '$lib/api/search';
	import { runLabels, subjectLabels, templateNavLabel } from '$lib/i18n/uiLabels';

	type ComboboxOption = {
		value: string;
		label: string;
		group?: string;
		href?: string;
	};

	type MbSelectDetail = { value: string; label: string; href?: string };

	type MbComboboxEl = HTMLElement & {
		value: string;
		options: ComboboxOption[];
		loading: boolean;
		open: boolean;
	};

	let {
		csrfToken,
		uiRunLabel,
		uiSubjectLabel
	}: {
		csrfToken?: string;
		uiRunLabel?: string | null;
		uiSubjectLabel?: string | null;
	} = $props();

	const DEBOUNCE_MS = 280;
	const SUGGEST_LIMIT = 12;

	let query = $state('');
	let options = $state<ComboboxOption[]>([]);
	let loading = $state(false);
	let combo: MbComboboxEl | undefined = $state();
	let debounceTimer = 0;
	let requestSeq = 0;

	const subject = $derived(subjectLabels(uiSubjectLabel));
	const run = $derived(runLabels(uiRunLabel));
	const templatesLabel = $derived(templateNavLabel(uiRunLabel));

	$effect(() => {
		if (page.url.pathname === '/search') {
			query = page.url.searchParams.get('q') ?? '';
		}
	});

	$effect(() => {
		return () => {
			window.clearTimeout(debounceTimer);
		};
	});

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

	function toOptions(results: SearchResult[], q: string): ComboboxOption[] {
		const items: ComboboxOption[] = results.map((item) => ({
			value: `${item.kind}:${item.id}`,
			label: item.subtitle ? `${item.title} — ${item.subtitle}` : item.title,
			group: kindLabel(item.kind),
			href: item.href
		}));
		if (!q) return items;
		items.push({
			value: '__all__',
			label: `Voir tous les résultats pour « ${q} »`,
			href: `/search?q=${encodeURIComponent(q)}`
		});
		return items;
	}

	function scheduleSearch(raw: string) {
		window.clearTimeout(debounceTimer);
		const q = raw.trim();
		if (!q) {
			requestSeq += 1;
			options = [];
			loading = false;
			if (combo) combo.open = false;
			return;
		}
		loading = true;
		if (combo) combo.open = true;
		const seq = ++requestSeq;
		debounceTimer = window.setTimeout(() => {
			void globalSearch({ q, limit: SUGGEST_LIMIT, csrfToken })
				.then((res) => {
					if (seq !== requestSeq) return;
					options = toOptions(res.results ?? [], q);
					loading = false;
					if (combo) combo.open = true;
				})
				.catch(() => {
					if (seq !== requestSeq) return;
					options = [
						{
							value: '__all__',
							label: `Ouvrir la recherche pour « ${q} »`,
							href: `/search?q=${encodeURIComponent(q)}`
						}
					];
					loading = false;
					if (combo) combo.open = true;
				});
		}, DEBOUNCE_MS);
	}

	function onInput(e: CustomEvent<{ value: string }>) {
		query = e.detail.value;
		scheduleSearch(query);
	}

	async function onSelect(e: CustomEvent<MbSelectDetail>) {
		const href = e.detail.href?.trim();
		if (!href) return;
		await goto(href, { keepFocus: false });
	}

	async function onSubmit(e: Event) {
		e.preventDefault();
		const q = query.trim();
		window.clearTimeout(debounceTimer);
		if (combo) combo.open = false;
		await goto(q ? `/search?q=${encodeURIComponent(q)}` : '/search', { keepFocus: true });
	}
</script>

<form class="app-header__search" role="search" onsubmit={onSubmit}>
	<mb-combobox
		bind:this={combo}
		label="Recherche"
		hide-label
		density="compact"
		type="search"
		name="q"
		placeholder="Rechercher…"
		empty-message="Aucun résultat"
		loading-message="Recherche…"
		value={query}
		{options}
		{loading}
		onmb-input={onInput}
		onmb-select={onSelect}
	></mb-combobox>
</form>
