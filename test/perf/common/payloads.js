// =============================================================================
// Reusable Realistic Payloads for RealEstate-Trust k6 Scenarios
// =============================================================================

export function createTransactionPayload(vuId = 1) {
  const nonce = `${Date.now()}-${vuId}-${Math.floor(Math.random() * 10000)}`;
  return JSON.stringify({
    propertyId: `prop-perf-${nonce}`,
    buyerId: `usr-buyer-${vuId}`,
    sellerId: `usr-seller-${vuId}`,
    totalAmount: 450000.00,
  });
}

export function createPropertyPayload(vuId = 1) {
  const nonce = `${Date.now()}-${vuId}`;
  return JSON.stringify({
    title: `Luxury Apartment Unit ${nonce}`,
    address: `${nonce} Innovation Way, Silicon Valley, CA`,
    valuation: 850000.00,
    ownerId: `usr-seller-${vuId}`,
  });
}

export function createFractionalPoolPayload(vuId = 1) {
  const nonce = `${Date.now()}-${vuId}`;
  return JSON.stringify({
    propertyId: `prop-pool-${nonce}`,
    totalShares: 1000,
    pricePerShare: 250.00,
  });
}

export function createKYCPayload(vuId = 1) {
  const nonce = `${Date.now()}-${vuId}`;
  return JSON.stringify({
    documentType: 'PASSPORT',
    documentReference: `PASS-US-${nonce}`,
  });
}

export function createLoanPayload(vuId = 1) {
  const nonce = `${Date.now()}-${vuId}`;
  return JSON.stringify({
    transactionId: `tx-loan-${nonce}`,
    applicantId: `usr-buyer-${vuId}`,
    loanAmount: 320000.00,
    propertyValue: 450000.00,
  });
}

export function createLedgerPayload(vuId = 1) {
  return JSON.stringify({
    payload: `PERF_AUDIT_LOG: VU ${vuId} executed deal verification at ${new Date().toISOString()}`,
  });
}

export function createFeedbackPayload(vuId = 1) {
  return JSON.stringify({
    brokerId: `broker-${vuId}`,
    reviewerId: `usr-buyer-${vuId}`,
    rating: 5,
    comment: 'Exceptional escrow closing speed and institutional-grade transparency.',
  });
}
