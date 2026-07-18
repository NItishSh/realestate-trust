# User Journeys - RealEstate Trust Platform

This guide outlines step-by-step user paths to explore, test, and demonstrate the integrated microservices inside the monorepo application.

---

## Journey 1: Home Buyer Escrow & Mortgage Onboarding
This workflow validates user identity management (`identity-service`), virtual accounts creation (`transaction-manager`), financing underwrites (`financing-engine`), and cryptographic record logs (`ledger-service`).

### Step-by-Step Path:
1. **Create Identity Profile**:
   - Go to the **KYC Onboarding** tab.
   - Enter your name (e.g. `Aryan Dev`) and email (`aryan@gmail.com`). Set role to **BUYER**.
   - Click **Register Account**.
2. **Submit Compliance KYC**:
   - Under the "Submit KYC Document" form, select **PAN_CARD** and enter a document reference code (e.g., `PAN10289B`).
   - Click **Submit Verification**. Your status shifts to `PENDING`.
   - Click **Simulate Approve** to instantly verify the account and mock compliance approval.
3. **Configure Escrow Instructions**:
   - Navigate to the **Escrow Accounts** page.
   - Enter a `PROPERTY ID` (e.g., `prop-202`) and enter your Buyer/Seller IDs. Enter an amount (e.g. `₹45,00,000`).
   - Click **Initialize Escrow**. The transaction starts in the `DRAFT` phase.
4. **Fund the Escrow Account**:
   - Select your transaction from the list on the left side.
   - Click **Initialize Escrow Account** to move to the `ESCROW` stage (virtual accounts are provisioned).
   - Click **Confirm Funding Deposit** to deposit funds into the account, moving it to `FUNDED`.
5. **Release Payouts**:
   - Click **Release Payouts & Close**. The status becomes `CLOSED`, signaling successful settlement.
   - Verify the ledger log index has sealed the transaction event under the **Ledger Logs** tab.

---

## Journey 2: Fractional Investor Asset Acquisition
This workflow validates fractional property token pools (`tokenization-engine`), transaction checks, and audit trails.

### Step-by-Step Path:
1. **Select Verified Profile**:
   - Switch to or register a user that has been **KYC Approved**. (Non-approved users are blocked from acquiring property tokens).
2. **Review Asset Pool Listings**:
   - Navigate to the **Fractional Pools** tab.
   - Examine the listed property cards (e.g., `PROP-101` at `₹450/share`). Note the subscription status percentage.
3. **Acquire Share Tokens**:
   - Select an asset pool from the list.
   - Under the "Invest in Asset Shares" form, type in the number of shares to acquire (e.g. `20` shares).
   - Click **Purchase Shares for ₹9,000**.
4. **Observe Returns & Logs**:
   - Watch the holding value and estimated annualized dividends card increment automatically.
   - Check the **Ledger Logs** to inspect the cryptographically signed block tracking your token purchase.

---

## Journey 3: Cryptographic Auditor Inspection
This workflow validates ledger records integrity (`ledger-service`).

### Step-by-Step Path:
1. **Search Ledger Logs**:
   - Go to the **Ledger Logs** tab.
   - Enter search phrases (e.g., `PROP-101` or specific transaction hashes) to filter matching block records.
2. **Add Manual Audit Event**:
   - Under the "Write Audit Log" form, enter a custom compliance notice (e.g., `Physical title deed inspections verified for prop-101`).
   - Click **Publish Audit Block**.
3. **Verify SHA256 Chaining**:
   - Inspect the newly minted block (`#2`, `#3`, etc.).
   - Verify that the block displays its own **Block Hash** and correctly references the **Prev Hash** of the preceding block, verifying the immutability of the chain.
