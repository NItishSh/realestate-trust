# --- BASE BUILDER STAGE ---
FROM golang:1.26.5-alpine AS builder-base
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .

# --- SPECIFIC BUILDER STAGES ---
FROM builder-base AS build-transaction-manager
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/app cmd/transaction-manager/main.go

FROM builder-base AS build-identity-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/app cmd/identity-service/main.go

FROM builder-base AS build-financing-engine
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/app cmd/financing-engine/main.go

FROM builder-base AS build-tokenization-engine
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/app cmd/tokenization-engine/main.go

FROM builder-base AS build-ledger-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/app cmd/ledger-service/main.go

FROM builder-base AS build-property-registry-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/app cmd/property-registry-service/main.go


# --- RUNTIME STAGES (DISTROLESS) ---
# Using distroless/static-debian11:nonroot which runs as user 'nonroot' (uid 65532)

# --- TARGET: Transaction Manager ---
FROM gcr.io/distroless/static-debian11:nonroot AS transaction-manager
COPY --from=build-transaction-manager /bin/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]

# --- TARGET: Identity Service ---
FROM gcr.io/distroless/static-debian11:nonroot AS identity-service
COPY --from=build-identity-service /bin/app /app
EXPOSE 8081
ENTRYPOINT ["/app"]

# --- TARGET: Financing Engine ---
FROM gcr.io/distroless/static-debian11:nonroot AS financing-engine
COPY --from=build-financing-engine /bin/app /app
EXPOSE 8082
ENTRYPOINT ["/app"]

# --- TARGET: Tokenization Engine ---
FROM gcr.io/distroless/static-debian11:nonroot AS tokenization-engine
COPY --from=build-tokenization-engine /bin/app /app
EXPOSE 8083
ENTRYPOINT ["/app"]

# --- TARGET: Ledger Service ---
FROM gcr.io/distroless/static-debian11:nonroot AS ledger-service
COPY --from=build-ledger-service /bin/app /app
EXPOSE 8084
ENTRYPOINT ["/app"]

# --- TARGET: Property Registry Service ---
FROM gcr.io/distroless/static-debian11:nonroot AS property-registry-service
COPY --from=build-property-registry-service /bin/app /app
EXPOSE 8085
ENTRYPOINT ["/app"]
