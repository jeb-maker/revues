/**
 * Soft-redirect legacy list `?q=` toward the global search page.
 * Documented decision for issue #290 — list APIs keep `q`, UI does not.
 */
export function searchHrefFromListQuery(q: string | null | undefined): string | null {
	const trimmed = (q ?? '').trim();
	if (!trimmed) return null;
	return `/search?q=${encodeURIComponent(trimmed)}`;
}
