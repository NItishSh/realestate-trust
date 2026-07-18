#!/usr/bin/env bash
# =============================================================================
# Vault & ESO Deployment Script — RealEstate Trust Platform
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MANIFESTS_DIR="${SCRIPT_DIR}/manifests"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log() { echo -e "${GREEN}[✓]${NC} $1"; }
err() { echo -e "${RED}[✗]${NC} $1"; exit 1; }

# 1. Add Helm repos
log "Adding Helm repositories..."
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo add external-secrets https://charts.external-secrets.io
helm repo update

# 2. Deploy HashiCorp Vault in dev mode
log "Deploying HashiCorp Vault (Dev Mode)..."
helm upgrade --install vault hashicorp/vault \
  --namespace vault \
  --create-namespace \
  --set "server.dev.enabled=true" \
  --set "injector.enabled=false" \
  --wait \
  --timeout 10m0s

# 3. Deploy External Secrets Operator (ESO)
log "Deploying External Secrets Operator..."
helm upgrade --install external-secrets external-secrets/external-secrets \
  --namespace external-secrets \
  --create-namespace \
  --set installCRDs=true \
  --wait \
  --timeout 10m0s

# 4. Wait for Vault pod to be Ready
log "Waiting for vault-0 pod to be ready..."
kubectl wait --for=condition=Ready pod/vault-0 -n vault --timeout=120s

# 5. Enable and Configure Kubernetes Authentication in Vault
log "Configuring Kubernetes authentication inside Vault..."
kubectl exec -i vault-0 -n vault -- vault auth enable kubernetes || true

kubectl exec -i vault-0 -n vault -- vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc:443"

# 6. Apply App Access Policy & Kubernetes Auth Role
log "Configuring read policies and roles inside Vault..."
kubectl exec -i vault-0 -n vault -- vault policy write realestate-trust-policy - <<EOF
path "secret/data/realestate-trust/*" {
  capabilities = ["read"]
}
EOF

kubectl exec -i vault-0 -n vault -- vault write auth/kubernetes/role/realestate-trust-role \
  bound_service_account_names="*" \
  bound_service_account_namespaces="realestate-trust,external-secrets" \
  policies="realestate-trust-policy" \
  ttl=24h

# 7. Seed Secrets from .env into Vault KV
log "Parsing local .env file and seeding secrets into Vault..."
if [ ! -f "${PROJECT_ROOT}/.env" ]; then
  err ".env file not found in project root!"
fi

args=()
while IFS= read -r line || [ -n "$line" ]; do
  # Skip comments and empty lines
  [[ "$line" =~ ^# ]] && continue
  [[ -z "$line" ]] && continue
  args+=("$line")
done < "${PROJECT_ROOT}/.env"

kubectl exec -i vault-0 -n vault -- vault kv put secret/realestate-trust/database "${args[@]}"
log "Secrets successfully seeded into Vault."

# 8. Create Application Namespace and Apply ESO CRDs
log "Applying SecretStore and ExternalSecret specifications..."
kubectl apply -f "${MANIFESTS_DIR}/namespace.yaml"
kubectl apply -f "${MANIFESTS_DIR}/vault-eso-resources.yaml"

log "Waiting for external secret database sync..."
sleep 5
kubectl wait --for=condition=Ready externalsecret/postgres-credentials-sync -n realestate-trust --timeout=60s

log "Vault and External Secrets Operator successfully configured!"
