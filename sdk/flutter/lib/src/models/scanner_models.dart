import 'package:meta/meta.dart';
import 'common_models.dart';

/// Scan type enumeration
enum ScanType {
  unknown,
  barcode,
  qrCode,
  dataMatrix,
  pdf417,
  aztec;

  static ScanType fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('SCAN_TYPE_', '');
    final map = {
      'BARCODE': barcode,
      'QR_CODE': qrCode,
      'DATA_MATRIX': dataMatrix,
      'PDF417': pdf417,
      'AZTEC': aztec,
    };
    return map[normalized] ?? ScanType.unknown;
  }
}

/// Barcode symbology enumeration
enum BarcodeSymbology {
  unknown,
  upca,
  upce,
  ean13,
  ean8,
  code39,
  code93,
  code128,
  itf,
  codabar,
  qr,
  dataMatrix,
  pdf417,
  aztec;

  static BarcodeSymbology fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('SYMBOLOGY_', '');
    return BarcodeSymbology.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized.replaceAll('_', ''),
      orElse: () => BarcodeSymbology.unknown,
    );
  }
}

/// Scanner mode enumeration
enum ScannerMode {
  manual,
  continuous,
  trigger;

  static ScannerMode fromString(String value) {
    return ScannerMode.values.firstWhere(
      (e) => e.name.toUpperCase() == value.toUpperCase(),
      orElse: () => ScannerMode.manual,
    );
  }
}

/// Scan result
@immutable
class ScanResult {
  final String data;
  final ScanType type;
  final BarcodeSymbology symbology;
  final String deviceId;
  final Timestamp timestamp;
  final int quality;
  final Map<String, dynamic>? metadata;

  const ScanResult({
    required this.data,
    required this.type,
    required this.symbology,
    required this.deviceId,
    required this.timestamp,
    required this.quality,
    this.metadata,
  });

  factory ScanResult.fromJson(Map<String, dynamic> json) {
    return ScanResult(
      data: json['data'] as String,
      type: ScanType.fromString(json['type'] as String),
      symbology: BarcodeSymbology.fromString(json['symbology'] as String),
      deviceId: json['device_id'] as String,
      timestamp: Timestamp.fromJson(json['timestamp']),
      quality: json['quality'] as int? ?? 0,
      metadata: json['metadata'] as Map<String, dynamic>?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'data': data,
      'type': type.name.toUpperCase(),
      'symbology': symbology.name.toUpperCase(),
      'device_id': deviceId,
      'timestamp': timestamp.toJson(),
      'quality': quality,
      if (metadata != null) 'metadata': metadata,
    };
  }

  @override
  String toString() =>
      'ScanResult(data: $data, type: $type, symbology: $symbology)';
}

/// Scanner capabilities
@immutable
class ScannerCapabilities {
  final List<BarcodeSymbology> supportedSymbologies;
  final bool supportsBarcode;
  final bool supportsQRCode;
  final bool supportsDataMatrix;
  final bool supportsPDF417;
  final bool supportsAztec;
  final bool supportsContinuousMode;
  final bool hasTriggerButton;
  final bool hasAimingLaser;

  const ScannerCapabilities({
    required this.supportedSymbologies,
    required this.supportsBarcode,
    required this.supportsQRCode,
    required this.supportsDataMatrix,
    required this.supportsPDF417,
    required this.supportsAztec,
    required this.supportsContinuousMode,
    required this.hasTriggerButton,
    required this.hasAimingLaser,
  });

  factory ScannerCapabilities.fromJson(Map<String, dynamic> json) {
    return ScannerCapabilities(
      supportedSymbologies: (json['supported_symbologies'] as List<dynamic>?)
              ?.map((e) => BarcodeSymbology.fromString(e as String))
              .toList() ??
          [],
      supportsBarcode: json['supports_barcode'] as bool? ?? false,
      supportsQRCode: json['supports_qr_code'] as bool? ?? false,
      supportsDataMatrix: json['supports_data_matrix'] as bool? ?? false,
      supportsPDF417: json['supports_pdf417'] as bool? ?? false,
      supportsAztec: json['supports_aztec'] as bool? ?? false,
      supportsContinuousMode: json['supports_continuous_mode'] as bool? ?? false,
      hasTriggerButton: json['has_trigger_button'] as bool? ?? false,
      hasAimingLaser: json['has_aiming_laser'] as bool? ?? false,
    );
  }

  @override
  String toString() =>
      'ScannerCapabilities(barcode: $supportsBarcode, qr: $supportsQRCode, continuous: $supportsContinuousMode)';
}

/// Scanner status
@immutable
class ScannerStatus {
  final String deviceId;
  final String name;
  final DeviceStatus status;
  final ScannerMode mode;
  final bool triggerPressed;
  final int? batteryLevel;
  final ScannerCapabilities? capabilities;
  final String? errorMessage;

  const ScannerStatus({
    required this.deviceId,
    required this.name,
    required this.status,
    required this.mode,
    required this.triggerPressed,
    this.batteryLevel,
    this.capabilities,
    this.errorMessage,
  });

  factory ScannerStatus.fromJson(Map<String, dynamic> json) {
    return ScannerStatus(
      deviceId: json['device_id'] as String,
      name: json['name'] as String,
      status: DeviceStatus.fromString(json['status'] as String),
      mode: ScannerMode.fromString(json['mode'] as String),
      triggerPressed: json['trigger_pressed'] as bool? ?? false,
      batteryLevel: json['battery_level'] as int?,
      capabilities: json['capabilities'] != null
          ? ScannerCapabilities.fromJson(
              json['capabilities'] as Map<String, dynamic>)
          : null,
      errorMessage: json['error_message'] as String?,
    );
  }

  bool get isReady => status == DeviceStatus.ready;
  bool get hasError => status == DeviceStatus.error || errorMessage != null;
  bool get isBatteryLow => batteryLevel != null && batteryLevel! < 20;

  @override
  String toString() =>
      'ScannerStatus(deviceId: $deviceId, status: $status, mode: $mode)';
}

/// Scan event for real-time scanning
@immutable
class ScanEvent {
  final String eventId;
  final ScanResult scan;
  final String? userId;
  final Map<String, dynamic>? context;

  const ScanEvent({
    required this.eventId,
    required this.scan,
    this.userId,
    this.context,
  });

  factory ScanEvent.fromJson(Map<String, dynamic> json) {
    return ScanEvent(
      eventId: json['event_id'] as String,
      scan: ScanResult.fromJson(json['scan'] as Map<String, dynamic>),
      userId: json['user_id'] as String?,
      context: json['context'] as Map<String, dynamic>?,
    );
  }

  @override
  String toString() => 'ScanEvent(eventId: $eventId, data: ${scan.data})';
}

/// Scanner configuration
@immutable
class ScannerConfig {
  final ScannerMode mode;
  final List<BarcodeSymbology>? enabledSymbologies;
  final int? timeout;
  final bool enableBeep;
  final bool enableVibration;
  final bool enableAimingLaser;
  final int? beepDuration;
  final int? vibrationDuration;

  const ScannerConfig({
    required this.mode,
    this.enabledSymbologies,
    this.timeout,
    this.enableBeep = true,
    this.enableVibration = true,
    this.enableAimingLaser = true,
    this.beepDuration,
    this.vibrationDuration,
  });

  Map<String, dynamic> toJson() {
    return {
      'mode': mode.name.toUpperCase(),
      if (enabledSymbologies != null)
        'enabled_symbologies':
            enabledSymbologies!.map((e) => e.name.toUpperCase()).toList(),
      if (timeout != null) 'timeout_seconds': timeout,
      'enable_beep': enableBeep,
      'enable_vibration': enableVibration,
      'enable_aiming_laser': enableAimingLaser,
      if (beepDuration != null) 'beep_duration_ms': beepDuration,
      if (vibrationDuration != null) 'vibration_duration_ms': vibrationDuration,
    };
  }

  @override
  String toString() =>
      'ScannerConfig(mode: $mode, beep: $enableBeep, vibration: $enableVibration)';
}

/// Scan options for one-time scans
@immutable
class ScanOptions {
  final Duration timeout;
  final List<BarcodeSymbology>? expectedSymbologies;
  final bool enableBeep;
  final bool enableVibration;

  const ScanOptions({
    this.timeout = const Duration(seconds: 30),
    this.expectedSymbologies,
    this.enableBeep = true,
    this.enableVibration = true,
  });

  Map<String, dynamic> toJson() {
    return {
      'timeout_seconds': timeout.inSeconds,
      if (expectedSymbologies != null)
        'expected_symbologies':
            expectedSymbologies!.map((e) => e.name.toUpperCase()).toList(),
      'enable_beep': enableBeep,
      'enable_vibration': enableVibration,
    };
  }

  @override
  String toString() => 'ScanOptions(timeout: ${timeout.inSeconds}s)';
}
