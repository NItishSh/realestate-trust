# Microservices Architecture Catalog

This document details the microservice layout, functional scopes, and tech stack configurations for the **Real Estate Trust & Escrow Platform**. All services are written in **Go (Golang)** and run containerized within a Kubernetes cluster.

---

## 1. Transaction & Escrow Manager (`transaction-manager`)

* **Scope**: Core payment engine and transaction state machine coordinator.
* **Backend Stack**: Go, gRPC (internal service-to-service APIs), REST (client-facing endpoints and webhook receivers), PostgreSQL database.
* **Database Tables Managed**: `transactions`, `escrow_accounts`, `properties` (metadata).

### Key Responsibilities:
* Spawning unique, transaction-bound **Virtual Accounts** via bank APIs for deposit tracking.
* Registering escrow agreement terms, milestones, and required digital signees (Buyer, Seller, Verifier).
* Checking reconciled deposit webhooks and transitioning transaction status to `FUNDED`.
* Coordinating and validating multi-party digital approvals (Buyer keys, Seller keys, and external Verifier).
* Executing payouts from the Escrow Virtual Account to the seller's bank details upon deed registration confirmation.
* Abstracting bank connectivity using a custom **Banking Adapter Factory** (`IBankingAdapter` interface).

---

## 2. Embedded Financing Engine (`financing-engine`)

* **Scope**: Mortgage loan orchestration and financial lender/NBFC portal integrations.
* **Backend Stack**: Go, gRPC (for transaction queries), REST (for lender webhook receivers), PostgreSQL database.
* **Database Tables Managed**: `financing_applications`, `lenders`, `loan_schedules`.

### Key Responsibilities:
* Collecting and forwarding buyer credit applications to NBFC partner APIs.
* Tracking loan application statuses (Applied, Approved, Disbursed, Rejected).
* Recording approved loan conditions (applied amounts, approved principal, interest rates).
* Validating NBFC disbursement callbacks and routing the disbursed funds directly into the corresponding transaction's virtual escrow account.

---

## 3. Fractional Tokenization Engine (`tokenization-engine`)

* **Scope**: Micro-investment tracking and asset equity fractionalization.
* **Backend Stack**: Go, gRPC (for internal ledger tracking), PostgreSQL database.
* **Database Tables Managed**: `fractional_pools`, `fractional_holdings`.

### Key Responsibilities:
* Creating property tokenization pools based on valuation, dividing asset equity into fractional ownership shares.
* Allocating share holdings, managing real-time share pool availability, and updating capitalization tables.
* Triggering rental yield distributions or yield payout requests.
* Interacting with the `transaction-manager` to lock pool funds in escrow until the property pool is 100% funded.

---

## 4. KYC & Identity Service (`identity-service`)

* **Scope**: User authentication, role-based access control (RBAC), and digital identity compliance.
* **Backend Stack**: Go, gRPC (used as middleware authorization by all services), PostgreSQL database.
* **Database Tables Managed**: `users`, `user_consent_logs`, `kyc_audit_trail`.

### Key Responsibilities:
* Managing user profiles, bank accounts, and authorization roles (`BUYER`, `SELLER`, `INVESTOR`, `LENDER`, `VERIFIER`, `ADMIN`).
* Integrating with third-party KYC platforms (Signzy, Karza) to validate user credentials (PAN cards, Aadhaar OTPs, GSTIN registries).
* Capturing, timestamping, and archiving **user consent signatures** before conducting regulatory checks to comply with the Indian DPDP (Digital Personal Data Protection) Act.
* Masking sensitive personally identifiable information (PII) before storage.

---

## 5. Ledger & Auditing Service (`ledger-service`)

* **Scope**: Trust layer, immutable system audit log.
* **Backend Stack**: Go (message queue consumers), PostgreSQL database (isolated, read-only/append-only tables).
* **Database Tables Managed**: `transaction_ledger` (Immutable append-only).

### Key Responsibilities:
* Subscribing to transaction state events published to RabbitMQ or Kafka.
* Appending ledger entries capturing details of the state transition, actor ID, and transaction metadata.
* Generating a cryptographic SHA256 hash for each row, chaining it to the hash of the preceding ledger entry (hash chaining).
* Running automated cron schedules to verify ledger chain integrity, reporting database tampering if a hash mismatch is detected.

---

## 6. Property Registry Service (`property-registry-service`)

* **Scope**: Land title verification, municipal deed records, and cadastral boundary validation.
* **Backend Stack**: Go, Echo v5 REST, PostgreSQL database.
* **Database Tables Managed**: `properties`, `property_title_verifications`.
* **Port**: `:8085`

### Key Responsibilities:
* Validating land parcel identifiers, survey numbers, and municipal boundaries against state registrar schemas.
* Tracking legal title verification status (`PENDING`, `VERIFIED`, `DISPUTED`).
* Enforcing role gates so only authenticated `SELLER`, `BROKER`, or `ADMIN` users can register properties.

---

## 7. Feedback & Reputation Service (`feedback-service`)

* **Scope**: Post-transaction satisfaction, counterparty rating boundaries, and platform reviews.
* **Backend Stack**: Go, Echo v5 REST, PostgreSQL database.
* **Database Tables Managed**: `feedback`.
* **Port**: `:8086`

### Key Responsibilities:
* Collecting stakeholder reviews after deal settlement (rating scale 1–5).
* Validating rating boundaries (1 to 5 integer) and sanitizing review comments against XSS injection.
* Providing aggregated seller and broker trust scores across the platform.
