# syntax=docker/dockerfile:1

# ── Stage 1: Build ─────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency files first for better layer caching
COPY go.mod ./

# Download dependencies (none for this module, but keeps the pattern)
RUN go mod download

# Copy source
COPY . .

# Build with ldflags to inject version information.
# VERSION, BUILD_TIME and GIT_COMMIT can be overridden at docker build time
# via --build-arg:
#   docker build \
#     --build-arg VERSION=$(git describe --tags --always) \
#     --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#     --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
#     -t myapp .
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w \
        -X main.Version=${VERSION} \
        -X main.BuildTime=${BUILD_TIME} \
        -X main.GitCommit=${GIT_COMMIT}" \
      -o /app/server .

# ── Stage 2: Final image ───────────────────────────────────────────────────
FROM alpine:3.20

# Add ca-certificates so HTTPS calls (if any) work inside the container
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server .

# Expose the default port
EXPOSE 8080

# Run as non-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ENTRYPOINT ["/app/server"]
