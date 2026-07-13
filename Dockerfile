# --- BUILD STAGE ---
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Compile all 6 microservices statically
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/transaction-manager cmd/transaction-manager/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/identity-service cmd/identity-service/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/financing-engine cmd/financing-engine/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/tokenization-engine cmd/tokenization-engine/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/ledger-service cmd/ledger-service/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/property-registry-service cmd/property-registry-service/main.go

# --- BASE RUNTIME STAGE ---
FROM alpine:latest AS base-runtime
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1001
WORKDIR /home/appuser
USER appuser

# --- TARGET: Transaction Manager ---
FROM base-runtime AS transaction-manager
COPY --from=builder --chown=appuser:appgroup /bin/transaction-manager .
EXPOSE 8080
CMD ["./transaction-manager"]

# --- TARGET: Identity Service ---
FROM base-runtime AS identity-service
COPY --from=builder --chown=appuser:appgroup /bin/identity-service .
EXPOSE 8081
CMD ["./identity-service"]

# --- TARGET: Financing Engine ---
FROM base-runtime AS financing-engine
COPY --from=builder --chown=appuser:appgroup /bin/financing-engine .
EXPOSE 8082
CMD ["./financing-engine"]

# --- TARGET: Tokenization Engine ---
FROM base-runtime AS tokenization-engine
COPY --from=builder --chown=appuser:appgroup /bin/tokenization-engine .
EXPOSE 8083
CMD ["./tokenization-engine"]

# --- TARGET: Ledger Service ---
FROM base-runtime AS ledger-service
COPY --from=builder --chown=appuser:appgroup /bin/ledger-service .
EXPOSE 8084
CMD ["./ledger-service"]

# --- TARGET: Property Registry Service ---
FROM base-runtime AS property-registry-service
COPY --from=builder --chown=appuser:appgroup /bin/property-registry-service .
EXPOSE 8085
CMD ["./property-registry-service"]
