# Device Bridge v2 - Windows Service Installation Script
# PowerShell script to install Device Bridge as a Windows Service
# Requires Administrator privileges

#Requires -RunAsAdministrator

param(
    [string]$InstallPath = "C:\Program Files\Device Bridge",
    [string]$ServiceName = "DeviceBridge",
    [string]$DisplayName = "Device Bridge v2",
    [string]$Description = "Hardware Device Integration Service for Payment Terminals, Scales, Scanners, and Printers"
)

$ErrorActionPreference = "Stop"

Write-Host "Device Bridge v2 - Windows Service Installation" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green
Write-Host ""

# Check if running as Administrator
$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Host "Error: This script must be run as Administrator" -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator'" -ForegroundColor Yellow
    exit 1
}

# Check if binary exists
if (-not (Test-Path ".\device-bridge.exe")) {
    Write-Host "Error: device-bridge.exe not found in current directory" -ForegroundColor Red
    Write-Host "Please build the Windows binary first: make build-windows" -ForegroundColor Yellow
    exit 1
}

Write-Host "Step 1: Creating installation directory..." -ForegroundColor Yellow
if (-not (Test-Path $InstallPath)) {
    New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    Write-Host "Created directory: $InstallPath" -ForegroundColor Green
} else {
    Write-Host "Directory already exists: $InstallPath" -ForegroundColor Green
}

Write-Host ""
Write-Host "Step 2: Copying files..." -ForegroundColor Yellow
Copy-Item ".\device-bridge.exe" -Destination "$InstallPath\device-bridge.exe" -Force
Write-Host "Copied device-bridge.exe" -ForegroundColor Green

# Create config directory
$ConfigPath = "$InstallPath\config"
if (-not (Test-Path $ConfigPath)) {
    New-Item -ItemType Directory -Path $ConfigPath -Force | Out-Null
}

# Copy configuration if exists
if (Test-Path ".\configs\payment\config.production.yaml") {
    Copy-Item ".\configs\payment\config.production.yaml" -Destination "$ConfigPath\config.yaml" -Force
    Write-Host "Copied configuration file" -ForegroundColor Green
}

# Create logs directory
$LogPath = "$InstallPath\logs"
if (-not (Test-Path $LogPath)) {
    New-Item -ItemType Directory -Path $LogPath -Force | Out-Null
}

# Create data directory
$DataPath = "$InstallPath\data"
if (-not (Test-Path $DataPath)) {
    New-Item -ItemType Directory -Path $DataPath -Force | Out-Null
}

Write-Host ""
Write-Host "Step 3: Creating Windows Service..." -ForegroundColor Yellow

# Check if service already exists
$existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($existingService) {
    Write-Host "Service already exists. Stopping and removing..." -ForegroundColor Yellow
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    & sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Create the service using NSSM (Non-Sucking Service Manager) or sc.exe
# Using sc.exe (built-in Windows tool)
$binaryPath = "`"$InstallPath\device-bridge.exe`" --config=`"$ConfigPath\config.yaml`" --log-level=info"

& sc.exe create $ServiceName binPath= $binaryPath DisplayName= $DisplayName start= auto
& sc.exe description $ServiceName $Description

# Set recovery options (restart on failure)
& sc.exe failure $ServiceName reset= 86400 actions= restart/5000/restart/10000/restart/30000

Write-Host "Created Windows Service: $ServiceName" -ForegroundColor Green

Write-Host ""
Write-Host "Step 4: Configuring firewall..." -ForegroundColor Yellow
# Add firewall rules for API port
try {
    New-NetFirewallRule -DisplayName "Device Bridge API" `
                        -Direction Inbound `
                        -Action Allow `
                        -Protocol TCP `
                        -LocalPort 8080 `
                        -ErrorAction SilentlyContinue | Out-Null
    Write-Host "Added firewall rule for port 8080" -ForegroundColor Green
} catch {
    Write-Host "Warning: Could not add firewall rule (may already exist)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Edit configuration: notepad `"$ConfigPath\config.yaml`""
Write-Host "  2. Set environment variables (if needed):"
Write-Host "     - MADA_TERMINAL_ID, MADA_MERCHANT_ID, MADA_AUTH_KEY"
Write-Host "     - KNET_TERMINAL_ID, KNET_MERCHANT_ID, KNET_AUTH_KEY"
Write-Host "  3. Start service: Start-Service -Name $ServiceName"
Write-Host "  4. Check status:  Get-Service -Name $ServiceName"
Write-Host "  5. View logs:     Get-Content `"$LogPath\device-bridge.log`" -Tail 50 -Wait"
Write-Host ""
Write-Host "Service Management Commands:" -ForegroundColor Cyan
Write-Host "  Start:   Start-Service -Name $ServiceName"
Write-Host "  Stop:    Stop-Service -Name $ServiceName"
Write-Host "  Restart: Restart-Service -Name $ServiceName"
Write-Host "  Status:  Get-Service -Name $ServiceName"
Write-Host ""
Write-Host "To uninstall, run: .\uninstall-service.ps1" -ForegroundColor Yellow
Write-Host ""

# Ask if user wants to start the service now
$start = Read-Host "Start the service now? (Y/N)"
if ($start -eq "Y" -or $start -eq "y") {
    Start-Service -Name $ServiceName
    Write-Host ""
    Write-Host "Service started successfully!" -ForegroundColor Green
    Get-Service -Name $ServiceName
}
