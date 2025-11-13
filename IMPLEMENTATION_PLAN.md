# Implementation Plan: Complete Acceptance Checklist

**Based on**: ACCEPTANCE_CHECKLIST_VALIDATION.md
**Date**: 2025-11-13
**Target**: Production-Ready Device Bridge for Middle East POS Market
**Current Status**: 65% Complete → Target: 95% Complete

---

## Executive Summary

This plan addresses the 35% gap identified in the acceptance checklist validation. Work is organized into 4 phases over 8-10 weeks:

- **Phase 1 (Week 1-3)**: Critical Blockers - Arabic printing, GetWeight fix, hardware tests
- **Phase 2 (Week 4-5)**: Transport Layer - USB/Serial support for ESC/POS
- **Phase 3 (Week 6-7)**: POS Features - Customer display, receipt history, end-to-end flows
- **Phase 4 (Week 8-10)**: Production Readiness - Load testing, CI/CD, documentation cleanup

**Critical Path**: Arabic printing (3 weeks) is the longest task and blocks Middle East deployment.

---

## Phase 1: Critical Blockers (Weeks 1-3)

### Priority: P0 - MUST COMPLETE BEFORE PRODUCTION

### 1.1 Arabic ESC/POS Printing (3 weeks, P0)

**Problem**: ESC/POS printer cannot render Arabic text properly. Code comment admits "TODO: Arabic text shaping".

**Impact**: Cannot deploy in Saudi Arabia, Kuwait, UAE, Egypt, or any Arabic-speaking market.

**Implementation Tasks**:

#### Week 1: Research & Prototype
- [ ] **Task 1.1.1**: Evaluate Arabic text shaping libraries
  - Options: `github.com/go-text/typesetting`, `github.com/benoitkugler/textlayout`, CGO with HarfBuzz
  - Criteria: ESC/POS compatibility, Go integration, license (prefer MIT/Apache)
  - Deliverable: Technical decision document
  - Effort: 2 days

- [ ] **Task 1.1.2**: Create Arabic text rendering prototype
  - Test cases:
    - "مرحباً" (Hello)
    - "كولا Coca-Cola 12.50" (mixed Arabic/English/numbers)
    - "إجمالي الفاتورة: ٢٥٫٥٠ ر.س" (invoice total with Eastern Arabic numerals)
    - Long text: "محل البقالة الكبير للمواد الغذائية والمشروبات" (wrapping test)
  - Deliverable: Prototype that prints correct glyph shapes
  - Effort: 3 days

#### Week 2: Integration
- [ ] **Task 1.1.3**: Integrate shaping library into `internal/drivers/printer_escpos/renderer.go`
  - Modify `renderText()` function (line 89-120)
  - Add text analysis: detect Arabic vs Latin vs mixed
  - Call shaping engine for Arabic runs
  - Preserve ESC/POS command generation
  - Effort: 3 days

- [ ] **Task 1.1.4**: Implement BiDi (bidirectional text) algorithm
  - Support mixed LTR/RTL runs
  - Line wrapping that respects character boundaries
  - Alignment adjustments for RTL
  - Reference: Unicode BiDi algorithm UAX#9
  - Effort: 2 days

#### Week 3: Testing & Validation
- [ ] **Task 1.1.5**: Unit tests for Arabic rendering
  - Test all Arabic letter forms (initial, medial, final, isolated)
  - Test ligatures (لا)
  - Test diacritics (harakat)
  - Test mixed scripts
  - Test alignment (right-align for Arabic money columns)
  - Deliverable: `renderer_arabic_test.go` with 20+ test cases
  - Effort: 2 days

- [ ] **Task 1.1.6**: Real printer testing
  - Test with 3 printer models:
    - Epson TM-T88 series
    - Cheaper Chinese brand (e.g., Xprinter)
    - Star Micronics TSP100
  - Print sample receipts in Arabic
  - Validate readability with native Arabic speakers
  - Deliverable: Photos of printed receipts, test report
  - Effort: 2 days

- [ ] **Task 1.1.7**: Performance testing
  - Benchmark Arabic vs English rendering
  - Ensure <200ms print job latency maintained
  - Optimize hot paths if needed
  - Effort: 1 day

**Files to Modify**:
- `internal/drivers/printer_escpos/renderer.go:89-120` - Add shaping
- `internal/drivers/printer_escpos/arabic.go` - New file for Arabic logic
- `internal/drivers/printer_escpos/renderer_arabic_test.go` - New test file
- `go.mod` - Add text shaping dependency

**Acceptance Criteria**:
- ✅ Arabic text prints with correct glyph shaping
- ✅ Mixed Arabic/English lines print correctly
- ✅ Long Arabic text wraps without breaking characters
- ✅ Right-to-left alignment works
- ✅ Store names in Arabic print clearly
- ✅ Tested on 3 different printer models
- ✅ Performance <200ms maintained

---

### 1.2 Fix Scale GetWeight Placeholder (2 days, P0)

**Problem**: `GetWeight` gRPC method returns hardcoded `0.0` instead of reading from scale driver.

**Impact**: Cannot sell weighed items (fruits, meat, deli, bulk goods).

**Implementation Tasks**:

- [ ] **Task 1.2.1**: Wire up scale driver to gRPC server (1 day)
  - Modify `internal/api/grpc/server.go:220-230`
  - Look up device by ID from device manager
  - Call driver's `GetWeight()` method
  - Handle errors (device not found, device offline, unstable weight)
  - Return proper units (kg, g, lb, oz)
  - Example implementation:
    ```go
    func (s *Server) GetWeight(ctx context.Context, req *pb.GetWeightRequest) (*pb.GetWeightResponse, error) {
        device, err := s.deviceManager.GetDevice(req.DeviceId)
        if err != nil {
            return nil, status.Errorf(codes.NotFound, "device not found: %v", err)
        }

        scaleDriver, ok := device.(*scale_serial.Driver)
        if !ok {
            return nil, status.Errorf(codes.InvalidArgument, "device is not a scale")
        }

        weight, unit, stable, err := scaleDriver.GetWeight(ctx)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to read weight: %v", err)
        }

        return &pb.GetWeightResponse{
            Weight: weight,
            Unit:   unit,
            Stable: stable,
        }, nil
    }
    ```

- [ ] **Task 1.2.2**: Add integration tests (1 day)
  - Test with virtual scale device
  - Test with real scale (Mettler Toledo if available)
  - Test error cases:
    - Device not found
    - Device is not a scale
    - Device offline
    - Weight unstable
  - Deliverable: `internal/api/grpc/server_scale_test.go`

- [ ] **Task 1.2.3**: Test REST and WebSocket endpoints
  - Verify REST `/v1/scales/{id}/weight` works
  - Consider WebSocket for live weight updates
  - Test from Flutter example app

**Files to Modify**:
- `internal/api/grpc/server.go:220-230` - Fix GetWeight
- `internal/api/grpc/server_test.go` - Add tests
- `api/proto/v1/device.proto` - Verify Weight message has `stable` field

**Acceptance Criteria**:
- ✅ GetWeight returns real values from scale driver
- ✅ Correct units returned (kg, g, lb, oz per scale config)
- ✅ Stable flag indicates when weight is settled
- ✅ Errors handled gracefully
- ✅ Tested with virtual device
- ✅ Tested with real scale (Mettler Toledo preferred)

---

### 1.3 Real Hardware Integration Testing (2 weeks, P0)

**Problem**: Tests exist but haven't been validated on real hardware matrix. Unknown bugs may exist.

**Impact**: Production surprises, device incompatibility, customer returns.

**Implementation Tasks**:

#### Week 2-3: Hardware Procurement & Testing

- [ ] **Task 1.3.1**: Create hardware test matrix (1 day)
  - Document required devices for comprehensive testing
  - Identify what's available in-house
  - Purchase missing devices (budget: $3000-5000)
  - Priority devices:
    1. ESC/POS printer (Epson TM-T88VI - TCP, USB)
    2. Cheap ESC/POS printer (Chinese brand - USB)
    3. Barcode scanner (Honeywell or Symbol - USB HID)
    4. Scale (Mettler Toledo or CAS - RS-232)
    5. Payment terminal (Mada certified terminal - TCP)
    6. ZPL label printer (Zebra ZD420 - TCP)
    7. Customer display (Epson DM-D110 - RS-232)
  - Deliverable: `docs/HARDWARE_MATRIX.md`

- [ ] **Task 1.3.2**: ESC/POS printer testing (2 days)
  - **TCP Printer**:
    - [ ] Connect and print test receipt
    - [ ] Test all text formatting (bold, underline, size)
    - [ ] Test alignment (left, center, right)
    - [ ] Test barcode printing (EAN-13, Code128, QR)
    - [ ] Test cash drawer pulse
    - [ ] Test auto-reconnection (unplug/replug network)
    - [ ] Print 100 receipts continuously (stability test)
  - **USB Printer**:
    - [ ] Same tests as TCP
    - [ ] Verify USB driver integration
  - **Arabic Test** (after Task 1.1 complete):
    - [ ] Print Arabic receipt
    - [ ] Verify readability
  - Deliverable: Test report with photos

- [ ] **Task 1.3.3**: ZPL label printer testing (1 day)
  - [ ] Connect Zebra printer
  - [ ] Print barcode labels
  - [ ] Print QR code labels
  - [ ] Test status queries
  - [ ] Print 50 labels (stability test)
  - Deliverable: Test report with photos

- [ ] **Task 1.3.4**: Barcode scanner testing (1 day)
  - [ ] Connect USB HID scanner
  - [ ] Scan EAN-13 barcodes
  - [ ] Scan Code128 barcodes
  - [ ] Scan QR codes
  - [ ] Test event streaming (gRPC, WebSocket)
  - [ ] Test disconnect/reconnect
  - [ ] Scan 100 items continuously
  - Deliverable: Test report

- [ ] **Task 1.3.5**: Scale testing (1 day)
  - [ ] Connect serial scale
  - [ ] Test GetWeight (after Task 1.2)
  - [ ] Test zero
  - [ ] Test tare
  - [ ] Test unit conversion
  - [ ] Test stability indicator
  - [ ] Weigh 20 items of different weights
  - Deliverable: Test report

- [ ] **Task 1.3.6**: Payment terminal testing (2 days)
  - [ ] Connect Mada terminal (TCP)
  - [ ] Test sale transaction (approved)
  - [ ] Test declined card
  - [ ] Test cancellation by cashier
  - [ ] Test timeout
  - [ ] Test void/refund
  - [ ] Test settlement
  - [ ] Verify Arabic translations on terminal display
  - [ ] Verify no PAN in logs (security check)
  - [ ] Run 50 transactions (stability test)
  - Deliverable: Test report, audit log review

- [ ] **Task 1.3.7**: Customer display testing (1 day)
  - [ ] Connect customer display
  - [ ] Test ShowDisplay with English text
  - [ ] Test ShowDisplay with Arabic text (after Task 1.1)
  - [ ] Test ClearDisplay
  - [ ] Test display during payment flow
  - Deliverable: Test report with photos

- [ ] **Task 1.3.8**: Document tested hardware (1 day)
  - Update `docs/HARDWARE_MATRIX.md` with results
  - List confirmed working models
  - Note any quirks or configuration requirements
  - Add photos of working setup
  - Deliverable: Approved hardware list

**Acceptance Criteria**:
- ✅ All device types tested with real hardware
- ✅ At least 2 models per critical device type (printers, scanners)
- ✅ Integration tests updated with real device results
- ✅ Known working hardware models documented
- ✅ Photos and test reports archived

---

### 1.4 Fix Misleading Documentation (3 days, P0)

**Problem**: README claims "100% MVP Complete" when 35% is incomplete. Multiple contradictory status files.

**Impact**: Lost credibility, wrong expectations, wasted developer time.

**Implementation Tasks**:

- [ ] **Task 1.4.1**: Create honest GAPS.md (1 day)
  - List all incomplete features
  - Reference code comments with "TODO"
  - Be specific: "Arabic ESC/POS printing - not implemented, see renderer.go:89"
  - Add estimated effort to complete each gap
  - Deliverable: `GAPS.md`

- [ ] **Task 1.4.2**: Update README.md (1 day)
  - Remove "100% MVP Complete" claims
  - Add "Status: 65% Production-Ready" (update as you fix things)
  - Add section "Tested Hardware" (after Task 1.3)
  - Add section "Known Limitations" linking to GAPS.md
  - Be honest about what works vs. what's planned
  - Example:
    ```markdown
    ## Implementation Status (65% Production-Ready)

    ### ✅ Production-Ready
    - Payment terminals (Mada, KNET) - ISO 8583, Arabic i18n
    - Serial scales (5 protocols: Mettler, Dibal, CAS, Toledo, Generic)
    - ESC/POS TCP printers (English text only, no Arabic yet)
    - USB HID barcode scanners
    - ZPL label printers
    - Security (TLS/mTLS, ACL)
    - Observability (Prometheus, Grafana)

    ### ⚠️ Partial / In Progress
    - Arabic ESC/POS printing (in progress, ETA: Week 3)
    - USB ESC/POS printers (driver exists, integration needed)
    - Customer display (minimal implementation)

    ### ❌ Not Yet Implemented
    - Serial ESC/POS printers
    - Receipt history / re-print
    - Image printing in ESC/POS

    See [GAPS.md](GAPS.md) for complete list.
    ```

- [ ] **Task 1.4.3**: Consolidate status files (1 day)
  - Repo has 9 different progress/status/roadmap markdown files
  - Consolidate into:
    - README.md - Overview, quick start
    - IMPLEMENTATION_PLAN.md - This file
    - GAPS.md - Known limitations
    - CHANGELOG.md - Version history
  - Archive or delete:
    - IMPLEMENTATION_STATUS.md (outdated)
    - PROGRESS.md (duplicate)
    - ROADMAP_*.md (multiple versions)
  - Deliverable: Clean, consistent documentation structure

**Acceptance Criteria**:
- ✅ README has honest status assessment
- ✅ GAPS.md lists all incomplete features
- ✅ No "100% complete" claims
- ✅ Consolidated to 4-5 key docs
- ✅ Documentation accurately reflects code

---

## Phase 2: Transport Layer Completion (Weeks 4-5)

### Priority: P1 - IMPORTANT FOR MOST DEPLOYMENTS

### 2.1 USB ESC/POS Printer Support (1 week, P1)

**Problem**: Only TCP transport implemented. Most POS setups use USB printers.

**Impact**: Cannot deploy in environments without networked printers (small shops, mobile POS).

**Implementation Tasks**:

- [ ] **Task 2.1.1**: Review existing USB printer driver (1 day)
  - Audit `internal/drivers/printer_usb/`
  - Check platform-specific implementations (Linux, Windows, macOS)
  - Verify USB device discovery
  - Check if integration with ESC/POS renderer exists

- [ ] **Task 2.1.2**: Integrate USB transport with ESC/POS driver (2 days)
  - Option A: Make `printer_escpos` support multiple transports (TCP, USB, Serial)
  - Option B: Create wrapper that uses USB driver + ESC/POS renderer
  - Prefer Option A for cleaner architecture
  - Modify config to support:
    ```yaml
    devices:
      - id: printer-usb-1
        type: printer
        kind: escpos
        transport: usb
        usb:
          vendor_id: 0x04b8  # Epson
          product_id: 0x0e15  # TM-T88
    ```

- [ ] **Task 2.1.3**: USB device auto-discovery (1 day)
  - Scan for ESC/POS printers on USB
  - Maintain device list
  - Handle hot-plug events
  - Deliverable: Auto-discovery feature

- [ ] **Task 2.1.4**: Testing (2 days)
  - Test with Epson TM-T88 USB
  - Test with cheap Chinese USB printer
  - Test all ESC/POS features (formatting, barcodes, QR, Arabic)
  - Test hot-plug (unplug/replug while daemon running)
  - Platform testing: Linux (priority), Windows (if needed)
  - Deliverable: Test report

**Files**:
- `internal/drivers/printer_escpos/driver.go` - Add transport abstraction
- `internal/drivers/printer_escpos/transport_usb.go` - New file
- `internal/drivers/printer_usb/` - Integrate or refactor

**Acceptance Criteria**:
- ✅ USB ESC/POS printers work
- ✅ All ESC/POS features supported (Arabic, formatting, barcodes)
- ✅ Auto-discovery of USB printers
- ✅ Hot-plug handling
- ✅ Tested on Linux with 2 printer models

---

### 2.2 Serial ESC/POS Printer Support (3 days, P1)

**Problem**: No serial transport for ESC/POS. Some legacy POS environments use RS-232 printers.

**Impact**: Cannot deploy in legacy environments.

**Implementation Tasks**:

- [ ] **Task 2.2.1**: Add serial transport to ESC/POS driver (1 day)
  - Use `github.com/tarm/serial` or similar
  - Config example:
    ```yaml
    devices:
      - id: printer-serial-1
        type: printer
        kind: escpos
        transport: serial
        serial:
          port: /dev/ttyS0  # or COM1 on Windows
          baud_rate: 9600
          data_bits: 8
          parity: none
          stop_bits: 1
    ```

- [ ] **Task 2.2.2**: Testing with real serial printer (1 day)
  - Acquire RS-232 ESC/POS printer or USB-to-serial adapter
  - Test all ESC/POS features
  - Deliverable: Test report

- [ ] **Task 2.2.3**: Documentation (1 day)
  - Add serial config examples
  - Document common baud rates (9600, 19200, 38400)
  - Troubleshooting guide for serial connections

**Acceptance Criteria**:
- ✅ Serial ESC/POS printers work
- ✅ Tested with real serial hardware or adapter
- ✅ Documented configuration

---

## Phase 3: POS Features (Weeks 6-7)

### Priority: P1-P2 - COMMON POS REQUIREMENTS

### 3.1 Customer Display Enhancement (1 week, P1)

**Problem**: Minimal implementation, unclear if production-ready.

**Impact**: Customer-facing display not working properly affects customer experience.

**Implementation Tasks**:

- [ ] **Task 3.1.1**: Audit existing customer display driver (1 day)
  - Check implementation status
  - Identify missing features
  - Test with real customer display (Epson DM-D110 or similar)

- [ ] **Task 3.1.2**: Implement missing features (2 days)
  - Multi-line support (typically 2x20 or 2x40 character displays)
  - Scrolling text for long messages
  - Special characters (currency symbols)
  - Brightness control (if supported)

- [ ] **Task 3.1.3**: Add Arabic support (1 day)
  - Apply Arabic rendering from Task 1.1 to customer display
  - Test with Arabic messages
  - Handle 2-line RTL layout

- [ ] **Task 3.1.4**: Testing (1 day)
  - Test with real customer display
  - Test common POS messages:
    - Welcome message
    - Item scanned (name + price)
    - Total amount
    - Payment prompts
    - Thank you message
  - Test English and Arabic

**Acceptance Criteria**:
- ✅ Customer display fully functional
- ✅ Arabic text supported
- ✅ Tested with real hardware
- ✅ Common POS messages working

---

### 3.2 Receipt History & Re-print (1 week, P2)

**Problem**: No receipt history or re-print capability. Common POS requirement.

**Impact**: Cannot re-print lost receipts for customers.

**Implementation Tasks**:

- [ ] **Task 3.2.1**: Design receipt storage (1 day)
  - Decision: In-memory cache vs. persistent storage (SQLite?)
  - Cache last N receipts (configurable, default 100)
  - Include: receipt content, timestamp, transaction ID, device ID
  - Deliverable: Design document

- [ ] **Task 3.2.2**: Implement receipt cache (2 days)
  - Add to print job flow
  - Store receipt JSON after successful print
  - Add receipt retrieval API
  - Add receipt list API (last N receipts)

- [ ] **Task 3.2.3**: Add gRPC/REST endpoints (1 day)
  - `ListReceipts(limit, offset)` - Get receipt history
  - `GetReceipt(receipt_id)` - Get specific receipt
  - `ReprintReceipt(receipt_id, device_id)` - Re-print to device

- [ ] **Task 3.2.4**: Flutter SDK integration (1 day)
  - Add receipt history widgets
  - Add re-print button
  - Update example app

- [ ] **Task 3.2.5**: Testing (1 day)
  - Print 50 receipts
  - Retrieve receipt history
  - Re-print old receipt
  - Test cache eviction (when >100 receipts)

**Files**:
- `internal/receipt/cache.go` - New package
- `internal/api/grpc/server.go` - Add RPC methods
- `api/proto/v1/receipt.proto` - New proto file
- `sdk/flutter/lib/src/services/receipt_service.dart` - New service

**Acceptance Criteria**:
- ✅ Last 100 receipts cached
- ✅ Re-print API works
- ✅ Flutter SDK supports re-print
- ✅ Example app demonstrates re-print

---

### 3.3 End-to-End POS Flow Testing (1 week, P1)

**Problem**: Individual components work, but orchestrated POS flows haven't been validated.

**Impact**: Integration bugs in production.

**Implementation Tasks**:

- [ ] **Task 3.3.1**: Create POS flow test suite (2 days)
  - Normal sale:
    1. Start transaction
    2. Scan items (use scanner)
    3. Display total on customer display
    4. Start payment
    5. Payment approved
    6. Print receipt
    7. Open cash drawer
    8. Complete transaction
  - Cancelled sale before payment
  - Refund/return flow
  - Scale sale (weighed items)
  - Mixed sale (scanned + weighed + manual entry)
  - Deliverable: Automated test suite or detailed manual test plan

- [ ] **Task 3.3.2**: Test with real hardware (2 days)
  - Set up complete POS station:
    - Printer (receipt)
    - Scanner (barcode)
    - Scale (weighing)
    - Customer display
    - Payment terminal
    - Cash drawer
  - Run all POS flows
  - Measure latency at each step
  - Identify bottlenecks

- [ ] **Task 3.3.3**: Create demo video (1 day)
  - Record complete POS sale
  - Show all devices working together
  - Show Arabic receipts
  - Deliverable: Demo video for documentation

**Acceptance Criteria**:
- ✅ All POS flows tested end-to-end
- ✅ Real hardware integration validated
- ✅ Latency targets met (<200ms print, <1s scale)
- ✅ Demo video created

---

## Phase 4: Production Readiness (Weeks 8-10)

### Priority: P1 - REQUIRED FOR PRODUCTION

### 4.1 Load & Stress Testing (1 week, P1)

**Problem**: Performance under load not validated. Unknown scalability limits.

**Impact**: System crashes during busy hours.

**Implementation Tasks**:

- [ ] **Task 4.1.1**: Create load testing suite (2 days)
  - Use `k6` or `ghz` (gRPC load testing)
  - Simulate busy hour:
    - 10 transactions/minute/lane
    - 4 lanes = 40 transactions/minute
    - Each transaction:
      - 5-10 scanned items
      - 1 print job
      - 1 payment
  - Sustained load for 1 hour
  - Deliverable: Load test scripts

- [ ] **Task 4.1.2**: Run load tests (2 days)
  - Monitor CPU, RAM, network
  - Monitor latency (print, scale, payment)
  - Monitor error rate
  - Identify bottlenecks
  - Deliverable: Load test report

- [ ] **Task 4.1.3**: Optimize bottlenecks (2 days)
  - Profile hot code paths
  - Optimize if needed
  - Re-run load tests
  - Deliverable: Optimization report

- [ ] **Task 4.1.4**: 24-hour stability test (1 day)
  - Run bridge under load for 24+ hours
  - Monitor for memory leaks
  - Monitor for connection leaks
  - Check goroutine count
  - Deliverable: Stability report

**Acceptance Criteria**:
- ✅ 40 transactions/minute sustained for 1 hour
- ✅ Average print latency <200ms
- ✅ Average scale read latency <1s
- ✅ CPU usage <50% under load
- ✅ No memory leaks in 24-hour test
- ✅ No crashes or errors

---

### 4.2 CI/CD Pipeline (1 week, P1)

**Problem**: No working CI/CD. Manual testing is slow and error-prone.

**Impact**: Regression bugs, slow releases.

**Implementation Tasks**:

- [ ] **Task 4.2.1**: Set up GitHub Actions (1 day)
  - Create `.github/workflows/ci.yml`
  - Jobs:
    - `lint` - Run golangci-lint
    - `test` - Run unit tests
    - `build` - Build binaries for Linux/Windows/macOS
    - `integration` - Run integration tests with virtual devices
  - Trigger: On push to any branch, on PR
  - Deliverable: Working CI pipeline

- [ ] **Task 4.2.2**: Set up Docker image builds (1 day)
  - Create `.github/workflows/docker.yml`
  - Build multi-arch Docker images (amd64, arm64)
  - Push to GitHub Container Registry or Docker Hub
  - Tag with commit SHA and version
  - Deliverable: Automated Docker builds

- [ ] **Task 4.2.3**: Set up releases (1 day)
  - Create `.github/workflows/release.yml`
  - Trigger: On git tag push (e.g., `v1.0.0`)
  - Build binaries for all platforms
  - Create GitHub release with binaries
  - Generate changelog from commits
  - Deliverable: Automated releases

- [ ] **Task 4.2.4**: Remove pre-compiled binaries from repo (1 day)
  - Delete `bridge` and `bridge-cli` binaries (44MB)
  - Add to `.gitignore`
  - Update README: "Download binaries from Releases page"
  - Commit cleanup
  - Deliverable: Clean repo

**Files**:
- `.github/workflows/ci.yml` - New file
- `.github/workflows/docker.yml` - New file
- `.github/workflows/release.yml` - New file
- `.gitignore` - Add binary exclusions

**Acceptance Criteria**:
- ✅ CI runs on every push
- ✅ All tests must pass before merge
- ✅ Docker images built automatically
- ✅ Releases automated with GitHub Releases
- ✅ Pre-compiled binaries removed from repo

---

### 4.3 Security Hardening (3 days, P1)

**Problem**: Security implemented but not audited. PCI-DSS claims not verified.

**Impact**: Compliance issues, security vulnerabilities.

**Implementation Tasks**:

- [ ] **Task 4.3.1**: Security self-assessment (1 day)
  - Review TLS implementation (minimum TLS 1.2?)
  - Review ACL implementation (test deny rules)
  - Review audit logging (verify PAN masking)
  - Check for common vulnerabilities:
    - SQL injection (if database added)
    - Command injection
    - Path traversal
    - Denial of service
  - Deliverable: Security assessment report

- [ ] **Task 4.3.2**: Add security tests (1 day)
  - Test ACL deny rules
  - Test mTLS client cert rejection
  - Test PAN masking in logs
  - Test rate limiting (if implemented)
  - Deliverable: `internal/security/*_test.go` updates

- [ ] **Task 4.3.3**: Documentation for PCI-DSS (1 day)
  - Document PCI-DSS compliance measures
  - Note: This is not certification, just documentation
  - Recommend: Engage security auditor for real certification
  - Deliverable: `docs/SECURITY.md`

**Acceptance Criteria**:
- ✅ Security self-assessment completed
- ✅ Security tests added
- ✅ Documentation updated
- ✅ Recommendation for external audit documented

---

### 4.4 Documentation & Training (1 week, P2)

**Problem**: Documentation extensive but disorganized and contradictory.

**Impact**: Difficult onboarding, support burden.

**Implementation Tasks**:

- [ ] **Task 4.4.1**: Update all documentation (2 days)
  - README.md - Quick start, status
  - QUICKSTART.md - Step-by-step setup
  - ARCHITECTURE.md - System design
  - API_REFERENCE.md - gRPC, REST, WebSocket docs
  - HARDWARE.md - Supported devices (from Task 1.3)
  - SECURITY.md - TLS, mTLS, ACL setup
  - DEPLOYMENT.md - Docker, systemd, production checklist
  - GAPS.md - Known limitations
  - Delete/archive outdated docs

- [ ] **Task 4.4.2**: Create video tutorials (2 days)
  - Setup & installation (10 min)
  - Configuration walkthrough (15 min)
  - Hardware setup (20 min)
  - Flutter app integration (15 min)
  - Troubleshooting common issues (10 min)
  - Deliverable: 5 video tutorials

- [ ] **Task 4.4.3**: Create operator manual (2 days)
  - For non-developers (POS operators, IT staff)
  - How to start/stop bridge
  - How to check device status
  - How to add a new printer
  - Troubleshooting guide
  - Deliverable: `docs/OPERATOR_MANUAL.md`

**Acceptance Criteria**:
- ✅ All documentation updated and consistent
- ✅ Video tutorials created
- ✅ Operator manual created
- ✅ Easy for new developer to onboard in <1 hour

---

### 4.5 Production Deployment Checklist (2 days, P1)

**Problem**: No clear checklist for production deployment. Easy to miss critical steps.

**Impact**: Production outages, security issues.

**Implementation Tasks**:

- [ ] **Task 4.5.1**: Create deployment checklist (1 day)
  - Pre-deployment:
    - [ ] Hardware tested (Task 1.3)
    - [ ] Load testing passed (Task 4.1)
    - [ ] Security hardened (Task 4.3)
    - [ ] TLS certificates obtained
    - [ ] ACL rules configured
    - [ ] Backup plan documented
  - Deployment:
    - [ ] Install binary or Docker image
    - [ ] Copy config.yaml
    - [ ] Set up systemd service (Linux)
    - [ ] Start daemon
    - [ ] Verify /metrics endpoint
    - [ ] Verify all devices register
  - Post-deployment:
    - [ ] Monitor logs for errors
    - [ ] Run smoke tests
    - [ ] Set up Prometheus/Grafana
    - [ ] Set up alerts
    - [ ] Document deployment in runbook
  - Deliverable: `docs/DEPLOYMENT_CHECKLIST.md`

- [ ] **Task 4.5.2**: Create systemd unit files (1 day)
  - Create `device-bridge.service`
  - Auto-restart on failure
  - Logging to journald
  - Run as non-root user
  - Example:
    ```ini
    [Unit]
    Description=Device Bridge
    After=network.target

    [Service]
    Type=simple
    User=bridge
    ExecStart=/usr/local/bin/bridge -config /etc/bridge/config.yaml
    Restart=on-failure
    RestartSec=5s

    [Install]
    WantedBy=multi-user.target
    ```
  - Deliverable: `configs/systemd/device-bridge.service`

**Acceptance Criteria**:
- ✅ Deployment checklist created
- ✅ Systemd unit file created
- ✅ Tested on clean Linux VM
- ✅ Service starts automatically on boot
- ✅ Service auto-restarts on crash

---

## Phase 5 (Optional): Nice-to-Have Features

### Priority: P2-P3 - FUTURE ENHANCEMENTS

These can be deferred to post-MVP or future versions:

### 5.1 Image Printing in ESC/POS (1 week, P2)
- Add logo printing capability
- Support PNG, JPEG
- Dithering for receipt printers
- Useful for: store logos, product images, promotional graphics

### 5.2 OpenTelemetry Tracing (3 days, P2)
- Config exists, implement tracing
- Trace request flows across gRPC/REST/WebSocket
- Integrate with Jaeger or similar
- Useful for: debugging latency issues

### 5.3 Dynamic Log Level Change (2 days, P2)
- Change log level without restart
- Add HTTP endpoint: `POST /api/v1/log-level` with body `{"level": "debug"}`
- Useful for: production debugging

### 5.4 Additional Payment Providers (per provider, 1 week, P2)
- Benefit (Bahrain)
- Additional regional providers
- Each provider needs research + implementation + testing

### 5.5 Kitchen Display System (KDS) Driver (2 weeks, P2)
- Support kitchen displays (order routing)
- Driver for KDS protocols
- Integration with print jobs
- Useful for: restaurants

---

## Timeline & Resource Allocation

### Gantt Chart (10 Weeks)

```
Week 1-3: Phase 1 - Critical Blockers
  Week 1: Arabic printing prototype
  Week 2: Arabic integration + Hardware procurement + GetWeight fix
  Week 3: Arabic testing + Hardware testing + Documentation fix

Week 4-5: Phase 2 - Transport Layer
  Week 4: USB ESC/POS + Serial ESC/POS start
  Week 5: Serial ESC/POS finish + Testing

Week 6-7: Phase 3 - POS Features
  Week 6: Customer display + Receipt history start
  Week 7: Receipt history finish + End-to-end testing

Week 8-10: Phase 4 - Production Readiness
  Week 8: Load testing + CI/CD setup
  Week 9: Security hardening + Documentation
  Week 10: Deployment checklist + Final validation
```

### Team Size & Skills

**Recommended team**: 2-3 developers

**Developer 1 (Senior)**: Focus on Arabic printing (critical path)
- Task 1.1 (Arabic ESC/POS) - 3 weeks
- Task 2.1 (USB ESC/POS) - 1 week
- Task 3.3 (End-to-end testing) - 1 week

**Developer 2 (Mid-level)**: Focus on integration & testing
- Task 1.2 (GetWeight fix) - 2 days
- Task 1.3 (Hardware testing) - 2 weeks
- Task 3.1 (Customer display) - 1 week
- Task 3.2 (Receipt history) - 1 week
- Task 4.1 (Load testing) - 1 week

**Developer 3 (Junior) or DevOps**: Focus on infrastructure
- Task 1.4 (Documentation fix) - 3 days
- Task 2.2 (Serial ESC/POS) - 3 days
- Task 4.2 (CI/CD) - 1 week
- Task 4.4 (Documentation & training) - 1 week
- Task 4.5 (Deployment checklist) - 2 days

**Skills Required**:
- Go programming (all developers)
- Text rendering / internationalization (Developer 1 for Arabic)
- Hardware integration experience (Developer 2)
- DevOps / CI/CD (Developer 3)
- Linux system administration (all)

---

## Risk Management

### High-Risk Items

**Risk 1: Arabic Text Rendering Complexity**
- **Probability**: Medium
- **Impact**: High (blocks Middle East deployment)
- **Mitigation**:
  - Start early (Week 1)
  - Allocate senior developer
  - Budget 3 weeks (longest task)
  - Have backup plan: Use external service for receipt generation if needed

**Risk 2: Hardware Procurement Delays**
- **Probability**: Medium
- **Impact**: Medium (delays testing)
- **Mitigation**:
  - Order hardware in Week 1
  - Use expedited shipping
  - Have virtual devices as fallback for initial testing

**Risk 3: Load Testing Reveals Performance Issues**
- **Probability**: Low-Medium
- **Impact**: High (may require architecture changes)
- **Mitigation**:
  - Architecture already looks solid (connection pooling, job queue)
  - Budget extra week for optimizations if needed
  - Start load testing in Week 8 to have buffer

**Risk 4: Arabic Printing Works on Some Printers, Not Others**
- **Probability**: Medium
- **Impact**: High (limits device compatibility)
- **Mitigation**:
  - Test with 3+ printer models (Epson, Star, Chinese)
  - Document printer requirements (Unicode support, etc.)
  - May need printer-specific rendering tweaks

---

## Success Metrics

### Completion Criteria

**Phase 1 Complete** when:
- ✅ Arabic receipts print correctly on 3 printer models
- ✅ GetWeight returns real values
- ✅ All device types tested with real hardware
- ✅ Documentation is honest and accurate

**Phase 2 Complete** when:
- ✅ USB ESC/POS printers work
- ✅ Serial ESC/POS printers work
- ✅ Tested on Linux with real hardware

**Phase 3 Complete** when:
- ✅ Customer display works with Arabic
- ✅ Receipt history and re-print work
- ✅ End-to-end POS flows validated

**Phase 4 Complete** when:
- ✅ 24-hour load test passed
- ✅ CI/CD pipeline running
- ✅ Security self-assessment completed
- ✅ Deployment checklist validated

### Production-Ready Definition

System is **production-ready** when:
1. ✅ All Phase 1 tasks completed (critical blockers fixed)
2. ✅ All Phase 2 tasks completed (transport layer complete)
3. ✅ All Phase 4 tasks completed (production readiness validated)
4. ✅ Acceptance checklist shows 90%+ pass rate
5. ✅ At least one pilot deployment successful

Phase 3 is important but can be completed in parallel with pilot deployment.

---

## Budget Estimate

### Hardware ($3,000 - $5,000)
- ESC/POS printer (Epson) - $300
- ESC/POS printer (Chinese) - $100
- USB barcode scanner - $100
- Serial scale - $400
- Payment terminal (Mada) - $1,000-2,000 (most expensive)
- ZPL label printer (Zebra) - $500
- Customer display - $200
- Cash drawer - $100
- Cables, adapters, etc. - $200

### Software/Services ($500 - $1,000)
- TLS certificates (if not using Let's Encrypt) - $100
- Cloud VM for testing (if needed) - $100/month x 3 months = $300
- Load testing service (if using external) - $100
- Total: ~$500-1,000

### Labor (2-3 developers x 10 weeks)
- Developer 1 (Senior): 10 weeks x $2,000-3,000/week = $20,000-30,000
- Developer 2 (Mid): 10 weeks x $1,500-2,500/week = $15,000-25,000
- Developer 3 (Junior/DevOps): 10 weeks x $1,000-2,000/week = $10,000-20,000
- **Total Labor**: $45,000-75,000

**Total Project Budget**: $48,500-81,000

---

## Next Steps

### Immediate Actions (This Week)

1. **Review & Approve This Plan**
   - Review with stakeholders
   - Adjust timeline if needed
   - Allocate budget

2. **Assemble Team**
   - Hire or assign developers
   - Ensure senior developer for Arabic text work

3. **Order Hardware**
   - Purchase devices from Task 1.3.1 list
   - Expedite shipping

4. **Set Up Development Environment**
   - Clone repo
   - Set up test environment
   - Verify build works

5. **Kick Off Task 1.1.1**
   - Start researching Arabic text shaping libraries
   - This is the critical path

### Weekly Progress Reviews

- **Every Friday**: Review completed tasks, blockers, adjust plan
- **Week 3**: Phase 1 review - Go/No-Go for Phase 2
- **Week 5**: Phase 2 review - Assess if on track
- **Week 7**: Phase 3 review - Production readiness assessment
- **Week 10**: Final review - Production deployment decision

---

## Conclusion

This plan addresses the 35% completion gap identified in the acceptance checklist validation. The critical path is **Arabic ESC/POS printing** (3 weeks), which is essential for Middle East deployment.

**Key Takeaways**:
1. **What's good**: Core architecture, payment terminal driver, scale driver, security, observability
2. **What's missing**: Arabic printing, hardware validation, some gRPC placeholders
3. **What's needed**: 8-10 weeks with 2-3 developers to reach 95% production-ready
4. **Investment**: ~$50K-80K total (hardware + labor)

**Recommendation**:
- Start immediately with Phase 1 (critical blockers)
- Consider pilot deployment after Phase 1 + Phase 4 if English-only is acceptable
- Complete Phase 2-3 based on deployment requirements

The foundation is solid (B+ grade). With focused effort on the identified gaps, this can become an A-grade, production-ready device bridge for Middle East POS markets.
