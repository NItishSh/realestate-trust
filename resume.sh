#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}" && pwd)"
MANIFESTS_DIR="${SCRIPT_DIR}/infra/kind/manifests"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

log() { echo -e "${GREEN}[✓]${NC} $1"; }
err() { echo -e "${RED}[✗]${NC} $1"; exit 1; }
step() { echo -e "\n${CYAN}━━━ $1 ━━━${NC}"; }

# 4. Wait for Vault pod to be Ready (it should be)
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

# 7. Deploy PostgreSQL (to resolve connection verification dependency)
log "Deploying PostgreSQL database..."
kubectl apply -f "${MANIFESTS_DIR}/postgres.yaml"
log "Waiting for PostgreSQL to be ready..."
sleep 5
kubectl wait --for=condition=Ready pods -l app=postgres -n realestate-trust --timeout=300s
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
kubectl exec -i vault-0 -n vault -- vault write -f transit/keys/kyc-key || true

# 8. Seed Secrets from .env into Vault KV (Fallback for static values)
log "Parsing local .env file and seeding secrets into Vault KV..."
if [ ! -f "${PROJECT_ROOT}/.env" ]; then
  err ".env file not found in project root!"
fi

args=()
while IFS= read -r line || [ -n "$line" ]; do
  [[ "$line" =~ ^# ]] && continue
  [[ -z "$line" ]] && continue
  args+=("$line")
done < "${PROJECT_ROOT}/.env"

kubectl exec -i vault-0 -n vault -- vault kv put secret/realestate-trust/database "${args[@]}"
kubectl exec -i vault-0 -n vault -- vault kv put secret/realestate-trust/grafana admin-password="dynamic_admin_pass"
log "Secrets successfully seeded into Vault KV."

# 9. Create Application Namespace and Apply ESO CRDs
log "Applying SecretStore and ExternalSecret specifications..."
kubectl apply -f "${MANIFESTS_DIR}/namespace.yaml"
kubectl apply -f "${MANIFESTS_DIR}/vault-eso-resources.yaml"

log "Waiting for external secret database syncs..."
sleep 5
ext_secrets=(
  "identity-service-db-secret-sync"
  "transaction-manager-db-secret-sync"
  "financing-engine-db-secret-sync"
  "tokenization-engine-db-secret-sync"
  "ledger-service-db-secret-sync"
  "property-registry-service-db-secret-sync"
)
for es in "${ext_secrets[@]}"; do
  kubectl wait --for=condition=Ready externalsecret/"${es}" -n realestate-trust --timeout=60s
done

log "Vault and External Secrets Operator successfully configured!"


# RESUME kind-up.sh from line 133

# Deploy RabbitMQ
kubectl apply -f "${MANIFESTS_DIR}/rabbitmq.yaml"

log "Waiting for RabbitMQ to be ready..."
sleep 5
kubectl wait --for=condition=Ready pods -l app=rabbitmq -n realestate-trust --timeout=300s

step "Installing Microservices via Helm"
services=("identity-service" "transaction-manager" "financing-engine" "tokenization-engine" "ledger-service" "property-registry-service" "frontend")
for svc in "${services[@]}"; do
    log "Installing ${svc}..."
    helm upgrade --install "${svc}" "${PROJECT_ROOT}/infra/helm/charts/microservice" \
        --namespace realestate-trust \
        --values "${SCRIPT_DIR}/infra/kind/values/${svc}.yaml" \
        --set image.pullPolicy=Never
done
log "All Helm charts deployed locally with automatic Istio sidecar injection."

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
log "Access Observability Dashboards (Unified Gateway):"
echo "  Kiali (Mesh Visualizer):    http://localhost:8080/kiali"
echo "  Grafana (Metrics & Traces): http://localhost:8080/grafana/ (Explore tab for Tempo)"
echo "  Prometheus (Raw Metrics):   http://localhost:8080/prometheus/"
echo ""
log "Starting ArgoCD Port-Forward..."
kubectl port-forward svc/argocd-server -n argocd 8081:443 > /dev/null 2>&1 &
echo "  ArgoCD UI:                  https://localhost:8081 (admin / admin)"
echo ""
