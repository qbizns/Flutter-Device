# USB HID Scanner Setup Guide

**Device Bridge v2 - Phase 3**

This guide covers setup and usage of USB HID barcode scanners across Linux, macOS, and Windows.

---

## Table of Contents

- [Overview](#overview)
- [Supported Scanners](#supported-scanners)
- [Platform Setup](#platform-setup)
  - [Linux Setup](#linux-setup)
  - [macOS Setup](#macos-setup)
  - [Windows Setup](#windows-setup)
- [Configuration](#configuration)
- [Usage](#usage)
- [Troubleshooting](#troubleshooting)
- [Advanced Topics](#advanced-topics)

---

## Overview

Device Bridge v2 supports USB HID barcode scanners through the `scanner_hid` driver. The driver:

- **Auto-detects** USB HID scanners
- **Parses** HID keyboard input into barcode data
- **Detects symbology** automatically (EAN-13, UPC-A, Code39, Code128, etc.)
- **Supports AIM identifiers** for advanced scanners
- **Handles hotplug** with automatic reconnection
- **Works across platforms** (Linux, macOS, Windows)

**Supported Modes:**
- ✅ HID Keyboard emulation (standard mode)
- ⏳ Serial mode (future)
- ⏳ USB CDC mode (future)

---

## Supported Scanners

### Known Scanner Database (22 models)

#### Symbol/Zebra
- LS2208 (VID:0x05e0, PID:0x1200)
- DS2208 (VID:0x05e0, PID:0x1300)
- DS9208 (VID:0x05e0, PID:0x1900)
- LS2208 Keyboard Wedge (VID:0x05e0, PID:0x1203)

#### Honeywell
- Voyager 1200g (VID:0x0c2e, PID:0x0200)
- Voyager 1400g (VID:0x0c2e, PID:0x0720)
- Xenon 1900 (VID:0x0c2e, PID:0x0b61)
- Xenon 1902 (VID:0x0c2e, PID:0x0b00)

#### Datalogic
- QuickScan QD2430 (VID:0x05f9, PID:0x4204)
- QuickScan QBT2430 (VID:0x05f9, PID:0x4206)
- Gryphon GD4430 (VID:0x05f9, PID:0x2206)
- Gryphon I GD4400 (VID:0x05f9, PID:0x2232)

#### Code Corporation
- CR2600 (VID:0x065a, PID:0x0001)
- CR2700 (VID:0x065a, PID:0x0002)

#### Opticon
- OPN-2001 (VID:0x065a, PID:0x0009)
- OPN-2002 (VID:0x065a, PID:0x0011)

#### Metrologic (now Honeywell)
- MS7120 Orbit (VID:0x0c2e, PID:0x0007)
- MS9520 Voyager (VID:0x0c2e, PID:0x0009)

#### Unitech
- MS840 (VID:0x1ec8, PID:0x0101)
- MS842 (VID:0x1ec8, PID:0x0102)

**Note**: Generic HID POS scanners (USB class 0x03) are also supported.

---

## Platform Setup

### Linux Setup

#### 1. Install System Dependencies

```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install -y libusb-1.0-0-dev pkg-config

# Fedora/RHEL/CentOS
sudo dnf install libusb-devel pkgconfig

# Arch Linux
sudo pacman -S libusb pkgconf
```

#### 2. Configure USB Permissions

**Option A: udev Rules (Recommended)**

```bash
# Create udev rule for all HID scanners
sudo tee /etc/udev/rules.d/99-barcode-scanners.rules <<EOF
# Allow non-root access to barcode scanners

# Symbol/Zebra scanners
SUBSYSTEM=="usb", ATTRS{idVendor}=="05e0", MODE="0666", GROUP="plugdev"

# Honeywell scanners
SUBSYSTEM=="usb", ATTRS{idVendor}=="0c2e", MODE="0666", GROUP="plugdev"

# Datalogic scanners
SUBSYSTEM=="usb", ATTRS{idVendor}=="05f9", MODE="0666", GROUP="plugdev"

# Code Corporation scanners
SUBSYSTEM=="usb", ATTRS{idVendor}=="065a", MODE="0666", GROUP="plugdev"

# Unitech scanners
SUBSYSTEM=="usb", ATTRS{idVendor}=="1ec8", MODE="0666", GROUP="plugdev"

# All HID devices (use with caution)
# SUBSYSTEM=="usb", ATTR{bInterfaceClass}=="03", MODE="0666", GROUP="plugdev"
EOF

# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger

# Add your user to plugdev group
sudo usermod -a -G plugdev $USER

# Log out and log back in for group changes to take effect
```

**Option B: Run as Root (Not Recommended)**

```bash
sudo ./bin/bridge
```

#### 3. Verify Scanner Detection

```bash
# List USB devices
lsusb

# Look for scanner in output (example):
# Bus 001 Device 005: ID 05e0:1200 Symbol Technologies, Inc. LS2208

# Check device permissions
ls -l /dev/bus/usb/001/005

# Should show:
# crw-rw-rw- 1 root plugdev ... /dev/bus/usb/001/005
```

---

### macOS Setup

#### 1. Install Homebrew (if not installed)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### 2. Install libusb

```bash
brew install libusb pkg-config
```

#### 3. Verify Scanner Detection

```bash
# List USB devices
system_profiler SPUSBDataType

# Or use ioreg
ioreg -p IOUSB -w0 | grep -i scanner
```

**Note**: macOS handles USB permissions automatically for most scanners. No additional setup required.

---

### Windows Setup

#### 1. Install WinUSB Driver (using Zadig)

Most scanners work with Windows HID drivers out of the box. If you need to use libusb:

1. **Download Zadig**: https://zadig.akeo.ie/
2. **Run Zadig** as Administrator
3. **Select your scanner** from the device list
4. **Choose WinUSB driver** (or libusb-win32)
5. **Click "Replace Driver"** or "Install Driver"

**Warning**: Only install WinUSB for specific scanners. Don't replace system HID drivers unless necessary.

#### 2. Verify Scanner Detection

```powershell
# PowerShell: List USB devices
Get-PnpDevice | Where-Object {$_.Class -eq "HIDClass"}

# Or use Device Manager:
# Win+X → Device Manager → Human Interface Devices
```

**Note**: libusb is included in the gousb package for Windows. No additional installation required.

---

## Configuration

### Device Configuration

Create a configuration file (`config.yaml`):

```yaml
devices:
  - id: scanner-front
    type: scanner.hid
    name: "Front Desk Scanner"
    config:
      vendor_id: 0x05e0      # Symbol/Zebra
      product_id: 0x1200     # LS2208
      # serial: "12345"      # Optional: specific scanner
      buffer_size: 10
      read_timeout: 1s
      reconnect_delay: 5s

  - id: scanner-back
    type: scanner.hid
    name: "Back Office Scanner"
    config:
      vendor_id: 0x0c2e      # Honeywell
      product_id: 0x0200     # Voyager 1200g

  # Auto-detect any HID scanner
  - id: scanner-auto
    type: scanner.hid
    name: "Auto-detected Scanner"
    config:
      # No vendor/product ID = auto-detect first HID scanner
      buffer_size: 10
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `vendor_id` | uint16 | 0 (auto) | USB Vendor ID (hex) |
| `product_id` | uint16 | 0 (auto) | USB Product ID (hex) |
| `serial` | string | "" (any) | Serial number (optional) |
| `buffer_size` | int | 10 | Scan event buffer size |
| `read_timeout` | duration | 1s | USB read timeout |
| `reconnect_delay` | duration | 5s | Reconnection delay |

---

## Usage

### Starting the Bridge

```bash
# With default config
./bin/bridge

# With custom config
./bin/bridge -config /path/to/config.yaml

# In Docker
docker run -p 50051:50051 -p 8080:8080 \
  --device=/dev/bus/usb \
  --privileged \
  -v /path/to/config.yaml:/etc/device-bridge/config.yaml \
  device-bridge:latest
```

**Note**: Docker requires `--device=/dev/bus/usb` and `--privileged` for USB access on Linux.

### Using the CLI

```bash
# List all scanners
bridge-cli devices list --type scanner.hid

# Get scanner details
bridge-cli devices get scanner-front

# Enumerate available scanners
bridge-cli devices discover --type scanner.hid

# Test scanner
bridge-cli test scan scanner-front
```

### API Examples

#### gRPC (Go)

```go
package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"google.golang.org/grpc"
)

func main() {
	conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
	defer conn.Close()

	client := pb.NewDeviceBridgeClient(conn)

	// Subscribe to scan events
	stream, _ := client.SubscribeScanEvents(context.Background(), &pb.SubscribeScanEventsRequest{
		DeviceId: "scanner-front",
	})

	for {
		event, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Scanned: %s (%s)\n", event.Data, event.Symbology)
	}
}
```

#### REST API

```bash
# Get scanner status
curl http://localhost:8080/v1/devices/scanner-front

# WebSocket: Subscribe to scan events
wscat -c ws://localhost:8080/ws/scanner/scanner-front

# Example output:
# {"type":"scanner.scan","device_id":"scanner-front","data":"5901234123457","symbology":"EAN13","timestamp":"2025-11-10T20:00:00Z"}
```

#### JavaScript (WebSocket)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/scanner/scanner-front');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(`Scanned: ${data.data} (${data.symbology})`);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

---

## Troubleshooting

### Scanner Not Detected

**Problem**: `no matching USB HID scanner found`

**Solutions**:

1. **Verify scanner is connected**:
   ```bash
   # Linux
   lsusb | grep -i scanner

   # macOS
   system_profiler SPUSBDataType | grep -i scanner

   # Windows
   Get-PnpDevice | Where-Object {$_.Class -eq "HIDClass"}
   ```

2. **Check permissions** (Linux):
   ```bash
   # Check device permissions
   ls -l /dev/bus/usb/*/*

   # Add udev rules (see Linux Setup)
   sudo tee /etc/udev/rules.d/99-barcode-scanners.rules <<EOF
   SUBSYSTEM=="usb", ATTRS{idVendor}=="05e0", MODE="0666"
   EOF

   sudo udevadm control --reload-rules
   sudo udevadm trigger
   ```

3. **Check vendor/product ID**:
   ```bash
   # Linux
   lsusb

   # Look for: Bus 001 Device 005: ID 05e0:1200 Symbol ...
   # Vendor ID: 0x05e0, Product ID: 0x1200
   ```

4. **Try auto-detect mode**:
   Remove `vendor_id` and `product_id` from config to auto-detect any HID scanner.

---

### Permission Denied (Linux)

**Problem**: `failed to open USB device: libusb: access denied`

**Solutions**:

1. **Add udev rules** (see Linux Setup)
2. **Add user to plugdev group**:
   ```bash
   sudo usermod -a -G plugdev $USER
   # Log out and log back in
   ```
3. **Check device permissions**:
   ```bash
   ls -l /dev/bus/usb/*/*
   # Should show: crw-rw-rw- or crw-rw-r--
   ```

---

### Scanner Not Scanning

**Problem**: Scanner connected but no scan events

**Solutions**:

1. **Verify scanner mode**: Ensure scanner is in HID keyboard emulation mode (not serial/USB CDC)
2. **Test in text editor**: Scan a barcode in notepad/text editor. Should type characters.
3. **Check logs**:
   ```bash
   tail -f /var/log/device-bridge/bridge.log | grep scanner
   ```
4. **Verify event subscription**:
   ```bash
   # gRPC
   bridge-cli test scan scanner-front

   # WebSocket
   wscat -c ws://localhost:8080/ws/scanner/scanner-front
   ```

---

### Driver Installation (Windows)

**Problem**: `failed to enumerate USB devices` on Windows

**Solution**: Install WinUSB driver using Zadig (see Windows Setup)

**Note**: Most scanners work with Windows HID drivers without Zadig. Only use Zadig if standard drivers don't work.

---

### Multiple Scanners

**Problem**: Want to use multiple scanners simultaneously

**Solution**: Configure each scanner with unique ID and serial number:

```yaml
devices:
  - id: scanner-1
    type: scanner.hid
    config:
      vendor_id: 0x05e0
      product_id: 0x1200
      serial: "ABC123"  # First scanner serial

  - id: scanner-2
    type: scanner.hid
    config:
      vendor_id: 0x05e0
      product_id: 0x1200
      serial: "XYZ789"  # Second scanner serial
```

---

## Advanced Topics

### AIM Identifiers

Some advanced scanners prefix barcodes with AIM (Association for Automatic Identification and Mobility) identifiers:

| Prefix | Symbology | Example |
|--------|-----------|---------|
| `]E0` | EAN-13 | `]E05901234123457` |
| `]E4` | EAN-8 | `]E412345678` |
| `]A0` | Code 39 | `]A0ABC123` |
| `]C0` | Code 128 | `]C0Test123` |
| `]Q3` | QR Code | `]Q3https://example.com` |

Device Bridge automatically strips AIM prefixes and reports the symbology separately.

### Custom Symbology Detection

Override symbology detection:

```go
package main

import "github.com/Macber-eg/Flutter-Device/internal/drivers/scanner_hid"

// Custom symbology detection
func detectCustomSymbology(barcode string) string {
	// Your custom logic
	if strings.HasPrefix(barcode, "99") {
		return "CUSTOM_CODE"
	}
	return scanner_hid.detectSymbology(barcode)
}
```

### Performance Tuning

For high-throughput scanning:

```yaml
config:
  buffer_size: 100        # Increase buffer
  read_timeout: 100ms     # Decrease timeout
```

**Benchmarks**:
- Single key parse: 19.65 ns/op (0 allocs)
- Complete barcode: 238.2 ns/op (2 allocs)

### Adding New Scanner Models

To add scanner to known database:

```go
// internal/drivers/scanner_hid/hid_linux.go (or hid_darwin.go, hid_windows.go)
var KnownScanners = []struct {
	VendorID  uint16
	ProductID uint16
	Vendor    string
	Model     string
}{
	// Add your scanner
	{0x1234, 0x5678, "YourVendor", "YourModel"},
	// ...
}
```

---

## Support

### Getting Help

- **Documentation**: https://docs.example.com/scanners
- **Issues**: https://github.com/Macber-eg/Flutter-Device/issues
- **Discussions**: https://github.com/Macber-eg/Flutter-Device/discussions

### Reporting Issues

When reporting scanner issues, include:

1. **Platform**: Linux/macOS/Windows
2. **Scanner model**: Vendor + Model
3. **USB IDs**: `lsusb` output (VID/PID)
4. **Error message**: Full error from logs
5. **Steps to reproduce**

### Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for:
- Adding new scanner models
- Improving symbology detection
- Platform-specific fixes
- Documentation improvements

---

## License

Apache 2.0 - See [LICENSE](../LICENSE) for details.
