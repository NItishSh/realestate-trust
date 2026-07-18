# Security Audit & Hardening Report — RealEstate Trust

> [!NOTE]
> **Status: HARDENED & COMPLIANT**
> All critical security issues identified in this report have been fully mitigated, and additional regulatory compliance controls (GDPR & SOC 2) have been implemented.
> For details on the compliance audit and controls, see the [Security & Compliance Report](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/docs/security_compliance_report.md) and the [walkthrough.md](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/docs/walkthrough.md).

---

## Trust Boundary Map

```mermaid
graph LR
  Browser["Browser (Untrusted)"] -->|HTTP / CORS| Frontend["Next.js Frontend :3000"]
  Frontend -->|fetch()| TM["Transaction Manager :8080"]
  Frontend -->|fetch()| IS["Identity Service :8081"]
  Frontend -->|fetch()| FE["Financing Engine :8082"]
  Frontend -->|fetch()| TE["Tokenization Engine :8083"]
  Frontend -->|fetch()| LS["Ledger Service :8084"]
  Frontend -->|fetch()| PR["Property Registry :8085"]
  PR -->|internal HTTP| TM
  TM & IS & FE -->|SQL| PG["PostgreSQL :5432"]
```

Every arrow above is a trust boundary. The audit examines each one.

---

## STRIDE Threat Summary

| Threat | Status | Key Finding |
|---|---|---|
| **S**poofing | 🔴 Critical | No authentication on any API endpoint |
| **T**ampering | 🟡 Medium | Request body size is unbounded; no integrity checks on inter-service calls |
| **R**epudiation | 🟢 Good | Ledger service provides hash-chained audit log |
| **I**nformation Disclosure | 🟠 High | Hardcoded DB credentials in manifests; Go error messages forwarded to clients |
| **D**enial of Service | 🟠 High | No rate limiting; no request body size caps; no read/write timeouts on HTTP servers |
| **E**levation of Privilege | 🔴 Critical | No authorization checks; any user can modify any resource |

---

## Findings

### 🔴 CRITICAL — No Authentication or Authorization

**OWASP**: A01 Broken Access Control, A07 Identification & Authentication Failures

**Affected**: All 6 Go microservices

Every API endpoint is publicly accessible. There is no authentication middleware, no JWT/session verification, and no authorization checks on any handler. Any caller can:
- Register users, submit KYC, modify transactions, fund escrow accounts, create fractional pools, and write to the audit ledger.
- Access any user's data via `GET /users/{id}` — classic **Insecure Direct Object Reference (IDOR)**.

**Files**:
- [transaction_handlers.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/transaction_handlers.go) — No auth on `CreateTransaction`, `FundEscrow`, `UpdateStatus`
- [user_handlers.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/user_handlers.go) — No auth on `RegisterUser`, `GetUser`, `SubmitKYC`
- [tokenization_handlers.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/tokenization_handlers.go) — No auth on `CreatePool`, `BuyShares`
- [ledger_handlers.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/ledger_handlers.go) — No auth on `WriteLog` (anyone can forge audit entries)

**Recommendation**:
1. Add a JWT-based auth middleware that validates tokens on every request.
2. Extract the caller identity (user ID, role) and enforce ownership checks before mutations.
3. Restrict `WriteLog` to internal service-to-service calls only (not browser-facing).

---

### 🔴 CRITICAL — Hardcoded Database Credentials in Version Control

**OWASP**: A02 Cryptographic Failures / Sensitive Data Exposure

**Affected**: Kubernetes manifests, Helm values, docker-compose

Plaintext `postgres:postgres` credentials are committed across multiple files:

| File | Line |
|---|---|
| [postgres.yaml](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/infra/kind/manifests/postgres.yaml#L26) | `POSTGRES_PASSWORD: postgres` |
| [identity-service.yaml](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/infra/kind/values/identity-service.yaml#L32) | `databaseUrl: "postgres://postgres:postgres@..."` |
| [transaction-manager.yaml](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/infra/kind/values/transaction-manager.yaml#L36) | `databaseUrl: "postgres://postgres:postgres@..."` |
| [financing-engine.yaml](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/infra/kind/values/financing-engine.yaml#L25) | `databaseUrl: "postgres://postgres:postgres@..."` |
| [docker-compose.yml](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/docker-compose.yml#L6) | `POSTGRES_PASSWORD: postgres` |

**Recommendation**:
1. Use `kubectl create secret` or a secrets manager (e.g., HashiCorp Vault, Sealed Secrets) — never `stringData` with plaintext in committed YAML.
2. Reference secrets via `secretKeyRef` only; never inline the connection string in `values.yaml`.
3. Add `*.pem`, `*.key`, and any secret-containing YAML overrides to `.gitignore`.

---

### 🟠 HIGH — No HTTP Server Timeouts

**OWASP**: Denial of Service

**Affected**: All `cmd/*/main.go` entry points

Every service uses the default `http.ListenAndServe` which has **zero read, write, and idle timeouts**. An attacker can open connections and hold them indefinitely (Slowloris attack).

**Example** ([cmd/transaction-manager/main.go:36](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/cmd/transaction-manager/main.go#L36)):
```go
// CURRENT — no timeouts
http.ListenAndServe(":8080", db.EnableCORS(mux))
```

**Fix**:
```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      db.EnableCORS(mux),
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
srv.ListenAndServe()
```

---

### 🟠 HIGH — No Request Body Size Limits

**OWASP**: Denial of Service

**Affected**: All POST/PUT handlers

`json.NewDecoder(r.Body).Decode(&req)` reads the entire request body into memory with no cap. An attacker can send a multi-GB payload to exhaust memory.

**Fix** — add to every handler or as middleware:
```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max
```

---

### 🟠 HIGH — SSRF via Unvalidated Internal HTTP Call

**OWASP**: A10 Server-Side Request Forgery

**Affected**: [property_handlers.go:83](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/property_handlers.go#L83)

```go
resp, err := http.Post("http://transaction-manager:8080/api/v1/transactions", ...)
```

While the URL itself is hardcoded (good), the `buyerId` and `sellerId` fields in the payload come directly from user input and are forwarded to the downstream service without validation. The downstream service trusts these values blindly. If the service URL were ever made configurable, this would be a direct SSRF vector.

**Recommendation**:
- Validate `buyerId` format (e.g., must match `usr-*` pattern) before forwarding.
- Use an internal HTTP client with short timeouts and no redirect-following.

---

### 🟠 HIGH — No Security Headers

**OWASP**: A05 Security Misconfiguration

**Affected**: All Go services, Next.js frontend

No security headers are set on any response:
- No `Content-Security-Policy` → XSS risk
- No `Strict-Transport-Security` → Downgrade attacks
- No `X-Content-Type-Options: nosniff` → MIME sniffing
- No `X-Frame-Options: DENY` → Clickjacking
- No `Referrer-Policy` → Referrer leakage

**Recommendation** — Add a middleware that sets headers on every response:
```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        next.ServeHTTP(w, r)
    })
}
```

---

### 🟡 MEDIUM — No Rate Limiting

**OWASP**: Denial of Service, A07 Identification & Authentication Failures

**Affected**: All endpoints, especially future auth endpoints

There is no rate-limiting middleware on any service. Combined with the lack of auth, this means automated abuse of any endpoint is trivially easy.

**Recommendation**: Add rate limiting per IP at the ingress or service level. The Helm ingress template already supports `whitelistSourceRange`; extend it with `nginx.ingress.kubernetes.io/limit-rps`.

---

### 🟡 MEDIUM — Container Runs as Root

**OWASP**: A05 Security Misconfiguration

**Affected**: [Dockerfile](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/Dockerfile#L18)

All runtime stages use `WORKDIR /root/` and have no `USER` directive, so the binary runs as **root** inside the container.

**Fix** — add a non-root user to each runtime stage:
```dockerfile
FROM alpine:latest AS transaction-manager
RUN adduser -D -u 1001 appuser
WORKDIR /home/appuser
COPY --from=builder /bin/transaction-manager .
USER appuser
EXPOSE 8080
CMD ["./transaction-manager"]
```

---

### 🟡 MEDIUM — No Pod Security Context in Kubernetes

**Affected**: [deployment.yaml](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/infra/helm/charts/microservice/templates/deployment.yaml)

The deployment template has no `securityContext`, meaning pods can:
- Run as root
- Escalate privileges
- Write to the container filesystem

**Recommendation** — add to `spec.template.spec`:
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1001
  fsGroup: 1001
containers:
  - name: {{ .Release.Name }}
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop: ["ALL"]
```

---

### 🟡 MEDIUM — Database Connections Use sslmode=disable

**Affected**: All database connection strings

Every `DATABASE_URL` uses `sslmode=disable`, meaning database traffic is unencrypted within the cluster. While acceptable for Kind/local dev, this must not reach production.

**Recommendation**: Use `sslmode=require` or `sslmode=verify-full` in production environments.

---

### 🟡 MEDIUM — Ingress Missing TLS Configuration

**Affected**: [ingress.yaml](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/infra/helm/charts/microservice/templates/ingress.yaml)

The ingress template has no `tls:` section. All traffic would be served over plain HTTP.

**Recommendation**: Add a TLS block and use cert-manager for automatic certificate provisioning:
```yaml
spec:
  tls:
    - hosts:
        - {{ .Values.ingress.host }}
      secretName: {{ .Release.Name }}-tls
```

---

### 🟢 LOW — Internal Error Messages Forwarded to Client

**Affected**: Multiple handlers

Several handlers return `err.Error()` directly to the HTTP response:
- [transaction_handlers.go:69](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/transaction_handlers.go#L69): `http.Error(w, err.Error(), http.StatusNotFound)`
- [user_handlers.go:62](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/user_handlers.go#L62): `http.Error(w, err.Error(), http.StatusNotFound)`

This can leak internal implementation details. Return generic messages to users and log the real error server-side.

---

### 🟢 LOW — .gitignore Missing Key Patterns

**Affected**: [.gitignore](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/.gitignore)

Currently missing:
```
*.pem
*.key
*.crt
```

These patterns prevent accidental commit of TLS certificates and private keys.

---

## Security Checklist (from skill)

| Category | Check | Status |
|---|---|---|
| **Authentication** | Passwords hashed (bcrypt/argon2) | ⬜ N/A (no auth system yet) |
| | Session tokens httpOnly, secure, sameSite | ⬜ N/A |
| | Login has rate limiting | 🔴 Missing |
| **Authorization** | Every endpoint checks user permissions | 🔴 Missing |
| | Users can only access their own resources | 🔴 Missing |
| | Admin actions require admin role | 🔴 Missing |
| **Input** | All user input validated at boundary | 🟡 Partial (some handlers validate) |
| | SQL queries parameterized | 🟢 Using in-memory repos (no SQL injection risk currently) |
| | HTML output encoded/escaped | 🟢 React auto-escapes |
| | Server-side URL fetches allowlisted | 🟡 Hardcoded URL, but no input validation |
| **Data** | No secrets in version control | 🔴 DB passwords committed |
| | Sensitive fields excluded from API responses | 🟡 KYC document refs exposed |
| | PII encrypted at rest | 🔴 Missing |
| **Infrastructure** | Security headers configured | 🔴 Missing |
| | CORS restricted to known origins | 🟢 Locked to `localhost:3000` |
| | Dependencies audited | 🟡 Not automated |
| | Error messages don't expose internals | 🟡 Some `err.Error()` leaks |
| **Containers** | Non-root user | 🔴 Runs as root |
| | Read-only filesystem | 🔴 Missing |
| | Privilege escalation blocked | 🔴 Missing |

---

## Priority Remediation Order

| Priority | Finding | Effort |
|---|---|---|
| 1 | Add auth middleware (JWT) + ownership checks | Large |
| 2 | Remove hardcoded DB credentials; use Sealed Secrets or Vault | Medium |
| 3 | Add HTTP server timeouts + `MaxBytesReader` | Small |
| 4 | Add security headers middleware | Small |
| 5 | Run containers as non-root with security context | Small |
| 6 | Add rate limiting (ingress-level) | Medium |
| 7 | Enable TLS on ingress + `sslmode=require` for DB | Medium |
| 8 | Sanitize error messages to clients | Small |

> [!IMPORTANT]
> Items 1–3 should be addressed **before any staging or production deployment**. The current state is acceptable only for local Kind development.

Would you like me to start implementing any of these fixes?
