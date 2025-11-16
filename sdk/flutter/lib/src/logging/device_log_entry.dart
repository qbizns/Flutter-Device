import 'package:meta/meta.dart';

/// Types of device interactions that can be logged
enum DeviceInteractionType {
  print,
  scan,
  weigh,
  rfidRead,
  rfidWrite,
  accessGrant,
  accessDeny,
  badgePrint,
  paymentStart,
  paymentComplete,
  paymentRefund,
  error,
  statusCheck,
  configuration,
}

/// Base class for device log entries
@immutable
abstract class DeviceLogEntry {
  /// Unique identifier for this log entry
  final String id;

  /// Timestamp when the interaction occurred
  final DateTime timestamp;

  /// ID of the device that performed the interaction
  final String deviceId;

  /// Type of interaction
  final DeviceInteractionType type;

  /// Whether the interaction was successful
  final bool success;

  /// Optional error message if interaction failed
  final String? errorMessage;

  /// Additional metadata
  final Map<String, dynamic> metadata;

  const DeviceLogEntry({
    required this.id,
    required this.timestamp,
    required this.deviceId,
    required this.type,
    this.success = true,
    this.errorMessage,
    this.metadata = const {},
  });

  /// Convert to JSON format
  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'timestamp': timestamp.toIso8601String(),
      'deviceId': deviceId,
      'type': type.name,
      'success': success,
      if (errorMessage != null) 'error': errorMessage,
      ...metadata,
    };
  }

  /// Format as a text line for .txt files
  String toTextLine();

  /// Format as a CSV row
  String toCsvRow();

  /// Get CSV header
  static String getCsvHeader(DeviceInteractionType type) {
    switch (type) {
      case DeviceInteractionType.scan:
        return 'ID,Timestamp,Device ID,Barcode Data,Symbology,Success,Error';
      case DeviceInteractionType.weigh:
        return 'ID,Timestamp,Device ID,Weight,Unit,Success,Error';
      case DeviceInteractionType.rfidRead:
      case DeviceInteractionType.rfidWrite:
        return 'ID,Timestamp,Device ID,Card UID,Card Type,Operation,Success,Error';
      case DeviceInteractionType.accessGrant:
      case DeviceInteractionType.accessDeny:
        return 'ID,Timestamp,Device ID,Door ID,Credential,Access Result,Error';
      case DeviceInteractionType.paymentStart:
      case DeviceInteractionType.paymentComplete:
      case DeviceInteractionType.paymentRefund:
        return 'ID,Timestamp,Device ID,Transaction ID,Amount,Currency,Payment Method,Status,Error';
      default:
        return 'ID,Timestamp,Device ID,Type,Success,Error,Metadata';
    }
  }
}

/// Log entry for printer interactions
class PrinterLogEntry extends DeviceLogEntry {
  final String jobId;
  final String content;
  final int? copies;

  const PrinterLogEntry({
    required super.id,
    required super.timestamp,
    required super.deviceId,
    required this.jobId,
    required this.content,
    this.copies,
    super.success,
    super.errorMessage,
    super.metadata,
  }) : super(type: DeviceInteractionType.print);

  @override
  String toTextLine() {
    final buffer = StringBuffer();
    buffer.writeln('=' * 80);
    buffer.writeln('Print Job: $jobId');
    buffer.writeln('Device: $deviceId');
    buffer.writeln('Timestamp: ${timestamp.toIso8601String()}');
    buffer.writeln('Status: ${success ? 'SUCCESS' : 'FAILED'}');
    if (copies != null && copies! > 1) {
      buffer.writeln('Copies: $copies');
    }
    if (!success && errorMessage != null) {
      buffer.writeln('Error: $errorMessage');
    }
    buffer.writeln('-' * 80);
    buffer.writeln(content);
    buffer.writeln('=' * 80);
    buffer.writeln();
    return buffer.toString();
  }

  @override
  String toCsvRow() {
    final escapedContent = content.replaceAll('"', '""').replaceAll('\n', '\\n');
    return '$id,${timestamp.toIso8601String()},$deviceId,$jobId,"$escapedContent",${copies ?? 1},$success,${errorMessage ?? ""}';
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      ...super.toJson(),
      'jobId': jobId,
      'content': content,
      if (copies != null) 'copies': copies,
    };
  }
}

/// Log entry for scanner interactions
class ScannerLogEntry extends DeviceLogEntry {
  final String barcodeData;
  final String symbology;

  const ScannerLogEntry({
    required super.id,
    required super.timestamp,
    required super.deviceId,
    required this.barcodeData,
    required this.symbology,
    super.success,
    super.errorMessage,
    super.metadata,
  }) : super(type: DeviceInteractionType.scan);

  @override
  String toTextLine() {
    return '${timestamp.toIso8601String()} | $deviceId | $symbology | $barcodeData | ${success ? 'SUCCESS' : 'FAILED'}';
  }

  @override
  String toCsvRow() {
    final escapedData = barcodeData.replaceAll('"', '""');
    return '$id,${timestamp.toIso8601String()},$deviceId,"$escapedData",$symbology,$success,${errorMessage ?? ""}';
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      ...super.toJson(),
      'barcodeData': barcodeData,
      'symbology': symbology,
    };
  }
}

/// Log entry for scale interactions
class ScaleLogEntry extends DeviceLogEntry {
  final double weight;
  final String unit;
  final bool stable;

  const ScaleLogEntry({
    required super.id,
    required super.timestamp,
    required super.deviceId,
    required this.weight,
    required this.unit,
    this.stable = true,
    super.success,
    super.errorMessage,
    super.metadata,
  }) : super(type: DeviceInteractionType.weigh);

  @override
  String toTextLine() {
    return '${timestamp.toIso8601String()} | $deviceId | $weight $unit | ${stable ? 'STABLE' : 'UNSTABLE'} | ${success ? 'SUCCESS' : 'FAILED'}';
  }

  @override
  String toCsvRow() {
    return '$id,${timestamp.toIso8601String()},$deviceId,$weight,$unit,$stable,$success,${errorMessage ?? ""}';
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      ...super.toJson(),
      'weight': weight,
      'unit': unit,
      'stable': stable,
    };
  }
}

/// Log entry for RFID reader interactions
class RFIDLogEntry extends DeviceLogEntry {
  final String cardUid;
  final String cardType;
  final String operation;
  final Map<String, dynamic>? cardData;

  const RFIDLogEntry({
    required super.id,
    required super.timestamp,
    required super.deviceId,
    required super.type,
    required this.cardUid,
    required this.cardType,
    required this.operation,
    this.cardData,
    super.success,
    super.errorMessage,
    super.metadata,
  });

  @override
  String toTextLine() {
    return '${timestamp.toIso8601String()} | $deviceId | $operation | $cardType | $cardUid | ${success ? 'SUCCESS' : 'FAILED'}';
  }

  @override
  String toCsvRow() {
    final escapedUid = cardUid.replaceAll('"', '""');
    return '$id,${timestamp.toIso8601String()},$deviceId,"$escapedUid",$cardType,$operation,$success,${errorMessage ?? ""}';
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      ...super.toJson(),
      'cardUid': cardUid,
      'cardType': cardType,
      'operation': operation,
      if (cardData != null) 'cardData': cardData,
    };
  }
}

/// Log entry for access control interactions
class AccessControlLogEntry extends DeviceLogEntry {
  final String doorId;
  final String credential;
  final String accessResult;

  const AccessControlLogEntry({
    required super.id,
    required super.timestamp,
    required super.deviceId,
    required super.type,
    required this.doorId,
    required this.credential,
    required this.accessResult,
    super.success,
    super.errorMessage,
    super.metadata,
  });

  @override
  String toTextLine() {
    return '${timestamp.toIso8601String()} | $deviceId | Door: $doorId | Credential: $credential | $accessResult | ${success ? 'SUCCESS' : 'FAILED'}';
  }

  @override
  String toCsvRow() {
    final escapedCredential = credential.replaceAll('"', '""');
    return '$id,${timestamp.toIso8601String()},$deviceId,$doorId,"$escapedCredential",$accessResult,$success,${errorMessage ?? ""}';
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      ...super.toJson(),
      'doorId': doorId,
      'credential': credential,
      'accessResult': accessResult,
    };
  }
}

/// Log entry for payment terminal interactions
class PaymentLogEntry extends DeviceLogEntry {
  final String transactionId;
  final double amount;
  final String currency;
  final String paymentMethod;
  final String status;

  const PaymentLogEntry({
    required super.id,
    required super.timestamp,
    required super.deviceId,
    required super.type,
    required this.transactionId,
    required this.amount,
    required this.currency,
    required this.paymentMethod,
    required this.status,
    super.success,
    super.errorMessage,
    super.metadata,
  });

  @override
  String toTextLine() {
    return '${timestamp.toIso8601String()} | $deviceId | TX: $transactionId | $amount $currency | $paymentMethod | $status | ${success ? 'SUCCESS' : 'FAILED'}';
  }

  @override
  String toCsvRow() {
    return '$id,${timestamp.toIso8601String()},$deviceId,$transactionId,$amount,$currency,$paymentMethod,$status,$success,${errorMessage ?? ""}';
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      ...super.toJson(),
      'transactionId': transactionId,
      'amount': amount,
      'currency': currency,
      'paymentMethod': paymentMethod,
      'status': status,
    };
  }
}

/// Log entry for badge printer interactions
class BadgePrinterLogEntry extends DeviceLogEntry {
  final String jobId;
  final String badgeType;
  final Map<String, dynamic> badgeData;

  const BadgePrinterLogEntry({
    required super.id,
    required super.timestamp,
    required super.deviceId,
    required this.jobId,
    required this.badgeType,
    required this.badgeData,
    super.success,
    super.errorMessage,
    super.metadata,
  }) : super(type: DeviceInteractionType.badgePrint);

  @override
  String toTextLine() {
    final buffer = StringBuffer();
    buffer.writeln('=' * 80);
    buffer.writeln('Badge Print Job: $jobId');
    buffer.writeln('Device: $deviceId');
    buffer.writeln('Timestamp: ${timestamp.toIso8601String()}');
    buffer.writeln('Badge Type: $badgeType');
    buffer.writeln('Status: ${success ? 'SUCCESS' : 'FAILED'}');
    if (!success && errorMessage != null) {
      buffer.writeln('Error: $errorMessage');
    }
    buffer.writeln('-' * 80);
    buffer.writeln('Badge Data:');
    badgeData.forEach((key, value) {
      buffer.writeln('  $key: $value');
    });
    buffer.writeln('=' * 80);
    buffer.writeln();
    return buffer.toString();
  }

  @override
  String toCsvRow() {
    return '$id,${timestamp.toIso8601String()},$deviceId,$jobId,$badgeType,$success,${errorMessage ?? ""}';
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      ...super.toJson(),
      'jobId': jobId,
      'badgeType': badgeType,
      'badgeData': badgeData,
    };
  }
}
