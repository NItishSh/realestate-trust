// =============================================================================
// k6 Performance Scenario: End-to-End User Journeys (100% Endpoint Coverage)
// =============================================================================
// Simulates realistic concurrent production user personas:
// - Buyer Journey (60%): Register -> KYC -> Browse -> Unlock Docs -> Transaction -> Fund Escrow -> Loan -> Feedback
// - Seller / Broker Journey (20%): Create Property -> Update Details -> Tokenize Pool -> Buy Shares -> Update Status
// - Compliance Officer Journey (10%): Audit Users -> Verify Insurance -> Review Loans
// - Auditor & Admin Journey (10%): Write Ledger Log -> Fetch Chain -> Verify Block -> Review Feedback -> List Tx -> Logout
// =============================================================================

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { getAuthToken } from '../common/auth.js';
import {
  createTransactionPayload,
  createPropertyPayload,
  createPropertyDetailsUpdatePayload,
  createUnlockDocsPayload,
  createTitleInsurancePayload,
  createFundEscrowPayload,
  createUpdateTxStatusPayload,
  createFractionalPoolPayload,
  createBuySharesPayload,
  createKYCPayload,
  createLoanPayload,
  createLedgerPayload,
  createFeedbackPayload,
  createUserRegistrationPayload,
} from '../common/payloads.js';
import { createSummaryHandler } from '../common/summary.js';

export const handleSummary = createSummaryHandler('User Journeys E2E Test');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const DURATION = __ENV.JOURNEY_DURATION || '1m';
const VUS = parseInt(__ENV.VUS || '10', 10);

export const options = {
  scenarios: {
    user_journeys: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '10s', target: VUS },
        { duration: DURATION, target: VUS },
        { duration: '10s', target: 0 },
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'], // Max 5% failure under multi-step stateful workflows
    http_req_duration: ['p(95)<5000', 'p(90)<2500'], // 90% under 2.5s, 95% under 5s across service mesh
  },
};

function getHeaders(token = null) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

// -----------------------------------------------------------------------------
// 1. Buyer Journey (60% of Virtual Users)
// -----------------------------------------------------------------------------
function runBuyerJourney(vuId, iteration) {
  const buyerId = `buyer-vu-${vuId}-${iteration}`;
  const buyerToken = getAuthToken('BUYER', buyerId);

  // 1.1 Frontend Portal Visit
  group('Buyer_01_FrontendPortal', () => {
    const res = http.get(`${BASE_URL}/`, { tags: { name: 'Frontend_Portal' } });
    check(res, { 'portal status 200': (r) => r.status === 200 });
  });

  // 1.2 User Registration
  group('Buyer_02_Register', () => {
    const regPayload = createUserRegistrationPayload(vuId, 'BUYER');
    const res = http.post(`${BASE_URL}/api/v1/users`, regPayload, {
      headers: getHeaders(),
      tags: { name: 'Identity_Register' },
    });
    check(res, { 'register responded': (r) => r.status === 201 || r.status === 409 || r.status === 429 });
  });

  // 1.3 Submit & Check KYC
  group('Buyer_03_KYC', () => {
    const kycPayload = createKYCPayload(vuId);
    const postRes = http.post(`${BASE_URL}/api/v1/users/${buyerId}/kyc`, kycPayload, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Identity_SubmitKYC' },
    });
    check(postRes, { 'kyc submit accepted 200-202': (r) => r.status === 200 || r.status === 201 || r.status === 202 });

    const getRes = http.get(`${BASE_URL}/api/v1/users/${buyerId}/kyc/status`, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Identity_GetKYCStatus' },
    });
    check(getRes, { 'kyc get status 200': (r) => r.status === 200 });
  });

  // 1.4 Browse & View Properties
  let selectedPropertyId = 'prop-seed-1';
  group('Buyer_04_PropertyDiscovery', () => {
    const listRes = http.get(`${BASE_URL}/api/v1/properties`, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Property_List' },
    });
    const passed = check(listRes, { 'list properties status 200': (r) => r.status === 200 });
    if (passed && listRes.json() && Array.isArray(listRes.json()) && listRes.json().length > 0) {
      selectedPropertyId = listRes.json()[0].id || selectedPropertyId;
    }

    const detailRes = http.get(`${BASE_URL}/api/v1/properties/${selectedPropertyId}`, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Property_GetDetail' },
    });
    check(detailRes, { 'property detail 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 1.5 Unlock Legal Documents
  group('Buyer_05_UnlockDocuments', () => {
    const unlockPayload = createUnlockDocsPayload(buyerId);
    const res = http.post(`${BASE_URL}/api/v1/properties/${selectedPropertyId}/documents/unlock`, unlockPayload, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Property_UnlockDocs' },
    });
    check(res, { 'unlock documents 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 1.6 Create Escrow Purchase Transaction
  let txId = `tx-vu-${vuId}-${iteration}`;
  group('Buyer_06_CreateTransaction', () => {
    const txPayload = createTransactionPayload(vuId, selectedPropertyId, buyerId, `seller-vu-${vuId}`);
    const res = http.post(`${BASE_URL}/api/v1/transactions`, txPayload, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Transaction_Create' },
    });
    const ok = check(res, { 'create tx 201': (r) => r.status === 201 });
    if (ok && res.json() && res.json().id) {
      txId = res.json().id;
    }

    const getRes = http.get(`${BASE_URL}/api/v1/transactions/${txId}`, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Transaction_Get' },
    });
    check(getRes, { 'get tx 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 1.7 Fund Escrow & Verify Balance
  group('Buyer_07_EscrowOperations', () => {
    const fundPayload = createFundEscrowPayload(50000.00);
    const fundRes = http.post(`${BASE_URL}/api/v1/transactions/${txId}/escrow/fund`, fundPayload, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Transaction_FundEscrow' },
    });
    check(fundRes, { 'fund escrow 200/404': (r) => r.status === 200 || r.status === 404 });

    const escrowRes = http.get(`${BASE_URL}/api/v1/transactions/${txId}/escrow`, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Transaction_GetEscrow' },
    });
    check(escrowRes, { 'get escrow 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 1.8 Apply for Financing Loan
  group('Buyer_08_ApplyFinancing', () => {
    const loanPayload = createLoanPayload(vuId, txId, buyerId);
    const loanRes = http.post(`${BASE_URL}/api/v1/loans`, loanPayload, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Financing_ApplyLoan' },
    });
    check(loanRes, { 'apply loan 201/400': (r) => r.status === 201 || r.status === 400 });
  });

  // 1.9 Submit Customer Feedback
  group('Buyer_09_Feedback', () => {
    const feedbackPayload = createFeedbackPayload(vuId);
    const res = http.post(`${BASE_URL}/api/v1/feedback`, feedbackPayload, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Feedback_Submit' },
    });
    check(res, { 'submit feedback 200/201': (r) => r.status === 200 || r.status === 201 });
  });
}

// -----------------------------------------------------------------------------
// 2. Seller & Broker Journey (20% of Virtual Users)
// -----------------------------------------------------------------------------
function runSellerBrokerJourney(vuId, iteration) {
  const sellerId = `seller-vu-${vuId}-${iteration}`;
  const sellerToken = getAuthToken('SELLER', sellerId);

  let propertyId = `prop-seller-${vuId}-${iteration}`;

  // 2.1 Create New Property Listing
  group('Seller_01_CreateProperty', () => {
    const propPayload = createPropertyPayload(vuId);
    const res = http.post(`${BASE_URL}/api/v1/properties`, propPayload, {
      headers: getHeaders(sellerToken),
      tags: { name: 'Property_Create' },
    });
    const ok = check(res, { 'create property 201': (r) => r.status === 201 });
    if (ok && res.json() && res.json().id) {
      propertyId = res.json().id;
    }
  });

  // 2.2 Update Property Details
  group('Seller_02_UpdateProperty', () => {
    const updatePayload = createPropertyDetailsUpdatePayload(vuId);
    const res = http.put(`${BASE_URL}/api/v1/properties/${propertyId}/details`, updatePayload, {
      headers: getHeaders(sellerToken),
      tags: { name: 'Property_UpdateDetails' },
    });
    check(res, { 'update property 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 2.3 Create Fractional Tokenization Pool
  let poolId = `pool-${vuId}-${iteration}`;
  group('Seller_03_CreatePool', () => {
    const poolPayload = createFractionalPoolPayload(vuId, propertyId);
    const res = http.post(`${BASE_URL}/api/v1/pools`, poolPayload, {
      headers: getHeaders(sellerToken),
      tags: { name: 'Tokenization_CreatePool' },
    });
    const ok = check(res, { 'create pool 201': (r) => r.status === 201 });
    if (ok && res.json() && res.json().id) {
      poolId = res.json().id;
    }

    const listPoolsRes = http.get(`${BASE_URL}/api/v1/pools`, {
      headers: getHeaders(sellerToken),
      tags: { name: 'Tokenization_ListPools' },
    });
    check(listPoolsRes, { 'list pools 200': (r) => r.status === 200 });
  });

  // 2.4 Buy Shares in Pool (Broker / Investor Flow)
  group('Seller_04_BuyPoolShares', () => {
    const buyerToken = getAuthToken('BUYER', `buyer-inv-${vuId}`);
    const buyPayload = createBuySharesPayload(`buyer-inv-${vuId}`, 5);
    const res = http.post(`${BASE_URL}/api/v1/pools/${poolId}/buy`, buyPayload, {
      headers: getHeaders(buyerToken),
      tags: { name: 'Tokenization_BuyShares' },
    });
    check(res, { 'buy shares 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 2.5 Update Transaction Status
  group('Seller_05_UpdateTxStatus', () => {
    const statusPayload = createUpdateTxStatusPayload('UNDER_REVIEW');
    const res = http.put(`${BASE_URL}/api/v1/transactions/tx-seed-1/status`, statusPayload, {
      headers: getHeaders(sellerToken),
      tags: { name: 'Transaction_UpdateStatus' },
    });
    check(res, { 'update tx status 200/400/404': (r) => r.status === 200 || r.status === 400 || r.status === 404 });
  });
}

// -----------------------------------------------------------------------------
// 3. Compliance Officer Journey (10% of Virtual Users)
// -----------------------------------------------------------------------------
function runOfficerJourney(vuId, iteration) {
  const officerToken = getAuthToken('OFFICER', `officer-vu-${vuId}`);

  // 3.1 Review Users Directory
  group('Officer_01_ListUsers', () => {
    const res = http.get(`${BASE_URL}/api/v1/users`, {
      headers: getHeaders(officerToken),
      tags: { name: 'Identity_ListUsers' },
    });
    check(res, { 'list users 200': (r) => r.status === 200 });
  });

  // 3.2 Inspect Specific User Details
  group('Officer_02_GetUser', () => {
    const res = http.get(`${BASE_URL}/api/v1/users/usr-seed-1`, {
      headers: getHeaders(officerToken),
      tags: { name: 'Identity_GetUser' },
    });
    check(res, { 'get user 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 3.3 Verify Title Insurance
  group('Officer_03_VerifyTitleInsurance', () => {
    const insurancePayload = createTitleInsurancePayload(`POL-AUTO-${vuId}`, 'Chicago Title Insurance');
    const res = http.post(`${BASE_URL}/api/v1/properties/prop-seed-1/insurance/verify`, insurancePayload, {
      headers: getHeaders(officerToken),
      tags: { name: 'Property_VerifyInsurance' },
    });
    check(res, { 'verify insurance 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 3.4 Review Loans & Inspect Detail
  group('Officer_04_ReviewLoans', () => {
    const loansRes = http.get(`${BASE_URL}/api/v1/loans`, {
      headers: getHeaders(officerToken),
      tags: { name: 'Financing_ListLoans' },
    });
    check(loansRes, { 'list loans 200': (r) => r.status === 200 });

    const loanDetailRes = http.get(`${BASE_URL}/api/v1/loans/loan-seed-1`, {
      headers: getHeaders(officerToken),
      tags: { name: 'Financing_GetLoan' },
    });
    check(loanDetailRes, { 'get loan 200/404': (r) => r.status === 200 || r.status === 404 });
  });
}

// -----------------------------------------------------------------------------
// 4. Auditor & Platform Admin Journey (10% of Virtual Users)
// -----------------------------------------------------------------------------
function runAdminAuditorJourney(vuId, iteration) {
  const adminToken = getAuthToken('ADMIN', `admin-vu-${vuId}`);

  // 4.1 Write Cryptographic Ledger Audit Log
  group('Admin_01_WriteLedger', () => {
    const ledgerPayload = createLedgerPayload(vuId);
    const res = http.post(`${BASE_URL}/api/v1/logs`, ledgerPayload, {
      headers: getHeaders(adminToken),
      tags: { name: 'Ledger_WriteLog' },
    });
    check(res, { 'write ledger log 200/201': (r) => r.status === 200 || r.status === 201 });
  });

  // 4.2 Fetch Immutable Chain
  group('Admin_02_GetLedgerChain', () => {
    const res = http.get(`${BASE_URL}/api/v1/logs`, {
      headers: getHeaders(adminToken),
      tags: { name: 'Ledger_GetLogs' },
    });
    check(res, { 'get ledger logs 200': (r) => r.status === 200 });

    const blockRes = http.get(`${BASE_URL}/api/v1/logs/0`, {
      headers: getHeaders(adminToken),
      tags: { name: 'Ledger_GetLogIndex' },
    });
    check(blockRes, { 'get block 0 200/404': (r) => r.status === 200 || r.status === 404 });
  });

  // 4.3 Review Customer Feedback
  group('Admin_03_ListFeedback', () => {
    const res = http.get(`${BASE_URL}/api/v1/feedback`, {
      headers: getHeaders(adminToken),
      tags: { name: 'Feedback_List' },
    });
    check(res, { 'list feedback 200': (r) => r.status === 200 });
  });

  // 4.4 List All Transactions Across Network
  group('Admin_04_ListTransactions', () => {
    const res = http.get(`${BASE_URL}/api/v1/transactions`, {
      headers: getHeaders(adminToken),
      tags: { name: 'Transaction_List' },
    });
    check(res, { 'list transactions 200': (r) => r.status === 200 });
  });

  // 4.5 Logout Test
  group('Admin_05_Logout', () => {
    const logoutPayload = JSON.stringify({ refreshToken: `ref-tok-${vuId}` });
    const res = http.post(`${BASE_URL}/api/v1/logout`, logoutPayload, {
      headers: getHeaders(adminToken),
      tags: { name: 'Identity_Logout' },
    });
    check(res, { 'logout 200': (r) => r.status === 200 });
  });
}

// -----------------------------------------------------------------------------
// Main Workload Execution (Weighted Persona Distribution)
// -----------------------------------------------------------------------------
export default function () {
  const vuId = __VU;
  const iteration = __ITER;

  // Modulo-based persona allocation (60% Buyer, 20% Seller, 10% Officer, 10% Admin)
  const personaBucket = (vuId + iteration) % 10;

  if (personaBucket < 6) {
    runBuyerJourney(vuId, iteration);
  } else if (personaBucket < 8) {
    runSellerBrokerJourney(vuId, iteration);
  } else if (personaBucket === 8) {
    runOfficerJourney(vuId, iteration);
  } else {
    runAdminAuditorJourney(vuId, iteration);
  }

  // Realistic user think time between workflows
  sleep(1);
}
