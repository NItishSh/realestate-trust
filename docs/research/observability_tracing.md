# Observability, Metrics & Distributed Tracing Research

This document outlines the observability strategy for the Real Estate Trust & Escrow Platform. Given the multi-service orchestration (Transaction Manager, Financing Engine, Tokenization, and Ledger Services), tracing requests end-to-end is essential for reliability and compliance.

---

## 1. Observability Stack (Grafana LGTM Stack)

The platform implements the standard open-source telemetry stack:
* **Loki**: Aggregates structured JSON application logs.
* **Grafana**: Unified dashboard dashboard visualization.
* **Tempo**: High-scale distributed tracing backend.
* **Mimir**: Long-term storage for Prometheus-compatible metrics.

```
+-------------------------------------------------------------------+
|                           GRAFANA                                 |
+---------+-------------------+------------------+------------------+
          ^                   ^                  ^
          |                   |                  |
   [ Loki Logs ]     [ Tempo Traces ]    [ Mimir Metrics ]
          ^                   ^                  ^
          |                   |                  |
      FluentBit          OpenTelemetry      Prometheus
                         Collector
```

---

## 2. Distributed Tracing (OpenTelemetry in Go)

Distributed tracing tracks the exact lifecycle of an escrow transaction as it travels between HTTP endpoints, queues, and background engines.

### 2.1 Implementing OTel in Go Services
In Go, we import the OpenTelemetry SDK (`go.opentelemetry.io/otel`). Below is the standard pattern to initialize tracing and generate span segments:

```go
package tracer

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// InitTracer configures the OTLP gRPC exporter and registers the global tracer provider.
func InitTracer(serviceName, collectorAddr string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Export spans via OTLP gRPC to the OpenTelemetry Collector
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(collectorAddr),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
```

### 2.2 Trace Context Propagation over Message Queues
When the `transaction-manager` publishes a `TransactionFundedEvent` to RabbitMQ/Kafka, it must inject the trace context into the message metadata (headers). 

The receiving `LedgerService` extracts this context from the message headers to start a child span. This ensures that the tracing span is not broken by asynchronous message queues.

---

## 3. Metrics Collection (Prometheus)

We export metrics via a standard HTTP `/metrics` handler using `prometheus/client_golang/prometheus/promhttp`.

### 3.1 Core System Metrics
* **Go Runtime Telemetry**: Monitor goroutine counts, heap memory allocations, and garbage collection frequency to prevent memory leaks in Kubernetes nodes.
* **Kubernetes Node Metrics**: Track CPU and memory utilization to verify Karpenter's auto-scaling triggers.

### 3.2 Key Business and Performance Metrics (KPIs)
To monitor the economic health and health-status of the platform, we track:

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `escrow_active_volume_inr` | Gauge | `bank_partner_id`, `state` | Total monetary volume of funds currently locked in escrow. |
| `escrow_transaction_state_changes` | Counter | `state_from`, `state_to` | Frequency of transaction lifecycle state changes. |
| `escrow_virtual_account_spawns` | Counter | `bank_partner_id`, `status` | Spawns of virtual accounts to check bank API health. |
| `bank_webhook_latency_seconds` | Histogram | `provider_slug`, `status_code` | Latency profile of banking callback reconciliations. |

---

## 4. PostgreSQL Database Audits & Connection Locks

During high-concurrency operations (e.g., fractional property token bidding), multiple users write to the database simultaneously. Monitoring locks is essential to prevent deadlock conditions.

* **Deadlock & Lock Telemetry**: Collect data from PostgreSQL system catalog tables (`pg_stat_activity`, `pg_locks`) using the Prometheus **Postgres Exporter**.
* **Slow Queries Tracking**: Enable `pg_stat_statements` on PostgreSQL to log queries taking longer than 200ms. Set up Grafana alerts to notify developers if write operations block the `transaction_ledger` queue.

---

## 5. Centralized Multi-Environment Observability Hub

Yes, implementing a single, centralized **Observability Cluster** shared across `dev`, `staging`, and `production` environments is an industry-standard pattern. It is highly cost-effective and operations-friendly, provided strict logical isolation is enforced.

```
 [ Dev Cluster ]      [ Staging Cluster ]     [ Production Cluster ]
(FluentBit / OTel)    (FluentBit / OTel)      (FluentBit / OTel)
         |                     |                       |
         +---------------------+-----------------------+
                               | (Secure mTLS Ingress)
                               v
               +-------------------------------+
               | Central Observability Cluster |
               |                               |
               | - Loki / Tempo / Mimir Ingest |
               | - Grafana Dashboard Console   |
               +-------------------------------+
```

### 5.1 Telemetry Routing & Data Isolation
To prevent telemetry data from bleeding between testing environments and production, enforce metadata tagging at the agent level before data is sent across the network.

* **Mandatory Labeling**: Every metric, log line, and span forwarded to the central cluster must carry the following tags:
  * `environment` (e.g., `dev`, `staging`, `production`)
  * `cluster` (e.g., `eks-dev`, `eks-prod`)
  * `namespace` (e.g., `realestate-trust`)
* **Query Scoping**: Configure Grafana datasources and dashboards with default query variables scoped to `$environment`. This prevents developers from viewing production metrics by default.

### 5.2 Access Control & Security Boundaries
Since production transaction logs can contain sensitive financial details and KYC references, enforce strict RBAC boundaries:
* **Grafana OIDC Integration**: Integrate Grafana with your identity provider (Google Workspace, Okta, Active Directory).
* **Role-Based Permissions**:
  * **Developers**: Granted `Read/Write` access to `dev` and `staging` metrics and log streams.
  * **On-Call SREs & Auditing Officers**: Granted exclusive read access to `production` logs and telemetry streams.
* **Encryption in Transit**: All agents pushing logs/traces to the Central Observability Hub must utilize HTTPS/gRPC channels secured with mutual TLS (mTLS) or validated API tokens.

### 5.3 Cost Control & Data Retention Rules
To prevent the high volume of development logs from bloating disk storage and database indexing costs, define environment-specific retention and pruning policies:

| Telemetry Type | Dev Environment | Staging Environment | Production Environment |
| :--- | :--- | :--- | :--- |
| **Ingestion Filter** | Drop `DEBUG` logs at source. | Drop `DEBUG` logs at source. | Store all log levels (`INFO` to `FATAL`). |
| **Log Retention** | 3 Days | 7 Days | 30 Days (plus long-term archive S3 bucket). |
| **Trace Retention** | 1 Day | 3 Days | 14 Days |
| **Metric Retention** | 7 Days | 15 Days | 90 Days (with Mimir downsampling). |
