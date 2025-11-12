# Zebra Badge Printer - Troubleshooting Guide

**Comprehensive troubleshooting guide for Zebra card/badge printers**

Quick reference for diagnosing and resolving common issues with Zebra card printers integrated with Device Bridge v2.

---

## Quick Diagnostics

### System Check Commands

```bash
# Check if Device Bridge is running
systemctl status device-bridge

# Check printer status
curl http://localhost:8080/v1/devices/badge-printer-01/badge/status

# Check recent logs
journalctl -u device-bridge -n 100 --no-pager

# List USB devices
lsusb | grep -i zebra

# Check network connectivity (TCP printers)
ping 192.168.1.100
telnet 192.168.1.100 9100
```

---

## Common Issues

### 1. Printer Not Detected

#### Symptoms
- Device Bridge shows "disconnected" status
- API returns "not found" error
- USB device not listed

#### Diagnosis

**USB Printers:**
```bash
# Check if USB device is recognized
lsusb | grep 0a5f

# Expected output:
# Bus 001 Device 005: ID 0a5f:0176 Zebra ZC300

# Check USB permissions
ls -l /dev/usb/lp*

# Check dmesg for errors
dmesg | grep -i zebra
```

**Network Printers:**
```bash
# Test network connectivity
ping 192.168.1.100

# Check if printer port is accessible
telnet 192.168.1.100 9100
nc -zv 192.168.1.100 9100

# Test with echo
echo "~HQES" | nc 192.168.1.100 9100
```

#### Solutions

**USB Not Detected:**
1. Reconnect USB cable
2. Try different USB port (use USB 2.0, not 3.0)
3. Check udev rules:
   ```bash
   sudo nano /etc/udev/rules.d/99-zebra-printers.rules
   # Add: SUBSYSTEM=="usb", ATTR{idVendor}=="0a5f", ATTR{idProduct}=="0176", MODE="0666"

   sudo udevadm control --reload-rules
   sudo udevadm trigger
   ```
4. Restart Device Bridge:
   ```bash
   systemctl restart device-bridge
   ```

**Network Not Reachable:**
1. Check printer IP configuration (via LCD panel)
2. Verify network cable connection
3. Check firewall:
   ```bash
   sudo ufw allow 9100/tcp
   sudo iptables -A INPUT -p tcp --dport 9100 -j ACCEPT
   ```
4. Verify printer is on same subnet
5. Check router/switch configuration

---

### 2. Ribbon Errors

#### Error: "No Ribbon"

**Diagnosis:**
- Ribbon not installed
- Ribbon not seated correctly
- Ribbon sensor dirty

**Solutions:**
1. Open printer cover
2. Remove and reinstall ribbon cartridge
3. Ensure ribbon clicks into place
4. Clean ribbon sensor with compressed air
5. Check ribbon compatibility with printer model

#### Error: "Wrong Ribbon Type"

**Diagnosis:**
- Ribbon doesn't match card type
- Using color ribbon for monochrome cards
- Using monochrome ribbon for color

**Solutions:**
1. Verify card type: Color (PVC) or monochrome
2. Install correct ribbon:
   - **YMCKO**: Full color cards
   - **KO**: Monochrome with overlay
   - **K**: Monochrome only
3. Update configuration if printer has multiple ribbon types

#### Error: "Ribbon Low"

**Warning Level:**
- < 20 panels: Warning
- < 10 panels: Critical

**Solutions:**
1. Monitor ribbon status via API:
   ```bash
   curl http://localhost:8080/v1/devices/badge-printer-01/badge/status | jq '.ribbon'
   ```
2. Replace ribbon when < 10% remaining
3. Keep spare ribbons in stock

---

### 3. Card Feed Issues

#### Error: "No Cards"

**Diagnosis:**
- Card feeder empty
- Cards not loaded correctly
- Card sensor blocked

**Solutions:**
1. Load cards into feeder (max 100 cards)
2. Fan cards before loading to prevent sticking
3. Ensure cards are aligned properly
4. Clean card feeder sensor

#### Error: "Card Jam"

**Diagnosis:**
- Card stuck in printer
- Bent or damaged card
- Wrong card thickness

**Solutions:**
1. Open printer cover
2. Carefully remove jammed card (follow path of travel)
3. Check for debris or torn card pieces
4. Verify card specifications:
   - Standard: CR80 (85.6 x 53.98 mm)
   - Thickness: 30 mil (0.76 mm)
   - Material: PVC
5. Run cleaning card cycle
6. Restart printer

#### Cards Feeding Multiple at Once

**Diagnosis:**
- Static electricity
- Cards stuck together
- Feeder adjustment needed

**Solutions:**
1. Fan cards thoroughly before loading
2. Use anti-static spray on cards
3. Reduce number of cards in feeder
4. Clean card feeder rollers
5. Adjust feeder tension (consult manual)

---

### 4. Print Quality Issues

#### Faded or Light Prints

**Diagnosis:**
- Ribbon depleted
- Print head worn
- Incorrect print density setting

**Solutions:**
1. Replace ribbon
2. Increase print density:
   ```yaml
   settings:
     print_speed: "quality"
     dpi: 600  # Higher DPI
   ```
3. Clean print head with cleaning pen
4. Replace print head if worn (every 50,000 prints)

#### White Spots or Streaks

**Diagnosis:**
- Dust on print head
- Debris on card
- Dirty rollers

**Solutions:**
1. Run printer cleaning cycle:
   - Use Zebra cleaning kit
   - Insert cleaning card
   - Run cleaning program
2. Clean print head manually with cleaning pen
3. Clean rollers with isopropyl alcohol
4. Use cards from sealed package (not exposed to dust)

#### Colors Incorrect or Washed Out

**Diagnosis:**
- Wrong ribbon type
- DPI setting too low
- Print quality setting incorrect

**Solutions:**
1. Verify YMCKO ribbon installed
2. Increase DPI:
   ```yaml
   settings:
     dpi: 600  # Use 600 for color
   ```
3. Set quality mode:
   ```yaml
   settings:
     print_speed: "quality"
   ```
4. Check color profiles (for advanced users)

#### Smudges or Scratches

**Diagnosis:**
- No overlay panel on ribbon
- Print head dirty
- Cards touching before overlay applied

**Solutions:**
1. Use ribbon with overlay (YMCKO, not YMCK)
2. Clean print head
3. Replace ribbon if overlay panel missing
4. Allow cards to cool before handling

---

### 5. Encoding Failures

#### Magnetic Stripe Encoding Fails

**Diagnosis:**
- Wrong card type (using LoCo cards with HiCo setting)
- Dirty magnetic stripe encoder
- Incorrect track format

**Error Codes:**
- `ERROR_CODE_ENCODER_ERROR`: General encoding failure
- Invalid track data format

**Solutions:**
1. Verify card coercivity matches setting:
   ```json
   {
     "encoding": {
       "magnetic_stripe": {
         "coercivity": "COERCIVITY_TYPE_HICO"  // or LOCO
       }
     }
   }
   ```
2. Clean magnetic stripe encoder with cleaning card
3. Verify track data format:
   - **Track 1**: `%B...?` format, alphanumeric
   - **Track 2**: `;...?` format, numeric
   - **Track 3**: `;...?` format, numeric
4. Check track length limits:
   - Track 1: 79 chars max
   - Track 2: 40 chars max
   - Track 3: 107 chars max

#### RFID Encoding Fails

**Diagnosis:**
- Wrong chip type
- Card not compatible
- RFID antenna misalignment

**Error Codes:**
- `ERROR_CODE_ENCODER_ERROR`
- "RFID write failed"

**Solutions:**
1. Verify RFID type matches card:
   ```json
   {
     "encoding": {
       "rfid": {
         "type": "RFID_TYPE_MIFARE_CLASSIC_1K"  // Match card type
       }
     }
   }
   ```
2. Check card compatibility:
   - Mifare 1K: 1KB storage
   - Mifare 4K: 4KB storage
   - HID iClass: Secure access
3. Ensure RFID encoder module is enabled
4. Test with known-good RFID card
5. Contact Zebra support for antenna alignment

---

### 6. Communication Errors

#### Error: "Communication Timeout"

**Diagnosis:**
- Slow network connection
- Printer not responding
- Print job too large

**Solutions:**
1. Increase timeout in configuration:
   ```yaml
   settings:
     print_timeout: "120s"  # Increase for complex jobs
     encode_timeout: "60s"
   ```
2. Check network latency:
   ```bash
   ping -c 10 192.168.1.100
   ```
3. Reduce batch size for large jobs
4. Restart printer
5. Check printer firmware version

#### Error: "USB Write Error"

**Diagnosis:**
- USB cable issue
- USB port power problem
- Driver conflict

**Solutions:**
1. Replace USB cable with high-quality cable
2. Use powered USB hub if needed
3. Try different USB port (preferably USB 2.0)
4. Check dmesg for USB errors:
   ```bash
   dmesg | tail -50 | grep -i usb
   ```
5. Reinstall USB drivers (Windows)

---

### 7. Application/API Errors

#### Error: "Invalid Request"

**Diagnosis:**
- Missing required fields
- Invalid template ID
- Malformed JSON

**Solutions:**
1. Validate JSON syntax:
   ```bash
   cat request.json | jq .
   ```
2. Check required fields:
   - `device_id` must match configured ID
   - `design` must be provided
   - Variables must match template requirements
3. Use example from documentation

#### Error: "Template Not Found"

**Diagnosis:**
- Template ID typo
- Template not loaded

**Solutions:**
1. List available templates:
   ```bash
   curl http://localhost:8080/v1/devices/badge-printer-01/badge/templates
   ```
2. Use exact template ID (case-sensitive):
   - `conference_attendee`
   - `vip_badge`
   - `staff_badge`
   - `visitor_badge`

#### Error: "Job Failed"

**Diagnosis:**
- Printer error during print
- Out of ribbon/cards
- Encoding failure

**Solutions:**
1. Check printer status:
   ```bash
   curl http://localhost:8080/v1/devices/badge-printer-01/badge/status
   ```
2. Review error details in response
3. Check ribbon and card levels
4. Verify encoding settings if using encoding
5. Retry print job

---

### 8. Performance Issues

#### Slow Print Speed

**Diagnosis:**
- Quality setting too high
- Network latency
- Complex design

**Solutions:**
1. Reduce quality for drafts:
   ```yaml
   settings:
     print_speed: "fast"
     dpi: 300
   ```
2. Simplify badge design (fewer elements)
3. Use batch printing for multiple badges
4. Upgrade network connection (use gigabit Ethernet)
5. Reduce image resolution

#### Queue Backlog

**Diagnosis:**
- High volume of print jobs
- Printer slower than expected
- Jobs not completing

**Solutions:**
1. Monitor queue via API:
   ```bash
   curl http://localhost:8080/v1/devices/badge-printer-01/badge/status | jq '.current_job_id'
   ```
2. Increase worker threads (if configurable)
3. Add additional printers for load balancing
4. Implement priority queues for VIP badges

---

## Diagnostic Procedures

### Print Test Card

```bash
# Simple text test
curl -X POST http://localhost:8080/v1/devices/badge-printer-01/badge/print \
  -H "Content-Type: application/json" \
  -d '{
    "design": {
      "custom": {
        "elements": [{
          "type": "ELEMENT_TYPE_TEXT",
          "position": {"x": 100, "y": 100},
          "content": "TEST CARD",
          "style": {"font_size": 40}
        }]
      }
    },
    "options": {"copies": 1}
  }'
```

### Check Event Log

```bash
# Subscribe to printer events
curl -N http://localhost:8080/v1/devices/badge-printer-01/badge/events

# Watch for errors:
# - EVENT_TYPE_ERROR
# - EVENT_TYPE_JOB_FAILED
# - EVENT_TYPE_RIBBON_LOW
# - EVENT_TYPE_CARDS_LOW
```

### Verify Encoding

```bash
# Test magnetic stripe encoding only
curl -X POST http://localhost:8080/v1/devices/badge-printer-01/badge/encode \
  -H "Content-Type: application/json" \
  -d '{
    "encoding": {
      "magnetic_stripe": {
        "enabled": true,
        "track2_data": "1234567890123456",
        "coercivity": "COERCIVITY_TYPE_HICO"
      }
    }
  }'

# Test RFID encoding only
curl -X POST http://localhost:8080/v1/devices/badge-printer-01/badge/encode \
  -H "Content-Type: application/json" \
  -d '{
    "encoding": {
      "rfid": {
        "enabled": true,
        "type": "RFID_TYPE_MIFARE_CLASSIC_1K",
        "uid": "01020304"
      }
    }
  }'
```

---

## Error Code Reference

| Code | Description | Action |
|------|-------------|--------|
| `ERROR_CODE_NO_RIBBON` | Ribbon not detected | Install/reseat ribbon |
| `ERROR_CODE_NO_CARDS` | Card feeder empty | Load cards |
| `ERROR_CODE_CARD_JAM` | Card stuck in printer | Clear jam, clean rollers |
| `ERROR_CODE_RIBBON_JAM` | Ribbon jammed | Clear jam, reinstall ribbon |
| `ERROR_CODE_PRINT_HEAD_ERROR` | Print head failure | Clean head, replace if needed |
| `ERROR_CODE_ENCODER_ERROR` | Encoding failed | Check card type, clean encoder |
| `ERROR_CODE_COMMUNICATION_ERROR` | Connection lost | Check network/USB, restart |
| `ERROR_CODE_DOOR_OPEN` | Printer cover open | Close cover securely |
| `ERROR_CODE_OVER_TEMPERATURE` | Print head overheated | Let cool, reduce print speed |
| `ERROR_CODE_INVALID_RIBBON` | Wrong ribbon type | Install correct ribbon |

---

## Maintenance Checklist

### Daily
- [ ] Check ribbon level
- [ ] Check card feeder level
- [ ] Visual inspection for errors
- [ ] Test print if needed

### Weekly
- [ ] Clean exterior
- [ ] Check for error messages
- [ ] Review print logs
- [ ] Verify network connectivity (TCP printers)

### Monthly
- [ ] Run printer cleaning cycle
- [ ] Clean print head with cleaning pen
- [ ] Clean rollers
- [ ] Inspect ribbon path for debris
- [ ] Update firmware if needed

### Quarterly
- [ ] Deep clean with full cleaning kit
- [ ] Professional service inspection
- [ ] Replace worn parts
- [ ] Calibrate printer
- [ ] Backup configuration

---

## Getting Support

### Logs to Collect

```bash
# Device Bridge logs
journalctl -u device-bridge --since "1 hour ago" > device-bridge.log

# Printer status
curl http://localhost:8080/v1/devices/badge-printer-01/badge/status > printer-status.json

# System info
uname -a > system-info.txt
lsusb > usb-devices.txt
```

### Information to Provide

1. **Printer Model**: ZC100, ZC300, ZC350, etc.
2. **Firmware Version**: From printer LCD or status API
3. **Device Bridge Version**: From `--version` flag
4. **Connection Type**: USB or Network (TCP)
5. **Error Message**: Exact error text/code
6. **Configuration**: Sanitized config.yaml
7. **Logs**: Last 100 lines before error
8. **What You Were Trying**: Print, encode, etc.

### Support Channels

- **Zebra Technical Support**: https://www.zebra.com/support
- **Device Bridge GitHub**: https://github.com/Macber-eg/Flutter-Device/issues
- **Emergency Support**: Check your Zebra warranty/support contract

---

## Advanced Troubleshooting

### Enable Debug Logging

```yaml
# config.yaml
logging:
  level: "debug"  # More verbose logging
  format: "json"
```

Restart Device Bridge:
```bash
systemctl restart device-bridge
journalctl -u device-bridge -f
```

### Packet Capture (Network Printers)

```bash
# Capture ZPL traffic
sudo tcpdump -i eth0 -A 'tcp port 9100' > zpl-capture.txt

# Analyze what's being sent
cat zpl-capture.txt | grep "^\^"
```

### USB Traffic Analysis

```bash
# Install usbmon
sudo modprobe usbmon

# Capture USB traffic
sudo cat /sys/kernel/debug/usb/usbmon/0u > usb-capture.txt

# Stop with Ctrl+C after capturing
```

---

**Last Updated**: 2025-11-12
**Document Version**: 1.0
