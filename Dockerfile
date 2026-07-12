# --- BUILD STAGE ---
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .

# Compile all 5 microservices statically
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/transaction-manager cmd/transaction-manager/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/identity-service cmd/identity-service/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/financing-engine cmd/financing-engine/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/tokenization-engine cmd/tokenization-engine/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/ledger-service cmd/ledger-service/main.go

# --- TARGET: Transaction Manager ---
FROM alpine:latest AS transaction-manager
WORKDIR /root/
COPY --from=builder /bin/transaction-manager .
EXPOSE 8080
CMD ["./transaction-manager"]

# --- TARGET: Identity Service ---
FROM alpine:latest AS identity-service
WORKDIR /root/
COPY --from=builder /bin/identity-service .
EXPOSE 8081
CMD ["./identity-service"]

# --- TARGET: Financing Engine ---
FROM alpine:latest AS financing-engine
WORKDIR /root/
COPY --from=builder /bin/financing-engine .
EXPOSE 8082
CMD ["./financing-engine"]

# --- TARGET: Tokenization Engine ---
FROM alpine:latest AS tokenization-engine
WORKDIR /root/
COPY --from=builder /bin/tokenization-engine .
EXPOSE 8083
CMD ["./tokenization-engine"]

# --- TARGET: Ledger Service ---
FROM alpine:latest AS ledger-service
WORKDIR /root/
COPY --from=builder /bin/ledger-service .
EXPOSE 8084
CMD ["./ledger-service"]
