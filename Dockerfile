# OffGridFlow API Dockerfile
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

# Copy source code
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

# Stage 2: Runtime
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 offgridflow && \
    adduser -D -u 1000 -G offgridflow offgridflow

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/offgridflow-api /app/offgridflow-api
COPY --from=builder /build/offgridflow-worker /app/offgridflow-worker

# Copy migrations and configs
COPY --from=builder /build/infra/db /app/infra/db

# Set ownership
RUN chown -R offgridflow:offgridflow /app

# Switch to non-root user
USER offgridflow

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["/app/offgridflow-api"]
CMD []

# Stage 3: Worker runtime
FROM alpine:3.18 AS worker

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 offgridflow && \
    adduser -D -u 1000 -G offgridflow offgridflow

# Set working directory
WORKDIR /app

# Copy worker binary from builder
COPY --from=builder /build/offgridflow-worker /app/offgridflow-worker

# Copy migrations and configs
COPY --from=builder /build/infra/db /app/infra/db

# Set ownership
RUN chown -R offgridflow:offgridflow /app

# Switch to non-root user
USER offgridflow

# Run the worker
ENTRYPOINT ["/app/offgridflow-worker"]
CMD []
