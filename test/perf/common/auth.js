// =============================================================================
// Authentication Helper for RealEstate-Trust k6 Performance Tests
// =============================================================================

import crypto from 'k6/crypto';
import encoding from 'k6/encoding';

// Development JWT Secret matching monorepo internal/db/middleware.go
const JWT_SECRET = __ENV.JWT_SECRET || 'super-secret-key-for-local-demo-only';

function base64Url(str) {
  return encoding.b64encode(str).replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_');
}

/**
 * Generate a valid, cryptographically signed HS256 JWT locally in k6.
 * Bypasses auth rate-limiting bottlenecks during high-throughput load tests.
 *
 * @param {string} role - BUYER, SELLER, BROKER, OFFICER, or ADMIN
 * @param {string} userId - Unique identifier for the subject claim
 * @returns {string} Signed JWT Bearer token
 */
export function getAuthToken(role = 'BUYER', userId = 'perf-user') {
  const header = base64Url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const now = Math.floor(Date.now() / 1000);
  const payload = base64Url(JSON.stringify({
    sub: userId,
    role: role,
    iat: now,
    exp: now + 86400, // 24-hour validity for sustained load tests
  }));

  const sigB64 = crypto.hmac('sha256', JWT_SECRET, `${header}.${payload}`, 'base64');
  const signature = sigB64.replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_');
  return `${header}.${payload}.${signature}`;
}
