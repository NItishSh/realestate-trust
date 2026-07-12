# Real Estate Trust & Escrow Platform (Go Monorepo)

A production-grade, highly secure, and containerized microservices platform designed to orchestrate high-value capital movements, compliance gatechecks (KYC/AML), mortgage lending pipelines, and property fractionalization.

Built in Go (**v1.26.5**) and designed to deploy natively to Kubernetes via Helm and ArgoCD GitOps pipelines.

---

## 1. System Architecture

The platform is structured as a single monorepo comprising five isolated Go microservices:

1. **Transaction & Escrow Manager** (`cmd/transaction-manager`): Orchestrates deal lifecycles (Draft -> Escrow -> Funded -> Closed) and manages virtual account integrations.
2. **User & Identity Service** (`cmd/identity-service`): Manages user registration, profiles, role configurations, and document-backed KYC/AML compliance checks.
3. **Embedded Financing Engine** (`cmd/financing-engine`): Interfaces with lenders to underwrite mortgages, checks LTV limits, and triggers automatic funding disbursements.
4. **Fractional Tokenization Engine** (`cmd/tokenization-engine`): Handles fractional pool creation, token purchases, and property digital share bounds.
5. **Immutable Audit Ledger** (`cmd/ledger-service`): Implements a SHA256 cryptographic logging ledger, guaranteeing tamperproof trails for all platform transactions.

---

## 2. Directory Layout

```
realestate-trust/
├── .github/                      # CI/CD Workflows (GitHub Actions)
│   └── workflows/
│       ├── ci.yml                # Parallel matrix test, build, scan, sign pipeline
│       └── release.yml           # Automated SemVer Release Please pipeline
├── cmd/                          # Service entrypoints (main.go)
│   ├── financing-engine/
│   ├── identity-service/
│   ├── ledger-service/
│   ├── tokenization-engine/
│   └── transaction-manager/
├── docs/                         # Architecture, research, and analysis documentation
│   └── research/                 # Observability, database HA, and network security specs
├── infra/                        # Infrastructure configurations
│   ├── gitops/                   # Declarative ArgoCD App-of-Apps manifests
│   └── helm/                     # Reusable microservice configurations chart templates
├── internal/                     # Private shared Go modules
│   ├── core/                     # Logic, validation boundaries, and state machines
│   └── db/                       # Decoupled repository queries and REST handlers
├── Dockerfile                    # Multi-stage, multi-target image builder
├── go.mod                        # Go dependency manifest (Go 1.26.5)
└── README.md
```

---

## 3. Engineering Best Practices

* **Zero-Dependency HTTP Routing**: Uses Go's native `http.ServeMux` (introduced in Go 1.22+) which handles REST wildcards (e.g. `/users/{id}`) without needing heavy external routers (like Gin or Chi).
* **Spec-First Strategy**: OpenAPI 3.0 specs inside `.loki/specs/` establish API contracts prior to writing server route implementations.
* **Underwriting Validation Bounds**: Validates mortgage criteria automatically at the domain logic level (e.g., enforcing an 80% maximum Loan-to-Value limit).
* **Decoupled Database Repositories**: Repository interfaces in `internal/db` separate SQL controllers from HTTP handlers, falling back to thread-safe memory stores for swift local testing.

---

## 4. Getting Started

### Prerequisites
* Go compiler installed (**v1.26.5** or newer)
* Docker
* Helm
* [pre-commit](https://pre-commit.com/) (run `pre-commit install && pre-commit install --hook-type commit-msg` to install hooks locally)

### Running Tests
To run the full unit and integration test suite across the workspace:
```bash
go test -v -race -cover ./...
```

### Compiling and Running Locally
To launch a microservice locally (e.g. Transaction Manager on `:8080`):
```bash
go run cmd/transaction-manager/main.go
```

### Running the Complete Stack (Docker Compose)
To spin up all five services, the PostgreSQL database, and run all SQL migrations sequentially:
```bash
docker compose up --build
```
This spins up the database, initializes the three service schemas, applies database migrations, and exposes ports `8080` (Transaction Manager), `8081` (Identity Service), `8082` (Financing Engine), `8083` (Tokenization Engine), and `8084` (Ledger Service) locally.

---

## 5. Containerization

We maintain a single, cache-efficient, multi-stage [Dockerfile](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/Dockerfile) containing separate target stages for compilation.

To build a specific microservice target (e.g., `transaction-manager`):
```bash
docker build --target transaction-manager -t realestate-trust/transaction-manager:latest .
```

---

## 6. Continuous Integration & Deployment (GitOps)

### CI Pipeline (`ci.yml`)
Runs on pushes and pull requests.
1. **Lints & Tests**: Runs format checks (`gofmt`), static analysis (`go vet`), execution checks (`go test`), and `golangci-lint`.
2. **Parallel Matrix Builds**: Compiles all 5 targets concurrently using Docker Buildx GHA layer caching.
3. **Trivy Vulnerability Scan**: Scans Docker images for High/Critical security vulnerabilities.
4. **GHCR Publishing**: Publishes secure images to the GitHub Container Registry (`ghcr.io`).
5. **Cosign Signatures**: Cryptographically signs the pushed images to prevent image-hijacking attacks.

### Release Automation (`release.yml`)
Uses **Release Please** to automate versioning based on [Conventional Commits](https://www.conventionalcommits.org/).
* Commits starting with `feat: ` trigger a minor version bump (e.g. `v1.0.0` -> `v1.1.0`).
* Commits starting with `fix: ` trigger a patch version bump (e.g. `v1.0.0` -> `v1.0.1`).
* On merges to `main`, a Release PR is generated. Merging this PR automatically tags the release and fires the CI Docker push pipeline.

### GitOps Sync (ArgoCD & Helm)
Deployments are handled automatically using a pull-based CD model:
1. Custom shared templates live in `infra/helm/charts/microservice`.
2. Declarative definitions live in [infra/gitops/service-apps.yaml](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/infra/gitops/service-apps.yaml), overwriting ports, replicas, limits, and IP whitelist ingress rules.
3. **ArgoCD** reconciles the Git status with your Kubernetes cluster, providing automated drift detection and self-healing.
