# Device Bridge v2 - Production Dockerfile
# Multi-stage build for optimized image size

# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev libusb-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o device-bridge \
    ./cmd/device-bridge

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    libusb \
    curl

# Create non-root user
RUN addgroup -S devicebridge && \
    adduser -S -G devicebridge -h /var/lib/device-bridge devicebridge

# Create directories
RUN mkdir -p /var/lib/device-bridge /var/log/device-bridge /etc/device-bridge && \
    chown -R devicebridge:devicebridge /var/lib/device-bridge /var/log/device-bridge /etc/device-bridge

# Copy binary from builder
COPY --from=builder /build/device-bridge /usr/local/bin/device-bridge
RUN chmod 755 /usr/local/bin/device-bridge

# Copy default configuration
COPY configs/payment/config.production.yaml /etc/device-bridge/config.yaml.example

# Switch to non-root user
USER devicebridge

# Set working directory
WORKDIR /var/lib/device-bridge

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/device-bridge"]

# Default arguments
CMD ["--config=/etc/device-bridge/config.yaml", "--log-level=info"]
