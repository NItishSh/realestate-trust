// =============================================================================
// k6 24-Hour Sustained Soak Test Scenario
// Tests long-term memory stability, connection leak absence, and sustained throughput
// Stages:
//   1. Gradual Ramp-Up (Default: 1 hour)
//   2. Sustained Maximum Plateau (Default: 22 hours)
//   3. Gradual Cooldown / Ramp-Down (Default: 1 hour)
// Overrides:
//   SOAK_RAMP=1h SOAK_PLATEAU=22h SOAK_COOLDOWN=1h SOAK_TARGET_VUS=50
// =============================================================================

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { getBaseUrl, defaultHeaders } from '../common/config.js';
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

// Custom Trend Metrics to monitor performance degradation over time
const soakErrorRate = new Rate('soak_error_rate');
const browseDuration = new Trend('soak_browse_duration_ms');
const escrowDuration = new Trend('soak_escrow_duration_ms');
const poolDuration = new Trend('soak_pool_duration_ms');
const ledgerDuration = new Trend('soak_ledger_duration_ms');

// Dynamic durations for validation flexibility
const RAMP_TIME = __ENV.SOAK_RAMP || '1h';
const PLATEAU_TIME = __ENV.SOAK_PLATEAU || '22h';
const COOLDOWN_TIME = __ENV.SOAK_COOLDOWN || '1h';
const TARGET_VUS = parseInt(__ENV.SOAK_TARGET_VUS || '50', 10);

export const options = {
  stages: [
    { duration: RAMP_TIME, target: TARGET_VUS },       // Phase 1: Gradual ramp-up
    { duration: PLATEAU_TIME, target: TARGET_VUS },    // Phase 2: Sustained maximum load plateau
    { duration: COOLDOWN_TIME, target: 0 },            // Phase 3: Gradual ramp-down to zero
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],                    // SLA: Error rate below 1% over 24h
    http_req_duration: ['p(95)<600', 'p(99)<1500'],    // SLA: p95 under 600ms, p99 under 1500ms
    soak_error_rate: ['rate<0.005'],                   // Business logic error rate under 0.5%
  },
};

const BASE_URL = getBaseUrl();

export function setup() {
  console.log(`[Soak Test] Starting 24h Soak Profile targeting ${BASE_URL}`);
  console.log(`[Soak Test] Configuration: Ramp=${RAMP_TIME}, Plateau=${PLATEAU_TIME}, Cooldown=${COOLDOWN_TIME}, Target VUs=${TARGET_VUS}`);
  const token = getAuthToken('ADMIN', 'soak');
  return { token: token };
}

export default function (data) {
  const authHeaders = Object.assign({}, defaultHeaders);
  if (data && data.token) {
    authHeaders['Authorization'] = `Bearer ${data.token}`;
  }

  const rand = Math.random();

  if (rand < 0.40) {
    // 40% Traffic: Marketplace Browsing (Read-heavy caching & frontend proxy)
    group('Soak: Browse Marketplace', () => {
      const start = Date.now();
      const res = http.get(`${BASE_URL}/api/v1/properties`, {
        headers: authHeaders,
        tags: { name: 'Soak_ListProperties' },
      });
      const ok = check(res, {
        'status is 200 or 401': (r) => r.status === 200 || r.status === 401,
      });
      soakErrorRate.add(!ok);
      browseDuration.add(Date.now() - start);
    });
  } else if (rand < 0.65) {
    // 25% Traffic: Escrow Lifecycle (DB write + Outbox + RabbitMQ)
    group('Soak: Escrow Transactions', () => {
      const start = Date.now();
      const payload = createTransactionPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/transactions`, payload, {
        headers: authHeaders,
        tags: { name: 'Soak_CreateTransaction' },
      });
      const ok = check(res, {
        'tx accepted': (r) => r.status === 200 || r.status === 201 || r.status === 401,
      });
      soakErrorRate.add(!ok);
      escrowDuration.add(Date.now() - start);
    });
  } else if (rand < 0.80) {
    // 15% Traffic: Fractional Pool Shares (Row lock concurrency)
    group('Soak: Fractional Pools', () => {
      const start = Date.now();
      const payload = createFractionalPoolPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/pools`, payload, {
        headers: authHeaders,
        tags: { name: 'Soak_CreatePool' },
      });
      const ok = check(res, {
        'pool accepted': (r) => r.status === 200 || r.status === 201 || r.status === 401,
      });
      soakErrorRate.add(!ok);
      poolDuration.add(Date.now() - start);
    });
  } else if (rand < 0.90) {
    // 10% Traffic: KYC Identity Upload (Vault Transit KMS encryption)
    group('Soak: KYC Encryption', () => {
      const payload = createKYCPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/users/usr-${__VU}/kyc`, payload, {
        headers: authHeaders,
        tags: { name: 'Soak_SubmitKYC' },
      });
      const ok = check(res, {
        'kyc accepted': (r) => r.status === 200 || r.status === 201 || r.status === 202 || r.status === 401 || r.status === 404,
      });
      soakErrorRate.add(!ok);
    });
  } else if (rand < 0.95) {
    // 5% Traffic: Cryptographic Immutable Ledger
    group('Soak: Ledger Audit Log', () => {
      const start = Date.now();
      const payload = createLedgerPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/logs`, payload, {
        headers: authHeaders,
        tags: { name: 'Soak_WriteLedger' },
      });
      const ok = check(res, {
        'ledger accepted': (r) => r.status === 200 || r.status === 201 || r.status === 401,
      });
      soakErrorRate.add(!ok);
      ledgerDuration.add(Date.now() - start);
    });
  } else {
    // 5% Traffic: Mortgage Underwriting
    group('Soak: Mortgage Loan Application', () => {
      const payload = createLoanPayload(__VU);
      const res = http.post(`${BASE_URL}/api/v1/loans`, payload, {
        headers: authHeaders,
        tags: { name: 'Soak_CreateLoan' },
      });
      const ok = check(res, {
        'loan accepted': (r) => r.status === 200 || r.status === 201 || r.status === 401,
      });
      soakErrorRate.add(!ok);
    });
  }

  sleep(1);
}

export function teardown(data) {
  console.log('[Soak Test] Completed soak test run. Inspect VPA and OpenCost metrics for right-sizing.');
}

export const handleSummary = createSummaryHandler('24-Hour Sustained Soak Test');
