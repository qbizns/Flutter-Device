import 'package:meta/meta.dart';
import 'common_models.dart';

/// Card type enumeration
enum CardType {
  unknown,
  hidProx,
  em4100,
  indala,
  mifareClassic1K,
  mifareClassic4K,
  mifareUltralight,
  mifareUltralightC,
  mifarePlus,
  mifareDesfire,
  hidIClass,
  hidIClassSE,
  hidSeos,
  ntag213,
  ntag215,
  ntag216,
  iso14443A,
  iso14443B,
  iso15693,
  felica;

  static CardType fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('CARD_TYPE_', '').replaceAll('_', '');
    return CardType.values.firstWhere(
      (e) => e.name.toUpperCase().replaceAll('_', '') == normalized,
      orElse: () => CardType.unknown,
    );
  }
}

/// Card technology type
enum CardTechnology {
  unknown,
  prox125khz,
  contactless13_56mhz,
  nfc,
  dualFreq;

  static CardTechnology fromString(String value) {
    final map = {
      'PROX_125KHZ': prox125khz,
      'CONTACTLESS_13_56MHZ': contactless13_56mhz,
      'NFC': nfc,
      'DUAL_FREQ': dualFreq,
    };
    return map[value.toUpperCase().replaceAll('CARD_TECHNOLOGY_', '')] ??
        CardTechnology.unknown;
  }
}

/// Card data from RFID reader
@immutable
class CardData {
  final CardType type;
  final String uid;
  final String serialNumber;
  final CardTechnology technology;
  final Map<String, dynamic>? protocolData;
  final CardCapabilities? capabilities;

  const CardData({
    required this.type,
    required this.uid,
    required this.serialNumber,
    required this.technology,
    this.protocolData,
    this.capabilities,
  });

  factory CardData.fromJson(Map<String, dynamic> json) {
    return CardData(
      type: CardType.fromString(json['type'] as String),
      uid: json['uid'] as String,
      serialNumber: json['serial_number'] as String,
      technology: CardTechnology.fromString(json['technology'] as String),
      protocolData: json['protocol_data'] as Map<String, dynamic>?,
      capabilities: json['capabilities'] != null
          ? CardCapabilities.fromJson(json['capabilities'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type.name.toUpperCase(),
      'uid': uid,
      'serial_number': serialNumber,
      'technology': technology.name.toUpperCase(),
      if (protocolData != null) 'protocol_data': protocolData,
      if (capabilities != null) 'capabilities': capabilities!.toJson(),
    };
  }

  @override
  String toString() =>
      'CardData(type: $type, uid: $uid, technology: $technology)';
}

/// Card capabilities
@immutable
class CardCapabilities {
  final bool writable;
  final bool requiresAuth;
  final bool supportsEncryption;
  final bool supportsNdef;
  final int? maxDataSize;
  final bool readOnly;

  const CardCapabilities({
    required this.writable,
    required this.requiresAuth,
    this.supportsEncryption = false,
    this.supportsNdef = false,
    this.maxDataSize,
    this.readOnly = false,
  });

  factory CardCapabilities.fromJson(Map<String, dynamic> json) {
    return CardCapabilities(
      writable: json['writable'] as bool? ?? false,
      requiresAuth: json['requires_auth'] as bool? ?? false,
      supportsEncryption: json['supports_encryption'] as bool? ?? false,
      supportsNdef: json['supports_ndef'] as bool? ?? false,
      maxDataSize: json['max_data_size'] as int?,
      readOnly: json['read_only'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'writable': writable,
      'requires_auth': requiresAuth,
      'supports_encryption': supportsEncryption,
      'supports_ndef': supportsNdef,
      if (maxDataSize != null) 'max_data_size': maxDataSize,
      'read_only': readOnly,
    };
  }
}

/// Card read response
@immutable
class CardReadResponse {
  final CardData card;
  final String deviceId;
  final Timestamp timestamp;
  final int signalStrength;

  const CardReadResponse({
    required this.card,
    required this.deviceId,
    required this.timestamp,
    required this.signalStrength,
  });

  factory CardReadResponse.fromJson(Map<String, dynamic> json) {
    return CardReadResponse(
      card: CardData.fromJson(json['card'] as Map<String, dynamic>),
      deviceId: json['device_id'] as String,
      timestamp: Timestamp.fromJson(json['timestamp']),
      signalStrength: json['signal_strength'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'card': card.toJson(),
      'device_id': deviceId,
      'timestamp': timestamp.toJson(),
      'signal_strength': signalStrength,
    };
  }

  @override
  String toString() =>
      'CardReadResponse(deviceId: $deviceId, cardType: ${card.type}, signalStrength: $signalStrength)';
}

/// Reader status
@immutable
class ReaderStatus {
  final String deviceId;
  final DeviceStatus status;
  final String mode;
  final List<CardType> supportedTypes;
  final bool cardPresent;
  final String? currentCardUid;
  final String? model;
  final String? manufacturer;
  final Map<String, dynamic>? statistics;

  const ReaderStatus({
    required this.deviceId,
    required this.status,
    required this.mode,
    required this.supportedTypes,
    required this.cardPresent,
    this.currentCardUid,
    this.model,
    this.manufacturer,
    this.statistics,
  });

  factory ReaderStatus.fromJson(Map<String, dynamic> json) {
    return ReaderStatus(
      deviceId: json['device_id'] as String,
      status: DeviceStatus.fromString(json['status'] as String),
      mode: json['mode'] as String,
      supportedTypes: (json['supported_types'] as List<dynamic>?)
              ?.map((e) => CardType.fromString(e as String))
              .toList() ??
          [],
      cardPresent: json['card_present'] as bool? ?? false,
      currentCardUid: json['current_card_uid'] as String?,
      model: json['model'] as String?,
      manufacturer: json['manufacturer'] as String?,
      statistics: json['statistics'] as Map<String, dynamic>?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'device_id': deviceId,
      'status': status.name.toUpperCase(),
      'mode': mode,
      'supported_types': supportedTypes.map((e) => e.name.toUpperCase()).toList(),
      'card_present': cardPresent,
      if (currentCardUid != null) 'current_card_uid': currentCardUid,
      if (model != null) 'model': model,
      if (manufacturer != null) 'manufacturer': manufacturer,
      if (statistics != null) 'statistics': statistics,
    };
  }

  @override
  String toString() =>
      'ReaderStatus(deviceId: $deviceId, status: $status, cardPresent: $cardPresent)';
}

/// Card read event (for streaming)
@immutable
class CardReadEvent {
  final String eventId;
  final String deviceId;
  final CardData card;
  final Timestamp timestamp;
  final int signalStrength;
  final int readTimeMs;

  const CardReadEvent({
    required this.eventId,
    required this.deviceId,
    required this.card,
    required this.timestamp,
    required this.signalStrength,
    required this.readTimeMs,
  });

  factory CardReadEvent.fromJson(Map<String, dynamic> json) {
    return CardReadEvent(
      eventId: json['event_id'] as String,
      deviceId: json['device_id'] as String,
      card: CardData.fromJson(json['card'] as Map<String, dynamic>),
      timestamp: Timestamp.fromJson(json['timestamp']),
      signalStrength: json['signal_strength'] as int,
      readTimeMs: json['read_time_ms'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'event_id': eventId,
      'device_id': deviceId,
      'card': card.toJson(),
      'timestamp': timestamp.toJson(),
      'signal_strength': signalStrength,
      'read_time_ms': readTimeMs,
    };
  }

  @override
  String toString() =>
      'CardReadEvent(eventId: $eventId, cardType: ${card.type})';
}

/// Authentication request credentials
@immutable
class AuthenticationCredentials {
  final List<String>? mifareKeysA;
  final List<String>? mifareKeysB;
  final String? iclassKey;

  const AuthenticationCredentials({
    this.mifareKeysA,
    this.mifareKeysB,
    this.iclassKey,
  });

  Map<String, dynamic> toJson() {
    return {
      if (mifareKeysA != null) 'mifare_keys_a': mifareKeysA,
      if (mifareKeysB != null) 'mifare_keys_b': mifareKeysB,
      if (iclassKey != null) 'iclass_key': iclassKey,
    };
  }
}

/// Authentication response
@immutable
class AuthenticationResponse {
  final bool authenticated;
  final String? reason;
  final Map<String, dynamic>? credentialInfo;
  final Timestamp timestamp;

  const AuthenticationResponse({
    required this.authenticated,
    this.reason,
    this.credentialInfo,
    required this.timestamp,
  });

  factory AuthenticationResponse.fromJson(Map<String, dynamic> json) {
    return AuthenticationResponse(
      authenticated: json['authenticated'] as bool,
      reason: json['reason'] as String?,
      credentialInfo: json['credential_info'] as Map<String, dynamic>?,
      timestamp: Timestamp.fromJson(json['timestamp']),
    );
  }

  @override
  String toString() =>
      'AuthenticationResponse(authenticated: $authenticated, reason: $reason)';
}
