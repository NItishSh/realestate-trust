#!/usr/bin/env bash
# =============================================================================
# Platform Reset Script — Recreate Cluster & Reload Deployments via Terraform
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TF_DIR="${SCRIPT_DIR}/infra/terraform"

echo -e "\033[0;36m━━━ Destroying existing local infrastructure ━━━\033[0m"
if [ -d "$TF_DIR/.terraform" ]; then
    cd "$TF_DIR"
    terraform destroy -auto-approve || echo -e "\033[1;33m[!] Destroy had issues or nothing to destroy. Continuing...\033[0m"
else
    # Fallback to destroy the kind cluster manually if terraform state doesn't exist
    kind delete cluster --name realestate-trust 2>/dev/null || true
fi

echo -e "\n\033[0;36m━━━ Provisioning local infrastructure via Terraform ━━━\033[0m"
cd "$TF_DIR"
terraform init
terraform apply -auto-approve

echo -e "\n\033[0;32m[✓]\033[0m Infrastructure provisioned successfully! ArgoCD will now sync all GitOps apps."
echo -e "You can monitor the ArgoCD sync status in the realestate-trust namespace."
