# Serial Scale Hardware Validation Script for Windows
# Tests serial scale functionality on Windows
#
# Usage: .\test-serial-scale.ps1 [port] [protocol]
# Example: .\test-serial-scale.ps1 COM1 mtsics

param(
    [string]$Port = "COM1",
    [string]$Protocol = "auto"
)

Write-Host "=========================================" -ForegroundColor Green
Write-Host "Serial Scale Hardware Validation (Windows)" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Target Port: $Port"
Write-Host "Protocol: $Protocol"
Write-Host ""

# Step 1: Check if running as Administrator
Write-Host "Step 1: Checking permissions..."
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if ($isAdmin) {
    Write-Host "✓ Running as Administrator" -ForegroundColor Green
} else {
    Write-Host "⚠ Not running as Administrator (may be OK)" -ForegroundColor Yellow
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

# Step 3: Check serial port
Write-Host "Step 3: Checking serial port..."
$portExists = Get-CimInstance -ClassName Win32_SerialPort | Where-Object { $_.DeviceID -eq $Port }
if ($portExists) {
    Write-Host "✓ Port found: $Port" -ForegroundColor Green
    Write-Host "  Description: $($portExists.Description)"
} else {
    Write-Host "✗ Port not found: $Port" -ForegroundColor Red
    Write-Host "Available serial ports:" -ForegroundColor Yellow
    Get-CimInstance -ClassName Win32_SerialPort | Format-Table DeviceID, Description
    exit 1
}
Write-Host ""

# Step 4: Build Device Bridge
Write-Host "Step 4: Building Device Bridge..."
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

# Step 5: Create test configuration
Write-Host "Step 5: Creating test configuration..."
$config = @"
server:
  grpc_port: 50051
  http_port: 8080
  metrics_port: 9090

devices:
  - id: scale-test
    name: "Serial Scale Hardware Test"
    kind: scale.serial
    enabled: true
    metadata:
      port: "$Port"
      baud_rate: 9600
      data_bits: 8
      parity: "none"
      stop_bits: 1
      protocol: "$Protocol"
      read_timeout: 1000
      retry_attempts: 3
      retry_delay: 100
      preferred_unit: "kg"

logging:
  level: debug
  format: text

telemetry:
  metrics: true
  tracing: false
"@
$config | Out-File -FilePath "$env:TEMP\test-scale-config.yaml" -Encoding UTF8
Write-Host "✓ Config created: $env:TEMP\test-scale-config.yaml" -ForegroundColor Green
Write-Host ""

# Step 6: Test scale connectivity
Write-Host "Step 6: Testing scale connectivity..."
Write-Host "Starting Device Bridge..."

# Start Device Bridge
$logFile = "$env:TEMP\scale-test.log"
$process = Start-Process -FilePath ".\bin\device-bridge.exe" -ArgumentList "-config", "$env:TEMP\test-scale-config.yaml" -RedirectStandardOutput $logFile -RedirectStandardError $logFile -PassThru -NoNewWindow

# Wait for startup
Start-Sleep -Seconds 3

# Check if process is running
if ($process.HasExited) {
    Write-Host "✗ Device Bridge failed to start" -ForegroundColor Red
    Write-Host "Log output:"
    Get-Content $logFile
    exit 1
}

Write-Host "✓ Device Bridge started (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

# Step 7: Run validation tests
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Step 7: Running Validation Tests" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Function to make gRPC request (requires grpcurl.exe)
function Invoke-GrpcRequest {
    param($Method, $Data)
    $result = & grpcurl -plaintext -d $Data localhost:50051 "devicebridge.v1.DeviceBridge/$Method" 2>&1
    return $result -join "`n"
}

# Check if grpcurl is available
$grpcurlAvailable = Get-Command grpcurl -ErrorAction SilentlyContinue
if (-not $grpcurlAvailable) {
    Write-Host "⚠ grpcurl not found - skipping API tests" -ForegroundColor Yellow
    Write-Host "Download from: https://github.com/fullstorydev/grpcurl/releases" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Manual testing instructions:" -ForegroundColor Yellow
    Write-Host "1. Put weight on scale"
    Write-Host "2. Check logs at: $logFile"
    Write-Host "3. Look for weight readings in logs"
    Write-Host ""
} else {
    # Test 1: Check device status
    Write-Host "Test 1: Device Status" -ForegroundColor Blue
    $status = Invoke-GrpcRequest "GetDeviceStatus" '{"device_id": "scale-test"}'
    if ($status -match "READY|ONLINE") {
        Write-Host "✓ Device is online" -ForegroundColor Green
    } else {
        Write-Host "⚠ Device status: $status" -ForegroundColor Yellow
    }
    Write-Host ""

    # Test 2: Read weight (immediate)
    Write-Host "Test 2: Read Weight (Immediate)" -ForegroundColor Blue
    Write-Host "Reading current weight..."
    $weight = Invoke-GrpcRequest "ReadWeight" '{"device_id": "scale-test", "stable_only": false}'
    if ($weight -match "value") {
        Write-Host "✓ Weight reading successful" -ForegroundColor Green
        $weight
    } else {
        Write-Host "⚠ Weight reading failed" -ForegroundColor Yellow
        $weight
    }
    Write-Host ""

    # Test 3: Read stable weight
    Write-Host "Test 3: Read Stable Weight" -ForegroundColor Blue
    Write-Host "Waiting for stable weight (put weight on scale)..."
    $stable = Invoke-GrpcRequest "ReadWeight" '{"device_id": "scale-test", "stable_only": true}'
    if ($stable -match "stable.*true") {
        Write-Host "✓ Stable weight reading successful" -ForegroundColor Green
        $stable
    } else {
        Write-Host "⚠ No stable weight detected" -ForegroundColor Yellow
    }
    Write-Host ""

    # Test 4: Zero operation
    Write-Host "Test 4: Zero Scale" -ForegroundColor Blue
    $zero = Invoke-GrpcRequest "ZeroScale" '{"device_id": "scale-test"}'
    Write-Host "✓ Zero command sent" -ForegroundColor Green
    Start-Sleep -Seconds 1
    Write-Host ""

    # Test 5: Tare operation
    Write-Host "Test 5: Tare Scale" -ForegroundColor Blue
    Write-Host "Place container on scale and press Enter..."
    Read-Host
    $tare = Invoke-GrpcRequest "TareScale" '{"device_id": "scale-test"}'
    Write-Host "✓ Tare command sent" -ForegroundColor Green
    Start-Sleep -Seconds 1
    Write-Host ""

    # Test 6: Continuous reading
    Write-Host "Test 6: Continuous Reading (10 samples)" -ForegroundColor Blue
    $successCount = 0
    for ($i = 1; $i -le 10; $i++) {
        $reading = Invoke-GrpcRequest "ReadWeight" '{"device_id": "scale-test", "stable_only": false}'
        if ($reading -match "value") {
            $successCount++
            Write-Host "  Reading $i: Success" -ForegroundColor Green
        } else {
            Write-Host "  Reading $i: Failed" -ForegroundColor Yellow
        }
        Start-Sleep -Milliseconds 500
    }
    Write-Host "✓ Successful readings: $successCount/10" -ForegroundColor Green
    Write-Host ""
}

# Cleanup
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Cleaning up..." -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
Write-Host "✓ Device Bridge stopped" -ForegroundColor Green
Write-Host ""

# Summary
Write-Host "=========================================" -ForegroundColor Green
Write-Host "VALIDATION SUMMARY" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Port: $Port"
Write-Host "Protocol: $Protocol"
Write-Host ""
Write-Host "Next steps:"
Write-Host "1. Review logs: $logFile"
Write-Host "2. Test with different protocols if needed"
Write-Host "3. Test continuous operation"
Write-Host "4. Test auto-reconnection"
Write-Host ""

# Offer to show logs
$showLogs = Read-Host "View detailed logs? (y/n)"
if ($showLogs -eq "y") {
    notepad $logFile
}
