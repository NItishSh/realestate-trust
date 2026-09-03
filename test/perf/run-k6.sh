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

mkdir -p "${SCRIPT_DIR}/reports"

# Optional Prometheus Remote Write Output
if [[ "${K6_PROMETHEUS_OUTPUT:-false}" == "true" ]]; then
    PROMETHEUS_URL="${K6_PROMETHEUS_RW_SERVER_URL:-http://localhost:9090/api/v1/write}"
    echo -e "\033[0;34m[ℹ]\033[0m Streaming live k6 metrics to Prometheus Remote Write: ${PROMETHEUS_URL}"
    ARGS+=("-o" "experimental-prometheus-rw")
fi

set +e
if command -v k6 &>/dev/null; then
    echo -e "\033[0;32m[✓]\033[0m Using local k6 binary: $(which k6)"
    k6 run -e TARGET_URL="${TARGET_URL}" "${ARGS[@]}"
    RUN_EXIT_CODE=$?
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
    RUN_EXIT_CODE=$?
else
    echo -e "\033[0;31m[✗]\033[0m Neither 'k6' nor 'docker' was found on your system. Please install one of them."
    exit 1
fi
set -e

# =============================================================================
# Benchmark History Archiver
# Extracts results from reports/latest.json and appends to reports/history.csv
# =============================================================================
if [ -f "${SCRIPT_DIR}/reports/latest.json" ]; then
    COMMIT_HASH="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
    TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

    RECORD=$(python3 -c '
import json, sys
try:
    with open("reports/latest.json") as f:
        d = json.load(f)
    m = d.get("metrics", {})
    reqs = m.get("http_reqs", {}).get("values", {})
    dur = m.get("http_req_duration", {}).get("values", {})
    failed = m.get("http_req_failed", {}).get("values", {})
    scenario = d.get("scenarioName", "Test")
    vus = int(m.get("vus", {}).get("values", {}).get("value", 1))
    total_reqs = int(reqs.get("count", 0))
    rate_val = reqs.get("rate", 0)
    rps = f"{rate_val:.2f}"
    p95_val = dur.get("p(95)", 0)
    p95 = f"{p95_val:.2f}"
    p99_val = dur.get("p(99)", 0)
    p99 = f"{p99_val:.2f}"
    fail_rate = failed.get("rate", 0) * 100
    err_rate = f"{fail_rate:.2f}%"
    status = "PASSED" if int(sys.argv[1]) == 0 else "FAILED"
    print(f"{sys.argv[2]},{sys.argv[3]},{scenario},{vus},{total_reqs},{rps},{p95},{p99},{err_rate},{status}")
except Exception as e:
    pass
' "$RUN_EXIT_CODE" "$TIMESTAMP" "$COMMIT_HASH")

    if [ -n "$RECORD" ]; then
        CSV_FILE="${SCRIPT_DIR}/reports/history.csv"
        if [ ! -f "$CSV_FILE" ]; then
            echo "timestamp,commit,scenario,vus,total_requests,rps,p95_ms,p99_ms,error_rate,status" > "$CSV_FILE"
        fi
        echo "$RECORD" >> "$CSV_FILE"
        echo -e "\033[0;32m[✓]\033[0m Performance benchmark archived to ${CSV_FILE}"

        # Archive timestamped copy of HTML report
        if [ -f "${SCRIPT_DIR}/reports/latest.html" ]; then
            REPORT_COPY="${SCRIPT_DIR}/reports/report-$(date +%Y%m%d-%H%M%S).html"
            cp "${SCRIPT_DIR}/reports/latest.html" "$REPORT_COPY"
            echo -e "\033[0;32m[✓]\033[0m HTML report saved to ${REPORT_COPY}"
        fi
    fi
fi

exit $RUN_EXIT_CODE
