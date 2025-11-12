import 'package:meta/meta.dart';
import 'common_models.dart';

/// Lock status enumeration
enum LockStatus {
  unknown,
  locked,
  unlocked,
  unlocking,
  locking,
  error;

  static LockStatus fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('LOCK_STATUS_', '');
    return LockStatus.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => LockStatus.unknown,
    );
  }
}

/// Door position enumeration
enum DoorPosition {
  unknown,
  closed,
  open,
  ajar,
  forced;

  static DoorPosition fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('DOOR_POSITION_', '');
    return DoorPosition.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => DoorPosition.unknown,
    );
  }
}

/// Door mode enumeration
enum DoorMode {
  unknown,
  normal,
  unlocked,
  locked,
  officeMode,
  lockdown,
  facilityCode,
  cardOnly;

  static DoorMode fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('DOOR_MODE_', '');
    final map = {
      'NORMAL': normal,
      'UNLOCKED': unlocked,
      'LOCKED': locked,
      'OFFICE_MODE': officeMode,
      'LOCKDOWN': lockdown,
      'FACILITY_CODE': facilityCode,
      'CARD_ONLY': cardOnly,
    };
    return map[normalized] ?? DoorMode.unknown;
  }
}

/// Access decision enumeration
enum AccessDecision {
  unknown,
  granted,
  denied,
  pending;

  static AccessDecision fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('ACCESS_DECISION_', '');
    return AccessDecision.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => AccessDecision.unknown,
    );
  }
}

/// Lockdown level enumeration
enum LockdownLevel {
  unknown,
  partial,
  full,
  evacuation;

  static LockdownLevel fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('LOCKDOWN_LEVEL_', '');
    return LockdownLevel.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => LockdownLevel.unknown,
    );
  }
}

/// Credential for access control
@immutable
class Credential {
  final String type;
  final String? cardUid;
  final String? pin;
  final String? qrCode;
  final int? facilityCode;
  final int? cardNumber;

  const Credential({
    required this.type,
    this.cardUid,
    this.pin,
    this.qrCode,
    this.facilityCode,
    this.cardNumber,
  });

  factory Credential.card({required String cardUid}) {
    return Credential(
      type: 'CREDENTIAL_TYPE_CARD',
      cardUid: cardUid,
    );
  }

  factory Credential.pin({required String pin}) {
    return Credential(
      type: 'CREDENTIAL_TYPE_PIN',
      pin: pin,
    );
  }

  factory Credential.cardAndPin({
    required String cardUid,
    required String pin,
  }) {
    return Credential(
      type: 'CREDENTIAL_TYPE_CARD_AND_PIN',
      cardUid: cardUid,
      pin: pin,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      if (cardUid != null) 'card_uid': cardUid,
      if (pin != null) 'pin': pin,
      if (qrCode != null) 'qr_code': qrCode,
      if (facilityCode != null) 'facility_code': facilityCode,
      if (cardNumber != null) 'card_number': cardNumber,
    };
  }

  @override
  String toString() => 'Credential(type: $type)';
}

/// User/credential information
@immutable
class CredentialInfo {
  final String userId;
  final String userName;
  final String userType;
  final String accessLevel;
  final Timestamp? validFrom;
  final Timestamp? validUntil;
  final List<String>? allowedDoors;
  final Map<String, dynamic>? attributes;

  const CredentialInfo({
    required this.userId,
    required this.userName,
    required this.userType,
    required this.accessLevel,
    this.validFrom,
    this.validUntil,
    this.allowedDoors,
    this.attributes,
  });

  factory CredentialInfo.fromJson(Map<String, dynamic> json) {
    return CredentialInfo(
      userId: json['user_id'] as String,
      userName: json['user_name'] as String,
      userType: json['user_type'] as String,
      accessLevel: json['access_level'] as String,
      validFrom: json['valid_from'] != null
          ? Timestamp.fromJson(json['valid_from'])
          : null,
      validUntil: json['valid_until'] != null
          ? Timestamp.fromJson(json['valid_until'])
          : null,
      allowedDoors: (json['allowed_doors'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      attributes: json['attributes'] as Map<String, dynamic>?,
    );
  }

  @override
  String toString() =>
      'CredentialInfo(userId: $userId, userName: $userName, accessLevel: $accessLevel)';
}

/// Door status
@immutable
class DoorStatus {
  final String deviceId;
  final String doorId;
  final String doorName;
  final LockStatus lockStatus;
  final DoorPosition doorPosition;
  final DoorMode mode;
  final Timestamp? lockedSince;
  final Timestamp? unlockedSince;
  final Map<String, dynamic>? statistics;

  const DoorStatus({
    required this.deviceId,
    required this.doorId,
    required this.doorName,
    required this.lockStatus,
    required this.doorPosition,
    required this.mode,
    this.lockedSince,
    this.unlockedSince,
    this.statistics,
  });

  factory DoorStatus.fromJson(Map<String, dynamic> json) {
    return DoorStatus(
      deviceId: json['device_id'] as String,
      doorId: json['door_id'] as String,
      doorName: json['door_name'] as String,
      lockStatus: LockStatus.fromString(json['lock_status'] as String),
      doorPosition: DoorPosition.fromString(json['door_position'] as String),
      mode: DoorMode.fromString(json['mode'] as String),
      lockedSince: json['locked_since'] != null
          ? Timestamp.fromJson(json['locked_since'])
          : null,
      unlockedSince: json['unlocked_since'] != null
          ? Timestamp.fromJson(json['unlocked_since'])
          : null,
      statistics: json['statistics'] as Map<String, dynamic>?,
    );
  }

  @override
  String toString() =>
      'DoorStatus(doorId: $doorId, lockStatus: $lockStatus, doorPosition: $doorPosition)';
}

/// Access check response
@immutable
class AccessCheckResponse {
  final AccessDecision decision;
  final String reason;
  final CredentialInfo? credentialInfo;
  final String? ruleId;
  final Timestamp timestamp;

  const AccessCheckResponse({
    required this.decision,
    required this.reason,
    this.credentialInfo,
    this.ruleId,
    required this.timestamp,
  });

  factory AccessCheckResponse.fromJson(Map<String, dynamic> json) {
    return AccessCheckResponse(
      decision: AccessDecision.fromString(json['decision'] as String),
      reason: json['reason'] as String,
      credentialInfo: json['credential_info'] != null
          ? CredentialInfo.fromJson(json['credential_info'] as Map<String, dynamic>)
          : null,
      ruleId: json['rule_id'] as String?,
      timestamp: Timestamp.fromJson(json['timestamp']),
    );
  }

  bool get isGranted => decision == AccessDecision.granted;
  bool get isDenied => decision == AccessDecision.denied;

  @override
  String toString() =>
      'AccessCheckResponse(decision: $decision, reason: $reason)';
}

/// Grant access response
@immutable
class GrantAccessResponse {
  final bool granted;
  final AccessDecision decision;
  final String reason;
  final bool doorUnlocked;
  final CredentialInfo? userInfo;
  final Timestamp timestamp;
  final String eventId;

  const GrantAccessResponse({
    required this.granted,
    required this.decision,
    required this.reason,
    required this.doorUnlocked,
    this.userInfo,
    required this.timestamp,
    required this.eventId,
  });

  factory GrantAccessResponse.fromJson(Map<String, dynamic> json) {
    return GrantAccessResponse(
      granted: json['granted'] as bool,
      decision: AccessDecision.fromString(json['decision'] as String),
      reason: json['reason'] as String,
      doorUnlocked: json['door_unlocked'] as bool,
      userInfo: json['user_info'] != null
          ? CredentialInfo.fromJson(json['user_info'] as Map<String, dynamic>)
          : null,
      timestamp: Timestamp.fromJson(json['timestamp']),
      eventId: json['event_id'] as String,
    );
  }

  @override
  String toString() =>
      'GrantAccessResponse(granted: $granted, doorUnlocked: $doorUnlocked, eventId: $eventId)';
}

/// Access event
@immutable
class AccessEvent {
  final String eventId;
  final String deviceId;
  final String doorId;
  final String type;
  final Timestamp timestamp;
  final Credential? credential;
  final CredentialInfo? userInfo;
  final AccessDecision? decision;
  final String? reason;
  final String severity;

  const AccessEvent({
    required this.eventId,
    required this.deviceId,
    required this.doorId,
    required this.type,
    required this.timestamp,
    this.credential,
    this.userInfo,
    this.decision,
    this.reason,
    required this.severity,
  });

  factory AccessEvent.fromJson(Map<String, dynamic> json) {
    return AccessEvent(
      eventId: json['event_id'] as String,
      deviceId: json['device_id'] as String,
      doorId: json['door_id'] as String,
      type: json['type'] as String,
      timestamp: Timestamp.fromJson(json['timestamp']),
      credential: json['credential'] != null
          ? Credential(
              type: json['credential']['type'] as String,
              cardUid: json['credential']['card_uid'] as String?,
              pin: json['credential']['pin'] as String?,
            )
          : null,
      userInfo: json['user_info'] != null
          ? CredentialInfo.fromJson(json['user_info'] as Map<String, dynamic>)
          : null,
      decision: json['decision'] != null
          ? AccessDecision.fromString(json['decision'] as String)
          : null,
      reason: json['reason'] as String?,
      severity: json['severity'] as String,
    );
  }

  @override
  String toString() =>
      'AccessEvent(eventId: $eventId, type: $type, doorId: $doorId)';
}

/// Controller status
@immutable
class ControllerStatus {
  final String deviceId;
  final String name;
  final DeviceStatus status;
  final String controllerType;
  final List<DoorStatus> doors;
  final String? model;
  final String? manufacturer;
  final bool lockdownActive;
  final LockdownLevel? lockdownLevel;
  final Map<String, dynamic>? statistics;

  const ControllerStatus({
    required this.deviceId,
    required this.name,
    required this.status,
    required this.controllerType,
    required this.doors,
    this.model,
    this.manufacturer,
    required this.lockdownActive,
    this.lockdownLevel,
    this.statistics,
  });

  factory ControllerStatus.fromJson(Map<String, dynamic> json) {
    return ControllerStatus(
      deviceId: json['device_id'] as String,
      name: json['name'] as String,
      status: DeviceStatus.fromString(json['status'] as String),
      controllerType: json['type'] as String,
      doors: (json['doors'] as List<dynamic>?)
              ?.map((e) => DoorStatus.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      model: json['model'] as String?,
      manufacturer: json['manufacturer'] as String?,
      lockdownActive: json['lockdown_active'] as bool? ?? false,
      lockdownLevel: json['lockdown_level'] != null
          ? LockdownLevel.fromString(json['lockdown_level'] as String)
          : null,
      statistics: json['statistics'] as Map<String, dynamic>?,
    );
  }

  @override
  String toString() =>
      'ControllerStatus(deviceId: $deviceId, status: $status, lockdownActive: $lockdownActive)';
}
