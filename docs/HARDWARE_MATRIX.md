# Hardware Testing Matrix & Procurement List

**Date**: 2025-11-13
**Purpose**: Phase 1.3 - Real Hardware Integration Testing
**Budget**: $3,000 - $5,000 USD
**Timeline**: Order Week 1, Receive Week 2, Test Week 2-3

---

## Executive Summary

This document lists **recommended hardware for comprehensive integration testing** of the Device Bridge. All devices listed are production-grade and commonly used in Middle East POS deployments.

**Priority**: P0 (Critical Blocker) - Must test with real hardware before production

**Why This Matters**:
- Virtual devices can only test so much
- Vendor-specific quirks need real validation
- Arabic printing requires actual printer testing
- Payment terminals MUST be tested with real banks/gateways

---

## Hardware Procurement List

### 1. ESC/POS Receipt Printers (2-3 units required)

#### 1.1 Epson TM-T88VI (High Priority) ⭐
**Purpose**: Industry standard, most common in Middle East

- **Model**: Epson TM-T88VI
- **Interfaces**: USB + Ethernet (both needed for testing)
- **Features**:
  - 80mm paper width
  - Auto-cutter
  - Cash drawer pulse
  - Code Page 864 (Arabic support)
- **Price**: ~$300-400 USD
- **Where to Buy**:
  - Amazon Business
  - Epson authorized dealers
  - Local POS equipment suppliers
- **Why This One**:
  - Most reliable ESC/POS printer
  - Excellent Arabic character set support
  - Well-documented
  - Used in 70%+ of retail deployments

**Testing Focus**:
- Arabic text rendering (critical!)
- TCP/IP connectivity
- USB connectivity (for Phase 2)
- Cash drawer trigger
- Auto-cutter
- Barcode/QR printing

---

#### 1.2 Star Micronics TSP143IIIU (Medium Priority)
**Purpose**: Alternative brand testing, popular in restaurants

- **Model**: Star Micronics TSP143IIIU
- **Interfaces**: USB
- **Features**:
  - 80mm paper width
  - Auto-cutter
  - USB interface
  - Fast printing (43 receipts/min)
- **Price**: ~$200-250 USD
- **Where to Buy**:
  - Amazon
  - Star Micronics website
  - POS equipment resellers
- **Why This One**:
  - Second most common brand
  - Different ESC/POS dialect (good for compatibility testing)
  - Popular in quick-service restaurants

**Testing Focus**:
- ESC/POS compatibility (Star has slight variations)
- USB-only deployment
- Speed under load

---

#### 1.3 Cheap Chinese ESC/POS Printer (High Priority) ⭐
**Purpose**: Test with low-cost hardware (common in small businesses)

- **Model**: Xprinter XP-80C or similar
- **Interfaces**: USB, Serial, Ethernet (varies by model)
- **Features**:
  - 80mm paper width
  - Auto-cutter
  - Very cheap ($60-100)
- **Price**: ~$60-100 USD
- **Where to Buy**:
  - AliExpress
  - Amazon (search "80mm thermal printer")
  - Local electronics markets
- **Why This One**:
  - Real-world testing (many small shops use cheap printers)
  - Test Arabic support on lower-end hardware
  - Verify we don't depend on high-end features

**Testing Focus**:
- Arabic text on low-end hardware
- Quirks and compatibility issues
- Character set limitations

---

### 2. Barcode Scanners (1-2 units)

#### 2.1 Honeywell Voyager 1200g (High Priority) ⭐
**Purpose**: USB HID scanner, industry standard

- **Model**: Honeywell Voyager 1200g
- **Interface**: USB HID
- **Features**:
  - 1D barcodes (all symbologies)
  - Laser scanner
  - Plug-and-play USB HID
- **Price**: ~$100-150 USD
- **Where to Buy**:
  - Amazon
  - Honeywell distributors
  - POS equipment suppliers
- **Why This One**:
  - Industry standard
  - Reliable USB HID implementation
  - Supports all common symbologies (EAN-13, Code128, etc.)

**Testing Focus**:
- USB HID event streaming
- gRPC and WebSocket scanner events
- Disconnect/reconnect handling
- Continuous scanning (100+ scans)

---

#### 2.2 Symbol/Zebra DS2208 (Optional)
**Purpose**: Alternative brand, 2D barcode support

- **Model**: Symbol/Zebra DS2208
- **Interface**: USB HID
- **Features**:
  - 1D and 2D barcodes
  - QR code scanning
  - Image capture
- **Price**: ~$150-200 USD
- **Where to Buy**:
  - Amazon
  - Zebra distributors
- **Why This One**:
  - 2D barcode support (QR codes common in Middle East)
  - Different USB HID implementation

**Testing Focus**:
- QR code scanning
- 2D barcode symbologies

---

### 3. Scales (1-2 units)

#### 3.1 Mettler Toledo Scale (High Priority) ⭐
**Purpose**: Industry standard for retail weighing

- **Model**: Mettler Toledo 8217 or similar
- **Interface**: RS-232 serial (USB-to-serial adapter needed)
- **Features**:
  - 15kg / 30lb capacity
  - 0.005kg precision
  - MT-SICS protocol
  - Backlit display
- **Price**: ~$300-500 USD
- **Where to Buy**:
  - Mettler Toledo authorized dealers
  - Lab equipment suppliers
  - Used/refurbished options on eBay
- **Why This One**:
  - Our driver has excellent MT-SICS protocol support
  - Industry standard
  - Most reliable for deli/produce

**Testing Focus**:
- GetWeight (now fixed!)
- Zero and Tare operations
- Stable weight detection
- Unit conversion (kg, g, lb, oz)
- MT-SICS protocol validation

---

#### 3.2 CAS S-2000 Jr (Alternative)
**Purpose**: Alternative brand, cheaper option

- **Model**: CAS S-2000 Jr
- **Interface**: RS-232 serial
- **Features**:
  - 60lb capacity
  - 0.02lb precision
  - CAS protocol
- **Price**: ~$200-300 USD
- **Where to Buy**:
  - Amazon
  - Restaurant supply stores
  - CAS distributors
- **Why This One**:
  - Our driver supports CAS protocol
  - More affordable
  - Common in grocery stores

**Testing Focus**:
- CAS protocol validation
- Alternative to Mettler Toledo

---

### 4. Payment Terminals (1-2 units) - MOST CRITICAL

#### 4.1 Mada-Certified Payment Terminal (CRITICAL) ⭐⭐⭐
**Purpose**: Test real Saudi payment gateway

- **Options**:
  - **Ingenico iCT250** (most common)
  - **Verifone VX520** (alternative)
  - **PAX S300** (budget option)
- **Interface**: TCP/IP (Ethernet/WiFi)
- **Features**:
  - Mada certification
  - ISO 8583 protocol
  - EMV chip reader
  - Contactless (NFC)
  - PIN pad
- **Price**: ~$1,000-2,000 USD (**most expensive item**)
- **Where to Buy**:
  - **CRITICAL**: Must purchase from authorized Saudi payment gateway provider
  - Contact: STC Pay, HyperPay, or local acquiring banks
  - Cannot buy on Amazon (needs gateway integration)
- **Why This One**:
  - Our payment driver is excellent (98% complete)
  - MUST test with real Mada transactions
  - Arabic i18n needs validation on real terminal display

**Setup Requirements**:
- Test merchant account with Saudi acquiring bank
- Gateway credentials (provided by payment provider)
- May need to coordinate with bank for test mode

**Testing Focus**:
- ISO 8583 transaction flow
- Sale, void, refund, pre-auth
- Arabic display on terminal
- Transaction timeout handling
- Reversal/cancellation
- Settlement/batch processing
- Audit log validation (PAN masking)
- 50+ test transactions

**⚠️ IMPORTANT**:
This is the most complex item to procure. May need 2-3 weeks lead time and coordination with payment provider.

---

#### 4.2 KNET Terminal (Kuwait) (Optional)
**Purpose**: Test Kuwaiti payment gateway

- **Options**: Similar to Mada (Ingenico, Verifone, PAX)
- **Price**: ~$1,000-1,500 USD
- **Where to Buy**: Kuwaiti acquiring banks or KNET-certified providers
- **Why This One**: Our driver supports KNET protocol

**Note**: Only if deploying in Kuwait. Otherwise skip for now.

---

### 5. ZPL Label Printer (1 unit)

#### 5.1 Zebra ZD421 (Medium Priority)
**Purpose**: Test ZPL driver for product labels

- **Model**: Zebra ZD421 (203 dpi)
- **Interface**: USB + Ethernet
- **Features**:
  - 4" x 6" label size
  - ZPL II programming language
  - Barcode and QR printing
  - 203 dpi resolution
- **Price**: ~$400-600 USD
- **Where to Buy**:
  - Amazon Business
  - Zebra distributors
  - Office supply stores
- **Why This One**:
  - Our ZPL driver is 95% complete
  - Standard for product labeling
  - Common in warehouses and retail

**Testing Focus**:
- ZPL command generation
- Barcode label printing (EAN, Code128)
- QR code labels
- TCP/IP connectivity
- Status queries

---

### 6. Customer Display (1 unit)

#### 6.1 Epson DM-D110 (Low Priority)
**Purpose**: Test customer-facing display

- **Model**: Epson DM-D110
- **Interface**: USB or RS-232
- **Features**:
  - 2 lines x 20 characters
  - VFD (Vacuum Fluorescent Display)
  - Bright, easy to read
- **Price**: ~$150-250 USD
- **Where to Buy**:
  - Amazon
  - Epson distributors
  - POS equipment suppliers
- **Why This One**:
  - Standard 2x20 character display
  - Good for testing Arabic display (depends on Gap #1)

**Testing Focus**:
- ShowDisplay / ClearDisplay APIs
- Arabic character display (after Arabic rendering is implemented)
- Common POS messages

**Note**: Lower priority since our driver is minimal (see GAPS.md). Can defer until after Arabic printing is working.

---

### 7. Cash Drawer (1 unit)

#### 7.1 APG Vasario Cash Drawer (Low Priority)
**Purpose**: Test drawer pulse from printer

- **Model**: APG Vasario 1616
- **Interface**: RJ-11/RJ-12 (connects to printer)
- **Features**:
  - 16" width
  - 5 bill / 5 coin compartments
  - Electronic lock
- **Price**: ~$100-200 USD
- **Where to Buy**:
  - Amazon
  - POS equipment suppliers
- **Why This One**:
  - Industry standard
  - Compatible with ESC/POS printers

**Testing Focus**:
- OpenDrawer API
- ESC/POS pulse command
- Multiple drawer support (if needed)

**Note**: Low priority - our driver works (pulse command implemented). Can use drawer that comes with printer bundle.

---

### 8. Accessories & Cables

#### 8.1 USB-to-Serial Adapters (2-3 units)
- **Purpose**: Connect RS-232 scales and displays to modern PCs
- **Model**: FTDI-based adapters (avoid cheap Prolific chips)
- **Price**: ~$15-25 each
- **Quantity**: 2-3 (scale, customer display, spare)

#### 8.2 Network Cables & Switch
- **Purpose**: Connect TCP/IP devices
- **Items**:
  - 5-port Ethernet switch (~$20)
  - Cat6 cables x 5 (~$20)

#### 8.3 Thermal Paper Rolls
- **Purpose**: For receipt printer testing
- **Size**: 80mm x 80mm thermal rolls
- **Quantity**: 10-20 rolls
- **Price**: ~$30-50 for pack of 20

#### 8.4 Barcode Test Sheet
- **Purpose**: Test different symbologies
- **Items**:
  - Pre-printed barcode test sheet (EAN-13, Code128, QR, etc.)
  - Or print from online barcode generator
- **Price**: Free (print) or $10-20 (pre-printed)

---

## Procurement Priority & Budget

### Critical Priority (P0) - Order Immediately

| Item | Quantity | Unit Price | Total | Notes |
|------|----------|------------|-------|-------|
| **Epson TM-T88VI** | 1 | $350 | $350 | For Arabic testing |
| **Cheap Chinese Printer** | 1 | $80 | $80 | Low-end testing |
| **Honeywell Scanner** | 1 | $120 | $120 | USB HID testing |
| **Mettler Toledo Scale** | 1 | $400 | $400 | GetWeight testing |
| **Mada Payment Terminal** | 1 | $1,500 | $1,500 | **Most critical** |
| **USB-Serial Adapters** | 3 | $20 | $60 | For scale/display |
| **Thermal Paper** | 1 pack | $40 | $40 | 20 rolls |
| **Network Switch & Cables** | 1 | $40 | $40 | For TCP devices |
| **SUBTOTAL (P0)** | | | **$2,590** | Essential items |

### High Priority (P1) - Order If Budget Allows

| Item | Quantity | Unit Price | Total | Notes |
|------|----------|------------|-------|-------|
| **Star TSP143 Printer** | 1 | $220 | $220 | Alternative brand |
| **Zebra ZD421 Label Printer** | 1 | $500 | $500 | ZPL testing |
| **CAS S-2000 Scale** | 1 | $250 | $250 | Alternative scale |
| **SUBTOTAL (P1)** | | | **$970** | If budget > $3K |

### Low Priority (P2) - Defer or Skip

| Item | Quantity | Unit Price | Total | Notes |
|------|----------|------------|-------|-------|
| **Epson Customer Display** | 1 | $200 | $200 | Driver minimal |
| **APG Cash Drawer** | 1 | $150 | $150 | Can use printer bundle |
| **Symbol DS2208 Scanner** | 1 | $180 | $180 | 2D barcodes |
| **KNET Terminal** | 1 | $1,200 | $1,200 | Only if Kuwait deploy |
| **SUBTOTAL (P2)** | | | **$1,730** | Optional |

---

## Budget Scenarios

### Scenario 1: Minimum Viable Testing ($2,600)
**Just P0 items**: 1 Epson printer, 1 cheap printer, 1 scanner, 1 scale, 1 payment terminal, accessories

**Coverage**: ~80% of critical functionality

### Scenario 2: Recommended ($3,600)
**P0 + some P1**: Add Star printer and Zebra label printer

**Coverage**: ~90% of critical functionality

### Scenario 3: Comprehensive ($5,300)
**P0 + P1 + select P2**: Add customer display, second scanner

**Coverage**: ~95% of all functionality

---

## Vendor Contacts & Resources

### General POS Equipment
- **Amazon Business**: Best for quick delivery, good return policy
- **POS-X**: https://www.pos-x.com/ (US distributor)
- **Barcode Discount**: https://www.barcodediscount.com/

### Payment Terminals (Mada/KNET)
- **Saudi Arabia (Mada)**:
  - STC Pay: https://stcpay.com.sa/en (acquiring services)
  - HyperPay: https://www.hyperpay.com/ (payment gateway)
  - Contact local banks: Al Rajhi, SNB, etc.
- **Kuwait (KNET)**:
  - KNET: https://www.knet.com.kw/
  - Contact Kuwaiti acquiring banks

**CRITICAL**: Payment terminal procurement requires merchant account setup. Start this process ASAP (2-3 weeks lead time).

### Scales
- **Mettler Toledo**: https://www.mt.com/ (find local distributor)
- **CAS**: https://www.cas-usa.com/

### Printers
- **Epson**: https://epson.com/pos (authorized resellers)
- **Star Micronics**: https://www.starmicronics.com/
- **Chinese Printers**: AliExpress, Amazon

---

## Procurement Timeline

### Week 1 (Now)
- ✅ Create this hardware matrix (DONE)
- ⏭️ Get budget approval ($2,600 - $5,300)
- ⏭️ Contact payment terminal provider (Mada - longest lead time)
- ⏭️ Order P0 items from Amazon (2-3 day shipping)

### Week 2
- ⏭️ Receive hardware
- ⏭️ Unbox and inventory
- ⏭️ Begin testing (see Test Plan below)
- ⏭️ Payment terminal setup with gateway

### Week 3
- ⏭️ Complete testing matrix
- ⏭️ Document results
- ⏭️ Update HARDWARE_MATRIX.md with "Tested" column

---

## Testing Checklist (To Be Completed Week 2-3)

For each device, test and mark: ✅ PASS / ❌ FAIL / ⚠️ PARTIAL / 🔴 NOT TESTED

### ESC/POS Printers

| Test | Epson TM-T88VI | Star TSP143 | Chinese Xprinter | Notes |
|------|----------------|-------------|------------------|-------|
| TCP/IP connection | 🔴 | - | 🔴 | |
| USB connection | 🔴 | 🔴 | 🔴 | |
| Text printing (English) | 🔴 | 🔴 | 🔴 | |
| Text formatting (bold, underline, size) | 🔴 | 🔴 | 🔴 | |
| Alignment (L/C/R) | 🔴 | 🔴 | 🔴 | |
| **Arabic text rendering** | 🔴 | 🔴 | 🔴 | **After Phase 1.1 complete** |
| Barcode (EAN-13) | 🔴 | 🔴 | 🔴 | |
| Barcode (Code128) | 🔴 | 🔴 | 🔴 | |
| QR code | 🔴 | 🔴 | 🔴 | |
| Auto-cutter | 🔴 | 🔴 | 🔴 | |
| Cash drawer pulse | 🔴 | 🔴 | 🔴 | |
| Auto-reconnect (unplug/replug) | 🔴 | 🔴 | 🔴 | |
| Continuous printing (100 receipts) | 🔴 | 🔴 | 🔴 | |

### Barcode Scanners

| Test | Honeywell 1200g | Symbol DS2208 | Notes |
|------|-----------------|---------------|-------|
| USB HID detection | 🔴 | 🔴 | |
| EAN-13 scanning | 🔴 | 🔴 | |
| Code128 scanning | 🔴 | 🔴 | |
| QR code scanning | - | 🔴 | 2D only |
| gRPC event stream | 🔴 | 🔴 | |
| WebSocket events | 🔴 | 🔴 | |
| Disconnect/reconnect | 🔴 | 🔴 | |
| Continuous scanning (100+ scans) | 🔴 | 🔴 | |

### Scales

| Test | Mettler Toledo | CAS S-2000 | Notes |
|------|----------------|------------|-------|
| Serial port connection | 🔴 | 🔴 | Via USB-serial adapter |
| GetWeight returns value | 🔴 | 🔴 | **Fixed in Phase 1.2!** |
| Stable weight detection | 🔴 | 🔴 | |
| Zero operation | 🔴 | 🔴 | |
| Tare operation | 🔴 | 🔴 | |
| Unit conversion (kg/g/lb/oz) | 🔴 | 🔴 | |
| Protocol auto-detection | 🔴 | 🔴 | |
| Weigh 20 different items | 🔴 | 🔴 | |

### Payment Terminals

| Test | Mada Terminal | KNET Terminal | Notes |
|------|---------------|---------------|-------|
| TCP/IP connection | 🔴 | - | |
| Sale transaction (approved) | 🔴 | - | |
| Declined card | 🔴 | - | |
| Cancellation by cashier | 🔴 | - | |
| Transaction timeout | 🔴 | - | |
| Void transaction | 🔴 | - | |
| Refund transaction | 🔴 | - | |
| Pre-authorization | 🔴 | - | |
| Settlement/batch | 🔴 | - | |
| **Arabic display on terminal** | 🔴 | - | **Critical for UX** |
| PAN masking in logs | 🔴 | - | **Security check** |
| 50+ test transactions | 🔴 | - | Stability |

### ZPL Label Printer

| Test | Zebra ZD421 | Notes |
|------|-------------|-------|
| TCP/IP connection | 🔴 | |
| USB connection | 🔴 | |
| Barcode label (Code128) | 🔴 | |
| QR code label | 🔴 | |
| Product label template | 🔴 | |
| Status query | 🔴 | |
| Print 50 labels | 🔴 | |

---

## Success Criteria

**Phase 1.3 is COMPLETE when**:
- ✅ All P0 hardware procured and received
- ✅ All P0 devices tested with results documented
- ✅ At least 2 ESC/POS printers tested with Arabic (after Phase 1.1 complete)
- ✅ Mada payment terminal tested with real transactions
- ✅ Scale GetWeight validated with real weight readings
- ✅ Scanner tested with continuous operation
- ✅ Test report created with photos
- ✅ HARDWARE_MATRIX.md updated with results

**Deliverables**:
1. This procurement list ✅ (DONE)
2. Purchase orders placed (Week 1)
3. Hardware received (Week 2)
4. Test results documented (Week 3)
5. Photos of working setup (Week 3)
6. Update GAPS.md with any new issues found (Week 3)

---

## Risk Mitigation

### Risk 1: Payment Terminal Procurement Delay
**Likelihood**: High
**Impact**: High (critical path)
**Mitigation**:
- Start payment terminal procurement NOW (longest lead time)
- Have backup: Use virtual payment terminal for initial testing
- Contact multiple providers in parallel

### Risk 2: Arabic Printing Fails on Some Printers
**Likelihood**: Medium
**Impact**: High
**Mitigation**:
- Test with 3 different printer models (Epson, Star, Chinese)
- Document which models work vs. don't work
- May need printer-specific rendering tweaks

### Risk 3: Budget Constraints
**Likelihood**: Medium
**Impact**: Medium
**Mitigation**:
- Start with P0 items only ($2,600)
- Add P1 items if budget allows
- Can rent some equipment (payment terminal especially)

### Risk 4: Shipping Delays
**Likelihood**: Low-Medium
**Impact**: Medium (delays testing)
**Mitigation**:
- Use Amazon Prime for fast shipping
- Order from multiple vendors
- Have contingency time buffer

---

## Appendix: Barcode Test Data

For testing scanners, use these standard test barcodes:

**EAN-13**:
- Coca-Cola: 5449000000996
- Pepsi: 012000010729

**Code128**:
- Test string: "ABC-12345-XYZ"

**QR Codes** (for testing payment, product info):
- Simple text: "https://example.com/product/12345"
- Arabic text: "تجربة رمز الاستجابة السريعة"

---

## Version History

- **2025-11-13**: Initial creation (Phase 1.3.1)
  - Identified 7 device categories
  - Recommended specific models
  - Created 3 budget scenarios ($2.6K - $5.3K)
  - Testing checklist template
  - Vendor contacts

**Next Update**: Week 3 after testing - mark all tests as PASS/FAIL

---

**See Also**:
- IMPLEMENTATION_PLAN.md - Phase 1.3 (Hardware Integration Testing)
- GAPS.md - Gap #3 (Hardware Integration Testing)
- ACCEPTANCE_CHECKLIST_VALIDATION.md - Section 2 (Device & Hardware Matrix)
