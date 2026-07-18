# Walkthrough — Reliability, Persistence, Istio Gateway, Log Correlation, CORS & Regulatory Compliance

We have successfully completed all core architectural and regulatory requirements to make the `realestate-trust` platform production-ready, including Phase 4: Request Tracing & Log Correlation, CORS resolutions, and Phase 5 compliance controls for GDPR & SOC 2.

---

## 🛠️ Changes Implemented

### 1. Database-Level Persistence & Migrations
- **SQL Repositories**: Wrote Postgres-backed repository implementations for all domains (users, transactions, escrow, financing, ledger, properties, tokenization, and feedback).
- **Idempotence**: Refactored seed logic to be idempotent, preventing duplicate key errors on database container restarts.

### 2. Queue Reliability & Resiliency
- **Dead Letter Queue (DLQ)**: Declared DLX (`transaction-events-dlx`) and DLQ (`transaction-events-dlq`) for processing resilience.
- **Publisher Confirms**: Implemented publisher confirms waiting for broker receipt via `ch.NotifyPublish()`.
- **Consumer Idempotency**: Gracefully handled unique constraint violations inside `ledger-service` to safely prune duplicate events.

### 3. Edge Routing with Istio Service Mesh (API Gateway)
- **Unified Ingress Gateway**: Consolidated separate service ports (8080-8086) to route via the Istio Ingress Gateway (`istio-ingress`) on port `8080` (NodePort `30080`).
- **Namespace Injection**: Labeled the `realestate-trust` namespace for automatic Envoy sidecar proxy injection.
- **Permissive mTLS & Plaintext Destinations**: Configured PeerAuthentication `mode: PERMISSIVE` and created `DestinationRules` to disable TLS for database (`postgres`) and broker (`rabbitmq`) pods.

### 4. End-to-End Tracing & Log Correlation
- **Frontend Client**: Updated `frontend/src/lib/api.ts` to automatically generate a unique request Correlation ID using the browser's crypto API (`crypto.randomUUID()`) and inject it as a header: `X-Correlation-ID`.
- **Go Middlewares**: Created `CorrelationIDMiddleware()` and `RequestLoggerMiddleware()` inside `internal/db/middleware.go` to automatically extract the request header, inject it into the request context, set the response header, and output structured JSON request logs.
- **Structured slog Integration**: Implemented a custom `SlogCorrelationHandler` that automatically retrieves the correlation ID from context values and adds it as a first-class structured log attribute (`"correlation_id"`).
- **RabbitMQ Context Propagation**: Modified events `Publish` and `Consume` to pass correlation IDs via message headers (`"correlation_id"`), ensuring end-to-end tracing context carries across asynchronous message boundaries.

### 5. Access Control (CORS) Fixes
- **Unified Origin Support**: Updated all 7 microservices' `main.go` entrypoints to allow `http://localhost:8080` (the unified Istio Edge port) in addition to `http://localhost:3000`.
- **CORS Allowed Headers**: Added the custom `"X-Correlation-ID"` header to each service's `AllowHeaders` config so that the browser does not block preflight precheck requests.

### 6. Regulatory Compliance Controls (GDPR & SOC 2)
- **Application-Level Encryption**: Implemented AES-256-GCM encryption/decryption utilities in [crypto.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/crypto.go).
- **KYC Safeguards**: Updated `SubmitKYC` inside [user_repository_sql.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/user_repository_sql.go) to automatically encrypt government document references using the AES-256-GCM wrapper prior to persisting them. Decryption is performed upon retrieval so that API outputs remain transparent.
- **GDPR Right to be Forgotten**:
  - Implemented `DeleteUser` database methods inside `SQLUserRepository` executing a cascade deletion of the target user ID.
  - Exposed a `DELETE /api/v1/users/:id` route inside `identity-service` that validates authenticated JWT token claims. The route restricts deletion exclusively to the owning user or an administrator.

---

## 🧪 Verification Results

### 1. Unit & Integration Tests
All tests compile and pass successfully, including a new `TestDeleteUser` integration test validating JWT owner authentication and `TestKYCEncryption` verifying encryption and decryption of string references:
```bash
go test -v ./internal/db/...
```
**Output:**
```
=== RUN   TestUserHandlersIntegration
--- PASS: TestUserHandlersIntegration (0.06s)
=== RUN   TestDeleteUser
=== RUN   TestDeleteUser/Forbidden_Delete
=== RUN   TestDeleteUser/Delete_Own_Profile
--- PASS: TestDeleteUser (0.00s)
    --- PASS: TestDeleteUser/Forbidden_Delete (0.00s)
    --- PASS: TestDeleteUser/Delete_Own_Profile (0.00s)
=== RUN   TestKYCEncryption
--- PASS: TestKYCEncryption (0.00s)
PASS
ok  	github.com/realestate-trust/monorepo/internal/db	1.074s
```

### 2. Live Database Encryption Verification
Submitting a KYC verification through the API gateway (`POST /api/v1/users/{id}/kyc`) with reference value `"PASSPORT-98765-SECRET"` yields a base64 encrypted string inside Postgres:
```sql
SELECT document_reference FROM kyc_verifications WHERE user_id = 'usr-delete-me-test@example.com';
```
**Database Value:**
`WCg07cbmGC2JAqV0ws1HrNnA49OyVvUdmCtG6HvcfkzesOb1o4Iokt0esz0vpNIOhA==`

### 3. GDPR Purge Verification (Right to Erasure)
Executing `DELETE /api/v1/users/usr-delete-me-test@example.com` successfully purged the records from both the `users` and `kyc_verifications` tables (under foreign key cascade settings):
```sql
SELECT count(*) FROM users WHERE id = 'usr-delete-me-test@example.com'; -- Returns 0
SELECT count(*) FROM kyc_verifications WHERE user_id = 'usr-delete-me-test@example.com'; -- Returns 0
```
All checks passed successfully!
