/**
 * Scénario mixte :
 *   ~70 % lecture
 *   ~20 % écriture légère (PATCH item / POST subject)
 *   ~10 % lancement de revue (POST …/runs — snapshot SQL)
 *
 *   ./scripts/load/run.sh mix
 *   VUS=30 DURATION=1m WRITE_RATIO=0.2 RUN_RATIO=0.1 ./scripts/load/run.sh mix
 */
import { sleep } from 'k6';
import {
  ensureAuth,
  healthCheck,
  prepareSessions,
  readJourney,
  writeMutation,
  launchRun,
  defaultThresholds,
} from './lib.js';

const VUS = Number(__ENV.VUS || 20);
const DURATION = __ENV.DURATION || '30s';
const WRITE_RATIO = Math.min(1, Math.max(0, Number(__ENV.WRITE_RATIO || 0.2)));
const RUN_RATIO = Math.min(1, Math.max(0, Number(__ENV.RUN_RATIO || 0.1)));

export const options = {
  scenarios: {
    mix: {
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
    // Snapshot + writes SQLite : p95 plus tolérant
    http_req_duration: ['p(95)<1500'],
  },
};

export function setup() {
  healthCheck();
  return { sessions: prepareSessions(VUS) };
}

export default function (data) {
  ensureAuth(data);
  const roll = Math.random();
  if (roll < RUN_RATIO) {
    launchRun();
  } else if (roll < RUN_RATIO + WRITE_RATIO) {
    writeMutation();
  } else {
    readJourney();
  }
  sleep(0.2 + Math.random() * 0.6);
}
