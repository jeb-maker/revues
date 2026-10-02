/**
 * Ambient typings for mb custom elements used in Svelte markup.
 * Keep in sync with `web/static/vendor/jeb-maker-mb/README.md`.
 */
declare module 'svelte/elements' {
	interface MbAttrs {
		variant?: string;
		type?: string;
		name?: string;
		value?: string;
		placeholder?: string;
		disabled?: boolean | string | null;
		required?: boolean | string | null;
		checked?: boolean | string | null;
		open?: boolean | string | null;
		href?: string;
		role?: string;
		max?: number | string;
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
