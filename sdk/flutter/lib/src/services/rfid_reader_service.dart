import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

import '../client.dart';
import '../models/rfid_models.dart';
import '../models/common_models.dart';
import '../exceptions.dart';

/// Service for interacting with RFID/NFC readers
///
/// Provides methods for reading cards, authenticating, and monitoring reader status.
///
/// Example:
/// ```dart
/// final reader = client.rfidReader('rfid-reader-01');
///
/// // Read a card
/// final card = await reader.readCard(timeout: Duration(seconds: 30));
/// print('Card UID: ${card.uid}');
///
/// // Subscribe to card reads
/// await for (final event in reader.subscribeCardReads()) {
///   print('Card detected: ${event.card.uid}');
/// }
/// ```
class RFIDReaderService {
  final DeviceBridgeClient _client;
  final String deviceId;

  RFIDReaderService(this._client, this.deviceId);

  /// Read a card from the reader
  ///
  /// [timeout] - Maximum time to wait for a card (default: 30 seconds)
  /// [expectedTypes] - Optional list of expected card types
  /// [readFullData] - Whether to read full card data (default: false)
  /// [authenticate] - Whether to authenticate before reading (default: false)
  /// [authKeys] - Authentication keys (required if authenticate is true)
  Future<CardReadResponse> readCard({
    Duration timeout = const Duration(seconds: 30),
    List<CardType>? expectedTypes,
    bool readFullData = false,
    bool authenticate = false,
    List<String>? authKeys,
  }) async {
    final body = <String, dynamic>{
      'device_id': deviceId,
      'timeout_seconds': timeout.inSeconds,
      if (expectedTypes != null && expectedTypes.isNotEmpty)
        'expected_types': expectedTypes.map((e) => e.name.toUpperCase()).toList(),
      'options': {
        'read_full_data': readFullData,
        'authenticate': authenticate,
        if (authKeys != null) 'auth_keys': authKeys,
        'beep_on_read': true,
      },
    };

    try {
      final response = await _client.post(
        '/v1/devices/$deviceId/read',
        body: body,
      );

      return CardReadResponse.fromJson(response);
    } on DeviceBridgeException {
      rethrow;
    } catch (e) {
      throw DeviceBridgeException('Failed to read card: $e');
    }
  }

  /// Authenticate a card
  ///
  /// [cardUid] - Card UID to authenticate
  /// [method] - Authentication method (default: 'MIFARE_KEY')
  /// [credentials] - Authentication credentials
  Future<AuthenticationResponse> authenticateCard({
    required String cardUid,
    String method = 'AUTH_METHOD_MIFARE_KEY',
    required AuthenticationCredentials credentials,
  }) async {
    final body = <String, dynamic>{
      'device_id': deviceId,
      'card_uid': cardUid,
      'method': method,
      'credentials': credentials.toJson(),
    };

    try {
      final response = await _client.post(
        '/v1/devices/$deviceId/authenticate',
        body: body,
      );

      return AuthenticationResponse.fromJson(response);
    } on DeviceBridgeException {
      rethrow;
    } catch (e) {
      throw DeviceBridgeException('Failed to authenticate card: $e');
    }
  }

  /// Write data to a card
  ///
  /// [cardUid] - Card UID to write to
  /// [blocks] - Data blocks to write
  /// [credentials] - Authentication credentials
  /// [verify] - Whether to verify after writing
  Future<bool> writeCard({
    required String cardUid,
    required List<Map<String, dynamic>> blocks,
    required AuthenticationCredentials credentials,
    bool verify = true,
  }) async {
    final body = <String, dynamic>{
      'device_id': deviceId,
      'card_uid': cardUid,
      'blocks': blocks,
      'credentials': credentials.toJson(),
      'options': {
        'verify': verify,
        'beep_on_success': true,
      },
    };

    try {
      final response = await _client.post(
        '/v1/devices/$deviceId/write',
        body: body,
      );

      return response['success'] as bool? ?? false;
    } on DeviceBridgeException {
      rethrow;
    } catch (e) {
      throw DeviceBridgeException('Failed to write card: $e');
    }
  }

  /// Get reader status
  Future<ReaderStatus> getStatus() async {
    try {
      final response = await _client.get('/v1/devices/$deviceId/status');
      return ReaderStatus.fromJson(response);
    } on DeviceBridgeException {
      rethrow;
    } catch (e) {
      throw DeviceBridgeException('Failed to get reader status: $e');
    }
  }

  /// Set reader mode
  ///
  /// [mode] - Reader mode ('CONTINUOUS', 'SINGLE_READ', 'IDLE')
  /// [beepOnDetect] - Beep when card is detected
  /// [autoRead] - Automatically read when card is detected
  Future<bool> setMode({
    required String mode,
    bool beepOnDetect = true,
    bool autoRead = true,
  }) async {
    final body = <String, dynamic>{
      'device_id': deviceId,
      'mode': mode,
      'options': {
        'beep_on_detect': beepOnDetect,
        'auto_read': autoRead,
      },
    };

    try {
      final response = await _client.post(
        '/v1/devices/$deviceId/mode',
        body: body,
      );

      return response['success'] as bool? ?? false;
    } on DeviceBridgeException {
      rethrow;
    } catch (e) {
      throw DeviceBridgeException('Failed to set reader mode: $e');
    }
  }

  /// Subscribe to card read events
  ///
  /// Returns a stream of card read events.
  /// The stream will continue until closed or an error occurs.
  ///
  /// Example:
  /// ```dart
  /// await for (final event in reader.subscribeCardReads()) {
  ///   print('Card: ${event.card.uid}');
  /// }
  /// ```
  Stream<CardReadEvent> subscribeCardReads({
    List<CardType>? filterTypes,
  }) {
    final controller = StreamController<CardReadEvent>();

    // Build WebSocket URL
    final wsUrl = _client.config.baseUrl
        .replaceFirst('http://', 'ws://')
        .replaceFirst('https://', 'wss://');

    final uri = Uri.parse('$wsUrl/v1/devices/$deviceId/subscribe');

    try {
      final channel = WebSocketChannel.connect(uri);

      // Listen to WebSocket messages
      channel.stream.listen(
        (message) {
          try {
            final json = jsonDecode(message as String) as Map<String, dynamic>;
            final event = CardReadEvent.fromJson(json);

            // Apply filter if specified
            if (filterTypes == null || filterTypes.contains(event.card.type)) {
              controller.add(event);
            }
          } catch (e) {
            controller.addError(
              DeviceBridgeException('Failed to parse card read event: $e'),
            );
          }
        },
        onError: (error) {
          controller.addError(
            DeviceBridgeException('WebSocket error: $error'),
          );
        },
        onDone: () {
          controller.close();
        },
      );

      // Clean up on cancel
      controller.onCancel = () {
        channel.sink.close();
      };
    } catch (e) {
      controller.addError(
        ConnectionException('Failed to connect to WebSocket: $e'),
      );
    }

    return controller.stream;
  }

  /// Get supported card types for this reader
  Future<List<CardType>> getSupportedCardTypes() async {
    try {
      final response = await _client.get('/v1/devices/$deviceId/types');

      final types = (response['card_types'] as List<dynamic>?)
          ?.map((e) => CardType.fromString(e['type'] as String))
          .toList();

      return types ?? [];
    } on DeviceBridgeException {
      rethrow;
    } catch (e) {
      throw DeviceBridgeException('Failed to get supported card types: $e');
    }
  }

  @override
  String toString() => 'RFIDReaderService(deviceId: $deviceId)';
}
