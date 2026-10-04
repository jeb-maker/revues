/**
 * Presets org `ui_run_label` / `ui_subject_label` (colonnes organizations).
 * Pas de marque dans le chrome : le hub runs est un onglet nav (preset `ui_run_label`).
 * Défaut produit conteneur = projet (code API : subjects).
 */

export type RunLabelPreset = 'revues' | 'listes_en_cours' | 'audits' | 'checklists';
export type SubjectLabelPreset = 'projet' | 'sujet' | 'cible' | 'entite' | 'asset';

export type RunUILabels = {
	preset: RunLabelPreset;
	/** Nav principale + H1 liste. */
	nav: string;
	/** Raccourci mobile éventuel. */
	navShort: string;
	singular: string;
	plural: string;
	article: string; // « une » / « un »
	noneArticle: string; // « Aucune » / « Aucun »
};

export type SubjectUILabels = {
	preset: SubjectLabelPreset;
	singular: string;
	plural: string;
};

const RUN_PRESETS: Record<RunLabelPreset, Omit<RunUILabels, 'preset'>> = {
	revues: {
		nav: 'Revues',
		navShort: 'Revues',
		singular: 'revue',
		plural: 'revues',
		article: 'une',
		noneArticle: 'Aucune'
	},
	listes_en_cours: {
		nav: 'Listes en cours',
		navShort: 'En cours',
		singular: 'liste en cours',
		plural: 'listes en cours',
		article: 'une',
		noneArticle: 'Aucune'
	},
	audits: {
		nav: 'Audits',
		navShort: 'Audits',
		singular: 'audit',
		plural: 'audits',
		article: 'un',
		noneArticle: 'Aucun'
	},
	checklists: {
		nav: 'Checklists',
		navShort: 'Checklists',
		singular: 'checklist',
		plural: 'checklists',
		article: 'une',
		noneArticle: 'Aucune'
	}
};

const SUBJECT_PRESETS: Record<SubjectLabelPreset, Omit<SubjectUILabels, 'preset'>> = {
	projet: { singular: 'Projet', plural: 'Projets' },
	sujet: { singular: 'Sujet', plural: 'Sujets' },
	cible: { singular: 'Cible', plural: 'Cibles' },
	entite: { singular: 'Entité', plural: 'Entités' },
	asset: { singular: 'Actif', plural: 'Actifs' }
};

function asRunPreset(raw: string | null | undefined): RunLabelPreset {
	const v = (raw ?? '').trim().toLowerCase();
	if (v in RUN_PRESETS) return v as RunLabelPreset;
	return 'revues';
}

function asSubjectPreset(raw: string | null | undefined): SubjectLabelPreset {
	const v = (raw ?? '').trim().toLowerCase();
	if (v in SUBJECT_PRESETS) return v as SubjectLabelPreset;
	return 'projet';
}

export function runLabels(preset?: string | null): RunUILabels {
	const p = asRunPreset(preset);
	return { preset: p, ...RUN_PRESETS[p] };
}

export function subjectLabels(preset?: string | null): SubjectUILabels {
	const p = asSubjectPreset(preset);
	return { preset: p, ...SUBJECT_PRESETS[p] };
}

/**
 * Vocabulaire Listes / Modèles (legacy listUI).
 * Sans flag `ShowSubjectColumn` côté API, on aligne le particulier seed
 * (`ui_run_label=listes_en_cours`) sur « Listes ».
 */
export function templateNavLabel(runPreset?: string | null): string {
	return asRunPreset(runPreset) === 'listes_en_cours' ? 'Listes' : 'Modèles';
}

/** CTA primaire listes / fiches : court — le H1 ou le contexte porte déjà le type. */
export function launchRunCTA(_run: RunUILabels): string {
	return 'Lancer';
}
