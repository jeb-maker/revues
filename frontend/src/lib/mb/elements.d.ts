/**
 * Ambient typings for mb custom elements used in Svelte markup.
 * Keep in sync with `web/static/vendor/jeb-maker-mb/README.md` (mb 0.4.4).
 *
 * Events: native `input` crosses the shadow DOM (composed) and `e.target` is the
 * host whose `.value` is already synced ; `change` does not, so `mb-select`,
 * `mb-checkbox`, `mb-radio-group` expose `mb-change` (`detail.value` / `detail.checked`).
 */
declare module 'svelte/elements' {
	type MbChangeDetail = { value: string; checked?: boolean; files?: FileList | null };

	interface MbAttrs {
		variant?: string;
		size?: string;
		type?: string;
		name?: string;
		value?: string;
		placeholder?: string;
		disabled?: boolean | string | null;
		required?: boolean | string | null;
		checked?: boolean | string | null;
		loading?: boolean | string | null;
		open?: boolean | string | null;
		href?: string;
		target?: string;
		rel?: string;
		role?: string;
		min?: number | string;
		max?: number | string;
		step?: number | string;
		minlength?: number | string;
		maxlength?: number | string;
		pattern?: string;
		readonly?: boolean | string | null;
		rows?: number | string;
		accept?: string;
		autocomplete?: string;
		label?: string;
		hint?: string;
		error?: string;
		heading?: string;
		'hide-label'?: boolean | string | null;
		'missing-message'?: string;
		'invalid-message'?: string;
		'close-label'?: string;
		'dismiss-label'?: string;
		'fallback-label'?: string;
		for?: string;
		'label-open'?: string;
		'label-close'?: string;
		percent?: number | string;
		columns?: string;
		align?: string;
		primary?: boolean | string | null;
		actions?: boolean | string | null;
		'sticky-header'?: boolean | string | null;
		oninput?: (e: Event) => void;
		onchange?: (e: Event) => void;
		onclick?: (e: MouseEvent) => void;
		'onmb-input'?: (e: CustomEvent<MbChangeDetail>) => void;
		'onmb-change'?: (e: CustomEvent<MbChangeDetail>) => void;
		'onmb-toggle'?: (e: CustomEvent<{ expanded: boolean }>) => void;
		class?: string;
		id?: string;
		slot?: string;
		title?: string;
		'aria-label'?: string;
		'aria-current'?: string;
	}

	export interface SvelteHTMLElements {
		'mb-alert': MbAttrs;
		'mb-avatar': MbAttrs;
		'mb-badge': MbAttrs;
		'mb-button': MbAttrs;
		'mb-card': MbAttrs;
		'mb-checkbox': MbAttrs;
		'mb-empty-state': MbAttrs;
		'mb-input': MbAttrs;
		'mb-modal': MbAttrs;
		'mb-nav': MbAttrs;
		'mb-nav-toggle': MbAttrs;
		'mb-pagination': MbAttrs;
		'mb-progress': MbAttrs;
		'mb-radio': MbAttrs;
		'mb-radio-group': MbAttrs;
		'mb-select': MbAttrs;
		'mb-spinner': MbAttrs;
		'mb-table': MbAttrs;
		'mb-table-cell': MbAttrs;
		'mb-table-row': MbAttrs;
		'mb-tag': MbAttrs;
		'mb-textarea': MbAttrs;
		'mb-toast': MbAttrs;
		'mb-toolbar': MbAttrs;
		'mb-segmented-control': MbAttrs;
		'mb-breadcrumbs': MbAttrs;
	}
}

export {};
