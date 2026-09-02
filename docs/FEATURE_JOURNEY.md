# 🗺️ RealEstate-Trust: Feature Evolution & Architectural Journey

An authoritative reference detailing the complete evolutionary history, architectural milestones, supported features, and security posture of the **RealEstate-Trust Monorepo Platform** across its first 23 Pull Requests and major milestone releases.

---

## 📑 Table of Contents
1. [Executive Summary & System Vision](#-executive-summary--system-vision)
2. [Evolutionary Timeline](#-evolutionary-timeline)
3. [Chronological Engineering Phases](#-chronological-engineering-phases)
4. [Supported Feature & Capability Matrix](#-supported-feature--capability-matrix)
5. [Pull Request (PR) Index (PRs #1 – #23)](#-pull-request-pr-index-prs-1--23)
6. [Core Architectural Patterns Implemented](#-core-architectural-patterns-implemented)
7. [Verification, CI/CD & Testing Infrastructure](#-verification-cicd--testing-infrastructure)

---

## 🎯 Executive Summary & System Vision

**RealEstate-Trust** is an institutional-grade, zero-trust digital real estate tokenization and escrow transaction platform. Built as a Go-based microservice monorepo with a Next.js web application frontend, it provides cryptographically verified property escrows, tokenized fractional investments, automated bank financing workflows, and an immutable audit ledger.

```mermaid
flowchart TB
    subgraph ClientLayer["🌐 Client & Delivery Layer"]
        Frontend["Next.js Web Application<br/>(Port 3000 / UID 1001)"]
        CLI["re-cli Command Line Utility"]
    end

    subgraph GatewayLayer["🛡️ Security & Zero-Trust Perimeter"]
        CORS["Dynamic CORS Filter"]
        RateLimit["Memory Token-Bucket Rate Limiter"]
        JWKS["Keycloak OIDC / Dynamic JWKS Client"]
        RBAC_ABAC["RBAC Role Gates + ABAC Ownership"]
    end

    subgraph Services["⚙️ Domain Microservices (Distroless / UID 65532)"]
        Identity["identity-service<br/>(:8081)"]
        TxManager["transaction-manager<br/>(:8080)"]
        Property["property-registry-service<br/>(:8085)"]
        Financing["financing-engine<br/>(:8082)"]
        Tokenization["tokenization-engine<br/>(:8083)"]
        Ledger["ledger-service<br/>(:8084)"]
        Feedback["feedback-service<br/>(:8086)"]
    end

    subgraph EventMesh["📬 Reliable Event Mesh"]
        Outbox["Transactional Outbox Table"]
        Relayer["Background Outbox Relayer"]
        RabbitMQ["RabbitMQ AMQP Broker<br/>(Topic: realestate.events)"]
    end

    subgraph DataLayer["💾 Persistence & Cryptography"]
        Postgres[("PostgreSQL Multi-Database<br/>(7 Isolated Schemas)")]
        Vault["HashiCorp Vault KMS / Local AES-GCM"]
    end

    ClientLayer --> GatewayLayer
    GatewayLayer --> Services
    TxManager --> Outbox
    Outbox --> Relayer --> RabbitMQ
    RabbitMQ --> Ledger
    Services --> Postgres
    Identity --> Vault
```

---

## ⏳ Evolutionary Timeline

```mermaid
timeline
    title RealEstate-Trust Platform Evolutionary Journey
    Phase 1 : Monorepo Scaffolding : 7 Go Microservices : Next.js Frontend : Echo Web Framework
    Phase 2 : Distributed Storage & Schema : 7 Isolated PostgreSQL Schemas : Idempotent Auto-Migrations : Golangci-lint
    Phase 3 : Cloud-Native & GitOps : Terraform Kind Provisioner : ArgoCD App-of-Apps : Keycloak IAM OIDC : External Secrets
    Phase 4 : Event-Driven Architecture : Transactional Outbox Pattern : Outbox Relay Daemon : CloudEvents Spec : AMQP Broker
    Phase 5 : Data Integrity & Concurrency : Distributed Ledger Idempotency : Row-Level Locking : Fail-Closed KMS Encryption
    Phase 6 : Production Hardening & Zero-Trust : HTTP Client Timeouts : Restricted Pod Security : RBAC/ABAC Gates : Rate Limiting
```

---

## 🚀 Chronological Engineering Phases

### **Phase 1: Monorepo Foundation & Microservices Scaffolding**
- **Focus**: Establishing clean domain microservice boundaries using Go, Echo, and a Next.js frontend.
- **Milestones**:
  - Structured 7 independent services (`identity-service`, `transaction-manager`, `property-registry-service`, `financing-engine`, `tokenization-engine`, `ledger-service`, `feedback-service`) and a shared `internal/` core layer.
  - Implemented initial domain entities, state machines, and REST API handlers.
  - Formatted codebase and established `golangci-lint` code quality baseline ([PR #1](https://github.com/NItishSh/realestate-trust/pull/1)).
  - **Release**: `v1.0.0` ([PR #2](https://github.com/NItishSh/realestate-trust/pull/2)).

---

### **Phase 2: Cloud-Native Infrastructure & GitOps Integration**
- **Focus**: Production-ready containerization, Kubernetes packaging, and GitOps pipelines.
- **Milestones**:
  - Upgraded External Secrets Operator to `v1beta1` ([PR #3](https://github.com/NItishSh/realestate-trust/pull/3)).
  - Standardized microservice Helm charts (`infra/helm/charts/microservice`) with dynamic smoke test jobs ([PR #5](https://github.com/NItishSh/realestate-trust/pull/5), [PR #6](https://github.com/NItishSh/realestate-trust/pull/6)).
  - Automated local Kubernetes cluster provisioning with Terraform (`infra/terraform`) and ArgoCD App-of-Apps GitOps integration ([PR #7](https://github.com/NItishSh/realestate-trust/pull/7)).
  - Configured Keycloak IAM integration for OpenID Connect (OIDC) realm provisioning.
  - Target revision GitOps alignment to `main` ([PR #9](https://github.com/NItishSh/realestate-trust/pull/9)).
  - Added ESO webhook warm-up retry logic and created the Cloud KMS Production Porting Guide ([PR #11](https://github.com/NItishSh/realestate-trust/pull/11)).
  - **Releases**: `v1.0.1` ([PR #4](https://github.com/NItishSh/realestate-trust/pull/4)), `v1.1.0` ([PR #8](https://github.com/NItishSh/realestate-trust/pull/8)), `v1.1.1` ([PR #10](https://github.com/NItishSh/realestate-trust/pull/10)), `v1.1.2` ([PR #12](https://github.com/NItishSh/realestate-trust/pull/12)).

---

### **Phase 3: Data Consistency & Distributed Concurrency Remediation**
- **Focus**: Eliminating distributed data races, race conditions, and event loss.
- **Milestones**:
  - **Distributed Ledger Idempotency**: Added database-level unique constraint on `(transaction_id, entry_type)` with atomic `ON CONFLICT DO NOTHING` handling to prevent duplicate ledger transactions under concurrent webhook replays ([PR #13](https://github.com/NItishSh/realestate-trust/pull/13)).
  - **Transactional Outbox Pattern**: Unified transaction state changes and outbox event logging into a single atomic PostgreSQL transaction (`BEGIN ... INSERT outbox_events ... COMMIT`), ending dual-write hazards ([PR #13](https://github.com/NItishSh/realestate-trust/pull/13)).
  - **Tokenization Concurrency Guards**: Applied `SELECT ... FOR UPDATE` row-level locking on fractional pools to prevent overselling shares during parallel buyer requests ([PR #13](https://github.com/NItishSh/realestate-trust/pull/13)).
  - **Standardized CloudEvents v1.0**: Standardized JSON payload serialization across the asynchronous message mesh ([PR #13](https://github.com/NItishSh/realestate-trust/pull/13)).
  - **Release**: `v1.2.0` ([PR #14](https://github.com/NItishSh/realestate-trust/pull/14)).

---

### **Phase 4: Cryptographic Security, KMS & Config Reloading**
- **Focus**: Zero-trust key management and automated configuration management.
- **Milestones**:
  - **Fail-Closed Production KMS**: Hardened KYC data encryption so that in `APP_ENV=production`, missing or unreachable HashiCorp Vault Transit KMS immediately fails closed without silent fallback to local dev keys ([PR #16](https://github.com/NItishSh/realestate-trust/pull/16)).
  - **Zero-Downtime Secret Rotation**: Integrated Stakater Reloader annotations (`reloader.stakater.com/auto: "true"`) into Helm deployment templates for rolling pod reloads on Secret/ConfigMap updates ([PR #16](https://github.com/NItishSh/realestate-trust/pull/16)).
  - **Automated E2E Suite Target**: Built `make compose-test-e2e` for spinning up the full 7-migration, 7-service stack and validating deal lifecycle transitions in CI ([PR #15](https://github.com/NItishSh/realestate-trust/pull/15), [PR #18](https://github.com/NItishSh/realestate-trust/pull/18)).
  - **Release**: `v1.3.0` ([PR #17](https://github.com/NItishSh/realestate-trust/pull/17)).

---

### **Phase 5: Production Code Auditing & Resilience**
- **Focus**: Eliminating socket leaks, unauthenticated endpoints, and unbounded input fields.
- **Milestones**:
  - **HTTP Client Timeouts & Resource Management**: Replaced default unbound HTTP clients with a 5-second timeout client and guaranteed `defer resp.Body.Close()` in `property_handlers.go` ([PR #19](https://github.com/NItishSh/realestate-trust/pull/19)).
  - **Bank Webhook Security**: Added `X-Webhook-Secret` verification to `POST /api/v1/loans/webhooks/bank` against `BANK_WEBHOOK_SECRET` ([PR #19](https://github.com/NItishSh/realestate-trust/pull/19)).
  - **Input Boundary Enforcement**: Enforced strict 1–5 range bounds on rating submissions in `feedback_handlers.go` ([PR #19](https://github.com/NItishSh/realestate-trust/pull/19)).
  - **Release**: `v1.4.0` ([PR #20](https://github.com/NItishSh/realestate-trust/pull/20)).

---

### **Phase 6: Zero-Trust Perimeter, RBAC/ABAC & Container Security**
- **Focus**: Hardening container profiles against kernel exploitation and implementing granular authorization gates.
- **Milestones**:
  - **Auth Rate Limiting**: Added token-bucket memory rate limiter (`10 req/s, Burst: 20`) on registration and login endpoints ([PR #21](https://github.com/NItishSh/realestate-trust/pull/21)).
  - **Dynamic CORS**: Replaced wildcard origins with comma-separated `CORS_ALLOWED_ORIGINS` across all 7 services ([PR #21](https://github.com/NItishSh/realestate-trust/pull/21)).
  - **Kubernetes Restricted Pod Security Standards**: Enforced `seccompProfile: { type: RuntimeDefault }`, `runAsUser: 65532` (Distroless `nonroot`), `readOnlyRootFilesystem: true`, and `capabilities: { drop: ["ALL"] }` across Helm charts ([PR #22](https://github.com/NItishSh/realestate-trust/pull/22)).
  - **Comprehensive RBAC & ABAC Gates**: Enforced role-based route middleware (`db.RBACMiddleware`) and attribute ownership validation (`sub == id || role == ADMIN`) across all endpoints ([PR #23](https://github.com/NItishSh/realestate-trust/pull/23)).
  - **XSS Sanitization**: Added `core.SanitizeString` to sanitize text inputs against HTML injection ([PR #23](https://github.com/NItishSh/realestate-trust/pull/23)).

---

## 📊 Supported Feature & Capability Matrix

| Feature | Owning Service | Port | Primary Endpoints | Storage Table | Security & Access Policy |
| :--- | :--- | :---: | :--- | :--- | :--- |
| **User Registration & Login** | `identity-service` | `8081` | `POST /api/v1/users`<br>`POST /api/v1/login` | `users`, `user_sessions` | Rate-limited (10 rps), BCrypt (Cost 12), Email regex |
| **KYC Identity Verification** | `identity-service` | `8081` | `POST /api/v1/users/:id/kyc`<br>`GET /api/v1/users/:id/kyc/status` | `kyc_verifications` | ABAC Owner/Admin check, HashiCorp Vault KMS encryption |
| **Escrow Transaction Management** | `transaction-manager` | `8080` | `POST /api/v1/transactions`<br>`PUT /api/v1/transactions/:id/status`<br>`POST /api/v1/transactions/:id/escrow/fund` | `transactions`, `escrow_accounts`, `outbox_events` | RBAC (Buyer/Seller/Broker/Officer/Admin), Outbox Relay Atomicity |
| **Property Title & Documents** | `property-registry-service` | `8085` | `POST /api/v1/properties`<br>`GET /api/v1/properties`<br>`POST /api/v1/properties/:id/insurance/verify`<br>`POST /api/v1/properties/:id/documents/unlock` | `properties`, `property_documents` | RBAC (Seller/Broker/Officer/Admin), HTTP client timeout |
| **Mortgage & Loan Underwriting** | `financing-engine` | `8082` | `POST /api/v1/loans`<br>`GET /api/v1/loans`<br>`POST /api/v1/loans/:id/disburse`<br>`POST /api/v1/loans/webhooks/bank` | `loan_applications` | LTV ratio validation (≤80%), `X-Webhook-Secret` verification, Officer disbursal gate |
| **Fractional Asset Tokenization** | `tokenization-engine` | `8083` | `POST /api/v1/pools`<br>`GET /api/v1/pools`<br>`POST /api/v1/pools/:id/buy` | `fractional_pools`, `fractional_holdings` | `SELECT FOR UPDATE` row locks, RBAC Seller/Buyer |
| **Cryptographic Immutable Ledger** | `ledger-service` | `8084` | `POST /api/v1/entries`<br>`GET /api/v1/entries` | `ledger_entries` | SHA-256 Merkle chain, Unique compound idempotency key |
| **Broker Reputation & Feedback** | `feedback-service` | `8086` | `POST /api/v1/reviews`<br>`GET /api/v1/brokers/:id/rating` | `broker_feedback` | 1–5 rating validation, CORS origin restriction |

---

## 📦 Pull Request (PR) Index (PRs #1 – #23)

| PR # | Type | Title | Key Files Changed | Impact |
| :---: | :---: | :--- | :--- | :--- |
| **[#1](https://github.com/NItishSh/realestate-trust/pull/1)** | `fix` | resolve remaining errcheck and govet issues | `cmd/*/main.go`, `internal/db/*.go` | Zero golangci-lint errors baseline |
| **[#2](https://github.com/NItishSh/realestate-trust/pull/2)** | `release` | Release 1.0.0 | Monorepo root | Initial stable version tag |
| **[#3](https://github.com/NItishSh/realestate-trust/pull/3)** | `fix` | update external-secrets apiVersion to v1beta1 | `infra/gitops/secret-store.yaml` | Kubernetes 1.28+ compatibility |
| **[#4](https://github.com/NItishSh/realestate-trust/pull/4)** | `release` | Release 1.0.1 | Monorepo root | Patch version release |
| **[#5](https://github.com/NItishSh/realestate-trust/pull/5)** | `fix` | Fix smoke test script ports to use default port 80 | `infra/helm/charts/microservice/` | ClusterIP Service routing fix |
| **[#6](https://github.com/NItishSh/realestate-trust/pull/6)** | `fix` | Fix local helm values smoke test ports | `infra/gitops/service-apps.yaml` | In-cluster smoke test stability |
| **[#7](https://github.com/NItishSh/realestate-trust/pull/7)** | `feat` | Terraform Kind provisioning, Keycloak IAM GitOps integration | `infra/terraform/`, `infra/gitops/` | Automated 1-click local Kind infrastructure |
| **[#8](https://github.com/NItishSh/realestate-trust/pull/8)** | `release` | Release 1.1.0 | Monorepo root | Minor version release |
| **[#9](https://github.com/NItishSh/realestate-trust/pull/9)** | `fix` | update targetRevision from feature branch to main | `infra/gitops/service-apps.yaml` | GitOps branch consistency |
| **[#10](https://github.com/NItishSh/realestate-trust/pull/10)** | `release` | Release 1.1.1 | Monorepo root | Patch version release |
| **[#11](https://github.com/NItishSh/realestate-trust/pull/11)** | `fix` | add retry loop for ESO webhook warm-up and add production porting guide | `infra/terraform/`, `docs/` | Flakeless ESO provisioning + AWS/GCP guide |
| **[#12](https://github.com/NItishSh/realestate-trust/pull/12)** | `release` | Release 1.1.2 | Monorepo root | Patch version release |
| **[#13](https://github.com/NItishSh/realestate-trust/pull/13)** | `feat` | remediate critical data consistency and distributed concurrency gaps (Phase 1 & 2) | `internal/db/`, `internal/events/` | Outbox atomicity, Ledger write idempotency, Pool row locks |
| **[#14](https://github.com/NItishSh/realestate-trust/pull/14)** | `release` | Release 1.2.0 | Monorepo root | Minor version release |
| **[#15](https://github.com/NItishSh/realestate-trust/pull/15)** | `docs` | update architectural gaps status and add compose-test-e2e make target | `docs/`, `Makefile` | Automated compose E2E target |
| **[#16](https://github.com/NItishSh/realestate-trust/pull/16)** | `feat` | complete Phase 3 gaps (GAP-08 fail-closed KMS, GAP-10 reloader rotation) | `internal/db/crypto.go`, `infra/helm/` | Zero-downtime config reloads + Fail-closed Vault KMS |
| **[#17](https://github.com/NItishSh/realestate-trust/pull/17)** | `release` | Release 1.3.0 | Monorepo root | Minor version release |
| **[#18](https://github.com/NItishSh/realestate-trust/pull/18)** | `fix` | make compose E2E test in CI deterministic and fix postgres healthcheck | `Makefile`, `docker-compose.yaml` | Flakeless Docker Compose E2E in GitHub Actions |
| **[#19](https://github.com/NItishSh/realestate-trust/pull/19)** | `feat` | harden HTTP client timeouts, webhook authentication, and input boundaries | `internal/db/*.go` | Sockets closed, bank webhook secret auth, rating boundaries |
| **[#20](https://github.com/NItishSh/realestate-trust/pull/20)** | `release` | Release 1.4.0 | Monorepo root | Minor version release |
| **[#21](https://github.com/NItishSh/realestate-trust/pull/21)** | `feat` | add auth rate limiting and dynamic CORS configuration | `internal/db/middleware.go`, `cmd/*/` | Sensitive endpoint rate limits (10 rps) & dynamic CORS |
| **[#22](https://github.com/NItishSh/realestate-trust/pull/22)** | `feat` | enforce Kubernetes Restricted Pod Security Standards across Helm microservice charts | `infra/helm/charts/microservice/` | `seccompProfile: RuntimeDefault`, `runAsUser: 65532`, `drop: [ALL]` |
| **[#23](https://github.com/NItishSh/realestate-trust/pull/23)** | `feat` | enforce RBAC/ABAC role gates, attribute ownership, and input sanitization | `cmd/*/main.go`, `internal/db/`, `internal/core/` | Zero-trust RBAC route middleware, ABAC ownership, XSS escaping |
| **[#24](https://github.com/NItishSh/realestate-trust/pull/24)** | `docs` | create comprehensive feature evolution and architectural journey guide | `docs/FEATURE_JOURNEY.md` | Authoritative engineering history and capability catalog |
| **[#25](https://github.com/NItishSh/realestate-trust/pull/25)** | `feat` | implement typed sentinel domain errors, enterprise HSTS/CSP security headers | `internal/core/errors.go`, `internal/db/`, `cmd/*/` | Idiomatic `errors.Is` error checking, HSTS/CSP middleware |
| **[#26](https://github.com/NItishSh/realestate-trust/pull/26)** | `feat` | implement centralized service config and deep Kubernetes readiness probes | `internal/core/config.go`, `internal/db/health.go`, `infra/helm/` | Dynamic Liveness/Readiness probes (`/health/live`, `/health/ready`) |

---

## 🏛️ Core Architectural Patterns Implemented

### 1. **Transactional Outbox Pattern**
To prevent dual-write inconsistencies between PostgreSQL and RabbitMQ, all transaction status modifications write an `outbox_events` row inside the **same atomic database transaction**. A dedicated background daemon (`OutboxProcessor`) polls pending events, transmits them over AMQP, and marks them `PROCESSED`.

### 2. **Distributed Ledger Idempotency & Hash Chaining**
Every state change creates an immutable `ledger_entries` record. Write idempotency is guaranteed by a compound unique constraint on `(transaction_id, entry_type)` with `ON CONFLICT DO NOTHING`. Each entry stores `previous_hash` and `current_hash` computed over `(prev_hash + tx_id + amount + timestamp)` creating a tamper-evident cryptographic chain.

### 3. **Dual-Mode Cryptographic KMS**
Sensitive KYC records (passports, national IDs) are encrypted via HashiCorp Vault Transit KMS. In local development (`APP_ENV != production`), an automatic fallback to local AES-256-GCM is permitted. In production (`APP_ENV=production`), the system **fails closed** if Vault Transit KMS is unreachable.

### 4. **Keycloak JWKS Verification with Dual-Mode Signature Support**
Supports both RS256/ES256 asymmetric cryptographic verification via Keycloak's dynamic JSON Web Key Set (`/realms/realestate-trust/protocol/openid-connect/certs`) and HS256 HMAC shared secrets for isolated microservice tests.

### 5. **Kubernetes Restricted Pod Security Standard (PSS)**
Workload pods run under Google Distroless with zero Linux capabilities (`drop: ["ALL"]`), read-only root filesystems (`readOnlyRootFilesystem: true`), non-root UID `65532`, and kernel syscall filtering (`seccompProfile: { type: RuntimeDefault }`).

---

## 🧪 Verification, CI/CD & Testing Infrastructure

The repository enforces strict multi-layered automated verification across all layers of the monorepo:

```bash
# 1. Monorepo Linter Suite (golangci-lint, hadolint, terraform validate, tflint, helm lint)
make lint

# 2. Complete Unit & Contract Test Suite
go test -v ./...

# 3. Full Local Docker Compose Stack & End-to-End Deal Lifecycle Test
make compose-test-e2e

# 4. Terraform Kind Cluster Provisioning & GitOps Bootstrap
make terraform-apply
```

### **GitHub Actions Automation**
Every Pull Request triggers:
1. **`Code Sanity Checks`**: Go unit tests, code formatting, `golangci-lint`, Helm lint, Dockerfile lint, Terraform validation.
2. **`Verify Docker Compose Stack`**: Boots all 7 migrations, 7 microservices, RabbitMQ, PostgreSQL, and executes the complete E2E deal lifecycle test suite in clean isolated containers.
3. **`Lint Pull Request Title`**: Semantic Conventional Commits enforcement.
4. **`Release Please`**: Automated SemVer tagging, changelog generation, and GitHub Release drafting.
