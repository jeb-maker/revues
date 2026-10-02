// See https://svelte.dev/docs/kit/types#app.d.ts

declare namespace svelteHTML {
	// Allow mb-* custom elements in Svelte templates.
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	interface IntrinsicElements {
		'mb-button': any;
		'mb-input': any;
		'mb-alert': any;
	}
}

declare global {
	namespace App {}
}

export {};
