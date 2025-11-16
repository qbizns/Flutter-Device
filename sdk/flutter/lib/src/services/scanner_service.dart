import 'dart:async';
import 'dart:convert';

import 'package:web_socket_channel/web_socket_channel.dart';

import '../client.dart';
import '../exceptions.dart';
import '../models/scanner_models.dart';
import '../logging/device_log_entry.dart';
import '../logging/device_interaction_logger.dart';

/// Service for interacting with barcode/QR scanners
///
/// Provides methods for scanning barcodes, QR codes, and other 2D codes
/// with support for continuous scanning mode.
///
/// Example:
/// ```dart
/// final scanner = client.scanner('scanner-01');
///
/// // One-time scan
/// final result = await scanner.scan(
///   options: ScanOptions(
///     timeout: Duration(seconds: 10),
///     expectedSymbologies: [BarcodeSymbology.qr, BarcodeSymbology.code128],
///   ),
/// );
/// print('Scanned: ${result.data}');
///
/// // Continuous scanning
/// await for (final event in scanner.subscribeScans()) {
///   print('Scanned: ${event.scan.data}');
/// }
/// ```
class ScannerService {
  final DeviceBridgeClient _client;
  final String deviceId;

  ScannerService(this._client, this.deviceId);

  /// Gets the current status of the scanner
  ///
  /// Returns comprehensive scanner status including mode, battery level,
  /// and capabilities
  ///
  /// Throws [DeviceNotFoundException] if scanner not found
  Future<ScannerStatus> getStatus() async {
    final response = await _client.get('/v1/devices/$deviceId/scanner/status');
    return ScannerStatus.fromJson(response);
  }

  /// Performs a single scan operation
  ///
  /// [options] - Optional scan options including timeout and symbologies
  ///
  /// Returns the scan result
  ///
  /// Throws [DeviceNotFoundException] if scanner not found
  /// Throws [DeviceBusyException] if scanner is already scanning
  /// Throws [DeviceTimeoutException] if no code is scanned within timeout
  ///
  /// Example:
  /// ```dart
  /// try {
  ///   final result = await scanner.scan(
  ///     options: ScanOptions(
  ///       timeout: Duration(seconds: 15),
  ///       expectedSymbologies: [BarcodeSymbology.qr],
  ///       enableBeep: true,
  ///     ),
  ///   );
  ///   print('Scanned QR: ${result.data}');
  /// } on DeviceTimeoutException {
  ///   print('No code scanned');
  /// }
  /// ```
  Future<ScanResult> scan({ScanOptions? options}) async {
    String? errorMessage;
    ScanResult? result;

    try {
      final body = options?.toJson() ?? {};

      final response = await _client.post(
        '/v1/devices/$deviceId/scanner/scan',
        body: body,
      );

      result = ScanResult.fromJson(response);

      // Log the scan interaction
      await _logScanInteraction(result);

      return result;
    } catch (e) {
      errorMessage = e.toString();

      // Log failed scan interaction
      if (result != null) {
        await _logScanInteraction(result, errorMessage: errorMessage);
      }

      rethrow;
    }
  }

  /// Log a scan interaction
  Future<void> _logScanInteraction(
    ScanResult result, {
    String? errorMessage,
  }) async {
    try {
      final logger = _client.deviceLogger;
      if (logger != null && logger.isEnabled) {
        final logEntry = ScannerLogEntry(
          id: DeviceInteractionLogger.generateId(),
          timestamp: result.timestamp ?? DateTime.now(),
          deviceId: deviceId,
          barcodeData: result.data,
          symbology: result.symbology?.name ?? 'unknown',
          success: errorMessage == null,
          errorMessage: errorMessage,
        );
        await logger.log(logEntry);
      }
    } catch (e) {
      // Silently fail logging to not disrupt main operations
    }
  }

  /// Starts continuous scanning mode
  ///
  /// [config] - Scanner configuration for continuous mode
  ///
  /// In continuous mode, the scanner will automatically scan codes
  /// without requiring a trigger press. Use [stopContinuousScan] to stop.
  ///
  /// Note: Use [subscribeScans] to receive scan events
  Future<void> startContinuousScan({ScannerConfig? config}) async {
    final body = config?.toJson() ??
        {
          'mode': 'CONTINUOUS',
        };

    await _client.post(
      '/v1/devices/$deviceId/scanner/start-continuous',
      body: body,
    );
  }

  /// Stops continuous scanning mode
  ///
  /// Returns the scanner to manual mode
  Future<void> stopContinuousScan() async {
    await _client.post('/v1/devices/$deviceId/scanner/stop-continuous');
  }

  /// Sets the scanner operating mode
  ///
  /// [mode] - The scanner mode to set
  /// [config] - Optional additional configuration
  ///
  /// Available modes:
  /// - manual: Scan on trigger press
  /// - continuous: Automatic continuous scanning
  /// - trigger: Scan while trigger is held
  Future<void> setMode(ScannerMode mode, {ScannerConfig? config}) async {
    final body = {
      'mode': mode.name.toUpperCase(),
      if (config != null) ...config.toJson(),
    };

    await _client.post(
      '/v1/devices/$deviceId/scanner/set-mode',
      body: body,
    );
  }

  /// Subscribes to real-time scan events
  ///
  /// [symbologyFilter] - Optional filter for specific symbologies
  ///
  /// Returns a stream of scan events
  ///
  /// The stream will continue until closed or an error occurs.
  /// This is useful for continuous scanning scenarios.
  ///
  /// Example:
  /// ```dart
  /// // Subscribe to all scans
  /// final subscription = scanner.subscribeScans().listen((event) {
  ///   print('Scanned: ${event.scan.data}');
  ///   print('Type: ${event.scan.symbology}');
  /// });
  ///
  /// // Later: cancel subscription
  /// await subscription.cancel();
  ///
  /// // Subscribe to specific symbologies
  /// final qrStream = scanner.subscribeScans(
  ///   symbologyFilter: [BarcodeSymbology.qr, BarcodeSymbology.dataMatrix],
  /// );
  /// ```
  Stream<ScanEvent> subscribeScans({
    List<BarcodeSymbology>? symbologyFilter,
  }) {
    final controller = StreamController<ScanEvent>();

    // Convert HTTP/HTTPS to WS/WSS
    final wsUrl = _client.config.baseUrl
        .replaceFirst('http://', 'ws://')
        .replaceFirst('https://', 'wss://');

    // Build query parameters
    final queryParams = <String, String>{};
    if (symbologyFilter != null && symbologyFilter.isNotEmpty) {
      queryParams['symbologies'] = symbologyFilter
          .map((e) => e.name.toUpperCase())
          .join(',');
    }

    final uri = Uri.parse('$wsUrl/v1/devices/$deviceId/scanner/subscribe')
        .replace(queryParameters: queryParams.isNotEmpty ? queryParams : null);

    try {
      final channel = WebSocketChannel.connect(uri);

      channel.stream.listen(
        (message) {
          try {
            final json = jsonDecode(message as String) as Map<String, dynamic>;
            final event = ScanEvent.fromJson(json);

            // Apply filter if specified
            if (symbologyFilter == null ||
                symbologyFilter.contains(event.scan.symbology)) {
              controller.add(event);
            }
          } catch (e) {
            controller.addError(
              DeviceBridgeException(
                'Failed to parse scan event: $e',
                cause: e,
              ),
            );
          }
        },
        onError: (error) {
          controller.addError(
            ConnectionException('WebSocket error: $error'),
          );
        },
        onDone: () {
          controller.close();
        },
        cancelOnError: false,
      );

      // Clean up WebSocket when stream is cancelled
      controller.onCancel = () {
        channel.sink.close();
      };
    } catch (e) {
      controller.addError(
        ConnectionException('Failed to connect to WebSocket: $e'),
      );
      controller.close();
    }

    return controller.stream;
  }

  /// Configures scanner settings
  ///
  /// [config] - Scanner configuration to apply
  ///
  /// Returns the updated scanner status
  Future<ScannerStatus> configure(ScannerConfig config) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/scanner/configure',
      body: config.toJson(),
    );

    return ScannerStatus.fromJson(response);
  }

  /// Gets scanner capabilities
  ///
  /// Returns detailed information about supported symbologies
  /// and scanner features
  Future<ScannerCapabilities> getCapabilities() async {
    final response =
        await _client.get('/v1/devices/$deviceId/scanner/capabilities');
    return ScannerCapabilities.fromJson(response);
  }

  /// Enables or disables specific barcode symbologies
  ///
  /// [symbologies] - List of symbologies to enable
  /// [enable] - Whether to enable (true) or disable (false)
  ///
  /// This allows you to restrict which code types the scanner will recognize
  Future<void> setSymbologies({
    required List<BarcodeSymbology> symbologies,
    required bool enable,
  }) async {
    final body = {
      'symbologies': symbologies.map((e) => e.name.toUpperCase()).toList(),
      'enable': enable,
    };

    await _client.post(
      '/v1/devices/$deviceId/scanner/set-symbologies',
      body: body,
    );
  }

  /// Triggers a scan programmatically
  ///
  /// This simulates a trigger button press. Useful for remote scanning.
  ///
  /// [options] - Optional scan options
  ///
  /// Returns the scan result
  Future<ScanResult> triggerScan({ScanOptions? options}) async {
    return scan(options: options);
  }

  /// Enables or disables the aiming laser
  ///
  /// [enabled] - Whether to enable the laser
  ///
  /// Note: Only works on scanners with aiming laser capability
  Future<void> setAimingLaser(bool enabled) async {
    await _client.post(
      '/v1/devices/$deviceId/scanner/set-aiming-laser',
      body: {'enabled': enabled},
    );
  }

  /// Enables or disables scan beep
  ///
  /// [enabled] - Whether to enable beep on successful scan
  /// [duration] - Optional beep duration in milliseconds
  Future<void> setBeep(bool enabled, {int? duration}) async {
    final body = {
      'enabled': enabled,
      if (duration != null) 'duration_ms': duration,
    };

    await _client.post(
      '/v1/devices/$deviceId/scanner/set-beep',
      body: body,
    );
  }

  /// Enables or disables scan vibration
  ///
  /// [enabled] - Whether to enable vibration on successful scan
  /// [duration] - Optional vibration duration in milliseconds
  Future<void> setVibration(bool enabled, {int? duration}) async {
    final body = {
      'enabled': enabled,
      if (duration != null) 'duration_ms': duration,
    };

    await _client.post(
      '/v1/devices/$deviceId/scanner/set-vibration',
      body: body,
    );
  }

  /// Resets the scanner to default settings
  ///
  /// This clears all configuration and returns scanner to factory defaults
  Future<void> reset() async {
    await _client.post('/v1/devices/$deviceId/scanner/reset');
  }

  /// Gets recent scan history
  ///
  /// [limit] - Maximum number of scans to return (default: 100)
  /// [symbologyFilter] - Optional filter for specific symbologies
  ///
  /// Returns a list of recent scan results
  Future<List<ScanResult>> getRecentScans({
    int limit = 100,
    List<BarcodeSymbology>? symbologyFilter,
  }) async {
    final queryParams = <String, dynamic>{
      'limit': limit.toString(),
      if (symbologyFilter != null && symbologyFilter.isNotEmpty)
        'symbologies':
            symbologyFilter.map((e) => e.name.toUpperCase()).join(','),
    };

    final response = await _client.get(
      '/v1/devices/$deviceId/scanner/recent-scans',
      queryParameters: queryParams,
    );

    final scans = response['scans'] as List<dynamic>;
    return scans
        .map((e) => ScanResult.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// Clears scan history
  Future<void> clearHistory() async {
    await _client.post('/v1/devices/$deviceId/scanner/clear-history');
  }
}
