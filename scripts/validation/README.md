# Device Bridge Validation Scripts

This directory contains automated validation scripts for testing Device Bridge v2 functionality with real hardware.

## Available Scripts

### USB HID Scanner Validation

#### Linux/macOS
```bash
./test-usb-scanner.sh [vendor_id] [product_id]
```

**Example:**
```bash
# Test Symbol/Zebra LS2208 scanner
./test-usb-scanner.sh 05e0 1200

# Test Honeywell Voyager scanner
./test-usb-scanner.sh 0c2e 0b61
```

#### Windows
```powershell
.\test-usb-scanner.ps1 [vendor_id] [product_id]
```

**Example:**
```powershell
# Test Symbol/Zebra LS2208 scanner (run as Administrator)
.\test-usb-scanner.ps1 05e0 1200
```

## What These Scripts Do

1. **Platform Detection** - Identify OS and architecture
2. **Dependency Check** - Verify Go, libusb, and other requirements
3. **Hardware Detection** - Find connected USB HID scanner
4. **Permission Check** - Validate USB access permissions
5. **Build Test** - Compile Device Bridge from source
6. **Configuration** - Create test config for the specific scanner
7. **Runtime Test** - Start Device Bridge and monitor for scan events

## Requirements

### Linux
- Go 1.22+
- libusb-1.0-0 and libusb-1.0-0-dev
- udev rules configured (see docs/USB_SCANNER_VALIDATION.md)
- User in `plugdev` group

### macOS
- Go 1.22+
- libusb (via Homebrew: `brew install libusb`)
- Xcode Command Line Tools

### Windows
- Go 1.22+
- Administrator privileges
- WinUSB driver installed (via Zadig)

## Common Scanner VID/PID

| Manufacturer | Model | VID | PID |
|--------------|-------|-----|-----|
| Symbol/Zebra | LS2208 | 05e0 | 1200 |
| Symbol/Zebra | DS2208 | 05e0 | 1900 |
| Honeywell | Voyager 1200g | 0c2e | 0b61 |
| Datalogic | QuickScan QD2430 | 05f9 | 4206 |

To find your scanner's VID/PID:
```bash
# Linux
lsusb

# macOS
system_profiler SPUSBDataType

# Windows
Get-PnpDevice | Where-Object {$_.FriendlyName -like "*Scanner*"}
```

## Troubleshooting

See [docs/USB_SCANNER_VALIDATION.md](../../docs/USB_SCANNER_VALIDATION.md) for comprehensive troubleshooting guides.

## Future Scripts

Coming soon:
- `test-serial-scale.sh` - Serial scale validation (Week 6)
- `test-usb-printer.sh` - USB printer validation (Week 3)
- `test-payment-terminal.sh` - Payment terminal validation (Week 9)
- `run-all-tests.sh` - Comprehensive hardware test suite
