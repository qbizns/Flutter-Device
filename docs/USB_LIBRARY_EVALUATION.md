# USB Library Evaluation for Device Bridge v2

**Date**: 2025-11-10
**Purpose**: Select USB library for Phase 3 USB HID Scanner implementation
**Decision**: ✅ **google/gousb** selected

---

## Comparison Matrix

| Feature | google/gousb | karalabe/usb | go-hid/hid |
|---------|--------------|--------------|------------|
| **Maturity** | High (Google maintained) | Medium | Medium |
| **Platform Support** | Linux, macOS, Windows | Linux, macOS, Windows | Linux, macOS, Windows |
| **Dependencies** | libusb-1.0 | libusb-1.0 | hidapi |
| **API Design** | Clean, Go-idiomatic | Clean, minimal | Simplified HID-only |
| **HID Support** | Full control | Full control | HID-specific |
| **Hotplug** | Manual polling | Manual polling | Manual polling |
| **Maintenance** | Active (Google) | Active | Active |
| **Documentation** | Excellent | Good | Good |
| **Stars** | ~800 | ~600 | ~300 |
| **Last Update** | Recent | Recent | Recent |

---

## Detailed Analysis

### 1. google/gousb

**Repository**: https://github.com/google/gousb
**License**: Apache 2.0
**Maintainer**: Google

#### Pros:
- ✅ Well-maintained by Google
- ✅ Comprehensive API covering full USB spec
- ✅ Excellent documentation and examples
- ✅ Supports HID, bulk, interrupt, isochronous transfers
- ✅ Context management for lifecycle
- ✅ Strong type safety
- ✅ Used in production by multiple projects
- ✅ Cross-platform (Linux, macOS, Windows)

#### Cons:
- ⚠️ Requires libusb-1.0 system dependency
- ⚠️ More complex API (but flexible)
- ⚠️ Manual device permission handling on Linux

#### Example Usage:
```go
ctx := gousb.NewContext()
defer ctx.Close()

dev, err := ctx.OpenDeviceWithVIDPID(0x05e0, 0x1200)
if err != nil {
    return err
}
defer dev.Close()

cfg, err := dev.Config(1)
defer cfg.Close()

intf, err := cfg.Interface(0, 0)
defer intf.Close()

ep, err := intf.InEndpoint(1)
buf := make([]byte, 64)
n, err := ep.Read(buf)
```

#### System Requirements:
```bash
# Linux
sudo apt-get install libusb-1.0-0-dev

# macOS
brew install libusb

# Windows
# libusb included in package
```

---

### 2. karalabe/usb

**Repository**: https://github.com/karalabe/usb
**License**: LGPL v3
**Maintainer**: Péter Szilágyi (Ethereum)

#### Pros:
- ✅ Simpler API than gousb
- ✅ Good cross-platform support
- ✅ Used in Ethereum projects
- ✅ Clean enumeration API
- ✅ Less boilerplate code

#### Cons:
- ⚠️ LGPL license (more restrictive)
- ⚠️ Smaller community
- ⚠️ Less flexible for complex USB operations
- ⚠️ Limited to HID-like devices

#### Example Usage:
```go
hids, err := usb.EnumerateHid(0x05e0, 0x1200)
for _, dev := range hids {
    device, err := dev.Open()
    defer device.Close()

    buf := make([]byte, 64)
    n, err := device.Read(buf)
}
```

---

### 3. go-hid/hid

**Repository**: https://github.com/sstallion/go-hid
**License**: MIT
**Maintainer**: Steven Stallion

#### Pros:
- ✅ Simplified HID-only API
- ✅ Based on hidapi (widely used)
- ✅ Good for simple HID devices
- ✅ MIT license (permissive)

#### Cons:
- ⚠️ HID-only (can't access other USB devices)
- ⚠️ Less control over low-level USB
- ⚠️ Smaller feature set
- ⚠️ Different dependency (hidapi vs libusb)

---

## Decision: google/gousb ✅

### Rationale:

1. **Maturity & Maintenance**: Google backing ensures long-term support
2. **Flexibility**: Full USB spec support allows future expansion
3. **Production Ready**: Used in real-world applications
4. **Documentation**: Comprehensive docs and examples
5. **License**: Apache 2.0 is permissive and enterprise-friendly
6. **Community**: Larger community for troubleshooting

### Trade-offs Accepted:

1. **Complexity**: More API surface, but well-documented
2. **Dependencies**: libusb-1.0 is widely available and standard
3. **Learning Curve**: Worth it for flexibility and future-proofing

---

## Implementation Plan

### Phase 1: Basic Integration (Current)
- [x] Add gousb dependency
- [ ] Implement device enumeration
- [ ] Implement HID interface detection
- [ ] Implement interrupt endpoint reading
- [ ] Add error handling and reconnection

### Phase 2: Advanced Features
- [ ] Hotplug detection (udev on Linux)
- [ ] Device filtering by class/subclass
- [ ] Configuration management
- [ ] Timeout and cancellation support

### Phase 3: Platform-Specific
- [ ] macOS implementation (hid_darwin.go)
- [ ] Windows implementation (hid_windows.go)
- [ ] Permission handling helpers

---

## Alternative Considered: karalabe/usb

While karalabe/usb has a simpler API, we chose gousb for:
- Better long-term support (Google maintained)
- More flexible API for future expansion
- Better documentation
- Apache 2.0 license (vs LGPL v3)

If gousb proves problematic, karalabe/usb is a viable fallback.

---

## Installation & Setup

### Developer Setup:

```bash
# Install system dependency
# Linux (Debian/Ubuntu)
sudo apt-get install libusb-1.0-0-dev

# Linux (Fedora/RHEL)
sudo dnf install libusb-devel

# macOS
brew install libusb

# Go dependency
go get github.com/google/gousb
```

### Production Deployment:

```dockerfile
# Dockerfile
FROM golang:1.23 AS builder
RUN apt-get update && apt-get install -y libusb-1.0-0-dev

# Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y libusb-1.0-0
```

### Linux Permissions:

```bash
# Add udev rule for scanners (optional, for non-root access)
echo 'SUBSYSTEM=="usb", ATTRS{idVendor}=="05e0", MODE="0666"' | \
  sudo tee /etc/udev/rules.d/99-barcode-scanners.rules

sudo udevadm control --reload-rules
sudo udevadm trigger
```

---

## References

- gousb documentation: https://pkg.go.dev/github.com/google/gousb
- libusb documentation: https://libusb.info/
- USB HID spec: https://www.usb.org/hid
- USB HID Usage Tables: https://usb.org/sites/default/files/hut1_3_0.pdf

---

## Conclusion

**google/gousb** selected as the USB library for Device Bridge v2 Phase 3. Implementation to begin in `hid_linux.go` with fallbacks to other platforms in subsequent iterations.

**Status**: ✅ Decision Made
**Next**: Implement Linux USB device handling
