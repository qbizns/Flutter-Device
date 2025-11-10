# Device Bridge v2

> A production-grade hardware abstraction layer for POS and retail devices

[![License](https://img.shields.io/badge/license-TBD-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.22%2B-blue.svg)](https://golang.org/dl/)
[![Status](https://img.shields.io/badge/status-mvp--complete-brightgreen.svg)](IMPLEMENTATION_STATUS.md)
[![Progress](https://img.shields.io/badge/progress-100%25-brightgreen.svg)](IMPLEMENTATION_STATUS.md)

---

## Overview

Device Bridge v2 is a local or network microservice that acts as a **unified hardware abstraction layer** between client applications (POS apps, browser UIs, back-end services) and physical devices (printers, scanners, scales, customer displays, payment terminals, etc.).

**Core Philosophy:** Clients never talk to hardware directly. They communicate with Device Bridge via a stable, versioned API.

**Current Status:** ✅ Phase 2 Enhancement - Device Drivers & Discovery (see [Quick Start](QUICKSTART.md) | [Phase 2 Roadmap](PHASE_2_ROADMAP.md))

**Phase 1 MVP (100% ✅):**
- ✅ Complete infrastructure (config, logging, metrics)
- ✅ Job scheduler & queue with workers
- ✅ Event bus (pub/sub system)
- ✅ Device registry with health monitoring
- ✅ ESC/POS printer driver (TCP transport)
- ✅ Virtual printer for testing
- ✅ gRPC API server (all 14 RPC methods)
- ✅ Full type converters
- ✅ Production-ready main daemon

**Phase 2 Progress (55% ✅):**
- ✅ REST/JSON API gateway (grpc-gateway)
- ✅ Swagger UI for API documentation
- ✅ CORS support for browser clients
- ✅ WebSocket server for real-time events
- ✅ Browser test clients (scanner, payment)
- ✅ Virtual scanner (auto-scan every 15s)
- ✅ Virtual scale (auto-weight every 5s, zero/tare)
- ✅ Auto-discovery framework (TCP scanner)
- ✅ Security features (mTLS, ACL, API Auth)
- ✅ CLI administration tool (bridge-cli)
- ⏳ ZPL label printer driver
- ⏳ Print template engine
- ⏳ Virtual display and drawer devices
- ⏳ Unit and integration tests
- ⏳ CI/CD pipeline (GitHub Actions)
- ⏳ Complete documentation

🎯 **Ready to use!** See [QUICKSTART.md](QUICKSTART.md)

---

## Key Features

- **Multi-Device Support**: Printers (ESC/POS, ZPL), scanners, scales, displays, payment terminals, cash drawers
- **Multiple Transports**: USB, Serial (RS-232), TCP/IP, with automatic discovery
- **Multi-Protocol APIs**: gRPC, REST/JSON, WebSocket, Server-Sent Events
- **Real Hardware Integration**: Not stubs - production-ready drivers for real devices
- **Fault Tolerant**: Automatic reconnection, health monitoring, graceful error handling
- **Observable**: Structured logging, Prometheus metrics, OpenTelemetry tracing
- **Secure**: mTLS, ACL, PCI-DSS compliant payment handling
- **Cloud Native**: Single binary, Docker support, Kubernetes ready

---

## Supported Devices

### Phase 1 (MVP) - Real Hardware Support

| Device Type | Protocols | Transports | Status |
|-------------|-----------|------------|--------|
| ESC/POS Printers | ESC/POS | USB, Serial, TCP | Planned |
| Kitchen Printers | ESC/POS | USB, Serial, TCP | Planned |
| Label Printers | ZPL, EPL, CPCL | USB, Serial, TCP | Planned |
| Barcode Scanners | HID, Serial | USB (HID/Serial), RS-232 | Planned |
| Scales | Dibal, Mettler, CAS | Serial, USB-Serial | Planned |
| Customer Displays | Generic Serial | RS-232, USB-Serial | Planned |
| Cash Drawers | ESC/POS Pulse | Via Printer | Planned |
| Payment Terminals | Mada/KNET/Benefit | TCP/IP, Serial | Planned |
| Virtual Devices | All | Software | Planned |

### Phase 2 (Future)

- RFID/NFC Readers
- Magstripe Card Readers
- Ticket/Badge Printers
- GPIO/Signalling Devices

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Applications                   │
│  (POS Apps, Web UIs, Back-end Services, Mobile Apps)       │
└───────────┬─────────────────────────────────────────────────┘
            │
            │ gRPC / REST / WebSocket / SSE
            │
┌───────────▼─────────────────────────────────────────────────┐
│                     Device Bridge v2                         │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   API Layer │  │  Job Queue  │  │  Event Bus  │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│         │                │                 │                │
│  ┌──────▼────────────────▼─────────────────▼──────┐        │
│  │          Device Registry & Scheduler            │        │
│  └──────┬──────────────────────────────────────────┘        │
│         │                                                    │
│  ┌──────▼────────────────────────────────────────────┐      │
│  │              Device Drivers                        │      │
│  │  [Printer] [Scanner] [Scale] [Display] [Payment] │      │
│  └──────┬────────────────────────────────────────────┘      │
│         │                                                    │
└─────────┼────────────────────────────────────────────────────┘
          │
          │ USB / Serial / TCP/IP
          │
┌─────────▼────────────────────────────────────────────────────┐
│                    Physical Devices                           │
│   [Printers]  [Scanners]  [Scales]  [Displays]  [Terminals] │
└──────────────────────────────────────────────────────────────┘
```

---

## Quick Start

### Prerequisites

- Go 1.22 or later
- Physical devices (or use virtual devices for testing)
- Linux, macOS, or Windows

### Installation

**Option 1: Download Binary** (coming soon)
```bash
# Download latest release
curl -L https://github.com/Macber-eg/Flutter-Device/releases/latest/download/device-bridge-linux-amd64 -o device-bridge
chmod +x device-bridge
./device-bridge --version
```

**Option 2: Build from Source**
```bash
# Clone repository
git clone https://github.com/Macber-eg/Flutter-Device.git
cd Flutter-Device

# Build
make build

# Run
./bin/device-bridge --config configs/config.yaml
```

**Option 3: Docker**
```bash
docker run -d \
  --name device-bridge \
  --privileged \
  -v $(pwd)/configs:/config \
  -p 50051:50051 \
  -p 8080:8080 \
  macber/device-bridge:latest
```

### Configuration

Create a `config.yaml`:

```yaml
server:
  grpc_port: 50051
  http_port: 8080
  metrics_port: 9090

devices:
  - id: printer-1
    name: Receipt Printer
    kind: printer.escpos
    transport: tcp
    address: 192.168.1.100:9100

  - id: scanner-1
    name: Barcode Scanner
    kind: scanner.hid
    vendor_id: 0x05e0
    product_id: 0x1200

  - id: scale-1
    name: Weighing Scale
    kind: scale.serial
    port: /dev/ttyUSB0
    baud_rate: 9600
    protocol: mettler

logging:
  level: info
  format: json

telemetry:
  metrics: true
  tracing: false
```

### Usage

**List Devices:**
```bash
grpcurl -plaintext localhost:50051 devicebridge.v1.DeviceBridge/ListDevices
```

**Print a Receipt:**
```bash
grpcurl -plaintext -d '{
  "device_id": "printer-1",
  "document": {
    "sections": [
      {
        "lines": [
          {
            "runs": [{"text": "Welcome to My Store", "style": {"bold": true}}],
            "alignment": "CENTER"
          },
          {
            "runs": [{"text": "Total: $99.99"}],
            "alignment": "RIGHT"
          }
        ]
      }
    ]
  }
}' localhost:50051 devicebridge.v1.DeviceBridge/Print
```

**Subscribe to Scanner Events:**
```bash
grpcurl -plaintext -d '{"device_id": "scanner-1"}' \
  localhost:50051 devicebridge.v1.DeviceBridge/SubscribeScanner
```

**Read Weight from Scale:**
```bash
grpcurl -plaintext -d '{"device_id": "scale-1"}' \
  localhost:50051 devicebridge.v1.DeviceBridge/GetWeight
```

---

## API Documentation

Device Bridge v2 exposes multiple API interfaces:

### gRPC API (Primary)

Protocol Buffers definitions: [`proto/devicebridge/v1/`](proto/devicebridge/v1/)

**Service:** `devicebridge.v1.DeviceBridge`

**Methods:**
- `ListDevices` - List all registered devices
- `GetDevice` - Get device details
- `Print` - Print a document
- `SubscribeScanner` - Stream barcode scan events
- `GetWeight` - Read weight from scale
- `ShowDisplay` - Show text on customer display
- `OpenDrawer` - Open cash drawer
- `StartPayment` - Initiate payment transaction
- `GetPaymentStatus` - Query payment status

### REST/JSON API

REST gateway available at: `http://localhost:8080/v1/`

**Examples:**
```bash
# List devices
curl http://localhost:8080/v1/devices

# Get device
curl http://localhost:8080/v1/devices/printer-1

# Print receipt
curl -X POST http://localhost:8080/v1/devices/printer-1/print \
  -H "Content-Type: application/json" \
  -d @receipt.json
```

### WebSocket API

Real-time event streaming: `ws://localhost:8080/ws/scanner/scanner-1`

**JavaScript Example:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/scanner/scanner-1');

ws.onmessage = (event) => {
  const scan = JSON.parse(event.data);
  console.log('Scanned:', scan.data, scan.symbology);
};
```

---

## Development

### Project Structure

```
device-bridge-v2/
├── cmd/                 # Entry points
│   ├── bridge/         # Main daemon
│   └── bridge-cli/     # Admin CLI
├── internal/           # Internal packages
│   ├── api/           # gRPC + REST + WebSocket handlers
│   ├── app/           # Application orchestration
│   ├── devices/       # Device interfaces
│   ├── drivers/       # Device driver implementations
│   ├── discovery/     # Device discovery
│   ├── jobs/          # Job scheduling
│   ├── events/        # Event pub/sub
│   ├── config/        # Configuration
│   ├── security/      # Security (mTLS, ACL)
│   └── telemetry/     # Logging, metrics, tracing
├── proto/             # Protocol Buffer definitions
├── test/              # Tests
│   ├── integration/   # Integration tests
│   └── virtual_devices/ # Virtual devices for testing
├── docs/              # Documentation
├── configs/           # Configuration examples
└── scripts/           # Build and deployment scripts
```

### Building

```bash
# Install dependencies
make deps

# Generate protobuf code
make proto

# Build binary
make build

# Run tests
make test

# Run integration tests (requires devices)
make test-integration

# Build Docker image
make docker-build
```

### Testing

**Unit Tests:**
```bash
go test ./internal/...
```

**Integration Tests (with virtual devices):**
```bash
go test ./test/integration/...
```

**Integration Tests (with real hardware):**
```bash
# Configure test devices in test/integration/config.yaml
REAL_HARDWARE=1 go test ./test/integration/...
```

### Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Development Workflow:**
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make lint` and `make test`
6. Submit a pull request

---

## Documentation

- **[Device Bridge v2 Specification](docs/DEVICE_BRIDGE_V2_SPECIFICATION.md)** - Complete technical specification
- **[API Reference](docs/API.md)** - Detailed API documentation (coming soon)
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment guide (coming soon)
- **[Driver Development](docs/DRIVER_DEVELOPMENT.md)** - How to add new device drivers (coming soon)
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions (coming soon)

---

## Observability

### Logging

Structured JSON logging with configurable levels:

```json
{
  "timestamp": "2025-11-10T12:34:56.789Z",
  "level": "INFO",
  "message": "Print job completed",
  "device_id": "printer-1",
  "job_id": "job_123",
  "duration_ms": 234
}
```

### Metrics

Prometheus metrics available at `http://localhost:9090/metrics`:

- `device_status{device_id, kind}` - Device status
- `jobs_total{device_id, type, status}` - Job counts
- `job_duration_seconds{device_id, type}` - Job duration histogram
- `payments_total{provider, status}` - Payment counts

**Grafana Dashboard:** (coming soon)

### Tracing

OpenTelemetry tracing support (optional):

```yaml
telemetry:
  tracing:
    enabled: true
    exporter: jaeger
    endpoint: http://jaeger:14268/api/traces
```

---

## Deployment

### Systemd Service (Linux)

```bash
sudo make install

# Start service
sudo systemctl start device-bridge

# Enable on boot
sudo systemctl enable device-bridge

# Check status
sudo systemctl status device-bridge
```

### Docker Compose

```yaml
version: '3.8'
services:
  device-bridge:
    image: macber/device-bridge:latest
    privileged: true
    volumes:
      - ./configs:/config
      - /dev:/dev
    ports:
      - "50051:50051"
      - "8080:8080"
      - "9090:9090"
    environment:
      - CONFIG_FILE=/config/config.yaml
```

### Kubernetes

```bash
# Install with Helm (coming soon)
helm repo add device-bridge https://macber-eg.github.io/device-bridge-helm
helm install device-bridge device-bridge/device-bridge
```

---

## Security

### Development Mode (Default)
- No TLS
- No authentication
- Listens on localhost only
- Verbose logging

### Production Mode

Enable in config:

```yaml
security:
  mode: production
  tls:
    enabled: true
    cert_file: /certs/server.crt
    key_file: /certs/server.key
    ca_file: /certs/ca.crt
  acl:
    enabled: true
    rules:
      - client_id: pos_app_1
        allowed_devices: ["printer-*", "drawer-*"]
        allowed_operations: ["print", "open_drawer"]
```

**Features:**
- mTLS (mutual TLS)
- Access Control Lists (ACL)
- PCI-DSS compliant payment handling
- No sensitive data logging

---

## Performance

**Target Performance:**
- 20+ print jobs/minute per lane
- 10+ concurrent payment flows
- 20+ scanner events/second per device
- < 200ms print job latency
- < 1s scale reading latency
- < 100MB memory (idle)
- < 5% CPU (idle)

---

## Roadmap

### Phase 1: MVP (Current)
- [x] Project structure and specification
- [ ] Core infrastructure (config, logging, metrics)
- [ ] gRPC API skeleton
- [ ] ESC/POS printer support
- [ ] Device registry
- [ ] Job scheduler
- [ ] Scanner support
- [ ] Scale support
- [ ] Display and drawer support
- [ ] Payment terminal support

### Phase 2: Enhancement
- [ ] REST/JSON gateway
- [ ] WebSocket support
- [ ] Label printer support (ZPL)
- [ ] Auto-discovery
- [ ] Arabic text support
- [ ] Print templates
- [ ] Security (mTLS, ACL)

### Phase 3: Production
- [ ] Virtual devices
- [ ] Integration tests
- [ ] CI/CD pipeline
- [ ] Docker images
- [ ] Helm charts
- [ ] Documentation
- [ ] v2.0.0 release

### Phase 4: Advanced
- [ ] RFID/NFC readers
- [ ] Magstripe readers
- [ ] GPIO devices
- [ ] Cloud integration
- [ ] Multi-node support
- [ ] Device health analytics

---

## License

TBD

---

## Support

- **Issues:** [GitHub Issues](https://github.com/Macber-eg/Flutter-Device/issues)
- **Discussions:** [GitHub Discussions](https://github.com/Macber-eg/Flutter-Device/discussions)
- **Email:** support@example.com

---

## Acknowledgments

Built with:
- [gRPC](https://grpc.io/)
- [Protocol Buffers](https://protobuf.dev/)
- [Go](https://golang.org/)
- [Prometheus](https://prometheus.io/)
- [OpenTelemetry](https://opentelemetry.io/)

---

## Related Projects

- [Flutter-Device v1](https://github.com/Macber-eg/Flutter-Device/tree/v1) - Previous version
- [ESC/POS Go Library](https://github.com/hennedo/escpos)
- [ZPL Utilities](https://github.com/SimulPiscator/ZPLPrinter)

---

**Device Bridge v2** - Bringing hardware and software together, reliably.
