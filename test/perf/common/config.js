// =============================================================================
// Common Configuration & Thresholds for RealEstate-Trust k6 Load Tests
// =============================================================================

export function getBaseUrl() {
  return __ENV.TARGET_URL || 'http://localhost:8080';
}

export const defaultHeaders = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
};

// Standard production SLA thresholds for RealEstate-Trust
export const standardThresholds = {
  http_req_failed: ['rate<0.01'],              // Less than 1% errors
  http_req_duration: ['p(95)<500', 'p(99)<1200'], // 95% < 500ms, 99% < 1.2s
};
