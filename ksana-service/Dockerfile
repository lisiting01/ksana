# Multi-stage build for ksana-service (Go API)

# 1) Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install build tools (just in case CGo or certificates are needed)
RUN apk add --no-cache git ca-certificates

# Copy go module files first for better caching
COPY ksana-service/go.mod ./

# Download dependencies (none declared beyond stdlib, but keep step for cache)
RUN go mod download

# Copy source
COPY ksana-service/ .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ksana-service .

# 2) Runtime stage
FROM alpine:3.20
WORKDIR /app

# Add CA certificates for outbound HTTPS requests
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -H -u 10001 appuser

# Create data and config dirs (can be mounted as volumes)
RUN mkdir -p /app/data /app/config && chown -R appuser:appuser /app

COPY --from=builder /out/ksana-service /app/ksana-service

USER appuser

# Default environment (can be overridden)
ENV PORT=7100 \
    DATA_DIR=/app/data \
    WORKERS=4 \
    DEFAULT_TIMEOUT=10s \
    MAX_RETRIES=3 \
    RETRY_BACKOFF=5s \
    LOG_LEVEL=info \
    AUTH_KEYS_FILE=/app/config/api_keys.txt

EXPOSE 7100

ENTRYPOINT ["/app/ksana-service"]
