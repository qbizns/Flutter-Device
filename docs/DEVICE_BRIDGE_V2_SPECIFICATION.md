# Device Bridge v2 - Official Specification

**Version:** 1.0
**Date:** 2025-11-10
**Status:** Official Specification

---

## Document Information

- **Purpose:** Define the complete architecture, requirements, and implementation approach for Device Bridge v2
- **Audience:** Developers, architects, hardware integrators, QA engineers
- **Technology Stack:** Go 1.22+, gRPC, REST/JSON, WebSocket/SSE
- **Repository:** [Macber-eg/Flutter-Device](https://github.com/Macber-eg/Flutter-Device)

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Device Catalogue & Phasing](#2-device-catalogue--phasing)
3. [Functional Requirements](#3-functional-requirements)
4. [Non-Functional Requirements](#4-non-functional-requirements)
5. [High-Level Architecture](#5-high-level-go-architecture)
6. [Deep Per-Device Integration](#6-deep-per-device-hardware-integration-notes)
7. [Implementation Roadmap](#7-implementation-roadmap)

---

## 1. System Overview

### 1.1 System Identity

**System Name:** Device Bridge v2
**Technology:** Go (Go 1.22+), gRPC, optional REST/JSON + WebSocket/SSE
**Role:** A local or network microservice that acts as a **hardware abstraction layer**

### 1.2 Core Purpose

Device Bridge v2 serves as a unified interface between:

- **Clients:** POS applications, browser UIs, back-end services
- **Devices:** Printers, scanners, scales, customer displays, payment terminals, cash drawers

### 1.3 Core Philosophy

**Clients never talk to hardware directly.** They communicate with Device Bridge via a stable, versioned API.

### 1.4 System Responsibilities

Device Bridge owns and manages:

1. **Device discovery & registration**
   - Static configuration
   - Dynamic discovery (USB, Serial, TCP, mDNS)
   - Manual API registration

2. **Connection management & retries**
   - Health monitoring
   - Automatic reconnection
   - Fault tolerance

3. **Driver implementation**
   - USB/Serial/TCP transport layers
   - Device-specific protocol handlers
   - Multi-vendor support

4. **Job scheduling**
   - Asynchronous job processing
   - Priority queuing
   - Timeout management
   - Idempotent execution

5. **Event streaming**
   - Real-time scanner events
   - Payment status updates
   - Device status changes

---

## 2. Device Catalogue & Phasing

### 2.1 Phase 1 – MVP (Must Have, Real Hardware)

These devices will be fully implemented with real hardware support:

#### 1. ESC/POS Receipt Printers
- **Transport:** USB, Serial (RS-232), TCP/IP (port 9100)
- **Use Cases:**
  - Receipt printing
  - Kitchen orders
  - Simple badges
- **Features:**
  - Text formatting (bold, underline, fonts)
  - Alignment (left, center, right)
  - QR codes and barcodes
  - Logo/bitmap images
  - Paper cutting
  - Cash drawer pulse

#### 2. Kitchen Printers
- **Transport:** Same as ESC/POS
- **Special Features:**
  - Buzzer support
  - Large font emphasis
  - Auto-cut after orders
  - Order routing by kitchen station

#### 3. Label / Barcode Printers (ZPL/EPL/CPCL)
- **Transport:** USB, Serial, TCP/IP
- **Protocols:**
  - ZPL (Zebra Programming Language)
  - EPL (Eltron Programming Language)
  - CPCL (Comtec Printer Control Language)
- **Use Cases:**
  - Product labels
  - Shipping labels
  - Shelf tags
  - Asset tags

#### 4. Barcode/QR Scanners
- **Modes:**
  - USB HID Keyboard mode (wedge)
  - USB HID Raw mode
  - USB Vendor-specific HID
  - Serial scanners (RS-232)
- **Features:**
  - Multiple symbology support
  - Configurable prefix/suffix
  - Scan event streaming

#### 5. Scales
- **Transport:** Serial (RS-232), USB-Serial
- **Protocols:**
  - At least 2-3 real vendor protocols:
    - Dibal scales
    - Mettler Toledo
    - CAS scales
- **Features:**
  - Weight reading with stability detection
  - Zero/tare commands
  - Unit conversion (kg, lb, oz)

#### 6. Customer Displays (Pole Displays)
- **Transport:** Serial (RS-232)
- **Specifications:**
  - 2×20 or 2×40 character displays
  - VFD (Vacuum Fluorescent Display)
  - LCD displays
- **Features:**
  - Line-by-line text display
  - Clear screen
  - Scrolling
  - Centered/aligned text
  - Brightness control

#### 7. Cash Drawers
- **Connection Methods:**
  - Via printer (ESC/POS pulse command)
  - Direct I/O (future phase)
- **Features:**
  - Open command
  - Status monitoring (if supported)

#### 8. Payment Terminals
- **Target:** At least one real EMV/QR provider
  - Options: Mada, Meeza, KNET, Benefit (pick one for MVP)
- **Transport:** TCP/IP, Serial
- **Features:**
  - Sale transactions
  - Refunds
  - Void
  - Settlement
  - Real-time status updates
- **Security:**
  - PCI-DSS compliance
  - No PAN logging
  - Encrypted communication

#### 9. Virtual Devices
- **Purpose:** Development and automated testing
- **Coverage:**
  - Virtual printer
  - Virtual scanner
  - Virtual scale
  - Virtual display
  - Virtual payment device
- **Benefits:**
  - CI/CD integration
  - Offline development
  - Automated testing

---

### 2.2 Phase 2 – Future Extensions

These devices are architecturally supported but deferred:

#### 10. RFID / NFC Readers
- USB HID / Serial
- MIFARE card support
- Tag read/write operations

#### 11. Magstripe Card Readers
- Keyboard wedge mode
- HID mode
- Track 1/2/3 data

#### 12. Ticket / Badge Printers
- Specialized printers not covered by ESC/POS or ZPL
- Dye-sublimation printers

#### 13. GPIO / Signalling Devices
- Signal lights (traffic lights)
- Buzzers and alarms
- Relay controllers
- Digital I/O modules

---

## 3. Functional Requirements

### 3.1 Device Management

#### FR-D1: Device Registry
The system SHALL maintain an internal registry of all known devices with:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique device identifier |
| `name` | string | Human-readable name |
| `kind` | string | Device type (e.g., "printer.escpos") |
| `vendor` | string | Manufacturer |
| `model` | string | Model number/name |
| `connection` | object | Connection details (type, address, port) |
| `status` | enum | Current device status |
| `last_seen` | timestamp | Last successful communication |
| `metadata` | map | Custom key-value pairs |

#### FR-D2: Provisioning Modes
The system SHALL support three device provisioning modes:

1. **Static Configuration**
   - Devices defined in YAML/JSON configuration file
   - Loaded at startup
   - Suitable for fixed installations

2. **Auto Discovery**
   - USB device enumeration
   - Serial port scanning
   - TCP network scanning
   - mDNS/Bonjour service discovery
   - Continuous background scanning

3. **API-Created Devices**
   - Administrative API for manual registration
   - Runtime device addition
   - Suitable for dynamic environments

#### FR-D3: Device Management APIs
The system SHALL expose APIs to:

| API | Method | Description |
|-----|--------|-------------|
| `ListDevices` | GET | List all registered devices with filters |
| `GetDevice` | GET | Get detailed device information |
| `EnableDevice` | POST | Enable a disabled device |
| `DisableDevice` | POST | Disable an active device |
| `UpdateDevice` | PUT | Update device metadata/tags |
| `DeleteDevice` | DELETE | Remove device from registry |
| `TagDevice` | POST | Add tags (e.g., "lane-1", "kitchen-A") |

#### FR-D4: Device Status Model
Device status SHALL be one of:

| Status | Description |
|--------|-------------|
| `unknown` | Device state cannot be determined |
| `ready` | Device is operational and accepting jobs |
| `degraded` | Device is operational but with issues |
| `offline` | Device is not reachable |
| `error` | Device is in error state |

Each status SHALL include:
- Timestamp of last status change
- Human-readable reason/message
- Error code (if applicable)

---

### 3.2 Job Scheduling & Execution

#### FR-J1: Job Model
All operations (print, weight read, display show, payment, etc.) SHALL be modeled as **jobs** with:

| Field | Type | Description |
|-------|------|-------------|
| `job_id` | string (UUID) | Unique job identifier |
| `device_id` | string | Target device ID |
| `type` | enum | Job type (print, scan, weigh, etc.) |
| `payload` | object | Job-specific data |
| `idempotency_key` | string | Client-provided uniqueness key |
| `created_at` | timestamp | Job creation time |
| `started_at` | timestamp | Job execution start |
| `completed_at` | timestamp | Job completion time |
| `status` | enum | Job status |
| `result` | object | Job result data |
| `error` | object | Error details (if failed) |

#### FR-J2: Asynchronous Processing
The bridge SHALL process jobs asynchronously via a bounded worker pool:

- Configurable concurrency per device type
  - Example: 4 concurrent print jobs, 10 scan streams
- Per-job timeout configuration
- Graceful shutdown with job completion
- Job queue persistence (optional, for restarts)

#### FR-J3: Idempotency
Jobs SHALL be **idempotent**:

- If a job with the same `idempotency_key` is submitted twice, it MUST not execute twice
- Return the original job result for duplicate requests
- Idempotency window: configurable (default 24 hours)

#### FR-J4: Job History
The system SHALL record job history for troubleshooting:

- Last N jobs per device (configurable, default 100)
- Status and error details
- Execution duration
- Searchable by device, status, time range
- Automatic cleanup of old history

#### FR-J5: Long-Running Operations
For long-running operations (payments, continuous scanning), the system SHALL support:

- Subscribe to updates via gRPC streams
- WebSocket connections for browser clients
- Server-Sent Events (SSE) as fallback
- Automatic reconnection handling

---

### 3.3 ESC/POS & Label Printing

#### FR-P1: Print Job Structure
The system SHALL accept high-level print jobs with:

**Receipt Elements:**
- Static text with formatting
  - Alignment: left, center, right
  - Styles: bold, underline, inverse, emphasized
  - Font sizes: normal, double height, double width, both
- Item lists (SKU, name, quantity, price)
- Totals and subtotals
- VAT/tax lines
- QR codes and barcodes (multiple formats)
- Logo/bitmap images (monochrome)
- Cut command
- Cash drawer pulse

**Template Support:**
```json
{
  "template_id": "receipt_standard",
  "data": {
    "store_name": "My Store",
    "items": [...],
    "total": 99.99,
    "tax": 15.00
  }
}
```

#### FR-P2: Rendering Pipeline
The system SHALL render print jobs into vendor-neutral objects:

```
PrintJob → Document → Sections → Lines → Runs
```

Where:
- **Document**: Top-level container
- **Section**: Logical grouping (header, items, footer)
- **Line**: Single line of output
- **Run**: Segment with consistent styling

#### FR-P3: Text Support
The system SHALL support:

- **Character Sets:**
  - ASCII (basic English)
  - UTF-8 (full Unicode support)
  - Code pages (CP437, CP850, CP858, CP1252, etc.)

- **Languages:**
  - Latin scripts
  - Arabic (RTL, text shaping, ligatures)
  - Other RTL languages (Hebrew, Urdu)
  - Bidirectional text (mixed LTR/RTL)

- **Arabic Rendering:**
  - Text shaping engine (contextual forms)
  - Bidirectional algorithm
  - Fallback to bitmap rendering if printer doesn't support Arabic codepoints

#### FR-P4: Command Translation
The system SHALL translate rendered documents into device-specific commands:

**ESC/POS Printers:**
- Initialize: `ESC @`
- Text encoding: `ESC t n`
- Alignment: `ESC a n`
- Bold: `ESC E n`
- Underline: `ESC - n`
- Font: `ESC M n`
- Size: `GS ! n`
- Barcode: `GS k m d1...dk NUL`
- QR code: `GS ( k ...`
- Image: Raster or column format
- Cut: `GS V m n`
- Drawer pulse: `ESC p m t1 t2`

**ZPL Label Printers:**
- Field origin: `^FO`
- Field data: `^FD...^FS`
- Barcode: `^B...`
- Graphics: `^GF`
- Print quantity: `^PQ`

**EPL Label Printers:**
- Similar command set to ZPL
- Different syntax

#### FR-P5: Print Templates
The system SHALL support templates for:

- **Receipt Templates:**
  - Standard receipt
  - Gift receipt
  - Return receipt
  - Refund receipt

- **Kitchen Tickets:**
  - Order ticket
  - Preparation instructions
  - Modifier handling

- **Labels:**
  - Product labels (with placeholders)
  - Price tags
  - Shelf labels
  - Shipping labels

Templates SHALL support:
- Variable substitution
- Conditional sections
- Loops (for line items)
- Custom formatting functions

---

### 3.4 Scanner Events

#### FR-S1: Event-Driven Model
The system SHALL treat scanners as **event-emitting devices**.

Scanners continuously emit events when barcodes/QR codes are read.

#### FR-S2: Event Streaming APIs
The system SHALL expose:

**gRPC Stream:**
```protobuf
rpc SubscribeScanner(SubscribeScannerRequest) returns (stream ScanEvent);
```

**WebSocket Endpoint:**
```
ws://bridge-host/ws/scanner/{device_id}
```

**Event Structure:**
```json
{
  "device_id": "scanner-1",
  "data": "1234567890123",
  "symbology": "EAN13",
  "timestamp": "2025-11-10T12:34:56Z",
  "event_id": "evt_123456"
}
```

**Supported Symbologies:**
- EAN13, EAN8, UPCA, UPCE
- Code39, Code128, Code93
- QR Code, Data Matrix, PDF417
- Codabar, ITF, MSI
- Aztec, MaxiCode

#### FR-S3: Scanner Modes
The system SHALL support:

1. **USB HID Keyboard Mode**
   - Scanner emulates keyboard
   - Decoded as string input
   - Configurable prefix/suffix
   - Limited control from bridge

2. **USB HID Raw Mode**
   - Direct HID communication
   - Full control over scanner
   - Read scan codes directly
   - Map HID reports to characters

3. **Serial Mode**
   - RS-232 connection
   - Vendor-specific protocols
   - Parse binary data streams

---

### 3.5 Scales

#### FR-SC1: Weight Reading API
The system SHALL expose a `GetWeight` API (unary RPC):

**Request:**
```json
{
  "device_id": "scale-1",
  "timeout_ms": 5000
}
```

**Response:**
```json
{
  "weight": 1.234,
  "unit": "kg",
  "stable": true,
  "timestamp": "2025-11-10T12:34:56Z"
}
```

#### FR-SC2: Scale Operations
The system SHALL support:

- **Stable Weight Detection:**
  - Debounce noisy readings
  - Configurable stability threshold
  - Configurable stability duration

- **Zero/Tare Commands:**
  - `ZeroScale(device_id)`: Set current reading as zero
  - `TareScale(device_id)`: Tare container weight

- **Continuous Reading (optional):**
  - Stream weight updates
  - Useful for live displays

#### FR-SC3: Vendor Protocol Support
The system SHALL handle vendor-specific protocols via pluggable "scale drivers":

**Protocol Interface:**
```go
type ScaleProtocol interface {
    RequestWeight(conn io.ReadWriter) (WeightReading, error)
    Zero(conn io.ReadWriter) error
    Tare(conn io.ReadWriter) error
    ParseResponse(data []byte) (WeightReading, error)
}
```

**Supported Protocols (Phase 1):**
- Dibal scales (serial protocol)
- Mettler Toledo (MT-SICS protocol)
- CAS scales (serial protocol)

---

### 3.6 Customer Display

#### FR-CD1: Display APIs
The system SHALL expose:

**Show Lines:**
```json
{
  "device_id": "display-1",
  "line1": "Welcome!",
  "line2": "Total: $99.99",
  "duration_ms": 5000  // optional, 0 = permanent
}
```

**Clear:**
```json
{
  "device_id": "display-1"
}
```

**Set Brightness (if supported):**
```json
{
  "device_id": "display-1",
  "brightness": 75  // percentage
}
```

#### FR-CD2: Text Formatting
The system SHALL implement:

- **Line Wrapping:** Auto-wrap text exceeding display width
- **Centering:** Center text within line width
- **Padding:** Pad short text with spaces
- **Truncation:** Truncate overlong text with "..."
- **Character Sanitization:** Replace unsupported characters with '?'

---

### 3.7 Cash Drawer

#### FR-DR1: Drawer Control API
The system SHALL expose:

```json
{
  "device_id": "drawer-1"
}
```

**Response:**
```json
{
  "success": true,
  "opened_at": "2025-11-10T12:34:56Z"
}
```

#### FR-DR2: Printer-Connected Drawers
For drawers connected via ESC/POS printers, the system SHALL:

- Send the correct ESC/POS pulse command:
  - Standard: `ESC p 0 50 50` (connector 1, 250ms pulse)
  - Configurable: connector, on time, off time
- Report errors if printer is not reachable
- Verify printer supports drawer pulse
- Log all open events

---

### 3.8 Payments

#### FR-PAY1: Payment API
The system SHALL expose a consistent payment API regardless of provider:

**Start Payment:**
```json
{
  "amount": 99.99,
  "currency": "USD",
  "reference": "order_123456",
  "metadata": {
    "order_id": "123456",
    "customer_id": "789"
  }
}
```

**Response:**
```json
{
  "payment_id": "pmt_abc123",
  "status": "initiated",
  "initiated_at": "2025-11-10T12:34:56Z"
}
```

**Cancel Payment:**
```json
{
  "payment_id": "pmt_abc123"
}
```

**Get Payment Status:**
```json
{
  "payment_id": "pmt_abc123"
}
```

**Response:**
```json
{
  "payment_id": "pmt_abc123",
  "status": "approved",
  "amount": 99.99,
  "currency": "USD",
  "card_type": "visa",
  "last_4": "1234",
  "approval_code": "123456",
  "completed_at": "2025-11-10T12:35:10Z"
}
```

#### FR-PAY2: Payment Status States
Payment status states SHALL at least be:

| Status | Description |
|--------|-------------|
| `initiated` | Payment request sent to terminal |
| `in_progress` | Customer interacting with terminal |
| `approved` | Payment successful |
| `declined` | Payment declined by issuer |
| `cancelled` | Payment cancelled by clerk/customer |
| `timeout` | Payment timed out |
| `error` | Technical error occurred |

#### FR-PAY3: Error Handling
The system SHALL handle:

- **Timeouts from Device:**
  - Configurable timeout (default 90 seconds)
  - Automatic status query on timeout
  - Prevent double charges

- **Terminal Not Reachable:**
  - Fast failure with clear error code
  - Automatic retry with backoff (optional)
  - Device health monitoring

- **Duplicate Requests:**
  - Detect duplicate merchant reference
  - Return existing payment status
  - Prevent duplicate charges

- **PCI-DSS Compliance:**
  - Never log PAN (Primary Account Number)
  - Never log CVV/track data
  - Encrypt sensitive data in transit
  - Mask card numbers in responses (last 4 only)

#### FR-PAY4: Event Streaming
The system SHALL stream payment events:

**Events:**
- `card_inserted`
- `pin_entry_started`
- `pin_entry_completed`
- `processing`
- `approved`
- `declined`
- `cancelled`
- `error`

---

### 3.9 API Interfaces

#### FR-API1: Multi-Protocol Support
All core functionality SHALL be exposed via:

1. **gRPC (Primary)**
   - High performance
   - Type-safe
   - Bidirectional streaming
   - Service definition in `.proto` files

2. **REST/JSON (via grpc-gateway)**
   - HTTP/1.1 compatibility
   - Standard REST verbs
   - JSON request/response
   - OpenAPI/Swagger documentation

3. **WebSocket**
   - For events (scans, payments)
   - Browser-friendly
   - Auto-reconnection support

4. **Server-Sent Events (SSE)**
   - Fallback for WebSocket
   - Simple one-way streaming
   - Works through proxies

#### FR-API2: Versioning
APIs SHALL be versioned:

- Package: `devicebridge.v1`
- URL path: `/v1/...`
- Proto package: `devicebridge.v1`
- Room for v2, v3 in future

**Versioning Policy:**
- Breaking changes require new major version
- Backward-compatible changes allowed in same version
- Support N-1 version for migration period

#### FR-API3: Observability
All APIs SHALL support:

- **Correlation IDs:**
  - Request ID in headers (`X-Request-ID`)
  - Propagated through all operations
  - Included in logs and traces

- **Request Tracing:**
  - OpenTelemetry spans
  - Distributed tracing
  - Performance monitoring

- **Error Codes:**
  - Machine-readable error codes
  - Human-readable error messages
  - Error details (when appropriate)

---

## 4. Non-Functional Requirements

### 4.1 Performance

#### NFR-P1: Throughput
Bridge MUST handle at least:

| Operation | Throughput |
|-----------|-----------|
| Print jobs per lane | 20 jobs/minute |
| Concurrent payment flows | 10 simultaneous |
| Scanner events per device | 20 events/second (burst) |
| Weight readings | 10 readings/second |
| Display updates | 30 updates/second |

#### NFR-P2: Latency
Under normal conditions:

| Operation | Target Latency | Maximum |
|-----------|----------------|---------|
| Print job submission to printer | < 200 ms | 500 ms |
| Weight reading | < 1 second | 3 seconds |
| Payment initiation | < 500 ms | 1 second |
| Scanner event delivery | < 50 ms | 100 ms |
| API response time (non-blocking) | < 100 ms | 300 ms |

#### NFR-P3: Resource Usage

| Resource | Target | Maximum |
|----------|--------|---------|
| Memory (idle) | < 100 MB | 200 MB |
| Memory (under load) | < 500 MB | 1 GB |
| CPU (idle) | < 5% | 10% |
| CPU (under load) | < 50% | 80% |
| Disk space (logs/history) | < 1 GB | 5 GB |

---

### 4.2 Reliability & Fault Tolerance

#### NFR-R1: Connection Resilience
Device connection failures MUST be retried with exponential backoff:

- Initial retry: 1 second
- Max retry interval: 60 seconds
- Max retry duration: Configurable (default: indefinite)
- Circuit breaker pattern for persistent failures

#### NFR-R2: Error Reporting
When a device is offline or unreachable, API MUST return clear error codes:

| Error Code | HTTP | gRPC | Description |
|------------|------|------|-------------|
| `DEVICE_NOT_FOUND` | 404 | NOT_FOUND | Device ID doesn't exist |
| `DEVICE_OFFLINE` | 503 | UNAVAILABLE | Device not reachable |
| `DEVICE_BUSY` | 409 | ABORTED | Device busy with another job |
| `CONNECTION_LOST` | 503 | UNAVAILABLE | Connection dropped during operation |
| `OPERATION_TIMEOUT` | 504 | DEADLINE_EXCEEDED | Operation timed out |

#### NFR-R3: Process Stability
The bridge process SHOULD be able to run for weeks without restart:

- No memory leaks
- Proper cleanup of resources
- Graceful handling of device disconnections
- Log rotation and cleanup
- Automatic recovery from transient errors

#### NFR-R4: Data Integrity
- Job history must be consistent
- No lost events (best effort delivery)
- Idempotency for critical operations
- Transaction logging for payments

---

### 4.3 Security

#### NFR-S1: Deployment Modes

**Development Mode:**
- No TLS (HTTP/gRPC)
- Listen on localhost only (127.0.0.1)
- No authentication
- Verbose logging
- Enabled by default

**Production Mode:**
- mTLS (mutual TLS) required
- ACL (Access Control Lists)
- Listen on configured interface
- Structured logging (no sensitive data)
- Requires explicit configuration

#### NFR-S2: Access Control
ACL rules MUST allow:

**Per-client/app configuration:**
```yaml
acl:
  - client_id: pos_app_1
    allowed_devices:
      - "printer-*"
      - "drawer-*"
      - "display-*"
    allowed_operations:
      - print
      - open_drawer
      - display_show

  - client_id: payment_app
    allowed_devices:
      - "payment-*"
    allowed_operations:
      - start_payment
      - cancel_payment
```

**Features:**
- Device pattern matching (wildcards)
- Operation whitelisting
- Client authentication via mTLS certificates
- API key support (optional)

#### NFR-S3: Sensitive Data Protection
Sensitive data (PAN, track data, PINs, etc.) MUST:

- **Never be logged** (even in debug mode)
- **Never be stored** in job history
- **Always be encrypted** in transit (TLS)
- **Be masked** in API responses (show last 4 digits only)
- **Follow PCI-DSS** requirements for payment data

#### NFR-S4: Security Hardening
- Minimal attack surface (only required ports)
- Input validation on all APIs
- Rate limiting (configurable)
- No SQL injection (we don't use SQL, but principle applies)
- No command injection (sanitize all device input)
- Secure defaults

---

### 4.4 Observability

#### NFR-O1: Logging
Structured JSON logging with:

**Log Levels:**
- `DEBUG`: Detailed debugging (dev mode only)
- `INFO`: Normal operations
- `WARN`: Recoverable errors
- `ERROR`: Operation failures
- `FATAL`: System failures

**Standard Fields:**
```json
{
  "timestamp": "2025-11-10T12:34:56.789Z",
  "level": "INFO",
  "message": "Print job completed",
  "device_id": "printer-1",
  "job_id": "job_123",
  "request_id": "req_456",
  "duration_ms": 234,
  "component": "printer_driver"
}
```

**Log Rotation:**
- Max file size: 100 MB
- Max files: 10
- Automatic cleanup

#### NFR-O2: Metrics
Prometheus metrics SHALL include:

**Device Metrics:**
- `device_status{device_id, kind}` - Device status (gauge)
- `device_up{device_id, kind}` - Device availability (gauge: 0 or 1)

**Job Metrics:**
- `jobs_total{device_id, type, status}` - Total jobs (counter)
- `job_duration_seconds{device_id, type}` - Job duration (histogram)
- `jobs_in_progress{device_id, type}` - Current jobs (gauge)

**Payment Metrics:**
- `payments_total{provider, status}` - Total payments (counter)
- `payment_amount_total{provider, currency}` - Payment amounts (counter)
- `payment_duration_seconds{provider}` - Payment duration (histogram)

**System Metrics:**
- `up` - Service up (gauge)
- `process_cpu_seconds_total` - CPU usage
- `process_resident_memory_bytes` - Memory usage
- `http_requests_total{method, path, status}` - HTTP requests
- `grpc_server_handled_total{method, code}` - gRPC requests

#### NFR-O3: Tracing
Optional OpenTelemetry tracing:

**Spans for:**
- API requests (gRPC, REST)
- Job execution
- Device operations (connect, send, receive)
- External calls (payment provider)

**Span Attributes:**
- `device.id`
- `device.kind`
- `job.id`
- `job.type`
- `payment.id`
- `error.message` (if failed)

**Exporters:**
- Jaeger
- Zipkin
- OTLP (OpenTelemetry Protocol)
- Console (debug)

#### NFR-O4: Health Checks
Standard health check endpoints:

**Liveness:**
- `GET /healthz` or gRPC `Health.Check`
- Returns 200 if process is alive
- Used by orchestrators (Kubernetes, Docker)

**Readiness:**
- `GET /readyz` or gRPC `Health.Watch`
- Returns 200 if ready to serve traffic
- Checks critical dependencies

---

### 4.5 Deployment

#### NFR-D1: Single Binary
The bridge MUST run as a single binary:

- No external dependencies except OS libraries
- Embedded assets (templates, configs)
- Cross-compilation for all targets
- Static linking where possible

#### NFR-D2: Supported Platforms

| OS | Architectures | Status |
|----|---------------|--------|
| Linux | amd64, arm64, arm | Tier 1 (full support) |
| macOS | amd64 (Intel), arm64 (Apple Silicon) | Tier 1 |
| Windows | amd64 | Tier 2 (best effort) |
| FreeBSD | amd64 | Tier 3 (community) |

**OS-Specific Features:**
- Linux: Full USB/Serial/TCP support
- macOS: Full USB/Serial/TCP support, System Extensions may be required
- Windows: Full support, driver installation may be required

#### NFR-D3: Deployment Methods

**1. Systemd Service (Linux)**
```ini
[Unit]
Description=Device Bridge v2
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/device-bridge --config /etc/device-bridge/config.yaml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

**2. Docker Container**
```dockerfile
FROM alpine:latest
COPY device-bridge /usr/local/bin/
EXPOSE 50051 8080
CMD ["device-bridge", "--config", "/config/config.yaml"]
```

Features:
- Privileged mode for USB access
- Volume mounts for config and logs
- Host network for device discovery

**3. Standalone Binary**
- Download and run
- No installation required
- Portable configuration

**4. Kubernetes (future)**
- Helm chart
- DaemonSet for per-node deployment
- Device plugin integration

#### NFR-D4: Configuration Management
Support multiple config sources:

1. **File** (YAML, JSON, TOML)
2. **Environment variables**
3. **Command-line flags**
4. **Remote config** (etcd, Consul - future)

Priority: CLI flags > Env vars > Config file > Defaults

---

## 5. High-Level Go Architecture

### 5.1 Directory Structure

```
device-bridge-v2/
├── cmd/
│   ├── bridge/              # Main daemon
│   │   └── main.go
│   └── bridge-cli/          # Admin/debugging CLI
│       └── main.go
├── internal/
│   ├── api/                 # gRPC + REST + WebSocket handlers
│   │   ├── grpc/
│   │   ├── rest/
│   │   └── ws/
│   ├── app/                 # Application orchestration
│   │   ├── registry.go      # Device registry
│   │   ├── scheduler.go     # Job scheduler
│   │   └── events.go        # Event bus
│   ├── devices/             # Device interfaces + registry
│   │   ├── device.go        # Base device interface
│   │   ├── printer/
│   │   │   ├── printer.go
│   │   │   └── document.go
│   │   ├── scanner/
│   │   │   ├── scanner.go
│   │   │   └── events.go
│   │   ├── scale/
│   │   │   └── scale.go
│   │   ├── display/
│   │   │   └── display.go
│   │   ├── drawer/
│   │   │   └── drawer.go
│   │   └── payment/
│   │       └── payment.go
│   ├── drivers/             # Concrete implementations
│   │   ├── printer_escpos/
│   │   │   ├── driver.go
│   │   │   ├── commands.go
│   │   │   └── renderer.go
│   │   ├── printer_zpl/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── scanner_hid/
│   │   │   ├── driver.go
│   │   │   └── decoder.go
│   │   ├── scanner_serial/
│   │   │   └── driver.go
│   │   ├── scale_serial/
│   │   │   ├── driver.go
│   │   │   └── protocols/
│   │   │       ├── dibal.go
│   │   │       ├── mettler.go
│   │   │       └── cas.go
│   │   ├── display_serial/
│   │   │   └── driver.go
│   │   └── payment_xxx/
│   │       └── driver.go
│   ├── discovery/           # Device discovery
│   │   ├── usb.go
│   │   ├── serial.go
│   │   ├── tcp.go
│   │   └── mdns.go
│   ├── jobs/                # Job queue & workers
│   │   ├── queue.go
│   │   ├── worker.go
│   │   └── history.go
│   ├── events/              # Pub/sub for events
│   │   ├── bus.go
│   │   └── subscriber.go
│   ├── config/              # Configuration
│   │   ├── config.go
│   │   ├── loader.go
│   │   └── validator.go
│   ├── security/            # Security
│   │   ├── tls.go
│   │   ├── acl.go
│   │   └── auth.go
│   └── telemetry/           # Observability
│       ├── logger.go
│       ├── metrics.go
│       └── tracing.go
├── proto/
│   └── devicebridge/
│       └── v1/
│           ├── device.proto
│           ├── printer.proto
│           ├── scanner.proto
│           ├── scale.proto
│           ├── display.proto
│           ├── drawer.proto
│           └── payment.proto
├── test/
│   ├── integration/         # Integration tests
│   │   ├── printer_test.go
│   │   ├── scanner_test.go
│   │   └── payment_test.go
│   └── virtual_devices/     # Mock devices for testing
│       ├── virtual_printer.go
│       ├── virtual_scanner.go
│       └── virtual_scale.go
├── configs/
│   ├── config.example.yaml
│   └── devices.example.yaml
├── docs/
│   ├── DEVICE_BRIDGE_V2_SPECIFICATION.md  # This document
│   ├── API.md
│   └── DEPLOYMENT.md
├── scripts/
│   ├── build.sh
│   ├── test.sh
│   └── install.sh
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

### 5.2 Key Interfaces

#### 5.2.1 Device Interface

```go
package devices

import "context"

// Device is the base interface all devices must implement
type Device interface {
    // ID returns the unique device identifier
    ID() string

    // Kind returns the device type (e.g., "printer.escpos", "scale.serial")
    Kind() string

    // Name returns the human-readable device name
    Name() string

    // Start initializes the device and begins operation
    Start(ctx context.Context) error

    // Stop gracefully shuts down the device
    Stop() error

    // Health returns the current health status
    Health() DeviceHealth

    // Metadata returns device-specific metadata
    Metadata() map[string]string
}

// DeviceHealth represents device health status
type DeviceHealth struct {
    Status    HealthStatus
    Message   string
    Timestamp time.Time
    Error     error
}

// HealthStatus represents device status
type HealthStatus int

const (
    HealthUnknown  HealthStatus = iota
    HealthReady
    HealthDegraded
    HealthOffline
    HealthError
)
```

#### 5.2.2 Printer Interface

```go
package printer

import (
    "context"
    "device-bridge-v2/internal/devices"
)

// Printer interface extends Device with print-specific methods
type Printer interface {
    devices.Device

    // Print sends a document to the printer
    Print(ctx context.Context, doc *PrintDocument) error

    // OpenDrawer opens the cash drawer (if connected)
    OpenDrawer(ctx context.Context) error

    // GetStatus queries printer status
    GetStatus(ctx context.Context) (*PrinterStatus, error)
}

// PrintDocument represents a high-level print job
type PrintDocument struct {
    Sections []Section
    Options  PrintOptions
}

// Section represents a logical section of the document
type Section struct {
    Lines []Line
}

// Line represents a single line of output
type Line struct {
    Runs      []Run
    Alignment Alignment
}

// Run represents styled text
type Run struct {
    Text  string
    Style TextStyle
}

// TextStyle defines text formatting
type TextStyle struct {
    Bold      bool
    Underline bool
    Inverse   bool
    DoubleWidth  bool
    DoubleHeight bool
    Font      Font
}

// Alignment defines text alignment
type Alignment int

const (
    AlignLeft   Alignment = iota
    AlignCenter
    AlignRight
)

// Font defines font selection
type Font int

const (
    FontA Font = iota  // Usually 12x24
    FontB              // Usually 9x17
)

// PrintOptions contains print job options
type PrintOptions struct {
    Cut           bool
    DrawerPulse   bool
    Copies        int
}

// PrinterStatus represents current printer state
type PrinterStatus struct {
    Online       bool
    PaperPresent bool
    DrawerOpen   bool
    Error        string
}
```

#### 5.2.3 Scanner Interface

```go
package scanner

import (
    "context"
    "device-bridge-v2/internal/devices"
)

// Scanner interface extends Device with scan event streaming
type Scanner interface {
    devices.Device

    // Events returns a channel of scan events
    // The channel is closed when the scanner stops
    Events(ctx context.Context) (<-chan ScanEvent, error)

    // Configure applies scanner settings
    Configure(ctx context.Context, config ScannerConfig) error
}

// ScanEvent represents a barcode/QR scan event
type ScanEvent struct {
    DeviceID  string
    Data      string
    Symbology string
    Timestamp time.Time
    EventID   string
}

// ScannerConfig contains scanner configuration
type ScannerConfig struct {
    Prefix    string
    Suffix    string
    Timeout   time.Duration
}
```

#### 5.2.4 Scale Interface

```go
package scale

import (
    "context"
    "device-bridge-v2/internal/devices"
)

// Scale interface extends Device with weight reading
type Scale interface {
    devices.Device

    // ReadWeight reads the current weight
    ReadWeight(ctx context.Context) (WeightReading, error)

    // Zero sets the current reading as zero point
    Zero(ctx context.Context) error

    // Tare subtracts container weight
    Tare(ctx context.Context) error
}

// WeightReading represents a weight measurement
type WeightReading struct {
    Weight    float64
    Unit      WeightUnit
    Stable    bool
    Timestamp time.Time
}

// WeightUnit defines weight units
type WeightUnit string

const (
    UnitKilogram WeightUnit = "kg"
    UnitGram     WeightUnit = "g"
    UnitPound    WeightUnit = "lb"
    UnitOunce    WeightUnit = "oz"
)

// ScaleProtocol defines vendor-specific scale protocols
type ScaleProtocol interface {
    // RequestWeight sends a weight request and returns the reading
    RequestWeight(conn io.ReadWriter) (WeightReading, error)

    // Zero sends zero command
    Zero(conn io.ReadWriter) error

    // Tare sends tare command
    Tare(conn io.ReadWriter) error

    // ParseResponse parses raw bytes into weight reading
    ParseResponse(data []byte) (WeightReading, error)
}
```

#### 5.2.5 Display Interface

```go
package display

import (
    "context"
    "device-bridge-v2/internal/devices"
)

// Display interface extends Device with display methods
type Display interface {
    devices.Device

    // ShowLines displays text on the customer display
    ShowLines(ctx context.Context, lines []string, duration time.Duration) error

    // Clear clears the display
    Clear(ctx context.Context) error

    // SetBrightness adjusts display brightness (0-100)
    SetBrightness(ctx context.Context, percent int) error
}
```

#### 5.2.6 Cash Drawer Interface

```go
package drawer

import (
    "context"
    "device-bridge-v2/internal/devices"
)

// Drawer interface extends Device with drawer control
type Drawer interface {
    devices.Device

    // Open opens the cash drawer
    Open(ctx context.Context) error

    // GetStatus returns drawer status (if supported)
    GetStatus(ctx context.Context) (*DrawerStatus, error)
}

// DrawerStatus represents drawer state
type DrawerStatus struct {
    Open bool
    LastOpened time.Time
}
```

#### 5.2.7 Payment Provider Interface

```go
package payment

import (
    "context"
    "device-bridge-v2/internal/devices"
)

// PaymentProvider interface extends Device with payment processing
type PaymentProvider interface {
    devices.Device

    // StartPayment initiates a payment transaction
    StartPayment(ctx context.Context, req PaymentRequest) (PaymentID, error)

    // CancelPayment cancels an in-progress payment
    CancelPayment(ctx context.Context, id PaymentID) error

    // Status returns the current payment status
    Status(ctx context.Context, id PaymentID) (PaymentStatus, error)

    // Events streams payment events
    Events(ctx context.Context, id PaymentID) (<-chan PaymentEvent, error)
}

// PaymentID is a unique payment identifier
type PaymentID string

// PaymentRequest represents a payment initiation
type PaymentRequest struct {
    Amount            float64
    Currency          string
    MerchantReference string
    Metadata          map[string]string
    Timeout           time.Duration
}

// PaymentStatus represents payment state
type PaymentStatus struct {
    ID            PaymentID
    Status        PaymentState
    Amount        float64
    Currency      string
    CardType      string
    Last4         string
    ApprovalCode  string
    Error         string
    InitiatedAt   time.Time
    CompletedAt   *time.Time
}

// PaymentState defines payment states
type PaymentState int

const (
    PaymentInitiated PaymentState = iota
    PaymentInProgress
    PaymentApproved
    PaymentDeclined
    PaymentCancelled
    PaymentTimeout
    PaymentError
)

// PaymentEvent represents a payment state change
type PaymentEvent struct {
    PaymentID PaymentID
    Event     string
    Status    PaymentState
    Message   string
    Timestamp time.Time
}
```

---

### 5.3 Application Layer

#### Registry + Scheduler

```go
package app

// Registry manages all devices
type Registry struct {
    devices map[string]devices.Device
    mu      sync.RWMutex
}

func (r *Registry) Register(dev devices.Device) error
func (r *Registry) Unregister(id string) error
func (r *Registry) Get(id string) (devices.Device, error)
func (r *Registry) List(filter DeviceFilter) []devices.Device

// Scheduler manages job execution
type Scheduler struct {
    registry *Registry
    queues   map[string]*jobs.Queue
    workers  *jobs.WorkerPool
}

func (s *Scheduler) Submit(job *jobs.Job) error
func (s *Scheduler) Cancel(jobID string) error
func (s *Scheduler) Status(jobID string) (*jobs.JobStatus, error)
```

---

### 5.4 API Layer

The API layer is thin - it just calls app services:

```go
package grpc

type DeviceBridgeServer struct {
    registry  *app.Registry
    scheduler *app.Scheduler
    events    *events.Bus
}

func (s *DeviceBridgeServer) ListDevices(ctx context.Context, req *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
    devices := s.registry.List(toFilter(req.Filter))
    return toProtoDeviceList(devices), nil
}

func (s *DeviceBridgeServer) Print(ctx context.Context, req *pb.PrintRequest) (*pb.PrintResponse, error) {
    job := toPrintJob(req)
    err := s.scheduler.Submit(job)
    return &pb.PrintResponse{JobId: job.ID}, err
}

func (s *DeviceBridgeServer) SubscribeScanner(req *pb.SubscribeScannerRequest, stream pb.DeviceBridge_SubscribeScannerServer) error {
    sub := s.events.Subscribe(req.DeviceId)
    defer sub.Unsubscribe()

    for event := range sub.Events() {
        if err := stream.Send(toProtoScanEvent(event)); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 6. Deep Per-Device Hardware Integration Notes

This section provides **real-world hardware integration details** for each device type.

---

### 6.1 ESC/POS Printers

#### 6.1.1 Transports

**USB:**
- Use `gousb` library or OS device paths
- Typical path: `/dev/usb/lp0`, `/dev/usb/lp1` on Linux
- Windows: Open printer by name via Win32 API
- macOS: IOKit or CUPS

**Serial:**
- Device: `/dev/ttyUSB0`, `/dev/ttyACM0`, `/dev/ttyS0` (Linux)
- Windows: `COM1`, `COM2`, etc.
- Settings: Typically `9600 8N1` or `115200 8N1`
- Use `github.com/tarm/serial` or `go.bug.st/serial`

**TCP/IP:**
- Raw socket to `printer_ip:9100`
- No protocol, just send ESC/POS bytes
- Implement timeout and retry logic
- Check connection with keepalive

#### 6.1.2 Driver Implementation

**Connection Management:**
```go
type ESCPOSDriver struct {
    config     ESCPOSConfig
    conn       io.ReadWriteCloser
    mu         sync.Mutex
    lastUsed   time.Time
}

func (d *ESCPOSDriver) Connect(ctx context.Context) error {
    switch d.config.Transport {
    case "tcp":
        conn, err := net.DialTimeout("tcp", d.config.Address, 5*time.Second)
        if err != nil {
            return fmt.Errorf("tcp connect failed: %w", err)
        }
        d.conn = conn

    case "serial":
        conn, err := serial.Open(&serial.Config{
            Name: d.config.Port,
            Baud: d.config.BaudRate,
            Parity: serial.ParityNone,
            StopBits: serial.Stop1,
            Size: 8,
        })
        if err != nil {
            return fmt.Errorf("serial open failed: %w", err)
        }
        d.conn = conn

    default:
        return fmt.Errorf("unsupported transport: %s", d.config.Transport)
    }

    return d.initialize()
}

func (d *ESCPOSDriver) initialize() error {
    // ESC @ - Initialize printer
    _, err := d.conn.Write([]byte{0x1B, 0x40})
    return err
}
```

**Print Implementation:**
```go
func (d *ESCPOSDriver) Print(ctx context.Context, doc *printer.PrintDocument) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Render document to ESC/POS commands
    commands := d.renderDocument(doc)

    // Send to printer with timeout
    deadline, ok := ctx.Deadline()
    if ok {
        if conn, ok := d.conn.(net.Conn); ok {
            conn.SetWriteDeadline(deadline)
        }
    }

    n, err := d.conn.Write(commands)
    if err != nil {
        return fmt.Errorf("write failed: %w", err)
    }
    if n != len(commands) {
        return fmt.Errorf("short write: %d/%d", n, len(commands))
    }

    d.lastUsed = time.Now()
    return nil
}
```

**Command Generation:**
```go
func (d *ESCPOSDriver) renderDocument(doc *printer.PrintDocument) []byte {
    buf := &bytes.Buffer{}

    // Initialize
    buf.Write([]byte{0x1B, 0x40}) // ESC @

    // Set charset (UTF-8 if supported, else code page)
    if d.config.SupportsUTF8 {
        buf.Write([]byte{0x1B, 0x74, 0x10}) // ESC t 16 (UTF-8)
    } else {
        buf.Write([]byte{0x1B, 0x74, 0x00}) // ESC t 0 (CP437)
    }

    // Render sections
    for _, section := range doc.Sections {
        for _, line := range section.Lines {
            d.renderLine(buf, line)
        }
    }

    // Cut (if requested)
    if doc.Options.Cut {
        buf.Write([]byte{0x1D, 0x56, 0x00}) // GS V 0 (full cut)
    }

    // Drawer pulse (if requested)
    if doc.Options.DrawerPulse {
        buf.Write([]byte{0x1B, 0x70, 0x00, 0x32, 0x32}) // ESC p 0 50 50
    }

    return buf.Bytes()
}

func (d *ESCPOSDriver) renderLine(buf *bytes.Buffer, line printer.Line) {
    // Set alignment
    switch line.Alignment {
    case printer.AlignLeft:
        buf.Write([]byte{0x1B, 0x61, 0x00}) // ESC a 0
    case printer.AlignCenter:
        buf.Write([]byte{0x1B, 0x61, 0x01}) // ESC a 1
    case printer.AlignRight:
        buf.Write([]byte{0x1B, 0x61, 0x02}) // ESC a 2
    }

    // Render runs
    for _, run := range line.Runs {
        d.renderRun(buf, run)
    }

    // Line feed
    buf.WriteByte('\n')
}

func (d *ESCPOSDriver) renderRun(buf *bytes.Buffer, run printer.Run) {
    // Set style
    if run.Style.Bold {
        buf.Write([]byte{0x1B, 0x45, 0x01}) // ESC E 1
    } else {
        buf.Write([]byte{0x1B, 0x45, 0x00}) // ESC E 0
    }

    if run.Style.Underline {
        buf.Write([]byte{0x1B, 0x2D, 0x01}) // ESC - 1
    } else {
        buf.Write([]byte{0x1B, 0x2D, 0x00}) // ESC - 0
    }

    // Set size
    var sizeCmd byte = 0x00
    if run.Style.DoubleWidth {
        sizeCmd |= 0x20
    }
    if run.Style.DoubleHeight {
        sizeCmd |= 0x10
    }
    buf.Write([]byte{0x1D, 0x21, sizeCmd}) // GS ! n

    // Write text
    buf.WriteString(run.Text)
}
```

#### 6.1.3 Arabic Text Support

For Arabic text, you need text shaping (contextual forms) and bidirectional reordering.

**Option 1: Text Shaping Library**
```go
import "github.com/benoitkugler/textlayout/harfbuzz"

func (d *ESCPOSDriver) shapeArabicText(text string) string {
    // Use HarfBuzz (via CGO binding) to shape Arabic text
    // This converts isolated forms to contextual forms
    shaped := harfbuzz.Shape(text, d.font)
    return shaped
}
```

**Option 2: Bitmap Rendering**
If the printer doesn't support Arabic codepoints:
```go
func (d *ESCPOSDriver) renderArabicAsBitmap(text string) []byte {
    // Render text to image
    img := d.renderTextToImage(text)

    // Convert to monochrome bitmap
    bitmap := d.imageToBitmap(img)

    // Generate raster graphics command
    return d.bitmapToESCPOS(bitmap)
}
```

#### 6.1.4 Barcodes and QR Codes

**1D Barcode (EAN13, Code128, etc.):**
```go
func (d *ESCPOSDriver) renderBarcode(data string, symbology BarcodeSymbology) []byte {
    buf := &bytes.Buffer{}

    // GS k m d1...dk NUL
    buf.WriteByte(0x1D) // GS
    buf.WriteByte(0x6B) // k
    buf.WriteByte(byte(symbology))
    buf.WriteString(data)
    buf.WriteByte(0x00) // NUL terminator

    return buf.Bytes()
}
```

**QR Code:**
```go
func (d *ESCPOSDriver) renderQRCode(data string, size int) []byte {
    buf := &bytes.Buffer{}

    // Model: GS ( k pL pH cn fn n
    buf.Write([]byte{0x1D, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00})

    // Size: GS ( k pL pH cn fn n
    buf.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, byte(size)})

    // Store data: GS ( k pL pH cn fn m d1...dk
    dataLen := len(data) + 3
    buf.Write([]byte{0x1D, 0x28, 0x6B, byte(dataLen % 256), byte(dataLen / 256), 0x31, 0x50, 0x30})
    buf.WriteString(data)

    // Print: GS ( k pL pH cn fn m
    buf.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30})

    return buf.Bytes()
}
```

#### 6.1.5 Error Handling

```go
func (d *ESCPOSDriver) handleError(err error) {
    if netErr, ok := err.(net.Error); ok {
        if netErr.Timeout() {
            // Network timeout - mark device as degraded
            d.setStatus(devices.HealthDegraded, "Network timeout")
        } else {
            // Other network error - mark as offline
            d.setStatus(devices.HealthOffline, fmt.Sprintf("Network error: %v", err))
        }
    } else if errors.Is(err, io.EOF) {
        // Connection closed - mark as offline
        d.setStatus(devices.HealthOffline, "Connection closed")
    } else {
        // Unknown error
        d.setStatus(devices.HealthError, fmt.Sprintf("Error: %v", err))
    }

    // Attempt reconnection
    go d.reconnect()
}

func (d *ESCPOSDriver) reconnect() {
    backoff := time.Second
    maxBackoff := 60 * time.Second

    for {
        time.Sleep(backoff)

        if err := d.Connect(context.Background()); err == nil {
            d.setStatus(devices.HealthReady, "Reconnected")
            return
        }

        // Exponential backoff
        backoff *= 2
        if backoff > maxBackoff {
            backoff = maxBackoff
        }
    }
}
```

---

### 6.2 Label / ZPL Printers

#### 6.2.1 ZPL Command Generation

```go
func (d *ZPLDriver) renderLabel(label *Label) string {
    zpl := &strings.Builder{}

    // Start label
    zpl.WriteString("^XA\n")

    // Set label dimensions
    zpl.WriteString(fmt.Sprintf("^PW%d\n", label.Width))
    zpl.WriteString(fmt.Sprintf("^LL%d\n", label.Height))

    // Add fields
    for _, field := range label.Fields {
        switch field.Type {
        case FieldTypeText:
            zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", field.X, field.Y))
            zpl.WriteString(fmt.Sprintf("^A0N,%d,%d\n", field.FontHeight, field.FontWidth))
            zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", field.Data))

        case FieldTypeBarcode:
            zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", field.X, field.Y))
            zpl.WriteString(fmt.Sprintf("^BY%d\n", field.BarcodeWidth))
            zpl.WriteString(fmt.Sprintf("^BCN,%d,Y,N,N\n", field.BarcodeHeight))
            zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", field.Data))

        case FieldTypeQRCode:
            zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", field.X, field.Y))
            zpl.WriteString(fmt.Sprintf("^BQN,2,%d\n", field.QRSize))
            zpl.WriteString(fmt.Sprintf("^FDQA,%s^FS\n", field.Data))
        }
    }

    // Print quantity
    zpl.WriteString(fmt.Sprintf("^PQ%d\n", label.Quantity))

    // End label
    zpl.WriteString("^XZ\n")

    return zpl.String()
}
```

---

### 6.3 Barcode Scanners

#### 6.3.1 USB HID Raw Mode

```go
import "github.com/google/gousb"

type HIDScanner struct {
    ctx      *gousb.Context
    device   *gousb.Device
    endpoint *gousb.InEndpoint
    events   chan scanner.ScanEvent
}

func (s *HIDScanner) Start(ctx context.Context) error {
    // Open USB device
    s.ctx = gousb.NewContext()
    dev, err := s.ctx.OpenDeviceWithVIDPID(s.config.VendorID, s.config.ProductID)
    if err != nil {
        return fmt.Errorf("failed to open scanner: %w", err)
    }
    s.device = dev

    // Claim interface
    intf, done, err := dev.DefaultInterface()
    if err != nil {
        return fmt.Errorf("failed to claim interface: %w", err)
    }
    defer done()

    // Open endpoint
    ep, err := intf.InEndpoint(s.config.EndpointAddr)
    if err != nil {
        return fmt.Errorf("failed to open endpoint: %w", err)
    }
    s.endpoint = ep

    // Start reading
    go s.readLoop(ctx)

    return nil
}

func (s *HIDScanner) readLoop(ctx context.Context) {
    buffer := make([]byte, 64)
    scanData := &strings.Builder{}

    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        // Read HID report
        n, err := s.endpoint.Read(buffer)
        if err != nil {
            log.Errorf("HID read error: %v", err)
            continue
        }

        // Parse HID report
        for i := 2; i < n; i++ { // Skip report ID and modifier
            keyCode := buffer[i]
            if keyCode == 0 {
                continue
            }

            // Map HID keycode to character
            char := s.hidKeycodeToChar(keyCode)
            if char == '\n' || char == '\r' {
                // End of scan
                data := scanData.String()
                scanData.Reset()

                s.events <- scanner.ScanEvent{
                    DeviceID:  s.ID(),
                    Data:      data,
                    Symbology: "UNKNOWN", // HID mode doesn't provide symbology
                    Timestamp: time.Now(),
                    EventID:   generateEventID(),
                }
            } else if char != 0 {
                scanData.WriteRune(char)
            }
        }
    }
}

func (s *HIDScanner) hidKeycodeToChar(keycode byte) rune {
    // HID Usage ID to ASCII mapping
    // This is a simplified version
    mapping := map[byte]rune{
        0x04: 'a', 0x05: 'b', 0x06: 'c', // ... etc
        0x1E: '1', 0x1F: '2', 0x20: '3', // ... etc
        0x28: '\n', // Enter
    }
    return mapping[keycode]
}
```

#### 6.3.2 Serial Scanner

```go
type SerialScanner struct {
    port     *serial.Port
    events   chan scanner.ScanEvent
}

func (s *SerialScanner) Start(ctx context.Context) error {
    port, err := serial.Open(&serial.Config{
        Name: s.config.Port,
        Baud: s.config.BaudRate,
    })
    if err != nil {
        return err
    }
    s.port = port

    go s.readLoop(ctx)
    return nil
}

func (s *SerialScanner) readLoop(ctx context.Context) {
    reader := bufio.NewReader(s.port)

    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        // Read until terminator (usually \r\n)
        line, err := reader.ReadString('\n')
        if err != nil {
            log.Errorf("Serial read error: %v", err)
            continue
        }

        data := strings.TrimSpace(line)
        if data == "" {
            continue
        }

        s.events <- scanner.ScanEvent{
            DeviceID:  s.ID(),
            Data:      data,
            Symbology: "UNKNOWN",
            Timestamp: time.Now(),
            EventID:   generateEventID(),
        }
    }
}
```

---

### 6.4 Scales

#### 6.4.1 Scale Protocol Interface

```go
type ScaleProtocol interface {
    RequestWeight(conn io.ReadWriter) (scale.WeightReading, error)
    Zero(conn io.ReadWriter) error
    Tare(conn io.ReadWriter) error
}
```

#### 6.4.2 Dibal Scale Protocol

```go
type DibalProtocol struct{}

func (p *DibalProtocol) RequestWeight(conn io.ReadWriter) (scale.WeightReading, error) {
    // Send weight request command
    _, err := conn.Write([]byte{0x05}) // ENQ
    if err != nil {
        return scale.WeightReading{}, err
    }

    // Read response (7 bytes)
    buf := make([]byte, 7)
    n, err := io.ReadFull(conn, buf)
    if err != nil {
        return scale.WeightReading{}, err
    }
    if n != 7 {
        return scale.WeightReading{}, fmt.Errorf("invalid response length: %d", n)
    }

    // Parse response
    // Format: STX (1) + Status (1) + Weight (5) + ETX (1)
    if buf[0] != 0x02 || buf[6] != 0x03 {
        return scale.WeightReading{}, fmt.Errorf("invalid frame")
    }

    status := buf[1]
    stable := (status & 0x01) == 0

    // Parse weight (BCD encoded)
    weightStr := fmt.Sprintf("%c%c%c%c%c", buf[2], buf[3], buf[4], buf[5], buf[6])
    weight, err := strconv.ParseFloat(weightStr, 64)
    if err != nil {
        return scale.WeightReading{}, err
    }

    return scale.WeightReading{
        Weight:    weight / 1000.0, // Convert to kg
        Unit:      scale.UnitKilogram,
        Stable:    stable,
        Timestamp: time.Now(),
    }, nil
}
```

#### 6.4.3 Mettler Toledo (MT-SICS) Protocol

```go
type MettlerProtocol struct{}

func (p *MettlerProtocol) RequestWeight(conn io.ReadWriter) (scale.WeightReading, error) {
    // Send "S" command (send stable weight)
    _, err := conn.Write([]byte("S\r\n"))
    if err != nil {
        return scale.WeightReading{}, err
    }

    // Read response
    reader := bufio.NewReader(conn)
    line, err := reader.ReadString('\n')
    if err != nil {
        return scale.WeightReading{}, err
    }

    // Parse response
    // Format: "S S     1.234 kg\r\n" (stable)
    //         "S D     1.234 kg\r\n" (dynamic)
    //         "S I\r\n" (invalid)

    parts := strings.Fields(line)
    if len(parts) < 4 {
        return scale.WeightReading{}, fmt.Errorf("invalid response: %s", line)
    }

    if parts[0] != "S" {
        return scale.WeightReading{}, fmt.Errorf("unexpected response: %s", line)
    }

    stable := parts[1] == "S"

    weight, err := strconv.ParseFloat(parts[2], 64)
    if err != nil {
        return scale.WeightReading{}, err
    }

    unit := scale.UnitKilogram
    if parts[3] == "lb" {
        unit = scale.UnitPound
    }

    return scale.WeightReading{
        Weight:    weight,
        Unit:      unit,
        Stable:    stable,
        Timestamp: time.Now(),
    }, nil
}

func (p *MettlerProtocol) Zero(conn io.ReadWriter) error {
    _, err := conn.Write([]byte("Z\r\n"))
    return err
}

func (p *MettlerProtocol) Tare(conn io.ReadWriter) error {
    _, err := conn.Write([]byte("T\r\n"))
    return err
}
```

#### 6.4.4 CAS Scale Protocol

```go
type CASProtocol struct{}

func (p *CASProtocol) RequestWeight(conn io.ReadWriter) (scale.WeightReading, error) {
    // CAS scales continuously send weight data
    // Read 16 bytes
    buf := make([]byte, 16)
    n, err := io.ReadFull(conn, buf)
    if err != nil {
        return scale.WeightReading{}, err
    }
    if n != 16 {
        return scale.WeightReading{}, fmt.Errorf("invalid response length: %d", n)
    }

    // Parse response
    // Byte 0: Status
    stable := (buf[0] & 0x20) != 0

    // Bytes 1-6: Weight (ASCII)
    weightStr := string(buf[1:7])
    weight, err := strconv.ParseFloat(strings.TrimSpace(weightStr), 64)
    if err != nil {
        return scale.WeightReading{}, err
    }

    // Determine unit from status byte
    unit := scale.UnitKilogram
    if (buf[0] & 0x01) != 0 {
        unit = scale.UnitPound
    }

    return scale.WeightReading{
        Weight:    weight,
        Unit:      unit,
        Stable:    stable,
        Timestamp: time.Now(),
    }, nil
}
```

---

### 6.5 Customer Display

#### 6.5.1 Serial Display Driver

```go
type SerialDisplay struct {
    port *serial.Port
}

func (d *SerialDisplay) ShowLines(ctx context.Context, lines []string, duration time.Duration) error {
    // Clear display first
    if err := d.Clear(ctx); err != nil {
        return err
    }

    // Send lines
    for i, line := range lines {
        if i >= 2 {
            break // Only 2 lines supported
        }

        // Position cursor
        cmd := []byte{0x1B, 0x6C, byte(i), 0x00} // ESC l row 0
        if _, err := d.port.Write(cmd); err != nil {
            return err
        }

        // Truncate/pad to 20 characters
        line = d.formatLine(line, 20)

        // Write text
        if _, err := d.port.Write([]byte(line)); err != nil {
            return err
        }
    }

    // Schedule auto-clear if duration > 0
    if duration > 0 {
        time.AfterFunc(duration, func() {
            d.Clear(context.Background())
        })
    }

    return nil
}

func (d *SerialDisplay) Clear(ctx context.Context) error {
    // Clear screen command
    cmd := []byte{0x0C} // Form feed
    _, err := d.port.Write(cmd)
    return err
}

func (d *SerialDisplay) formatLine(text string, width int) string {
    // Remove non-ASCII characters
    cleaned := strings.Map(func(r rune) rune {
        if r < 32 || r > 126 {
            return -1
        }
        return r
    }, text)

    // Truncate or pad
    if len(cleaned) > width {
        return cleaned[:width]
    }
    return cleaned + strings.Repeat(" ", width-len(cleaned))
}
```

---

### 6.6 Cash Drawer

#### 6.6.1 Drawer via Printer

```go
type PrinterDrawer struct {
    printer printer.Printer
}

func (d *PrinterDrawer) Open(ctx context.Context) error {
    // Delegate to printer's OpenDrawer method
    return d.printer.OpenDrawer(ctx)
}

// In ESC/POS driver:
func (p *ESCPOSDriver) OpenDrawer(ctx context.Context) error {
    // ESC p m t1 t2
    // m = connector (0 or 1)
    // t1 = ON time (units of 2ms)
    // t2 = OFF time (units of 2ms)
    cmd := []byte{0x1B, 0x70, 0x00, 0x32, 0x32} // 100ms pulse

    _, err := p.conn.Write(cmd)
    return err
}
```

---

### 6.7 Payment Terminals

#### 6.7.1 Payment Driver Architecture

```go
type PaymentDriver struct {
    config   PaymentConfig
    conn     net.Conn
    payments map[payment.PaymentID]*PaymentState
    mu       sync.RWMutex
}

type PaymentState struct {
    ID        payment.PaymentID
    Status    payment.PaymentStatus
    Events    chan payment.PaymentEvent
    CancelCtx context.CancelFunc
}
```

#### 6.7.2 Example: Mada Integration (Saudi Arabia)

This is a simplified example. Real integration requires provider SDK/documentation.

```go
func (d *MadaDriver) StartPayment(ctx context.Context, req payment.PaymentRequest) (payment.PaymentID, error) {
    // Generate payment ID
    paymentID := payment.PaymentID(generateUUID())

    // Create payment state
    state := &PaymentState{
        ID:     paymentID,
        Events: make(chan payment.PaymentEvent, 10),
    }
    ctx, cancel := context.WithCancel(ctx)
    state.CancelCtx = cancel

    d.mu.Lock()
    d.payments[paymentID] = state
    d.mu.Unlock()

    // Start payment flow in background
    go d.processPayment(ctx, paymentID, req, state)

    return paymentID, nil
}

func (d *MadaDriver) processPayment(ctx context.Context, id payment.PaymentID, req payment.PaymentRequest, state *PaymentState) {
    // Send payment request to terminal
    msg := d.buildPaymentMessage(req)
    if err := d.sendMessage(msg); err != nil {
        state.Events <- payment.PaymentEvent{
            PaymentID: id,
            Event:     "error",
            Status:    payment.PaymentError,
            Message:   err.Error(),
            Timestamp: time.Now(),
        }
        return
    }

    state.Events <- payment.PaymentEvent{
        PaymentID: id,
        Event:     "initiated",
        Status:    payment.PaymentInitiated,
        Timestamp: time.Now(),
    }

    // Poll terminal for updates
    timeout := time.After(req.Timeout)
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // Cancelled
            d.cancelOnTerminal(id)
            state.Events <- payment.PaymentEvent{
                PaymentID: id,
                Event:     "cancelled",
                Status:    payment.PaymentCancelled,
                Timestamp: time.Now(),
            }
            return

        case <-timeout:
            // Timeout
            state.Events <- payment.PaymentEvent{
                PaymentID: id,
                Event:     "timeout",
                Status:    payment.PaymentTimeout,
                Timestamp: time.Now(),
            }
            return

        case <-ticker.C:
            // Query status
            status, err := d.queryTerminal(id)
            if err != nil {
                log.Errorf("Failed to query terminal: %v", err)
                continue
            }

            // Parse status and emit events
            if status.Final {
                finalEvent := payment.PaymentEvent{
                    PaymentID: id,
                    Timestamp: time.Now(),
                }

                if status.Approved {
                    finalEvent.Event = "approved"
                    finalEvent.Status = payment.PaymentApproved
                } else {
                    finalEvent.Event = "declined"
                    finalEvent.Status = payment.PaymentDeclined
                }

                state.Events <- finalEvent
                return
            } else {
                // Intermediate event
                state.Events <- payment.PaymentEvent{
                    PaymentID: id,
                    Event:     status.Event,
                    Status:    payment.PaymentInProgress,
                    Message:   status.Message,
                    Timestamp: time.Now(),
                }
            }
        }
    }
}

func (d *MadaDriver) buildPaymentMessage(req payment.PaymentRequest) []byte {
    // Build protocol-specific message
    // This is highly vendor-specific
    msg := &bytes.Buffer{}

    // Example: ISO 8583-like message
    // (Actual format depends on provider)

    return msg.Bytes()
}

func (d *MadaDriver) queryTerminal(id payment.PaymentID) (*TerminalStatus, error) {
    // Send status query to terminal
    // Parse response

    return &TerminalStatus{}, nil
}
```

---

## 7. Implementation Roadmap

This section provides a **clean, phased approach** to implementing Device Bridge v2.

---

### Step 1: Skeleton & Infrastructure (Week 1)

**Goal:** Get a runnable service with no real device support yet.

**Tasks:**
1. Create repository structure (as in section 5.1)
2. Initialize Go modules, add initial dependencies:
   - `google.golang.org/grpc`
   - `github.com/spf13/viper` (config)
   - `go.uber.org/zap` (logging)
   - `github.com/prometheus/client_golang` (metrics)
3. Implement config loader (YAML support)
4. Implement structured logger
5. Implement basic Prometheus metrics
6. Create gRPC server skeleton (no services yet)
7. Add `Ping` API to verify everything works
8. Write `Dockerfile` and `docker-compose.yaml`
9. Write `Makefile` for build/test/run

**Deliverables:**
- Running service on `localhost:50051`
- `grpcurl localhost:50051 devicebridge.v1.DeviceBridge/Ping` works
- Health check endpoint works
- Prometheus metrics endpoint works

---

### Step 2: ESC/POS Printer (Single Device) (Week 2-3)

**Goal:** Print a simple receipt on one real TCP printer.

**Tasks:**
1. Define `Printer` interface (section 5.2.2)
2. Implement `PrintDocument` model
3. Implement `ESCPOSDriver` for TCP transport
4. Implement basic rendering (text, alignment, bold)
5. Add `Print` gRPC API
6. Test on real hardware until rock solid

**Deliverables:**
- Print a simple ASCII receipt on a real printer
- Handle connection errors gracefully
- Log all operations

---

### Step 3: Device Registry & Static Config (Week 3-4)

**Goal:** Support multiple devices from config.

**Tasks:**
1. Implement `Registry` (section 5.3)
2. Add YAML device configuration
3. Load devices from config on startup
4. Add `ListDevices` API
5. Add `GetDevice` API
6. Add device health monitoring

**Deliverables:**
- Configure 2-3 printers in YAML
- `ListDevices` returns all configured devices
- Device health is tracked and exposed

---

### Step 4: Job Scheduler (Week 4-5)

**Goal:** Asynchronous job processing.

**Tasks:**
1. Implement `Job` model (section 3.2)
2. Implement job queue (channel-based)
3. Implement worker pool
4. Add job history storage
5. Add idempotency support
6. Add job status APIs

**Deliverables:**
- Print jobs are processed asynchronously
- Job history is queryable
- Duplicate jobs are handled correctly

---

### Step 5: Scanner Support (Week 5-6)

**Goal:** Stream barcode scans to clients.

**Tasks:**
1. Define `Scanner` interface (section 5.2.3)
2. Implement USB HID scanner driver
3. Implement serial scanner driver
4. Add gRPC streaming API (`SubscribeScanner`)
5. Add WebSocket endpoint for browsers
6. Test with real scanners

**Deliverables:**
- Scan barcodes and stream events via gRPC
- WebSocket endpoint works in browser
- Support at least 2 scanner models

---

### Step 6: Scales (Week 6-7)

**Goal:** Read weight from real scales.

**Tasks:**
1. Define `Scale` interface (section 5.2.4)
2. Implement 2-3 scale protocols (Dibal, Mettler, CAS)
3. Add `GetWeight` API
4. Add stable weight detection
5. Test with real scales

**Deliverables:**
- Read stable weight from real scale
- Handle unstable readings correctly
- Support at least 2 scale models

---

### Step 7: Customer Display & Cash Drawer (Week 7-8)

**Goal:** Control customer display and cash drawer.

**Tasks:**
1. Define `Display` interface (section 5.2.5)
2. Implement serial display driver
3. Add `ShowLines` and `Clear` APIs
4. Define `Drawer` interface (section 5.2.6)
5. Implement drawer control via printer
6. Add `OpenDrawer` API

**Deliverables:**
- Show text on customer display
- Open cash drawer via printer
- Handle errors gracefully

---

### Step 8: Payment Provider (Week 8-10)

**Goal:** Process payments on real terminal.

**Tasks:**
1. Define `PaymentProvider` interface (section 5.2.7)
2. Pick one real provider (Mada, KNET, etc.)
3. Implement provider driver
4. Add payment APIs (`StartPayment`, `CancelPayment`, `Status`)
5. Add payment event streaming
6. Test on real terminal

**Deliverables:**
- Process a real payment end-to-end
- Stream payment events to client
- Handle all error cases
- PCI-DSS compliant logging

---

### Step 9: REST/JSON Gateway (Week 10-11)

**Goal:** Expose REST APIs for non-gRPC clients.

**Tasks:**
1. Add `grpc-gateway` (protoc plugin)
2. Generate REST/JSON gateway
3. Add OpenAPI/Swagger docs
4. Test all APIs via HTTP

**Deliverables:**
- All gRPC APIs available via REST
- Swagger UI for API exploration
- Browser-friendly APIs

---

### Step 10: Advanced Features (Week 11-12)

**Goal:** Polish and advanced features.

**Tasks:**
1. Add Arabic text support for printers
2. Add QR/barcode generation
3. Add print templates
4. Add label printer support (ZPL)
5. Add auto-discovery (USB, TCP, mDNS)
6. Add ACL/security

**Deliverables:**
- Print Arabic receipts
- Print QR codes
- Use templates for receipts
- Print labels (ZPL)
- Auto-discover devices
- Secure production deployment

---

### Step 11: Virtual Devices & Testing (Week 12-13)

**Goal:** Support offline development and CI.

**Tasks:**
1. Implement virtual printer
2. Implement virtual scanner
3. Implement virtual scale
4. Write integration tests using virtual devices
5. Add CI pipeline (GitHub Actions)

**Deliverables:**
- Full test suite with virtual devices
- CI runs on every commit
- Code coverage > 70%

---

### Step 12: Documentation & Release (Week 13-14)

**Goal:** Production-ready release.

**Tasks:**
1. Write deployment guides
2. Write API documentation
3. Create example clients (Go, Python, JavaScript)
4. Create Helm chart (if Kubernetes support needed)
5. Write troubleshooting guide
6. Tag v2.0.0 release

**Deliverables:**
- Complete documentation
- Example code for clients
- Production deployment guide
- v2.0.0 release

---

## Conclusion

This specification defines Device Bridge v2 as a **production-grade hardware abstraction layer** for POS and retail devices.

**Key Principles:**
1. **Real hardware, not stubs** - Every device type has real, tested implementations
2. **Clean architecture** - Clear separation between interfaces, drivers, and business logic
3. **Observability first** - Logging, metrics, and tracing built-in from day 1
4. **Fault tolerance** - Graceful error handling and automatic recovery
5. **Phased approach** - Start small, iterate, grow

**Success Criteria:**
- ✅ Supports 8+ device types with real hardware
- ✅ Handles 20+ print jobs/minute per lane
- ✅ Runs for weeks without restart
- ✅ Clear APIs (gRPC, REST, WebSocket)
- ✅ Production-ready security (mTLS, ACL)
- ✅ Complete observability (logs, metrics, traces)
- ✅ Comprehensive documentation
- ✅ Virtual devices for offline development

---

**Version History:**

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-11-10 | Initial specification |

---

**Maintainers:**
- Device Bridge Team

**License:**
- TBD

---

*End of Document*
