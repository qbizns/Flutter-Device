# RFID/NFC Reader - Troubleshooting Guide

**Comprehensive troubleshooting guide for RFID/NFC readers integrated with Device Bridge v2**

Quick reference for diagnosing and resolving common issues with RFID/NFC card readers.

---

## Quick Diagnostics

### System Check Commands

```bash
# Check if Device Bridge is running
systemctl status device-bridge

# Check reader status
curl http://localhost:8080/v1/devices/rfid-reader-01/status

# Check recent logs
journalctl -u device-bridge -n 100 --no-pager

# List USB devices
lsusb | grep -i "072f\|076b\|054c"

# Check PC/SC service (for smart card readers)
systemctl status pcscd

# Check network connectivity (TCP readers)
ping 192.168.1.50
telnet 192.168.1.50 10001
```

---

## Common Issues

### 1. Reader Not Detected

#### Symptoms
- Device Bridge shows "disconnected" status
- API returns "not found" error
- USB device not listed

#### Diagnosis

**USB Readers:**
```bash
# Check if USB device is recognized
lsusb | grep 072f

# Expected output (ACR122U):
# Bus 001 Device 005: ID 072f:2200 ACS ACR122U

# Check USB permissions
ls -l /dev/bus/usb/001/005

# Check dmesg for errors
dmesg | grep -i usb | tail -20

# Check if device is claimed by another driver
lsusb -t
```

**Network Readers:**
```bash
# Test network connectivity
ping 192.168.1.50

# Check if reader port is accessible
telnet 192.168.1.50 10001
nc -zv 192.168.1.50 10001

# Test with echo
echo "status" | nc 192.168.1.50 10001
```

**Serial Readers:**
```bash
# List serial ports
ls -l /dev/ttyUSB* /dev/ttyACM*

# Check port permissions
ls -l /dev/ttyUSB0

# Test serial communication
minicom -D /dev/ttyUSB0 -b 115200
```

#### Solutions

**USB Not Detected:**

1. **Reconnect USB cable**
   - Unplug and replug the reader
   - Try different USB port
   - Use USB 2.0 port (not 3.0) if having issues

2. **Check udev rules:**
   ```bash
   sudo nano /etc/udev/rules.d/99-rfid-readers.rules
   # Add:
   SUBSYSTEM=="usb", ATTR{idVendor}=="072f", ATTR{idProduct}=="2200", MODE="0666"

   sudo udevadm control --reload-rules
   sudo udevadm trigger
   ```

3. **Install/restart PC/SC service (for smart card readers):**
   ```bash
   sudo apt-get install pcscd
   sudo systemctl restart pcscd
   ```

4. **Check for kernel driver conflicts:**
   ```bash
   # Remove conflicting driver
   sudo rmmod pn533_usb
   sudo rmmod pn533

   # Blacklist if needed
   echo "blacklist pn533_usb" | sudo tee -a /etc/modprobe.d/blacklist.conf
   ```

5. **Restart Device Bridge:**
   ```bash
   systemctl restart device-bridge
   ```

**Network Not Reachable:**

1. **Check reader IP configuration:**
   - Use reader's LCD/web interface
   - Verify IP address, subnet, gateway

2. **Verify network cable:**
   - Check link lights on reader and switch
   - Try different cable

3. **Check firewall:**
   ```bash
   sudo ufw allow 10001/tcp
   sudo iptables -A INPUT -p tcp --dport 10001 -j ACCEPT
   ```

4. **Verify reader is on same subnet:**
   - Reader: 192.168.1.50
   - Server: 192.168.1.100
   - Both should be on 192.168.1.0/24

**Serial Not Working:**

1. **Check user permissions:**
   ```bash
   sudo usermod -a -G dialout $USER
   # Log out and log back in
   ```

2. **Verify baud rate:**
   - Common rates: 9600, 19200, 38400, 115200
   - Check reader documentation

3. **Test serial port:**
   ```bash
   stty -F /dev/ttyUSB0 115200
   cat /dev/ttyUSB0 &
   # Present card and check for output
   ```

---

### 2. Card Not Reading

#### Symptoms
- Reader detected but card not recognized
- Timeout errors when trying to read
- Intermittent reads

#### Diagnosis

```bash
# Enable debug logging
# In config.yaml:
logging:
  level: "debug"

# Restart and watch logs
journalctl -u device-bridge -f

# Test card read
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/read \
  -H "Content-Type: application/json" \
  -d '{"timeout_seconds": 10}'
```

#### Solutions

1. **Card Orientation:**
   - Try different orientations
   - Ensure card is flat against reader
   - Remove from wallet/case

2. **Card Distance:**
   - Optimal distance: 0-10cm for NFC
   - Try moving card closer/further

3. **Card Condition:**
   - Check for physical damage
   - Clean card with soft cloth
   - Replace damaged cards

4. **Electromagnetic Interference:**
   - Move reader away from metal objects
   - Keep away from other RF devices
   - Check for nearby computers/monitors

5. **Wrong Card Type:**
   ```bash
   # Check supported types
   curl http://localhost:8080/v1/devices/rfid-reader-01/types

   # Verify card type matches supported protocols
   ```

6. **Reader Antenna:**
   - Internal antenna may be faulty
   - Test with known-good card
   - Try different reader if available

---

### 3. Authentication Failures

#### Symptoms
- "Authentication failed" errors
- "Invalid key" errors
- Can read UID but not card data

#### Diagnosis

```bash
# Test with factory default keys
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/authenticate \
  -d '{
    "card_uid": "01020304",
    "method": "AUTH_METHOD_MIFARE_KEY",
    "credentials": {
      "mifare_keys_a": ["FFFFFFFFFFFF"]
    }
  }'
```

#### Solutions

**Mifare Classic:**

1. **Try default keys first:**
   - Key A: FF FF FF FF FF FF (factory default)
   - Key B: FF FF FF FF FF FF

2. **Try common keys:**
   ```python
   common_keys = [
       "FFFFFFFFFFFF",  # Factory default
       "A0A1A2A3A4A5",  # Common custom
       "D3F7D3F7D3F7",  # MAD key
       "000000000000",  # All zeros
       "AABBCCDDEEFF",  # Sequential
   ]
   ```

3. **Check sector access bits:**
   - Sector may be locked
   - Use Mifare Classic Tool to check

4. **Key diversification:**
   - Some systems use derived keys
   - Check with card issuer

**iClass:**

1. **Verify key format:**
   - iClass keys are 8 bytes (16 hex characters)
   - Example: `0102030405060708`

2. **Check key derivation:**
   - Some iClass cards use diversified keys
   - May need card-specific calculation

**DESFire:**

1. **Application-level authentication:**
   - DESFire requires authenticating to specific application
   - Default key: `00 00 00 00 00 00 00 00` (DES)
   - Default key: `00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00` (AES)

2. **Key settings:**
   - Check master key settings
   - May require multiple authentication steps

---

### 4. Write Failures

#### Symptoms
- "Write failed" errors
- Partial writes
- Verification errors

#### Diagnosis

```bash
# Test write with verification
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/write \
  -d '{
    "card_uid": "01020304",
    "blocks": [
      {"block_number": 4, "data": "00010203040506070809101112131415"}
    ],
    "credentials": {
      "mifare_keys_a": ["FFFFFFFFFFFF"]
    },
    "options": {
      "verify": true
    }
  }'
```

#### Solutions

1. **Card is Read-Only:**
   - **Prox cards**: Always read-only
   - **NTAG with lock**: Check if permanently locked
   - **Mifare with access bits**: Sector may be write-protected

2. **Authentication Required:**
   - Provide correct key for write operation
   - Some sectors require Key B for write

3. **Block Protection:**
   - **Sector trailer** (last block of sector): Special handling required
   - **Manufacturer block** (block 0): Read-only
   - **NDEF lock bytes**: Check if NDEF is locked

4. **Card Distance:**
   - Writing requires stable connection
   - Keep card completely still during write
   - Reduce distance to < 5cm

5. **Power Issues:**
   - Insufficient power from reader
   - Try powered USB hub
   - Check reader power supply

6. **Timeout:**
   ```yaml
   settings:
     write_timeout_sec: 30  # Increase timeout
   ```

---

### 5. Slow Read Performance

#### Symptoms
- Read operations take >2 seconds
- High latency
- Timeouts in busy environments

#### Diagnosis

```bash
# Check average read time
curl http://localhost:8080/v1/devices/rfid-reader-01/status | jq '.statistics.avg_read_time_ms'

# Monitor real-time performance
curl -N http://localhost:8080/v1/devices/rfid-reader-01/subscribe
```

#### Solutions

1. **Optimize Read Interval:**
   ```yaml
   settings:
     read_interval_ms: 300  # Reduce from 500ms
   ```

2. **Reduce Read Scope:**
   ```bash
   # Read only UID (fast)
   curl -X POST .../read -d '{
     "options": {
       "read_full_data": false
     }
   }'

   # Read specific sectors only
   curl -X POST .../read -d '{
     "options": {
       "read_full_data": true,
       "sectors": [1, 2]  # Instead of all sectors
     }
   }'
   ```

3. **Reader Mode:**
   ```yaml
   settings:
     mode: "single_read"  # Read once then idle
   ```

4. **Network Latency (TCP readers):**
   - Use wired connection instead of WiFi
   - Check network congestion
   - Reduce other traffic on network

5. **USB Issues:**
   - Use USB 2.0 port
   - Avoid USB hubs if possible
   - Update USB drivers

6. **Multiple Readers:**
   - Distribute load across multiple readers
   - Dedicate readers to specific areas

---

### 6. Connection Lost / Intermittent

#### Symptoms
- Reader disconnects randomly
- "Connection lost" errors
- Needs frequent reconnection

#### Diagnosis

```bash
# Watch for disconnect events
journalctl -u device-bridge -f | grep -i "disconnect\|error"

# Check USB stability
dmesg -w | grep -i usb

# Network connection stability (TCP)
ping -i 0.2 192.168.1.50  # Continuous ping
```

#### Solutions

**USB Disconnections:**

1. **USB Power Management:**
   ```bash
   # Disable USB autosuspend
   echo -1 | sudo tee /sys/module/usbcore/parameters/autosuspend

   # Disable for specific device
   echo 'on' | sudo tee /sys/bus/usb/devices/1-1.2/power/control
   ```

2. **USB Cable Quality:**
   - Use high-quality shielded cable
   - Keep cable length < 2 meters
   - Avoid cable near power cables

3. **USB Hub:**
   - Use powered USB hub
   - Avoid unpowered hubs
   - Connect directly to computer if possible

4. **USB Port:**
   - Some ports share bandwidth
   - Try different port
   - Use USB 2.0 port

**Network Disconnections:**

1. **Network Configuration:**
   ```bash
   # Set static IP on reader
   # Avoid DHCP lease expiration

   # Increase TCP keepalive
   echo 60 | sudo tee /proc/sys/net/ipv4/tcp_keepalive_time
   ```

2. **Switch/Router Issues:**
   - Check switch port configuration
   - Disable port auto-negotiation
   - Force 100Mbps full-duplex

3. **PoE Issues (if applicable):**
   - Check PoE power budget
   - Verify PoE injector/switch capacity
   - Test with external power

---

### 7. Multiple Card Detection

#### Symptoms
- Reader detects multiple cards
- "Collision" errors
- Inconsistent reads

#### Diagnosis

```bash
# Check for multiple cards in field
curl http://localhost:8080/v1/devices/rfid-reader-01/status
# Look at card_present and current_card_uid

# Enable collision detection logging
# config.yaml:
logging:
  level: "debug"
```

#### Solutions

1. **Physical Separation:**
   - Ensure only one card near reader
   - Remove cards from wallet
   - Separate reader from card stack

2. **Anti-Collision Algorithm:**
   - Most readers handle this automatically
   - May need firmware update

3. **Reader Configuration:**
   ```yaml
   settings:
     mode: "single_read"  # Read first card only
   ```

4. **Card Stacking:**
   - In high-volume scenarios, use card dispensers
   - Present cards individually

---

### 8. NDEF Read/Write Issues

#### Symptoms
- Cannot read NDEF records
- NDEF write fails
- "Not NDEF formatted" error

#### Diagnosis

```bash
# Check if card supports NDEF
curl -X POST .../read -d '{
  "options": {
    "read_full_data": true
  }
}'

# Check capabilities.supports_ndef in response
```

#### Solutions

1. **Card Not NDEF Formatted:**
   ```bash
   # Format card as NDEF (NTAG example)
   curl -X POST .../write -d '{
     "blocks": [
       {"block_number": 4, "data": "03..."}  # NDEF capability container
     ]
   }'
   ```

2. **NDEF Lock Byte:**
   - Check if NDEF is locked
   - Locked NDEF is read-only
   - Cannot unlock once locked

3. **Card Type:**
   - **NTAG**: Native NDEF support
   - **Mifare Classic**: Requires NDEF formatting (MAD)
   - **Mifare DESFire**: Application-level NDEF

4. **NDEF Size:**
   - Check available memory
   - NTAG213: 144 bytes user memory
   - NTAG215: 504 bytes
   - NTAG216: 888 bytes

---

## Error Code Reference

| Error | Description | Action |
|-------|-------------|--------|
| `ERR_READER_NOT_FOUND` | Reader device not detected | Check USB/network connection |
| `ERR_READ_TIMEOUT` | No card presented within timeout | Increase timeout, check card |
| `ERR_CARD_NOT_SUPPORTED` | Card type not supported | Check supported_protocols |
| `ERR_AUTH_FAILED` | Authentication failed | Verify keys, check card type |
| `ERR_WRITE_FAILED` | Write operation failed | Check authentication, card writable |
| `ERR_CARD_REMOVED` | Card removed during operation | Keep card steady |
| `ERR_COLLISION` | Multiple cards detected | Present one card at a time |
| `ERR_COMMUNICATION` | Reader communication error | Check connection, restart reader |
| `ERR_INVALID_BLOCK` | Block number out of range | Check block number, card size |
| `ERR_PERMISSION_DENIED` | USB/serial permission denied | Add user to dialout group |

---

## Diagnostic Procedures

### Test Card Reading

**Basic UID Read**:
```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/read \
  -H "Content-Type: application/json" \
  -d '{
    "timeout_seconds": 10
  }'
```

**Full Data Read**:
```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/read \
  -d '{
    "options": {
      "read_full_data": true,
      "authenticate": true,
      "auth_keys": ["FFFFFFFFFFFF"],
      "beep_on_read": true
    }
  }'
```

### Test Authentication

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/authenticate \
  -d '{
    "card_uid": "01020304",
    "method": "AUTH_METHOD_MIFARE_KEY",
    "credentials": {
      "mifare_keys_a": ["FFFFFFFFFFFF"]
    }
  }'
```

### Test Writing

```bash
curl -X POST http://localhost:8080/v1/devices/rfid-reader-01/write \
  -d '{
    "card_uid": "01020304",
    "blocks": [
      {
        "block_number": 4,
        "data": "48656C6C6F20526669642054657374"
      }
    ],
    "credentials": {
      "mifare_keys_a": ["FFFFFFFFFFFF"]
    },
    "options": {
      "verify": true
    }
  }'
```

### Monitor Events

```bash
# Subscribe to all events
curl -N http://localhost:8080/v1/devices/rfid-reader-01/events

# Watch for specific event types:
# - READER_EVENT_TYPE_CARD_READ
# - READER_EVENT_TYPE_CARD_PRESENT
# - READER_EVENT_TYPE_CARD_REMOVED
# - READER_EVENT_TYPE_READER_ERROR
# - READER_EVENT_TYPE_AUTH_FAILURE
```

---

## Advanced Troubleshooting

### Enable Verbose Logging

```yaml
# config.yaml
logging:
  level: "debug"  # Most verbose
  format: "json"
```

Restart Device Bridge:
```bash
systemctl restart device-bridge
journalctl -u device-bridge -f
```

### USB Traffic Analysis

```bash
# Install usbmon
sudo modprobe usbmon

# Capture USB traffic
sudo cat /sys/kernel/debug/usb/usbmon/0u > usb-capture.txt

# Stop with Ctrl+C after capturing

# Analyze with Wireshark
wireshark usb-capture.txt
```

### Network Traffic Analysis (TCP Readers)

```bash
# Capture TCP traffic to reader
sudo tcpdump -i eth0 -A 'tcp port 10001' -w rfid-capture.pcap

# Analyze with Wireshark
wireshark rfid-capture.pcap

# Or view in terminal
tcpdump -r rfid-capture.pcap -A
```

### PC/SC Reader Analysis

```bash
# List PC/SC readers
pcsc_scan

# Test with PC/SC tools
opensc-tool --list-readers
opensc-tool --reader 0 --atr

# Monitor PC/SC daemon
journalctl -u pcscd -f
```

### Check Reader Firmware

```bash
# Some readers support firmware version query
# Example for ACR122U:
echo -e '\xFF\x00\x48\x00\x00' | sudo tee /dev/bus/usb/001/005
```

---

## Maintenance Checklist

### Daily
- [ ] Check reader status via API
- [ ] Review error count in statistics
- [ ] Test with known-good card

### Weekly
- [ ] Clean reader surface
- [ ] Check connection cables
- [ ] Review access logs
- [ ] Verify network connectivity (TCP readers)

### Monthly
- [ ] Update Device Bridge software
- [ ] Review and rotate authentication keys
- [ ] Test backup readers
- [ ] Clean reader contacts (if applicable)
- [ ] Update reader firmware (if available)

### Quarterly
- [ ] Full system audit
- [ ] Replace worn cables
- [ ] Test failover procedures
- [ ] Security key rotation
- [ ] Performance benchmarking

---

## Getting Support

### Logs to Collect

```bash
# Device Bridge logs
journalctl -u device-bridge --since "1 hour ago" > device-bridge.log

# Reader status
curl http://localhost:8080/v1/devices/rfid-reader-01/status > reader-status.json

# System info
uname -a > system-info.txt
lsusb > usb-devices.txt
ip addr > network-config.txt

# Configuration (sanitized)
cat config.yaml | grep -v "key" > config-sanitized.yaml
```

### Information to Provide

1. **Reader Model**: ACR122U, ProxPoint Plus, etc.
2. **Firmware Version**: From reader or PC/SC
3. **Device Bridge Version**: From `--version` flag
4. **Connection Type**: USB, TCP, or Serial
5. **Card Type**: Mifare, Prox, iClass, NTAG
6. **Error Message**: Exact error text/code
7. **Configuration**: Sanitized config.yaml
8. **Logs**: Last 100 lines before error
9. **What You Were Trying**: Read, write, authenticate

### Support Channels

- **Device Bridge GitHub**: https://github.com/Macber-eg/Flutter-Device/issues
- **Reader Manufacturer Support**: Check reader documentation
- **Community Forums**: Reddit /r/RFID, NFC Forum

---

## Common Hardware-Specific Issues

### ACR122U (ACS)

**Issue**: Reader LED stays red
- **Solution**: Unplug for 10 seconds, replug

**Issue**: PC/SC conflict
```bash
# Stop pcscd
sudo systemctl stop pcscd
# Restart Device Bridge
systemctl restart device-bridge
```

### HID ProxPoint Plus

**Issue**: No Wiegand output
- **Solution**: Check Wiegand configuration switches
- Verify Wiegand wiring (green=D0, white=D1)

### OmniKey CardMan

**Issue**: iClass authentication fails
- **Solution**: Update firmware to latest version
- Check key diversification settings

### Sony PaSoRi

**Issue**: FeliCa cards not detected
- **Solution**: Ensure system code matches card
- Install nfcpy Python library if using PC/SC

---

**Last Updated**: 2025-11-12
**Document Version**: 1.0
