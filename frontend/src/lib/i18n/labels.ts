/**
 * Libellés FR des codes métier renvoyés par l'API (`schema.d.ts`, docs/RBAC.md).
 * Les codes inconnus sont renvoyés tels quels pour ne jamais masquer une valeur.
 */

export type ItemStatus = 'pending' | 'ok' | 'nok' | 'na';
export type RunStatus = 'draft' | 'in_progress' | 'done' | 'overdue' | 'archived';
export type BadgeVariant = 'neutral' | 'success' | 'danger' | 'warning' | 'info';

const ITEM_STATUS: Record<ItemStatus, string> = {
	pending: 'En attente',
	ok: 'Validé',
	nok: 'Non validé',
	na: 'Non applicable'
};

const RUN_STATUS: Record<RunStatus, string> = {
	draft: 'Brouillon',
	in_progress: 'En cours',
	done: 'Terminée',
	overdue: 'En retard',
	archived: 'Archivée'
};

/** Rôles globaux (users), org (members) et sujet (subject_members). */
const ROLE: Record<string, string> = {
	owner: 'Propriétaire',
	admin: 'Administrateur',
	member: 'Membre',
	editor: 'Éditeur',
	reader: 'Lecteur',
	lead: 'Référent',
	contributor: 'Contributeur',
	viewer: 'Observateur'
};

const VISIBILITY: Record<string, string> = {
	normal: 'Normal',
	private: 'Privé'
};

const DELIVERY_STATE: Record<string, string> = {
	pending: 'En attente',
	done: 'Livrée',
	poison: 'Abandonnée'
};

export function formatItemStatus(status: string): string {
	return ITEM_STATUS[status as ItemStatus] ?? status;
}

export function formatRunStatus(status: string): string {
	return RUN_STATUS[status as RunStatus] ?? status;
}

export function formatRole(role: string): string {
	return ROLE[role] ?? role;
}

export function formatVisibility(visibility: string): string {
	return VISIBILITY[visibility] ?? visibility;
}

export function formatDeliveryState(state: string): string {
	return DELIVERY_STATE[state] ?? state;
}

export function itemStatusVariant(status: string): BadgeVariant {
	switch (status) {
		case 'ok':
			return 'success';
		case 'nok':
			return 'danger';
		case 'na':
			return 'neutral';
		default:
			return 'warning';
	}
}

export function runStatusVariant(status: string): BadgeVariant {
	switch (status) {
		case 'done':
			return 'success';
		case 'overdue':
			return 'danger';
		case 'in_progress':
			return 'info';
		default:
			return 'neutral';
	}
}

/** Options `<mb-select>` / `<option>` typées pour un ensemble de codes. */
export function roleOptions<T extends string>(roles: readonly T[]): { value: T; label: string }[] {
	return roles.map((value) => ({ value, label: formatRole(value) }));
}
