# Known Limitations and Gaps

**Last Updated**: 2025-11-13
**Status**: 65% Production-Ready

This document provides an honest assessment of incomplete features and known limitations in the Device Bridge project. Use this as a reference when planning deployments or estimating completion effort.

---

## Critical Gaps (P0 - Blockers for Production)

### 1. Arabic ESC/POS Printing ❌ **NOT IMPLEMENTED**

**Status**: Code has explicit TODO acknowledging this is missing

**Evidence**:
```go
// internal/drivers/printer_escpos/renderer.go:89-94
// TODO: Implement proper text shaping for Arabic/complex scripts
// Currently assumes all text is LTR and doesn't handle:
// - Arabic character joining (initial, medial, final, isolated forms)
// - Right-to-left text direction
// - BiDi text mixing Arabic and Latin
// Consider using: github.com/go-text/typesetting or similar
```

**Impact**:
- Cannot deploy in Middle East markets (Saudi Arabia, Kuwait, UAE, Egypt, etc.)
- Arabic text will print as broken/disconnected characters
- Store names, addresses, product names in Arabic will be unreadable
- This is a **CRITICAL BLOCKER** for Arabic-speaking regions

**What Works**:
- Payment terminal Arabic i18n is excellent (253 lines of translations)
- Transaction receipts from payment terminal show Arabic correctly
- But general ESC/POS printer cannot render Arabic text

**What's Missing**:
- Text shaping engine (HarfBuzz, FriBidi, or Go equivalent)
- Glyph joining rules (initial/medial/final/isolated forms)
- Right-to-left (RTL) text direction
- Bidirectional (BiDi) algorithm for mixed Arabic/English
- Arabic-aware line wrapping
- RTL-aware text alignment

**Effort to Complete**: 3 weeks (see IMPLEMENTATION_PLAN.md Phase 1.1)

**Workaround**: Use English-only receipts, or pre-render receipts as images (but image printing also not implemented - see below)

---

### 2. Customer Display Implementation ⚠️ MINIMAL

**Status**: Placeholder implementation

**Evidence**:
```go
// internal/api/grpc/server.go:263-267
func (s *Server) ShowDisplay(ctx context.Context, req *pb.ShowDisplayRequest) (*pb.ShowDisplayResponse, error) {
    s.logger.Info("show display request", telemetry.String("device_id", req.DeviceId))
    return &pb.ShowDisplayResponse{Success: true}, nil  // Placeholder!
}
```

**Impact**:
- Customer-facing displays won't show transaction info
- Affects customer experience at checkout
- Not critical for basic POS but expected feature

**What's Missing**:
- Actual driver implementation for customer displays
- Multi-line support (2x20, 2x40 character displays)
- Arabic text support (depends on Gap #1)
- Display control (brightness, clear, positioning)

**Effort to Complete**: 1 week (see IMPLEMENTATION_PLAN.md Phase 3.1)

**Workaround**: Use without customer display, or implement custom driver

---

### 3. Hardware Integration Testing 🔴 NOT VALIDATED

**Status**: Tests exist but haven't been run on real hardware

**Evidence**:
- Test files present: `test/integration/hardware/`
- Virtual devices used for testing: `test/virtual_devices/`
- No test reports with real device results
- README lists devices as "Planned" but some are actually implemented

**Impact**:
- Unknown bugs may exist with real hardware
- Device compatibility uncertain
- May have surprises in production

**What's Missing**:
- Test results from real Epson/Star/Chinese printers
- Test results from real Mettler/CAS/Dibal scales
- Test results from real Mada/KNET payment terminals
- Test results from real Honeywell/Symbol scanners
- Hardware compatibility matrix

**Effort to Complete**: 2 weeks + $3K-5K hardware budget (see IMPLEMENTATION_PLAN.md Phase 1.3)

**Workaround**: Test with your specific hardware before production deployment

---

## Important Gaps (P1 - Limits Deployment Options)

### 4. USB ESC/POS Printers ⚠️ INCOMPLETE INTEGRATION

**Status**: USB driver code exists but integration unclear

**Evidence**:
- USB driver exists: `internal/drivers/printer_usb/`
- ESC/POS driver only has TCP transport: `internal/drivers/printer_escpos/driver.go`
- No config examples for USB ESC/POS printers

**Impact**:
- Most small businesses use USB printers
- Cannot deploy without networked printers
- Limits deployment to setups with network infrastructure

**What Works**:
- TCP/IP ESC/POS printers work well
- USB driver code exists separately

**What's Missing**:
- Integration of USB transport into ESC/POS driver
- USB device auto-discovery
- Hot-plug support
- Config schema for USB printers

**Effort to Complete**: 1 week (see IMPLEMENTATION_PLAN.md Phase 2.1)

**Workaround**: Use network-connected ESC/POS printers (requires network setup)

---

### 5. Serial (RS-232) ESC/POS Printers ❌ NOT IMPLEMENTED

**Status**: No serial transport for ESC/POS

**Evidence**:
- ESC/POS driver only has TCP: `internal/drivers/printer_escpos/driver.go:248`
- No serial port code in ESC/POS driver

**Impact**:
- Cannot use legacy RS-232 printers
- Some industrial/kitchen environments only have serial
- Limits hardware compatibility

**What's Missing**:
- Serial port transport for ESC/POS
- Baud rate configuration
- Serial port discovery/enumeration

**Effort to Complete**: 3 days (see IMPLEMENTATION_PLAN.md Phase 2.2)

**Workaround**: Use TCP or USB printers

---

### 6. Receipt History & Re-print ❌ NOT IMPLEMENTED

**Status**: No receipt caching or history

**Evidence**:
- No receipt storage code in codebase
- Print jobs go directly to device without caching
- No API for retrieving past receipts

**Impact**:
- Cannot re-print lost receipts
- Common customer request cannot be fulfilled
- May need manual receipt recreation

**What's Missing**:
- Receipt cache (in-memory or persistent)
- ListReceipts API
- ReprintReceipt API
- Receipt storage configuration

**Effort to Complete**: 1 week (see IMPLEMENTATION_PLAN.md Phase 3.2)

**Workaround**: Keep receipt data in POS application, regenerate and print if needed

---

## Nice-to-Have Gaps (P2 - Future Enhancements)

### 7. Image Printing in ESC/POS ❌ NOT IMPLEMENTED

**Status**: Not implemented

**Evidence**:
```go
// internal/drivers/printer_escpos/renderer.go
// No image rendering code
// Image element exists in proto but not implemented
```

**Impact**:
- Cannot print logos
- Cannot print product images
- Cannot print promotional graphics

**What's Missing**:
- Image format support (PNG, JPEG)
- Dithering for monochrome printers
- Image scaling/positioning
- Raster bit-image command generation

**Effort to Complete**: 1 week

**Workaround**: Use text-only receipts, or external receipt rendering

---

### 8. OpenTelemetry Tracing ⚠️ CONFIG ONLY

**Status**: Configuration exists but implementation unclear

**Evidence**:
- Config has `tracing.enabled` flag
- No obvious OpenTelemetry instrumentation in code

**Impact**:
- Cannot trace requests across services
- Harder to debug latency issues

**What's Missing**:
- OpenTelemetry SDK integration
- Span creation for requests
- Trace context propagation
- Exporter configuration (Jaeger, etc.)

**Effort to Complete**: 3 days

**Workaround**: Use structured logs and Prometheus metrics (which are excellent)

---

### 9. Dynamic Log Level Change ❌ NOT IMPLEMENTED

**Status**: Requires daemon restart to change log level

**Evidence**:
- Log level set at startup from config
- No runtime API to change level

**Impact**:
- Cannot enable debug logging without restart
- Harder to troubleshoot production issues

**What's Missing**:
- HTTP endpoint to change log level
- Runtime log level control

**Effort to Complete**: 2 days

**Workaround**: Restart daemon with new config (may lose in-flight transactions)

---

### 10. Systemd Service Files ❌ NOT INCLUDED

**Status**: Documented but no service files in repo

**Evidence**:
- README mentions systemd
- No `.service` files in repo

**Impact**:
- Manual setup required for Linux deployments
- No auto-start on boot
- No auto-restart on crash

**What's Missing**:
- `device-bridge.service` systemd unit file
- Installation documentation
- Service user/permissions setup

**Effort to Complete**: 1 day

**Workaround**: Create service file manually or use Docker

---

## Documentation Gaps

### 11. Misleading Completeness Claims 🔴 CRITICAL ISSUE

**Status**: README claims "100% MVP Complete" but ~35% is incomplete

**Evidence**:
- README has "100% MVP Complete" badges
- Multiple status files contradict each other
- 9 different progress/roadmap markdown files

**Impact**:
- Lost credibility
- Wrong deployment expectations
- Wasted developer time investigating "complete" features

**What's Wrong**:
- Overstated completion percentage
- Too many contradictory status documents
- Features listed as "complete" are actually "planned"

**Fix Required**:
- Remove "100% complete" claims
- Create honest status section (see IMPLEMENTATION_PLAN.md Phase 1.4)
- Consolidate to 4-5 key documentation files
- Add this GAPS.md to README

**Effort to Complete**: 3 days (in progress)

---

### 12. Pre-compiled Binaries in Repo 🔴 BAD PRACTICE

**Status**: 44MB of binaries committed to git

**Evidence**:
- `bridge` binary (23.6 MB)
- `bridge-cli` binary (20.5 MB)

**Impact**:
- Large repo size
- Binaries may be stale
- Poor development practice
- Harder to trust source code

**What's Wrong**:
- Binaries should not be in source control
- Should use GitHub Releases instead

**Fix Required**:
- Delete binaries from repo
- Add to `.gitignore`
- Set up automated releases (see IMPLEMENTATION_PLAN.md Phase 4.2)

**Effort to Complete**: 1 day

---

## What Actually Works (Production-Ready)

To balance this list, here's what IS production-ready:

### ✅ Excellent Components

1. **Payment Terminal Driver** (98% complete)
   - ISO 8583 protocol
   - Mada (Saudi) and KNET (Kuwait) providers
   - 253 lines of Arabic translations
   - Full transaction types (Sale, Void, Refund, PreAuth, etc.)
   - Audit logging with PAN masking
   - Well-tested (9 test files)

2. **Serial Scale Driver** (95% complete)
   - 5 protocols: Mettler Toledo MT-SICS, Dibal, CAS, Toledo, Generic
   - Auto-protocol detection
   - Comprehensive tests
   - Weight reading, zero, tare all implemented
   - (Just needed GetWeight wired to gRPC - now fixed!)

3. **ZPL Label Printer** (95% complete)
   - Full ZPL command generation
   - Barcode and QR code support
   - Builder pattern for layouts
   - TCP/IP transport

4. **USB HID Scanner** (90% complete)
   - Platform-specific implementations (Linux/Windows/macOS)
   - Event streaming (gRPC and WebSocket)
   - Auto-reconnection
   - Multiple symbologies

5. **ESC/POS TCP Printer** (70% complete - missing Arabic)
   - Text formatting (bold, underline, sizes)
   - Alignment (left, center, right)
   - Barcodes (8 symbologies)
   - QR codes
   - Cash drawer pulse
   - Auto-reconnection
   - **BUT: No Arabic text support**

6. **Security** (90% complete)
   - TLS/mTLS with client cert verification
   - ACL with pattern matching
   - API key authentication
   - Audit logging
   - (Not externally audited for PCI-DSS)

7. **Observability** (95% complete)
   - Prometheus metrics (all key metrics)
   - Grafana dashboards included
   - Structured JSON logging
   - All standard metrics implemented

8. **API Layer** (85% complete)
   - gRPC service (most methods implemented)
   - REST gateway (grpc-gateway)
   - WebSocket for real-time events
   - (Some methods like ShowDisplay are placeholders)

9. **Flutter SDK** (95% complete)
   - Type-safe Dart SDK
   - All device services
   - Widgets for scanner, RFID, etc.
   - Comprehensive example app

### ✅ What's Been Fixed Recently

- **GetWeight, ZeroScale, TareScale** - Fixed 2025-11-13
  - Was returning placeholder values
  - Now properly calls scale driver
  - Comprehensive tests added

---

## Summary by Priority

### Must Fix for Production (P0)
1. Arabic ESC/POS printing (3 weeks)
2. Hardware integration testing (2 weeks + hardware budget)
3. Customer display implementation (1 week)
4. Documentation accuracy (3 days)

### Should Fix for Most Deployments (P1)
5. USB ESC/POS support (1 week)
6. Serial ESC/POS support (3 days)
7. Receipt history & re-print (1 week)

### Nice to Have (P2)
8. Image printing (1 week)
9. OpenTelemetry tracing (3 days)
10. Dynamic log level (2 days)
11. Systemd service files (1 day)

---

## How to Use This Document

**For Deployment Planning**:
1. Check if any P0 gaps affect your market (e.g., Arabic for Middle East)
2. Verify P1 gaps against your hardware setup (e.g., USB vs TCP printers)
3. Plan workarounds or schedule completion work

**For Development Planning**:
- Use effort estimates in IMPLEMENTATION_PLAN.md
- Prioritize based on your deployment needs
- Each gap has clear acceptance criteria in implementation plan

**For Honest Communication**:
- Reference this doc when discussing project status
- Don't claim features are complete if listed here
- Update this doc as gaps are closed

---

## Reporting New Gaps

If you discover a gap not listed here:

1. Check if it's actually a bug vs. unimplemented feature
2. Verify by checking code and tests
3. Add to this document with:
   - Clear evidence (file paths, code quotes)
   - Impact assessment
   - Effort estimate (if known)
4. Submit PR with GAPS.md update

---

## Version History

- **2025-11-13**: Initial creation based on acceptance checklist validation
  - Assessed 8 major sections
  - Identified 12 major gaps
  - Added effort estimates
- **2025-11-13**: Fixed GetWeight/ZeroScale/TareScale (no longer a gap)

---

**See Also**:
- ACCEPTANCE_CHECKLIST_VALIDATION.md - Detailed validation results
- IMPLEMENTATION_PLAN.md - 10-week plan to close these gaps
- README.md - Project overview (being updated for accuracy)
