import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:logging/logging.dart';

import 'config.dart';
import 'exceptions.dart';
import 'services/printer_service.dart';
import 'services/scanner_service.dart';
import 'services/rfid_reader_service.dart';
import 'services/access_control_service.dart';
import 'services/badge_printer_service.dart';
import 'services/payment_terminal_service.dart';

/// Main client for interacting with Device Bridge
///
/// This is the entry point for all hardware operations.
/// Create an instance and use it to access different hardware services.
///
/// Example:
/// ```dart
/// final client = DeviceBridgeClient(
///   baseUrl: 'http://localhost:8080',
/// );
///
/// // Access printer
/// final printer = client.printer('printer-01');
/// await printer.print(content: 'Hello World\n');
///
/// // Access RFID reader
/// final reader = client.rfidReader('rfid-reader-01');
/// final card = await reader.readCard();
/// ```
class DeviceBridgeClient {
  /// Configuration for this client
  final DeviceBridgeConfig config;

  /// HTTP client for making requests
  final http.Client _httpClient;

  /// Logger
  final Logger _logger;

  /// Create a new Device Bridge client
  ///
  /// [config] - Configuration for the client
  /// [httpClient] - Optional custom HTTP client
  DeviceBridgeClient({
    required String baseUrl,
    String? apiKey,
    bool useTls = false,
    Duration timeout = const Duration(seconds: 30),
    bool debug = false,
    Map<String, String>? headers,
    http.Client? httpClient,
  })  : config = DeviceBridgeConfig(
          baseUrl: baseUrl,
          apiKey: apiKey,
          useTls: useTls,
          timeout: timeout,
          debug: debug,
          headers: headers,
        ),
        _httpClient = httpClient ?? http.Client(),
        _logger = Logger('DeviceBridgeClient') {
    if (debug) {
      Logger.root.level = Level.ALL;
      Logger.root.onRecord.listen((record) {
        print('${record.level.name}: ${record.time}: ${record.message}');
      });
    }
  }

  /// Create a client with custom configuration
  DeviceBridgeClient.withConfig(
    this.config, {
    http.Client? httpClient,
  })  : _httpClient = httpClient ?? http.Client(),
        _logger = Logger('DeviceBridgeClient') {
    if (config.debug) {
      Logger.root.level = Level.ALL;
      Logger.root.onRecord.listen((record) {
        print('${record.level.name}: ${record.time}: ${record.message}');
      });
    }
  }

  // ========================================
  // Service Accessors
  // ========================================

  /// Get a printer service for a specific device
  ///
  /// [deviceId] - The ID of the printer device
  PrinterService printer(String deviceId) {
    return PrinterService(this, deviceId);
  }

  /// Get a scanner service for a specific device
  ///
  /// [deviceId] - The ID of the scanner device
  ScannerService scanner(String deviceId) {
    return ScannerService(this, deviceId);
  }

  /// Get an RFID reader service for a specific device
  ///
  /// [deviceId] - The ID of the RFID reader device
  RFIDReaderService rfidReader(String deviceId) {
    return RFIDReaderService(this, deviceId);
  }

  /// Get an access control service for a specific controller
  ///
  /// [deviceId] - The ID of the access controller
  AccessControlService accessControl(String deviceId) {
    return AccessControlService(this, deviceId);
  }

  /// Get a badge printer service for a specific device
  ///
  /// [deviceId] - The ID of the badge printer device
  BadgePrinterService badgePrinter(String deviceId) {
    return BadgePrinterService(this, deviceId);
  }

  /// Get a payment terminal service for a specific device
  ///
  /// [deviceId] - The ID of the payment terminal device
  PaymentTerminalService paymentTerminal(String deviceId) {
    return PaymentTerminalService(this, deviceId);
  }

  // ========================================
  // HTTP Request Methods
  // ========================================

  /// Make a GET request
  ///
  /// [path] - API endpoint path
  /// [queryParameters] - Optional query parameters
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? queryParameters,
  }) async {
    final uri = _buildUri(path, queryParameters);
    _logger.fine('GET $uri');

    try {
      final response = await _httpClient
          .get(uri, headers: config.getHeaders())
          .timeout(config.timeout);

      return _handleResponse(response);
    } catch (e) {
      throw DeviceBridgeException('GET request failed: $e');
    }
  }

  /// Make a POST request
  ///
  /// [path] - API endpoint path
  /// [body] - Request body
  /// [queryParameters] - Optional query parameters
  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? body,
    Map<String, dynamic>? queryParameters,
  }) async {
    final uri = _buildUri(path, queryParameters);
    _logger.fine('POST $uri');

    try {
      final response = await _httpClient
          .post(
            uri,
            headers: config.getHeaders(),
            body: body != null ? jsonEncode(body) : null,
          )
          .timeout(config.timeout);

      return _handleResponse(response);
    } catch (e) {
      throw DeviceBridgeException('POST request failed: $e');
    }
  }

  /// Make a PUT request
  ///
  /// [path] - API endpoint path
  /// [body] - Request body
  Future<Map<String, dynamic>> put(
    String path, {
    required Map<String, dynamic> body,
  }) async {
    final uri = _buildUri(path);
    _logger.fine('PUT $uri');

    try {
      final response = await _httpClient
          .put(
            uri,
            headers: config.getHeaders(),
            body: jsonEncode(body),
          )
          .timeout(config.timeout);

      return _handleResponse(response);
    } catch (e) {
      throw DeviceBridgeException('PUT request failed: $e');
    }
  }

  /// Make a DELETE request
  ///
  /// [path] - API endpoint path
  Future<Map<String, dynamic>> delete(String path) async {
    final uri = _buildUri(path);
    _logger.fine('DELETE $uri');

    try {
      final response = await _httpClient
          .delete(uri, headers: config.getHeaders())
          .timeout(config.timeout);

      return _handleResponse(response);
    } catch (e) {
      throw DeviceBridgeException('DELETE request failed: $e');
    }
  }

  // ========================================
  // Helper Methods
  // ========================================

  /// Build URI from path and query parameters
  Uri _buildUri(String path, [Map<String, dynamic>? queryParameters]) {
    final url = config.getUrl(path);
    final uri = Uri.parse(url);

    if (queryParameters != null && queryParameters.isNotEmpty) {
      return uri.replace(
        queryParameters:
            queryParameters.map((key, value) => MapEntry(key, value.toString())),
      );
    }

    return uri;
  }

  /// Handle HTTP response
  Map<String, dynamic> _handleResponse(http.Response response) {
    _logger.fine('Response status: ${response.statusCode}');

    if (response.statusCode >= 200 && response.statusCode < 300) {
      if (response.body.isEmpty) {
        return {};
      }

      try {
        return jsonDecode(response.body) as Map<String, dynamic>;
      } catch (e) {
        throw DeviceBridgeException('Failed to parse response: $e');
      }
    } else {
      // Handle error response
      String errorMessage = 'Request failed with status ${response.statusCode}';

      try {
        final errorData = jsonDecode(response.body) as Map<String, dynamic>;
        errorMessage = errorData['error']?.toString() ??
            errorData['message']?.toString() ??
            errorMessage;
      } catch (_) {
        // If we can't parse error, use the body as message
        if (response.body.isNotEmpty) {
          errorMessage = response.body;
        }
      }

      throw DeviceBridgeException(
        errorMessage,
        statusCode: response.statusCode,
      );
    }
  }

  /// Close the client and clean up resources
  void close() {
    _httpClient.close();
  }
}
