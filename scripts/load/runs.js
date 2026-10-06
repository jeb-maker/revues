/**
 * Scénario dédié : lancement de revues (POST /subjects/{id}/runs).
 * Sollicite le snapshot transactionnel SQLite — le chemin write le plus lourd.
 *
 *   ./scripts/load/run.sh runs
 *   VUS=20 DURATION=1m ./scripts/load/run.sh runs
 */
import { sleep } from 'k6';
import {
  ensureAuth,
  healthCheck,
  prepareSessions,
  launchRun,
  defaultThresholds,
} from './lib.js';

const VUS = Number(__ENV.VUS || 15);
const DURATION = __ENV.DURATION || '30s';

export const options = {
  scenarios: {
    runs: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: Math.max(1, Math.floor(VUS / 2)) },
        { duration: DURATION, target: VUS },
        { duration: '10s', target: 0 },
      ],
      gracefulRampDown: '5s',
    },
  },
  thresholds: {
    ...defaultThresholds(),
    http_req_duration: ['p(95)<2000'],
  },
};

export function setup() {
  healthCheck();
  return { sessions: prepareSessions(VUS) };
}

export default function (data) {
  ensureAuth(data);
  launchRun();
  sleep(0.4 + Math.random() * 0.8);
}
