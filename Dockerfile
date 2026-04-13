# OffGridFlow Dockerfile
# Multi-stage build for production

# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Invalidate cache for source copy
ARG CACHEBUST=1
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to strip debug info and reduce size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION:-dev} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o offgridflow-api \
    ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION:-dev} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o offgridflow-worker \
    ./cmd/worker

# Stage 2: Worker runtime (named target — use with --target worker)
FROM alpine:3.18 AS worker

RUN apk add --no-cache ca-certificates tzdata curl
RUN addgroup -g 1000 offgridflow && \
    adduser -D -u 1000 -G offgridflow offgridflow

WORKDIR /app
COPY --from=builder /build/offgridflow-worker /app/offgridflow-worker
COPY --from=builder /build/infra/db /app/infra/db
RUN chown -R offgridflow:offgridflow /app
USER offgridflow
ENTRYPOINT ["/app/offgridflow-worker"]
CMD []

# Stage 3: API runtime (default — last stage, used by Railway)
FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata curl
RUN addgroup -g 1000 offgridflow && \
    adduser -D -u 1000 -G offgridflow offgridflow

WORKDIR /app

COPY --from=builder /build/offgridflow-api /app/offgridflow-api
COPY --from=builder /build/offgridflow-worker /app/offgridflow-worker
COPY --from=builder /build/railway-start.sh /app/railway-start.sh
COPY --from=builder /build/infra/db /app/infra/db

RUN chmod +x /app/railway-start.sh && \
    chown -R offgridflow:offgridflow /app

USER offgridflow

EXPOSE 8090

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:${OFFGRIDFLOW_HTTP_PORT:-8090}/health || exit 1

ENTRYPOINT ["/app/offgridflow-api"]
CMD []

