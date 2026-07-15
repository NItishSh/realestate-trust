# Best Practices - Real Estate Trust & Escrow Platform

This document outlines the recommended architectural, security, operational, and cost-optimization practices for running and maintaining the Real Estate Trust & Escrow Platform.

Given the high transaction volumes and values handled by the system, these principles ensure a secure, resilient, and cost-effective deployment.

---

## 1. Manageability & Observability (Maintainability)

### 1.1 Structured JSON Logging
* **Standard Logger**: Use Go's built-in structured logger (`slog` or a high-performance library like `zap`) to output logs in JSON format.
* **Correlated Fields**: Ensure every log event contains tracing keys:
  * `transaction_id`: Links log lines to a specific escrow transaction.
  * `escrow_account_id`: Identifies the related bank virtual account.
  * `bank_partner_id`: Indicates which bank partner was used for the execution.
  * `correlation_id`: Set at the API Gateway level to trace individual requests.

### 1.2 Distributed Tracing
* **OpenTelemetry (OTel)**: Implement OpenTelemetry trace injection and extraction.
* **Trace Propagation**: Propagate the context from the client request through the API Gateway, inside asynchronous event queues (RabbitMQ/Kafka), down to the final database transaction ledger updates.

### 1.3 Database Schema Control
* **Declarative Migrations**: Never manually run DDL statements on production databases. Use a schema migration engine like `dbmate` or `golang-migrate`.
* **Deployment Hooks**: Run migrations automatically using Kubernetes **Init Containers** or deployment pipeline hooks so that DB schemas are upgraded before services start serving requests.

### 1.4 Centralized Contracts
* **gRPC Protocol Buffers**: Use gRPC for all internal microservice calls. It guarantees contract synchronization at compile-time and minimizes network overhead inside the cluster.
* **OpenAPI Specs**: Generate and publish OpenAPI specs for public-facing webhooks and frontend clients.

---

## 2. Platform Security & Trust (Compliance)

### 2.1 Idempotency Architecture
* **Strict Double-Spend Protection**: Mandate an `Idempotency-Key` (UUIDv4) header for all state-changing endpoints (especially payouts).
* **Storage and TTL**: Save idempotency request payloads and status responses in PostgreSQL or a fast key-value store (Redis) with a TTL of 24–48 hours to automatically handle client retries safely.

### 2.2 Vault Separation
* **Secrets Management**: Never commit bank certificates, API keys, or database credentials to git repositories or store them in raw ConfigMaps.
* **External Secrets Operator (ESO)**: Deploy ESO on the Kubernetes cluster. ESO automatically syncs secrets from secure providers (e.g., HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager) straight into native Kubernetes secrets, mounting them as environment variables or transient files.

### 2.3 Webhook Validation
* **HMAC verification**: Validate all incoming deposit notifications from partner banks by computing the SHA256 HMAC of the request body using the configured bank secret, verifying it against the signature headers.
* **mTLS (Mutual TLS)**: Configure the Kubernetes Ingress controller to terminate and validate client certificates for incoming bank communication channels.

### 2.4 Cryptographic Ledger Audits
* **Audit Chain Integrity**: Ensure the transaction ledger continues hash-chaining state updates.
* **Scheduled Validation**: Run a daily Kubernetes CronJob to recalculate the hashes in `transaction_ledger` and trigger alert webhooks if any record has been modified out of order.

---

## 3. Cost-Effectiveness & Resource Optimization (FinOps)

### 3.1 Connection Optimization
* **Client Pooling**: Configure database connections on the Go client using `pgxpool` with appropriate max connection limits.
* **PgBouncer**: Deploy PgBouncer as a middleware proxy between Kubernetes pods and the database cluster to mitigate the memory overhead of maintaining direct PostgreSQL connections.

### 3.2 Kubernetes Auto-Scaling
* **Karpenter Node Provisioner**: Use Karpenter for AWS/GCP nodes rather than standard auto-scalers. Karpenter dynamically matches incoming pod resource specs to launch the most cost-efficient compute configurations.
* **KEDA (Kubernetes Event-driven Autoscaling)**: Scale ledger and message queue consumer microservices based on actual queue depth (RabbitMQ message count/Kafka log lag) rather than CPU/RAM parameters. This allows pods to scale down to zero when there are no transactions.

### 3.3 Container Footprint Optimization
* **Distroless & Scratch Images**: Use multi-stage Docker builds. Build statically compiled Go binaries (`CGO_ENABLED=0`) and execute them inside `scratch` or `distroless` runner containers.
* **Efficiency Gains**: Reduces images to under **15 MB**, cutting down container storage costs, decreasing pod launch latency, and reducing the container's security attack surface.
