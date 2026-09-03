SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

# Colors for terminal output
COLOR_RESET  := \033[0m
COLOR_INFO   := \033[36m
COLOR_TITLE  := \033[1;33m
COLOR_SUCCESS:= \033[32m
COLOR_WARN   := \033[31m

SERVICES := transaction-manager identity-service financing-engine tokenization-engine ledger-service property-registry-service feedback-service re-cli

.PHONY: help
help: ## Show this help message
	@echo -e "$(COLOR_TITLE)=======================================================$(COLOR_RESET)"
	@echo -e "$(COLOR_TITLE) RealEstate Trust - Development & Automation Makefile   $(COLOR_RESET)"
	@echo -e "$(COLOR_TITLE)=======================================================$(COLOR_RESET)"
	@echo -e "Usage: $(COLOR_INFO)make <target>$(COLOR_RESET)\n"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  $(COLOR_INFO)%-26s$(COLOR_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ==============================================================================
# Dependency Management & Setup
# ==============================================================================

.PHONY: deps
deps: deps-go deps-frontend ## Download Go and frontend dependencies
	@echo -e "$(COLOR_SUCCESS)✓ All project dependencies installed.$(COLOR_RESET)"

.PHONY: deps-go
deps-go: ## Download Go module dependencies and verify
	@echo -e "$(COLOR_INFO)Downloading Go dependencies...$(COLOR_RESET)"
	go mod download
	go mod verify

.PHONY: deps-frontend
deps-frontend: ## Install frontend npm dependencies
	@if [ -d "frontend" ]; then \
		echo -e "$(COLOR_INFO)Installing frontend dependencies...$(COLOR_RESET)"; \
		npm --prefix frontend install; \
	fi

.PHONY: deps-tools
deps-tools: ## Check if required development tools are installed on host
	@echo -e "$(COLOR_INFO)Checking local toolchain...$(COLOR_RESET)"
	@tools="go docker terraform helm kind golangci-lint hadolint trivy checkov tflint pre-commit kubectl"; \
	for tool in $$tools; do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo -e "  [$(COLOR_SUCCESS)OK$(COLOR_RESET)] $$tool ($$(command -v $$tool))"; \
		else \
			echo -e "  [$(COLOR_WARN)MISSING$(COLOR_RESET)] $$tool"; \
		fi; \
	done

.PHONY: precommit-init
precommit-init: ## Install git pre-commit hooks into .git/hooks
	@echo -e "$(COLOR_INFO)Initializing pre-commit hooks...$(COLOR_RESET)"
	pre-commit install --install-hooks
	pre-commit install --hook-type commit-msg
	@echo -e "$(COLOR_SUCCESS)✓ Pre-commit hooks installed.$(COLOR_RESET)"

# ==============================================================================
# Linting, Formatting & Pre-Commit Checks
# ==============================================================================

.PHONY: precommit
precommit: ## Run all pre-commit hooks on all files
	@echo -e "$(COLOR_INFO)Running all pre-commit hooks...$(COLOR_RESET)"
	pre-commit run --all-files

.PHONY: lint
lint: lint-go lint-frontend typecheck-frontend lint-helm lint-docker lint-tf lint-compose ## Run all static linters across the repo
	@echo -e "$(COLOR_SUCCESS)✓ All linters passed successfully.$(COLOR_RESET)"

.PHONY: lint-go
lint-go: ## Run golangci-lint on Go packages
	@echo -e "$(COLOR_INFO)Running golangci-lint...$(COLOR_RESET)"
	golangci-lint run ./cmd/... ./internal/... ./test/...

.PHONY: lint-frontend
lint-frontend: ## Run ESLint on Next.js frontend
	@echo -e "$(COLOR_INFO)Running ESLint on frontend...$(COLOR_RESET)"
	npm --prefix frontend run lint

.PHONY: typecheck-frontend
typecheck-frontend: ## Run TypeScript type-checking on frontend
	@echo -e "$(COLOR_INFO)Running TypeScript type-check on frontend...$(COLOR_RESET)"
	npm --prefix frontend run type-check

.PHONY: test-frontend
test-frontend: ## Run Vitest unit tests on frontend
	@echo -e "$(COLOR_INFO)Running Vitest frontend unit tests...$(COLOR_RESET)"
	npm --prefix frontend run test

.PHONY: lint-helm
lint-helm: ## Run helm lint on charts
	@echo -e "$(COLOR_INFO)Linting Helm charts...$(COLOR_RESET)"
	helm lint infra/helm/charts/microservice

.PHONY: lint-docker
lint-docker: ## Run hadolint on Dockerfiles
	@echo -e "$(COLOR_INFO)Linting Dockerfiles with hadolint...$(COLOR_RESET)"
	hadolint Dockerfile frontend/Dockerfile

.PHONY: lint-tf
lint-tf: ## Run tflint and terraform validate
	@echo -e "$(COLOR_INFO)Validating Terraform configurations...$(COLOR_RESET)"
	terraform -chdir=infra/terraform validate
	tflint --chdir=infra/terraform --config=$(CURDIR)/.tflint.hcl

.PHONY: lint-compose
lint-compose: ## Validate docker-compose file configuration
	@echo -e "$(COLOR_INFO)Validating docker-compose configuration...$(COLOR_RESET)"
	docker compose config -q

.PHONY: fmt
fmt: ## Format Go code and Terraform files
	@echo -e "$(COLOR_INFO)Formatting Go code...$(COLOR_RESET)"
	gofmt -s -w .
	go mod tidy
	@echo -e "$(COLOR_INFO)Formatting Terraform files...$(COLOR_RESET)"
	terraform -chdir=infra/terraform fmt

# ==============================================================================
# Security & Vulnerability Scanning
# ==============================================================================

.PHONY: scan-vuln
scan-vuln: ## Scan filesystem for package vulnerabilities using Trivy
	@echo -e "$(COLOR_INFO)Scanning repo for high/critical vulnerabilities...$(COLOR_RESET)"
	trivy fs --severity HIGH,CRITICAL .

.PHONY: scan-secrets
scan-secrets: ## Scan filesystem for unencrypted secrets using Trivy
	@echo -e "$(COLOR_INFO)Scanning repo for exposed secrets...$(COLOR_RESET)"
	trivy fs --security-checks secret .

.PHONY: scan-docker
scan-docker: ## Scan built microservice container images for vulnerabilities
	@echo -e "$(COLOR_INFO)Scanning container images with Trivy...$(COLOR_RESET)"
	@for svc in $(SERVICES); do \
		if [ "$$svc" != "re-cli" ]; then \
			echo -e "Scanning $$svc:latest..."; \
			trivy image --severity HIGH,CRITICAL $$svc:latest || true; \
		fi; \
	done

# ==============================================================================
# Build Binaries & Images
# ==============================================================================

.PHONY: build
build: $(SERVICES) ## Build all Go microservices and CLI into bin/
	@echo -e "$(COLOR_SUCCESS)✓ All Go binaries built in ./bin/$(COLOR_RESET)"

$(SERVICES):
	@echo -e "$(COLOR_INFO)Building $@...$(COLOR_RESET)"
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/$@ ./cmd/$@/main.go

.PHONY: build-frontend
build-frontend: ## Build Next.js frontend production bundle
	@echo -e "$(COLOR_INFO)Building frontend bundle...$(COLOR_RESET)"
	npm --prefix frontend run build

.PHONY: docker-build
docker-build: ## Build all container images using local build script
	@echo -e "$(COLOR_INFO)Building microservice Docker images...$(COLOR_RESET)"
	./infra/terraform/scripts/build_images.sh

# ==============================================================================
# Testing & Code Coverage
# ==============================================================================

.PHONY: test
test: ## Run unit and integration tests with race detector
	@echo -e "$(COLOR_INFO)Running test suite...$(COLOR_RESET)"
	go test -v -race -coverprofile=coverage.out ./cmd/... ./internal/...

.PHONY: test-cover
test-cover: test ## Run tests and generate HTML code coverage report
	@echo -e "$(COLOR_INFO)Generating coverage report...$(COLOR_RESET)"
	go tool cover -html=coverage.out -o coverage.html
	@echo -e "$(COLOR_SUCCESS)✓ Coverage report generated at coverage.html$(COLOR_RESET)"

.PHONY: test-e2e
test-e2e: ## Run E2E test suite (requires running services or compose)
	@echo -e "$(COLOR_INFO)Running E2E tests...$(COLOR_RESET)"
	go test -v -tags=e2e ./test/e2e/...

.PHONY: compose-test-e2e
compose-test-e2e: ## Build, start Docker Compose stack, run E2E tests, and tear down cleanly
	@echo -e "$(COLOR_INFO)Building and starting Docker Compose infrastructure...$(COLOR_RESET)"
	@docker compose down -v >/dev/null 2>&1 || true
	docker compose up -d --build postgres rabbitmq
	@echo -e "$(COLOR_INFO)Waiting for database and message broker to be healthy...$(COLOR_RESET)"
	@sleep 5
	@echo -e "$(COLOR_INFO)Running database schema migrations...$(COLOR_RESET)"
	docker compose up identity-migrate transactions-migrate financing-migrate feedback-migrate tokenization-migrate ledger-migrate properties-migrate
	@echo -e "$(COLOR_INFO)Starting microservices and frontend...$(COLOR_RESET)"
	docker compose up -d
	@sleep 5
	@echo -e "$(COLOR_INFO)Running E2E tests...$(COLOR_RESET)"
	@go test -v -tags=e2e ./test/e2e/... || (docker compose down -v && exit 1)
	@echo -e "$(COLOR_SUCCESS)✓ E2E tests passed successfully.$(COLOR_RESET)"
	@echo -e "$(COLOR_INFO)Stopping Docker Compose stack...$(COLOR_RESET)"
	docker compose down -v

.PHONY: test-contract
test-contract: ## Run contract tests
	@echo -e "$(COLOR_INFO)Running contract tests...$(COLOR_RESET)"
	go test -v ./test/contract/...

# ==============================================================================
# Docker Compose & Database Helpers
# ==============================================================================

.PHONY: compose-up
compose-up: ## Start local Docker Compose stack in background
	@echo -e "$(COLOR_INFO)Starting Docker Compose stack...$(COLOR_RESET)"
	docker compose up -d

.PHONY: compose-down
compose-down: ## Stop and remove local Docker Compose stack and volumes
	@echo -e "$(COLOR_INFO)Stopping Docker Compose stack...$(COLOR_RESET)"
	docker compose down -v

.PHONY: compose-logs
compose-logs: ## Follow Docker Compose container logs
	docker compose logs -f

.PHONY: compose-ps
compose-ps: ## List running Docker Compose services
	docker compose ps

.PHONY: db-psql
db-psql: ## Open interactive psql shell into Docker Compose postgres container
	@echo -e "$(COLOR_INFO)Connecting to PostgreSQL...$(COLOR_RESET)"
	docker compose exec postgres psql -U postgres -d postgres

# ==============================================================================
# Kubernetes / Kind & GitOps Cluster Management
# ==============================================================================

.PHONY: cluster-up
cluster-up: ## Provision Kind cluster, build images, and bootstrap ArgoCD GitOps
	@echo -e "$(COLOR_INFO)Bootstrapping Kind cluster and GitOps stack...$(COLOR_RESET)"
	./reset.sh

.PHONY: cluster-down
cluster-down: ## Destroy Kind cluster and clean up Terraform state
	@echo -e "$(COLOR_INFO)Destroying Kind cluster...$(COLOR_RESET)"
	cd infra/terraform && terraform destroy -auto-approve

.PHONY: cluster-status
cluster-status: ## Show status of all pods across all namespaces
	@echo -e "$(COLOR_INFO)Cluster Pod Status:$(COLOR_RESET)"
	kubectl get pods -A

.PHONY: cluster-apps
cluster-apps: ## List all ArgoCD application statuses
	@echo -e "$(COLOR_INFO)ArgoCD Applications:$(COLOR_RESET)"
	kubectl get applications.argoproj.io -n argocd

.PHONY: cluster-endpoints endpoints
endpoints: cluster-endpoints
cluster-endpoints: ## Display all application, API health, IAM, and observability endpoints
	@echo -e "\n$(COLOR_TITLE)=========================================================================$(COLOR_RESET)"
	@echo -e "$(COLOR_TITLE) 🏢 RealEstate-Trust — Cluster Endpoints & Monitoring Dashboard           $(COLOR_RESET)"
	@echo -e "$(COLOR_TITLE)=========================================================================$(COLOR_RESET)"
	@echo -e "\n$(COLOR_INFO)🚪 Unified Istio Ingress Gateway (http://localhost:8080):$(COLOR_RESET)"
	@echo -e "  • Frontend Web UI       : $(COLOR_SUCCESS)http://localhost:8080/$(COLOR_RESET) (Marketplace & Portals)"
	@echo -e "  • Grafana Dashboard     : $(COLOR_SUCCESS)http://localhost:8080/grafana/$(COLOR_RESET) (admin / make grafana-password)"
	@echo -e "  • Kiali Service Mesh    : $(COLOR_SUCCESS)http://localhost:8080/kiali/$(COLOR_RESET)"
	@echo -e "  • Keycloak OIDC Realm   : $(COLOR_SUCCESS)http://localhost:8080/realms/realestate-trust$(COLOR_RESET)"
	@echo -e "  • Identity Service API  : $(COLOR_SUCCESS)http://localhost:8080/api/v1/users$(COLOR_RESET)"
	@echo -e "  • Transaction API       : $(COLOR_SUCCESS)http://localhost:8080/api/v1/transactions$(COLOR_RESET)"
	@echo -e "  • Financing API         : $(COLOR_SUCCESS)http://localhost:8080/api/v1/loans$(COLOR_RESET)"
	@echo -e "  • Tokenization API      : $(COLOR_SUCCESS)http://localhost:8080/api/v1/pools$(COLOR_RESET)"
	@echo -e "  • Audit Ledger API      : $(COLOR_SUCCESS)http://localhost:8080/api/v1/logs$(COLOR_RESET)"
	@echo -e "  • Property Registry API : $(COLOR_SUCCESS)http://localhost:8080/api/v1/properties$(COLOR_RESET)"
	@echo -e "  • Feedback Reviews API  : $(COLOR_SUCCESS)http://localhost:8080/api/v1/feedback$(COLOR_RESET)"
	@echo -e "\n$(COLOR_INFO)🔌 Dedicated Port-Forward Helpers (Run in terminal to open browser):$(COLOR_RESET)"
	@echo -e "  • make port-forward-frontend  -> $(COLOR_SUCCESS)http://localhost:3000$(COLOR_RESET) (Next.js Web UI)"
	@echo -e "  • make port-forward-grafana   -> $(COLOR_SUCCESS)http://localhost:3001$(COLOR_RESET) (admin / make grafana-password)"
	@echo -e "  • make port-forward-opencost  -> $(COLOR_SUCCESS)http://localhost:9003$(COLOR_RESET) (FinOps UI)"
	@echo -e "  • make port-forward-keycloak  -> $(COLOR_SUCCESS)http://localhost:8088$(COLOR_RESET) (admin / admin)"
	@echo -e "  • make port-forward-kiali     -> $(COLOR_SUCCESS)http://localhost:20001$(COLOR_RESET) (Mesh topology)"
	@echo -e "  • make port-forward-vault     -> $(COLOR_SUCCESS)http://localhost:8200$(COLOR_RESET) (token: root)"
	@echo -e "  • make port-forward-argocd    -> $(COLOR_SUCCESS)http://localhost:8087$(COLOR_RESET) (admin / make argocd-password)"
	@echo -e "  • make port-forward-prometheus-> $(COLOR_SUCCESS)http://localhost:9090$(COLOR_RESET) (Prometheus Graph)"
	@echo -e "  • make port-forward-rabbitmq  -> $(COLOR_SUCCESS)http://localhost:15672$(COLOR_RESET) (guest / guest)"
	@echo -e "\n$(COLOR_INFO)🔍 Verification & Diagnostic Commands:$(COLOR_RESET)"
	@echo -e "  • make cluster-status         : View pod status across all namespaces"
	@echo -e "  • make cluster-apps           : View ArgoCD GitOps application sync state"
	@echo -e "  • make finops-rightsize-dryrun : Preview FinOps VPA resource allocations"
	@echo -e "$(COLOR_TITLE)=========================================================================$(COLOR_RESET)\n"

# ==============================================================================
# Cluster Access, Credentials & Port-Forwarding
# ==============================================================================

.PHONY: argocd-password
argocd-password: ## Fetch ArgoCD admin password from Kubernetes secret
	@echo -e "$(COLOR_INFO)ArgoCD Admin Password:$(COLOR_RESET)"
	@kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo ""

.PHONY: grafana-password
grafana-password: ## Fetch Grafana admin password from Kubernetes secret
	@echo -e "$(COLOR_INFO)Grafana Admin Password:$(COLOR_RESET)"
	@kubectl -n observability get secret grafana-admin-secret -o jsonpath="{.data.admin-password}" | base64 -d 2>/dev/null || echo "admin"

.PHONY: port-forward-argocd
port-forward-argocd: ## Port-forward ArgoCD server to localhost:8087
	@echo -e "$(COLOR_INFO)Port-forwarding ArgoCD UI on http://localhost:8087 (admin / run 'make argocd-password')...$(COLOR_RESET)"
	kubectl port-forward svc/argocd-server -n argocd 8087:80

.PHONY: port-forward-grafana
port-forward-grafana: ## Port-forward Grafana dashboard to localhost:3001
	@echo -e "$(COLOR_INFO)Port-forwarding Grafana UI on http://localhost:3001...$(COLOR_RESET)"
	kubectl port-forward svc/grafana -n observability 3001:80

.PHONY: port-forward-keycloak
port-forward-keycloak: ## Port-forward Keycloak UI to localhost:8088
	@echo -e "$(COLOR_INFO)Port-forwarding Keycloak UI on http://localhost:8088 (admin / admin)...$(COLOR_RESET)"
	kubectl port-forward svc/keycloak -n realestate-trust 8088:8080

.PHONY: port-forward-kiali
port-forward-kiali: ## Port-forward Kiali Istio console to localhost:20001
	@echo -e "$(COLOR_INFO)Port-forwarding Kiali UI on http://localhost:20001...$(COLOR_RESET)"
	kubectl port-forward svc/kiali -n istio-system 20001:20001

.PHONY: port-forward-vault
port-forward-vault: ## Port-forward HashiCorp Vault to localhost:8200
	@echo -e "$(COLOR_INFO)Port-forwarding Vault on http://localhost:8200...$(COLOR_RESET)"
	kubectl port-forward svc/vault -n vault 8200:8200

.PHONY: port-forward-frontend
port-forward-frontend: ## Port-forward Next.js Frontend UI to localhost:3000
	@echo -e "$(COLOR_INFO)Port-forwarding Frontend UI on http://localhost:3000...$(COLOR_RESET)"
	kubectl port-forward svc/frontend -n realestate-trust 3000:3000

.PHONY: port-forward-opencost
port-forward-opencost: ## Port-forward OpenCost FinOps UI to localhost:9003
	@echo -e "$(COLOR_INFO)Port-forwarding OpenCost UI on http://localhost:9003...$(COLOR_RESET)"
	kubectl port-forward svc/opencost -n observability 9003:9090

.PHONY: port-forward-prometheus
port-forward-prometheus: ## Port-forward Prometheus server to localhost:9090
	@echo -e "$(COLOR_INFO)Port-forwarding Prometheus UI on http://localhost:9090...$(COLOR_RESET)"
	kubectl port-forward svc/prometheus-server -n istio-system 9090:80

.PHONY: port-forward-rabbitmq
port-forward-rabbitmq: ## Port-forward RabbitMQ Management Console to localhost:15672
	@echo -e "$(COLOR_INFO)Port-forwarding RabbitMQ UI on http://localhost:15672 (guest / guest)...$(COLOR_RESET)"
	kubectl port-forward svc/rabbitmq -n realestate-trust 15672:15672

.PHONY: finops-rightsize
finops-rightsize: ## Execute automated FinOps right-sizing policy engine against cluster
	@echo -e "$(COLOR_INFO)Running FinOps Workload Right-Sizing Engine...$(COLOR_RESET)"
	go run ./cmd/re-cli finops rightsize

.PHONY: finops-rightsize-dryrun
finops-rightsize-dryrun: ## Preview FinOps right-sizing recommendations without modifying files
	@echo -e "$(COLOR_INFO)Previewing FinOps Workload Right-Sizing Recommendations...$(COLOR_RESET)"
	go run ./cmd/re-cli finops rightsize --dry-run

# ==============================================================================
# Performance, Load & Soak Testing Suite
# ==============================================================================

.PHONY: perf-smoke
perf-smoke: ## Run quick 1-minute k6 smoke test to verify endpoints and SLAs
	@echo -e "$(COLOR_INFO)Executing k6 Smoke Test...$(COLOR_RESET)"
	bash ./test/perf/run-k6.sh ./scenarios/smoke.js

.PHONY: perf-load
perf-load: ## Run 12-minute k6 load test simulating standard production traffic
	@echo -e "$(COLOR_INFO)Executing k6 Standard Load Test...$(COLOR_RESET)"
	bash ./test/perf/run-k6.sh ./scenarios/load.js

.PHONY: perf-stress
perf-stress: ## Run 23-minute k6 stress test to discover breaking points and saturation limits
	@echo -e "$(COLOR_INFO)Executing k6 Stress & Saturation Test...$(COLOR_RESET)"
	bash ./test/perf/run-k6.sh ./scenarios/stress.js

.PHONY: perf-soak
perf-soak: ## Run 24-hour k6 soak test to detect memory leaks and connection degradation
	@echo -e "$(COLOR_INFO)Executing k6 24-Hour Sustained Soak Test...$(COLOR_RESET)"
	bash ./test/perf/run-k6.sh ./scenarios/soak_24h.js

.PHONY: perf-history
perf-history: ## Display historical performance benchmark trend comparison table
	@echo -e "$(COLOR_INFO)Performance Benchmark History (test/perf/reports/history.csv):$(COLOR_RESET)\n"
	@if [ -f test/perf/reports/history.csv ]; then \
		column -s, -t < test/perf/reports/history.csv; \
	else \
		echo -e "$(COLOR_WARN)No benchmark history found. Run a test like 'make perf-smoke' first.$(COLOR_RESET)"; \
	fi

.PHONY: perf-report
perf-report: ## Open the latest generated HTML performance report in default browser
	@if [ -f test/perf/reports/latest.html ]; then \
		echo -e "$(COLOR_INFO)Opening latest performance report...$(COLOR_RESET)"; \
		open test/perf/reports/latest.html 2>/dev/null || xdg-open test/perf/reports/latest.html 2>/dev/null || echo "Report available at: test/perf/reports/latest.html"; \
	else \
		echo -e "$(COLOR_WARN)No HTML report found. Run 'make perf-smoke' first.$(COLOR_RESET)"; \
	fi

# ==============================================================================
# Clean & Housekeeping
# ==============================================================================

.PHONY: clean
clean: ## Clean build binaries, coverage files, and temporary artifacts
	@echo -e "$(COLOR_INFO)Cleaning build artifacts...$(COLOR_RESET)"
	rm -rf bin/ coverage.out coverage.html re-cli
	@echo -e "$(COLOR_SUCCESS)✓ Workspace cleaned.$(COLOR_RESET)"
