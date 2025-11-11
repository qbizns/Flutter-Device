# USB HID Scanner - Platform Validation Guide

**Date:** 2025-11-11
**Week:** 2 of 14-Week Completion Plan
**Status:** 🧪 Software Complete - Awaiting Hardware Validation

---

## Overview

This document provides comprehensive validation procedures for the USB HID scanner driver across Linux, macOS, and Windows platforms. The driver is **software complete** but requires physical hardware testing to validate cross-platform functionality.

### Current Status

| Component | Status | Notes |
|-----------|--------|-------|
| **Software Implementation** | ✅ Complete | All platform code written |
| **Linux Support** | ✅ Expected to work | gousb works well on Linux |
| **macOS Support** | ⚠️ **Needs validation** | gousb compatibility uncertain |
| **Windows Support** | ⚠️ **Needs validation** | gousb compatibility uncertain |
| **Unit Tests** | ✅ Complete | Parser tests passing |
| **Hardware Tests** | ⏳ Pending | Requires physical scanner |

---

## Technical Due Diligence Findings

### Known Concern: Windows/macOS USB Compatibility

**From Due Diligence Report:**
> "The USB-HID scanner driver files for windows/darwin import github.com/google/gousb, which is primarily Linux/libusb-oriented. In practice, HID on Windows/macOS needs different stacks (hidapi, win32 HID, IOHID). Expect non-functional scanner support on those OSes without substantial rework."

### Our Assessment

**gousb Library Analysis:**
- **Linux:** ✅ Uses libusb (standard, well-supported)
- **macOS:** ⚠️ Uses libusb via Homebrew (may work but not native)
- **Windows:** ⚠️ Uses libusb-win32 or WinUSB (driver installation required)

**Potential Issues:**
1. **Windows:** Requires WinUSB or libusb-win32 driver (not inbox)
2. **macOS:** May require Homebrew libusb installation
3. **HID-specific:** gousb is generic USB, not HID-optimized
4. **Permissions:** May require admin/root on all platforms

**Alternative Libraries (if gousb fails):**
- **hidapi** (cross-platform HID-specific)
- **karalabe/usb** (alternative Go USB library)
- Platform-native: win32 HID API (Windows), IOHIDManager (macOS)

---

## Prerequisites for Testing

### 1. Hardware Requirements

**USB HID Barcode Scanner (any of):**
- Symbol/Zebra LS2208 (~$100)
- Symbol/Zebra DS2208 (~$150)
- Honeywell Voyager 1200g (~$120)
- Datalogic QuickScan QD2430 (~$130)
- Generic USB HID POS scanner

**Specifications:**
- Interface: USB HID (keyboard emulation)
- Connection: USB Type-A or USB-C (with adapter)
- Power: USB bus-powered (no external power needed)
- Protocol: USB HID POS (not serial-emulation)

### 2. Software Requirements

#### Linux
```bash
# Install libusb (most distros have it pre-installed)
# Ubuntu/Debian:
sudo apt-get install libusb-1.0-0 libusb-1.0-0-dev

# Fedora/RHEL:
sudo dnf install libusb libusbx-devel

# Arch:
sudo pacman -S libusb

# Verify libusb
pkg-config --modversion libusb-1.0
```

#### macOS
```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install libusb
brew install libusb

# Verify
brew list libusb
```

#### Windows
```powershell
# Option 1: Install Zadig (WinUSB driver installer)
# Download from: https://zadig.akeo.ie/

# Option 2: Use libusb-win32
# Download from: https://sourceforge.net/projects/libusb-win32/

# Note: You'll need to replace the scanner's inbox driver with WinUSB/libusb-win32
```

### 3. Permissions Setup

#### Linux (udev rules)
```bash
# Create udev rule for scanner access
sudo tee /etc/udev/rules.d/99-barcode-scanner.rules << 'EOF'
# Symbol/Zebra scanners
SUBSYSTEM=="usb", ATTR{idVendor}=="05e0", MODE="0666", GROUP="plugdev"

# Honeywell scanners
SUBSYSTEM=="usb", ATTR{idVendor}=="0c2e", MODE="0666", GROUP="plugdev"

# Datalogic scanners
SUBSYSTEM=="usb", ATTR{idVendor}=="05f9", MODE="0666", GROUP="plugdev"

# Generic HID scanners (Class 03)
SUBSYSTEM=="usb", ATTR{bInterfaceClass}=="03", MODE="0666", GROUP="plugdev"
EOF

# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger

# Add user to plugdev group
sudo usermod -a -G plugdev $USER
# Log out and back in for group change to take effect
```

#### macOS
```bash
# No special permissions usually needed
# If issues, grant Terminal/app Full Disk Access in System Preferences
```

#### Windows
```powershell
# Run as Administrator
# Install WinUSB driver via Zadig for scanner device
```

---

## Validation Test Plan

### Phase 1: Pre-Hardware Checks

**Goal:** Verify build and dependencies without scanner

```bash
# 1. Verify Go version
go version
# Should be go1.22 or later

# 2. Check gousb dependency
go list -m github.com/google/gousb
# Should show: github.com/google/gousb v1.1.3

# 3. Build for current platform
cd /path/to/Flutter-Device
go build ./internal/drivers/scanner_hid/...

# 4. Run unit tests
go test ./internal/drivers/scanner_hid/... -v
# Should show: PASS (all parser tests)

# 5. Check for build tags
go list -f '{{.GoFiles}}' ./internal/drivers/scanner_hid/
# Should show appropriate platform file (hid_linux.go, hid_darwin.go, or hid_windows.go)
```

**Expected Results:**
- ✅ Build succeeds on all platforms
- ✅ Unit tests pass (8 tests)
- ✅ Correct platform file compiled

### Phase 2: USB Device Detection

**Goal:** Detect scanner without opening

```bash
# Linux
lsusb | grep -E "05e0|0c2e|05f9"
# Should show scanner if connected

# macOS
system_profiler SPUSBDataType | grep -A 10 "Barcode"
# Should show scanner details

# Windows (PowerShell)
Get-PnpDevice | Where-Object {$_.FriendlyName -like "*Barcode*"}
# Should show scanner
```

**Expected Scanner Output (example):**
```
Bus 001 Device 005: ID 05e0:1200 Symbol Technologies, Inc. Barcode Scanner
```

### Phase 3: gousb Enumeration Test

**Goal:** Test gousb can find the scanner

Create test file `test_usb_enum.go`:
```go
package main

import (
	"fmt"
	"log"

	"github.com/google/gousb"
)

func main() {
	// Create USB context
	ctx := gousb.NewContext()
	defer ctx.Close()

	// List all USB devices
	devices, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// Match HID devices
		for _, cfg := range desc.Configs {
			for _, intf := range cfg.Interfaces {
				for _, setting := range intf.AltSettings {
					if setting.Class == gousb.ClassHID {
						fmt.Printf("Found HID device: %04x:%04x\n", desc.Vendor, desc.Product)
						return true
					}
				}
			}
		}
		return false
	})

	if err != nil {
		log.Fatalf("Failed to enumerate USB devices: %v", err)
	}

	fmt.Printf("Found %d HID device(s)\n", len(devices))

	for i, dev := range devices {
		desc, _ := dev.Desc()
		manufacturer, _ := dev.Manufacturer()
		product, _ := dev.Product()
		serial, _ := dev.SerialNumber()

		fmt.Printf("Device %d:\n", i+1)
		fmt.Printf("  Vendor:       %04x (%s)\n", desc.Vendor, manufacturer)
		fmt.Printf("  Product:      %04x (%s)\n", desc.Product, product)
		fmt.Printf("  Serial:       %s\n", serial)
		fmt.Printf("  Class:        %02x\n", desc.Class)
		fmt.Printf("  USB Version:  %s\n", desc.Spec)

		dev.Close()
	}
}
```

Run:
```bash
go run test_usb_enum.go
```

**Expected Output (if working):**
```
Found HID device: 05e0:1200
Found 1 HID device(s)
Device 1:
  Vendor:       05e0 (Symbol Technologies)
  Product:      1200 (Barcode Scanner)
  Serial:       1234567890
  Class:        00
  USB Version:  2.00
```

**If enumeration fails:**
- **Linux:** Check permissions (udev rules)
- **macOS:** Check libusb installation, permissions
- **Windows:** Check WinUSB driver installation

### Phase 4: Scanner Driver Test

**Goal:** Test full scanner driver functionality

```bash
# Build device-bridge
cd /path/to/Flutter-Device
make build

# Create test config
cat > config.test-scanner.yaml << 'EOF'
server:
  grpc_port: 50051
  http_port: 8080

devices:
  - id: scanner-test
    name: "Test USB Scanner"
    kind: scanner.hid
    vendor_id: 0x05e0    # Symbol/Zebra (adjust for your scanner)
    product_id: 0x1200   # Adjust for your scanner
    buffer_size: 10
    reconnect_delay: 5s
    read_timeout: 1s

logging:
  level: debug
  format: text
EOF

# Run device-bridge
./bin/device-bridge -config config.test-scanner.yaml
```

**Expected Output:**
```
INFO  Starting Device Bridge v2
INFO  Loading configuration from config.test-scanner.yaml
DEBUG USB HID scanner driver: Searching for device 05e0:1200
DEBUG Found scanner: Symbol Barcode Scanner (S/N: 1234567890)
INFO  Scanner device registered: scanner-test
INFO  Scanner device started successfully
INFO  gRPC server listening on :50051
INFO  REST API listening on :8080
```

### Phase 5: Scan Event Test

**Goal:** Verify scanner can read barcodes

**In another terminal:**
```bash
# Subscribe to scanner events via CLI
./bin/bridge-cli test scan scanner-test
```

**Or via grpcurl:**
```bash
grpcurl -plaintext -d '{"device_id": "scanner-test"}' \
  localhost:50051 devicebridge.v1.DeviceBridge/SubscribeScanner
```

**Test Procedure:**
1. Scan a barcode (e.g., product UPC, QR code)
2. Observe output

**Expected Output:**
```json
{
  "event": {
    "data": "1234567890123",
    "symbology": "EAN13",
    "timestamp": "2025-11-11T14:30:00Z",
    "deviceId": "scanner-test"
  }
}
```

**Scan multiple barcodes of different types:**
- UPC-A: 012345678901
- EAN-13: 1234567890128
- Code 128: Various
- QR Code: Various

**Expected Results:**
- ✅ All scans received
- ✅ Correct symbology detection
- ✅ No data corruption
- ✅ Low latency (< 100ms)

---

## Platform-Specific Validation

### Linux Validation

**Systems to Test:**
- Ubuntu 22.04 LTS (primary)
- Debian 12
- Fedora 39
- Arch Linux (latest)

**Validation Steps:**
1. Install libusb dependencies ✓
2. Set up udev rules ✓
3. Test enumeration ✓
4. Test scanning ✓
5. Test hotplug (unplug/replug scanner) ✓
6. Test multiple scanners ✓

**Expected Result:** ✅ Full functionality

### macOS Validation

**Systems to Test:**
- macOS 13 Ventura (Intel)
- macOS 14 Sonoma (Intel)
- macOS 14 Sonoma (Apple Silicon)

**Validation Steps:**
1. Install libusb via Homebrew ✓
2. Test enumeration ✓
3. Test scanning ✓
4. Test System Preferences permissions ✓
5. Test hotplug ✓

**Known Issues to Check:**
- IOHIDManager conflicts
- System extension requirements
- Permissions prompts

**Expected Result:** ⚠️ May work with libusb, document any issues

### Windows Validation

**Systems to Test:**
- Windows 10 (build 19045)
- Windows 11 (latest)

**Validation Steps:**
1. Install Zadig and WinUSB driver ✓
2. Replace inbox HID driver with WinUSB ✓
3. Test enumeration ✓
4. Test scanning ✓
5. Test hotplug ✓
6. Test driver restore (revert to inbox HID if needed) ✓

**Known Issues to Check:**
- WinUSB driver installation
- libusb-win32 compatibility
- Administrator privileges required
- Conflicts with Windows HID stack

**Expected Result:** ⚠️ May work with WinUSB, may need alternative approach

---

## Troubleshooting Guide

### Issue: Scanner not detected

**Symptoms:**
- `no matching USB HID scanner found`
- Empty device list

**Solutions:**

**Linux:**
```bash
# Check if scanner is connected
lsusb

# Check permissions
ls -l /dev/bus/usb/*/
# Should show 0666 permissions for scanner

# Check udev rules
cat /etc/udev/rules.d/99-barcode-scanner.rules

# Reload udev
sudo udevadm control --reload-rules
sudo udevadm trigger

# Try as root (temporarily)
sudo ./bin/device-bridge -config config.yaml
```

**macOS:**
```bash
# Check libusb
brew list libusb

# Reinstall libusb
brew reinstall libusb

# Check permissions
# System Preferences → Security & Privacy → Full Disk Access
# Add Terminal or your app

# Try with sudo (temporarily)
sudo ./bin/device-bridge -config config.yaml
```

**Windows:**
```powershell
# Check if scanner is in Device Manager
devmgmt.msc

# Install WinUSB driver via Zadig:
# 1. Download Zadig from https://zadig.akeo.ie/
# 2. Run as Administrator
# 3. Options → List All Devices
# 4. Select your scanner
# 5. Select WinUSB driver
# 6. Click "Replace Driver"

# Run as Administrator
.\bin\device-bridge.exe -config config.yaml
```

### Issue: Scanner detected but no scans received

**Symptoms:**
- Driver starts successfully
- No events when scanning

**Solutions:**

1. **Check HID report format**
   ```bash
   # Linux: Use usbhid-dump
   sudo usbhid-dump -d 05e0:1200
   # Scan a barcode, observe output
   ```

2. **Enable debug logging**
   ```yaml
   logging:
     level: debug  # or trace
   ```

3. **Test parser manually**
   ```go
   // Create test program
   package main

   import (
       "fmt"
       "github.com/Macber-eg/Flutter-Device/internal/drivers/scanner_hid"
   )

   func main() {
       // Sample HID report (adjust based on usbhid-dump)
       report := []byte{0x00, 0x00, 0x1E, 0x00, 0x00, 0x00, 0x00, 0x00}

       data, symbology := scanner_hid.ParseHIDReport(report)
       fmt.Printf("Data: %s, Symbology: %s\n", data, symbology)
   }
   ```

### Issue: gousb fails on Windows/macOS

**Symptoms:**
- Enumeration fails
- `libusb_init` errors
- Access denied

**Solutions:**

**Option 1: Try alternative library (karalabe/usb)**
```go
// Replace gousb import
import "github.com/karalabe/usb"
```

**Option 2: Use platform-native APIs**
- Windows: Implement using win32 HID API (hidsdi.h)
- macOS: Implement using IOHIDManager

**Option 3: Use hidapi**
```go
import "github.com/sstallion/go-hid"
```

**See:** `docs/USB_ALTERNATIVE_LIBS.md` (to be created in Week 3)

---

## Performance Benchmarks

### Expected Performance

| Metric | Target | Notes |
|--------|--------|-------|
| **Enumeration Time** | < 100ms | Time to find scanner |
| **Connection Time** | < 500ms | Time to open device |
| **Scan Latency** | < 50ms | Scan to event time |
| **CPU Usage (idle)** | < 1% | With scanner connected |
| **Memory Usage** | < 5MB | Per scanner device |
| **Concurrent Scanners** | 10+ | Multiple devices |

### Benchmark Test

```bash
# Run benchmark tests
go test ./internal/drivers/scanner_hid/... -bench=. -benchmem

# Expected output:
# BenchmarkParseHIDReport-8    5000000    250 ns/op    0 B/op    0 allocs/op
```

---

## Validation Checklist

### Pre-Hardware Checklist
- [ ] Go 1.22+ installed
- [ ] Dependencies installed (`go mod download`)
- [ ] Code compiles on target platform
- [ ] Unit tests pass
- [ ] Platform-specific file selected by build tags

### Hardware Checklist
- [ ] USB HID scanner procured
- [ ] libusb installed (macOS/Windows)
- [ ] Permissions configured (udev rules, etc.)
- [ ] Scanner detected by OS
- [ ] Scanner detected by gousb enumeration

### Functional Checklist
- [ ] Scanner enumerated by driver
- [ ] Connection established
- [ ] Scans received via gRPC
- [ ] Scans received via REST
- [ ] Scans received via WebSocket
- [ ] Symbology detection correct
- [ ] Hotplug detection works
- [ ] Multiple scanners supported
- [ ] Error recovery works

### Platform Validation Matrix
| Test Case | Linux | macOS | Windows |
|-----------|-------|-------|---------|
| Build | ⏳ | ⏳ | ⏳ |
| Enumerate | ⏳ | ⏳ | ⏳ |
| Connect | ⏳ | ⏳ | ⏳ |
| Scan | ⏳ | ⏳ | ⏳ |
| Hotplug | ⏳ | ⏳ | ⏳ |
| Multi-device | ⏳ | ⏳ | ⏳ |

Legend: ✅ Validated | ❌ Failed | ⚠️ Partial | ⏳ Pending

---

## Next Steps

### If Validation Succeeds (gousb works)
1. Update platform compatibility matrix
2. Document tested scanner models
3. Create setup guides for each platform
4. Move to Week 3 (USB Printer)

### If Validation Fails (gousb doesn't work)
1. Document specific failure modes
2. Evaluate alternative libraries:
   - karalabe/usb
   - hidapi (sstallion/go-hid)
   - Platform-native APIs
3. Create migration plan
4. Update Week 2 timeline

---

## References

- [gousb Documentation](https://github.com/google/gousb)
- [USB HID Specification](https://www.usb.org/hid)
- [libusb Documentation](https://libusb.info/)
- [Zadig (Windows WinUSB)](https://zadig.akeo.ie/)
- [USB Library Evaluation](../USB_LIBRARY_EVALUATION.md)

---

**Status:** 📋 Documentation Complete - Ready for Hardware Testing
**Updated:** 2025-11-11
**Next Review:** When hardware arrives
