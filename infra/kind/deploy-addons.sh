#!/usr/bin/env bash
set -euo pipefail

ISTIO_VERSION="1.30.3"
ADDONS=("prometheus" "grafana" "jaeger" "kiali")
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "━━━ Deploying Istio Observability Addons ━━━"

for addon in "${ADDONS[@]}"; do
  echo "Installing ${addon}..."
  kubectl apply -f "https://raw.githubusercontent.com/istio/istio/${ISTIO_VERSION}/samples/addons/${addon}.yaml"
done

echo "Waiting for observability deployments to roll out..."
sleep 5

# 1. Patch Grafana (Subpath configuration and Jaeger integration)
echo "Patching Grafana ConfigMap & Deployment..."
cat << 'EOF' > /tmp/patch_grafana.py
import sys, json
try:
    cm = json.load(sys.stdin)
    out_lines = [
        "apiVersion: 1",
        "datasources:",
        "- access: proxy",
        "  editable: true",
        "  isDefault: true",
        "  jsonData:",
        "    timeInterval: 15s",
        "  name: Prometheus",
        "  orgId: 1",
        "  type: prometheus",
        "  url: http://prometheus:9090/prometheus",
        "- access: proxy",
        "  editable: true",
        "  isDefault: false",
        "  jsonData:",
        "    timeInterval: 5s",
        "  name: Loki",
        "  orgId: 1",
        "  type: loki",
        "  url: http://loki:3100",
        "- access: proxy",
        "  editable: true",
        "  isDefault: false",
        "  name: Jaeger",
        "  orgId: 1",
        "  type: jaeger",
        "  url: http://tracing:80/jaeger"
    ]
    cm["data"]["datasources.yaml"] = "\n".join(out_lines) + "\n"
    print(json.dumps(cm))
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    sys.exit(1)
EOF
kubectl get configmap grafana -n istio-system -o json | python3 /tmp/patch_grafana.py | kubectl apply -f -
rm -f /tmp/patch_grafana.py

# Patch the Deployment environment variables for Grafana subpath routing
kubectl patch deployment grafana -n istio-system --type=json -p '[{"op": "add", "path": "/spec/template/spec/containers/0/env/-", "value": {"name": "GF_SERVER_ROOT_URL", "value": "http://localhost:8080/grafana/"}}, {"op": "add", "path": "/spec/template/spec/containers/0/env/-", "value": {"name": "GF_SERVER_SERVE_FROM_SUB_PATH", "value": "true"}}]'

# 2. Patch Jaeger (Subpath configuration)
echo "Patching Jaeger for /jaeger subpath..."
kubectl patch deployment jaeger -n istio-system --type=json -p '[{"op": "add", "path": "/spec/template/spec/containers/0/env", "value": [{"name": "QUERY_BASE_PATH", "value": "/jaeger"}]}]'

# 3. Patch Prometheus (Subpath configuration & readiness probes)
echo "Patching Prometheus for /prometheus subpath and readiness/liveness paths..."
kubectl patch deployment prometheus -n istio-system --type=json -p '[{"op": "add", "path": "/spec/template/spec/containers/1/args/-", "value": "--web.external-url=http://localhost:8080/prometheus/"}, {"op": "add", "path": "/spec/template/spec/containers/1/args/-", "value": "--web.route-prefix=/prometheus/"}, {"op": "replace", "path": "/spec/template/spec/containers/1/livenessProbe/httpGet/path", "value": "/prometheus/-/healthy"}, {"op": "replace", "path": "/spec/template/spec/containers/1/readinessProbe/httpGet/path", "value": "/prometheus/-/ready"}]'

# 4. Patch Kiali (Configure subpath integrations for Prometheus & Grafana)
echo "Patching Kiali ConfigMap for subpath integrations..."
cat << 'EOF' > /tmp/patch_kiali.py
import sys, json
try:
    cm = json.load(sys.stdin)
    config_yaml = cm["data"]["config.yaml"]
    lines = config_yaml.splitlines()
    out_lines = []
    in_ext = False
    skip_next_enabled = False
    for line in lines:
        if line.strip() == "external_services:":
            in_ext = True
            out_lines.append(line)
            continue
        if in_ext:
            if line and not line.startswith(" "):
                in_ext = False
            if in_ext:
                if "prometheus:" in line:
                    out_lines.append("  prometheus:")
                    out_lines.append('    url: "http://prometheus.istio-system.svc.cluster.local:9090/prometheus"')
                    skip_next_enabled = True
                    continue
                if skip_next_enabled and "enabled: true" in line:
                    out_lines.append("    enabled: true")
                    skip_next_enabled = False
                    continue
                if "tracing:" in line:
                    out_lines.append("  grafana:")
                    out_lines.append("    enabled: true")
                    out_lines.append('    url: "http://localhost:8080/grafana"')
                    out_lines.append('    in_cluster_url: "http://grafana.istio-system.svc.cluster.local:3000/grafana"')
        out_lines.append(line)
    cm["data"]["config.yaml"] = "\n".join(out_lines) + "\n"
    print(json.dumps(cm))
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    sys.exit(1)
EOF
kubectl get configmap kiali -n istio-system -o json | python3 /tmp/patch_kiali.py | kubectl apply -f -
rm -f /tmp/patch_kiali.py

# 5. Install Loki and Promtail
echo "Installing Loki and Promtail via Helm..."
helm upgrade --install loki grafana/loki-stack \
  --namespace istio-system \
  --set loki.persistence.enabled=false \
  --set promtail.enabled=true \
  --set loki.config.limits_config.volume_enabled=true \
  --set loki.image.tag=2.9.3 \
  --set loki.readinessProbe.timeoutSeconds=10 \
  --set loki.livenessProbe.timeoutSeconds=10

# 6. Apply Ingress Routing Gateway VirtualService
echo "Deploying gateway routing VirtualService..."
kubectl apply -f "${SCRIPT_DIR}/manifests/observability-gateway.yaml"

# 7. Wait for all rollouts to complete
echo "Waiting for observability pods to be ready..."
kubectl rollout restart deployment/kiali -n istio-system
kubectl rollout status deployment/prometheus -n istio-system --timeout=300s
kubectl rollout status deployment/grafana -n istio-system --timeout=300s
kubectl rollout status deployment/jaeger -n istio-system --timeout=300s
kubectl rollout status deployment/kiali -n istio-system --timeout=300s
kubectl rollout status statefulset/loki -n istio-system --timeout=300s
kubectl rollout status daemonset/loki-promtail -n istio-system --timeout=300s

echo "[✓] Observability addons and Loki log aggregation successfully installed!"
