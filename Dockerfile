# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bridge ./cmd/bridge
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bridge-cli ./cmd/bridge-cli

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 bridge && \
    adduser -D -u 1000 -G bridge bridge

# Copy binaries from builder
COPY --from=builder /bridge /usr/local/bin/bridge
COPY --from=builder /bridge-cli /usr/local/bin/bridge-cli

# Copy example configuration
COPY configs/config.example.yaml /etc/device-bridge/config.yaml

# Create data directory
RUN mkdir -p /var/lib/device-bridge && \
    chown -R bridge:bridge /var/lib/device-bridge /etc/device-bridge

# Switch to non-root user
USER bridge

# Expose ports
EXPOSE 50051 8080 9090

# Set environment variables
ENV CONFIG_FILE=/etc/device-bridge/config.yaml
ENV DATA_DIR=/var/lib/device-bridge

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD bridge-cli --server localhost:50051 ping || exit 1

# Run the bridge
ENTRYPOINT ["bridge"]
CMD ["-config", "/etc/device-bridge/config.yaml"]
