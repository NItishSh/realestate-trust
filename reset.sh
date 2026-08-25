#!/usr/bin/env bash
# =============================================================================
# Platform Reset Script — Recreate Cluster & Reload Deployments
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"${SCRIPT_DIR}/infra/kind/kind-up.sh" --reset "$@"
