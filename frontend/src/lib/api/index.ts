/**
 * Front API surface — generated OpenAPI types + thin wrappers.
 * Domain helpers (auth, orgs, …) live alongside as non-generated modules until
 * the matching operations land in `api/openapi/openapi.yaml`.
 */
export {
	API_V1_BASE,
	api,
	createApiClient,
	getHealth,
	type HealthResponse,
	type paths
} from './client';
