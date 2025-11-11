# USB Printer Setup Guide

Complete setup and configuration guide for USB ESC/POS printers with Device Bridge v2.

## Table of Contents

- [Overview](#overview)
- [Supported Printers](#supported-printers)
- [Requirements](#requirements)
- [Platform Setup](#platform-setup)
  - [Linux](#linux-setup)
  - [macOS](#macos-setup)
  - [Windows](#windows-setup)
- [Finding Printer USB IDs](#finding-printer-usb-ids)
- [Configuration](#configuration)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [Advanced Configuration](#advanced-configuration)

## Overview

Device Bridge v2 supports USB ESC/POS thermal printers through the `printer.usb` driver. This driver provides direct USB communication with thermal receipt printers, bypassing operating system printer drivers for maximum performance and control.

**Key Features:**
- Direct USB communication via USB Printer Class (0x07)
- Full ESC/POS command support
- Text formatting (bold, underline, size, alignment)
- Barcode printing (EAN, UPC, Code39, Code128, etc.)
- QR code printing
- Cash drawer control
- Status monitoring (online, paper, drawer)
- Cross-platform support (Linux, macOS, Windows)

## Supported Printers

The USB printer driver supports any thermal printer that implements the **ESC/POS protocol** and uses the **USB Printer Class** (0x07). This includes:

### Major Brands

**Epson** (Industry Standard)
- TM-T88 series (II, III, IV, V, VI) - Most popular receipt printer
- TM-T20 series - Budget-friendly option
- TM-U220 - Impact (dot matrix) printer
- TM-H6000 - Hybrid thermal/impact

**Star Micronics**
- TSP100/143 - Entry-level thermal
- TSP600/650/700 - High-performance
- TSP800 - Wide format

**Citizen**
- CT-S310/601/801 - Compact design
- CT-S4000 - High-speed

**Bixolon**
- SRP-350/350plus - Value option
- SRP-275/280 - Compact
- SRP-Q300 - Kitchen printer

**Others**
- Custom VKP80
- Posiflex PP-8000
- Most Chinese ESC/POS compatible printers

### Identifying Compatible Printers

A printer is compatible if:
1. It uses thermal or impact printing technology
2. It supports ESC/POS commands
3. It has a USB interface with Printer Class (0x07)
4. Marketed as "POS printer", "receipt printer", or "ESC/POS compatible"

## Requirements

### All Platforms

- Device Bridge v2
- USB cable (usually Type-B or Type-C)
- Thermal paper (usually 80mm width for receipts)
- Power supply for the printer

### Linux

- Go 1.22 or later
- libusb 1.0
- udev rules configured (for non-root access)
- User in `plugdev` group (recommended)

### macOS

- Go 1.22 or later
- libusb (via Homebrew)
- Xcode Command Line Tools
- May require running as root or granting USB permissions

### Windows

- Go 1.22 or later
- WinUSB driver (installed via Zadig)
- Administrator privileges for initial setup

## Platform Setup

### Linux Setup

#### 1. Install Dependencies

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install libusb-1.0-0 libusb-1.0-0-dev
```

**Fedora/RHEL:**
```bash
sudo dnf install libusb libusbx-devel
```

**Arch Linux:**
```bash
sudo pacman -S libusb
```

#### 2. Configure Permissions

**Option A: udev Rules (Recommended)**

Create a udev rule to allow non-root access:

```bash
sudo nano /etc/udev/rules.d/99-usb-printers.rules
```

Add rules for your printer(s):

```
# Epson printers
SUBSYSTEM=="usb", ATTR{idVendor}=="04b8", MODE="0666", GROUP="plugdev"

# Star Micronics printers
SUBSYSTEM=="usb", ATTR{idVendor}=="0519", MODE="0666", GROUP="plugdev"

# Citizen printers
SUBSYSTEM=="usb", ATTR{idVendor}=="2730", MODE="0666", GROUP="plugdev"

# Bixolon printers
SUBSYSTEM=="usb", ATTR{idVendor}=="1504", MODE="0666", GROUP="plugdev"

# Generic rule for all USB printers (Class 0x07)
SUBSYSTEM=="usb", ATTR{bDeviceClass}=="07", MODE="0666", GROUP="plugdev"
```

Reload udev rules:

```bash
sudo udevadm control --reload-rules
sudo udevadm trigger
```

**Option B: Add User to plugdev Group**

```bash
sudo usermod -a -G plugdev $USER
```

Log out and log back in for changes to take effect.

#### 3. Verify Setup

Check if printer is detected:

```bash
lsusb | grep -i printer
# or search by vendor name:
lsusb | grep -i epson
```

### macOS Setup

#### 1. Install Dependencies

Install libusb via Homebrew:

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install libusb
brew install libusb
```

#### 2. Install Xcode Command Line Tools

```bash
xcode-select --install
```

#### 3. Handle Built-in Drivers (Optional)

macOS may automatically load its own printer drivers. If you encounter conflicts:

```bash
# Temporarily disable built-in printer service
sudo launchctl unload -w /System/Library/LaunchDaemons/com.apple.printd.plist

# Re-enable when needed:
sudo launchctl load -w /System/Library/LaunchDaemons/com.apple.printd.plist
```

#### 4. Grant USB Permissions

For macOS 10.14+ (Mojave and later), you may need to:
1. System Preferences → Security & Privacy
2. Privacy tab → Full Disk Access
3. Add your application or Terminal

#### 5. Verify Setup

```bash
system_profiler SPUSBDataType | grep -A 10 "Printer"
```

### Windows Setup

#### 1. Install Zadig (WinUSB Driver)

Device Bridge uses libusb on Windows, which requires the WinUSB driver.

**Download and Install:**

1. Download Zadig from https://zadig.akeo.ie/
2. Run as Administrator
3. Options → List All Devices
4. Select your printer from the dropdown
5. Select "WinUSB" from the driver list
6. Click "Replace Driver" or "Install Driver"
7. Wait for installation to complete

**IMPORTANT:** After installing WinUSB, the printer will no longer work with Windows built-in printer drivers or standard Windows printing. This is for direct ESC/POS communication only.

#### 2. Verify Installation

Open Device Manager:
1. Press Win + X → Device Manager
2. Look for your printer under "Universal Serial Bus devices"
3. Should show "WinUSB" as the driver

#### 3. Reverting to Windows Driver (if needed)

To restore Windows printer functionality:
1. Open Device Manager
2. Right-click your printer → Update Driver
3. Browse → Let me pick from available drivers
4. Select the original manufacturer driver

## Finding Printer USB IDs

Every USB printer has a Vendor ID (VID) and Product ID (PID) that uniquely identify it.

### Linux

```bash
lsusb
```

Example output:
```
Bus 001 Device 005: ID 04b8:0e15 Seiko Epson Corp. TM-T88V
```

- **Vendor ID:** 0x04b8
- **Product ID:** 0x0e15

### macOS

```bash
system_profiler SPUSBDataType
```

Look for your printer and find:
```
Vendor ID: 0x04b8
Product ID: 0x0e15
```

### Windows

1. Open Device Manager (Win + X → Device Manager)
2. Find your printer (usually under "Printers" or "USB devices")
3. Right-click → Properties
4. Details tab → Hardware IDs

Example:
```
USB\VID_04B8&PID_0E15
```

- **Vendor ID:** 0x04b8
- **Product ID:** 0x0e15

## Configuration

### Basic Configuration

Create or edit your configuration file:

```yaml
# config.yaml
devices:
  - id: main-printer
    name: "Main Receipt Printer"
    kind: printer.usb
    enabled: true
    metadata:
      vendor_id: 0x04b8        # Your printer's vendor ID
      product_id: 0x0e15       # Your printer's product ID
      timeout: 5000            # USB timeout (ms)
```

### Multiple Printers

To use multiple printers of the same model, specify serial numbers:

```yaml
devices:
  - id: printer-1
    name: "Register 1 Printer"
    kind: printer.usb
    enabled: true
    metadata:
      vendor_id: 0x04b8
      product_id: 0x0e15
      serial: "ABC123"         # First printer's serial
      timeout: 5000

  - id: printer-2
    name: "Register 2 Printer"
    kind: printer.usb
    enabled: true
    metadata:
      vendor_id: 0x04b8
      product_id: 0x0e15
      serial: "XYZ789"         # Second printer's serial
      timeout: 5000
```

Find serial numbers using:
- **Linux:** `lsusb -v`
- **macOS:** `system_profiler SPUSBDataType`
- **Windows:** Device Manager → Properties → Details → Device Instance Path

### Advanced Configuration

```yaml
devices:
  - id: advanced-printer
    name: "Advanced Receipt Printer"
    kind: printer.usb
    enabled: true
    metadata:
      # USB Configuration
      vendor_id: 0x04b8
      product_id: 0x0e15
      timeout: 5000
      reconnect_delay: 5000

      # Printer Capabilities
      max_width: 48              # Characters per line
      dpi: 203                   # Dots per inch
      supports_drawer: true      # Cash drawer support

      # Application Metadata
      location: "Store 1, Register 3"
      department: "Sales"
      printer_type: "receipt"
```

## Testing

### 1. Start Device Bridge

```bash
./bin/device-bridge -config config.yaml
```

Check logs for:
```
INFO USB printer started successfully vendor_id=0x04b8 product_id=0x0e15
```

### 2. Verify Device Registration

Using bridge-cli (if available):
```bash
bridge-cli devices list
```

Using gRPC:
```bash
grpcurl -plaintext localhost:50051 devicebridge.v1.DeviceBridge/ListDevices
```

### 3. Print Test Receipt

Using gRPC:
```bash
grpcurl -plaintext -d '{
  "device_id": "main-printer",
  "document": {
    "sections": [{
      "elements": [{
        "type": "line",
        "line": {
          "alignment": "center",
          "runs": [{
            "text": "TEST RECEIPT",
            "style": {"bold": true, "double_height": true, "double_width": true}
          }]
        }
      }]
    }],
    "options": {"cut": true}
  }
}' localhost:50051 devicebridge.v1.DeviceBridge/Print
```

### 4. Test Cash Drawer

```bash
grpcurl -plaintext -d '{
  "device_id": "main-printer"
}' localhost:50051 devicebridge.v1.DeviceBridge/OpenDrawer
```

### 5. Check Status

```bash
grpcurl -plaintext -d '{
  "device_id": "main-printer"
}' localhost:50051 devicebridge.v1.DeviceBridge/GetPrinterStatus
```

## Troubleshooting

### "No matching USB printer found"

**Possible Causes:**
- Printer not powered on
- USB cable not connected
- Wrong vendor_id or product_id in config
- Permission issues (Linux)
- WinUSB driver not installed (Windows)

**Solutions:**
1. Verify printer is on and connected
2. Check USB IDs with `lsusb` or equivalent
3. On Linux: check udev rules and permissions
4. On Windows: install WinUSB via Zadig
5. Try running as root/administrator to isolate permission issues

### "Failed to set configuration" (Windows)

**Cause:** WinUSB driver not installed

**Solution:** Use Zadig to install WinUSB driver (see Windows Setup above)

### "Permission denied" (Linux)

**Cause:** User doesn't have USB access

**Solutions:**
1. Add user to plugdev group: `sudo usermod -a -G plugdev $USER`
2. Create udev rules (see Linux Setup above)
3. Temporarily run as root: `sudo ./device-bridge -config config.yaml`

### Printer prints garbled text

**Possible Causes:**
- Printer not in ESC/POS mode
- Printer using different protocol
- Firmware issues

**Solutions:**
1. Check printer manual for ESC/POS mode setting
2. Reset printer to factory defaults
3. Update printer firmware
4. Verify printer is actually ESC/POS compatible

### Cash drawer not opening

**Possible Causes:**
- Drawer not connected
- Wrong cable type
- Drawer lock engaged
- Printer doesn't support drawer

**Solutions:**
1. Verify drawer is connected to printer's RJ11/RJ12 port
2. Check cable connections
3. Test drawer manually (many have a key)
4. Verify `supports_drawer: true` in config
5. Some printers require specific drawer models

### Slow printing or timeouts

**Possible Causes:**
- USB cable quality
- USB hub issues
- Power supply problems
- Printer buffer full

**Solutions:**
1. Increase timeout in config: `timeout: 10000`
2. Use high-quality, short USB cable
3. Connect directly to computer (avoid USB hubs)
4. Check printer power supply
5. Disable USB power management:
   - Linux: `/etc/udev/rules.d/50-usb-power.rules`
   - Windows: Device Manager → USB → Power Management → Uncheck "Allow computer to turn off"

### "Device disconnected, will attempt reconnection"

**Possible Causes:**
- Loose USB connection
- Power issues
- USB power management
- Cable defect

**Solutions:**
1. Check USB cable and connections
2. Try different USB port
3. Check printer power supply
4. Disable USB power management
5. Increase `reconnect_delay` in config

### macOS: "Operation not permitted"

**Cause:** USB permission issues

**Solutions:**
1. Run as root: `sudo ./device-bridge`
2. Grant Full Disk Access in System Preferences
3. Disable System Integrity Protection (not recommended)

## Advanced Configuration

### Multiple Stations

```yaml
devices:
  # Front desk
  - id: front-desk-printer
    kind: printer.usb
    metadata:
      vendor_id: 0x04b8
      product_id: 0x0e15
      location: "Front Desk"

  # Kitchen
  - id: kitchen-printer
    kind: printer.usb
    metadata:
      vendor_id: 0x0519
      product_id: 0x0011
      location: "Kitchen"

  # Bar
  - id: bar-printer
    kind: printer.usb
    metadata:
      vendor_id: 0x0519
      product_id: 0x0011
      serial: "BAR001"
      location: "Bar"
```

### Auto-Discovery (Future Feature)

Device auto-discovery is planned for Week 4-5. This will automatically detect connected USB printers and add them to the configuration.

### Hot-Plugging

The USB printer driver supports automatic reconnection when a printer is disconnected and reconnected. Configure with:

```yaml
metadata:
  reconnect_delay: 5000    # Retry every 5 seconds
```

## Common USB IDs Reference

| Manufacturer    | Model              | Vendor ID | Product ID |
|-----------------|--------------------|-----------| ----------|
| Epson           | TM-T88V            | 0x04b8    | 0x0e15    |
| Epson           | TM-T20             | 0x04b8    | 0x0e03    |
| Epson           | TM-U220            | 0x04b8    | 0x0e01    |
| Star Micronics  | TSP100             | 0x0519    | 0x0001    |
| Star Micronics  | TSP650             | 0x0519    | 0x0011    |
| Star Micronics  | TSP700II           | 0x0519    | 0x0017    |
| Citizen         | CT-S310            | 0x2730    | 0x0fff    |
| Citizen         | CT-S801            | 0x2730    | 0x200f    |
| Bixolon         | SRP-350plus        | 0x1504    | 0x0011    |
| Bixolon         | SRP-275            | 0x1504    | 0x001d    |

For unlisted models, use `lsusb` or equivalent to find the USB IDs.

## Next Steps

- Review [API Documentation](API.md) for printing commands
- See [Configuration Examples](../configs/config.printer-usb.example.yaml)
- Check [Platform Compatibility Matrix](PLATFORM_COMPATIBILITY_MATRIX.md)
- Read [ESC/POS Protocol Reference](ESCPOS_PROTOCOL.md) (if available)

## Support

For issues and questions:
- GitHub Issues: https://github.com/Macber-eg/Flutter-Device/issues
- Documentation: https://github.com/Macber-eg/Flutter-Device/tree/main/docs
