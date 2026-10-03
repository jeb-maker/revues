// See https://svelte.dev/docs/kit/types#app.d.ts
/// <reference path="./lib/mb/elements.d.ts" />

import type { Session } from '$lib/auth/session';

declare global {
	namespace App {
		interface PageData {
			/** Session partagée chargée par `routes/+layout.ts`. */
			boot: Session;
			/** Renseigné quand l'API ne répond pas (bootstrap en échec). */
			apiError?: string;
		}
	}
}

export {};
