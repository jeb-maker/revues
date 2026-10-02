/**
 * miniature-broccoli (mb) — tokens CSS + custom elements for SvelteKit.
 *
 * Assets stay under `web/static/vendor/jeb-maker-mb/` (embedded by Go at
 * `/static/...`). In Vite dev, `vite.config.ts` proxies `/static` → `:8080`.
 * Production SPA is same-origin with Go, so the same URLs work.
 *
 * Call `ensureMb()` once from the root layout (or import side-effect).
 */

export const MB_VENDOR_BASE = '/static/vendor/jeb-maker-mb';

const TOKEN_HREF = `${MB_VENDOR_BASE}/tokens/tokens-core.css`;
const BRIDGE_HREF = `${MB_VENDOR_BASE}/mb-bridge.css`;
const BOOT_SRC = `${MB_VENDOR_BASE}/mb-boot.js`;

let loaded = false;

/** Inject mb tokens + bridge CSS and load CE bundle (idempotent). */
export async function ensureMb(): Promise<void> {
	if (typeof document === 'undefined') return;
	if (loaded) return;
	loaded = true;

	ensureStylesheet(TOKEN_HREF);
	ensureStylesheet(BRIDGE_HREF);

	if (!document.querySelector(`script[data-mb-boot]`)) {
		await new Promise<void>((resolve, reject) => {
			const s = document.createElement('script');
			s.type = 'module';
			s.src = BOOT_SRC;
			s.dataset.mbBoot = '1';
			s.onload = () => resolve();
			s.onerror = () => reject(new Error(`failed to load ${BOOT_SRC}`));
			document.head.appendChild(s);
		});
	}
}

function ensureStylesheet(href: string): void {
	if (document.querySelector(`link[href="${href}"]`)) return;
	const link = document.createElement('link');
	link.rel = 'stylesheet';
	link.href = href;
	document.head.appendChild(link);
}
