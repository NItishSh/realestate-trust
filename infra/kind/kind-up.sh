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
    local targets=("transaction-manager" "identity-service" "financing-engine" "tokenization-engine" "ledger-service")
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

print_status() {
    step "Cluster Status"
    echo ""
    kubectl get pods -n realestate-trust -o wide
    echo ""
    log "Access ArgoCD UI:"
    echo "  kubectl port-forward svc/argocd-server -n argocd 8080:443"
    echo "  Password: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath=\"{.data.password}\" | base64 -d"
    echo ""
    log "Useful commands:"
    echo "  kubectl get applications -n argocd           # List ArgoCD apps"
    echo "  kubectl get pods -n realestate-trust         # List application pods"
}

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
            deploy_argocd_and_postgres
            print_status
            ;;
        *)
            check_dependencies
            cluster_up
            build_and_load_images
            deploy_argocd_and_postgres
            print_status
            ;;
    esac
}

main "$@"
