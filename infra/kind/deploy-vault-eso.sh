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

# 1. Wait for Vault StatefulSet to be created by ArgoCD
log "Waiting for Vault to be deployed by ArgoCD..."
while ! kubectl get statefulset vault -n vault > /dev/null 2>&1; do
  sleep 2
done

# 2. Wait for Vault pod to be created and Ready
log "Waiting for vault-0 pod to be created..."
while ! kubectl get pod vault-0 -n vault > /dev/null 2>&1; do
  sleep 2
done
log "Waiting for vault-0 pod to be ready..."
kubectl wait --for=condition=Ready pod/vault-0 -n vault --timeout=600s

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
path "database/creds/realestate-trust-role" {
  capabilities = ["read"]
}
path "transit/encrypt/kyc-key" {
  capabilities = ["update"]
}
path "transit/decrypt/kyc-key" {
  capabilities = ["update"]
}
EOF

kubectl exec -i vault-0 -n vault -- vault write auth/kubernetes/role/realestate-trust-role \
  bound_service_account_names="*" \
  bound_service_account_namespaces="realestate-trust,external-secrets,observability" \
  policies="realestate-trust-policy" \
  ttl=24h

# 7. Wait for PostgreSQL (deployed by ArgoCD)
log "Waiting for PostgreSQL to be deployed by ArgoCD..."
while ! kubectl get pod postgres-0 -n realestate-trust > /dev/null 2>&1; do
  sleep 2
done
log "Waiting for postgres-0 pod to be ready..."
kubectl wait --for=condition=Ready pod/postgres-0 -n realestate-trust --timeout=900s
log "Postgres pod is Ready. Waiting 15s for database initialization to complete..."
sleep 15

# 8. Enable and Configure Vault Secrets Engines (Database & Transit)
log "Enabling and configuring Vault Database Secrets Engine..."
kubectl exec -i vault-0 -n vault -- vault secrets enable database || true

kubectl exec -i vault-0 -n vault -- vault write database/config/postgres \
  plugin_name=postgresql-database-plugin \
  allowed_roles="realestate-trust-role" \
  connection_url="postgresql://{{username}}:{{password}}@postgres.realestate-trust.svc.cluster.local:5432/postgres?sslmode=disable" \
  username="postgres" \
  password="postgres"

kubectl exec -i vault-0 -n vault -- vault write database/roles/realestate-trust-role \
  db_name=postgres \
  creation_statements="CREATE ROLE \"{{name}}\" WITH SUPERUSER LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';" \
  default_ttl="720h" \
  max_ttl="720h"

log "Enabling and configuring Vault Transit Secrets Engine..."
kubectl exec -i vault-0 -n vault -- vault secrets enable transit || true
kubectl exec -i vault-0 -n vault -- vault write -f transit/keys/kyc-key

# 8. Seed Secrets from .env into Vault KV (Fallback for static values)
log "Parsing local .env file and seeding secrets into Vault KV..."
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
kubectl exec -i vault-0 -n vault -- vault kv put secret/realestate-trust/grafana admin-password="dynamic_admin_pass"
log "Secrets successfully seeded into Vault KV."

# 9. Apply SecretStores and ExternalSecrets
log "Ensuring namespaces and applying SecretStores & ExternalSecrets..."
kubectl create namespace observability --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace realestate-trust --dry-run=client -o yaml | kubectl apply -f -

log "Waiting for External Secrets Operator webhook to be ready..."
while [ $(kubectl get pods -l app.kubernetes.io/name=external-secrets-webhook -n external-secrets --no-headers 2>/dev/null | wc -l) -eq 0 ]; do
  sleep 2
done
kubectl wait --for=condition=Ready pods -l app.kubernetes.io/name=external-secrets-webhook -n external-secrets --timeout=420s
sleep 5

log "Applying vault-eso-resources.yaml (with retry for webhook warm-up)..."
applied=false
for i in $(seq 1 12); do
  if kubectl apply -f "${SCRIPT_DIR}/eso-manifests/vault-eso-resources.yaml"; then
    applied=true
    break
  fi
  echo "External Secrets admission webhook is warming up. Retrying in 5s (attempt $i/12)..."
  sleep 5
done

if [ "$applied" = false ]; then
  err "Failed to apply vault-eso-resources.yaml after multiple attempts."
fi

log "Waiting for external secret database syncs..."
kubectl wait --for=condition=Ready externalsecret --all -n realestate-trust --timeout=300s
kubectl wait --for=condition=Ready externalsecret --all -n observability --timeout=300s

log "Vault and External Secrets Operator successfully configured!"
