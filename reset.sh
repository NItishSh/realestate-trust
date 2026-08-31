#!/usr/bin/env bash
# =============================================================================
# Platform Reset Script — Recreate Cluster & Reload Deployments via Terraform
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TF_DIR="${SCRIPT_DIR}/infra/terraform"

echo -e "\033[0;36m━━━ Destroying existing local infrastructure ━━━\033[0m"
kind delete cluster --name realestate-trust 2>/dev/null || true
rm -f "$TF_DIR/terraform.tfstate"* "$TF_DIR/realestate-trust-config"

echo -e "\n\033[0;36m━━━ Provisioning local infrastructure via Terraform ━━━\033[0m"
cd "$TF_DIR"
terraform init
terraform apply -auto-approve

echo -e "\n\033[0;32m[✓]\033[0m Infrastructure provisioned successfully! ArgoCD will now sync all GitOps apps."
echo -e "You can monitor the ArgoCD sync status in the realestate-trust namespace."
