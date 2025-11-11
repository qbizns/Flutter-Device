# Hardware Integration Tests

This directory contains integration tests that require physical hardware devices.

## Overview

Hardware tests are tagged with `// +build hardware` and are **not run** by default during normal testing. They must be explicitly enabled using build tags.

## Running Hardware Tests

### Prerequisites

1. **Connect physical hardware**:
   - USB HID barcode scanner (Symbol, Honeywell, Datalogic, etc.)
   - Serial scale (for scale tests)
   - Payment terminal (for payment tests)

2. **Set up USB permissions** (Linux only):
   ```bash
   # Add udev rules
   sudo tee /etc/udev/rules.d/99-barcode-scanners.rules <<EOF
   SUBSYSTEM=="usb", ATTRS{idVendor}=="05e0", MODE="0666"
   SUBSYSTEM=="usb", ATTRS{idVendor}=="0c2e", MODE="0666"
   EOF

   sudo udevadm control --reload-rules
   sudo udevadm trigger
   ```

3. **Verify scanner detected**:
   ```bash
   # Linux
   lsusb | grep -i scanner

   # macOS
   system_profiler SPUSBDataType | grep -i scanner

   # Windows
   Get-PnpDevice | Where-Object {$_.Class -eq "HIDClass"}
   ```

### Running Scanner Tests

#### Auto-detect scanner:
```bash
go test -tags=hardware -v ./test/hardware/
```

#### Specify scanner by vendor/product ID:
```bash
export SCANNER_VENDOR_ID=05e0  # Symbol/Zebra
export SCANNER_PRODUCT_ID=1200 # LS2208
go test -tags=hardware -v ./test/hardware/
```

#### Filter by serial number (for multiple scanners):
```bash
export SCANNER_VENDOR_ID=05e0
export SCANNER_PRODUCT_ID=1200
export SCANNER_SERIAL=ABC123
go test -tags=hardware -v ./test/hardware/
```

### Run Specific Tests

#### Test scanner connection only:
```bash
go test -tags=hardware -v -run TestScannerConnection ./test/hardware/
```

#### Test scan events:
```bash
go test -tags=hardware -v -run TestScannerScanEvent ./test/hardware/
```

#### Test multiple scans:
```bash
go test -tags=hardware -v -run TestScannerMultipleScans ./test/hardware/
```

#### Test symbology detection:
```bash
go test -tags=hardware -v -run TestScannerSymbologyDetection ./test/hardware/
```

#### Test multiple scanners:
```bash
go test -tags=hardware -v -run TestMultipleScanners ./test/hardware/
```

### Benchmarks

#### Benchmark scan throughput:
```bash
go test -tags=hardware -bench=BenchmarkScanThroughput -benchtime=30s ./test/hardware/
```

## Test Descriptions

### TestScannerAvailable
- **Purpose**: Detect if USB HID scanners are connected
- **Duration**: < 1 second
- **Requires**: Scanner connected
- **Action**: None (automatic detection)

### TestScannerConnection
- **Purpose**: Test connecting to scanner
- **Duration**: < 10 seconds
- **Requires**: Scanner connected
- **Action**: None (automatic)

### TestScannerScanEvent
- **Purpose**: Test receiving scan events
- **Duration**: Up to 60 seconds
- **Requires**: Scanner connected
- **Action**: **Scan 1 barcode within 60 seconds**

### TestScannerMultipleScans
- **Purpose**: Test multiple consecutive scans
- **Duration**: Up to 120 seconds
- **Requires**: Scanner connected
- **Action**: **Scan 5 barcodes within 120 seconds**

### TestScannerSymbologyDetection
- **Purpose**: Test symbology detection for different barcode types
- **Duration**: Up to 180 seconds
- **Requires**: Scanner connected + barcodes of different types
- **Action**: **Scan EAN-13, UPC-A, Code39, Code128, QR codes**

### TestScannerReconnection
- **Purpose**: Test automatic reconnection after disconnect
- **Duration**: Up to 180 seconds
- **Requires**: Scanner connected
- **Action**: **Manually disconnect/reconnect scanner during test**
- **Status**: Skipped by default (manual test)

### TestMultipleScanners
- **Purpose**: Test using multiple scanners simultaneously
- **Duration**: Up to 30 seconds
- **Requires**: 2+ scanners connected
- **Action**: **Scan barcodes on different scanners**

### BenchmarkScanThroughput
- **Purpose**: Measure scan processing throughput
- **Duration**: 30 seconds (configurable)
- **Requires**: Scanner connected
- **Action**: **Scan barcodes rapidly for 30 seconds**

## Test Results

### Expected Output

```bash
$ go test -tags=hardware -v ./test/hardware/

=== RUN   TestScannerAvailable
    scanner_hid_test.go:38: Found 1 scanner(s):
    scanner_hid_test.go:40:   [1] Symbol LS2208 S/N:12345 [bus 1 addr 5]
    scanner_hid_test.go:42:        Known model: Symbol LS2208
--- PASS: TestScannerAvailable (0.05s)

=== RUN   TestScannerConnection
    scanner_hid_test.go:67: Scanner connected successfully
    scanner_hid_test.go:68: Device: Test Scanner
    scanner_hid_test.go:69: Status: ready
--- PASS: TestScannerConnection (0.12s)

=== RUN   TestScannerScanEvent
    scanner_hid_test.go:87: Scanner ready. Please scan a barcode within 60 seconds...
    scanner_hid_test.go:91: ✅ Scan received!
    scanner_hid_test.go:92:    Barcode:   5901234123457
    scanner_hid_test.go:93:    Symbology: EAN13
    scanner_hid_test.go:94:    Timestamp: 2025-11-11T10:30:00Z
--- PASS: TestScannerScanEvent (3.42s)

=== RUN   TestScannerMultipleScans
    scanner_hid_test.go:118: Scanner ready. Please scan 5 barcodes within 120 seconds...
    scanner_hid_test.go:125: ✅ Scan 1/5: 5901234123457 (EAN13)
    scanner_hid_test.go:125: ✅ Scan 2/5: 012345678905 (UPCA)
    scanner_hid_test.go:125: ✅ Scan 3/5: ABC123 (CODE39)
    scanner_hid_test.go:125: ✅ Scan 4/5: Test123 (CODE128)
    scanner_hid_test.go:125: ✅ Scan 5/5: https://example.com (QRCODE)
--- PASS: TestScannerMultipleScans (15.23s)

PASS
ok      github.com/Macber-eg/Flutter-Device/test/hardware    18.820s
```

## Troubleshooting

### Scanner not detected

**Error**: `No USB HID scanners detected`

**Solution**:
1. Verify scanner is connected: `lsusb` (Linux) or `system_profiler SPUSBDataType` (macOS)
2. Check USB permissions (Linux): see setup above
3. Try different USB port
4. Check scanner power (some require external power)

### Permission denied (Linux)

**Error**: `failed to open USB device: libusb: access denied`

**Solution**:
1. Add udev rules (see setup above)
2. Add user to plugdev group: `sudo usermod -a -G plugdev $USER`
3. Log out and log back in
4. Or run with sudo (not recommended): `sudo go test -tags=hardware ...`

### Test timeout

**Error**: `Timeout waiting for scan event`

**Solution**:
1. Ensure you scanned a barcode during the test
2. Verify scanner is in HID keyboard emulation mode
3. Test scanner in text editor (should type characters)
4. Check scanner configuration (may need programming barcode)

### Scanner not scanning

**Problem**: Scanner connected but no scan events

**Solution**:
1. Test in text editor - scan should type characters
2. Check scanner mode (HID vs Serial vs USB CDC)
3. Verify scanner beeps/LED lights when scanning
4. Try different barcode or distance

### Multiple scanners conflict

**Problem**: Only one scanner works when multiple connected

**Solution**:
1. Use `SCANNER_SERIAL` to target specific scanner
2. Check for USB bus conflicts
3. Try different USB ports/controllers
4. Verify both scanners have unique serial numbers

## Adding New Hardware Tests

### Structure

```go
// +build hardware

package hardware

import (
    "testing"
    // ...
)

func TestMyHardwareFeature(t *testing.T) {
    // 1. Get config
    config := getTestConfig(t)

    // 2. Create driver
    driver := scanner_hid.NewDriver(...)

    // 3. Start driver
    ctx := context.Background()
    err := driver.Start(ctx)
    if err != nil {
        t.Fatalf("Failed: %v", err)
    }
    defer driver.Stop()

    // 4. Test feature
    // ...

    // 5. Verify results
    // ...
}
```

### Guidelines

1. **Always use build tag**: `// +build hardware`
2. **Skip if no hardware**: Check and skip gracefully
3. **Clear instructions**: Log what user needs to do
4. **Reasonable timeouts**: Don't wait forever
5. **Cleanup**: Always defer `Stop()`
6. **Document**: Add to this README

## CI/CD Integration

Hardware tests are **excluded** from normal CI pipelines because they require physical devices.

### Optional: Dedicated Hardware Test Runner

For organizations with dedicated test hardware:

```yaml
# .github/workflows/hardware-tests.yml
name: Hardware Tests

on:
  workflow_dispatch:  # Manual trigger only

jobs:
  test-scanners:
    runs-on: [self-hosted, hardware-scanner]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - name: Run scanner tests
        run: go test -tags=hardware -v ./test/hardware/
        env:
          SCANNER_VENDOR_ID: "05e0"
          SCANNER_PRODUCT_ID: "1200"
```

## Future Tests

Planned hardware test additions:

### Week 4-6: Serial Scales
- `test/hardware/scale_serial_test.go`
- Test serial communication
- Test weight reading
- Test tare/zero
- Test different protocols

### Week 7-9: Payment Terminals
- `test/hardware/payment_tcp_test.go`
- Test TCP connection
- Test sale transaction
- Test void/refund
- Test receipt printing

## Resources

- [Scanner Setup Guide](../../docs/SCANNER_SETUP.md)
- [Building Guide](../../docs/BUILDING.md)
- [USB Library Evaluation](../../docs/USB_LIBRARY_EVALUATION.md)
- [Phase 3 Roadmap](../../PHASE_3_ROADMAP.md)

## Support

For issues with hardware tests:
1. Check troubleshooting section above
2. Review scanner setup documentation
3. Open issue: https://github.com/Macber-eg/Flutter-Device/issues
4. Tag with: `hardware`, `testing`, `scanner`
