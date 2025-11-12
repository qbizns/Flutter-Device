# Device Bridge v2 - Windows Service Uninstallation Script
# PowerShell script to uninstall Device Bridge Windows Service
# Requires Administrator privileges

#Requires -RunAsAdministrator

param(
    [string]$InstallPath = "C:\Program Files\Device Bridge",
    [string]$ServiceName = "DeviceBridge"
)

$ErrorActionPreference = "Stop"

Write-Host "Device Bridge v2 - Windows Service Uninstallation" -ForegroundColor Yellow
Write-Host "==================================================" -ForegroundColor Yellow
Write-Host ""

# Confirm uninstallation
$confirm = Read-Host "Are you sure you want to uninstall Device Bridge? (yes/no)"
if ($confirm -ne "yes") {
    Write-Host "Uninstallation cancelled" -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "Step 1: Stopping service..." -ForegroundColor Yellow
try {
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    Write-Host "Service stopped" -ForegroundColor Green
} catch {
    Write-Host "Service was not running" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Step 2: Removing service..." -ForegroundColor Yellow
try {
    & sc.exe delete $ServiceName
    Write-Host "Service removed" -ForegroundColor Green
} catch {
    Write-Host "Warning: Could not remove service" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Step 3: Removing firewall rules..." -ForegroundColor Yellow
try {
    Remove-NetFirewallRule -DisplayName "Device Bridge API" -ErrorAction SilentlyContinue
    Write-Host "Firewall rules removed" -ForegroundColor Green
} catch {
    Write-Host "Warning: Could not remove firewall rules" -ForegroundColor Yellow
}

Write-Host ""
$removeFiles = Read-Host "Remove all files and configuration? (yes/no)"
if ($removeFiles -eq "yes") {
    Write-Host "Step 4: Removing files..." -ForegroundColor Yellow
    if (Test-Path $InstallPath) {
        Remove-Item -Path $InstallPath -Recurse -Force
        Write-Host "Removed installation directory" -ForegroundColor Green
    }
} else {
    Write-Host "Files preserved in: $InstallPath" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Uninstallation complete!" -ForegroundColor Green
