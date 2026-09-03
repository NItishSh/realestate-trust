# 📈 RealEstate-Trust: Capacity & Pod Right-Sizing Report

* **Generated**: 2026-09-03T11:01:25.714Z
* **Target Namespace**: `realestate-trust`
* **Telemetry Engine**: OpenCost (Open Source FinOps for Kubernetes)
* **Measured Active Cost**: $0.0283/hr across monitored workloads

---

## 🔍 Live Pod Utilization & OpenCost Efficiency

| Pod Name | Avg CPU Usage | Avg RAM Usage | CPU Efficiency | RAM Efficiency | Recommended FinOps Sizing |
| :--- | :---: | :---: | :---: | :---: | :---: |
| `feedback-service-6494457c4d-c8xkf` | 4.9m | 45.1 MiB | 3.3% | 23.5% | **`50m / 64Mi`** |
| `feedback-service-6494457c4d-fjnl9` | 1.4m | 41.1 MiB | 0.9% | 21.4% | **`50m / 64Mi`** |
| `feedback-service-6494457c4d-lqw79` | 4.7m | 47.7 MiB | 3.1% | 24.9% | **`50m / 64Mi`** |
| `financing-engine-78bd549dbc-rzmb5` | 7.7m | 46.6 MiB | 5.1% | 24.3% | **`50m / 64Mi`** |
| `financing-engine-78bd549dbc-xjx68` | 1.0m | 49.5 MiB | 0.7% | 25.8% | **`50m / 64Mi`** |
| `frontend-56ccd6c57-4sfzw` | 24.0m | 117.5 MiB | 12.0% | 45.9% | **`50m / 147Mi`** |
| `frontend-56ccd6c57-c9g54` | 5.3m | 71.3 MiB | 2.6% | 27.8% | **`50m / 90Mi`** |
| `frontend-56ccd6c57-h6vrb` | 3.9m | 73.3 MiB | 2.0% | 28.6% | **`50m / 92Mi`** |
| `identity-service-67b46b859-6rhpl` | 17.8m | 47.8 MiB | 5.1% | 18.7% | **`50m / 64Mi`** |
| `identity-service-67b46b859-crv7p` | 7.0m | 48.4 MiB | 2.0% | 18.9% | **`50m / 64Mi`** |
| `keycloak-0` | 58.0m | 380.7 MiB | 58.0% | 297.4% | **`250m / 512Mi`** |
| `ledger-service-5c48f5f678-px5bn` | 4.1m | 46.1 MiB | 2.7% | 24.0% | **`50m / 64Mi`** |
| `ledger-service-f6d9485fb-5vbdj` | 4.2m | 43.7 MiB | 2.8% | 22.8% | **`50m / 64Mi`** |
| `ledger-service-f6d9485fb-9w9fb` | 5.6m | 41.8 MiB | 3.7% | 21.8% | **`50m / 64Mi`** |
| `postgres-0` | 27.4m | 38.3 MiB | 11.0% | 14.9% | **`250m / 256Mi`** |
| `property-registry-service-5d84567d6c-n2s9s` | 1.6m | 54.9 MiB | 1.1% | 28.6% | **`50m / 69Mi`** |
| `property-registry-service-5d84567d6c-v8xqr` | 1.1m | 45.5 MiB | 0.8% | 23.7% | **`50m / 64Mi`** |
| `property-registry-service-845494fc66-l22cf` | 2.6m | 50.6 MiB | 1.8% | 27.4% | **`50m / 64Mi`** |
| `property-registry-service-845494fc66-x578x` | 8.9m | 47.2 MiB | 6.3% | 25.5% | **`50m / 64Mi`** |
| `rabbitmq-56d7d88fc7-2hmkz` | 10.4m | 61.6 MiB | 10.4% | 12.0% | **`200m / 256Mi`** |
| `rabbitmq-6d65b85b96-6fwg6` | 14.4m | 89.3 MiB | 14.4% | 17.4% | **`200m / 256Mi`** |
| `rabbitmq-6d65b85b96-s4xg4` | 0.0m | 48.1 MiB | 0.0% | 9.4% | **`200m / 256Mi`** |
| `tokenization-engine-595d5f99f6-6fnn9` | 5.3m | 44.1 MiB | 3.5% | 23.0% | **`50m / 64Mi`** |
| `tokenization-engine-595d5f99f6-jkrxk` | 5.0m | 44.8 MiB | 3.3% | 23.4% | **`50m / 64Mi`** |
| `tokenization-engine-b55bd7c84-mm8l7` | 4.4m | 46.7 MiB | 0.9% | 14.6% | **`50m / 64Mi`** |
| `transaction-manager-65fd5575d5-4hq4w` | 9.0m | 46.8 MiB | 1.5% | 12.2% | **`50m / 64Mi`** |
| `transaction-manager-65fd5575d5-96cmm` | 1.5m | 52.7 MiB | 0.2% | 13.7% | **`50m / 66Mi`** |

---

## 🎯 FinOps Right-Sizing Methodology
* **CPU Request**: $P_{95}(\text{CPU}) \times 1.20$ (20% safety headroom above peak).
* **Memory Request**: $P_{99}(\text{RAM}) \times 1.25$ (25% safety buffer to guarantee against OOMKills).
* **Memory Limit**: $1.50 \times \text{Memory Request}$.
