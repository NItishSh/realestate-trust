# Kubernetes Infrastructure & Deployment Research

This document outlines infrastructure design, containerization details, node autoscaling setups, load balancing protocols, and database High Availability (HA) specifications for running the Go microservices on Kubernetes.

---

## 1. Kubernetes Architecture (EKS/GKE Cluster)

We deploy on a managed Kubernetes service (AWS EKS or Google GKE) to ensure resilience and automated master node scaling.

### 1.1 CNI (Container Network Interface)
* **Standard CNI**: Use **Cilium** or **AWS VPC CNI**. Cilium is preferred for its high-performance eBPF-based networking, low resource consumption, and secure IP-address-management.
* **Network Policies**: Implement strict namespace-level network policies. The `LedgerService` and `transaction-manager` should restrict outbound internet traffic, allowing connections only to whitelisted banking endpoints and PostgreSQL instances.

### 1.2 Karpenter Node Auto-scaler (EKS/GKE/IKS)
Instead of standard cluster auto-scalers that scale based on fixed node groups, Karpenter monitors the pod scheduler queue and directly provisions the exact VM sizes requested. This is supported natively on AWS EKS and can be integrated into GKE or IKS setups to optimize compute resources.

* **Spot vs. On-Demand Allocation**:
  * **On-Demand**: Used for PostgreSQL database statefulsets and the core `transaction-manager` to ensure zero disruption.
  * **Spot (Preemptive/Spot Nodes)**: Karpenter dynamically schedules stateless, queue-driven workers (e.g., `LedgerService`, tokenization trackers) onto Spot instances, cutting compute costs by up to 70-90%.

---

## 2. Ingress & Load Balancing (Webhook Isolation)

Banking partners require secure webhook endpoints. The platform isolates this traffic using the **Ingress-NGINX** controller.

### 2.1 Webhook IP-Whitelisting Configuration
To prevent denial-of-service (DoS) and spoofing attacks, restrict access to the bank webhooks route so only the bank's IP ranges are permitted. This is enforced using Ingress annotations:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: bank-webhook-ingress
  namespace: realestate-trust
  annotations:
    kubernetes.io/ingress.class: "nginx"
    # Whitelist specific Yes Bank and ICICI Bank subnet ranges
    nginx.ingress.kubernetes.io/whitelist-source-range: "125.16.0.0/16, 202.54.30.0/24"
    nginx.ingress.kubernetes.io/backend-protocol: "HTTP"
spec:
  rules:
  - host: api.trust.realestate.com
    http:
      paths:
      - path: /api/v1/webhooks/bank
        pathType: Prefix
        backend:
          service:
            name: transaction-manager
            port:
              number: 80
```

### 2.2 Webhook Routing Flow
```
Bank Hook Call -> Edge Load Balancer (ALB/NLB) -> Nginx Ingress (IP Check) -> transaction-manager Pod
```

---

## 3. Database Architecture (PostgreSQL HA & Pooling)

High-value financial ledger data must have high durability and transaction consistency.

### 3.1 PostgreSQL High Availability
* **Cloud Managed**: **AWS RDS Multi-AZ** or **GCP Cloud SQL HA** (replicated synchronously to a standby instance in a different availability zone).
* **Kubernetes Native**: If deploying the database inside the Kubernetes cluster, use the **CloudNativePG** operator. It automates failovers, manages pg_hba configurations, handles WAL archiving to S3, and schedules zero-downtime rolling updates.

### 3.2 PgBouncer Connection Pooling
Since Go services spawn separate connection pools (`pgxpool`), a spike in pods during high-traffic events (e.g., fractional asset launches) can overwhelm PostgreSQL.

* **PgBouncer Deployment**: Deploy PgBouncer as a Kubernetes deployment inside the same namespace.
* **Service Config**: Adjust services to communicate with PgBouncer (`port 6432`) rather than directly with PostgreSQL (`port 5432`).
* PgBouncer manages connection pooling efficiently, keeping PostgreSQL resource consumption low and preventing memory-exhaustion crashes.

---

## 4. Secret Management (External Secrets Operator)

To prevent security compliance failures, the platform implements **ESO (External Secrets Operator)**.

* **ESO Architecture**:
  1. Developers define a `SecretStore` that connects to AWS Secrets Manager / Vault using IAM Roles for Service Accounts (IRSA).
  2. Developers deploy an `ExternalSecret` manifest mapping remote secret keys to Kubernetes secrets.
  3. ESO automatically queries AWS Secrets Manager, fetches the keys, and creates a local Kubernetes Secret dynamically.
  4. The Go app container mounts this secret as an environment variable or as a volume.

---

## 5. Event-Driven Autoscaling (KEDA)

Traditional horizontal pod autoscaling (HPA) scales based on resource utilization metrics (CPU/Memory). However, for event-driven workers like `LedgerService` and `transaction-manager` that process queue messages asynchronously, resource usage does not represent real-time traffic demand.

* **KEDA Integration**: Deploy KEDA (Kubernetes Event-driven Autoscaler) to connect directly to the RabbitMQ broker.
* **Queue Metrics**: Define a `ScaledObject` that monitors the `transaction-events` queue length. If a burst of transactions occurs and the queue size increases, KEDA immediately scales out the `ledger-service` subscriber pods to drain the backlog, and scales them down to zero when the queue is idle.

---

## 6. Infrastructure as Code (Terraform)

Production environments require automated provisioning and infrastructure consistency.

* **Terraform IaC**: Declaratively define the GKE, EKS, or IKS cluster resources, VPC subnet topologies, IAM execution roles, and node provisioning policies using Terraform modules. This ensures cluster rebuilds are fast, secure, and reproducible.
