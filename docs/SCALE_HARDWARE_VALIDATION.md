# Serial Scale Hardware Validation Guide

Complete hardware validation checklist and procedures for serial scale drivers.

## Table of Contents

- [Overview](#overview)
- [Pre-Validation Requirements](#pre-validation-requirements)
- [Validation Checklist](#validation-checklist)
- [Protocol-Specific Tests](#protocol-specific-tests)
- [Performance Benchmarks](#performance-benchmarks)
- [Long-Running Tests](#long-running-tests)
- [Troubleshooting](#troubleshooting)
- [Hardware Compatibility Matrix](#hardware-compatibility-matrix)

## Overview

This guide provides comprehensive procedures for validating serial scale drivers with physical hardware. Complete this validation before deploying scales in production environments.

**Validation Goals:**
- ✅ Verify all 5 protocols work correctly
- ✅ Validate zero/tare operations
- ✅ Confirm stable weight detection
- ✅ Test auto-reconnection
- ✅ Measure performance and reliability
- ✅ Document hardware compatibility

**Timeline:** 3-5 days with hardware access

## Pre-Validation Requirements

### Hardware Needed

**Minimum (1 scale):**
- 1x Serial scale (any supported protocol)
- 1x USB-to-Serial adapter (if using USB port)
- 1x Serial cable (usually included with scale)
- Test weights (calibrated if possible)

**Recommended (3+ scales):**
- 1x Mettler Toledo scale (MT-SICS protocol)
- 1x CAS scale (CAS protocol)
- 1x Dibal or Toledo scale
- Set of calibrated test weights (100g, 500g, 1kg, 5kg)
- Multiple serial cables for quick swapping

### Software Requirements

- Device Bridge v2 built and ready
- `grpcurl` installed (for API testing)
- Terminal access
- Validation scripts from `scripts/validation/`

### Environment Setup

**Linux:**
```bash
# Add user to dialout group
sudo usermod -a -G dialout $USER
# Logout and login

# Verify ports
ls -l /dev/tty{USB,S,ACM}*
```

**macOS:**
```bash
# Install libusb
brew install libusb

# Verify ports
ls /dev/{tty,cu}.*
```

**Windows:**
```powershell
# Run as Administrator
# Check Device Manager for COM ports
Get-CimInstance -ClassName Win32_SerialPort
```

## Validation Checklist

### Phase 1: Basic Connectivity (30 minutes per scale)

For each scale model:

- [ ] **Physical Setup**
  - [ ] Scale powered on
  - [ ] Serial cable connected properly
  - [ ] Port identified (e.g., /dev/ttyUSB0, COM1)
  - [ ] Baud rate configured correctly (check scale manual)

- [ ] **Port Detection**
  - [ ] Port appears in system (ls /dev/tty* or Device Manager)
  - [ ] Correct permissions set (Linux: dialout group)
  - [ ] No conflicts with other devices

- [ ] **Driver Connection**
  - [ ] Device Bridge starts without errors
  - [ ] Driver connects to scale
  - [ ] Initial weight reading succeeds
  - [ ] Logs show no connection errors

**Test Command:**
```bash
./scripts/validation/test-serial-scale.sh /dev/ttyUSB0 auto
```

### Phase 2: Protocol Validation (1 hour per protocol)

For each supported protocol:

#### MT-SICS Protocol (Mettler Toledo)

- [ ] **Basic Operations**
  - [ ] Immediate weight read (SI command)
  - [ ] Stable weight read (S command)
  - [ ] Zero operation (Z command)
  - [ ] Tare operation (T command)

- [ ] **Response Parsing**
  - [ ] Stable status detected (S)
  - [ ] Unstable status detected (D)
  - [ ] Overload detected (+)
  - [ ] Underload detected (-)
  - [ ] Unit parsing (g, kg, lb, oz)
  - [ ] Decimal places handled correctly

- [ ] **Error Handling**
  - [ ] Invalid command rejected
  - [ ] Timeout handling works
  - [ ] Reconnection after disconnect

#### CAS Protocol

- [ ] **Basic Operations**
  - [ ] Weight request (STX W ETX)
  - [ ] Zero operation
  - [ ] Tare operation

- [ ] **Response Parsing**
  - [ ] STX/ETX framing correct
  - [ ] Status byte (S/U/E) detected
  - [ ] 6-digit weight parsed
  - [ ] Unit extracted correctly
  - [ ] Checksum validated (if enabled)

- [ ] **Variants**
  - [ ] With checksum
  - [ ] Without checksum
  - [ ] Different decimal places (configurable)

#### Dibal Protocol

- [ ] **Continuous Output**
  - [ ] Receives continuous stream
  - [ ] STX/ETX framing correct
  - [ ] Status (+/-) detected
  - [ ] Weight parsed correctly

- [ ] **Response Handling**
  - [ ] Stable weight (+)
  - [ ] Unstable weight (-)
  - [ ] Unit parsing
  - [ ] Decimal places configuration

#### Toledo 8217 Protocol

- [ ] **Continuous ASCII**
  - [ ] Receives CR/LF terminated data
  - [ ] Status byte (S/D/M/E) parsed
  - [ ] Fixed-width format
  - [ ] Variable-width format

- [ ] **Motion Detection**
  - [ ] Motion status detected (M)
  - [ ] Returns to stable (S)

#### Generic Protocol

- [ ] **Fallback Mode**
  - [ ] Receives data
  - [ ] Parses numbers
  - [ ] Basic weight extraction

### Phase 3: Functional Testing (2 hours per scale)

- [ ] **Weight Reading**
  - [ ] Empty scale reads zero (or near-zero)
  - [ ] Known weight reads correctly (±tolerance)
  - [ ] Multiple weights tested (100g, 500g, 1kg, 5kg)
  - [ ] Accuracy within manufacturer spec
  - [ ] Consistent readings (10 samples, < 1% variation)

- [ ] **Zero Operation**
  - [ ] Empty scale: zero resets to 0.000
  - [ ] With light object: zero resets to 0.000
  - [ ] After zero: weight readings accurate
  - [ ] Zero persists across readings

- [ ] **Tare Operation**
  - [ ] Place container: reads container weight
  - [ ] Tare: display shows 0.000
  - [ ] Add product: shows only product weight
  - [ ] Remove product: shows negative container weight
  - [ ] Tare persists until reset

- [ ] **Stable Weight Detection**
  - [ ] Static weight: marked as stable
  - [ ] During placement: marked as unstable
  - [ ] Settles to stable within expected time
  - [ ] `stable_only` flag works correctly

- [ ] **Unit Conversion**
  - [ ] Reads in grams (g)
  - [ ] Reads in kilograms (kg)
  - [ ] Reads in pounds (lb)
  - [ ] Reads in ounces (oz)
  - [ ] Preferred unit setting works

- [ ] **Error Conditions**
  - [ ] Overload: detected and reported
  - [ ] Underload: detected and reported
  - [ ] No weight: handled gracefully
  - [ ] Scale disconnected: detected
  - [ ] Invalid command: rejected

### Phase 4: Reliability Testing (4-8 hours)

- [ ] **Continuous Operation**
  - [ ] Read weight 1000 times: all succeed
  - [ ] No memory leaks (monitor RAM)
  - [ ] No connection drops
  - [ ] Response time consistent

- [ ] **Stress Testing**
  - [ ] Rapid weight changes: all detected
  - [ ] Quick succession reads (100/sec): handled
  - [ ] Concurrent API requests: serialized correctly
  - [ ] High frequency zero/tare: no errors

- [ ] **Auto-Reconnection**
  - [ ] Unplug scale: disconnect detected
  - [ ] Replug scale: reconnection automatic
  - [ ] Reads work after reconnection
  - [ ] No restart required

- [ ] **Long-Running Test (24 hours)**
  - [ ] Set up: read weight every 30 seconds
  - [ ] Leave running overnight
  - [ ] Check logs: no errors
  - [ ] Performance: consistent response times
  - [ ] Memory: stable (no leaks)

### Phase 5: Performance Benchmarks (1 hour)

- [ ] **Response Time**
  - [ ] Immediate read: < 200ms average
  - [ ] Stable read: < 2s average (depends on scale)
  - [ ] Zero operation: < 500ms
  - [ ] Tare operation: < 500ms

- [ ] **Throughput**
  - [ ] Continuous reads: 5-10 per second minimum
  - [ ] No dropped readings
  - [ ] Consistent timing

- [ ] **Resource Usage**
  - [ ] CPU: < 5% during normal operation
  - [ ] Memory: < 50MB per scale driver
  - [ ] No memory leaks over 24 hours

## Protocol-Specific Tests

### MT-SICS Detailed Tests

```bash
# Test immediate weight
grpcurl -plaintext -d '{
  "device_id": "scale-test",
  "stable_only": false
}' localhost:50051 devicebridge.v1.DeviceBridge/ReadWeight

# Test stable weight (wait for stability)
grpcurl -plaintext -d '{
  "device_id": "scale-test",
  "stable_only": true
}' localhost:50051 devicebridge.v1.DeviceBridge/ReadWeight

# Test zero
grpcurl -plaintext -d '{
  "device_id": "scale-test"
}' localhost:50051 devicebridge.v1.DeviceBridge/ZeroScale

# Test tare
grpcurl -plaintext -d '{
  "device_id": "scale-test"
}' localhost:50051 devicebridge.v1.DeviceBridge/TareScale
```

### CAS Detailed Tests

```bash
# Test with checksum validation
# Edit config: checksum: true
# Restart bridge
# Test weight read

# Test without checksum
# Edit config: checksum: false
# Restart bridge
# Test weight read

# Test different decimal places
# Edit config: decimal_places: 2 (or 0, 1, 3)
# Restart bridge
# Verify weight precision
```

### Auto-Protocol Detection Test

```bash
# Set protocol to "auto" in config
# Start bridge
# Check logs for detected protocol
# Verify correct protocol selected
# Test weight reading works
```

## Performance Benchmarks

### Expected Performance

| Metric | Target | Acceptable | Unacceptable |
|--------|--------|------------|--------------|
| Immediate read latency | < 100ms | < 200ms | > 500ms |
| Stable read latency | < 1s | < 2s | > 5s |
| Zero/Tare latency | < 200ms | < 500ms | > 1s |
| Throughput | 10 reads/sec | 5 reads/sec | < 2 reads/sec |
| CPU usage | < 2% | < 5% | > 10% |
| Memory per scale | < 20MB | < 50MB | > 100MB |
| Reconnect time | < 2s | < 5s | > 10s |
| 24h uptime | 100% | 99.9% | < 99% |

### Benchmark Test Script

```bash
#!/bin/bash
# Benchmark script
echo "Running performance benchmark..."

# Test 1: Latency (100 reads)
echo "Test 1: Read latency (100 samples)"
for i in {1..100}; do
    /usr/bin/time -f "%e" grpcurl -plaintext -d '{"device_id": "scale-test"}' \
        localhost:50051 devicebridge.v1.DeviceBridge/ReadWeight 2>&1 | grep -E "[0-9]+\.[0-9]+"
done | awk '{sum+=$1; count++} END {print "Average:", sum/count "s"}'

# Test 2: Throughput (10 seconds)
echo "Test 2: Throughput (10 seconds)"
START=$(date +%s)
COUNT=0
while [ $(($(date +%s) - START)) -lt 10 ]; do
    grpcurl -plaintext -d '{"device_id": "scale-test"}' \
        localhost:50051 devicebridge.v1.DeviceBridge/ReadWeight > /dev/null 2>&1
    ((COUNT++))
done
echo "Reads per second: $((COUNT / 10))"

# Test 3: Resource usage
echo "Test 3: Resource usage"
ps aux | grep device-bridge | grep -v grep
```

## Long-Running Tests

### 24-Hour Soak Test

**Setup:**
```yaml
# config-soak-test.yaml
devices:
  - id: scale-soak-test
    kind: scale.serial
    metadata:
      port: "/dev/ttyUSB0"
      protocol: "mtsics"
      # ... other settings
```

**Test Script:**
```bash
#!/bin/bash
# 24-hour soak test
LOG_FILE="soak-test-$(date +%Y%m%d-%H%M%S).log"

echo "Starting 24-hour soak test..." | tee -a $LOG_FILE
echo "Start time: $(date)" | tee -a $LOG_FILE

# Run for 24 hours (86400 seconds)
# Read every 30 seconds (2880 reads total)
END_TIME=$(($(date +%s) + 86400))

READ_COUNT=0
ERROR_COUNT=0

while [ $(date +%s) -lt $END_TIME ]; do
    TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")

    # Read weight
    RESULT=$(grpcurl -plaintext -d '{"device_id": "scale-soak-test"}' \
        localhost:50051 devicebridge.v1.DeviceBridge/ReadWeight 2>&1)

    if echo "$RESULT" | grep -q "value"; then
        VALUE=$(echo "$RESULT" | grep -o '"value":[0-9.]*' | cut -d: -f2)
        echo "[$TIMESTAMP] Read #$READ_COUNT: $VALUE" | tee -a $LOG_FILE
        ((READ_COUNT++))
    else
        echo "[$TIMESTAMP] ERROR: Read failed" | tee -a $LOG_FILE
        ((ERROR_COUNT++))
    fi

    # Memory check every 100 reads
    if [ $((READ_COUNT % 100)) -eq 0 ]; then
        MEM=$(ps aux | grep device-bridge | grep -v grep | awk '{print $6}')
        echo "[$TIMESTAMP] Memory: ${MEM}KB" | tee -a $LOG_FILE
    fi

    sleep 30
done

echo "Test completed: $(date)" | tee -a $LOG_FILE
echo "Total reads: $READ_COUNT" | tee -a $LOG_FILE
echo "Errors: $ERROR_COUNT" | tee -a $LOG_FILE
echo "Success rate: $(echo "scale=2; 100 * ($READ_COUNT - $ERROR_COUNT) / $READ_COUNT" | bc)%" | tee -a $LOG_FILE
```

**Success Criteria:**
- ✅ Success rate: > 99.9%
- ✅ No memory leaks (stable memory usage)
- ✅ No connection drops
- ✅ Consistent response times

## Troubleshooting

### Common Issues

#### No Weight Readings

**Symptoms:** Driver connects but no weight data
**Checks:**
1. Verify baud rate matches scale configuration
2. Check data bits, parity, stop bits settings
3. Try different protocols (auto-detection)
4. Check serial cable (try different cable)
5. Verify scale is in correct output mode

**Solution:**
```bash
# Try auto-protocol detection
# Set protocol: "auto" in config
# Check logs for detected protocol

# Or try each protocol manually
for protocol in mtsics cas dibal toledo; do
    echo "Testing $protocol..."
    # Update config with protocol
    # Restart bridge
    # Check if readings work
done
```

#### Unstable Readings

**Symptoms:** Weight fluctuates constantly
**Checks:**
1. Scale on stable surface
2. No vibrations or air currents
3. Scale properly calibrated
4. No electromagnetic interference

**Solution:**
- Use stable_only: true for reads
- Increase settling time
- Calibrate scale
- Move to stable location

#### Connection Drops

**Symptoms:** Regular disconnections
**Checks:**
1. Loose cable connections
2. USB-to-Serial adapter issues
3. Power supply problems
4. USB power management

**Solution:**
```bash
# Disable USB power management (Linux)
echo 'SUBSYSTEM=="usb", TEST=="power/control", ATTR{power/control}="on"' | \
    sudo tee /etc/udev/rules.d/50-usb-power.rules

# Reconnect timeout
# Increase in config: reconnect_delay: 10s
```

#### Slow Response Times

**Symptoms:** Reads take > 1 second
**Checks:**
1. Network latency (if remote)
2. High system load
3. Scale in wrong mode
4. Baud rate too low

**Solution:**
- Increase baud rate (if scale supports)
- Reduce retry_attempts
- Optimize system load

## Hardware Compatibility Matrix

### Tested Configurations

| Manufacturer | Model | Protocol | Baud | Status | Notes |
|--------------|-------|----------|------|--------|-------|
| Mettler Toledo | ICS685 | MT-SICS | 9600 | ✅ Validated | Full support |
| Mettler Toledo | PS60 | MT-SICS | 9600 | ⏳ Pending | Awaiting hardware |
| CAS | CL5000 | CAS | 9600 | ⏳ Pending | Awaiting hardware |
| CAS | MW-II | CAS | 19200 | ⏳ Pending | Awaiting hardware |
| Dibal | D-955 | Dibal | 9600 | ⏳ Pending | Awaiting hardware |
| Toledo | 8217 | Toledo | 9600 | ⏳ Pending | Awaiting hardware |

### Testing Status Legend

- ✅ **Validated** - Fully tested and working
- ⏳ **Pending** - Awaiting hardware for testing
- ⚠️ **Limited** - Partial functionality
- ❌ **Incompatible** - Does not work

### Add Your Results

When you validate a new scale model, add it to the matrix:

```yaml
# Format:
| Manufacturer | Model | Protocol | Baud | Status | Notes |
| ------------ | ----- | -------- | ---- | ------ | ----- |
| Your Company | ABC-123 | mtsics | 9600 | ✅ Validated | Works perfectly |
```

## Next Steps

After completing validation:

1. **Document Results**
   - Update Hardware Compatibility Matrix
   - Add specific model notes to SCALE_SETUP.md
   - Report any issues or limitations

2. **Performance Tuning**
   - Optimize baud rate
   - Adjust timeouts
   - Configure retry logic

3. **Production Deployment**
   - Use validated configurations
   - Set up monitoring
   - Plan maintenance procedures

4. **Feedback**
   - Report hardware compatibility results
   - Suggest protocol improvements
   - Share best practices

## Related Documentation

- [Scale Setup Guide](SCALE_SETUP.md)
- [Discovery Guide](DISCOVERY.md)
- [Platform Compatibility](PLATFORM_COMPATIBILITY_MATRIX.md)
- [Validation Scripts](../scripts/validation/README.md)
