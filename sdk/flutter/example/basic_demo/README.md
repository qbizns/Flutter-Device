# Device Bridge Example App

A comprehensive example Flutter application demonstrating all features of the Flutter Device Bridge SDK.

## Overview

This example app showcases integration with all six hardware types supported by Device Bridge:

1. **Printer** - ESC/POS receipt printing
2. **Scanner** - Barcode and QR code scanning
3. **RFID Reader** - RFID/NFC card reading
4. **Access Control** - Door control and access management
5. **Badge Printer** - ID card and badge printing
6. **Payment Terminal** - Payment processing

## Running the Example

### Prerequisites

1. Install Flutter SDK (>=3.0.0)
2. Device Bridge server running on `http://localhost:8080`
3. Hardware devices configured in Device Bridge

### Installation

```bash
cd example/basic_demo
flutter pub get
```

### Run on Different Platforms

```bash
# Android
flutter run -d android

# iOS
flutter run -d ios

# Web
flutter run -d chrome

# Desktop (macOS)
flutter run -d macos

# Desktop (Windows)
flutter run -d windows

# Desktop (Linux)
flutter run -d linux
```

## Features Demonstrated

### Printer Example

- Building receipts using `ReceiptBuilder`
- Printing text with alignment
- Adding dividers and line feeds
- Creating key-value pairs
- Adding QR codes
- Paper cutting

**Code Location**: `lib/main.dart` - `PrinterExamplePage`

### Scanner Example

- Real-time barcode/QR code scanning
- Using `ScannerListener` widget
- Displaying scan results
- Automatic event handling

**Code Location**: `lib/main.dart` - `ScannerExamplePage`

### RFID Reader Example

- Reading RFID/NFC cards
- Using `RFIDReaderWidget`
- Displaying card information (UID, type, technology)
- Error handling

**Code Location**: `lib/main.dart` - `RFIDExamplePage`

### Access Control Example

- Door control (lock/unlock)
- Using `AccessControlWidget`
- Real-time door status
- Access event handling

**Code Location**: `lib/main.dart` - `AccessControlExamplePage`

### Badge Printer Example

- Creating badge layouts with positioned elements
- Text elements with custom fonts and colors
- Printing high-quality badges
- Job status tracking

**Code Location**: `lib/main.dart` - `BadgePrinterExamplePage`

### Payment Terminal Example

- Processing sale transactions
- Displaying payment results
- Transaction status handling
- Card brand detection

**Code Location**: `lib/main.dart` - `PaymentTerminalExamplePage`

## Configuration

To connect to a different Device Bridge server, modify the client initialization in `lib/main.dart`:

```dart
client = DeviceBridgeClient(
  baseUrl: 'https://your-device-bridge-server.com',
  apiKey: 'your-api-key',
  useTls: true,
);
```

## Device Configuration

Make sure your Device Bridge server has the following devices configured:

- `printer-01` - ESC/POS printer
- `scanner-01` - Barcode/QR scanner
- `reader-01` - RFID/NFC reader
- `controller-01` - Access control controller with `main-entrance` door
- `zebra-01` - Badge/card printer
- `terminal-01` - Payment terminal

You can configure these using the Device Bridge CLI or API.

## Architecture

The example app follows Flutter best practices:

- **Stateful Widgets**: For managing UI state
- **Async/Await**: For asynchronous hardware operations
- **Error Handling**: Try-catch blocks with user feedback
- **Material Design**: Using Material Design 3 components
- **Navigation**: Using Navigator for page transitions
- **Widgets**: Leveraging SDK-provided widgets where appropriate

## Extending the Example

### Adding Your Own Features

1. Create a new page widget
2. Initialize the appropriate service from `client`
3. Implement your hardware operation
4. Add error handling
5. Update UI based on results

Example:

```dart
class MyCustomPage extends StatefulWidget {
  final DeviceBridgeClient client;

  const MyCustomPage({super.key, required this.client});

  @override
  State<MyCustomPage> createState() => _MyCustomPageState();
}

class _MyCustomPageState extends State<MyCustomPage> {
  Future<void> _performOperation() async {
    try {
      final service = widget.client.yourService('device-id');
      final result = await service.yourMethod();
      // Update UI
    } catch (e) {
      // Handle error
    }
  }

  @override
  Widget build(BuildContext context) {
    // Build your UI
  }
}
```

### Testing with Mock Data

For testing without physical hardware, you can:

1. Use Device Bridge simulation mode
2. Create mock implementations of services
3. Use test doubles with `mockito`

## Troubleshooting

### Connection Issues

**Problem**: Cannot connect to Device Bridge server

**Solutions**:
- Verify server is running: `curl http://localhost:8080/health`
- Check `baseUrl` in client initialization
- Ensure firewall allows connections
- Enable debug logging: `DeviceBridgeClient(debug: true)`

### Device Not Found

**Problem**: DeviceNotFoundException errors

**Solutions**:
- List available devices: `device-bridge devices list`
- Check device IDs match configuration
- Verify devices are connected and powered on
- Check Device Bridge server logs

### Widget Errors

**Problem**: Widget not displaying correctly

**Solutions**:
- Check widget is inside MaterialApp
- Verify context is valid
- Ensure async operations complete before updating UI
- Check console for error messages

## Learn More

- [Device Bridge SDK Documentation](../../README.md)
- [Device Bridge Server Documentation](https://docs.devicebridge.io)
- [Flutter Documentation](https://docs.flutter.dev)

## Support

For issues or questions:

- Open an issue on GitHub
- Check existing issues for solutions
- Review Device Bridge documentation
- Join the Device Bridge community

## License

This example is provided under the same license as the Device Bridge SDK (MIT License).
