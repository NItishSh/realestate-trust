# ⚡ RealEstate-Trust: Performance, Capacity & 24-Hour Soak Testing Guide

An institutional guide to performance verification, saturation threshold discovery, long-duration soak testing, and empirical Kubernetes workload right-sizing for the **RealEstate-Trust Platform**.

---

## 🎯 Objectives & Architecture

The performance testing framework fulfills three critical engineering requirements:
1. **Capacity & Throughput Discovery**: Determines the exact requests-per-second (RPS) and concurrency ceilings achievable given a specific cluster configuration ($X$ CPU cores and $Y$ GB RAM).
2. **24-Hour Soak & Memory Leak Detection**: Verifies that microservices, PostgreSQL connections, RabbitMQ channels, and Istio sidecar proxies maintain rock-solid stability under sustained load with zero memory creep.
3. **Empirical Workload Sizing**: Replaces speculative CPU/memory guesswork by feeding empirical $p95$ and $p99$ telemetry from long-running tests directly into the **Vertical Pod Autoscaler (VPA)**, **OpenCost**, and our automated **`re-cli finops rightsize`** engine.

```mermaid
flowchart LR
    subgraph TestSuite["🧪 k6 Performance Suite"]
        Smoke["make perf-smoke<br/>(1-2m Sanity)"]
        Load["make perf-load<br/>(12m Baseline)"]
        Stress["make perf-stress<br/>(23m Saturation)"]
        Soak["make perf-soak<br/>(24h Soak Plateau)"]
    end

    subgraph Cluster["☸️ RealEstate-Trust Kubernetes Cluster"]
        Gateway["Istio Gateway (:8080)"]
        Services["7 Go Microservices<br/>+ Next.js Frontend"]
        Telemetry["Prometheus + Grafana"]
        VPA["VPA Recommendations"]
    end

    subgraph FinOps["💰 Continuous Right-Sizing"]
        Rightsize["re-cli finops rightsize"]
        Values["infra/kind/values/*.yaml<br/>(CPU / Memory Tuned)"]
    end

    TestSuite --> Gateway
    Gateway --> Services
    Services -.-> Telemetry
    Services -.-> VPA
    VPA -.-> Rightsize
    Rightsize --> Values
```

---

## 📊 Test Hierarchy

| Target | Command | Duration | Concurrency (VUs) | Primary Goal |
| :--- | :--- | :---: | :---: | :--- |
| **Smoke Test** | `make perf-smoke` | 1 minute | 3 VUs | Verifies route connectivity, JWT auth flow, and basic endpoint availability. |
| **User Journeys (100% Coverage)** | `make perf-journeys` | Configurable (default 1m) | 10 VUs | Stateful persona journeys covering 100% of all microservice REST endpoints. |
| **Standard Load** | `make perf-load` | 12 minutes | 20 VUs (~50 RPS) | Establishes standard production baseline $p95$ and $p99$ latency SLAs. |
| **Stress & Saturation** | `make perf-stress` | 23 minutes | Up to 200 VUs (300+ RPS) | Ramps past capacity to find breaking points, queue buildup, and bottleneck services. |
| **24-Hour Sustained Soak** | `make perf-soak` | 24 hours | 50 VUs sustained | Sustained peak plateau to detect memory leaks, connection pool exhaustion, and long-term degradation. |

---

## 🚀 Running Performance Tests

Tests can be executed with **zero host dependencies** (using the official `grafana/k6:latest` container) or using a local `k6` binary if installed.

### 1. Smoke Test (Sanity)
```bash
make perf-smoke
```

### 2. End-to-End User Journeys Test (100% Endpoint Coverage)
Exercises all 30 REST endpoints across all 7 microservices and the Next.js frontend across 4 distinct personas:
* **Buyer Persona (60%)**: Register $\rightarrow$ Submit KYC $\rightarrow$ Search $\rightarrow$ Unlock Docs $\rightarrow$ Purchase $\rightarrow$ Fund Escrow $\rightarrow$ Apply Loan $\rightarrow$ Feedback.
* **Seller / Broker Persona (20%)**: Create Property Listing $\rightarrow$ Edit Details $\rightarrow$ Fractional Pool Tokenization $\rightarrow$ Purchase Shares $\rightarrow$ Update Status.
* **Compliance Officer Persona (10%)**: Inspect Users Registry $\rightarrow$ User Detail $\rightarrow$ Verify Title Insurance $\rightarrow$ Review Loan Portfolio.
* **Auditor & Admin Persona (10%)**: Write Cryptographic Block $\rightarrow$ Fetch Chain $\rightarrow$ Block Verification $\rightarrow$ Review Feedback $\rightarrow$ List Transactions $\rightarrow$ Logout.

```bash
make perf-journeys

# Custom concurrency & duration
JOURNEY_DURATION=5m VUS=25 make perf-journeys
```

### 3. Standard Load Test
```bash
make perf-load
```

### 3. Stress / Breaking Point Test
```bash
make perf-stress
```

### 4. 24-Hour Sustained Soak Test
By default, `make perf-soak` runs a 24-hour cycle:
- **1 Hour**: Gradual ramp-up to peak load
- **22 Hours**: Sustained maximum plateau
- **1 Hour**: Gradual ramp-down to zero

```bash
make perf-soak
```

#### Custom Soak Test Duration (Validation Dry-Runs)
You can customize the soak durations via environment variables for rapid validation:
```bash
# Example: 10-minute soak validation run (1m ramp, 8m plateau, 1m cooldown)
SOAK_RAMP=1m SOAK_PLATEAU=8m SOAK_COOLDOWN=1m SOAK_TARGET_VUS=25 make perf-soak
```

---

## 📈 Monitoring Live Telemetry During Runs

While the test is executing, monitor real-time behavior via the built-in observability stack:

* **Unified Gateway Portal**: `http://localhost:8080`
* **Grafana Dashboards**: `http://localhost:8080/grafana/` (User: `admin`, Password: run `make grafana-password`)
  * *Kubernetes / Compute Resources / Pod*: Inspect CPU and memory slopes over time to detect memory leaks.
  * *Istio Service Mesh*: Inspect HTTP request volume, error rates, and $p95$/$p99$ response times per microservice.
* **Kiali Mesh Visualizer**: `http://localhost:8080/kiali/`
* **Prometheus Metrics**: `http://localhost:8080/prometheus/`

---

## 🔄 Translating Soak Results into Optimal Pod Resources

Once the sustained load test has run, you can extract exact CPU and Memory recommendations:

### 1. Inspect Vertical Pod Autoscaler (VPA)
Query the in-cluster VPA recommendations:
```bash
kubectl get vpa -n realestate-trust
```

Output displays the empirically derived sizing recommendations:
```
NAME                          MODE   CPU     MEM       AGE
transaction-manager-vpa       Off    52m     68Mi      12h
identity-service-vpa          Off    45m     62Mi      12h
property-registry-service-vpa Off    35m     54Mi      12h
tokenization-engine-vpa       Off    40m     58Mi      12h
ledger-service-vpa            Off    38m     52Mi      12h
financing-engine-vpa          Off    32m     48Mi      12h
feedback-service-vpa          Off    25m     40Mi      12h
```

### 2. Preview and Apply Automated Right-Sizing
Run our FinOps CLI engine, which ingests the empirical peak data, adds a **+20% safety headroom**, sets memory limits at 1.5x, and preserves YAML formatting:
```bash
# Dry-run preview
make finops-rightsize-dryrun

# Apply changes to infra/kind/values/*.yaml
make finops-rightsize
```

### 3. GitOps Deployment
Commit the updated values. ArgoCD automatically syncs the newly tuned CPU and memory allocations to the cluster.

---

## 📜 Historical Benchmark Tracking & Trend Comparison

Performance results are automatically preserved across all runs to track regressions over time:

### 1. Version-Controlled Benchmark History (`test/perf/reports/history.csv`)
Every performance run automatically appends key benchmark metrics (timestamp, commit hash, scenario, VUs, total requests, RPS, $p95$, $p99$, error rate, and status) to `test/perf/reports/history.csv`.

To view the historical comparison table directly in your terminal:
```bash
make perf-history
```

Example output:
```
timestamp             commit   scenario       vus  total_requests  rps    p95_ms  p99_ms  error_rate  status
2026-09-03T07:37:36Z  46394da  smoke          3    227             10.37  366.35  512.10  0.00%       PASSED
2026-09-03T08:15:00Z  da550a3  Standard Load  20   36000           50.00  240.12  410.50  0.02%       PASSED
```

### 2. Standalone Interactive HTML Reports
Every run generates a self-contained, interactive HTML dashboard (`test/perf/reports/latest.html`) and timestamped snapshot (`test/perf/reports/report-<timestamp>.html`).

To view the latest visual report:
```bash
make perf-report
```

### 3. Streaming to In-Cluster Prometheus (Remote Write)
For live real-time graph streaming into Prometheus and Grafana during tests, enable `K6_PROMETHEUS_OUTPUT`:
```bash
K6_PROMETHEUS_OUTPUT=true make perf-load
```
k6 will stream metrics directly into `prometheus-server` via Prometheus Remote Write.
