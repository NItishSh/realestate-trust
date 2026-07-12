Viewed architecture_design.md:92-141

Before proceeding to the implementation phase, there are several key operational, integration, and engineering details that need to be finalized. These details represent the boundary lines between our core application logic and external systems:

---

### 1. Integration Specifications & API Sandboxes
* **Banking APIs**: We need the actual sandbox documentation and credentials for the target banks (e.g., ICICI Corporate Banking, Yes Bank API Banking, HDFC, or aggregator APIs like RazorpayX/Cashfree). We specifically need:
  * The payload formats for creating Virtual Accounts.
  * The exact hashing algorithms used by the bank to sign webhooks (e.g., RSA-SHA256, HMAC-SHA256).
  * Outbound payout API specs (IMPS, NEFT, RTGS) and transaction status check query contracts.
* **NBFC & Lender APIs**: We need the API documentation for our embedded financing partners to understand:
  * Credit profile submission formats (PAN/CIBIL query inputs).
  * Direct disbursement tracking webhooks.

---

### 2. Legal Document & Land Registry (SRO) Verification Model
* **Verification Method**: How will the platform confirm that a deed registration has been successfully executed?
  * **Automated Integration**: Do we have access to government land records/Sub-Registrar Office (SRO) APIs (e.g., IGR services)?
  * **Manual Auditing**: Or do we build a **Verifier Portal** where internal legal officers review uploaded registration deeds, cross-reference them manually, and sign off with their platform credentials? *(Recommended for the first phase, as state land record APIs vary drastically in availability and uptime)*.

---

### 3. Identity Verification (KYC & AML)
* **KYC Partner**: Real estate transactions require strict Anti-Money Laundering (AML) and Know Your Customer (KYC) checks on buyers, sellers, and fractional investors.
  * We need details of the KYC verification provider to integrate (e.g., Signzy, Karza, HyperVerge, Digilocker APIs).
  * We need to define KYC validation rules (e.g., PAN verification, Aadhaar OTP verification, corporate GSTIN verification).

---

### 4. Fractional Ledger Tech Stack
* **Ledger Mechanics**: How should the ownership shares for fractional properties be logged and transferred?
  * **Database-backed Ledger**: PostgreSQL double-entry bookkeeping tables (highly secure, simple, fast).
  * **Blockchain/DLT Ledger**: Smart contracts (e.g., ERC-3643 or ERC-20 on an EVM-compatible chain like Polygon or Hyperledger) to programmatically represent fraction shares. If choosing this, we will need to decide on wallet management (custodial vs. non-custodial via integrations like Fireblocks or Web3Auth).

---

### 5. Technical Architecture & Tech Stack Preferences
* **Backend Language**: Do you want to build the backend in **TypeScript (Node.js/NestJS)**, **Python (FastAPI)**, or **Go**?
* **Deployment & Compliance Hosting**: Because of financial transaction data storage guidelines (e.g., RBI localization rules in India), we need to define the cloud hosting setup (e.g., AWS Mumbai region) and data security/encryption models for databases containing bank accounts and user identities.