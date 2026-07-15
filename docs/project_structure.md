# Monorepo, Infrastructure & UI Architecture Spec

To keep development agile, type-safe, and version-controlled, we recommend a **Monorepo** structure. This layout houses the Go microservices, next-generation UI code, and Infrastructure as Code (IaC) files under a single repository, making it simple to share API contracts and deploy changes synchronously.

---

## 1. Directory Structure Blueprint

Here is the proposed repository layout for `realestate-trust`:

```
realestate-trust/
├── .github/                      # CI/CD Workflows (GitHub Actions)
├── cmd/                          # Entry points for Go microservices
│   ├── transaction-manager/      # Core Escrow Manager main.go
│   ├── financing-engine/        # Financing Engine main.go
│   ├── tokenization-engine/      # Fractional Tokenization main.go
│   ├── identity-service/         # User KYC & Consent main.go
│   └── ledger-service/           # Immutable Audit Ledger main.go
├── docs/                         # System architecture, research, and walkthroughs
├── frontend/                     # Next.js (React) Dashboard Portal
│   ├── src/
│   │   ├── app/                  # Next.js app router paths
│   │   ├── components/           # Reusable UI widgets (cards, forms, loaders)
│   │   ├── hooks/                # Custom React Hooks
│   │   └── lib/                  # API clients and state utilities
│   ├── package.json
│   └── tailwind.config.js
├── infra/                        # Infrastructure as Code (IaC)
│   ├── terraform/                # Cloud resource provisioning
│   │   ├── environments/         # dev, staging, prod configs
│   │   └── modules/              # Reusable modules (VPC, EKS, RDS, MQ)
│   ├── helm/                     # Kubernetes app configuration packaging
│   │   └── charts/               # Custom charts for Go services
│   └── gitops/                   # ArgoCD manifest definitions
├── internal/                     # Private shared Go modules
│   ├── config/                   # Configuration parsing
│   ├── db/                       # PostgreSQL query wrappers (sqlc generated)
│   ├── model/                    # Go data entities
│   └── pb/                       # Generated gRPC Protobuf codes
├── proto/                        # Protocol Buffer definitions (.proto files)
├── go.mod                        # Go dependency manifest
└── Makefile                      # Build, test, and code-generation tasks
```

---

## 2. Containerization Strategy: Separate Containers

Yes! **Each microservice must compile into a distinct Docker container image and run as a separate Deployment entity inside the Kubernetes cluster.**

We do not bundle multiple services inside a single container. Here is why this design is critical:

### 2.1 Benefits of Service Isolation
* **Independent Scalability**: `transaction-manager` and `tokenization-engine` will scale up and down based on API request spikes (e.g. during peak investment hours). `ledger-service` runs as a background message queue subscriber and scales based on queue depth (RabbitMQ/Kafka message lag).
* **Least-Privilege Security (IAM Isolation)**: Kubernetes assigns IAM Roles directly to Pod **ServiceAccounts** (via IAM Roles for Service Accounts - IRSA).
  * `identity-service` needs access to KYC secrets and database vaults.
  * `transaction-manager` needs access to banking API secrets.
  * By splitting them into separate deployments, they only get the specific permissions they need, reducing the blast radius of a credential leak.
* **Zero-Downtime Rolling Deployments**: If you fix a bug in the `financing-engine` (e.g., modifying an NBFC webhook parsing rule), you can deploy a new image tag to the `financing-engine` deployment. The other four services remain completely untouched and online, guaranteeing zero platform downtime.
* **Granular Resource Allocation**: Allocate different CPU and Memory limits per container (`limits` and `requests`), ensuring heavy ledger writes or token calculations do not choke core transactional routes.

### 2.2 Multi-Target Dockerfile
Using Go in a monorepo allows us to maintain a single, highly cache-efficient `Dockerfile` that uses **multi-target builds** to compile and pack individual microservice images:

```dockerfile
# --- BUILD STAGE ---
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Compile all microservice binaries statically
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/transaction-manager cmd/transaction-manager/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/financing-engine cmd/financing-engine/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/tokenization-engine cmd/tokenization-engine/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/identity-service cmd/identity-service/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/ledger-service cmd/ledger-service/main.go

# --- TARGET: Transaction Manager ---
FROM alpine:3.18 AS transaction-manager
WORKDIR /root/
COPY --from=builder /bin/transaction-manager .
EXPOSE 8080
CMD ["./transaction-manager"]

# --- TARGET: Financing Engine ---
FROM alpine:3.18 AS financing-engine
WORKDIR /root/
COPY --from=builder /bin/financing-engine .
EXPOSE 8080
CMD ["./financing-engine"]

# --- TARGET: Tokenization Engine ---
FROM alpine:3.18 AS tokenization-engine
WORKDIR /root/
COPY --from=builder /bin/tokenization-engine .
EXPOSE 8080
CMD ["./tokenization-engine"]

# --- TARGET: Identity Service ---
FROM alpine:3.18 AS identity-service
WORKDIR /root/
COPY --from=builder /bin/identity-service .
EXPOSE 8080
CMD ["./identity-service"]

# --- TARGET: Ledger Service ---
FROM alpine:3.18 AS ledger-service
WORKDIR /root/
COPY --from=builder /bin/ledger-service .
CMD ["./ledger-service"]
```

When building, trigger the specific target:
```bash
docker build --target transaction-manager -t realestate-trust/transaction-manager:v1.0.0 .
docker build --target ledger-service -t realestate-trust/ledger-service:v1.0.0 .
```

---

## 3. Kubernetes Deployment & Service Mapping

Each microservice maps to different Kubernetes objects depending on its network exposure:

| Microservice | Kubernetes Deployment | Kubernetes Service | Ingress Routes | Scaler (HPA) Metrics |
| :--- | :--- | :--- | :--- | :--- |
| `transaction-manager` | Yes | Yes (ClusterIP, Port 80) | `/api/v1/escrow/*` | CPU (75%) + HTTP Latency |
| `financing-engine` | Yes | Yes (ClusterIP, Port 80) | `/api/v1/financing/*` | CPU (75%) |
| `tokenization-engine`| Yes | Yes (ClusterIP, Port 80) | `/api/v1/tokenization/*` | CPU (75%) |
| `identity-service` | Yes | Yes (ClusterIP, Port 80) | None (Internal gRPC only) | CPU (70%) |
| `ledger-service` | Yes | **No** (Direct queue consumer) | None | Message queue backlog lag |

---

## 4. Infrastructure as Code (IaC) Strategy

To deploy and maintain the containerized microservices in a secure, repeatable, and audit-compliant way, we use a tiered GitOps strategy:

### 4.1 Terraform (Cloud Provisioning)
* **Scope**: Automates creation of foundational resources.
* **Resources Managed**:
  * Private VPCs with strict public/private subnet routing.
  * Managed EKS/GKE Kubernetes cluster control planes.
  * Multi-AZ PostgreSQL instances (RDS or Cloud SQL) with encryptions enabled.
  * AWS IAM Roles for Service Accounts (IRSA) to enable secure K8s-to-Vault communication.
  * Event Broker instances (Managed RabbitMQ or Amazon MQ).

### 4.2 Helm (Configuration Templating)
* **Scope**: Defines how Go microservices are configured and run in Kubernetes.
* **Charts Created**: A unified internal Helm chart parameterized to toggle replicas, environment variables, Secrets manager linkages, ingress routes, and resource limits per environment.

### 4.3 GitOps (ArgoCD Deployment)
* **Scope**: Continuous Delivery engine.
* **Process**:
  1. Any merge to the `main` branch updates the service image tags in the `infra/gitops/` manifests.
  2. ArgoCD detects the change and automatically reconciles the Kubernetes cluster resources to pull the new container and apply configurations without manual CLI intervention.

---

## 5. UI Frontend Architecture Spec

The user interface must support high-capital workflows, investor dashboards, and legal verifier checks.

### 5.1 Tech Stack
* **Framework**: **Next.js 14+ (React)** with TypeScript. Next.js App Router provides Server-Side Rendering (SSR) for public property portals (SEO optimization) and fast Client-Side Hydration for dynamic user dashboards.
* **State Management**: **Zustand** or **Redux Toolkit** for lightweight, decoupled dashboard states (e.g., tracking current active transaction stages, uploaded compliance PDFs, and investor balance indicators).
* **Styling**: TailwindCSS implementing a custom visual design system (sleek dark mode, glassmorphism card metrics, micro-animations for transaction state progression).
* **Communication**: **gRPC-Web** or typed REST clients. Next.js APIs act as a Backend-for-Frontend (BFF) proxy to authenticate sessions and route requests downstream to the Go microservices.

### 5.2 Key User Journeys Designed in the UI
* **Escrow Tracker Widget**: A visual timeline showing the transaction progress through DRAFT, FUNDED, DUE_DILIGENCE, SRO_REGISTRATION, and RELEASED (using real-time WebSocket state subscriptions).
* **KYC & Onboarding Form**: Multi-step verification form integrating camera capture (for Video KYC face-matching) and PAN details verification fields.
* **Fractional Portfolio Board**: Interactive charts showcasing property share distributions, cap tables, estimated yield payouts, and active investment plots.
