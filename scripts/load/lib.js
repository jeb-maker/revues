/**
 * Helpers partagés pour les scénarios k6 Revues.
 *
 * Env :
 *   BASE_URL, LOAD_PASSWORD, LOAD_EMAIL_PREFIX, SESSION_POOL, VUS
 *
 * Auth : sessions créées en setup() (séquentiel) pour éviter le rate-limit
 * login/register (30/min). Désactiver REVUES_DEV_AUTH pour un run concurrent propre.
 */
import http from 'k6/http';
import { check, fail, sleep } from 'k6';

export const BASE_URL = (__ENV.BASE_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');
export const LOAD_PASSWORD = __ENV.LOAD_PASSWORD || 'loadtest-pass-12';
export const LOAD_EMAIL_PREFIX = __ENV.LOAD_EMAIL_PREFIX || 'loadtest-vu';

const vuAuth = {
  csrf: '',
  ready: false,
  email: '',
  session: '',
};

const okStatuses = http.expectedStatuses(200, 201, 403, 409);

export function apiURL(path) {
  return `${BASE_URL}/api/v1${path}`;
}

export function jsonHeaders(csrf) {
  const h = { 'Content-Type': 'application/json', Accept: 'application/json' };
  if (csrf) {
    h['X-CSRF-Token'] = csrf;
  }
  return h;
}

function parseJSON(res, label) {
  try {
    return res.json();
  } catch (e) {
    fail(`${label}: body JSON invalide (status ${res.status}): ${String(res.body).slice(0, 200)}`);
  }
}

function cookieValue(jar, name) {
  const cookies = jar.cookiesForURL(BASE_URL) || {};
  const values = cookies[name];
  if (Array.isArray(values) && values.length > 0) {
    return values[0];
  }
  return '';
}

function clearJar() {
  http.cookieJar().clear(BASE_URL);
}

function authenticateAccount(email) {
  for (let attempt = 0; attempt < 8; attempt++) {
    clearJar();
    const boot = http.get(apiURL('/bootstrap'), {
      headers: { Accept: 'application/json' },
      tags: { name: 'setup bootstrap' },
    });
    if (boot.status !== 200) {
      fail(`setup bootstrap ${boot.status}`);
    }
    const bootBody = parseJSON(boot, 'setup bootstrap');

    if (bootBody.authenticated && bootBody.csrf_token) {
      return finalizeSetupSession(email, bootBody.csrf_token, true);
    }

    const guestCSRF = bootBody.csrf_token;
    const reg = http.post(
      apiURL('/auth/register'),
      JSON.stringify({
        email,
        password: LOAD_PASSWORD,
        password_confirm: LOAD_PASSWORD,
      }),
      {
        headers: jsonHeaders(guestCSRF),
        tags: { name: 'setup register' },
        responseCallback: http.expectedStatuses(200, 201, 400, 403, 429),
      },
    );

    if (reg.status === 200 || reg.status === 201) {
      const body = parseJSON(reg, 'setup register');
      return finalizeSetupSession(email, body.csrf_token, false);
    }

    if (reg.status === 429) {
      console.warn(`rate-limit register, pause 65s (tentative ${attempt + 1})…`);
      sleep(65);
      continue;
    }

    clearJar();
    const boot2 = http.get(apiURL('/bootstrap'), {
      headers: { Accept: 'application/json' },
      tags: { name: 'setup bootstrap2' },
    });
    const boot2Body = parseJSON(boot2, 'setup bootstrap2');
    if (boot2Body.authenticated && boot2Body.csrf_token) {
      return finalizeSetupSession(email, boot2Body.csrf_token, true);
    }

    const login = http.post(
      apiURL('/auth/login'),
      JSON.stringify({ email, password: LOAD_PASSWORD }),
      {
        headers: jsonHeaders(boot2Body.csrf_token),
        tags: { name: 'setup login' },
        responseCallback: http.expectedStatuses(200, 401, 403, 429),
      },
    );

    if (login.status === 200) {
      const body = parseJSON(login, 'setup login');
      return finalizeSetupSession(email, body.csrf_token, false);
    }

    if (login.status === 429) {
      console.warn(`rate-limit login, pause 65s (tentative ${attempt + 1})…`);
      sleep(65);
      continue;
    }

    fail(`setup auth échoué (${login.status}) pour ${email}: ${String(login.body).slice(0, 300)}`);
  }
  fail(`setup auth abandonné après retries pour ${email}`);
}

function finalizeSetupSession(email, csrf, devAuth) {
  let session = cookieValue(http.cookieJar(), 'revues_session');
  if (!session) {
    fail(`setup sans cookie session (${email})`);
  }

  const orgs = http.get(apiURL('/orgs'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'setup orgs' },
  });
  if (orgs.status === 200) {
    const body = parseJSON(orgs, 'setup orgs');
    const list = body.organizations || [];
    if (list.length === 0 && body.can_create) {
      const slug = `load-${email.replace(/[^a-z0-9]+/gi, '-').replace(/^-|-$/g, '').slice(0, 40)}`;
      const created = http.post(
        apiURL('/orgs'),
        JSON.stringify({ name: `Load ${email}`, slug }),
        {
          headers: jsonHeaders(csrf),
          tags: { name: 'setup create org' },
          responseCallback: http.expectedStatuses(200, 201, 400, 403),
        },
      );
      if (created.status !== 200 && created.status !== 201) {
        fail(`setup create org (${created.status}): ${String(created.body).slice(0, 300)}`);
      }
      const refreshed = cookieValue(http.cookieJar(), 'revues_session');
      if (refreshed) {
        session = refreshed;
      }
      seedMinimalData(csrf);
      const afterSeed = cookieValue(http.cookieJar(), 'revues_session');
      if (afterSeed) {
        session = afterSeed;
      }
      const me = http.get(apiURL('/me'), { headers: { Accept: 'application/json' } });
      if (me.status === 200) {
        const meBody = parseJSON(me, 'setup me');
        if (meBody.csrf_token) {
          csrf = meBody.csrf_token;
        }
      }
    } else {
      // Org déjà là : s'assurer qu'il y a sujet + modèle pour lancer des revues.
      ensureLaunchFixtures(csrf);
      const me = http.get(apiURL('/me'), { headers: { Accept: 'application/json' } });
      if (me.status === 200) {
        const meBody = parseJSON(me, 'setup me');
        if (meBody.csrf_token) {
          csrf = meBody.csrf_token;
        }
      }
      const refreshed = cookieValue(http.cookieJar(), 'revues_session');
      if (refreshed) {
        session = refreshed;
      }
    }
  }

  return { email, session, csrf, devAuth };
}

function seedMinimalData(csrf) {
  const subject = http.post(
    apiURL('/subjects'),
    JSON.stringify({
      name: 'Load Subject',
      description: 'seed k6',
      domains: ['web'],
    }),
    {
      headers: jsonHeaders(csrf),
      tags: { name: 'setup seed subject' },
      responseCallback: http.expectedStatuses(200, 201, 403),
    },
  );
  if (subject.status !== 201 && subject.status !== 200) {
    console.warn(`seed subject skip (${subject.status})`);
    return;
  }
  const subjectBody = parseJSON(subject, 'seed subject');

  const tpl = http.post(
    apiURL('/templates'),
    JSON.stringify({
      name: 'Load Template',
      domains: ['web'],
      items: [
        { section: 'A', label: 'Point 1', required: false },
        { section: 'A', label: 'Point 2', required: false },
        { section: 'B', label: 'Point 3', required: false },
        { section: 'B', label: 'Point 4', required: false },
        { section: 'C', label: 'Point 5', required: false },
      ],
    }),
    {
      headers: jsonHeaders(csrf),
      tags: { name: 'setup seed template' },
      responseCallback: http.expectedStatuses(200, 201, 403),
    },
  );
  if (tpl.status !== 201 && tpl.status !== 200) {
    console.warn(`seed template skip (${tpl.status})`);
    return;
  }
  const tplBody = parseJSON(tpl, 'seed template');

  const run = http.post(
    apiURL(`/subjects/${subjectBody.id}/runs`),
    JSON.stringify({ template_id: tplBody.id }),
    {
      headers: jsonHeaders(csrf),
      tags: { name: 'setup seed run' },
      responseCallback: http.expectedStatuses(200, 201, 403),
    },
  );
  if (run.status !== 201 && run.status !== 200) {
    console.warn(`seed run skip (${run.status})`);
  }
}

/** Garantit au moins un sujet + un modèle lançable. */
function ensureLaunchFixtures(csrf) {
  const subjects = http.get(apiURL('/subjects'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'setup list subjects' },
  });
  if (subjects.status !== 200) {
    return;
  }
  const subjectList = parseJSON(subjects, 'setup subjects').subjects || [];
  const templates = http.get(apiURL('/templates'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'setup list templates' },
  });
  const templateList =
    templates.status === 200 ? parseJSON(templates, 'setup templates').templates || [] : [];

  if (subjectList.length > 0 && templateList.length > 0) {
    return;
  }
  seedMinimalData(csrf);
}

export function prepareSessions(vuCount) {
  const poolSize = Math.max(1, Math.min(Number(__ENV.SESSION_POOL || 10), vuCount));
  const sessions = [];
  console.log(`préparation de ${poolSize} session(s) loadtest…`);

  for (let i = 1; i <= poolSize; i++) {
    const email = `${LOAD_EMAIL_PREFIX}${i}@example.com`;
    const account = authenticateAccount(email);
    if (account.devAuth) {
      console.warn(
        '⚠ REVUES_DEV_AUTH actif — sessions admin partagées. Désactive-le pour un run propre.',
      );
      sessions.push(account);
      break;
    }
    sessions.push(account);
    if (i < poolSize) {
      sleep(2.2);
    }
  }

  console.log(`sessions prêtes: ${sessions.length}`);
  return sessions;
}

export function applySession(data) {
  if (!data || !data.sessions || data.sessions.length === 0) {
    fail('aucune session préparée (setup)');
  }
  const account = data.sessions[(__VU - 1) % data.sessions.length];
  const jar = http.cookieJar();
  jar.set(BASE_URL, 'revues_session', account.session, { path: '/' });
  vuAuth.email = account.email;
  vuAuth.session = account.session;
  vuAuth.csrf = account.csrf;
  vuAuth.ready = true;
  return vuAuth;
}

export function ensureAuth(data) {
  if (vuAuth.ready && vuAuth.session) {
    http.cookieJar().set(BASE_URL, 'revues_session', vuAuth.session, { path: '/' });
    return vuAuth;
  }
  return applySession(data);
}

export function freshCSRF() {
  ensureAuth();
  const me = http.get(apiURL('/me'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /me (csrf)' },
  });
  if (me.status === 200) {
    const body = parseJSON(me, 'me csrf');
    if (body.csrf_token) {
      vuAuth.csrf = body.csrf_token;
      return vuAuth.csrf;
    }
  }
  if (vuAuth.session) {
    http.cookieJar().set(BASE_URL, 'revues_session', vuAuth.session, { path: '/' });
    const me2 = http.get(apiURL('/me'), {
      headers: { Accept: 'application/json' },
      tags: { name: 'GET /me (csrf retry)' },
    });
    if (me2.status === 200) {
      const body = parseJSON(me2, 'me csrf retry');
      vuAuth.csrf = body.csrf_token;
      return vuAuth.csrf;
    }
  }
  fail(`freshCSRF: /me ${me.status}`);
}

/** Navigation lecture typique. */
export function readJourney() {
  const me = http.get(apiURL('/me'), { headers: { Accept: 'application/json' }, tags: { name: 'GET /me' } });
  check(me, { 'me 200': (r) => r.status === 200 });
  if (me.status === 200) {
    const body = parseJSON(me, 'me');
    if (body.csrf_token) {
      vuAuth.csrf = body.csrf_token;
    }
  }

  const subjects = http.get(apiURL('/subjects'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /subjects' },
  });
  check(subjects, { 'subjects 200': (r) => r.status === 200 });
  const subjectList = subjects.status === 200 ? parseJSON(subjects, 'subjects').subjects || [] : [];

  const templates = http.get(apiURL('/templates'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /templates' },
  });
  check(templates, { 'templates 200': (r) => r.status === 200 });

  const runs = http.get(apiURL('/runs?limit=25'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /runs' },
  });
  check(runs, { 'runs 200': (r) => r.status === 200 });
  const runList = runs.status === 200 ? parseJSON(runs, 'runs').runs || [] : [];

  if (subjectList.length > 0) {
    const sid = subjectList[Math.floor(Math.random() * subjectList.length)].id;
    const detail = http.get(apiURL(`/subjects/${sid}`), {
      headers: { Accept: 'application/json' },
      tags: { name: 'GET /subjects/{id}' },
    });
    check(detail, { 'subject detail 200': (r) => r.status === 200 });
  }

  if (runList.length > 0) {
    const rid = runList[Math.floor(Math.random() * runList.length)].id;
    const detail = http.get(apiURL(`/runs/${rid}`), {
      headers: { Accept: 'application/json' },
      tags: { name: 'GET /runs/{id}' },
    });
    check(detail, { 'run detail 200': (r) => r.status === 200 });
  }

  return { subjectList, runList };
}

/**
 * Lance une revue : POST /subjects/{id}/runs (snapshot transactionnel).
 * C'est le chemin write le plus lourd côté SQLite.
 */
export function launchRun() {
  let subjects = http.get(apiURL('/subjects'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /subjects (launch prep)' },
  });
  if (subjects.status !== 200) {
    check(subjects, { 'launch prep subjects 200': (r) => r.status === 200 });
    return;
  }
  let subjectList = parseJSON(subjects, 'launch subjects').subjects || [];
  if (subjectList.length === 0) {
    seedMinimalData(freshCSRF());
    subjects = http.get(apiURL('/subjects'), {
      headers: { Accept: 'application/json' },
      tags: { name: 'GET /subjects (launch prep)' },
    });
    if (subjects.status !== 200) {
      return;
    }
    subjectList = parseJSON(subjects, 'launch subjects retry').subjects || [];
    if (subjectList.length === 0) {
      return;
    }
  }

  const subject = subjectList[Math.floor(Math.random() * subjectList.length)];
  const catalog = http.get(apiURL(`/subjects/${subject.id}/run-templates`), {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /subjects/{id}/run-templates' },
  });
  check(catalog, { 'run-templates 200': (r) => r.status === 200 });
  if (catalog.status !== 200) {
    return;
  }
  const catalogBody = parseJSON(catalog, 'run-templates');
  let templateID = 0;
  const templates = catalogBody.templates || [];
  if (templates.length > 0 && catalogBody.can_launch !== false) {
    templateID = templates[Math.floor(Math.random() * templates.length)].id;
  } else {
    // Pas de modèle matchant : en créer un (domaines vides = tous les sujets).
    const tpl = http.post(
      apiURL('/templates'),
      JSON.stringify({
        name: `Load Tpl ${__VU}-${__ITER}`,
        domains: [],
        items: [
          { section: 'A', label: 'Check 1', required: true },
          { section: 'A', label: 'Check 2', required: false },
          { section: 'B', label: 'Check 3', required: false },
        ],
      }),
      {
        headers: jsonHeaders(freshCSRF()),
        tags: { name: 'POST /templates (launch prep)' },
        responseCallback: okStatuses,
      },
    );
    if (tpl.status !== 201 && tpl.status !== 200) {
      check(tpl, { 'launch prep template 201': (r) => r.status === 201 });
      return;
    }
    templateID = parseJSON(tpl, 'launch prep template').id;
  }

  const create = http.post(
    apiURL(`/subjects/${subject.id}/runs`),
    JSON.stringify({ template_id: templateID }),
    {
      headers: jsonHeaders(freshCSRF()),
      tags: { name: 'POST /subjects/{id}/runs' },
      responseCallback: okStatuses,
    },
  );
  check(create, { 'create run 201': (r) => r.status === 201 });

  if (create.status === 201) {
    const run = parseJSON(create, 'create run');
    const detail = http.get(apiURL(`/runs/${run.id}`), {
      headers: { Accept: 'application/json' },
      tags: { name: 'GET /runs/{id} (after launch)' },
    });
    check(detail, { 'run detail after launch 200': (r) => r.status === 200 });
  }
}

/** Mutation légère : PATCH item, sinon POST subject. */
export function writeMutation() {
  const runs = http.get(apiURL('/runs?status=in_progress&limit=25'), {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /runs (write prep)' },
  });

  if (runs.status === 200) {
    const list = parseJSON(runs, 'runs write').runs || [];
    if (list.length > 0) {
      const run = list[Math.floor(Math.random() * list.length)];
      const detail = http.get(apiURL(`/runs/${run.id}`), {
        headers: { Accept: 'application/json' },
        tags: { name: 'GET /runs/{id} (write prep)' },
      });
      if (detail.status === 200) {
        const body = parseJSON(detail, 'run detail write');
        const items = body.items || [];
        if (items.length > 0 && body.capabilities && body.capabilities.can_update_items) {
          const item = items[Math.floor(Math.random() * items.length)];
          const nextStatus = item.status === 'ok' ? 'pending' : 'ok';
          let csrf = freshCSRF();
          let patch = http.patch(
            apiURL(`/runs/${run.id}/items/${item.id}`),
            JSON.stringify({
              status: nextStatus,
              updated_at: item.updated_at,
            }),
            {
              headers: jsonHeaders(csrf),
              tags: { name: 'PATCH /runs/{id}/items/{itemId}' },
              responseCallback: okStatuses,
            },
          );
          if (patch.status === 403) {
            csrf = freshCSRF();
            patch = http.patch(
              apiURL(`/runs/${run.id}/items/${item.id}`),
              JSON.stringify({ status: nextStatus }),
              {
                headers: jsonHeaders(csrf),
                tags: { name: 'PATCH /runs/{id}/items/{itemId} retry' },
                responseCallback: okStatuses,
              },
            );
          }
          check(patch, {
            'patch item ok|conflict': (r) => r.status === 200 || r.status === 409,
          });
          if (patch.status === 200 || patch.status === 409) {
            return;
          }
        }
      }
    }
  }

  const name = `load-${__VU}-${__ITER}-${Date.now()}`;
  let create = http.post(
    apiURL('/subjects'),
    JSON.stringify({
      name,
      description: 'sujet créé par k6 load mix',
      domains: ['web'],
    }),
    {
      headers: jsonHeaders(freshCSRF()),
      tags: { name: 'POST /subjects' },
      responseCallback: okStatuses,
    },
  );
  if (create.status === 403) {
    create = http.post(
      apiURL('/subjects'),
      JSON.stringify({
        name: `${name}-retry`,
        description: 'sujet créé par k6 load mix',
        domains: ['web'],
      }),
      {
        headers: jsonHeaders(freshCSRF()),
        tags: { name: 'POST /subjects retry' },
        responseCallback: okStatuses,
      },
    );
  }
  check(create, { 'create subject 201': (r) => r.status === 201 });
}

export function healthCheck() {
  const res = http.get(`${BASE_URL}/healthz`);
  if (res.status !== 200) {
    fail(`API injoignable sur ${BASE_URL}/healthz (status ${res.status})`);
  }
}

export function defaultThresholds() {
  return {
    http_req_failed: ['rate<0.02'],
    http_req_duration: ['p(95)<800'],
    checks: ['rate>0.95'],
  };
}
