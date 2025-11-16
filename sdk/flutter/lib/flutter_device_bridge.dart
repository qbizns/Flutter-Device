/// Official Flutter SDK for Device Bridge
///
/// Universal hardware integration for POS, RFID, access control, badge printers, and more.
///
/// This SDK provides a type-safe, idiomatic Dart/Flutter interface to Device Bridge,
/// a production-ready Go microservice for universal hardware integration.
///
/// Features:
/// - Type-safe API for all hardware types
/// - Support for REST, gRPC, and WebSocket protocols
/// - Real-time event streaming
/// - Flutter widgets for common hardware operations
/// - Comprehensive error handling
/// - Full async/await support
///
/// Example:
/// ```dart
/// // Initialize client
/// final bridge = DeviceBridgeClient(baseUrl: 'http://localhost:8080');
///
/// // Print receipt
/// final printer = bridge.printer('printer-01');
/// await printer.print(content: 'Hello World\n', cut: true);
///
/// // Read RFID card
/// final reader = bridge.rfidReader('rfid-reader-01');
/// final card = await reader.readCard(timeout: Duration(seconds: 30));
/// print('Card UID: ${card.uid}');
///
/// // Grant access
/// final accessControl = bridge.accessControl('controller-01');
/// final result = await accessControl.grantAccess(
///   doorId: 'main-entrance',
///   cardUid: card.uid,
/// );
/// ```
library flutter_device_bridge;

// Core client
export 'src/client.dart';
export 'src/config.dart';
export 'src/env_config.dart';

// Services
export 'src/services/printer_service.dart';
export 'src/services/scanner_service.dart';
export 'src/services/rfid_reader_service.dart';
export 'src/services/access_control_service.dart';
export 'src/services/badge_printer_service.dart';
export 'src/services/payment_terminal_service.dart';

// Models
export 'src/models/printer_models.dart';
export 'src/models/scanner_models.dart';
export 'src/models/rfid_models.dart';
export 'src/models/access_control_models.dart';
export 'src/models/badge_printer_models.dart';
export 'src/models/payment_models.dart';
export 'src/models/common_models.dart';

// Widgets
export 'src/widgets/scanner_listener.dart';
export 'src/widgets/device_status_widget.dart';
export 'src/widgets/rfid_reader_widget.dart';
export 'src/widgets/access_control_widget.dart';

// Exceptions
export 'src/exceptions.dart';

// Device Interaction Logging
export 'src/logging/device_interaction_logger.dart';
export 'src/logging/device_log_entry.dart';

// Utilities
export 'src/utils/logger.dart';
