# Zebra Badge Printer Setup Guide

**Comprehensive setup guide for Zebra card/badge printers with Device Bridge v2**

This guide covers hardware installation, driver configuration, and troubleshooting for Zebra card printers (ZC100, ZC300, ZC350, ZXP Series).

---

## Table of Contents

1. [Supported Hardware](#supported-hardware)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Testing](#testing)
6. [Badge Templates](#badge-templates)
7. [Encoding Options](#encoding-options)
8. [Troubleshooting](#troubleshooting)
9. [Best Practices](#best-practices)

---

## Supported Hardware

### Zebra Card Printers

| Model | Type | Features | Use Case |
|-------|------|----------|----------|
| **ZC100** | Entry-level | Single-sided, USB | Visitor badges, temp badges |
| **ZC300** | Mid-range | Dual-sided, USB/Network, Encoding | Conference badges, staff ID |
| **ZC350** | High-volume | Dual-sided, Encoding, Lamination | Employee badges, access cards |
| **ZXP Series 7** | Premium | Retransfer, High-quality, Encoding | VIP badges, government ID |

### Encoding Options

- **Magnetic Stripe**: Track 1, 2, 3 (HiCo/LoCo)
- **RFID/NFC**: Mifare 1K/4K, HID iClass, NTAG, Prox
- **Smart Card**: Contact/Contactless chips

---

## Prerequisites

### Software Requirements

**Linux:**
```bash
# Install libusb for USB connection
sudo apt-get update
sudo apt-get install libusb-1.0-0-dev pkg-config

# Verify USB device detection
lsusb | grep -i zebra
```

**macOS:**
```bash
# Install libusb via Homebrew
brew install libusb pkg-config

# Check USB devices
system_profiler SPUSBDataType | grep -i zebra
```

**Windows:**
```powershell
# Install Zebra card printer driver from:
# https://www.zebra.com/us/en/support-downloads.html

# Verify printer in Device Manager
Get-PnpDevice | Where-Object {$_.FriendlyName -like "*Zebra*"}
```

### Hardware Setup

1. **Unpack printer** and remove all protective materials
2. **Load ribbon** (YMCKO for color, K for monochrome)
3. **Load cards** into feeder (up to 100 cards)
4. **Connect power** and turn on printer
5. **Connect to computer**:
   - **USB**: Use provided USB cable
   - **Network**: Connect Ethernet cable, configure IP

### Network Configuration (TCP/IP Printers)

```bash
# Set static IP on printer (via LCD panel):
# Menu → Network → TCP/IP → Static IP
# IP Address: 192.168.1.100
# Subnet: 255.255.255.0
# Gateway: 192.168.1.1

# Test connectivity
ping 192.168.1.100
telnet 192.168.1.100 9100
```

---

## Installation

### 1. Clone and Build Device Bridge

```bash
git clone https://github.com/Macber-eg/Flutter-Device.git
cd Flutter-Device

# Install dependencies
go mod download

# Build
make build

# Or build directly
go build -o device-bridge ./cmd/bridge
```

### 2. USB Permissions (Linux only)

Create udev rule for Zebra printers:

```bash
sudo nano /etc/udev/rules.d/99-zebra-printers.rules
```

Add:
```
# Zebra ZC100
SUBSYSTEM=="usb", ATTR{idVendor}=="0a5f", ATTR{idProduct}=="0165", MODE="0666"

# Zebra ZC300
SUBSYSTEM=="usb", ATTR{idVendor}=="0a5f", ATTR{idProduct}=="0176", MODE="0666"

# Zebra ZC350
SUBSYSTEM=="usb", ATTR{idVendor}=="0a5f", ATTR{idProduct}=="0177", MODE="0666"

# Zebra ZXP Series 7
SUBSYSTEM=="usb", ATTR{idVendor}=="0a5f", ATTR{idProduct}=="0179", MODE="0666"
```

Reload rules:
```bash
sudo udevadm control --reload-rules
sudo udevadm trigger

# Verify
ls -l /dev/usb/lp*
```

### 3. Configure Device Bridge

Create configuration file:

```bash
cp configs/config.badge-printer.example.yaml configs/config.yaml
nano configs/config.yaml
```

**USB Configuration:**
```yaml
devices:
  - id: "badge-printer-01"
    name: "Main Badge Printer"
    kind: "badge_printer.zebra"
    enabled: true
    transport: "usb"
    vendor_id: 0x0a5f    # Zebra
    product_id: 0x0176   # ZC300

    settings:
      dual_sided: true
      magnetic_stripe: false
      rfid: false
      dpi: 300
      print_speed: "normal"
```

**Network (TCP) Configuration:**
```yaml
devices:
  - id: "badge-printer-01"
    name: "Main Badge Printer"
    kind: "badge_printer.zebra"
    enabled: true
    transport: "tcp"
    address: "192.168.1.100"
    port: 9100

    settings:
      dual_sided: true
      rfid: true
      rfid_types: ["mifare_1k", "iclass"]
      dpi: 600
      print_speed: "quality"
```

### 4. Start Device Bridge

```bash
./device-bridge --config configs/config.yaml

# Or with systemd
sudo systemctl start device-bridge
```

---

## Configuration

### Basic Settings

```yaml
settings:
  # Print quality
  dpi: 300              # 300 or 600
  print_speed: "normal" # fast, normal, quality

  # Capabilities (match your printer)
  dual_sided: true      # Can print on both sides
  magnetic_stripe: true # Has mag stripe encoder
  rfid: true            # Has RFID encoder
  lamination: true      # Has lamination module

  # RFID types supported
  rfid_types:
    - "mifare_1k"
    - "mifare_4k"
    - "iclass"
    - "ntag213"

  # Timeouts
  connect_timeout: "30s"
  print_timeout: "60s"
  encode_timeout: "30s"
```

### Advanced Settings

**Ribbon Types:**
- `YMCKO` - Full color (Yellow, Magenta, Cyan, Black, Overlay)
- `KO` - Black with overlay (monochrome)
- `K` - Black only (fastest)

**Card Types:**
- `CR80` - Standard credit card size (85.6 x 53.98 mm)
- `CR79` - Slightly smaller
- Custom sizes

---

## Testing

### 1. Check Printer Status

```bash
# Via REST API
curl http://localhost:8080/v1/devices/badge-printer-01/badge/status

# Expected response:
{
  "device_id": "badge-printer-01",
  "status": "DEVICE_STATUS_READY",
  "ribbon": {
    "type": "YMCKO",
    "panels_remaining": 450,
    "panels_capacity": 500,
    "percent_remaining": 90,
    "ribbon_installed": true
  },
  "cards_remaining": 95,
  "cards_capacity": 100
}
```

### 2. Print Test Badge

```bash
# Get available templates
curl http://localhost:8080/v1/devices/badge-printer-01/badge/templates

# Print conference attendee badge
curl -X POST http://localhost:8080/v1/devices/badge-printer-01/badge/print \
  -H "Content-Type: application/json" \
  -d '{
    "design": {
      "template_id": "conference_attendee",
      "variables": {
        "name": "John Doe",
        "company": "Acme Corp",
        "badge_id": "ATT-12345",
        "qr_data": "https://event.example.com/attendee/12345"
      }
    },
    "options": {
      "side": "PRINT_SIDE_FRONT_ONLY",
      "quality": "PRINT_QUALITY_NORMAL",
      "copies": 1
    }
  }'
```

### 3. Subscribe to Events

```bash
# Using curl with Server-Sent Events
curl -N http://localhost:8080/v1/devices/badge-printer-01/badge/events

# Or via WebSocket (JavaScript)
const ws = new WebSocket('ws://localhost:8080/ws');
ws.send(JSON.stringify({
  action: 'subscribe',
  device_id: 'badge-printer-01',
  event_type: 'badge_printer'
}));
```

---

## Badge Templates

### Available Templates

1. **conference_attendee** - Standard conference badge
   - Name, company, QR code, badge ID barcode
   - Variables: `name`, `company`, `badge_id`, `qr_data`

2. **vip_badge** - VIP badge with gold border
   - Name, title, access level
   - Variables: `name`, `title`, `access_level`, `qr_data`

3. **staff_badge** - Staff ID with photo
   - Photo, name, department, employee ID barcode
   - Variables: `name`, `department`, `employee_id`, `photo`, `position`

4. **visitor_badge** - Temporary visitor badge
   - Name, company, host, date, expiry
   - Variables: `name`, `company`, `host`, `date`, `expires`

### Using Templates

```json
{
  "design": {
    "template_id": "vip_badge",
    "variables": {
      "name": "Alice Johnson",
      "title": "Keynote Speaker",
      "access_level": "All Access",
      "qr_data": "VIP-987654"
    }
  }
}
```

### Custom Badge Design

```json
{
  "design": {
    "custom": {
      "orientation": "ORIENTATION_PORTRAIT",
      "background_color": "#FFFFFF",
      "elements": [
        {
          "type": "ELEMENT_TYPE_TEXT",
          "position": {"x": 100, "y": 100},
          "size": {"width": 800, "height": 60},
          "content": "My Custom Badge",
          "style": {
            "font_size": 40,
            "bold": true,
            "alignment": "TEXT_ALIGN_CENTER"
          }
        },
        {
          "type": "ELEMENT_TYPE_QR_CODE",
          "position": {"x": 400, "y": 300},
          "size": {"width": 200, "height": 200},
          "content": "https://example.com/verify/12345"
        }
      ]
    }
  }
}
```

---

## Encoding Options

### Magnetic Stripe Encoding

```json
{
  "encoding": {
    "magnetic_stripe": {
      "enabled": true,
      "track1_data": "%B1234567890123456^DOE/JOHN^2512101?",
      "track2_data": "1234567890123456=25121011234567890?",
      "coercivity": "COERCIVITY_TYPE_HICO"
    }
  }
}
```

**Track Formats:**
- **Track 1**: Alphanumeric, 79 chars, used for name/account
- **Track 2**: Numeric, 40 chars, used for account number
- **Track 3**: Numeric, 107 chars, rarely used

**Coercivity:**
- **HiCo** (High Coercivity 2750 Oe): More durable, harder to erase
- **LoCo** (Low Coercivity 300 Oe): Cheaper, easier to erase

### RFID/NFC Encoding

```json
{
  "encoding": {
    "rfid": {
      "enabled": true,
      "type": "RFID_TYPE_MIFARE_CLASSIC_1K",
      "uid": "01020304",
      "blocks": [
        {
          "block_number": 4,
          "data": "000102030405060708090A0B0C0D0E0F"
        },
        {
          "block_number": 5,
          "data": "FFEEDDCCBBAA99887766554433221100"
        }
      ]
    }
  }
}
```

**RFID Types:**
- **Mifare Classic 1K/4K**: Most common, 1KB/4KB storage
- **HID iClass**: Secure access control
- **NTAG**: NFC Forum Type 2, phone-compatible
- **Mifare DESFire**: High security, cryptography

---

## Troubleshooting

### Printer Not Detected

**USB:**
```bash
# Check if printer is recognized
lsusb | grep 0a5f

# Check permissions
ls -l /dev/usb/lp*

# Re-add udev rules
sudo udevadm control --reload-rules
```

**Network:**
```bash
# Ping printer
ping 192.168.1.100

# Check if port is open
telnet 192.168.1.100 9100
nc -zv 192.168.1.100 9100

# Check firewall
sudo ufw allow 9100/tcp
```

### Ribbon Errors

- **"No Ribbon"**: Install/reseat ribbon cartridge
- **"Wrong Ribbon"**: Check ribbon type matches card type
- **"Ribbon Low"**: Replace ribbon soon

### Card Jams

1. Open printer cover
2. Remove jammed card carefully
3. Check card thickness (should be CR80 30mil)
4. Check for bent/damaged cards
5. Clean rollers with cleaning card

### Print Quality Issues

- **Faded prints**: Replace ribbon
- **White spots**: Clean print head with cleaning pen
- **Smudges**: Use overlay panel on ribbon
- **Colors off**: Use YMCKO ribbon, check DPI setting

### Encoding Failures

**Magnetic Stripe:**
- Use HiCo cards with HiCo setting
- Ensure track data format is correct
- Clean mag stripe encoder

**RFID:**
- Use correct card type (Mifare, iClass, etc.)
- Check RFID antenna alignment
- Verify block numbers are valid
- Use proper access keys for secured chips

### Connection Lost

```bash
# Check status
systemctl status device-bridge

# Check logs
journalctl -u device-bridge -f

# Restart service
systemctl restart device-bridge
```

---

## Best Practices

### Card Handling

1. **Store cards properly**: Cool, dry place, away from magnetic fields
2. **Use quality cards**: CR80 30mil PVC cards
3. **Clean regularly**: Use cleaning cards every 1000 prints
4. **Load correctly**: Fan cards before loading

### Ribbon Management

1. **Match ribbon to cards**: YMCKO for color, K for monochrome
2. **Check capacity**: Replace when <10% remaining
3. **Store unused ribbons**: In sealed packaging

### Maintenance Schedule

- **Daily**: Check ribbon/card levels
- **Weekly**: Visual inspection, clean exterior
- **Monthly**: Deep cleaning with cleaning kit
- **Quarterly**: Professional service check

### Security

1. **Restrict physical access** to printer
2. **Enable API authentication** in Device Bridge
3. **Use HTTPS** for network communication
4. **Encrypt RFID data** on cards
5. **Log all print jobs** for audit trail

### Performance Optimization

1. **Batch printing**: Print multiple badges at once
2. **Use appropriate quality**: "fast" for drafts, "quality" for final
3. **Network stability**: Use wired Ethernet, not WiFi
4. **Pre-load templates**: Cache frequently used designs

---

## API Reference

### Print Badge

```http
POST /v1/devices/{device_id}/badge/print
Content-Type: application/json

{
  "design": {...},
  "encoding": {...},
  "options": {...}
}
```

### Get Status

```http
GET /v1/devices/{device_id}/badge/status
```

### Get Templates

```http
GET /v1/devices/{device_id}/badge/templates
```

### Encode Card

```http
POST /v1/devices/{device_id}/badge/encode
Content-Type: application/json

{
  "encoding": {...}
}
```

### Batch Print

```http
POST /v1/devices/{device_id}/badge/batch
Content-Type: application/json

{
  "badges": [
    {"badge_id": "1", "design": {...}},
    {"badge_id": "2", "design": {...}}
  ]
}
```

---

## Support

### Resources

- **Zebra Support**: https://www.zebra.com/support
- **Device Bridge Docs**: https://github.com/Macber-eg/Flutter-Device
- **ZPL Programming Guide**: https://www.zebra.com/zpl

### Getting Help

1. Check this documentation
2. Search GitHub Issues
3. Contact Zebra technical support
4. Open GitHub issue with:
   - Printer model
   - Device Bridge version
   - Error logs
   - Configuration file (sanitized)

---

**Last Updated**: 2025-11-12
**Document Version**: 1.0
