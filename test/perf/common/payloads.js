// =============================================================================
// Reusable Realistic Payloads for RealEstate-Trust k6 Scenarios
// =============================================================================

export function createTransactionPayload(vuId = 1, propertyId = null, buyerId = null, sellerId = null) {
  const nonce = `${Date.now()}-${vuId}-${Math.floor(Math.random() * 10000)}`;
  return JSON.stringify({
    propertyId: propertyId || `prop-perf-${nonce}`,
    buyerId: buyerId || `usr-buyer-${vuId}`,
    sellerId: sellerId || `usr-seller-${vuId}`,
    totalAmount: 450000.00,
  });
}

export function createPropertyPayload(vuId = 1) {
  const nonce = `${Date.now()}-${vuId}-${Math.floor(Math.random() * 10000)}`;
  return JSON.stringify({
    address: `${vuId * 100 + 12} Innovation Way, Silicon Valley, CA`,
    description: `Luxury Smart Residence unit ${nonce} with smart metering.`,
    value: 850000.00,
    thumbnail: 'https://images.unsplash.com/photo-1518780664697-55e3ad937233?w=800',
    sqft: 2400,
    bedrooms: 3,
    bathrooms: 2,
    yearBuilt: 2022,
    propertyType: 'RESIDENTIAL',
  });
}

export function createPropertyDetailsUpdatePayload(vuId = 1) {
  return JSON.stringify({
    address: `Updated 742 Evergreen Terrace #${vuId}`,
    description: 'Fully refurbished property with smart home automation.',
    value: 950000.00,
    thumbnail: 'https://images.unsplash.com/photo-1518780664697-55e3ad937233?w=800',
    sqft: 2800,
    bedrooms: 4,
    bathrooms: 3,
    yearBuilt: 2021,
    propertyType: 'RESIDENTIAL',
  });
}

export function createUnlockDocsPayload(buyerId = 'usr-buyer-1') {
  return JSON.stringify({
    buyerId: buyerId,
  });
}

export function createTitleInsurancePayload(policyNumber = 'POL-99281', company = 'First American Title') {
  return JSON.stringify({
    policy: policyNumber,
    company: company,
  });
}

export function createFundEscrowPayload(amount = 50000.00) {
  return JSON.stringify({
    amount: amount,
  });
}

export function createUpdateTxStatusPayload(newState = 'UNDER_REVIEW') {
  return JSON.stringify({
    newState: newState,
  });
}

export function createFractionalPoolPayload(vuId = 1, propertyId = null) {
  const nonce = `${Date.now()}-${vuId}-${Math.floor(Math.random() * 10000)}`;
  return JSON.stringify({
    propertyId: propertyId || `prop-pool-${nonce}`,
    totalTokens: 1000,
    tokenPrice: 250.00,
  });
}

export function createBuySharesPayload(investorId = 'usr-buyer-1', tokenCount = 10) {
  return JSON.stringify({
    investorId: investorId,
    tokenCount: tokenCount,
  });
}

export function createKYCPayload(vuId = 1) {
  const nonce = `${Date.now()}-${vuId}`;
  return JSON.stringify({
    documentType: 'PASSPORT',
    documentReference: `PASS-US-${nonce}`,
  });
}

export function createLoanPayload(vuId = 1, transactionId = null, userId = null) {
  const nonce = `${Date.now()}-${vuId}-${Math.floor(Math.random() * 10000)}`;
  return JSON.stringify({
    transactionId: transactionId || `tx-loan-${nonce}`,
    userId: userId || `usr-buyer-${vuId}`,
    requestedAmount: 320000.00,
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
    category: 'ESCROW',
    rating: 5,
    message: `Exceptional escrow closing speed and institutional-grade transparency for VU ${vuId}.`,
  });
}

export function createUserRegistrationPayload(vuId = 1, role = 'BUYER') {
  const nonce = `${Date.now()}-${vuId}-${Math.floor(Math.random() * 100000)}`;
  return JSON.stringify({
    email: `user-${nonce}@perf-test.local`,
    password: `PerfPass_${nonce}!`,
    fullName: `Performance Tester ${vuId}`,
    role: role,
  });
}
