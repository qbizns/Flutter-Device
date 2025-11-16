# Flutter Device Bridge SDK

Official Flutter/Dart SDK for Device Bridge - Universal hardware integration for Flutter applications.

## Overview

The Flutter Device Bridge SDK provides type-safe, idiomatic Dart/Flutter interfaces to all Device Bridge hardware capabilities. Integrate printers, scanners, RFID readers, access control systems, badge printers, and payment terminals into your Flutter applications with ease.

## Features

- ✅ **Type-Safe**: Full type safety with comprehensive Dart models
- ✅ **Real-Time Events**: WebSocket streaming for scanner, RFID, and access control events
- ✅ **Cross-Platform**: Works on Android, iOS, Web, Windows, macOS, and Linux
- ✅ **Easy to Use**: Intuitive APIs with builder patterns and helper methods
- ✅ **Comprehensive**: Supports 6 major hardware categories with 60+ operations
- ✅ **Production Ready**: Error handling, retries, and timeout support
- ✅ **Flutter Widgets**: Pre-built UI components for common operations
- ✅ **Device Logging**: Automatic logging with on/off toggle for audit trails
- ✅ **Environment Configuration**: .env file support for easy configuration

## Supported Hardware

| Hardware Type | Features | Status |
|--------------|----------|--------|
| **ESC/POS Printers** | Receipts, barcodes, QR codes, images | ✅ Complete |
| **Barcode/QR Scanners** | 1D/2D codes, continuous mode, events | ✅ Complete |
| **RFID/NFC Readers** | 19+ card types, authentication, encoding | ✅ Complete |
| **Access Control** | Doors, locks, credentials, lockdown | ✅ Complete |
| **Badge Printers** | Cards, magnetic stripe, RFID encoding | ✅ Complete |
| **Payment Terminals** | Swipe, chip, contactless, refunds | ✅ Complete |

## Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  flutter_device_bridge: ^1.0.0
```

Then run:

```bash
flutter pub get
```

## Quick Start

### 1. Setup Environment Configuration

Copy the example environment file:

```bash
cd sdk/flutter
cp .env.example .env
```

Edit `.env` to configure your Device Bridge connection:

```env
# Device Bridge Configuration
DEVICE_BRIDGE_BASE_URL=http://localhost:8080
DEVICE_BRIDGE_DEBUG=true

# Device Interaction Logging (ON/OFF)
DEVICE_LOGGING_ENABLED=true
PRINTER_LOGGING_ENABLED=true
SCANNER_LOGGING_ENABLED=true
```

### 2. Initialize in Your App

```dart
import 'package:flutter/material.dart';
import 'package:flutter_device_bridge/flutter_device_bridge.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize environment configuration
  await EnvConfig.initialize();

  runApp(MyApp());
}
```

### 3. Initialize the Client

```dart
// Using environment configuration
final client = DeviceBridgeClient(
  baseUrl: EnvConfig.baseUrl,
  debug: EnvConfig.debug,
  enableDeviceLogging: EnvConfig.deviceLoggingEnabled,
);

// Or configure manually
final client = DeviceBridgeClient(
  baseUrl: 'http://localhost:8080',
  debug: true,
  enableDeviceLogging: true,
);
```

### 4. Use Hardware Services

> **Note:** All device interactions are automatically logged when enabled. Logs are saved in device-specific formats (TXT for receipts, CSV for scans/weights/transactions).

#### Print a Receipt

```dart
final printer = client.printer('printer-01');

final receipt = ReceiptBuilder()
  .header('My Store')
  .text('123 Main St')
  .text('City, ST 12345')
  .divider()
  .keyValue('Item', 'Price')
  .keyValue('Coffee', '\$3.50')
  .keyValue('Tax', '\$0.30')
  .divider()
  .keyValue('Total', '\$3.80')
  .lineFeed(lines: 2)
  .qrCode('https://example.com/receipt/123')
  .text('Thank you!', alignment: TextAlignment.center)
  .cut()
  .build();

final result = await printer.print(receipt);
print('Print job: ${result.jobId}');

// Receipt is automatically logged to:
// device_logs/receipts/printer-01_2025-11-16.txt
```

#### Scan a Barcode

```dart
final scanner = client.scanner('scanner-01');

// One-time scan
final result = await scanner.scan(
  options: ScanOptions(
    timeout: Duration(seconds: 10),
    expectedSymbologies: [BarcodeSymbology.qr, BarcodeSymbology.code128],
  ),
);
print('Scanned: ${result.data}');

// Scan is automatically logged to:
// device_logs/scans/scanner-01_2025-11-16.csv

// Continuous scanning with events
await for (final event in scanner.subscribeScans()) {
  print('Scanned: ${event.scan.data}');
}
```

#### Read an RFID Card

```dart
final reader = client.rfidReader('reader-01');

final response = await reader.readCard(
  timeout: Duration(seconds: 30),
  expectedTypes: [CardType.mifareClassic1K, CardType.hidProx],
);

print('Card UID: ${response.card.uid}');
print('Card Type: ${response.card.type}');
print('Technology: ${response.card.technology}');
```

#### Control Access

```dart
final controller = client.accessControl('controller-01');

// Grant access with card credential
final result = await controller.grantAccess(
  doorId: 'main-entrance',
  credential: Credential.card(cardUid: '01020304'),
);

if (result.granted) {
  print('Access granted to ${result.userInfo?.userName}');
}

// Subscribe to access events
await for (final event in controller.subscribeAccessEvents()) {
  print('Event: ${event.type} at ${event.doorId}');
}
```

#### Print a Badge

```dart
final badgePrinter = client.badgePrinter('zebra-01');

final layout = CardLayout(
  cardSize: CardSize.cr80,
  orientation: CardOrientation.landscape,
  elements: [
    BadgeTextElement(
      position: Position(x: 100, y: 150),
      size: Size(width: 400, height: 50),
      text: 'John Doe',
      fontSize: 24,
      bold: true,
    ),
    BadgeImageElement(
      position: Position(x: 50, y: 50),
      size: Size(width: 150, height: 150),
      imageData: photoBase64,
    ),
  ],
);

final result = await badgePrinter.printBadge(
  BadgePrintJob(
    frontLayout: layout,
    quality: PrintQuality.high,
  ),
);
```

#### Process a Payment

```dart
final terminal = client.paymentTerminal('terminal-01');

final request = PaymentRequest(
  amount: 1050, // $10.50 in cents
  currency: 'USD',
  type: TransactionType.sale,
  orderId: 'ORDER-123',
);

final result = await terminal.processPayment(request);

if (result.payment.isApproved) {
  print('Payment approved: ${result.payment.approvalCode}');
  print('Card: ${result.payment.cardData?.brand}');
}
```

## Flutter Widgets

The SDK includes pre-built widgets for common operations:

### ScannerListener

```dart
ScannerListener(
  scanner: client.scanner('scanner-01'),
  onScan: (event) {
    setState(() {
      scannedCode = event.scan.data;
    });
  },
  child: YourAppWidget(),
)
```

### RFIDReaderWidget

```dart
RFIDReaderWidget(
  readerService: client.rfidReader('reader-01'),
  onCardRead: (response) {
    print('Card: ${response.card.uid}');
  },
)
```

### AccessControlWidget

```dart
AccessControlWidget(
  accessControlService: client.accessControl('controller-01'),
  doorId: 'main-entrance',
  onAccessGranted: (response) {
    print('Access granted');
  },
)
```

### DeviceStatusWidget

```dart
DeviceStatusWidget(
  deviceInfo: DeviceInfo(
    deviceId: 'printer-01',
    name: 'Receipt Printer',
    kind: 'printer',
    status: DeviceStatus.ready,
  ),
)
```

## Device Interaction Logging

All device interactions can be automatically logged for audit trails, debugging, and compliance:

### Features

- **Multiple Formats**: TXT for receipts, CSV for structured data, JSON for APIs
- **Per-Device Control**: Enable/disable logging for specific device types
- **Automatic Rotation**: Files rotated based on size limits
- **Log Retention**: Configurable retention periods and cleanup
- **Privacy-Friendly**: Completely disable when not needed

### Configuration

```env
# Enable/disable all logging
DEVICE_LOGGING_ENABLED=true

# Per-device type settings
PRINTER_LOGGING_ENABLED=true
PRINTER_LOG_FORMAT=txt
PRINTER_LOG_PATH=device_logs/receipts

SCANNER_LOGGING_ENABLED=true
SCANNER_LOG_FORMAT=csv
SCANNER_LOG_PATH=device_logs/scans

# Log retention
LOG_MAX_FILE_SIZE_MB=10
LOG_RETENTION_DAYS=30
```

### Accessing Logs

```dart
final logger = await DeviceInteractionLogger.getInstance();

// Get log files for a device
final files = await logger.getLogFilesForDevice('printer-01');

// Get log directory
final dir = await logger.getLogDirectoryForType(DeviceInteractionType.scan);
print('Logs at: ${dir.path}');

// Clear logs
await logger.clearLogsForDevice('scanner-01');
await logger.clearAllLogs();
```

### Log Formats

| Device Type | Format | Example |
|-------------|--------|---------|
| Printer | TXT | Full receipt content, human-readable |
| Scanner | CSV | Timestamp, barcode data, symbology |
| Scale | CSV | Timestamp, weight, unit |
| RFID | CSV | Timestamp, card UID, type, operation |
| Access Control | CSV | Timestamp, door, credential, result |
| Payment | CSV | Timestamp, transaction details |

For complete logging documentation, see [LOGGING.md](./LOGGING.md).

## Advanced Usage

### Error Handling

```dart
try {
  final result = await printer.print(receipt);
} on DeviceNotFoundException catch (e) {
  print('Printer not found: ${e.deviceId}');
} on DeviceBusyException catch (e) {
  print('Printer is busy');
} on DeviceTimeoutException catch (e) {
  print('Operation timed out after ${e.timeout.inSeconds}s');
} on DeviceBridgeException catch (e) {
  print('Error: ${e.message}');
}
```

### Custom Configuration

```dart
final config = DeviceBridgeConfig(
  baseUrl: 'https://device-bridge.example.com',
  apiKey: 'your-api-key',
  useTls: true,
  timeout: Duration(seconds: 60),
  maxRetries: 3,
  retryDelay: Duration(seconds: 2),
  headers: {
    'X-Custom-Header': 'value',
  },
);

final client = DeviceBridgeClient.fromConfig(config);
```

### Batch Operations

```dart
// Process multiple print jobs
final jobs = [receipt1, receipt2, receipt3];

for (final job in jobs) {
  try {
    await printer.print(job);
  } catch (e) {
    print('Failed to print: $e');
  }
}
```

### Real-Time Event Streams

```dart
// Combine multiple event streams
Stream<String> getAllEvents() async* {
  final scanStream = scanner.subscribeScans();
  final rfidStream = reader.subscribeCardReads();
  final accessStream = controller.subscribeAccessEvents();

  await for (final event in scanStream) {
    yield 'Scan: ${event.scan.data}';
  }
}
```

## API Reference

### Client

- `DeviceBridgeClient(...)` - Create a new client instance
- `client.printer(deviceId)` - Get printer service
- `client.scanner(deviceId)` - Get scanner service
- `client.rfidReader(deviceId)` - Get RFID reader service
- `client.accessControl(deviceId)` - Get access control service
- `client.badgePrinter(deviceId)` - Get badge printer service
- `client.paymentTerminal(deviceId)` - Get payment terminal service

### Services

Each service provides methods specific to its hardware type. See the [API documentation](https://pub.dev/documentation/flutter_device_bridge/latest/) for complete details.

## Examples

Complete example applications are available in the `/example` directory:

- `example/basic_app` - Basic integration examples
- `example/retail_pos` - Retail point-of-sale application
- `example/access_control_demo` - Access control system demo
- `example/badge_printing` - Badge printing kiosk

## Testing

The SDK includes comprehensive test coverage:

```bash
# Run all tests
flutter test

# Run specific test file
flutter test test/services/printer_service_test.dart

# Run with coverage
flutter test --coverage
```

## Troubleshooting

### Connection Issues

If you're having trouble connecting to Device Bridge:

1. Verify the server is running: `curl http://localhost:8080/health`
2. Check firewall settings
3. Enable debug logging: `DeviceBridgeClient(debug: true)`
4. Review logs in Device Bridge server

### Device Not Found

If devices aren't being detected:

1. Check device is properly connected
2. Verify device driver is installed
3. Check Device Bridge device configuration
4. Use Device Bridge CLI to list devices: `device-bridge devices list`

### Timeout Errors

If operations are timing out:

1. Increase timeout: `await printer.print(job, timeout: Duration(minutes: 2))`
2. Check device is powered on and ready
3. Verify network latency is acceptable
4. Review Device Bridge server logs

## Performance

The SDK is optimized for production use:

- **HTTP Connection Pooling**: Reuses connections for better performance
- **WebSocket Streaming**: Efficient real-time event delivery
- **Lazy Loading**: Services are created on demand
- **Memory Efficient**: Immutable models with const constructors
- **Null Safety**: Full null safety support

## Security

Security best practices:

- **Use TLS**: Always use HTTPS/WSS in production
- **API Keys**: Protect your API keys (use environment variables)
- **Network Security**: Run Device Bridge on isolated network
- **Input Validation**: SDK validates all inputs
- **Error Messages**: Sensitive data is not exposed in errors

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Support

- **Documentation**: https://docs.devicebridge.io
- **API Reference**: https://pub.dev/documentation/flutter_device_bridge
- **Issues**: https://github.com/your-org/flutter-device/issues
- **Discussions**: https://github.com/your-org/flutter-device/discussions

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and migration guides.

---

Made with ❤️ by the Device Bridge team
