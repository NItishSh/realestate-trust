# Implementation Plan — Regulatory Compliance Gaps

This plan outlines the changes required to resolve the compliance gaps identified in the **Regulatory Compliance & Security Audit Report**:
1. **Application-Level Encryption for KYC Data** (Gap 2.1)
2. **Exposing account deletion endpoint for GDPR Right to be Forgotten** (Gap 2.2)

---

## User Review Required

> [!IMPORTANT]
> - **Encryption Key**: We will retrieve the AES-256 key from the `KYC_ENCRYPTION_KEY` environment variable. For local development and testing simplicity, a default 32-byte fallback key (`"dev-key-must-be-32-bytes-long!!!"`) will be defined. In production, this key must be set to a secure, randomly generated 32-byte value.
> - **Account Deletion Scope**: The `DELETE /api/v1/users/:id` route will purge the user from the `users` table. Because of `ON DELETE CASCADE` foreign keys, this automatically deletes all active sessions (`refresh_token_sessions`) and KYC submissions (`kyc_verifications`) associated with the user, satisfying GDPR data erasure mandates. Financial logs in `ledger-service` are preserved under their hashed signatures for financial compliance.

---

## Open Questions
*No open questions at this stage.*

---

## Proposed Changes

### Private Shared Library (`internal/`)

---

#### [NEW] [crypto.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/crypto.go)
Create a new file containing AES-256-GCM encryption/decryption utilities for sensitive string fields.
- Retrieve key from `os.Getenv("KYC_ENCRYPTION_KEY")`.
- Default to `"dev-key-must-be-32-bytes-long!!!"` if unset.

#### [MODIFY] [user_repository.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/user_repository.go)
- Add `DeleteUser(id string) error` method to the `UserRepository` interface.
- Implement `DeleteUser` in `InMemoryUserRepository`.

#### [MODIFY] [user_repository_sql.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/user_repository_sql.go)
- Implement `DeleteUser(id string) error` in `SQLUserRepository` executing:
  ```sql
  DELETE FROM users WHERE id = $1
  ```
- Update `SubmitKYC` to:
  1. Encrypt the raw document reference (`docRef`) before executing the database write.
  2. Decrypt the returned encrypted value scan so the returned `*KYCVerification` struct remains populated with the plaintext value for the API response.

#### [MODIFY] [user_handlers.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/internal/db/user_handlers.go)
- Implement the `DeleteUser(c *echo.Context) error` handler:
  1. Extract the authenticated user ID from JWT claims (`sub`).
  2. Check if the authenticated user matches the target `:id` or has the role `ADMIN`. If not, return `403 Forbidden`.
  3. Call `h.Repo.DeleteUser(userID)`.
  4. Return `204 No Content` on success.

---

### microservices (`cmd/`)

---

#### [MODIFY] [main.go](file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/cmd/identity-service/main.go)
- Register `protected.DELETE("/users/:id", handler.DeleteUser)` route.

---

## Verification Plan

### Automated Tests
- Create a unit test `TestDeleteUser` in `internal/db/handlers_test.go` asserting:
  - Deletion succeeds when the authenticated user deletes their own profile.
  - Deletion fails with `403 Forbidden` if a different user attempts it.
  - Deletion succeeds if an `ADMIN` requests it.
- Create a unit test `TestKYCEncryption` verifying that:
  - Document references are stored encrypted in the database (or verifying the encrypt/decrypt output).
  - Retrieving / submitting returns the correct plaintext value.

```bash
go test -v ./internal/db/...
```

### Manual Verification
- Deploy to the cluster, register a user, submit KYC, verify the values in the postgres DB are encrypted, delete the user, and verify all records (sessions, KYC) are purged from database tables.
