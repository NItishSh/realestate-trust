# Infrastructure and Deployment Strategy Research

This document captures the analysis and strategic decisions regarding the infrastructure, deployment model, and observability stack for the Real Estate Trust & Escrow Platform.

## 1. Observability Stack: Shared vs. Dedicated Cluster
**Analysis:** Should the observability stack be implemented on the same Kubernetes cluster as the application, or a different one? Can we have a self-hosted Monitoring Cluster common across all environments?

**Recommendation:** A **Dedicated, Self-Hosted Monitoring Cluster** is highly recommended.
* **Why not the same cluster?** If the application cluster experiences a catastrophic failure (e.g., OOM, CPU exhaustion, network partition), the observability stack running on the same cluster will likely fail as well, leaving you blind precisely when you need logs and metrics the most.
* **Common Cluster Across Environments:** Having a single, centralized Monitoring Cluster (running Prometheus, Grafana, Loki/Elasticsearch, Jaeger) that scrapes metrics from Dev, Staging, and Prod is a very common and cost-effective pattern.
* **Implementation Details:** You achieve separation by tagging all incoming metrics and logs with `environment` labels (e.g., `env=prod`, `env=staging`). You must ensure secure, cross-cluster network communication (e.g., via VPC peering or secure ingress) so the centralized observability cluster can pull metrics/logs from the application clusters safely.

## 2. Microservices Architecture & Containerization
**Analysis:** How many microservices are planned? Should each be a different Docker container and Deployment in Kubernetes?

**Recommendation:** **Yes**, each microservice should be built as a distinct Docker image and deployed as an independent `Deployment` entity in Kubernetes. This enables independent scaling, isolated fault domains, and decoupled deployment lifecycles.

Here are the planned core microservices for this platform:

1. **API Gateway / BFF (Backend-for-Frontend)**
   * **Description:** The single entry point for all client requests. It handles rate limiting, basic JWT validation, request routing, and payload aggregation.

2. **Transaction & Escrow Manager**
   * **Description:** The core financial engine. Manages state transitions for deals (Draft -> Escrow -> Funded -> Closed), interacts with external banking APIs for Virtual Accounts (VAs), and orchestrates money movement.

3. **Embedded Financing Engine**
   * **Description:** Handles loan applications, credit scoring integrations, and interacts with NBFCs/Banks to approve and track mortgages or financing for properties.

4. **Tokenization Engine (Optional/Future)**
   * **Description:** Manages the fractional ownership logic. Issues digital shares for real estate assets, tracks fractional cap tables, and handles secondary market trading logic.

5. **User & Identity Service**
   * **Description:** Manages user profiles, KYC/AML verification statuses, role-based access control (Buyer, Seller, Title Officer), and authentication.

6. **Property & Listing Service**
   * **Description:** Manages the catalog of real estate properties, media (images/videos), legal documents, and property metadata.

7. **Notification Service**
   * **Description:** Asynchronous service handling emails, SMS, and push notifications for escrow milestones, funding approvals, and system alerts.

## 3. Infrastructure as Code (IaC) and UI Code Strategy
**Analysis:** How should the infrastructure code and UI code be handled and structured?

**Recommendation:**
* **Infrastructure Code:** The infrastructure should be strictly declarative using tools like **Terraform** or **OpenTofu**. This code should live in its own dedicated repository (or a clearly separated `infra/` root directory in a monorepo). It will provision the VPCs, Kubernetes clusters (EKS/GKE/AKS), managed PostgreSQL databases, and IAM roles. Deployment to Kubernetes should be managed by a GitOps tool like **ArgoCD** or **Flux**, which constantly monitors your manifests and synchronizes the cluster state.
* **UI Code:** The frontend should be separated into its own repository (or monorepo workspace). Depending on the platform:
   * **Web App:** Built with Next.js or React, deployed via Vercel, Netlify, or as containerized instances within the Kubernetes cluster.
   * **Mobile App:** Built with Flutter or React Native to support both iOS and Android natively, interfacing with the same API Gateway.
