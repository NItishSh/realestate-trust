#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

if [ ! -f "${PROJECT_ROOT}/.env" ]; then
    echo "Error: .env file not found in project root!"
    exit 1
fi

kubectl create namespace realestate-trust --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic postgres-credentials \
  --namespace realestate-trust \
  --from-env-file="${PROJECT_ROOT}/.env" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Successfully created postgres-credentials secret from .env file."
