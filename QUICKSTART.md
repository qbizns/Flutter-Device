# Device Bridge v2 - Quick Start Guide

This guide will help you build and run Device Bridge v2 MVP.

---

## Prerequisites

- **Go 1.22+**: `go version`
- **Protocol Buffer Compiler**: Required for code generation
- **Make**: For building (optional, can use `go build` directly)

---

## Step 1: Install Protocol Buffer Compiler

### macOS
```bash
brew install protobuf
```

### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install -y protobuf-compiler
```

### Windows
Download from: https://github.com/protocolbuffers/protobuf/releases

### Verify Installation
```bash
protoc --version
# Should show: libprotoc 3.x.x or later
```

---

## Step 2: Generate Protocol Buffer Code

This step generates Go code from the `.proto` files:

```bash
# Method 1: Use the script (recommended)
./scripts/generate-proto.sh

# Method 2: Use Makefile
make proto

# Method 3: Manual
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$PATH:$(go env GOPATH)/bin"

protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    proto/devicebridge/v1/*.proto
```

**Expected Output:**
```
proto/devicebridge/v1/common.pb.go
proto/devicebridge/v1/printer.pb.go
proto/devicebridge/v1/scanner.pb.go
proto/devicebridge/v1/scale.pb.go
proto/devicebridge/v1/display.pb.go
proto/devicebridge/v1/payment.pb.go
proto/devicebridge/v1/service.pb.go
proto/devicebridge/v1/service_grpc.pb.go
```

---

## Step 3: Install Dependencies

```bash
go mod download
go mod tidy
```

---

## Step 4: Build

```bash
# Method 1: Use Makefile
make build

# Method 2: Use Go directly
go build -o bin/device-bridge ./cmd/bridge

# Verify
ls -lh bin/device-bridge
```

---

## Step 5: Configure

Edit `configs/config.example.yaml` to match your setup:

```yaml
server:
  grpc_port: 50051
  http_port: 8080
  metrics_port: 9090
  host: "localhost"

devices:
  # For testing with virtual printer (no hardware needed)
  - id: virtual-printer-1
    name: "Virtual Test Printer"
    kind: "printer.virtual"
    transport: "virtual"
    enabled: true

  # For real ESC/POS printer (requires hardware)
  # - id: printer-1
  #   name: "Receipt Printer"
  #   kind: "printer.escpos"
  #   transport: "tcp"
  #   address: "192.168.1.100"  # Your printer's IP
  #   port: 9100
  #   enabled: true

logging:
  level: "info"
  format: "text"  # or "json"

telemetry:
  metrics:
    enabled: true
```

---

## Step 6: Run

```bash
# Method 1: Use Makefile
make run

# Method 2: Run binary directly
./bin/device-bridge --config configs/config.example.yaml

# Method 3: Run with go run
go run ./cmd/bridge --config configs/config.example.yaml
```

**Expected Output:**
```
Loading configuration from: configs/config.example.yaml

╔═══════════════════════════════════════════════════════════╗
║           Device Bridge v2.0.0-dev                        ║
╠═══════════════════════════════════════════════════════════╣
║  Status: Running                                          ║
║  gRPC:   localhost:50051                                  ║
║  HTTP:   http://localhost:8080 (REST gateway TODO)       ║
║  Metrics: http://localhost:9090/metrics                  ║
║                                                           ║
║  Devices: 1 registered                                    ║
║                                                           ║
║  Press Ctrl+C to stop                                     ║
╚═══════════════════════════════════════════════════════════╝
```

---

## Step 7: Test with grpcurl

### Install grpcurl
```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Windows
# Download from: https://github.com/fullstorydev/grpcurl/releases
```

### List Available Services
```bash
grpcurl -plaintext localhost:50051 list
```

**Output:**
```
devicebridge.v1.DeviceBridge
grpc.health.v1.Health
grpc.reflection.v1alpha.ServerReflection
```

### Ping the Service
```bash
grpcurl -plaintext localhost:50051 devicebridge.v1.DeviceBridge/Ping
```

**Output:**
```json
{
  "message": "Device Bridge is running",
  "version": "2.0.0-dev"
}
```

### List Devices
```bash
grpcurl -plaintext localhost:50051 devicebridge.v1.DeviceBridge/ListDevices
```

**Output:**
```json
{
  "devices": [
    {
      "id": "virtual-printer-1",
      "name": "Virtual Test Printer",
      "kind": "printer.virtual",
      "health": {
        "status": "DEVICE_STATUS_READY",
        "message": "virtual printer ready",
        "timestamp": "2025-11-10T..."
      },
      "lastSeen": "2025-11-10T...",
      "metadata": {
        "virtual": "true"
      }
    }
  ]
}
```

### Print a Test Receipt
```bash
grpcurl -plaintext -d '{
  "device_id": "virtual-printer-1",
  "document": {
    "sections": [{
      "elements": [{
        "line": {
          "runs": [{"text": "=== TEST RECEIPT ===", "style": {"bold": true}}],
          "alignment": "ALIGNMENT_CENTER"
        }
      }, {
        "line": {
          "runs": [{"text": "Device Bridge v2"}],
          "alignment": "ALIGNMENT_CENTER"
        }
      }, {
        "line": {
          "runs": [{"text": ""}],
          "alignment": "ALIGNMENT_LEFT"
        }
      }, {
        "line": {
          "runs": [{"text": "Item 1", "style": {}}, {"text": "         $10.00", "style": {}}],
          "alignment": "ALIGNMENT_LEFT"
        }
      }, {
        "line": {
          "runs": [{"text": "Item 2", "style": {}}, {"text": "         $15.00", "style": {}}],
          "alignment": "ALIGNMENT_LEFT"
        }
      }, {
        "line": {
          "runs": [{"text": ""}],
          "alignment": "ALIGNMENT_LEFT"
        }
      }, {
        "line": {
          "runs": [{"text": "TOTAL:", "style": {"bold": true}}, {"text": "        $25.00", "style": {"bold": true}}],
          "alignment": "ALIGNMENT_RIGHT"
        }
      }, {
        "qr_code": {
          "data": "https://github.com/Macber-eg/Flutter-Device",
          "size": 6
        }
      }]
    }],
    "options": {
      "cut": true
    }
  }
}' localhost:50051 devicebridge.v1.DeviceBridge/Print
```

**Output:**
```json
{
  "jobId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Console Output (Virtual Printer):**
```
===== VIRTUAL PRINTER OUTPUT =====
            === TEST RECEIPT ===
             Device Bridge v2

Item 1         $10.00
Item 2         $15.00

                         TOTAL:        $25.00
[QR CODE: https://github.com/Macber-eg/Flutter-Device]
--- CUT ---
===== END VIRTUAL PRINTER OUTPUT =====
```

### Get Job Status
```bash
grpcurl -plaintext -d '{
  "job_id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:50051 devicebridge.v1.DeviceBridge/GetJob
```

---

## Step 8: Test with REST/JSON API

Device Bridge v2 also exposes a complete REST/JSON API via grpc-gateway.

### Ping the Service
```bash
curl http://localhost:8080/v1/ping
```

**Output:**
```json
{
  "message": "Device Bridge is running",
  "version": "2.0.0-dev"
}
```

### List Devices
```bash
curl http://localhost:8080/v1/devices
```

**Output:**
```json
{
  "devices": [
    {
      "id": "virtual-printer-1",
      "name": "Virtual Test Printer",
      "kind": "printer.virtual",
      "health": {
        "status": "DEVICE_STATUS_READY",
        "message": "virtual printer ready",
        "timestamp": "2025-11-10T..."
      },
      "lastSeen": "2025-11-10T...",
      "metadata": {
        "virtual": "true"
      }
    }
  ]
}
```

### Get Single Device
```bash
curl http://localhost:8080/v1/devices/virtual-printer-1
```

### Print a Receipt (REST)
```bash
curl -X POST http://localhost:8080/v1/devices/virtual-printer-1/print \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "sections": [{
        "elements": [{
          "line": {
            "runs": [{"text": "=== REST API TEST ===", "style": {"bold": true}}],
            "alignment": "ALIGNMENT_CENTER"
          }
        }, {
          "line": {
            "runs": [{"text": "Device Bridge v2"}],
            "alignment": "ALIGNMENT_CENTER"
          }
        }, {
          "line": {
            "runs": [{"text": ""}],
            "alignment": "ALIGNMENT_LEFT"
          }
        }, {
          "line": {
            "runs": [{"text": "Coffee", "style": {}}, {"text": "         $3.50", "style": {}}],
            "alignment": "ALIGNMENT_LEFT"
          }
        }, {
          "line": {
            "runs": [{"text": "Muffin", "style": {}}, {"text": "         $2.50", "style": {}}],
            "alignment": "ALIGNMENT_LEFT"
          }
        }, {
          "line": {
            "runs": [{"text": ""}],
            "alignment": "ALIGNMENT_LEFT"
          }
        }, {
          "line": {
            "runs": [{"text": "TOTAL:", "style": {"bold": true}}, {"text": "         $6.00", "style": {"bold": true}}],
            "alignment": "ALIGNMENT_RIGHT"
          }
        }]
      }],
      "options": {
        "cut": true
      }
    }
  }'
```

**Output:**
```json
{
  "jobId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

### Get Job Status (REST)
```bash
curl http://localhost:8080/v1/jobs/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

### Browse API with Swagger UI

Open in your browser: **http://localhost:8080/swagger/**

The Swagger UI provides:
- Complete API documentation
- Interactive "Try it out" for all endpoints
- Request/response examples
- Schema definitions
- Download OpenAPI spec

---

## Step 9: Test with WebSocket (Real-time Events)

Device Bridge v2 also provides WebSocket endpoints for real-time event streaming.

### WebSocket Endpoints

- **Scanner Events**: `ws://localhost:8080/ws/scanner/:device_id`
- **Payment Events**: `ws://localhost:8080/ws/payment/:device_id`
- **Device Status**: `ws://localhost:8080/ws/devices`

### Using Browser Test Clients

Open the included HTML test clients in your browser:

**Scanner Test Client:**
```bash
# Open in browser
open test/client/scanner.html
# or
firefox test/client/scanner.html
# or
chrome test/client/scanner.html
```

**Payment Test Client:**
```bash
# Open in browser
open test/client/payment.html
```

The test clients provide:
- Real-time event display
- Connection status indicator
- Auto-reconnect support
- Event history with timestamps
- Beautiful, responsive UI

### JavaScript Example

```javascript
// Connect to scanner WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/scanner/virtual-scanner-1');

ws.onopen = function() {
    console.log('Connected to scanner');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Scan event:', data);
    /*
    {
        "type": "BARCODE_SCANNED",
        "device_id": "virtual-scanner-1",
        "timestamp": "2025-11-10T15:30:00Z",
        "data": {
            "barcode": "1234567890",
            "symbology": "CODE128"
        }
    }
    */
};

ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};

ws.onclose = function() {
    console.log('Disconnected from scanner');
};
```

### WebSocket Features

- **Heartbeat**: Automatic ping/pong every 54 seconds
- **Auto-cleanup**: Connections cleaned up on disconnect
- **Event Broadcasting**: Events sent to all connected clients
- **CORS Support**: Works from browser applications
- **Multiple Clients**: Unlimited concurrent connections

---

## Step 10: View Metrics

Open in browser: http://localhost:9090/metrics

**Sample Metrics:**
```
# HELP device_bridge_up Service is up (1) or down (0)
# TYPE device_bridge_up gauge
device_bridge_up 1

# HELP device_bridge_device_up Device availability (0=down, 1=up)
# TYPE device_bridge_device_up gauge
device_bridge_device_up{device_id="virtual-printer-1",kind="printer.virtual"} 1

# HELP device_bridge_jobs_total Total number of jobs by device, type, and status
# TYPE device_bridge_jobs_total counter
device_bridge_jobs_total{device_id="virtual-printer-1",status="completed",type="print"} 1

# HELP device_bridge_job_duration_seconds Job execution duration in seconds
# TYPE device_bridge_job_duration_seconds histogram
device_bridge_job_duration_seconds_bucket{device_id="virtual-printer-1",type="print",le="0.001"} 0
device_bridge_job_duration_seconds_bucket{device_id="virtual-printer-1",type="print",le="0.002"} 1
...
```

---

## Troubleshooting

### Proto Generation Fails

**Error:** `protoc: command not found`
**Solution:** Install protobuf compiler (see Step 1)

**Error:** `protoc-gen-go: program not found or is not executable`
**Solution:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Build Fails

**Error:** `cannot find package "...proto/devicebridge/v1"`
**Solution:** Generate proto code first (`make proto`)

**Error:** `go.mod file not found`
**Solution:** Run `go mod download && go mod tidy`

### Runtime Errors

**Error:** `failed to load configuration`
**Solution:** Check config file path and YAML syntax

**Error:** `failed to listen: address already in use`
**Solution:** Change port in config or kill process using port
```bash
# Find process
lsof -i :50051

# Kill it
kill -9 <PID>
```

**Error:** `device not found`
**Solution:** Check device ID in request matches config

---

## Next Steps

1. **Try with Real Hardware:**
   - Edit config to add your ESC/POS printer
   - Set `address` to printer IP
   - Set `port` to 9100 (standard ESC/POS port)

2. **Build a Client:**
   - See example in `examples/client/` (coming soon)
   - Use any gRPC client library
   - Available for: Go, Python, JavaScript, Java, C#, etc.

3. **Add More Devices:**
   - Scanners (coming soon)
   - Scales (coming soon)
   - Payment terminals (coming soon)

4. **Production Deployment:**
   - Enable TLS in config
   - Set up ACL rules
   - Use JSON logging format
   - Deploy with Docker/Kubernetes

---

## Useful Commands

```bash
# Build
make build

# Run
make run

# Test
make test

# Clean
make clean

# Generate proto
make proto

# Format code
make fmt

# Lint
make lint

# View logs (if running as service)
journalctl -u device-bridge -f

# Check version
./bin/device-bridge --version
```

---

## API Documentation

### gRPC Methods

- **Ping**: Health check
- **ListDevices**: Get all devices
- **GetDevice**: Get device by ID
- **Print**: Send print job
- **OpenDrawer**: Open cash drawer
- **SubscribeScanner**: Stream scan events (TODO)
- **GetWeight**: Read scale weight (TODO)
- **ShowDisplay**: Show text on display (TODO)
- **StartPayment**: Start payment (TODO)
- **GetJob**: Get job status

See `proto/devicebridge/v1/service.proto` for full API definition.

---

## Support

- **Documentation:** `docs/`
- **Issues:** GitHub Issues
- **Specification:** `docs/DEVICE_BRIDGE_V2_SPECIFICATION.md`
- **Progress:** `IMPLEMENTATION_STATUS.md` and `MVP_PROGRESS.md`

---

**You're ready to go! 🚀**
