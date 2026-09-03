// =============================================================================
// Authentication Helper for RealEstate-Trust k6 Performance Tests
// =============================================================================

import http from 'k6/http';
import { getBaseUrl, defaultHeaders } from './config.js';

export function getAuthToken(role = 'BUYER', prefix = 'perf') {
  const baseUrl = getBaseUrl();
  const id = `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1000000)}`;
  const email = `${id}@test.local`;
  const password = `PerfPass123!`;

  // 1. Register user with desired role
  const regPayload = JSON.stringify({
    email: email,
    password: password,
    fullName: `Perf User ${id}`,
    role: role,
  });

  const regRes = http.post(`${baseUrl}/api/v1/users`, regPayload, {
    headers: defaultHeaders,
    tags: { name: 'RegisterUser' },
  });

  if (regRes.status !== 200 && regRes.status !== 201 && regRes.status !== 409) {
    return null;
  }

  // 2. Login to obtain JWT Bearer token
  const loginPayload = JSON.stringify({
    email: email,
    password: password,
  });

  const loginRes = http.post(`${baseUrl}/api/v1/login`, loginPayload, {
    headers: defaultHeaders,
    tags: { name: 'LoginUser' },
  });

  if (loginRes.status === 200) {
    try {
      const body = loginRes.json();
      return body.token || null;
    } catch (e) {
      return null;
    }
  }

  return null;
}
