#!/usr/bin/env bash
# =============================================================================
# Kind Bootstrap Script — RealEstate Trust Platform
# =============================================================================
# Usage:
#   ./infra/kind/kind-up.sh          # Create cluster + deploy everything
#   ./infra/kind/kind-up.sh --down   # Tear down the cluster
#   ./infra/kind/kind-up.sh --reset  # Destroy + recreate from scratch
# =============================================================================
set -euo pipefail

CLUSTER_NAME="realestate-trust"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
KIND_CONFIG="${SCRIPT_DIR}/kind-config.yaml"
MANIFESTS_DIR="${SCRIPT_DIR}/manifests"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; exit 1; }
step() { echo -e "\n${CYAN}━━━ $1 ━━━${NC}"; }

# ---------------------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------------------
check_dependencies() {
    step "Pre-flight checks"
    for cmd in docker kind kubectl; do
        if ! command -v "$cmd" &> /dev/null; then
            err "$cmd is not installed. Please install it first."
        fi
        log "$cmd found: $(command -v "$cmd")"
    done
}

# ---------------------------------------------------------------------------
# Tear down
# ---------------------------------------------------------------------------
cluster_down() {
    step "Tearing down Kind cluster: ${CLUSTER_NAME}"
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        kind delete cluster --name "${CLUSTER_NAME}"
        log "Cluster '${CLUSTER_NAME}' deleted."
    else
        warn "Cluster '${CLUSTER_NAME}' does not exist. Nothing to delete."
    fi
}

# ---------------------------------------------------------------------------
# Create cluster
# ---------------------------------------------------------------------------
cluster_up() {
    step "Creating Kind cluster: ${CLUSTER_NAME}"
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        warn "Cluster '${CLUSTER_NAME}' already exists. Skipping creation."
        warn "Use --reset to destroy and recreate."
    else
        kind create cluster --config "${KIND_CONFIG}" --name "${CLUSTER_NAME}"
        log "Cluster '${CLUSTER_NAME}' created successfully."
    fi

    # Ensure kubectl context points to our cluster
    kubectl cluster-info --context "kind-${CLUSTER_NAME}"
}

# ---------------------------------------------------------------------------
# Build and load images into Kind
# ---------------------------------------------------------------------------
build_and_load_images() {
    step "Building Docker images"

    # Build backend services (multi-target Dockerfile)
    local targets=("transaction-manager" "identity-service" "financing-engine" "tokenization-engine" "ledger-service")
    for target in "${targets[@]}"; do
        local image_name="realestate-trust-${target}:latest"
        log "Building ${image_name}..."
        docker build -t "${image_name}" --target "${target}" "${PROJECT_ROOT}"
    done

    # Build frontend
    log "Building realestate-trust-frontend:latest..."
    docker build -t "realestate-trust-frontend:latest" "${PROJECT_ROOT}/frontend"

    step "Loading images into Kind cluster"
    local images=(
        "realestate-trust-transaction-manager:latest"
        "realestate-trust-identity-service:latest"
        "realestate-trust-financing-engine:latest"
        "realestate-trust-tokenization-engine:latest"
        "realestate-trust-ledger-service:latest"
        "realestate-trust-frontend:latest"
    )
    for img in "${images[@]}"; do
        log "Loading ${img}..."
        kind load docker-image "${img}" --name "${CLUSTER_NAME}"
    done
    log "All images loaded into Kind."
}

# ---------------------------------------------------------------------------
# Deploy manifests
# ---------------------------------------------------------------------------
deploy_manifests() {
    step "Deploying Kubernetes manifests via Kustomize"
    kubectl apply -k "${MANIFESTS_DIR}"
    log "All manifests applied."

    step "Waiting for pods to be ready"
    kubectl wait --for=condition=Ready pods --all \
        -n realestate-trust \
        --timeout=120s \
        2>/dev/null || warn "Some pods may still be starting up."

    log "Deployment complete!"
}

# ---------------------------------------------------------------------------
# Print status
# ---------------------------------------------------------------------------
print_status() {
    step "Cluster Status"
    echo ""
    kubectl get pods -n realestate-trust -o wide
    echo ""
    kubectl get svc -n realestate-trust
    echo ""
    log "Access your application:"
    echo "  Frontend:            http://localhost:3000"
    echo "  Transaction Manager: http://localhost:8080/api/v1/health"
    echo "  Identity Service:    http://localhost:8081/api/v1/health"
    echo "  Financing Engine:    http://localhost:8082/api/v1/health"
    echo "  Tokenization Engine: http://localhost:8083/api/v1/health"
    echo "  Ledger Service:      http://localhost:8084/api/v1/health"
    echo ""
    log "Useful commands:"
    echo "  kubectl get pods -n realestate-trust         # List pods"
    echo "  kubectl logs -n realestate-trust -l app=XX   # View logs"
    echo "  kubectl describe pod -n realestate-trust XX   # Debug a pod"
    echo "  ./infra/kind/kind-up.sh --down               # Tear down"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    case "${1:-}" in
        --down)
            check_dependencies
            cluster_down
            ;;
        --reset)
            check_dependencies
            cluster_down
            cluster_up
            build_and_load_images
            deploy_manifests
            print_status
            ;;
        *)
            check_dependencies
            cluster_up
            build_and_load_images
            deploy_manifests
            print_status
            ;;
    esac
}

main "$@"
