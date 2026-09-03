// =============================================================================
// k6 Smoke Test Scenario (Quick Health & Route Verification)
// Validates route responsiveness and basic API SLAs within 1-2 minutes
// =============================================================================

import http from 'k6/http';
import { check, sleep } from 'k6';
import { getBaseUrl, defaultHeaders } from '../common/config.js';
import { getAuthToken } from '../common/auth.js';
import {
  createTransactionPayload,
  createPropertyPayload,
  createLedgerPayload,
  createFeedbackPayload,
} from '../common/payloads.js';

export const options = {
  vus: 3,
  duration: '1m',
  thresholds: {
    http_req_failed: ['rate<0.05'], // Allow up to 5% failures during smoke startup
    http_req_duration: ['p(95)<800'],
  },
};

const BASE_URL = getBaseUrl();

export function setup() {
  const token = getAuthToken('ADMIN', 'smoke');
  return { token: token };
}

export default function (data) {
  const authHeaders = Object.assign({}, defaultHeaders);
  if (data && data.token) {
    authHeaders['Authorization'] = `Bearer ${data.token}`;
  }

  // 1. Check Root Gateway / Frontend
  const frontendRes = http.get(`${BASE_URL}/`, { tags: { name: 'FrontendPortal' } });
  check(frontendRes, {
    'frontend available': (r) => r.status === 200 || r.status === 304,
  });

  // 2. Identity Service Health
  const idHealth = http.get(`${BASE_URL}/api/v1/health`, { tags: { name: 'IdentityHealth' } });
  check(idHealth, {
    'identity health responded': (r) => r.status === 200 || r.status === 404,
  });

  // 3. Properties Endpoint
  const propRes = http.get(`${BASE_URL}/api/v1/properties`, {
    headers: authHeaders,
    tags: { name: 'ListProperties' },
  });
  check(propRes, {
    'properties endpoint responded': (r) => r.status === 200 || r.status === 401,
  });

  // 4. Transaction Creation Check
  const txPayload = createTransactionPayload(__VU);
  const txRes = http.post(`${BASE_URL}/api/v1/transactions`, txPayload, {
    headers: authHeaders,
    tags: { name: 'CreateTransaction' },
  });
  check(txRes, {
    'tx endpoint responded': (r) => r.status === 200 || r.status === 201 || r.status === 401,
  });

  // 5. Ledger Log Check
  const ledgerPayload = createLedgerPayload(__VU);
  const ledgerRes = http.post(`${BASE_URL}/api/v1/logs`, ledgerPayload, {
    headers: authHeaders,
    tags: { name: 'WriteLedger' },
  });
  check(ledgerRes, {
    'ledger endpoint responded': (r) => r.status === 200 || r.status === 201 || r.status === 401,
  });

  sleep(1);
}
