# USB HID Scanner Validation Script for Windows
# Tests USB HID scanner functionality on Windows
#
# Usage: .\test-usb-scanner.ps1 [vendor_id] [product_id]
# Example: .\test-usb-scanner.ps1 05e0 1200

param(
    [string]$VendorId = "05e0",
    [string]$ProductId = "1200"
)

Write-Host "=========================================" -ForegroundColor Green
Write-Host "USB HID Scanner Validation (Windows)" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Target Scanner: VID=$VendorId PID=$ProductId"
Write-Host ""

# Step 1: Check if running as Administrator
Write-Host "Step 1: Checking permissions..."
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if ($isAdmin) {
    Write-Host "✓ Running as Administrator" -ForegroundColor Green
} else {
    Write-Host "✗ Not running as Administrator" -ForegroundColor Red
    Write-Host "Please run PowerShell as Administrator" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

# Step 2: Check Go version
Write-Host "Step 2: Checking Go version..."
try {
    $goVersion = go version
    Write-Host "✓ Go installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go not installed" -ForegroundColor Red
    Write-Host "Install from: https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

# Step 3: Check for USB scanner
Write-Host "Step 3: Checking for USB scanner..."
$vidPid = "$VendorId`.+$ProductId"
$devices = Get-PnpDevice | Where-Object { $_.InstanceId -match "USB\\VID_$VendorId&PID_$ProductId" }
if ($devices) {
    Write-Host "✓ Scanner found:" -ForegroundColor Green
    foreach ($device in $devices) {
        Write-Host "  - $($device.FriendlyName) [$($device.Status)]" -ForegroundColor Cyan
    }
} else {
    Write-Host "⚠ Scanner not found (VID=$VendorId PID=$ProductId)" -ForegroundColor Yellow
    Write-Host "Available USB devices:" -ForegroundColor Yellow
    Get-PnpDevice -Class USB | Format-Table FriendlyName, Status, InstanceId
    exit 1
}
Write-Host ""

# Step 4: Check WinUSB driver
Write-Host "Step 4: Checking USB driver..."
$driver = Get-PnpDevice | Where-Object { $_.InstanceId -match "USB\\VID_$VendorId&PID_$ProductId" } | Get-PnpDeviceProperty -KeyName "DEVPKEY_Device_Service"
if ($driver.Data -eq "WinUSB") {
    Write-Host "✓ WinUSB driver installed" -ForegroundColor Green
} else {
    Write-Host "⚠ Driver: $($driver.Data)" -ForegroundColor Yellow
    Write-Host "WinUSB driver not detected. gousb requires WinUSB." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "To install WinUSB driver:" -ForegroundColor Yellow
    Write-Host "1. Download Zadig from https://zadig.akeo.ie/" -ForegroundColor Yellow
    Write-Host "2. Run as Administrator" -ForegroundColor Yellow
    Write-Host "3. Options → List All Devices" -ForegroundColor Yellow
    Write-Host "4. Select your scanner" -ForegroundColor Yellow
    Write-Host "5. Select WinUSB driver" -ForegroundColor Yellow
    Write-Host "6. Click 'Replace Driver'" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Do you want to continue anyway? (y/n)" -ForegroundColor Yellow
    $continue = Read-Host
    if ($continue -ne "y") {
        exit 1
    }
}
Write-Host ""

# Step 5: Build Device Bridge
Write-Host "Step 5: Building Device Bridge..."
Set-Location "$PSScriptRoot\..\.."
try {
    go build -o bin\device-bridge.exe .\cmd\bridge\
    Write-Host "✓ Build successful" -ForegroundColor Green
} catch {
    Write-Host "✗ Build failed" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit 1
}
Write-Host ""

# Step 6: Create test configuration
Write-Host "Step 6: Creating test configuration..."
$config = @"
server:
  grpc_port: 50051
  http_port: 8080
  metrics_port: 9090

devices:
  - id: scanner-test
    name: "USB HID Scanner Test"
    kind: scanner.hid
    vendor_id: 0x$VendorId
    product_id: 0x$ProductId
    buffer_size: 10
    reconnect_delay: 5s
    read_timeout: 1s

logging:
  level: debug
  format: text

telemetry:
  metrics: true
  tracing: false
"@
$config | Out-File -FilePath "$env:TEMP\test-scanner-config.yaml" -Encoding UTF8
Write-Host "✓ Config created: $env:TEMP\test-scanner-config.yaml" -ForegroundColor Green
Write-Host ""

# Step 7: Run Device Bridge
Write-Host "Step 7: Starting Device Bridge (press Ctrl+C to stop)..."
Write-Host "----------------------------------------" -ForegroundColor Cyan
Write-Host ""

try {
    .\bin\device-bridge.exe -config "$env:TEMP\test-scanner-config.yaml"
} catch {
    Write-Host ""
    Write-Host "✗ Device Bridge failed to start" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit 1
}

Write-Host ""
Write-Host "Test completed" -ForegroundColor Green
