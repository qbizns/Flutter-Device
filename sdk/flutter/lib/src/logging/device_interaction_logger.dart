import 'dart:io';
import 'dart:convert';
import 'package:path/path.dart' as path;
import 'package:path_provider/path_provider.dart';
import 'package:logging/logging.dart';
import 'package:uuid/uuid.dart';

import '../env_config.dart';
import 'device_log_entry.dart';

/// Service for logging device interactions to files
///
/// Logs are written to local storage in various formats:
/// - Printer: TXT format (full receipt content)
/// - Scanner: CSV format (timestamp, barcode data, symbology)
/// - Scale: CSV format (timestamp, weight, unit)
/// - RFID: CSV format (timestamp, card UID, type, operation)
/// - Access Control: CSV format (timestamp, door, credential, result)
/// - Payment: CSV format (timestamp, transaction details)
/// - Badge Printer: TXT format (badge details)
///
/// Each device type has its own subdirectory and log files.
/// Logging can be enabled/disabled globally or per device type via .env configuration.
class DeviceInteractionLogger {
  static final Logger _logger = Logger('DeviceInteractionLogger');
  static final Uuid _uuid = const Uuid();
  static DeviceInteractionLogger? _instance;

  final bool _enabled;
  final String _basePath;
  final Map<DeviceInteractionType, bool> _deviceTypeEnabled;
  final Map<DeviceInteractionType, String> _deviceTypePaths;
  final Map<DeviceInteractionType, String> _deviceTypeFormats;

  // File handles for CSV files (to manage headers)
  final Map<String, bool> _csvHeadersWritten = {};

  DeviceInteractionLogger._({
    required bool enabled,
    required String basePath,
    required Map<DeviceInteractionType, bool> deviceTypeEnabled,
    required Map<DeviceInteractionType, String> deviceTypePaths,
    required Map<DeviceInteractionType, String> deviceTypeFormats,
  })  : _enabled = enabled,
        _basePath = basePath,
        _deviceTypeEnabled = deviceTypeEnabled,
        _deviceTypePaths = deviceTypePaths,
        _deviceTypeFormats = deviceTypeFormats;

  /// Get singleton instance
  static Future<DeviceInteractionLogger> getInstance() async {
    if (_instance == null) {
      await _initialize();
    }
    return _instance!;
  }

  /// Initialize the logger from environment configuration
  static Future<void> _initialize() async {
    if (!EnvConfig.isInitialized) {
      await EnvConfig.initialize();
    }

    final deviceTypeEnabled = {
      DeviceInteractionType.print: EnvConfig.printerLoggingEnabled,
      DeviceInteractionType.scan: EnvConfig.scannerLoggingEnabled,
      DeviceInteractionType.weigh: EnvConfig.scaleLoggingEnabled,
      DeviceInteractionType.rfidRead: EnvConfig.rfidLoggingEnabled,
      DeviceInteractionType.rfidWrite: EnvConfig.rfidLoggingEnabled,
      DeviceInteractionType.accessGrant: EnvConfig.accessControlLoggingEnabled,
      DeviceInteractionType.accessDeny: EnvConfig.accessControlLoggingEnabled,
      DeviceInteractionType.badgePrint: EnvConfig.badgePrinterLoggingEnabled,
      DeviceInteractionType.paymentStart: EnvConfig.paymentLoggingEnabled,
      DeviceInteractionType.paymentComplete: EnvConfig.paymentLoggingEnabled,
      DeviceInteractionType.paymentRefund: EnvConfig.paymentLoggingEnabled,
    };

    final deviceTypePaths = {
      DeviceInteractionType.print: EnvConfig.printerLogPath,
      DeviceInteractionType.scan: EnvConfig.scannerLogPath,
      DeviceInteractionType.weigh: EnvConfig.scaleLogPath,
      DeviceInteractionType.rfidRead: EnvConfig.rfidLogPath,
      DeviceInteractionType.rfidWrite: EnvConfig.rfidLogPath,
      DeviceInteractionType.accessGrant: EnvConfig.accessControlLogPath,
      DeviceInteractionType.accessDeny: EnvConfig.accessControlLogPath,
      DeviceInteractionType.badgePrint: EnvConfig.badgePrinterLogPath,
      DeviceInteractionType.paymentStart: EnvConfig.paymentLogPath,
      DeviceInteractionType.paymentComplete: EnvConfig.paymentLogPath,
      DeviceInteractionType.paymentRefund: EnvConfig.paymentLogPath,
    };

    final deviceTypeFormats = {
      DeviceInteractionType.print: EnvConfig.printerLogFormat,
      DeviceInteractionType.scan: EnvConfig.scannerLogFormat,
      DeviceInteractionType.weigh: EnvConfig.scaleLogFormat,
      DeviceInteractionType.rfidRead: EnvConfig.rfidLogFormat,
      DeviceInteractionType.rfidWrite: EnvConfig.rfidLogFormat,
      DeviceInteractionType.accessGrant: EnvConfig.accessControlLogFormat,
      DeviceInteractionType.accessDeny: EnvConfig.accessControlLogFormat,
      DeviceInteractionType.badgePrint: EnvConfig.badgePrinterLogFormat,
      DeviceInteractionType.paymentStart: EnvConfig.paymentLogFormat,
      DeviceInteractionType.paymentComplete: EnvConfig.paymentLogFormat,
      DeviceInteractionType.paymentRefund: EnvConfig.paymentLogFormat,
    };

    _instance = DeviceInteractionLogger._(
      enabled: EnvConfig.deviceLoggingEnabled,
      basePath: EnvConfig.loggingBasePath,
      deviceTypeEnabled: deviceTypeEnabled,
      deviceTypePaths: deviceTypePaths,
      deviceTypeFormats: deviceTypeFormats,
    );

    _logger.info('Device interaction logger initialized (enabled: ${EnvConfig.deviceLoggingEnabled})');
  }

  /// Log a device interaction
  ///
  /// [entry] - The log entry to write
  ///
  /// Returns true if the log was written successfully
  Future<bool> log(DeviceLogEntry entry) async {
    if (!_enabled) {
      _logger.fine('Logging disabled, skipping entry: ${entry.id}');
      return false;
    }

    final typeEnabled = _deviceTypeEnabled[entry.type] ?? false;
    if (!typeEnabled) {
      _logger.fine('Logging disabled for ${entry.type}, skipping entry: ${entry.id}');
      return false;
    }

    try {
      final logPath = _deviceTypePaths[entry.type] ?? _basePath;
      final format = _deviceTypeFormats[entry.type] ?? 'txt';

      await _writeLog(entry, logPath, format);
      _logger.fine('Logged ${entry.type} interaction: ${entry.id}');
      return true;
    } catch (e, stackTrace) {
      _logger.severe('Failed to log ${entry.type} interaction: ${entry.id}', e, stackTrace);
      return false;
    }
  }

  /// Write log entry to file
  Future<void> _writeLog(
    DeviceLogEntry entry,
    String logPath,
    String format,
  ) async {
    final directory = await _getLogDirectory(logPath);
    final fileName = _getLogFileName(entry, format);
    final filePath = path.join(directory.path, fileName);

    final file = File(filePath);

    // Ensure directory exists
    if (!await directory.exists()) {
      await directory.create(recursive: true);
    }

    if (format == 'csv') {
      await _writeCsvLog(file, entry);
    } else if (format == 'json') {
      await _writeJsonLog(file, entry);
    } else {
      // Default to text format
      await _writeTextLog(file, entry);
    }

    // Check file size and rotate if needed
    await _checkAndRotateLog(file);
  }

  /// Write log entry in CSV format
  Future<void> _writeCsvLog(File file, DeviceLogEntry entry) async {
    final fileExists = await file.exists();
    final csvKey = '${file.path}_${entry.type.name}';

    // Write header if file doesn't exist or hasn't been written yet
    if (!fileExists || !_csvHeadersWritten.containsKey(csvKey)) {
      final header = DeviceLogEntry.getCsvHeader(entry.type);
      await file.writeAsString('$header\n', mode: FileMode.write);
      _csvHeadersWritten[csvKey] = true;
    }

    // Append CSV row
    final row = entry.toCsvRow();
    await file.writeAsString('$row\n', mode: FileMode.append);
  }

  /// Write log entry in text format
  Future<void> _writeTextLog(File file, DeviceLogEntry entry) async {
    final content = entry.toTextLine();
    await file.writeAsString(content, mode: FileMode.append);
  }

  /// Write log entry in JSON format
  Future<void> _writeJsonLog(File file, DeviceLogEntry entry) async {
    final jsonLine = jsonEncode(entry.toJson());
    await file.writeAsString('$jsonLine\n', mode: FileMode.append);
  }

  /// Get log directory
  Future<Directory> _getLogDirectory(String logPath) async {
    final documentsDir = await getApplicationDocumentsDirectory();
    final logDir = Directory(path.join(documentsDir.path, logPath));
    return logDir;
  }

  /// Get log file name
  String _getLogFileName(DeviceLogEntry entry, String format) {
    final date = DateTime.now();
    final dateStr = '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';

    if (format == 'csv') {
      return '${entry.deviceId}_$dateStr.$format';
    } else if (format == 'json') {
      return '${entry.deviceId}_$dateStr.$format';
    } else {
      // For text format, use device ID and date
      return '${entry.deviceId}_$dateStr.txt';
    }
  }

  /// Check file size and rotate if needed
  Future<void> _checkAndRotateLog(File file) async {
    try {
      final stat = await file.stat();
      final sizeMB = stat.size / (1024 * 1024);

      if (sizeMB >= EnvConfig.logMaxFileSizeMB) {
        _logger.info('Log file ${file.path} exceeds size limit, rotating...');
        await _rotateLog(file);
      }
    } catch (e) {
      _logger.warning('Failed to check log file size: $e');
    }
  }

  /// Rotate log file
  Future<void> _rotateLog(File file) async {
    try {
      final timestamp = DateTime.now().millisecondsSinceEpoch;
      final newPath = '${file.path}.$timestamp';
      await file.rename(newPath);

      // Clean up old logs
      if (EnvConfig.logCleanupEnabled) {
        await _cleanupOldLogs(file.parent);
      }
    } catch (e) {
      _logger.severe('Failed to rotate log file: $e');
    }
  }

  /// Clean up old log files
  Future<void> _cleanupOldLogs(Directory directory) async {
    try {
      final now = DateTime.now();
      final cutoffDate = now.subtract(Duration(days: EnvConfig.logRetentionDays));

      final files = await directory.list().toList();
      for (final entity in files) {
        if (entity is File) {
          final stat = await entity.stat();
          if (stat.modified.isBefore(cutoffDate)) {
            _logger.info('Deleting old log file: ${entity.path}');
            await entity.delete();
          }
        }
      }

      // Also limit number of files per device
      await _limitFilesPerDevice(directory);
    } catch (e) {
      _logger.warning('Failed to cleanup old logs: $e');
    }
  }

  /// Limit number of files per device
  Future<void> _limitFilesPerDevice(Directory directory) async {
    try {
      final files = await directory
          .list()
          .where((entity) => entity is File)
          .cast<File>()
          .toList();

      // Group files by device ID
      final Map<String, List<File>> filesByDevice = {};
      for (final file in files) {
        final fileName = path.basename(file.path);
        final deviceId = fileName.split('_').first;
        filesByDevice.putIfAbsent(deviceId, () => []).add(file);
      }

      // For each device, keep only the most recent N files
      for (final deviceFiles in filesByDevice.values) {
        if (deviceFiles.length > EnvConfig.logMaxFilesPerDevice) {
          // Sort by modification time, newest first
          deviceFiles.sort((a, b) {
            return b.lastModifiedSync().compareTo(a.lastModifiedSync());
          });

          // Delete excess files
          final filesToDelete = deviceFiles.skip(EnvConfig.logMaxFilesPerDevice);
          for (final file in filesToDelete) {
            _logger.info('Deleting excess log file: ${file.path}');
            await file.delete();
          }
        }
      }
    } catch (e) {
      _logger.warning('Failed to limit files per device: $e');
    }
  }

  /// Generate a unique log entry ID
  static String generateId() {
    return _uuid.v4();
  }

  /// Check if logging is enabled
  bool get isEnabled => _enabled;

  /// Check if logging is enabled for a specific device type
  bool isEnabledForType(DeviceInteractionType type) {
    return _enabled && (_deviceTypeEnabled[type] ?? false);
  }

  /// Get log directory for a device type
  Future<Directory> getLogDirectoryForType(DeviceInteractionType type) async {
    final logPath = _deviceTypePaths[type] ?? _basePath;
    return _getLogDirectory(logPath);
  }

  /// Get all log files for a device
  Future<List<File>> getLogFilesForDevice(String deviceId) async {
    try {
      final documentsDir = await getApplicationDocumentsDirectory();
      final logDir = Directory(path.join(documentsDir.path, _basePath));

      if (!await logDir.exists()) {
        return [];
      }

      final files = <File>[];
      await for (final entity in logDir.list(recursive: true)) {
        if (entity is File) {
          final fileName = path.basename(entity.path);
          if (fileName.startsWith(deviceId)) {
            files.add(entity);
          }
        }
      }

      return files;
    } catch (e) {
      _logger.warning('Failed to get log files for device $deviceId: $e');
      return [];
    }
  }

  /// Clear all logs
  Future<void> clearAllLogs() async {
    try {
      final documentsDir = await getApplicationDocumentsDirectory();
      final logDir = Directory(path.join(documentsDir.path, _basePath));

      if (await logDir.exists()) {
        await logDir.delete(recursive: true);
        _logger.info('All logs cleared');
      }
    } catch (e) {
      _logger.severe('Failed to clear logs: $e');
    }
  }

  /// Clear logs for a specific device
  Future<void> clearLogsForDevice(String deviceId) async {
    try {
      final files = await getLogFilesForDevice(deviceId);
      for (final file in files) {
        await file.delete();
      }
      _logger.info('Logs cleared for device: $deviceId');
    } catch (e) {
      _logger.severe('Failed to clear logs for device $deviceId: $e');
    }
  }
}
