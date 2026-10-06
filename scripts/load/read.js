/**
 * Scénario lecture seule — capacité navigation typique.
 *
 *   ./scripts/load/run.sh read
 *   VUS=50 DURATION=1m ./scripts/load/run.sh read
 */
import { sleep } from 'k6';
import {
  ensureAuth,
  healthCheck,
  prepareSessions,
  readJourney,
  defaultThresholds,
} from './lib.js';

const VUS = Number(__ENV.VUS || 20);
const DURATION = __ENV.DURATION || '30s';

export const options = {
  scenarios: {
    read: {
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
  thresholds: defaultThresholds(),
};

export function setup() {
  healthCheck();
  return { sessions: prepareSessions(VUS) };
}

export default function (data) {
  ensureAuth(data);
  readJourney();
  sleep(0.3 + Math.random() * 0.7);
}
