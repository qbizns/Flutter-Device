import 'dart:async';
import 'dart:convert';

import 'package:web_socket_channel/web_socket_channel.dart';

import '../client.dart';
import '../exceptions.dart';
import '../models/access_control_models.dart';

/// Service for interacting with access control devices
///
/// Provides methods for door control, access validation, event monitoring,
/// and lockdown management.
///
/// Example:
/// ```dart
/// final controller = client.accessControl('controller-01');
///
/// // Grant access with card credential
/// final result = await controller.grantAccess(
///   doorId: 'main-entrance',
///   credential: Credential.card(cardUid: '01020304'),
/// );
///
/// if (result.granted) {
///   print('Access granted: ${result.reason}');
/// }
///
/// // Subscribe to access events
/// await for (final event in controller.subscribeAccessEvents()) {
///   print('Event: ${event.type} at ${event.doorId}');
/// }
/// ```
class AccessControlService {
  final DeviceBridgeClient _client;
  final String deviceId;

  AccessControlService(this._client, this.deviceId);

  /// Unlocks a specific door
  ///
  /// [doorId] - The door identifier to unlock
  /// [duration] - Optional duration to keep door unlocked (null = manual lock required)
  /// [reason] - Optional reason for unlocking
  ///
  /// Returns the updated door status after unlocking
  ///
  /// Throws [DeviceNotFoundException] if device or door not found
  /// Throws [DeviceBusyException] if device is busy
  /// Throws [DeviceErrorException] if unlock fails
  Future<DoorStatus> unlockDoor({
    required String doorId,
    Duration? duration,
    String? reason,
  }) async {
    final body = <String, dynamic>{
      'door_id': doorId,
      if (duration != null) 'duration_seconds': duration.inSeconds,
      if (reason != null) 'reason': reason,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/access-control/doors/$doorId/unlock',
      body: body,
    );

    return DoorStatus.fromJson(response['door_status'] as Map<String, dynamic>);
  }

  /// Locks a specific door
  ///
  /// [doorId] - The door identifier to lock
  /// [reason] - Optional reason for locking
  ///
  /// Returns the updated door status after locking
  ///
  /// Throws [DeviceNotFoundException] if device or door not found
  /// Throws [DeviceBusyException] if device is busy
  /// Throws [DeviceErrorException] if lock fails
  Future<DoorStatus> lockDoor({
    required String doorId,
    String? reason,
  }) async {
    final body = <String, dynamic>{
      'door_id': doorId,
      if (reason != null) 'reason': reason,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/access-control/doors/$doorId/lock',
      body: body,
    );

    return DoorStatus.fromJson(response['door_status'] as Map<String, dynamic>);
  }

  /// Gets the current status of a door
  ///
  /// [doorId] - The door identifier
  ///
  /// Returns the current door status including lock state, position, and mode
  ///
  /// Throws [DeviceNotFoundException] if device or door not found
  Future<DoorStatus> getDoorStatus(String doorId) async {
    final response = await _client.get(
      '/v1/devices/$deviceId/access-control/doors/$doorId/status',
    );

    return DoorStatus.fromJson(response);
  }

  /// Checks if a credential would be granted access without actually granting it
  ///
  /// [doorId] - The door identifier to check access for
  /// [credential] - The credential to validate
  /// [context] - Optional context information (time, location, etc.)
  ///
  /// Returns an access check response with the decision and reason
  ///
  /// This is useful for pre-validation before granting access
  Future<AccessCheckResponse> checkAccess({
    required String doorId,
    required Credential credential,
    Map<String, dynamic>? context,
  }) async {
    final body = <String, dynamic>{
      'door_id': doorId,
      'credential': credential.toJson(),
      if (context != null) 'context': context,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/access-control/check-access',
      body: body,
    );

    return AccessCheckResponse.fromJson(response);
  }

  /// Grants access to a door with the provided credential
  ///
  /// [doorId] - The door identifier
  /// [credential] - The credential for access (card, PIN, or card+PIN)
  /// [unlockDoor] - Whether to physically unlock the door (default: true)
  /// [unlockDuration] - How long to keep door unlocked (default: 5 seconds)
  /// [context] - Optional context information
  ///
  /// Returns the grant access response with decision, user info, and event ID
  ///
  /// Example:
  /// ```dart
  /// // Card-only access
  /// final result = await controller.grantAccess(
  ///   doorId: 'door-1',
  ///   credential: Credential.card(cardUid: '01020304'),
  /// );
  ///
  /// // Card + PIN access
  /// final result = await controller.grantAccess(
  ///   doorId: 'door-1',
  ///   credential: Credential.cardAndPin(
  ///     cardUid: '01020304',
  ///     pin: '1234',
  ///   ),
  /// );
  /// ```
  Future<GrantAccessResponse> grantAccess({
    required String doorId,
    required Credential credential,
    bool unlockDoor = true,
    Duration unlockDuration = const Duration(seconds: 5),
    Map<String, dynamic>? context,
  }) async {
    final body = <String, dynamic>{
      'door_id': doorId,
      'credential': credential.toJson(),
      'unlock_door': unlockDoor,
      'unlock_duration_seconds': unlockDuration.inSeconds,
      if (context != null) 'context': context,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/access-control/grant-access',
      body: body,
    );

    return GrantAccessResponse.fromJson(response);
  }

  /// Sets the operating mode for a door
  ///
  /// [doorId] - The door identifier
  /// [mode] - The door mode to set
  /// [reason] - Optional reason for mode change
  ///
  /// Returns the updated door status
  ///
  /// Modes:
  /// - normal: Standard card access mode
  /// - unlocked: Door always unlocked
  /// - locked: Door always locked
  /// - officeMode: Unlocked during business hours
  /// - lockdown: Emergency lockdown mode
  /// - facilityCode: Require facility code validation
  /// - cardOnly: Card access without PIN
  Future<DoorStatus> setDoorMode({
    required String doorId,
    required DoorMode mode,
    String? reason,
  }) async {
    final body = <String, dynamic>{
      'door_id': doorId,
      'mode': 'DOOR_MODE_${mode.name.toUpperCase()}',
      if (reason != null) 'reason': reason,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/access-control/doors/$doorId/set-mode',
      body: body,
    );

    return DoorStatus.fromJson(response['door_status'] as Map<String, dynamic>);
  }

  /// Activates lockdown mode on the controller
  ///
  /// [level] - The lockdown level (partial, full, evacuation)
  /// [doorIds] - Optional specific doors to lock down (null = all doors)
  /// [reason] - Required reason for lockdown
  /// [notifyAuthorities] - Whether to notify authorities (default: false)
  ///
  /// Returns the updated controller status
  ///
  /// Example:
  /// ```dart
  /// // Full facility lockdown
  /// await controller.activateLockdown(
  ///   level: LockdownLevel.full,
  ///   reason: 'Emergency situation',
  ///   notifyAuthorities: true,
  /// );
  ///
  /// // Partial lockdown of specific areas
  /// await controller.activateLockdown(
  ///   level: LockdownLevel.partial,
  ///   doorIds: ['east-wing-1', 'east-wing-2'],
  ///   reason: 'Maintenance in east wing',
  /// );
  /// ```
  Future<ControllerStatus> activateLockdown({
    required LockdownLevel level,
    List<String>? doorIds,
    required String reason,
    bool notifyAuthorities = false,
  }) async {
    final body = <String, dynamic>{
      'level': 'LOCKDOWN_LEVEL_${level.name.toUpperCase()}',
      if (doorIds != null) 'door_ids': doorIds,
      'reason': reason,
      'notify_authorities': notifyAuthorities,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/access-control/lockdown/activate',
      body: body,
    );

    return ControllerStatus.fromJson(response);
  }

  /// Deactivates lockdown mode on the controller
  ///
  /// [reason] - Required reason for deactivating lockdown
  /// [restoreNormalMode] - Whether to restore doors to normal mode (default: true)
  ///
  /// Returns the updated controller status
  Future<ControllerStatus> deactivateLockdown({
    required String reason,
    bool restoreNormalMode = true,
  }) async {
    final body = <String, dynamic>{
      'reason': reason,
      'restore_normal_mode': restoreNormalMode,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/access-control/lockdown/deactivate',
      body: body,
    );

    return ControllerStatus.fromJson(response);
  }

  /// Gets the current status of the access controller
  ///
  /// Returns comprehensive controller status including all doors,
  /// lockdown state, and statistics
  Future<ControllerStatus> getControllerStatus() async {
    final response = await _client.get(
      '/v1/devices/$deviceId/access-control/status',
    );

    return ControllerStatus.fromJson(response);
  }

  /// Gets all doors managed by this controller
  ///
  /// Returns a list of door status objects
  Future<List<DoorStatus>> listDoors() async {
    final response = await _client.get(
      '/v1/devices/$deviceId/access-control/doors',
    );

    final doors = response['doors'] as List<dynamic>;
    return doors
        .map((e) => DoorStatus.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// Subscribes to real-time access control events
  ///
  /// [doorIds] - Optional filter for specific doors
  /// [eventTypes] - Optional filter for specific event types
  ///
  /// Returns a stream of access events including:
  /// - Access granted/denied events
  /// - Door open/close events
  /// - Forced door events
  /// - Lockdown events
  /// - Tamper alerts
  ///
  /// The stream will continue until closed or an error occurs.
  ///
  /// Example:
  /// ```dart
  /// // Subscribe to all events
  /// await for (final event in controller.subscribeAccessEvents()) {
  ///   if (event.decision == AccessDecision.denied) {
  ///     print('Access denied: ${event.reason}');
  ///   }
  /// }
  ///
  /// // Subscribe to specific doors
  /// final stream = controller.subscribeAccessEvents(
  ///   doorIds: ['main-entrance', 'back-door'],
  /// );
  /// ```
  Stream<AccessEvent> subscribeAccessEvents({
    List<String>? doorIds,
    List<String>? eventTypes,
  }) {
    final controller = StreamController<AccessEvent>();

    // Convert HTTP/HTTPS to WS/WSS
    final wsUrl = _client.config.baseUrl
        .replaceFirst('http://', 'ws://')
        .replaceFirst('https://', 'wss://');

    // Build query parameters
    final queryParams = <String, String>{};
    if (doorIds != null && doorIds.isNotEmpty) {
      queryParams['door_ids'] = doorIds.join(',');
    }
    if (eventTypes != null && eventTypes.isNotEmpty) {
      queryParams['event_types'] = eventTypes.join(',');
    }

    final uri = Uri.parse('$wsUrl/v1/devices/$deviceId/access-control/subscribe')
        .replace(queryParameters: queryParams.isNotEmpty ? queryParams : null);

    try {
      final channel = WebSocketChannel.connect(uri);

      channel.stream.listen(
        (message) {
          try {
            final json = jsonDecode(message as String) as Map<String, dynamic>;
            final event = AccessEvent.fromJson(json);
            controller.add(event);
          } catch (e) {
            controller.addError(
              DeviceBridgeException(
                'Failed to parse access event: $e',
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

  /// Gets credential information by user ID
  ///
  /// [userId] - The user identifier
  ///
  /// Returns credential information including access level, valid dates,
  /// and allowed doors
  ///
  /// Throws [DeviceNotFoundException] if user not found
  Future<CredentialInfo> getCredentialInfo(String userId) async {
    final response = await _client.get(
      '/v1/devices/$deviceId/access-control/credentials/$userId',
    );

    return CredentialInfo.fromJson(response);
  }

  /// Gets recent access events with optional filtering
  ///
  /// [doorIds] - Optional filter for specific doors
  /// [userIds] - Optional filter for specific users
  /// [startTime] - Optional start time for event range
  /// [endTime] - Optional end time for event range
  /// [limit] - Maximum number of events to return (default: 100)
  ///
  /// Returns a list of access events
  Future<List<AccessEvent>> getRecentEvents({
    List<String>? doorIds,
    List<String>? userIds,
    DateTime? startTime,
    DateTime? endTime,
    int limit = 100,
  }) async {
    final queryParams = <String, dynamic>{
      'limit': limit.toString(),
      if (doorIds != null && doorIds.isNotEmpty) 'door_ids': doorIds.join(','),
      if (userIds != null && userIds.isNotEmpty) 'user_ids': userIds.join(','),
      if (startTime != null) 'start_time': startTime.toIso8601String(),
      if (endTime != null) 'end_time': endTime.toIso8601String(),
    };

    final response = await _client.get(
      '/v1/devices/$deviceId/access-control/events',
      queryParameters: queryParams,
    );

    final events = response['events'] as List<dynamic>;
    return events
        .map((e) => AccessEvent.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// Performs an emergency door override
  ///
  /// This bypasses normal access control rules for emergency situations.
  /// Requires special authorization.
  ///
  /// [doorId] - The door to override
  /// [action] - The action to perform ('unlock' or 'lock')
  /// [reason] - Required reason for override
  /// [authCode] - Required authorization code
  ///
  /// Returns the updated door status
  ///
  /// Throws [AuthorizationException] if auth code is invalid
  Future<DoorStatus> emergencyOverride({
    required String doorId,
    required String action,
    required String reason,
    required String authCode,
  }) async {
    final body = <String, dynamic>{
      'door_id': doorId,
      'action': action,
      'reason': reason,
      'auth_code': authCode,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/access-control/emergency-override',
      body: body,
    );

    return DoorStatus.fromJson(response['door_status'] as Map<String, dynamic>);
  }
}
