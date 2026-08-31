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
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  $(COLOR_INFO)%-22s$(COLOR_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

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
	@tools="go docker terraform helm kind golangci-lint hadolint trivy checkov tflint pre-commit"; \
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
lint: lint-go lint-helm lint-docker lint-tf lint-compose ## Run all static linters across the repo
	@echo -e "$(COLOR_SUCCESS)✓ All linters passed successfully.$(COLOR_RESET)"

.PHONY: lint-go
lint-go: ## Run golangci-lint on Go packages
	@echo -e "$(COLOR_INFO)Running golangci-lint...$(COLOR_RESET)"
	golangci-lint run ./cmd/... ./internal/... ./test/...

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
	tflint --chdir=infra/terraform

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

.PHONY: test-contract
test-contract: ## Run contract tests
	@echo -e "$(COLOR_INFO)Running contract tests...$(COLOR_RESET)"
	go test -v ./test/contract/...

# ==============================================================================
# Docker Compose Management
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

.PHONY: argocd-apps
cluster-apps: ## List all ArgoCD application statuses
	@echo -e "$(COLOR_INFO)ArgoCD Applications:$(COLOR_RESET)"
	kubectl get applications.argoproj.io -n argocd

# ==============================================================================
# Clean & Housekeeping
# ==============================================================================

.PHONY: clean
clean: ## Clean build binaries, coverage files, and temporary artifacts
	@echo -e "$(COLOR_INFO)Cleaning build artifacts...$(COLOR_RESET)"
	rm -rf bin/ coverage.out coverage.html re-cli
	@echo -e "$(COLOR_SUCCESS)✓ Workspace cleaned.$(COLOR_RESET)"
