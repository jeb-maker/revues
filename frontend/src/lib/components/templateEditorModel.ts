/** Shared helpers for template editor rows. */
export type EditorItem = {
	key: string;
	section: string;
	label: string;
	help_text: string;
	required: boolean;
};

export function newEditorKey(): string {
	return `r-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export function emptyEditorItem(section = ''): EditorItem {
	return { key: newEditorKey(), section, label: '', help_text: '', required: false };
}

export function itemsFromDetail(
	items: Array<{ section?: string; label: string; help_text?: string; required?: boolean }>
): EditorItem[] {
	if (!items.length) return [emptyEditorItem()];
	return items.map((it) => ({
		key: newEditorKey(),
		section: it.section ?? '',
		label: it.label,
		help_text: it.help_text ?? '',
		required: Boolean(it.required)
	}));
}

export function toWriteItems(items: EditorItem[]) {
	return items
		.filter((it) => it.label.trim() !== '')
		.map((it) => ({
			section: it.section.trim(),
			label: it.label.trim(),
			help_text: it.help_text.trim(),
			required: it.required
		}));
}

/** Aligné sur IsStructuralItemChange côté Go (help_text ignoré). */
export function isStructuralItemChange(
	baseline: Array<{ section: string; label: string; required: boolean }>,
	next: Array<{ section: string; label: string; required: boolean }>
): boolean {
	if (baseline.length !== next.length) return true;
	for (let i = 0; i < baseline.length; i++) {
		if (
			baseline[i].section !== next[i].section ||
			baseline[i].label !== next[i].label ||
			baseline[i].required !== next[i].required
		) {
			return true;
		}
	}
	return false;
}
