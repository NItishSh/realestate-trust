# Architectural Proposal: Hardening RealEstate Trust with Vault Engines

This document outlines how we can leverage advanced HashiCorp Vault engines to upgrade our security posture from static secrets to a zero-trust model.

---

## 1. Dynamic Database Credentials (Database Engine)

### Current Architecture
Currently, all five microservices connect to PostgreSQL using a shared, static connection string:
`postgres://postgres:postgres@postgres:5432/..._db`
- **Security Risks**: The credentials are long-lived, shared, and have superuser (root) privileges. If one service is compromised, the entire database instance is vulnerable.

### Hardened Architecture
Using Vault's **Database Secrets Engine**, Vault dynamically creates unique, least-privilege PostgreSQL users for each service on startup.

```mermaid
sequenceDiagram
    participant App as Identity Service
    participant ESO as External Secrets Operator
    participant Vault as Vault Server
    participant DB as PostgreSQL Database

    ESO->>Vault: 1. Request Database Credentials (Role: identity-service)
    Note over Vault: Generate temporary SQL User<br/>& Password (TTL: 1h)
    Vault->>DB: 2. Create User: "v-k8s-identity-xyz" with SELECT/INSERT
    Vault-->>ESO: 3. Return Credentials
    ESO->>App: 4. Populate K8s Secret: "postgres-credentials"
    App->>DB: 5. Connect using ephemeral user
    Note over Vault: After 1h, delete SQL User
    Vault->>DB: 6. DROP USER "v-k8s-identity-xyz"
```

### Setup Steps

1. **Enable the Database Engine in Vault**:
   ```bash
   vault secrets enable database
   ```

2. **Configure Vault connection to PostgreSQL**:
   ```bash
   vault write database/config/postgres \
       plugin_name=postgresql-database-plugin \
       allowed_roles="identity-service-role" \
       connection_url="postgresql://{{username}}:{{password}}@postgres.realestate-trust.svc.cluster.local:5432/postgres?sslmode=disable" \
       username="postgres" \
       password="postgres-root-password"
   ```

3. **Define a Role with dynamic SQL creation statement**:
   Vault runs this statement in PostgreSQL to provision the temporary user:
   ```bash
   vault write database/roles/identity-service-role \
       db_name=postgres \
       creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
                            GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
       default_ttl="1h" \
       max_ttl="24h"
   ```

4. **Update the ExternalSecret manifest**:
   Instruct ESO to fetch the dynamic credentials:
   ```yaml
   apiVersion: external-secrets.io/v1
   kind: ExternalSecret
   metadata:
     name: db-credentials-sync
     namespace: realestate-trust
   spec:
     refreshInterval: "5m" # Automatically request new creds before TTL expires
     secretStoreRef:
       name: vault-backend
       kind: SecretStore
     target:
       name: postgres-credentials
     data:
       - secretKey: POSTGRES_USER
         remoteRef:
           key: database/creds/identity-service-role
           property: username
       - secretKey: POSTGRES_PASSWORD
         remoteRef:
           key: database/creds/identity-service-role
           property: password
   ```

---

## 2. Cryptography-as-a-Service (Transit Engine)

### Current Architecture
For GDPR compliance, we encrypt sensitive KYC fields using application-side AES-256-GCM.
- **Security Risks**: The encryption keys (`KYC_ENCRYPTION_KEY`) must be distributed to the Go containers. If a container is breached or an operator gains access, the key is compromised, allowing decryption of all records.

### Hardened Architecture
Using Vault's **Transit Secrets Engine**, the microservice offloads cryptographic operations to Vault. The key **never leaves Vault's secure memory**.

```mermaid
flowchart LR
    GoApp[Identity Service]
    Vault[Vault Server]
    DB[(PostgreSQL)]

    GoApp -->|1. Encrypt: 'Passport123'| Vault
    Vault -->|2. Returns: 'vault:v1:aBcd987...'| GoApp
    GoApp -->|3. Store Ciphertext| DB
```

### Implementation Steps

1. **Enable the Transit Secrets Engine**:
   ```bash
   vault secrets enable transit
   ```

2. **Create a named encryption key**:
   ```bash
   vault write -f transit/keys/kyc-encryption-key
   ```

3. **Update the Go application code**:
   Instead of encrypting locally, the application calls the Vault API over HTTP/gRPC. We can use the official `hashicorp/vault/api` Go SDK:

   ```go
   import (
       vault "github.com/hashicorp/vault/api"
   )

   func EncryptKYCField(plaintext string) (string, error) {
       config := vault.DefaultConfig()
       client, _ := vault.NewClient(config)

       data := map[string]interface{}{
           "plaintext": base64.StdEncoding.EncodeToString([]byte(plaintext)),
       }

       secret, err := client.Logical().Write("transit/encrypt/kyc-encryption-key", data)
       if err != nil {
           return "", err
       }

       return secret.Data["ciphertext"].(string), nil
   }
   ```

### Major Advantages:
- **Zero Key Leaks**: The Go microservice does not store, read, or know the cryptographic key.
- **Automated Key Rotation**: Keys can be rotated on Vault (`vault write -f transit/keys/kyc-encryption-key/rotate`) without modifying a single line of application code or configuration.
- **Cryptographic Audit Trail**: Every encryption/decryption request produces a log entry showing which service account requested the operation.
