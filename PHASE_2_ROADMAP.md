# Device Bridge v2 - Phase 2: Enhancement Roadmap

**Date:** 2025-11-10
**Status:** 🚧 In Progress
**Target:** Production-ready with REST, WebSocket, more devices, and security

---

## Phase 2 Overview

Phase 2 builds on the solid MVP foundation to add:
- Multi-protocol API support (REST, WebSocket)
- Additional device drivers (scanners, scales, displays)
- Auto-discovery capabilities
- Security features (mTLS, ACL)
- Advanced printer features
- CLI administration tool
- Comprehensive testing

---

## Implementation Plan

### Priority 1: API Extensions (Week 1-2)

#### 1.1 REST/JSON Gateway
**Goal:** Expose all gRPC APIs via REST/JSON for browser and HTTP clients

**Tasks:**
- [ ] Add grpc-gateway dependencies
- [ ] Update proto files with REST annotations
- [ ] Generate REST gateway code
- [ ] Integrate gateway into main daemon
- [ ] Add OpenAPI/Swagger documentation
- [ ] Test all endpoints with curl
- [ ] Add CORS support

**Files:**
- `proto/devicebridge/v1/service.proto` - Add HTTP annotations (already has some)
- `internal/api/rest/gateway.go` - REST gateway server
- `internal/api/rest/middleware.go` - CORS, logging middleware
- `cmd/bridge/main.go` - Wire REST gateway
- `docs/api/swagger.yaml` - Generated OpenAPI spec

**Deliverables:**
- All gRPC APIs accessible via HTTP/JSON at `http://localhost:8080/v1/`
- Swagger UI at `http://localhost:8080/swagger/`
- CORS enabled for browser clients

---

#### 1.2 WebSocket Server
**Goal:** Real-time event streaming for browser clients

**Tasks:**
- [ ] Create WebSocket server
- [ ] Add scanner event streaming endpoint
- [ ] Add payment event streaming endpoint
- [ ] Add device status streaming endpoint
- [ ] Implement connection management
- [ ] Add authentication support
- [ ] Test with browser clients

**Files:**
- `internal/api/ws/server.go` - WebSocket server
- `internal/api/ws/handlers.go` - Event handlers
- `internal/api/ws/client.go` - Client connection management
- `test/client/scanner.html` - Browser test client
- `test/client/payment.html` - Browser test client

**Endpoints:**
- `ws://localhost:8080/ws/scanner/:device_id` - Scanner events
- `ws://localhost:8080/ws/payment/:device_id` - Payment events
- `ws://localhost:8080/ws/devices` - Device status changes

**Deliverables:**
- Real-time scanner events in browser
- Real-time payment updates in browser
- Connection pooling and cleanup
- Heartbeat/ping-pong support

---

### Priority 2: Device Drivers (Week 3-5)

#### 2.1 Scanner Drivers
**Goal:** Support USB HID and Serial barcode scanners

**Tasks:**
- [ ] Implement USB HID scanner driver
  - USB device enumeration
  - HID input parsing
  - Multiple scanners support
- [ ] Implement Serial scanner driver
  - Serial port communication
  - Protocol support (Honeywell, Zebra, etc.)
  - Configurable baud rates
- [ ] Test with real scanners
- [ ] Add scanner to device registry
- [ ] Update documentation

**Files:**
- `internal/drivers/scanner_hid/driver.go` - HID scanner driver
- `internal/drivers/scanner_serial/driver.go` - Serial scanner driver
- `test/virtual_devices/virtual_scanner.go` - Virtual scanner for testing

**Supported Scanners:**
- Honeywell Voyager (USB HID)
- Zebra DS2208 (USB HID)
- Generic Serial scanners

**Deliverables:**
- Stream barcode scans via gRPC, WebSocket, REST
- Support at least 2 real scanner models
- Virtual scanner for testing

---

#### 2.2 Scale Drivers
**Goal:** Support weighing scales with multiple protocols

**Tasks:**
- [ ] Implement serial scale driver base
- [ ] Add Dibal protocol support
- [ ] Add Mettler Toledo protocol support
- [ ] Add CAS protocol support
- [ ] Implement stable weight detection
- [ ] Add zero/tare operations
- [ ] Test with real scales
- [ ] Add virtual scale for testing

**Files:**
- `internal/drivers/scale_serial/driver.go` - Serial scale driver
- `internal/drivers/scale_serial/protocols/dibal.go` - Dibal protocol
- `internal/drivers/scale_serial/protocols/mettler.go` - Mettler protocol
- `internal/drivers/scale_serial/protocols/cas.go` - CAS protocol
- `test/virtual_devices/virtual_scale.go` - Virtual scale

**Protocols:**
- Dibal (Spain) - Common in retail
- Mettler Toledo - High precision
- CAS - Asian markets

**Deliverables:**
- Read stable weight from real scales
- Support kg, g, lb, oz units
- Zero and tare operations
- Stable weight detection algorithm

---

#### 2.3 Display and Drawer Drivers
**Goal:** Control customer displays and cash drawers

**Tasks:**
- [ ] Implement serial display driver
- [ ] Add LCD 2x20 support
- [ ] Add VFD display support
- [ ] Implement drawer driver (via printer)
- [ ] Test with real hardware
- [ ] Add virtual devices for testing

**Files:**
- `internal/drivers/display_serial/driver.go` - Serial display driver
- `internal/drivers/drawer/driver.go` - Cash drawer driver
- `test/virtual_devices/virtual_display.go` - Virtual display
- `test/virtual_devices/virtual_drawer.go` - Virtual drawer

**Deliverables:**
- Show text on customer displays
- Clear display operation
- Open cash drawer via printer
- Status monitoring

---

### Priority 3: Auto-Discovery (Week 6)

#### 3.1 Device Discovery System
**Goal:** Automatically discover devices on USB, Serial, TCP, mDNS

**Tasks:**
- [ ] Create discovery framework
- [ ] Implement USB device discovery
- [ ] Implement Serial port scanning
- [ ] Implement TCP network scanning
- [ ] Implement mDNS service discovery
- [ ] Add background discovery worker
- [ ] Add discovery configuration
- [ ] Test discovery with real devices

**Files:**
- `internal/discovery/discovery.go` - Discovery manager
- `internal/discovery/usb.go` - USB enumeration
- `internal/discovery/serial.go` - Serial port scanning
- `internal/discovery/tcp.go` - TCP network scanning
- `internal/discovery/mdns.go` - mDNS/Bonjour discovery

**Discovery Methods:**
- **USB:** Enumerate USB devices, match vendor/product IDs
- **Serial:** Scan COM/tty ports, probe with commands
- **TCP:** Scan common ports (9100 for printers)
- **mDNS:** Discover network devices advertising services

**Deliverables:**
- Auto-discover USB printers/scanners
- Auto-discover network printers
- Background scanning every 30 seconds
- Add discovered devices to registry

---

### Priority 4: Security (Week 7)

#### 4.1 Security Features
**Goal:** Production-ready security (mTLS, ACL, authentication)

**Tasks:**
- [ ] Implement mTLS support
  - Server certificates
  - Client certificates
  - Certificate validation
- [ ] Implement ACL system
  - Device access control
  - Operation permissions
  - Role-based access
- [ ] Implement API authentication
  - Token-based auth
  - API key support
  - Session management
- [ ] Add audit logging
- [ ] Add PCI-DSS compliance features
- [ ] Security documentation

**Files:**
- `internal/security/tls.go` - TLS configuration and management
- `internal/security/acl.go` - Access control lists
- `internal/security/auth.go` - Authentication
- `internal/security/audit.go` - Audit logging
- `configs/security.example.yaml` - Security configuration example

**Features:**
- mTLS for gRPC connections
- ACL rules per device/operation
- API key authentication
- Audit trail for sensitive operations
- No sensitive data in logs (PCI-DSS)

**Deliverables:**
- Secure production deployment
- ACL configuration
- Certificate management guide
- Security best practices doc

---

### Priority 5: Advanced Features (Week 8-9)

#### 5.1 Label Printer Support (ZPL)
**Goal:** Support Zebra ZPL label printers

**Tasks:**
- [ ] Define label document model
- [ ] Implement ZPL renderer
- [ ] Add TCP transport for ZPL printers
- [ ] Add barcode/QR code support
- [ ] Test with real Zebra printer
- [ ] Documentation and examples

**Files:**
- `internal/devices/printer/label.go` - Label document model
- `internal/drivers/printer_zpl/driver.go` - ZPL printer driver
- `internal/drivers/printer_zpl/renderer.go` - ZPL command generation

**Deliverables:**
- Print shipping labels
- Print product labels
- Barcode labels (Code 128, EAN, etc.)
- Support Zebra ZT/GK/GX series

---

#### 5.2 Print Templates
**Goal:** Template-based receipt generation

**Tasks:**
- [ ] Design template format (Go templates)
- [ ] Implement template engine
- [ ] Add template registry
- [ ] Create example templates
- [ ] Add template validation
- [ ] Documentation

**Files:**
- `internal/templates/engine.go` - Template engine
- `internal/templates/registry.go` - Template registry
- `templates/receipt.tmpl` - Example receipt template
- `templates/invoice.tmpl` - Example invoice template

**Features:**
- Dynamic data injection
- Conditional sections
- Loops for items
- Custom functions (formatting, etc.)
- Template validation

**Deliverables:**
- Load templates from files
- Render receipts from data
- Template library (receipt, invoice, label)

---

#### 5.3 Arabic Text Support
**Goal:** Print Arabic text on receipts

**Tasks:**
- [ ] Add Arabic text shaping library
- [ ] Implement bidirectional text support
- [ ] Add Arabic-capable font support
- [ ] Test with real printers
- [ ] Documentation and examples

**Files:**
- `internal/drivers/printer_escpos/arabic.go` - Arabic text handling
- `internal/fonts/arabic.go` - Arabic font data

**Deliverables:**
- Print Arabic receipts
- Bidirectional text (Arabic + English)
- Proper character shaping

---

### Priority 6: CLI Tool (Week 10)

#### 6.1 Administration CLI
**Goal:** Command-line tool for administration

**Tasks:**
- [ ] Create CLI structure with cobra
- [ ] Add device management commands
- [ ] Add job management commands
- [ ] Add discovery commands
- [ ] Add config validation command
- [ ] Add testing commands
- [ ] Documentation

**Files:**
- `cmd/bridge-cli/main.go` - CLI entry point
- `cmd/bridge-cli/cmd/devices.go` - Device commands
- `cmd/bridge-cli/cmd/jobs.go` - Job commands
- `cmd/bridge-cli/cmd/config.go` - Config commands

**Commands:**
```bash
bridge-cli devices list
bridge-cli devices get <device-id>
bridge-cli devices discover
bridge-cli jobs list
bridge-cli jobs get <job-id>
bridge-cli config validate
bridge-cli test print <device-id>
bridge-cli test scan <device-id>
```

**Deliverables:**
- Full CLI tool
- Interactive device discovery
- Job monitoring
- Configuration testing

---

### Priority 7: Testing & CI (Week 11-12)

#### 7.1 Comprehensive Test Suite
**Goal:** High test coverage and CI pipeline

**Tasks:**
- [ ] Write unit tests for all packages
- [ ] Create integration tests
- [ ] Add virtual devices for all types
- [ ] Set up GitHub Actions CI
- [ ] Add test coverage reporting
- [ ] Add benchmarks
- [ ] Performance tests

**Files:**
- `internal/*/\*_test.go` - Unit tests
- `test/integration/\*_test.go` - Integration tests
- `test/virtual_devices/\*.go` - All virtual devices
- `.github/workflows/ci.yaml` - CI pipeline

**Tests:**
- Unit tests (target: >80% coverage)
- Integration tests with virtual devices
- Performance benchmarks
- Load tests (concurrent operations)

**CI Pipeline:**
- Run tests on every commit
- Build binaries for Linux/macOS/Windows
- Docker image build
- Security scanning
- Linting (golangci-lint)

**Deliverables:**
- Comprehensive test suite
- CI runs on every PR
- Code coverage >80%
- Performance benchmarks

---

### Priority 8: Documentation (Week 13)

#### 8.1 Complete Documentation
**Goal:** Production-ready documentation

**Tasks:**
- [ ] API reference documentation
- [ ] Deployment guides
- [ ] Hardware compatibility matrix
- [ ] Troubleshooting guide
- [ ] Example clients (Go, Python, JavaScript)
- [ ] Performance tuning guide
- [ ] Security hardening guide

**Files:**
- `docs/api/README.md` - API reference
- `docs/deployment/README.md` - Deployment guides
- `docs/deployment/docker.md` - Docker deployment
- `docs/deployment/kubernetes.md` - Kubernetes deployment
- `docs/deployment/systemd.md` - Systemd service
- `docs/hardware/compatibility.md` - Hardware matrix
- `docs/troubleshooting.md` - Troubleshooting
- `docs/security.md` - Security guide
- `examples/client/go/` - Go client example
- `examples/client/python/` - Python client example
- `examples/client/javascript/` - JavaScript client example

**Deliverables:**
- Complete API documentation
- Step-by-step deployment guides
- Hardware compatibility list
- Example client code
- Troubleshooting runbook

---

## Phase 2 Success Criteria

- ✅ REST/JSON API fully functional
- ✅ WebSocket streaming works
- ✅ Scanners working (HID + Serial)
- ✅ Scales working (3 protocols)
- ✅ Auto-discovery operational
- ✅ Security features enabled
- ✅ CLI tool complete
- ✅ Test coverage >80%
- ✅ CI pipeline working
- ✅ Complete documentation
- ✅ Example clients for 3 languages

---

## Timeline

| Week | Focus | Deliverables |
|------|-------|-------------|
| 1-2 | API Extensions | REST gateway, WebSocket server |
| 3-5 | Device Drivers | Scanners, Scales, Displays |
| 6 | Auto-Discovery | USB, Serial, TCP, mDNS discovery |
| 7 | Security | mTLS, ACL, Authentication |
| 8-9 | Advanced Features | ZPL, Templates, Arabic text |
| 10 | CLI Tool | Administration CLI |
| 11-12 | Testing & CI | Tests, CI pipeline, benchmarks |
| 13 | Documentation | API docs, guides, examples |

**Total Duration:** ~13 weeks (3 months)

---

## Progress Tracking

**Current Status:** 0% (Just started)

Track progress in [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)

---

## Next Session

Start with **Priority 1.1: REST/JSON Gateway** - this provides immediate value by opening the API to more clients.

**Immediate Next Steps:**
1. Update proto files with HTTP annotations (if needed)
2. Add grpc-gateway dependencies
3. Generate REST gateway code
4. Implement REST server
5. Add to main daemon
6. Test with curl
7. Add Swagger UI

---

*Let's build a production-ready device bridge!*
