# RFID/NFC Reader - Setup and Configuration Guide

**Complete guide for integrating RFID/NFC readers with Device Bridge v2**

Support for 125 kHz proximity cards, 13.56 MHz contactless cards, and NFC tags for event access control and authentication.

---

## Table of Contents

1. [Overview](#overview)
2. [Supported Hardware](#supported-hardware)
3. [Prerequisites](#prerequisites)
4. [Installation](#installation)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Card Types and Protocols](#card-types-and-protocols)
8. [API Reference](#api-reference)
9. [Best Practices](#best-practices)
10. [Integration Patterns](#integration-patterns)

---

## Overview

Device Bridge v2 provides comprehensive RFID/NFC reader support for:

- **Access Control**: Employee badges, visitor management
- **Event Management**: Conference check-in, session tracking
- **Payment Systems**: Transit cards, cashless payments
- **Asset Tracking**: Inventory management, equipment tracking
- **Authentication**: Multi-factor authentication, secure access

### Supported Technologies

- **125 kHz Proximity**: HID Prox, EM4100, Indala
- **13.56 MHz Contactless**: Mifare Classic, iClass, DESFire
- **NFC Tags**: NTAG213/215/216, Mifare Ultralight
- **Regional**: FeliCa (Japan transit/payment)

---

## Supported Hardware

### USB RFID Readers

| Manufacturer | Model | VID:PID | Protocols | Notes |
|--------------|-------|---------|-----------|-------|
| ACS | ACR122U | 072f:2200 | Mifare, NTAG, DESFire, FeliCa | Most popular NFC reader |
| HID Global | ProxPoint Plus | 076b:5340 | HID Prox | 125 kHz proximity |
| OmniKey | CardMan 5121 | 076b:5121 | iClass, Mifare | High-security |
| Sony | PaSoRi RC-S380 | 054c:06c3 | FeliCa | Japan market |
| Identiv | uTrust 3700 F | 04e6:5790 | Mifare, NTAG, DESFire | Dual interface |

### Network RFID Readers

- **TCP/IP Readers**: Connect via Ethernet (port 10001 typical)
- **PoE Readers**: Power over Ethernet support
- **Wiegand Output**: Traditional access control integration

### Serial RFID Readers

- **RS-232/RS-485**: Industrial applications
- **TTL Serial**: Embedded systems
- **Baud Rates**: 9600-115200 typical

---

## Prerequisites

### Linux

```bash
# Install libusb for USB readers
sudo apt-get update
sudo apt-get install libusb-1.0-0-dev libudev-dev

# Install libpcsclite for PC/SC readers
sudo apt-get install libpcsclite-dev pcscd

# For serial readers
sudo apt-get install setserial

# Add user to dialout group (for serial/USB access)
sudo usermod -a -G dialout $USER

# Create udev rules for USB readers
sudo nano /etc/udev/rules.d/99-rfid-readers.rules
```

**udev rules** (`/etc/udev/rules.d/99-rfid-readers.rules`):

```
# ACS ACR122U
SUBSYSTEM=="usb", ATTR{idVendor}=="072f", ATTR{idProduct}=="2200", MODE="0666"

# HID ProxPoint Plus
SUBSYSTEM=="usb", ATTR{idVendor}=="076b", ATTR{idProduct}=="5340", MODE="0666"

# OmniKey CardMan 5121
SUBSYSTEM=="usb", ATTR{idVendor}=="076b", ATTR{idProduct}=="5121", MODE="0666"

# Sony PaSoRi RC-S380
SUBSYSTEM=="usb", ATTR{idVendor}=="054c", ATTR{idProduct}=="06c3", MODE="0666"
```

Reload udev rules:

```bash
sudo udevadm control --reload-rules
sudo udevadm trigger
```

### macOS

```bash
# Install libusb via Homebrew
brew install libusb

# For PC/SC readers (built-in support)
# No additional packages needed

# For serial readers
brew install minicom
```

### Windows

```bash
# Install libusb driver using Zadig
# Download from: https://zadig.akeo.ie/

# For PC/SC readers
# Windows has built-in Smart Card service
# Ensure service is running:
sc query SCardSvr

# For serial readers
# Install appropriate driver from manufacturer
```

---

## Installation

### 1. Verify Reader Connection

**USB Readers**:

```bash
# Linux/macOS
lsusb | grep -i "072f\|076b\|054c"

# Expected output (ACR122U):
# Bus 001 Device 005: ID 072f:2200 ACS ACR122U

# Windows (PowerShell)
Get-PnpDevice | Where-Object {$_.Class -eq "SmartCardReader"}
```

**Network Readers**:

```bash
# Test TCP connection
ping 192.168.1.50
telnet 192.168.1.50 10001

# Or with netcat
nc -zv 192.168.1.50 10001
```

**Serial Readers**:

```bash
# List serial ports (Linux)
ls -l /dev/ttyUSB* /dev/ttyACM*

# List serial ports (macOS)
ls -l /dev/tty.usbserial*

# List serial ports (Windows)
mode
```

### 2. Configure Device Bridge

Create configuration file `config.yaml`:

```yaml
server:
  grpc_port: 50051
  http_port: 8080

devices:
  - id: "rfid-reader-01"
    name: "Main Entrance Reader"
    kind: "rfid_reader.acr122u"
    enabled: true
    transport: "usb"
    vendor_id: 0x072f
    product_id: 0x2200

    settings:
      supported_protocols:
        - "mifare"
        - "ntag"

      mode: "continuous"
      beep_on_detect: true
      auto_read: true
      read_interval_ms: 500
```

### 3. Start Device Bridge

```bash
# Start Device Bridge
./device-bridge --config config.yaml

# Verify reader is connected
curl http://localhost:8080/v1/devices/rfid-reader-01/status
```

---

## Configuration

### Basic USB Reader

```yaml
devices:
  - id: "rfid-usb-01"
    name: "USB NFC Reader"
    kind: "rfid_reader.usb"
    enabled: true
    transport: "usb"
    vendor_id: 0x072f
    product_id: 0x2200

    settings:
      supported_protocols:
        - "mifare"
        - "ntag"

      can_read: true
      can_write: true
      has_buzzer: true
      has_led: true

      mode: "continuous"
      beep_on_detect: true
      read_interval_ms: 500
      read_timeout_sec: 30
```

### Network Reader (TCP)

```yaml
devices:
  - id: "rfid-network-01"
    name: "Network RFID Reader"
    kind: "rfid_reader.network"
    enabled: true
    transport: "tcp"
    address: "192.168.1.50"
    port: 10001

    settings:
      supported_protocols:
        - "mifare"
        - "iclass"

      mode: "continuous"
      beep_on_detect: true
      auto_read: true
```

### Serial Reader (RS-232)

```yaml
devices:
  - id: "rfid-serial-01"
    name: "Serial RFID Reader"
    kind: "rfid_reader.serial"
    enabled: true
    transport: "serial"
    serial_dev: "/dev/ttyUSB0"
    baud_rate: 115200

    settings:
      supported_protocols:
        - "prox"

      mode: "continuous"
      read_interval_ms: 500
```

### High-Security Configuration (with Authentication)

```yaml
devices:
  - id: "rfid-secure-01"
    name: "Secure Access Reader"
    kind: "rfid_reader.secure"
    enabled: true
    transport: "usb"
    vendor_id: 0x076b
    product_id: 0x5121

    settings:
      supported_protocols:
        - "iclass"
        - "desfire"

      can_authenticate: true

      # Custom authentication keys
      default_mifare_key_a: "A0A1A2A3A4A5"
      default_mifare_key_b: "B0B1B2B3B4B5"
      default_iclass_key: "SECUREKEY001"

      mode: "continuous"
      beep_on_detect: true
```

---

## Usage Examples

### 1. Read a Card (Blocking)

**HTTP API**:

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/read \
  -H "Content-Type: application/json" \
  -d '{
    "timeout_seconds": 30
  }'
```

**Response**:

```json
{
  "card": {
    "type": "CARD_TYPE_MIFARE_CLASSIC_1K",
    "uid": "01020304",
    "serial_number": "01020304",
    "technology": "CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ",
    "mifare": {
      "size_bytes": 1024,
      "num_sectors": 16,
      "sak": 8,
      "atqa": 4
    },
    "capabilities": {
      "writable": true,
      "requires_auth": true
    }
  },
  "device_id": "rfid-reader-01",
  "timestamp": "2025-11-12T10:30:00Z",
  "signal_strength": 95
}
```

### 2. Subscribe to Card Reads (Streaming)

**HTTP SSE**:

```bash
curl -N http://localhost:8080/v1/devices/rfid-reader-01/subscribe
```

**Stream Output**:

```
event: card_read
data: {"card": {"type": "CARD_TYPE_MIFARE_CLASSIC_1K", "uid": "01020304"}, "timestamp": "2025-11-12T10:30:00Z"}

event: card_read
data: {"card": {"type": "CARD_TYPE_NTAG_213", "uid": "04AABBCCDD"}, "timestamp": "2025-11-12T10:30:15Z"}
```

### 3. Read Card with Full Data

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/read \
  -H "Content-Type: application/json" \
  -d '{
    "options": {
      "read_full_data": true,
      "auth_keys": ["FFFFFFFFFFFF"],
      "beep_on_read": true
    }
  }'
```

### 4. Write Data to Card

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/write \
  -H "Content-Type: application/json" \
  -d '{
    "card_uid": "01020304",
    "blocks": [
      {
        "block_number": 4,
        "data": "48656C6C6F20576F726C6421"
      }
    ],
    "credentials": {
      "mifare_keys_a": ["FFFFFFFFFFFF"]
    },
    "options": {
      "verify": true,
      "beep_on_success": true
    }
  }'
```

### 5. Authenticate Card

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/authenticate \
  -H "Content-Type: application/json" \
  -d '{
    "card_uid": "01020304",
    "method": "AUTH_METHOD_MIFARE_KEY",
    "credentials": {
      "mifare_keys_a": ["A0A1A2A3A4A5"]
    }
  }'
```

**Response**:

```json
{
  "authenticated": true,
  "credential_info": {
    "user_id": "EMP001",
    "user_name": "John Doe",
    "access_level": "employee",
    "valid_until": "2026-12-31T23:59:59Z"
  },
  "timestamp": "2025-11-12T10:30:00Z"
}
```

### 6. Get Reader Status

```bash
curl http://localhost:8080/v1/devices/rfid-reader-01/status
```

**Response**:

```json
{
  "device_id": "rfid-reader-01",
  "status": "DEVICE_STATUS_READY",
  "mode": "READER_MODE_CONTINUOUS",
  "supported_types": [
    "CARD_TYPE_MIFARE_CLASSIC_1K",
    "CARD_TYPE_MIFARE_CLASSIC_4K",
    "CARD_TYPE_NTAG_213"
  ],
  "capabilities": {
    "can_read": true,
    "can_write": true,
    "can_authenticate": true,
    "has_buzzer": true,
    "has_led": true
  },
  "card_present": false,
  "model": "ACR122U",
  "manufacturer": "ACS",
  "statistics": {
    "total_reads": 1523,
    "successful_reads": 1520,
    "failed_reads": 3,
    "avg_read_time_ms": 45
  }
}
```

### 7. Set Reader Mode

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/mode \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "READER_MODE_SINGLE_READ",
    "options": {
      "beep_on_detect": true,
      "auto_read": true
    }
  }'
```

---

## Card Types and Protocols

### 125 kHz Proximity Cards

#### HID Prox

**Characteristics**:
- Read-only
- 26-bit Wiegand format (facility code + card number)
- 4-8 byte UID
- No encryption

**Use Cases**:
- Basic access control
- Time & attendance
- Legacy systems

**Example**:

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-prox-01/read
```

```json
{
  "card": {
    "type": "CARD_TYPE_HID_PROX",
    "uid": "0A123456",
    "prox": {
      "facility_code": 10,
      "card_number": 4660,
      "wiegand_bits": 26,
      "format_name": "26-bit Wiegand"
    }
  }
}
```

### 13.56 MHz Contactless Cards

#### Mifare Classic

**Characteristics**:
- 1K (1024 bytes) or 4K (4096 bytes)
- 16 sectors (1K) or 40 sectors (4K)
- Requires authentication (Key A or Key B)
- Read/write

**Use Cases**:
- Conference badges
- Employee access cards
- Transit cards

**Memory Structure**:
- Sector 0: Manufacturer data (read-only)
- Sectors 1-15: User data (1K) or 1-39 (4K)
- Each sector: 4 blocks × 16 bytes

**Example - Read with Authentication**:

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/read \
  -d '{
    "options": {
      "read_full_data": true,
      "authenticate": true,
      "auth_keys": ["FFFFFFFFFFFF"],
      "sectors": [1, 2, 3]
    }
  }'
```

#### HID iClass

**Characteristics**:
- 2KB-32KB memory
- Encrypted (3DES)
- High security
- Read/write with authentication

**Use Cases**:
- High-security access
- Government facilities
- Financial institutions

**Example**:

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-iclass-01/authenticate \
  -d '{
    "card_uid": "0102030405060708",
    "method": "AUTH_METHOD_ICLASS_KEY",
    "credentials": {
      "iclass_key": "SECUREKEY001"
    }
  }'
```

#### Mifare DESFire

**Characteristics**:
- 2KB-8KB memory
- Multiple applications per card
- AES/3DES encryption
- File-based storage

**Use Cases**:
- Multi-application cards (access + payment)
- Secure transit systems
- Campus cards

**Example - Read Applications**:

```json
{
  "card": {
    "type": "CARD_TYPE_MIFARE_DESFIRE",
    "desfire": {
      "applications": [
        {
          "aid": 1,
          "name": "Access Control",
          "files": [
            {"file_id": 0, "file_type": "standard", "size": 32}
          ]
        },
        {
          "aid": 2,
          "name": "Payment",
          "files": [
            {"file_id": 0, "file_type": "value", "size": 4}
          ]
        }
      ]
    }
  }
}
```

### NFC Tags

#### NTAG (213/215/216)

**Characteristics**:
- NDEF-compatible
- 144-888 bytes user memory
- Optional password protection
- Read/write

**Use Cases**:
- NFC stickers/tags
- Smart posters
- Product authentication

**NTAG Variants**:
- **NTAG213**: 144 bytes user memory
- **NTAG215**: 504 bytes user memory
- **NTAG216**: 888 bytes user memory

**Example - Write NDEF Record**:

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/write \
  -d '{
    "ndef_records": [
      {
        "type": "NDEF_RECORD_TYPE_URI",
        "uri": "https://conference.example.com/attendee/12345"
      },
      {
        "type": "NDEF_RECORD_TYPE_TEXT",
        "text": "John Doe - VIP Attendee"
      }
    ]
  }'
```

#### FeliCa

**Characteristics**:
- Popular in Japan
- Used for transit (Suica, PASMO)
- Service/block structure
- Fast transaction speed

**Use Cases**:
- Japanese transit systems
- Payment systems
- Access control

**Example**:

```json
{
  "card": {
    "type": "CARD_TYPE_FELICA",
    "felica": {
      "idm": "0102030405060708",
      "system_code": 35012,
      "services": [
        {"service_code": 4112, "blocks": [...]}
      ]
    }
  }
}
```

---

## API Reference

### gRPC Service

```protobuf
service RFIDReader {
  rpc ReadCard(ReadCardRequest) returns (ReadCardResponse);
  rpc SubscribeCardReads(SubscribeCardReadsRequest) returns (stream CardReadEvent);
  rpc WriteCard(WriteCardRequest) returns (WriteCardResponse);
  rpc AuthenticateCard(AuthenticateCardRequest) returns (AuthenticateCardResponse);
  rpc GetReaderStatus(GetReaderStatusRequest) returns (ReaderStatus);
  rpc SetReaderMode(SetReaderModeRequest) returns (SetReaderModeResponse);
}
```

### REST API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/devices/{id}/read` | Read a card (blocking) |
| GET | `/v1/devices/{id}/subscribe` | Subscribe to card reads (SSE) |
| POST | `/v1/devices/{id}/write` | Write data to card |
| POST | `/v1/devices/{id}/authenticate` | Authenticate card |
| GET | `/v1/devices/{id}/status` | Get reader status |
| POST | `/v1/devices/{id}/mode` | Set reader mode |

---

## Best Practices

### 1. Reader Placement

**Optimal Distance**:
- **125 kHz Prox**: 2-10 cm
- **13.56 MHz/NFC**: 0-10 cm
- **Long-range readers**: Up to 1 meter

**Environmental Factors**:
- Avoid metal surfaces (interference)
- Keep away from other RF devices
- Maintain stable power supply
- Consider ambient temperature

### 2. Security

**Authentication Keys**:
```yaml
# Change default Mifare keys
default_mifare_key_a: "A0A1A2A3A4A5"  # NOT factory default
default_mifare_key_b: "B0B1B2B3B4B5"

# Rotate keys periodically
key_rotation_days: 90
```

**Key Management**:
- Never use factory default keys in production
- Store keys securely (environment variables, secrets manager)
- Implement key rotation
- Use different keys per sector

**Access Control**:
```yaml
security:
  api_auth:
    enabled: true
    api_keys:
      - "production-key-random-string"
  tls:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
```

### 3. Performance Optimization

**Read Interval**:
```yaml
# Fast polling (100ms) - high CPU usage
read_interval_ms: 100

# Normal polling (500ms) - recommended
read_interval_ms: 500

# Slow polling (1000ms) - low traffic areas
read_interval_ms: 1000
```

**Connection Pooling** (for multiple readers):
```yaml
# Use separate readers for high-traffic areas
devices:
  - id: "entrance-1"
    location: "Main Entrance"
  - id: "entrance-2"
    location: "Side Entrance"
```

### 4. Error Handling

**Timeouts**:
```yaml
settings:
  read_timeout_sec: 30   # Wait up to 30 seconds for card
  write_timeout_sec: 10  # Write operations timeout
```

**Retry Logic** (application-level):
```python
import time

def read_card_with_retry(device_id, max_retries=3):
    for attempt in range(max_retries):
        try:
            return api.read_card(device_id, timeout=10)
        except TimeoutError:
            if attempt < max_retries - 1:
                time.sleep(1)
            else:
                raise
```

### 5. Logging and Monitoring

**Enable Debug Logging**:
```yaml
logging:
  level: "debug"
  format: "json"
```

**Monitor Metrics**:
```bash
# Prometheus metrics endpoint
curl http://localhost:9090/metrics | grep rfid

# Example metrics:
# rfid_total_reads{device="rfid-reader-01"} 1523
# rfid_successful_reads{device="rfid-reader-01"} 1520
# rfid_failed_reads{device="rfid-reader-01"} 3
# rfid_avg_read_time_ms{device="rfid-reader-01"} 45
```

---

## Integration Patterns

### Pattern 1: Access Control System

```python
import requests
import json

def check_access(device_id):
    """Poll for card and authenticate"""

    # Read card
    response = requests.post(
        f"http://localhost:8080/v1/devices/{device_id}/read",
        json={"timeout_seconds": 30}
    )

    if response.status_code == 200:
        card = response.json()["card"]

        # Authenticate against database
        auth_response = requests.post(
            f"http://localhost:8080/v1/devices/{device_id}/authenticate",
            json={
                "card_uid": card["uid"],
                "method": "AUTH_METHOD_DATABASE",
                "credentials": {
                    "database": {
                        "connection_string": "postgresql://...",
                        "query": "SELECT * FROM credentials WHERE uid = $1"
                    }
                }
            }
        )

        if auth_response.json()["authenticated"]:
            grant_access(card)
        else:
            deny_access(card)
```

### Pattern 2: Event Check-In

```javascript
// Subscribe to card reads
const eventSource = new EventSource('http://localhost:8080/v1/devices/rfid-reader-01/subscribe');

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);

    if (data.card) {
        checkInAttendee(data.card.uid);
    }
};

function checkInAttendee(uid) {
    // Mark attendee as checked in
    fetch('/api/check-in', {
        method: 'POST',
        body: JSON.stringify({
            uid: uid,
            timestamp: new Date().toISOString()
        })
    });
}
```

### Pattern 3: Badge Printing + RFID Encoding

```python
def issue_badge(attendee_data):
    """Print badge and encode RFID card"""

    # 1. Print badge (using badge printer driver)
    print_response = requests.post(
        "http://localhost:8080/v1/devices/badge-printer-01/print",
        json={
            "design": {
                "template_id": "conference_attendee",
                "variables": attendee_data
            }
        }
    )

    # 2. Wait for card presentation
    read_response = requests.post(
        "http://localhost:8080/v1/devices/rfid-reader-01/read",
        json={"timeout_seconds": 60}
    )

    card_uid = read_response.json()["card"]["uid"]

    # 3. Write attendee data to card
    write_response = requests.post(
        "http://localhost:8080/v1/devices/rfid-reader-01/write",
        json={
            "card_uid": card_uid,
            "blocks": [
                {
                    "block_number": 4,
                    "data": encode_attendee_data(attendee_data)
                }
            ],
            "credentials": {
                "mifare_keys_a": ["CONFERENCE2025"]
            }
        }
    )

    return card_uid
```

---

**Last Updated**: 2025-11-12
**Document Version**: 1.0
