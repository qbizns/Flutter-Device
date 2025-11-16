import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:logging/logging.dart';

/// Environment configuration loader and validator
///
/// Loads configuration from .env file and provides type-safe access
/// to environment variables with validation.
class EnvConfig {
  static final Logger _logger = Logger('EnvConfig');
  static bool _isInitialized = false;

  /// Initialize environment configuration
  ///
  /// Must be called before using any configuration values.
  /// Validates that .env file exists and required variables are set.
  ///
  /// Example:
  /// ```dart
  /// void main() async {
  ///   await EnvConfig.initialize();
  ///   runApp(MyApp());
  /// }
  /// ```
  static Future<void> initialize() async {
    if (_isInitialized) {
      _logger.warning('.env already initialized');
      return;
    }

    try {
      await dotenv.load(fileName: '.env');
      _isInitialized = true;
      _logger.info('.env file loaded successfully');

      // Validate critical configuration
      _validateConfig();
    } catch (e) {
      _logger.severe('Failed to load .env file: $e');
      _logger.warning('Using default configuration values');
      _isInitialized = true; // Continue with defaults
    }
  }

  /// Validate that .env file exists and has required configuration
  static void _validateConfig() {
    final baseUrl = dotenv.env['DEVICE_BRIDGE_BASE_URL'];
    if (baseUrl == null || baseUrl.isEmpty) {
      _logger.warning(
        'DEVICE_BRIDGE_BASE_URL not set in .env, using default: http://localhost:8080',
      );
    }
  }

  /// Check if environment config is initialized
  static bool get isInitialized => _isInitialized;

  // ========================================
  // Device Bridge Configuration
  // ========================================

  static String get baseUrl =>
      dotenv.env['DEVICE_BRIDGE_BASE_URL'] ?? 'http://localhost:8080';

  static String? get apiKey => dotenv.env['DEVICE_BRIDGE_API_KEY'];

  static bool get useTls =>
      dotenv.env['DEVICE_BRIDGE_USE_TLS']?.toLowerCase() == 'true';

  static int get timeoutSeconds =>
      int.tryParse(dotenv.env['DEVICE_BRIDGE_TIMEOUT_SECONDS'] ?? '30') ?? 30;

  static bool get debug =>
      dotenv.env['DEVICE_BRIDGE_DEBUG']?.toLowerCase() == 'true';

  // ========================================
  // Device Logging Configuration
  // ========================================

  static bool get deviceLoggingEnabled =>
      dotenv.env['DEVICE_LOGGING_ENABLED']?.toLowerCase() != 'false';

  static String get loggingBasePath =>
      dotenv.env['DEVICE_LOGGING_BASE_PATH'] ?? 'device_logs';

  // Printer Logging
  static bool get printerLoggingEnabled =>
      dotenv.env['PRINTER_LOGGING_ENABLED']?.toLowerCase() != 'false';

  static String get printerLogFormat =>
      dotenv.env['PRINTER_LOG_FORMAT'] ?? 'txt';

  static String get printerLogPath =>
      dotenv.env['PRINTER_LOG_PATH'] ?? 'device_logs/receipts';

  // Scanner Logging
  static bool get scannerLoggingEnabled =>
      dotenv.env['SCANNER_LOGGING_ENABLED']?.toLowerCase() != 'false';

  static String get scannerLogFormat =>
      dotenv.env['SCANNER_LOG_FORMAT'] ?? 'csv';

  static String get scannerLogPath =>
      dotenv.env['SCANNER_LOG_PATH'] ?? 'device_logs/scans';

  // Scale Logging
  static bool get scaleLoggingEnabled =>
      dotenv.env['SCALE_LOGGING_ENABLED']?.toLowerCase() != 'false';

  static String get scaleLogFormat =>
      dotenv.env['SCALE_LOG_FORMAT'] ?? 'csv';

  static String get scaleLogPath =>
      dotenv.env['SCALE_LOG_PATH'] ?? 'device_logs/scale';

  // RFID Logging
  static bool get rfidLoggingEnabled =>
      dotenv.env['RFID_LOGGING_ENABLED']?.toLowerCase() != 'false';

  static String get rfidLogFormat =>
      dotenv.env['RFID_LOG_FORMAT'] ?? 'csv';

  static String get rfidLogPath =>
      dotenv.env['RFID_LOG_PATH'] ?? 'device_logs/rfid';

  // Access Control Logging
  static bool get accessControlLoggingEnabled =>
      dotenv.env['ACCESS_CONTROL_LOGGING_ENABLED']?.toLowerCase() != 'false';

  static String get accessControlLogFormat =>
      dotenv.env['ACCESS_CONTROL_LOG_FORMAT'] ?? 'csv';

  static String get accessControlLogPath =>
      dotenv.env['ACCESS_CONTROL_LOG_PATH'] ?? 'device_logs/access_control';

  // Badge Printer Logging
  static bool get badgePrinterLoggingEnabled =>
      dotenv.env['BADGE_PRINTER_LOGGING_ENABLED']?.toLowerCase() != 'false';

  static String get badgePrinterLogFormat =>
      dotenv.env['BADGE_PRINTER_LOG_FORMAT'] ?? 'txt';

  static String get badgePrinterLogPath =>
      dotenv.env['BADGE_PRINTER_LOG_PATH'] ?? 'device_logs/badges';

  // Payment Logging
  static bool get paymentLoggingEnabled =>
      dotenv.env['PAYMENT_LOGGING_ENABLED']?.toLowerCase() != 'false';

  static String get paymentLogFormat =>
      dotenv.env['PAYMENT_LOG_FORMAT'] ?? 'csv';

  static String get paymentLogPath =>
      dotenv.env['PAYMENT_LOG_PATH'] ?? 'device_logs/payments';

  // ========================================
  // Log Retention Settings
  // ========================================

  static int get logMaxFileSizeMB =>
      int.tryParse(dotenv.env['LOG_MAX_FILE_SIZE_MB'] ?? '10') ?? 10;

  static int get logMaxFilesPerDevice =>
      int.tryParse(dotenv.env['LOG_MAX_FILES_PER_DEVICE'] ?? '100') ?? 100;

  static bool get logCleanupEnabled =>
      dotenv.env['LOG_CLEANUP_ENABLED']?.toLowerCase() != 'false';

  static int get logRetentionDays =>
      int.tryParse(dotenv.env['LOG_RETENTION_DAYS'] ?? '30') ?? 30;

  /// Get all configuration as a map (for debugging)
  static Map<String, dynamic> toMap() {
    return {
      'deviceBridge': {
        'baseUrl': baseUrl,
        'useTls': useTls,
        'timeout': timeoutSeconds,
        'debug': debug,
      },
      'deviceLogging': {
        'enabled': deviceLoggingEnabled,
        'basePath': loggingBasePath,
        'printer': {
          'enabled': printerLoggingEnabled,
          'format': printerLogFormat,
          'path': printerLogPath,
        },
        'scanner': {
          'enabled': scannerLoggingEnabled,
          'format': scannerLogFormat,
          'path': scannerLogPath,
        },
        'scale': {
          'enabled': scaleLoggingEnabled,
          'format': scaleLogFormat,
          'path': scaleLogPath,
        },
        'rfid': {
          'enabled': rfidLoggingEnabled,
          'format': rfidLogFormat,
          'path': rfidLogPath,
        },
        'accessControl': {
          'enabled': accessControlLoggingEnabled,
          'format': accessControlLogFormat,
          'path': accessControlLogPath,
        },
        'badgePrinter': {
          'enabled': badgePrinterLoggingEnabled,
          'format': badgePrinterLogFormat,
          'path': badgePrinterLogPath,
        },
        'payment': {
          'enabled': paymentLoggingEnabled,
          'format': paymentLogFormat,
          'path': paymentLogPath,
        },
      },
      'logRetention': {
        'maxFileSizeMB': logMaxFileSizeMB,
        'maxFilesPerDevice': logMaxFilesPerDevice,
        'cleanupEnabled': logCleanupEnabled,
        'retentionDays': logRetentionDays,
      },
    };
  }
}
