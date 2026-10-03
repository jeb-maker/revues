/**
 * miniature-broccoli (mb) — tokens CSS + custom elements for SvelteKit.
 *
 * Assets stay under `web/static/vendor/jeb-maker-mb/` (embedded by Go at
 * `/static/...`). In Vite dev, `vite.config.ts` proxies `/static` → `:8080`.
 * Production SPA is same-origin with Go, so the same URLs work.
 *
 * `ensureMb()` is awaited by the root layout load so that every page renders
 * with the custom elements already defined (no FOUC, Svelte binds properties
 * instead of attributes). `app.html` preloads the same assets.
 */

export const MB_VENDOR_BASE = '/static/vendor/jeb-maker-mb';

const TOKEN_HREF = `${MB_VENDOR_BASE}/tokens/tokens-core.css`;
const BRIDGE_HREF = `${MB_VENDOR_BASE}/mb-bridge.css`;
const BOOT_SRC = `${MB_VENDOR_BASE}/mb-boot.js`;

let pending: Promise<void> | null = null;

/** Inject mb tokens + bridge CSS and load the CE bundle (idempotent). */
export function ensureMb(): Promise<void> {
	if (typeof document === 'undefined') return Promise.resolve();
	pending ??= loadMb().catch((e: unknown) => {
		pending = null;
		throw e;
	});
	return pending;
}

async function loadMb(): Promise<void> {
	ensureStylesheet(TOKEN_HREF);
	ensureStylesheet(BRIDGE_HREF);

	if (customElements.get('mb-button')) return;
	if (!document.querySelector(`script[src="${BOOT_SRC}"]`)) {
		await new Promise<void>((resolve, reject) => {
			const s = document.createElement('script');
			s.type = 'module';
			s.src = BOOT_SRC;
			s.onload = () => resolve();
			s.onerror = () => reject(new Error(`failed to load ${BOOT_SRC}`));
			document.head.appendChild(s);
		});
	}
	await customElements.whenDefined('mb-button');
}

function ensureStylesheet(href: string): void {
	if (document.querySelector(`link[rel="stylesheet"][href="${href}"]`)) return;
	const link = document.createElement('link');
	link.rel = 'stylesheet';
	link.href = href;
	document.head.appendChild(link);
}

/** Valeur courante d'un `mb-input` / `mb-textarea` depuis son événement `input`. */
export function inputValue(e: Event): string {
	return (e.target as HTMLInputElement).value;
}
