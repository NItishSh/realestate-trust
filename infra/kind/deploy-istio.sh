#!/usr/bin/env bash
set -euo pipefail

# Deploy Istio using Helm
echo "━━━ Deploying Istio Service Mesh ━━━"

echo "Installing Istio Base..."
helm upgrade --install istio-base istio/base -n istio-system --create-namespace --wait

echo "Installing Istiod (Control Plane)..."
helm upgrade --install istiod istio/istiod -n istio-system --wait

echo "Installing Istio Ingress Gateway (mapping NodePort 30080)..."
helm upgrade --install istio-ingress istio/gateway \
  -n istio-system \
  --set service.type=NodePort \
  --set "service.ports[0].name=http2" \
  --set "service.ports[0].port=80" \
  --set "service.ports[0].targetPort=80" \
  --set "service.ports[0].nodePort=30080" \
  --wait

echo "Labeling realestate-trust namespace for sidecar injection..."
kubectl label namespace realestate-trust istio-injection=enabled --overwrite

echo "[✓] Istio successfully installed and namespace configured!"
