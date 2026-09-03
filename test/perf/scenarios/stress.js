// =============================================================================
// k6 Stress & Saturation Test Scenario
// Ramps beyond standard capacity to discover breaking points and saturation limits
// =============================================================================

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { getBaseUrl, defaultHeaders } from '../common/config.js';
import { getAuthToken } from '../common/auth.js';
import {
  createTransactionPayload,
  createPropertyPayload,
  createFractionalPoolPayload,
  createKYCPayload,
  createLedgerPayload,
} from '../common/payloads.js';

const errorRate = new Rate('stress_error_rate');
const latencyTrend = new Trend('stress_latency_ms');

export const options = {
  stages: [
    { duration: '2m', target: 20 },   // Warm-up to 20 VUs
    { duration: '5m', target: 50 },   // Normal high load
    { duration: '5m', target: 100 },  // Escalating stress
    { duration: '5m', target: 150 },  // Heavy saturation
    { duration: '3m', target: 200 },  // Extreme peak breaking point test
    { duration: '3m', target: 0 },    // Cooldown
  ],
  thresholds: {
    // In stress testing, we expect degradation; thresholds record acceptable boundaries before full outage
    http_req_failed: ['rate<0.15'],    // Allow up to 15% error rate at absolute peak saturation
    http_req_duration: ['p(90)<1500'], // 90% within 1.5s under maximum stress
  },
};

const BASE_URL = getBaseUrl();

export function setup() {
  const token = getAuthToken('ADMIN', 'stress');
  return { token: token };
}

export default function (data) {
  const authHeaders = Object.assign({}, defaultHeaders);
  if (data && data.token) {
    authHeaders['Authorization'] = `Bearer ${data.token}`;
  }

  const rand = Math.random();
  const start = Date.now();

  if (rand < 0.35) {
    // Properties query
    group('High Concurrency Queries', () => {
      const res = http.get(`${BASE_URL}/api/v1/properties`, {
        headers: authHeaders,
        tags: { name: 'Stress_ListProperties' },
      });
      const ok = check(res, { 'status <= 499': (r) => r.status < 500 });
      errorRate.add(!ok);
    });
  } else if (rand < 0.65) {
    // Concurrent Transaction creation
    group('High Concurrency Escrows', () => {
      const payload = createTransactionPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/transactions`, payload, {
        headers: authHeaders,
        tags: { name: 'Stress_CreateTx' },
      });
      const ok = check(res, { 'status <= 499': (r) => r.status < 500 });
      errorRate.add(!ok);
    });
  } else if (rand < 0.85) {
    // Fractional token pool transactions
    group('High Concurrency Pools', () => {
      const payload = createFractionalPoolPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/pools`, payload, {
        headers: authHeaders,
        tags: { name: 'Stress_CreatePool' },
      });
      const ok = check(res, { 'status <= 499': (r) => r.status < 500 });
      errorRate.add(!ok);
    });
  } else {
    // Cryptographic audit logs
    group('High Concurrency Ledger', () => {
      const payload = createLedgerPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/logs`, payload, {
        headers: authHeaders,
        tags: { name: 'Stress_WriteLedger' },
      });
      const ok = check(res, { 'status <= 499': (r) => r.status < 500 });
      errorRate.add(!ok);
    });
  }

  latencyTrend.add(Date.now() - start);
  sleep(0.5); // Tight sleep for stress testing
}
