#!/usr/bin/env bash
# =============================================================================
# Universal k6 Performance Test Runner
# Supports local k6 binary or Docker container execution with zero setup
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Always execute from test/perf directory so relative scenario paths resolve identically
cd "${SCRIPT_DIR}"

# Normalize arguments (e.g. if test/perf/scenarios/... is passed, trim test/perf/)
ARGS=()
for arg in "$@"; do
    if [[ "$arg" == test/perf/* ]]; then
        ARGS+=("./${arg#test/perf/}")
    elif [[ "$arg" == ./test/perf/* ]]; then
        ARGS+=("./${arg#./test/perf/}")
    else
        ARGS+=("$arg")
    fi
done

# Detect default target URL
TARGET_URL="${TARGET_URL:-http://localhost:8080}"

if command -v k6 &>/dev/null; then
    echo -e "\033[0;32m[✓]\033[0m Using local k6 binary: $(which k6)"
    k6 run -e TARGET_URL="${TARGET_URL}" "${ARGS[@]}"
elif command -v docker &>/dev/null; then
    echo -e "\033[0;34m[ℹ]\033[0m Local k6 not found. Executing via grafana/k6 Docker container..."

    # Check OS for Docker networking
    DOCKER_NET_ARGS=("--network=host")
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # On macOS, host.docker.internal resolves to the host machine
        if [[ "$TARGET_URL" == "http://localhost:8080"* || "$TARGET_URL" == "http://127.0.0.1:8080"* ]]; then
            TARGET_URL="http://host.docker.internal:8080"
        fi
        DOCKER_NET_ARGS=("-e" "TARGET_URL=${TARGET_URL}" "--add-host=host.docker.internal:host-gateway")
    else
        DOCKER_NET_ARGS=("--network=host" "-e" "TARGET_URL=${TARGET_URL}")
    fi

    docker run --rm -i \
        "${DOCKER_NET_ARGS[@]}" \
        -v "${SCRIPT_DIR}:/scripts" \
        -w /scripts \
        grafana/k6:latest run "${ARGS[@]}"
else
    echo -e "\033[0;31m[✗]\033[0m Neither 'k6' nor 'docker' was found on your system. Please install one of them."
    exit 1
fi
