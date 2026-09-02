# Real Estate Trust & Escrow Platform (Go Monorepo)

A production-grade, highly secure, and containerized microservices platform designed to orchestrate high-value capital movements, compliance gatechecks (KYC/AML), mortgage lending pipelines, property fractionalization, and immutable cryptographic ledgers.

Built in Go (**v1.26.5**) and Next.js, and designed to deploy natively to Kubernetes via Helm and ArgoCD GitOps pipelines.

---

## 1. System Architecture & Microservices

The platform is structured as a single monorepo comprising **7 isolated Go microservices**, a **Next.js frontend**, and an administrative **support CLI**:

1. **Transaction & Escrow Manager** (`cmd/transaction-manager` | `:8080`): Orchestrates deal lifecycles (`DRAFT` $\rightarrow$ `ESCROW` $\rightarrow$ `FUNDED` $\rightarrow$ `CLOSED`), manages virtual escrow accounts, and dispatches events via the Transactional Outbox pattern.
2. **User & Identity Service** (`cmd/identity-service` | `:8081`): Manages user authentication, JWT JWKS issuer/verifier, role configurations (`BUYER`, `SELLER`, `BROKER`, `OFFICER`, `ADMIN`), and Vault Transit KMS AES-256-GCM encrypted KYC/AML compliance.
3. **Embedded Financing Engine** (`cmd/financing-engine` | `:8082`): Interfaces with lenders/NBFCs to underwrite mortgages, verifies LTV limits ($\le 80\%$), validates HMAC bank webhooks, and triggers automated loan disbursements.
4. **Fractional Tokenization Engine** (`cmd/tokenization-engine` | `:8083`): Handles property fractionalization pool creation, equity digital share bounds, and RBAC-restricted investment holdings.
5. **Immutable Audit Ledger** (`cmd/ledger-service` | `:8084`): Implements an immutable SHA-256 cryptographic hash-chained ledger, providing tamper-evident audit trails with compound idempotency.
6. **Property Registry Service** (`cmd/property-registry-service` | `:8085`): Manages verified land registry titles, municipal boundary validation, and deed ownership transfers.
7. **Feedback & Reputation Service** (`cmd/feedback-service` | `:8086`): Collects and moderates post-transaction stakeholder reviews with strict 1–5 rating validation.

> 📖 **Read the Definitive Architecture Guides:**
> - [Feature Evolution & Engineering Journey](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/docs/FEATURE_JOURNEY.md) (All PRs, capabilities, and design patterns)
> - [System Manual](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/docs/system_manual.md) (Deep dive into state machines, deployment patterns, and security models)
> - [Microservices Catalog](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/docs/microservices.md) (Detailed API contracts, tables, and schemas)

---

## 2. Directory Layout

```
realestate-trust/
├── .github/                      # CI/CD Workflows (GitHub Actions)
│   └── workflows/
│       ├── ci.yml                # Parallel matrix test, build, scan, sign pipeline
│       ├── finops-rightsize.yaml # Automated weekly VPA FinOps right-sizing pipeline
│       └── release.yml           # Automated SemVer Release Please pipeline
├── cmd/                          # Service entrypoints (main.go)
│   ├── feedback-service/         # Port :8086
│   ├── financing-engine/         # Port :8082
│   ├── identity-service/         # Port :8081
│   ├── ledger-service/           # Port :8084
│   ├── property-registry-service/# Port :8085
│   ├── re-cli/                   # Support, RCA, and FinOps CLI tool
│   ├── tokenization-engine/      # Port :8083
│   └── transaction-manager/      # Port :8080
├── docs/                         # Architecture, research, and analysis documentation
│   ├── FEATURE_JOURNEY.md        # Definitive engineering and capability roadmap
│   ├── microservices.md          # Detailed microservice catalog
│   └── system_manual.md          # Definitive System Manual and architectural guide
├── frontend/                     # Next.js 14 React Web Application (Port :3000)
├── infra/                        # Infrastructure & GitOps configurations
│   ├── gitops/                   # Declarative ArgoCD App-of-Apps manifests
│   ├── helm/                     # Reusable microservice chart with VPA, KEDA, and NetworkPolicies
│   ├── kind/                     # Local Kubernetes (Kind) values and bootstrap scripts
│   └── terraform/                # Infrastructure as Code (Terraform)
├── internal/                     # Private shared Go modules
│   ├── core/                     # Domain entities, validation boundaries, typed errors, config
│   ├── db/                       # Decoupled repository queries, REST handlers, middleware
│   └── events/                   # CloudEvents 1.0 schemas and RabbitMQ broker clients
├── test/                         # Integration, Contract, and E2E Test Suites
│   ├── contract/                 # OpenAPI 3.0 contract compliance tests
│   └── e2e/                      # Full deal lifecycle end-to-end test suite
├── Dockerfile                    # Multi-stage, multi-target distroless image builder
├── go.mod                        # Go dependency manifest (Go 1.26.5)
└── Makefile                      # Developer workflow orchestration
```

---

## 3. Core Architectural Patterns

* **Transactional Outbox Pattern**: Atomic dual-write prevention between PostgreSQL and RabbitMQ CloudEvents relay.
* **Dual-Mode Cryptographic KMS**: Envelope encryption with HashiCorp Vault Transit KMS in production (`AES-256-GCM`), with safe local fallback.
* **Role-Based & Attribute-Based Access Control (RBAC/ABAC)**: Zero-trust route middleware enforcing role gates (`BUYER`, `SELLER`, `BROKER`, `OFFICER`, `ADMIN`) and resource ownership.
* **Enterprise Security Headers**: Strict HSTS (`max-age=31536000`), CSP (`default-src 'self'`), `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`.
* **Deep Kubernetes Health Probes**: Decoupled `/api/v1/health/live` (process liveness) and `/api/v1/health/ready` (active DB pool ping and RabbitMQ connection checks).
* **Automated FinOps Right-Sizing**: Ingests historical VPA telemetry to auto-tune Helm resource requests with a +20% safety margin and memory limit safeguards.

---

## 4. Getting Started

### Prerequisites
* Go compiler (**v1.26.5** or newer)
* Docker & Docker Compose
* Helm & Kubectl
* [pre-commit](https://pre-commit.com/) (`pre-commit install && pre-commit install --hook-type commit-msg`)

### Running Tests
```bash
# Run all unit and contract tests monorepo-wide
go test -v -race ./...

# Run linters (golangci-lint, helm lint, hadolint, terraform validate)
make lint

# Run end-to-end deal lifecycle test suite in Docker Compose
make compose-test-e2e
```

### Local Development (Docker Compose)
The fastest way to spin up the entire application stack:
```bash
make compose-up
```
Exposes:
* **Frontend UI**: `http://localhost:3000`
* **Transaction Manager**: `:8080`
* **Identity Service**: `:8081`
* **Financing Engine**: `:8082`
* **Tokenization Engine**: `:8083`
* **Ledger Service**: `:8084`
* **Property Registry**: `:8085`
* **Feedback Service**: `:8086`
* **PostgreSQL Database**: `:5432`
* **RabbitMQ Management**: `:15672` (guest / guest)

---

## 5. Kubernetes, GitOps & FinOps

### Kind Cluster Bootstrap
```bash
make cluster-up
```
Provisions **Istio Service Mesh**, **HashiCorp Vault**, **External Secrets Operator (ESO)**, **Keycloak IAM**, **Prometheus**, **Grafana**, **Loki**, **Tempo**, **OpenCost**, and bootstraps **ArgoCD**.

### Port-Forwarding & UIs
```bash
make port-forward-argocd   # ArgoCD UI on https://localhost:8080
make port-forward-grafana  # Grafana Dashboards on http://localhost:3001
make port-forward-opencost # OpenCost FinOps UI on http://localhost:9003
make port-forward-keycloak # Keycloak IAM on http://localhost:8088
make port-forward-kiali    # Kiali Service Mesh on http://localhost:20001
make port-forward-vault    # HashiCorp Vault on http://localhost:8200
```

### FinOps Workload Right-Sizing
```bash
# Preview right-sizing suggestions natively in Go:
make finops-rightsize-dryrun

# Apply right-sized allocations with +20% safety buffer to Helm values:
make finops-rightsize
```

---

## 6. Support CLI (`re-cli`)

The platform includes a native Go CLI utility in `cmd/re-cli` for SRE/support operations and FinOps automation:

```bash
# Build the CLI
go build -o re-cli ./cmd/re-cli

# Fetch transaction status
./re-cli transaction inspect <txn_id>

# Check underwriting financing status
./re-cli finance check <loan_id>

# Verify cryptographic ledger integrity
./re-cli ledger verify

# Run FinOps right-sizing
./re-cli finops rightsize --dry-run
```
