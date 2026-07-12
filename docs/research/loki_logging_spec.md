# Grafana Loki Logging & LogQL Specification Research

This document outlines the design and operational specifications for the log aggregation system based on **Grafana Loki**. In a high-value real estate transaction platform, structured logs are critical for legal auditing, debugging bank discrepancies, and detecting fraud.

---

## 1. Loki Logging Architecture

We implement a pull-push aggregation pipeline using **Promtail** as a DaemonSet to collect logs and forward them to the central Loki storage cluster.

```
 [ Pod Containers ] ---> Writes to stdout (Structured JSON)
         |
         v (Written to Host Path)
 `/var/log/pods/`
         |
         v (Scraped by)
 [ Promtail DaemonSet ] ---> Parses JSON & applies labels
         |
         v (Pushes over secure gRPC)
 [ Central Loki Ingestion Hub ] ---> Index & Chunks Storage (AWS S3)
```

---

## 2. Standardized JSON Log Schema

All Go microservices must log exclusively in structured JSON format to stdout. The platform standardizes on the following JSON schema:

```json
{
  "time": "2026-07-11T13:28:00Z",
  "level": "INFO",
  "service": "transaction-manager",
  "caller": "internal/escrow/manager.go:142",
  "transaction_id": "d03f5db9-6014-41d3-a3d2-40df48fe75e0",
  "escrow_account_id": "c3b9a7c6-7a1a-46be-8b26-bb2b2c9c7f66",
  "bank_partner_id": "2b2a1a0c-5d6e-7f8a-9b0c-1d2e3f4a5b6c",
  "message": "Escrow virtual account funded successfully",
  "details": {
    "amount_received": 10000000.00,
    "utr": "UTRN88291038102"
  }
}
```

### Go Configuration (`slog` Integration)
In Go, we configure the `log/slog` library to structure all outputs:

```go
package logging

import (
	"os"
	"log/slog"
)

// InitStructuredLogger initializes the global logger to format as JSON.
func InitStructuredLogger(serviceName string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs([]slog.Attr{
		slog.String("service", serviceName),
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
```

---

## 3. Promtail Scraping Configuration

Promtail runs as a DaemonSet on every node in the application cluster. It reads log logs, parses the JSON structure, and extracts key fields as indexing labels in Loki.

#### `promtail-config.yaml` Snippet:
```yaml
scrape_configs:
- job_name: kubernetes-pods
  kubernetes_sd_configs:
  - role: pod
  pipeline_stages:
  # 1. Parse JSON log payload
  - json:
      expressions:
        level: level
        service: service
        transaction_id: transaction_id
        escrow_account_id: escrow_account_id
  # 2. Extract parsed fields as Loki search index labels
  - labels:
      level:
      service:
      transaction_id:
      escrow_account_id:
  # 3. Suppress logs matching high-volume debug logs in non-dev namespaces
  - match:
      selector: '{namespace!="dev", level="DEBUG"}'
      action: drop
```

*Note: Limiting the number of high-cardinality labels (like `transaction_id` or `escrow_account_id`) is essential to prevent Loki index bloat. Highly unique IDs should be parsed dynamically inside LogQL queries using Loki's `| json` filter stage rather than indexing them permanently.*

---

## 4. LogQL Audit & Debugging Queries

Grafana Loki uses LogQL (Log Query Language). Below are critical queries for tracing and auditing payments:

### 4.1 Trace a Specific Transaction's Lifecycle
Queries the logs of all services for events matching a specific `transaction_id`:
```logql
{namespace="realestate-trust"} | json | transaction_id = "d03f5db9-6014-41d3-a3d2-40df48fe75e0"
```

### 4.2 Track Bank Webhook Failures
Filters logs from the `transaction-manager` to locate failed incoming webhooks:
```logql
{service="transaction-manager", level="ERROR"} | json | message =~ "(?i).*webhook.*failed.*"
```

### 4.3 Aggregate Error Rates per Microservice
Computes the count of errors occurring per service over a rolling 1-hour window:
```logql
sum by (service) (count_over_time({namespace="realestate-trust", level="ERROR"}[1h]))
```

### 4.4 Find Latency Outliers on Bank Payouts
Finds payout confirmation logs where the duration took longer than 2.5 seconds:
```logql
{service="transaction-manager"} | json | message =~ ".*payout.*completed.*" | duration > 2.5s
```
