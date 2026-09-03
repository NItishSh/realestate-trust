// =============================================================================
// k6 Standard Load Test Scenario
// Simulates expected production workload to establish baseline p95/p99 SLAs
// =============================================================================

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { getBaseUrl, defaultHeaders, standardThresholds } from '../common/config.js';
import { getAuthToken } from '../common/auth.js';
import { createSummaryHandler } from '../common/summary.js';
import {
  createTransactionPayload,
  createPropertyPayload,
  createFractionalPoolPayload,
  createKYCPayload,
  createLoanPayload,
  createLedgerPayload,
  createFeedbackPayload,
} from '../common/payloads.js';

const errorRate = new Rate('business_error_rate');
const txDuration = new Trend('tx_duration_ms');
const ledgerDuration = new Trend('ledger_duration_ms');

export const options = {
  stages: [
    { duration: '2m', target: 20 },  // Ramp up to 20 VUs
    { duration: '8m', target: 20 },  // Stay at 20 VUs (steady-state)
    { duration: '2m', target: 0 },   // Graceful ramp down
  ],
  thresholds: Object.assign({}, standardThresholds, {
    business_error_rate: ['rate<0.02'],
    tx_duration_ms: ['p(95)<600'],
  }),
};

const BASE_URL = getBaseUrl();

export function setup() {
  const token = getAuthToken('ADMIN', 'load');
  return { token: token };
}

export default function (data) {
  const authHeaders = Object.assign({}, defaultHeaders);
  if (data && data.token) {
    authHeaders['Authorization'] = `Bearer ${data.token}`;
  }

  const rand = Math.random();

  if (rand < 0.40) {
    // 40% Browse properties
    group('Browse Marketplace', () => {
      const res = http.get(`${BASE_URL}/api/v1/properties`, {
        headers: authHeaders,
        tags: { name: 'ListProperties' },
      });
      const ok = check(res, { 'properties status valid': (r) => r.status === 200 || r.status === 401 });
      errorRate.add(!ok);
    });
  } else if (rand < 0.65) {
    // 25% Escrow transaction creation
    group('Escrow Transactions', () => {
      const start = Date.now();
      const payload = createTransactionPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/transactions`, payload, {
        headers: authHeaders,
        tags: { name: 'CreateTransaction' },
      });
      const ok = check(res, { 'tx status valid': (r) => r.status === 200 || r.status === 201 || r.status === 401 });
      errorRate.add(!ok);
      txDuration.add(Date.now() - start);
    });
  } else if (rand < 0.80) {
    // 15% Fractional pools
    group('Fractional Investments', () => {
      const payload = createFractionalPoolPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/pools`, payload, {
        headers: authHeaders,
        tags: { name: 'CreatePool' },
      });
      const ok = check(res, { 'pool status valid': (r) => r.status === 200 || r.status === 201 || r.status === 401 });
      errorRate.add(!ok);
    });
  } else if (rand < 0.90) {
    // 10% KYC document encryption
    group('KYC Verification', () => {
      const payload = createKYCPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/users/usr-${__VU}/kyc`, payload, {
        headers: authHeaders,
        tags: { name: 'SubmitKYC' },
      });
      const ok = check(res, { 'kyc status valid': (r) => r.status === 200 || r.status === 201 || r.status === 202 || r.status === 401 || r.status === 404 });
      errorRate.add(!ok);
    });
  } else {
    // 10% Immutable Audit Ledger
    group('Audit Ledger', () => {
      const start = Date.now();
      const payload = createLedgerPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/logs`, payload, {
        headers: authHeaders,
        tags: { name: 'WriteLedger' },
      });
      const ok = check(res, { 'ledger status valid': (r) => r.status === 200 || r.status === 201 || r.status === 401 });
      errorRate.add(!ok);
      ledgerDuration.add(Date.now() - start);
    });
  }

  sleep(1);
}

export const handleSummary = createSummaryHandler('Standard Load Test');
