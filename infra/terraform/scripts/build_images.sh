#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="$1"
PROJECT_ROOT="$2"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
step() { echo -e "\n${CYAN}━━━ $1 ━━━${NC}"; }

step "Building Docker images"
targets=("transaction-manager" "identity-service" "financing-engine" "tokenization-engine" "ledger-service" "property-registry-service" "feedback-service")
for target in "${targets[@]}"; do
    image_name="ghcr.io/realestate-trust/monorepo/${target}:latest"
    log "Building ${image_name}..."
    docker build -q -t "${image_name}" --target "${target}" "${PROJECT_ROOT}"
done
docker build -q -t "ghcr.io/realestate-trust/monorepo/frontend:latest" "${PROJECT_ROOT}/frontend"

step "Loading images into Kind cluster"
third_party_images=(
    "postgres:15-alpine"
    "rabbitmq:3.13-management-alpine"
    "hashicorp/vault:1.16.1"
    "quay.io/keycloak/keycloak:26.7.2"
)
for img in "${third_party_images[@]}"; do
    log "Pulling and loading ${img} into Kind..."
    docker pull -q "${img}" || true
    kind load docker-image "${img}" --name "${CLUSTER_NAME}" || true
done

images=(
    "ghcr.io/realestate-trust/monorepo/transaction-manager:latest"
    "ghcr.io/realestate-trust/monorepo/identity-service:latest"
    "ghcr.io/realestate-trust/monorepo/financing-engine:latest"
    "ghcr.io/realestate-trust/monorepo/tokenization-engine:latest"
    "ghcr.io/realestate-trust/monorepo/ledger-service:latest"
    "ghcr.io/realestate-trust/monorepo/property-registry-service:latest"
    "ghcr.io/realestate-trust/monorepo/feedback-service:latest"
    "ghcr.io/realestate-trust/monorepo/frontend:latest"
)

for img in "${images[@]}"; do
    log "Loading ${img}..."
    kind load docker-image "${img}" --name "${CLUSTER_NAME}"
done
log "All images loaded into Kind."
