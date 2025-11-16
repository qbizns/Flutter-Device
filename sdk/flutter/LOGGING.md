# Device Interaction Logging

The Flutter Device Bridge SDK includes comprehensive logging of all device interactions. This feature allows you to track and audit all hardware operations including printing receipts, scanning barcodes, weighing items, and more.

## Features

- **Automatic Logging**: All device interactions are automatically logged when enabled
- **Multiple Formats**: Support for TXT, CSV, and JSON log formats
- **Per-Device Configuration**: Enable/disable logging for specific device types
- **Log Rotation**: Automatic file rotation based on size limits
- **Log Retention**: Configurable retention periods and cleanup
- **Privacy-Friendly**: Logging can be completely disabled via configuration

## Table of Contents

1. [Quick Start](#quick-start)
2. [Configuration](#configuration)
3. [Log Formats](#log-formats)
4. [Log Storage](#log-storage)
5. [Log Management](#log-management)
6. [Best Practices](#best-practices)

## Quick Start

### 1. Setup Environment Variables

Create a `.env` file in your project root (copy from `.env.example`):

```bash
cp .env.example .env
```

### 2. Configure Logging

Edit `.env` to enable/disable logging:

```env
# Enable global device logging
DEVICE_LOGGING_ENABLED=true

# Enable logging for specific devices
PRINTER_LOGGING_ENABLED=true
SCANNER_LOGGING_ENABLED=true
SCALE_LOGGING_ENABLED=true
RFID_LOGGING_ENABLED=true
ACCESS_CONTROL_LOGGING_ENABLED=true
BADGE_PRINTER_LOGGING_ENABLED=true
PAYMENT_LOGGING_ENABLED=true
```

### 3. Initialize in Your App

```dart
import 'package:flutter_device_bridge/flutter_device_bridge.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize environment configuration
  await EnvConfig.initialize();

  runApp(MyApp());
}
```

### 4. Create Client with Logging

```dart
final client = DeviceBridgeClient(
  baseUrl: EnvConfig.baseUrl,
  enableDeviceLogging: EnvConfig.deviceLoggingEnabled,
  debug: EnvConfig.debug,
);
```

That's it! All device interactions will now be automatically logged.

## Configuration

### Environment Variables

All logging configuration is done via the `.env` file:

#### Global Settings

```env
# Master switch for all device logging
DEVICE_LOGGING_ENABLED=true

# Base path for all log files (relative to app documents directory)
DEVICE_LOGGING_BASE_PATH=device_logs
```

#### Per-Device Type Settings

Each device type can be configured independently:

```env
# Printer Logging
PRINTER_LOGGING_ENABLED=true
PRINTER_LOG_FORMAT=txt              # txt, csv, or json
PRINTER_LOG_PATH=device_logs/receipts

# Scanner Logging
SCANNER_LOGGING_ENABLED=true
SCANNER_LOG_FORMAT=csv
SCANNER_LOG_PATH=device_logs/scans

# Scale Logging
SCALE_LOGGING_ENABLED=true
SCALE_LOG_FORMAT=csv
SCALE_LOG_PATH=device_logs/scale

# RFID Logging
RFID_LOGGING_ENABLED=true
RFID_LOG_FORMAT=csv
RFID_LOG_PATH=device_logs/rfid

# Access Control Logging
ACCESS_CONTROL_LOGGING_ENABLED=true
ACCESS_CONTROL_LOG_FORMAT=csv
ACCESS_CONTROL_LOG_PATH=device_logs/access_control

# Badge Printer Logging
BADGE_PRINTER_LOGGING_ENABLED=true
BADGE_PRINTER_LOG_FORMAT=txt
BADGE_PRINTER_LOG_PATH=device_logs/badges

# Payment Terminal Logging
PAYMENT_LOGGING_ENABLED=true
PAYMENT_LOG_FORMAT=csv
PAYMENT_LOG_PATH=device_logs/payments
```

#### Log Retention Settings

```env
# Maximum file size before rotation (in MB)
LOG_MAX_FILE_SIZE_MB=10

# Maximum number of log files per device
LOG_MAX_FILES_PER_DEVICE=100

# Enable automatic cleanup of old logs
LOG_CLEANUP_ENABLED=true

# Number of days to retain logs
LOG_RETENTION_DAYS=30
```

### Programmatic Configuration

You can also control logging programmatically:

```dart
// Disable logging for a specific operation
final client = DeviceBridgeClient(
  baseUrl: 'http://localhost:8080',
  enableDeviceLogging: false, // Disable all logging
);

// Check if logging is enabled
final logger = await DeviceInteractionLogger.getInstance();
if (logger.isEnabled) {
  print('Device logging is enabled');
}

// Check if logging is enabled for a specific device type
if (logger.isEnabledForType(DeviceInteractionType.print)) {
  print('Printer logging is enabled');
}
```

## Log Formats

### TXT Format (Receipts, Badges)

Best for human-readable logs with full content. Used for printers and badge printers.

**Example (Receipt):**
```
================================================================================
Print Job: job-12345
Device: printer-01
Timestamp: 2025-11-16T10:30:45.123Z
Status: SUCCESS
Copies: 1
--------------------------------------------------------------------------------
My Store
123 Main Street
City, ST 12345
--------------------------------------------------------------------------------
Coffee               $3.50
Tax                  $0.30
--------------------------------------------------------------------------------
Total                $3.80

Thank you!
================================================================================
```

### CSV Format (Scans, Weights, Transactions)

Best for structured data that needs to be analyzed or imported. Used for scanners, scales, RFID, access control, and payments.

**Example (Scanner):**
```csv
ID,Timestamp,Device ID,Barcode Data,Symbology,Success,Error
550e8400-e29b-41d4-a716-446655440000,2025-11-16T10:30:45.123Z,scanner-01,"1234567890128",EAN13,true,
550e8400-e29b-41d4-a716-446655440001,2025-11-16T10:31:12.456Z,scanner-01,"QR-CODE-DATA",QR,true,
```

**Example (Scale):**
```csv
ID,Timestamp,Device ID,Weight,Unit,Success,Error
550e8400-e29b-41d4-a716-446655440000,2025-11-16T10:30:45.123Z,scale-01,1.250,kg,true,
550e8400-e29b-41d4-a716-446655440001,2025-11-16T10:31:15.789Z,scale-01,0.500,kg,true,
```

**Example (Access Control):**
```csv
ID,Timestamp,Device ID,Door ID,Credential,Access Result,Error
550e8400-e29b-41d4-a716-446655440000,2025-11-16T10:30:45.123Z,controller-01,main-entrance,"AB:CD:EF:01:23:45",GRANTED,
550e8400-e29b-41d4-a716-446655440001,2025-11-16T10:31:20.456Z,controller-01,main-entrance,"00:11:22:33:44:55",DENIED,Invalid credential
```

### JSON Format

Best for programmatic processing and integration with log management systems.

**Example:**
```json
{"id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2025-11-16T10:30:45.123Z","deviceId":"scanner-01","type":"scan","success":true,"barcodeData":"1234567890128","symbology":"EAN13"}
{"id":"550e8400-e29b-41d4-a716-446655440001","timestamp":"2025-11-16T10:31:12.456Z","deviceId":"scanner-01","type":"scan","success":true,"barcodeData":"QR-CODE-DATA","symbology":"QR"}
```

## Log Storage

### File Organization

Logs are stored in the application's documents directory under the configured base path:

```
<App Documents Directory>/
└── device_logs/
    ├── receipts/
    │   ├── printer-01_2025-11-16.txt
    │   └── printer-02_2025-11-16.txt
    ├── scans/
    │   ├── scanner-01_2025-11-16.csv
    │   └── scanner-02_2025-11-16.csv
    ├── scale/
    │   └── scale-01_2025-11-16.csv
    ├── rfid/
    │   └── rfid-reader-01_2025-11-16.csv
    ├── access_control/
    │   └── controller-01_2025-11-16.csv
    ├── badges/
    │   └── badge-printer-01_2025-11-16.txt
    └── payments/
        └── payment-terminal-01_2025-11-16.csv
```

### File Naming Convention

Log files are named using the pattern: `{deviceId}_{date}.{format}`

- `deviceId`: The ID of the device that generated the log
- `date`: The date in YYYY-MM-DD format
- `format`: The file extension (txt, csv, json)

### File Rotation

When a log file exceeds `LOG_MAX_FILE_SIZE_MB`, it is automatically rotated:

```
printer-01_2025-11-16.txt
printer-01_2025-11-16.txt.1700123456789
printer-01_2025-11-16.txt.1700123457890
```

The timestamp suffix is the Unix epoch milliseconds when the file was rotated.

## Log Management

### Accessing Logs

```dart
// Get logger instance
final logger = await DeviceInteractionLogger.getInstance();

// Get all log files for a specific device
final files = await logger.getLogFilesForDevice('printer-01');
for (final file in files) {
  print('Log file: ${file.path}');
  print('Size: ${await file.length()} bytes');
  print('Modified: ${await file.lastModified()}');
}

// Get log directory for a specific device type
final directory = await logger.getLogDirectoryForType(DeviceInteractionType.scan);
print('Scanner logs: ${directory.path}');
```

### Clearing Logs

```dart
final logger = await DeviceInteractionLogger.getInstance();

// Clear all logs
await logger.clearAllLogs();

// Clear logs for a specific device
await logger.clearLogsForDevice('printer-01');
```

### Exporting Logs

You can use the `share_plus` package to share log files:

```dart
import 'package:share_plus/share_plus.dart';

final logger = await DeviceInteractionLogger.getInstance();
final files = await logger.getLogFilesForDevice('scanner-01');

if (files.isNotEmpty) {
  await Share.shareXFiles(
    files.map((f) => XFile(f.path)).toList(),
    subject: 'Scanner Logs',
  );
}
```

## Best Practices

### 1. Enable Logging in Development

Always enable logging during development and testing:

```env
DEVICE_LOGGING_ENABLED=true
```

### 2. Choose Appropriate Formats

- Use **TXT** for receipts and documents where you need to see the full content
- Use **CSV** for data that needs analysis (scans, weights, transactions)
- Use **JSON** for integration with log management systems

### 3. Set Reasonable Retention Periods

Balance between audit requirements and storage:

```env
# For POS systems with audit requirements
LOG_RETENTION_DAYS=90
LOG_MAX_FILES_PER_DEVICE=500

# For general use
LOG_RETENTION_DAYS=30
LOG_MAX_FILES_PER_DEVICE=100
```

### 4. Monitor Log Storage

Implement a log monitoring screen in your app:

```dart
class LogMonitorScreen extends StatefulWidget {
  @override
  _LogMonitorScreenState createState() => _LogMonitorScreenState();
}

class _LogMonitorScreenState extends State<LogMonitorScreen> {
  int _totalFiles = 0;
  int _totalSize = 0;

  @override
  void initState() {
    super.initState();
    _calculateLogStats();
  }

  Future<void> _calculateLogStats() async {
    final logger = await DeviceInteractionLogger.getInstance();
    // Calculate total files and size across all devices
    // Update UI
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Log Statistics')),
      body: Column(
        children: [
          ListTile(
            title: Text('Total Log Files'),
            trailing: Text('$_totalFiles'),
          ),
          ListTile(
            title: Text('Total Storage Used'),
            trailing: Text('${(_totalSize / 1024 / 1024).toStringAsFixed(2)} MB'),
          ),
          ElevatedButton(
            onPressed: () async {
              final logger = await DeviceInteractionLogger.getInstance();
              await logger.clearAllLogs();
              _calculateLogStats();
            },
            child: Text('Clear All Logs'),
          ),
        ],
      ),
    );
  }
}
```

### 5. Privacy Considerations

- Disable logging in production if handling sensitive data:
  ```env
  DEVICE_LOGGING_ENABLED=false
  ```

- Or disable only sensitive device types:
  ```env
  PAYMENT_LOGGING_ENABLED=false
  RFID_LOGGING_ENABLED=false
  ```

- Implement secure log storage for regulated industries (PCI-DSS, HIPAA, etc.)

### 6. Performance Considerations

Logging is designed to be non-blocking:

- Log writes happen asynchronously
- Failed logging doesn't interrupt device operations
- Minimal performance impact on device interactions

### 7. Testing Logging

Test your logging configuration:

```dart
void testLogging() async {
  final logger = await DeviceInteractionLogger.getInstance();

  // Verify logging is enabled
  assert(logger.isEnabled, 'Logging should be enabled');

  // Verify device types are configured correctly
  assert(logger.isEnabledForType(DeviceInteractionType.print));
  assert(logger.isEnabledForType(DeviceInteractionType.scan));

  // Test a print operation
  final printer = client.printer('printer-01');
  await printer.printText('Test receipt');

  // Verify log was created
  final files = await logger.getLogFilesForDevice('printer-01');
  assert(files.isNotEmpty, 'Log file should be created');
}
```

## Troubleshooting

### Logs Not Being Created

1. **Check if logging is enabled**:
   ```dart
   final logger = await DeviceInteractionLogger.getInstance();
   print('Logging enabled: ${logger.isEnabled}');
   ```

2. **Verify .env file is loaded**:
   ```dart
   print('Env initialized: ${EnvConfig.isInitialized}');
   print('Logging enabled: ${EnvConfig.deviceLoggingEnabled}');
   ```

3. **Check permissions**: Ensure your app has permission to write to storage

4. **Verify device type is enabled**:
   ```dart
   print('Printer logging: ${EnvConfig.printerLoggingEnabled}');
   ```

### Log Files Too Large

1. Reduce `LOG_MAX_FILE_SIZE_MB`:
   ```env
   LOG_MAX_FILE_SIZE_MB=5
   ```

2. Enable cleanup:
   ```env
   LOG_CLEANUP_ENABLED=true
   LOG_RETENTION_DAYS=7
   ```

3. Reduce files per device:
   ```env
   LOG_MAX_FILES_PER_DEVICE=50
   ```

### Cannot Find Log Files

```dart
// Get the exact path where logs are stored
final logger = await DeviceInteractionLogger.getInstance();
final directory = await logger.getLogDirectoryForType(DeviceInteractionType.print);
print('Logs are at: ${directory.path}');

// On Android: /data/user/0/com.example.app/app_flutter/device_logs/receipts
// On iOS: /var/mobile/Containers/Data/Application/{UUID}/Documents/device_logs/receipts
```

## Examples

See the [example app](./example/basic_demo) for a complete demonstration of device interaction logging.

## API Reference

For detailed API documentation, see:
- [EnvConfig](./lib/src/env_config.dart)
- [DeviceInteractionLogger](./lib/src/logging/device_interaction_logger.dart)
- [DeviceLogEntry](./lib/src/logging/device_log_entry.dart)
