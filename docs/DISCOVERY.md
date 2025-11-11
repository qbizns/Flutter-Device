# Device Auto-Discovery

Device Bridge v2 provides comprehensive auto-discovery for USB, serial, and network devices. The discovery system automatically detects connected devices and can optionally register them for use.

## Table of Contents

- [Overview](#overview)
- [Supported Transports](#supported-transports)
- [Configuration](#configuration)
- [Discovery Events](#discovery-events)
- [USB Discovery](#usb-discovery)
- [Serial Discovery](#serial-discovery)
- [Network Discovery](#network-discovery)
- [Manual Discovery](#manual-discovery)
- [Troubleshooting](#troubleshooting)

## Overview

The discovery system consists of:

1. **Transport Scanners** - Scan specific transports (USB, Serial, mDNS, TCP)
2. **Discovery Manager** - Coordinates scanners, publishes events
3. **Background Worker** - Continuous scanning at configured intervals
4. **Event System** - Publishes discovery/removal events
5. **Auto-Registration** - Optionally auto-registers discovered devices

### Key Features

- Multi-transport discovery (USB, Serial, Network)
- Protocol auto-detection for serial scales
- Vendor/product database for device identification
- Configurable scan intervals
- Event-driven architecture
- Duplicate detection
- Stale device removal
- Hot-plug support

## Supported Transports

| Transport | Discovers | Auto-Detect | Notes |
|-----------|-----------|-------------|-------|
| **USB** | Printers, Scanners | Yes | VID/PID matching |
| **Serial** | Scales | Yes | Protocol probing |
| **mDNS** | Network Printers | Yes | Bonjour/Zeroconf |
| **TCP** | Network Devices | No | Manual subnet scan |

## Configuration

### Basic Configuration

```yaml
discovery:
  enabled: true
  scan_interval: 30s        # Scan every 30 seconds
  auto_register: false      # Don't auto-register (manual approval)

  # Enable specific transports
  transports:
    - usb
    - serial
    - mdns
```

### Advanced Configuration

```yaml
discovery:
  enabled: true
  scan_interval: 30s
  auto_register: true       # Automatically register devices

  transports:
    - usb
    - serial
    - mdns
    - tcp

  # USB Discovery
  usb:
    discover_printers: true
    discover_scanners: true

  # Serial Discovery
  serial:
    discover_scales: true
    baud_rates:           # Baud rates to try
      - 9600
      - 19200
      - 4800
    probe_timeout: 2s

  # mDNS Discovery
  mdns:
    services:             # Service types to discover
      - "_ipp._tcp"       # IPP printers
      - "_printer._tcp"   # Generic printers
    timeout: 3s

  # TCP Discovery (manual subnet scan)
  tcp:
    subnets:
      - "192.168.1.0/24"
    ports:
      - 9100              # Raw printing
      - 515               # LPD
```

### Production Configuration

For production environments, use longer scan intervals to reduce system load:

```yaml
discovery:
  enabled: true
  scan_interval: 5m         # Scan every 5 minutes
  auto_register: false      # Manual approval for security

  transports:
    - usb                   # USB only for POS stations

  usb:
    discover_printers: true
    discover_scanners: true
```

## Discovery Events

The discovery system publishes events when devices are found or removed.

### Event Types

| Event Type | Description |
|------------|-------------|
| `discovery.discovered` | New device found |
| `discovery.removed` | Device no longer detected (stale) |

### Event Data

```json
{
  "device_id": "usb-printer-04b8-0e15-ABC123",
  "name": "Epson TM-T88V",
  "kind": "printer.usb",
  "transport": "usb",
  "address": "",
  "vendor_id": 1208,
  "product_id": 3605
}
```

### Subscribing to Events

Via gRPC:

```bash
grpcurl -plaintext -d '{
  "device_id": "*"
}' localhost:50051 devicebridge.v1.DeviceBridge/SubscribeEvents
```

Via WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.send(JSON.stringify({
  type: 'subscribe',
  device_id: '*',
  event_types: ['discovery.discovered', 'discovery.removed']
}));

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Discovery event:', data);
};
```

## USB Discovery

Auto-discovers USB printers and HID scanners by enumerating USB devices.

### How It Works

1. Enumerate all USB devices
2. Filter by USB class (0x07 for printers, 0x03 for HID)
3. Match against known vendor/product IDs
4. Extract device information (vendor, model, serial)
5. Publish discovery event

### Supported Devices

**Printers:**
- Epson (TM-T88 series, TM-T20, TM-U220)
- Star Micronics (TSP series)
- Citizen (CT-S series)
- Bixolon (SRP series)
- Generic ESC/POS printers

**Scanners:**
- Symbol/Zebra (LS2208, DS2208, DS9208)
- Honeywell (Voyager, Xenon series)
- Datalogic (QuickScan series)
- Generic HID POS scanners

### Configuration

```yaml
discovery:
  transports:
    - usb

  usb:
    discover_printers: true
    discover_scanners: true
```

### Example Discovered Device

```json
{
  "id": "usb-printer-04b8-0e15",
  "name": "Epson TM-T88V",
  "kind": "printer.usb",
  "transport": "usb",
  "vendor_id": 1208,
  "product_id": 3605,
  "serial": "ABC123",
  "metadata": {
    "vendor": "Epson",
    "model": "TM-T88V",
    "bus": "1",
    "address": "5"
  }
}
```

## Serial Discovery

Auto-discovers serial scales by probing ports with different protocols.

### How It Works

1. Enumerate all serial ports (COM/tty)
2. Skip virtual ports (Bluetooth)
3. Try each port with common baud rates (9600, 19200, 4800)
4. Probe with each protocol (MT-SICS, CAS, Dibal, Toledo)
5. Identify protocol from response format
6. Publish discovery event

### Supported Protocols

| Protocol | Manufacturer | Detection Method |
|----------|--------------|------------------|
| **MT-SICS** | Mettler Toledo | Send "SI\r\n", check for status + weight + unit |
| **CAS** | CAS Corporation | Send STX W ETX, check for STX status 6-digits ETX |
| **Dibal** | Dibal | Read continuous output, check for STX +/- weight ETX |
| **Toledo** | Toledo | Read continuous output, check for S/D status format |

### Configuration

```yaml
discovery:
  transports:
    - serial

  serial:
    discover_scales: true
    baud_rates:
      - 9600
      - 19200
      - 4800
      - 38400
    probe_timeout: 2s
```

### Example Discovered Device

```json
{
  "id": "serial-dev-ttyUSB0",
  "name": "Mettler Toledo Scale on /dev/ttyUSB0",
  "kind": "scale.serial",
  "transport": "serial",
  "address": "/dev/ttyUSB0",
  "metadata": {
    "port": "/dev/ttyUSB0",
    "baud_rate": "9600",
    "protocol": "mtsics"
  }
}
```

## Network Discovery

Auto-discovers network devices via mDNS (Bonjour/Zeroconf).

### How It Works

1. Query mDNS for specific service types
2. Collect service announcements (name, hostname, IP, port)
3. Extract device information from TXT records
4. Publish discovery event

### Supported Services

| Service Type | Description | Typical Devices |
|-------------|-------------|-----------------|
| `_ipp._tcp` | Internet Printing Protocol | Network printers |
| `_printer._tcp` | Generic printer service | HP, Canon, Epson printers |
| `_pdl-datastream._tcp` | HP PDL | HP printers |
| `_http._tcp` | HTTP service | Web-enabled devices |

### Configuration

```yaml
discovery:
  transports:
    - mdns

  mdns:
    services:
      - "_ipp._tcp"
      - "_printer._tcp"
      - "_pdl-datastream._tcp"
    timeout: 3s
```

### Example Discovered Device

```json
{
  "id": "mdns-HP-LaserJet-hp-laserjet.local",
  "name": "HP LaserJet",
  "kind": "printer.ipp",
  "transport": "network",
  "address": "192.168.1.100",
  "port": 631,
  "metadata": {
    "hostname": "hp-laserjet.local",
    "service": "_ipp._tcp"
  }
}
```

## Manual Discovery

Trigger one-time discovery scans via API or CLI.

### Via gRPC

```bash
grpcurl -plaintext localhost:50051 devicebridge.v1.DeviceBridge/DiscoverDevices
```

### Via REST API

```bash
curl http://localhost:8080/v1/discovery/scan
```

### Via CLI (future)

```bash
bridge-cli devices discover
bridge-cli devices discover --usb
bridge-cli devices discover --serial
bridge-cli devices discover --network
```

## Troubleshooting

### No Devices Discovered

**USB Devices:**
1. Check USB cable and power
2. Verify permissions (Linux: udev rules, plugdev group)
3. Check vendor_id/product_id with `lsusb`
4. Ensure WinUSB driver installed (Windows)

**Serial Devices:**
1. Check serial cable connections
2. Verify port permissions (Linux: dialout group)
3. Try manual serial communication to test port
4. Check baud rate configuration

**Network Devices:**
1. Ensure devices are on same network/VLAN
2. Check firewall rules (mDNS uses UDP 5353)
3. Verify mDNS/Bonjour is enabled on devices
4. Test with `dns-sd` (macOS) or `avahi-browse` (Linux)

### Discovery Too Slow

1. Increase `scan_interval` to reduce frequency
2. Disable unused transports
3. Reduce baud_rate list for serial discovery
4. Reduce mDNS timeout

### Devices Keep Disappearing

1. Check for intermittent USB connections
2. Increase stale threshold (currently 5 minutes)
3. Check power management settings (USB suspend)
4. For serial: ensure continuous connection

### False Positives

1. Bluetooth serial ports detected as scales
   - Solution: Serial scanner filters these automatically
2. Non-printer USB devices detected
   - Solution: Discovery filters by USB class
3. Generic network services detected
   - Solution: Limit mDNS service types in config

## Performance Considerations

### Scan Intervals

| Environment | Recommended Interval | Rationale |
|-------------|---------------------|-----------|
| Development | 10-30s | Fast iteration |
| Production | 5-10m | Reduce system load |
| Kiosk/POS | 1-2m | Balance responsiveness & load |

### Resource Usage

| Scanner | CPU Impact | Network Impact | Disk I/O |
|---------|------------|----------------|----------|
| USB | Low | None | None |
| Serial | Medium | None | None |
| mDNS | Low | Low | None |
| TCP | High | High | None |

### Optimization Tips

1. **Disable unused transports** - Don't scan for devices you don't use
2. **Increase scan interval** - Less frequent scans reduce load
3. **Limit baud rates** - Fewer baud rates = faster serial discovery
4. **Use auto-register carefully** - Manual approval prevents unwanted devices

## Next Steps

- Configure discovery in your `config.yaml`
- Test with `bridge-cli devices discover` (when available)
- Subscribe to discovery events in your application
- Set up auto-registration policies
- Monitor discovery performance

## Related Documentation

- [USB Printer Setup](USB_PRINTER_SETUP.md)
- [USB Scanner Validation](USB_SCANNER_VALIDATION.md)
- [Scale Setup](SCALE_SETUP.md)
- [Platform Compatibility](PLATFORM_COMPATIBILITY_MATRIX.md)
