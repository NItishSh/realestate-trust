#!/usr/bin/env bash
# =============================================================================
# Kind Bootstrap Script — RealEstate Trust Platform
# =============================================================================
set -euo pipefail

CLUSTER_NAME="realestate-trust"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
KIND_CONFIG="${SCRIPT_DIR}/kind-config.yaml"
MANIFESTS_DIR="${SCRIPT_DIR}/manifests"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; exit 1; }
step() { echo -e "\n${CYAN}━━━ $1 ━━━${NC}"; }

check_dependencies() {
    step "Pre-flight checks"
    for cmd in docker kind kubectl; do
        if ! command -v "$cmd" &> /dev/null; then
            err "$cmd is not installed. Please install it first."
        fi
        log "$cmd found: $(command -v "$cmd")"
    done
}

cluster_down() {
    step "Tearing down Kind cluster: ${CLUSTER_NAME}"
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        kind delete cluster --name "${CLUSTER_NAME}"
        log "Cluster '${CLUSTER_NAME}' deleted."
    else
        warn "Cluster '${CLUSTER_NAME}' does not exist. Nothing to delete."
    fi
}

cluster_up() {
    step "Creating Kind cluster: ${CLUSTER_NAME}"
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        warn "Cluster '${CLUSTER_NAME}' already exists. Skipping creation."
    else
        kind create cluster --config "${KIND_CONFIG}" --name "${CLUSTER_NAME}"
        log "Cluster '${CLUSTER_NAME}' created successfully."
    fi
    kubectl cluster-info --context "kind-${CLUSTER_NAME}"
}

build_and_load_images() {
    step "Building Docker images"
    local targets=("transaction-manager" "identity-service" "financing-engine" "tokenization-engine" "ledger-service" "property-registry-service")
    for target in "${targets[@]}"; do
        local image_name="ghcr.io/realestate-trust/monorepo/${target}:latest"
        log "Building ${image_name}..."
        docker build -t "${image_name}" --target "${target}" "${PROJECT_ROOT}"
    done
    docker build -t "ghcr.io/realestate-trust/monorepo/frontend:latest" "${PROJECT_ROOT}/frontend"

    step "Loading images into Kind cluster"
    local images=(
        "ghcr.io/realestate-trust/monorepo/transaction-manager:latest"
        "ghcr.io/realestate-trust/monorepo/identity-service:latest"
        "ghcr.io/realestate-trust/monorepo/financing-engine:latest"
        "ghcr.io/realestate-trust/monorepo/tokenization-engine:latest"
        "ghcr.io/realestate-trust/monorepo/ledger-service:latest"
        "ghcr.io/realestate-trust/monorepo/property-registry-service:latest"
        "ghcr.io/realestate-trust/monorepo/frontend:latest"
    )
    for img in "${images[@]}"; do
        log "Loading ${img}..."
        kind load docker-image "${img}" --name "${CLUSTER_NAME}"
    done
    log "All images loaded into Kind."
}

deploy_argocd_and_postgres() {
    step "Deploying ArgoCD and PostgreSQL"

    # Create target namespace
    kubectl apply -f "${MANIFESTS_DIR}/namespace.yaml"

    # Create Secrets
    "${SCRIPT_DIR}/create-secrets.sh"

    # Deploy PostgreSQL (Dependency for ArgoCD apps)
    kubectl apply -f "${MANIFESTS_DIR}/postgres.yaml"

    # Install ArgoCD (using server-side apply to avoid CRD size limits)
    kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml --server-side --force-conflicts

    log "Waiting for ArgoCD Server to be ready..."
    kubectl wait --for=condition=Available deployment/argocd-server -n argocd --timeout=300s

    step "Applying ArgoCD GitOps Root Application"
    kubectl apply -f "${PROJECT_ROOT}/infra/gitops/root-application.yaml"

    log "ArgoCD will now sync the applications from Git."
}

deploy_helm_local_and_postgres() {
	step "Deploying Local Helm Charts, PostgreSQL, and Istio Service Mesh"

	# Create target namespace
	kubectl apply -f "${MANIFESTS_DIR}/namespace.yaml"

	# Deploy Istio Service Mesh
	"${SCRIPT_DIR}/deploy-istio.sh"

	# Apply Istio Gateway, PeerAuthentication, and DestinationRules
	kubectl apply -f "${MANIFESTS_DIR}/istio-gateway.yaml"

	# Create Secrets
	"${SCRIPT_DIR}/create-secrets.sh"

	# Deploy PostgreSQL & RabbitMQ
	kubectl apply -f "${MANIFESTS_DIR}/postgres.yaml"
	kubectl apply -f "${MANIFESTS_DIR}/rabbitmq.yaml"

	log "Waiting for PostgreSQL to be ready..."
	sleep 5
	kubectl wait --for=condition=Ready pods -l app=postgres -n realestate-trust --timeout=300s

	log "Waiting for RabbitMQ to be ready..."
	kubectl wait --for=condition=Ready pods -l app=rabbitmq -n realestate-trust --timeout=300s

	step "Installing Microservices via Helm"
	local services=("identity-service" "transaction-manager" "financing-engine" "tokenization-engine" "ledger-service" "property-registry-service" "frontend")
	for svc in "${services[@]}"; do
		log "Installing ${svc}..."
		helm upgrade --install "${svc}" "${PROJECT_ROOT}/infra/helm/charts/microservice" \
			--namespace realestate-trust \
			--values "${SCRIPT_DIR}/values/${svc}.yaml" \
			--set image.pullPolicy=Never
	done
	log "All Helm charts deployed locally with automatic Istio sidecar injection."
}

print_status() {
	step "Cluster Status"
	echo ""
	kubectl get pods -n realestate-trust -o wide
	echo ""
	log "Access your application via the Istio Ingress Gateway:"
	echo "  Unified Portal (Frontend):  http://localhost:8080"
	echo "  Identity Service:           http://localhost:8080/api/v1/users"
	echo "  Transaction Manager:        http://localhost:8080/api/v1/transactions"
	echo "  Financing Engine:           http://localhost:8080/api/v1/loans"
	echo "  Tokenization Engine:        http://localhost:8080/api/v1/pools"
	echo "  Ledger Service:             http://localhost:8080/api/v1/logs"
	echo "  Property Registry:          http://localhost:8080/api/v1/properties"
	echo ""
	log "Useful commands:"
	echo "  kubectl get pods -n realestate-trust         # List pods"
	echo "  ./infra/kind/kind-up.sh --down               # Tear down"
}

main() {
    local mode="helm"

    for arg in "$@"; do
        case $arg in
            --down)
                check_dependencies
                cluster_down
                exit 0
                ;;
            --reset)
                check_dependencies
                cluster_down
                cluster_up
                build_and_load_images
                ;;
            --argocd)
                mode="argocd"
                ;;
        esac
    done

    # If not resetting but cluster doesn't exist, create it
    if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        check_dependencies
        cluster_up
        build_and_load_images
    fi

    if [ "$mode" = "argocd" ]; then
        deploy_argocd_and_postgres
    else
        deploy_helm_local_and_postgres
    fi
    print_status
}

main "$@"
