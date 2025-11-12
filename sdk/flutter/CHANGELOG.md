# Changelog

All notable changes to the Flutter Device Bridge SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-12

### Added

#### Core Infrastructure
- `DeviceBridgeClient` - Main client class with HTTP/gRPC/WebSocket support
- `DeviceBridgeConfig` - Flexible configuration for different environments
- Exception hierarchy with specific error types:
  - `DeviceNotFoundException`
  - `DeviceTimeoutException`
  - `DeviceBusyException`
  - `DeviceErrorException`
  - `AuthenticationException`
  - `AuthorizationException`
  - `ValidationException`
  - `ConnectionException`
  - `UnsupportedOperationException`

#### Common Models
- `DeviceStatus` enum and `DeviceInfo` class
- `Position`, `Size`, `RGBColor` for UI positioning
- `Timestamp` wrapper for consistent date handling
- `PaginationInfo` for paginated results
- `Result<T>` generic wrapper for success/failure handling

#### RFID/NFC Reader Service
- Support for 19+ card types (MIFARE, HID, EM4100, NTAG, etc.)
- `readCard()` - Read card with timeout and type filtering
- `authenticateCard()` - Authenticate with keys
- `writeCard()` - Write data with verification
- `getStatus()` - Get reader status
- `setMode()` - Configure reader mode
- `subscribeCardReads()` - Real-time card read events via WebSocket
- `getSupportedCardTypes()` - Query capabilities

#### Access Control Service
- `unlockDoor()` - Unlock with optional duration
- `lockDoor()` - Lock door
- `getDoorStatus()` - Get door status (lock state, position, mode)
- `checkAccess()` - Pre-validate credentials
- `grantAccess()` - Grant access with credentials (card, PIN, card+PIN)
- `setDoorMode()` - Set operating mode (normal, locked, unlocked, office, lockdown, etc.)
- `activateLockdown()` - Activate emergency lockdown (partial, full, evacuation)
- `deactivateLockdown()` - Deactivate lockdown
- `subscribeAccessEvents()` - Real-time access events via WebSocket
- `getCredentialInfo()` - Get user/credential information
- `getRecentEvents()` - Query event history
- `emergencyOverride()` - Emergency door control

#### Printer Service (ESC/POS)
- `print()` - Print job with multiple elements
- `printText()` - Quick text printing
- `printReceipt()` - Print using builder pattern
- `printBarcode()` - Print 1D barcodes (UPC-A, EAN-13, Code 128, etc.)
- `printQRCode()` - Print QR codes
- `printImage()` - Print images
- `feedPaper()` - Paper feed control
- `cutPaper()` - Cut paper (full or partial)
- `openCashDrawer()` - Open cash drawer
- `ReceiptBuilder` - Fluent API for receipt creation with:
  - Text with alignment, size, bold, underline, invert
  - Line feeds
  - Barcodes and QR codes
  - Images
  - Dividers and headers
  - Key-value pairs

#### Badge Printer Service
- `printBadge()` - Print badges with front/back layouts
- `printFromTemplate()` - Template-based printing
- `encodeMagneticStripe()` - Magnetic stripe encoding with verification
- `encodeRFID()` - RFID encoding with verification
- Badge layout with positioned elements:
  - `BadgeTextElement` - Text with fonts, colors, alignment
  - `BadgeImageElement` - Images with aspect ratio control
  - `BadgeBarcodeElement` - Barcodes
  - `BadgeQRCodeElement` - QR codes
- Support for card sizes: CR80, CR79, CR100
- Print quality settings: draft, normal, high
- Duplex printing, overlay application
- Template management (list, get, save, delete)

#### Scanner Service
- `scan()` - One-time scan with timeout
- `startContinuousScan()` - Continuous scanning mode
- `stopContinuousScan()` - Stop continuous mode
- `setMode()` - Set scanner mode (manual, continuous, trigger)
- `subscribeScans()` - Real-time scan events via WebSocket
- Support for symbologies:
  - 1D: UPC-A, UPC-E, EAN-13, EAN-8, Code 39/93/128, ITF, Codabar
  - 2D: QR Code, Data Matrix, PDF417, Aztec
- `setSymbologies()` - Enable/disable specific symbologies
- `setAimingLaser()` - Control aiming laser
- `setBeep()` - Configure scan beep
- `setVibration()` - Configure scan vibration
- `getRecentScans()` - Query scan history

#### Payment Terminal Service
- `processPayment()` - Process sale/authorization
- `processRefund()` - Process full or partial refunds
- `voidTransaction()` - Void same-day transactions
- `adjustTip()` - Add tip to existing transaction
- `cancelPayment()` - Cancel in-progress payment
- Support for payment methods:
  - Card swipe (magnetic stripe)
  - Chip (EMV)
  - Contactless (NFC)
  - Manual entry
  - QR code payments
- `getTransaction()` - Get transaction details
- `getRecentTransactions()` - Query transaction history
- `closeBatch()` - End-of-day settlement
- `getBatchSummary()` - Current batch information
- `displayMessage()` - Show message on terminal screen
- `printReceipt()` - Print transaction receipt
- `getReceiptData()` - Get receipt data (text or HTML)

#### Flutter Widgets
- `ScannerListener` - Auto-managing scanner event listener
- `ScanResultDisplay` - Animated scan result display
- `RFIDReaderWidget` - Complete RFID reading UI
- `RFIDCardListener` - Auto-managing RFID event listener
- `AccessControlWidget` - Door control UI with status
- `AccessControlEventListener` - Auto-managing access event listener
- `DeviceStatusWidget` - Device status display with icon
- `DeviceStatusCard` - Detailed device status card
- `DeviceOperationLoader` - Loading indicator for operations

### Technical Features
- Full null safety support
- Immutable models with const constructors
- Factory constructors for all models from JSON
- Type-safe enum parsing with fallback to unknown
- WebSocket streaming for real-time events
- Automatic error handling and retry logic
- Consistent API design across all services
- Builder patterns for complex operations
- Comprehensive toString() implementations
- Platform support: Android, iOS, Web, Windows, macOS, Linux

### Dependencies
- `http: ^1.1.0` - HTTP client
- `grpc: ^3.2.0` - gRPC support
- `protobuf: ^3.1.0` - Protocol buffers
- `web_socket_channel: ^2.4.0` - WebSocket streaming
- `json_annotation: ^4.8.1` - JSON serialization
- `rxdart: ^0.27.7` - Reactive extensions
- `logging: ^1.2.0` - Logging support

### Dev Dependencies
- `build_runner: ^2.4.0` - Code generation
- `json_serializable: ^6.7.0` - JSON serialization generator
- `mockito: ^5.4.0` - Mocking for tests
- `test: ^1.24.0` - Testing framework
- `flutter_lints: ^3.0.0` - Linting rules

## [Unreleased]

### Planned
- Additional widgets for payment terminal UI
- Batch operations helper methods
- Offline queue for failed operations
- Advanced caching strategies
- gRPC streaming support
- Protobuf code generation integration
- Example applications for each hardware type
- Comprehensive test coverage
- CI/CD integration examples
- Performance benchmarks

---

[1.0.0]: https://github.com/your-org/flutter-device/releases/tag/v1.0.0
