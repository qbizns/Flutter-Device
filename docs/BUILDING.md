# Building Device Bridge v2

This document describes how to build Device Bridge v2 from source.

## Prerequisites

### Go

- Go 1.22 or later
- See [go.mod](../go.mod) for exact version

### System Dependencies

#### Linux

```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install -y libusb-1.0-0-dev pkg-config

# Fedora/RHEL/CentOS
sudo dnf install libusb-devel pkgconfig

# Arch Linux
sudo pacman -S libusb pkgconf
```

#### macOS

```bash
# Using Homebrew
brew install libusb pkg-config
```

#### Windows

- libusb is included in the gousb package for Windows
- No additional setup required

## Building

### Build All Binaries

```bash
make build
```

This builds:
- `bin/bridge` - Main device bridge daemon
- `bin/bridge-cli` - CLI administration tool

### Build for Specific Platform

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 make build

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 make build

# Windows AMD64
GOOS=windows GOARCH=amd64 make build
```

### Build All Platforms

```bash
make build-all
```

Outputs:
- `bin/bridge-linux-amd64`
- `bin/bridge-linux-arm64`
- `bin/bridge-darwin-amd64`
- `bin/bridge-darwin-arm64`
- `bin/bridge-windows-amd64.exe`
- `bin/bridge-windows-arm64.exe`

## Testing

### Unit Tests

```bash
make test-unit
```

### Integration Tests

```bash
make test-integration
```

### All Tests

```bash
make test
```

### Test Coverage

```bash
make test-coverage
```

Generates:
- `coverage.out` - Coverage profile
- `coverage.html` - HTML coverage report

### Benchmarks

```bash
make test-bench
```

## Development Workflow

### Format Code

```bash
make fmt
```

### Lint Code

```bash
make lint
```

Requires: `golangci-lint` (install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

### Vet Code

```bash
make vet
```

### Generate Protobuf

```bash
make proto
```

Requires: `protoc` and Go plugins

### Local CI Pipeline

```bash
make ci
```

Runs: lint, test, build (simulates GitHub Actions locally)

## Docker

### Build Docker Image

```bash
make docker
```

or

```bash
docker build -t device-bridge:latest .
```

### Run Docker Container

```bash
docker run -p 50051:50051 -p 8080:8080 device-bridge:latest
```

With config:

```bash
docker run -p 50051:50051 -p 8080:8080 \
  -v $(pwd)/configs/production.yaml:/etc/device-bridge/config.yaml \
  device-bridge:latest
```

## USB Device Access (Linux)

For USB HID scanners to work without root, add udev rules:

```bash
# Create udev rule
sudo tee /etc/udev/rules.d/99-barcode-scanners.rules <<EOF
# Symbol/Zebra scanners
SUBSYSTEM=="usb", ATTRS{idVendor}=="05e0", MODE="0666"

# Honeywell scanners
SUBSYSTEM=="usb", ATTRS{idVendor}=="0c2e", MODE="0666"

# Datalogic scanners
SUBSYSTEM=="usb", ATTRS{idVendor}=="05f9", MODE="0666"

# All HID devices (use with caution)
# SUBSYSTEM=="usb", ATTR{bInterfaceClass}=="03", MODE="0666"
EOF

# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger
```

Alternatively, run as root (not recommended for production):

```bash
sudo ./bin/bridge
```

## Troubleshooting

### libusb not found

**Error**: `Package libusb-1.0 was not found in the pkg-config search path`

**Solution**:
```bash
# Install libusb development package (see Prerequisites above)
sudo apt-get install libusb-1.0-0-dev pkg-config
```

### USB Permission Denied

**Error**: `failed to open USB device: libusb: access denied`

**Solution**:
1. Add udev rules (see USB Device Access section)
2. Or run with sudo (not recommended)
3. Or add your user to `plugdev` group:
   ```bash
   sudo usermod -a -G plugdev $USER
   # Logout and login again
   ```

### Cannot find USB device

**Error**: `no matching USB HID scanner found`

**Solution**:
1. Check device is connected: `lsusb`
2. Check device permissions: `ls -l /dev/bus/usb/*/*`
3. Try enumerate scanners:
   ```bash
   bridge-cli devices discover
   ```

## IDE Setup

### VS Code

Recommended extensions:
- Go (golang.go)
- Protobuf (zxh404.vscode-proto3)

Settings (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.vetOnSave": "workspace"
}
```

### GoLand/IntelliJ IDEA

- Enable Go modules: Settings → Go → Go Modules → Enable
- Set GOROOT: Settings → Go → GOROOT
- Enable go fmt on save: Settings → Tools → File Watchers

## Continuous Integration

GitHub Actions workflow: `.github/workflows/ci.yml`

Runs on:
- push to main
- pull requests
- tags (triggers release)

## Release Process

### Create Release

```bash
# Tag release
git tag -a v0.3.0 -m "Release v0.3.0 - Phase 3 USB HID Scanner"
git push origin v0.3.0
```

GitHub Actions will:
1. Build multi-platform binaries
2. Create Docker images
3. Publish GitHub release
4. Upload artifacts

### Manual Release (using GoReleaser)

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Create release
goreleaser release --clean
```

## Performance Profiling

### CPU Profile

```bash
go test -cpuprofile=cpu.prof -bench=. ./internal/drivers/scanner_hid/...
go tool pprof cpu.prof
```

### Memory Profile

```bash
go test -memprofile=mem.prof -bench=. ./internal/drivers/scanner_hid/...
go tool pprof mem.prof
```

### Benchstat Comparison

```bash
# Baseline
go test -bench=. ./internal/drivers/scanner_hid/... > old.txt

# After changes
go test -bench=. ./internal/drivers/scanner_hid/... > new.txt

# Compare
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

## License

Apache 2.0 - See [LICENSE](../LICENSE) for details.
