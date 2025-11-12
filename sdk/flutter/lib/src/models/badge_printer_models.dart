import 'package:meta/meta.dart';
import 'common_models.dart';

/// Card size enumeration
enum CardSize {
  unknown,
  cr80,
  cr79,
  cr100;

  static CardSize fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('CARD_SIZE_', '');
    return CardSize.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => CardSize.unknown,
    );
  }

  /// Returns card dimensions in millimeters [width, height]
  List<double> get dimensions {
    switch (this) {
      case cr80:
        return [85.6, 53.98]; // Standard credit card size
      case cr79:
        return [79.0, 49.0];
      case cr100:
        return [98.5, 67.0];
      default:
        return [0, 0];
    }
  }
}

/// Print side enumeration
enum PrintSide {
  front,
  back,
  both;

  static PrintSide fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('PRINT_SIDE_', '');
    return PrintSide.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => PrintSide.front,
    );
  }
}

/// Card orientation enumeration
enum CardOrientation {
  portrait,
  landscape;

  static CardOrientation fromString(String value) {
    return CardOrientation.values.firstWhere(
      (e) => e.name.toUpperCase() == value.toUpperCase(),
      orElse: () => CardOrientation.landscape,
    );
  }
}

/// Print quality enumeration
enum PrintQuality {
  draft,
  normal,
  high;

  static PrintQuality fromString(String value) {
    return PrintQuality.values.firstWhere(
      (e) => e.name.toUpperCase() == value.toUpperCase(),
      orElse: () => PrintQuality.normal,
    );
  }
}

/// Magnetic stripe track enumeration
enum MagneticTrack {
  track1,
  track2,
  track3;

  static MagneticTrack fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('TRACK_', '').replaceAll('TRACK', '');
    final map = {
      '1': track1,
      '2': track2,
      '3': track3,
    };
    return map[normalized] ?? MagneticTrack.track1;
  }
}

/// Ribbon type enumeration
enum RibbonType {
  unknown,
  ymcko,
  kO,
  mono,
  halfPanel;

  static RibbonType fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('RIBBON_TYPE_', '');
    final map = {
      'YMCKO': ymcko,
      'KO': kO,
      'MONO': mono,
      'HALF_PANEL': halfPanel,
    };
    return map[normalized] ?? RibbonType.unknown;
  }
}

/// Magnetic stripe data
@immutable
class MagneticStripeData {
  final String? track1;
  final String? track2;
  final String? track3;
  final bool verify;

  const MagneticStripeData({
    this.track1,
    this.track2,
    this.track3,
    this.verify = true,
  });

  Map<String, dynamic> toJson() {
    return {
      if (track1 != null) 'track1': track1,
      if (track2 != null) 'track2': track2,
      if (track3 != null) 'track3': track3,
      'verify': verify,
    };
  }

  @override
  String toString() => 'MagneticStripeData(verify: $verify)';
}

/// RFID encoding data
@immutable
class RFIDEncodingData {
  final String uid;
  final Map<int, String>? blockData;
  final List<String>? authKeys;
  final bool verify;

  const RFIDEncodingData({
    required this.uid,
    this.blockData,
    this.authKeys,
    this.verify = true,
  });

  Map<String, dynamic> toJson() {
    return {
      'uid': uid,
      if (blockData != null)
        'block_data': blockData!.map((k, v) => MapEntry(k.toString(), v)),
      if (authKeys != null) 'auth_keys': authKeys,
      'verify': verify,
    };
  }

  @override
  String toString() => 'RFIDEncodingData(uid: $uid, verify: $verify)';
}

/// Badge element base class
@immutable
abstract class BadgeElement {
  final Position position;
  final Size size;

  const BadgeElement({
    required this.position,
    required this.size,
  });

  Map<String, dynamic> toJson();
}

/// Text element for badge
@immutable
class BadgeTextElement extends BadgeElement {
  final String text;
  final String fontFamily;
  final int fontSize;
  final RGBColor color;
  final bool bold;
  final bool italic;
  final String alignment;

  const BadgeTextElement({
    required super.position,
    required super.size,
    required this.text,
    this.fontFamily = 'Arial',
    this.fontSize = 12,
    this.color = const RGBColor(red: 0, green: 0, blue: 0),
    this.bold = false,
    this.italic = false,
    this.alignment = 'left',
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'text',
      'position': position.toJson(),
      'size': size.toJson(),
      'text': text,
      'font_family': fontFamily,
      'font_size': fontSize,
      'color': color.toJson(),
      'bold': bold,
      'italic': italic,
      'alignment': alignment,
    };
  }

  @override
  String toString() => 'BadgeTextElement(text: $text, fontSize: $fontSize)';
}

/// Image element for badge
@immutable
class BadgeImageElement extends BadgeElement {
  final String imageData; // Base64 encoded
  final bool maintainAspectRatio;

  const BadgeImageElement({
    required super.position,
    required super.size,
    required this.imageData,
    this.maintainAspectRatio = true,
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'image',
      'position': position.toJson(),
      'size': size.toJson(),
      'image_data': imageData,
      'maintain_aspect_ratio': maintainAspectRatio,
    };
  }

  @override
  String toString() => 'BadgeImageElement(size: ${size.width}x${size.height})';
}

/// Barcode element for badge
@immutable
class BadgeBarcodeElement extends BadgeElement {
  final String data;
  final String format;
  final bool showText;

  const BadgeBarcodeElement({
    required super.position,
    required super.size,
    required this.data,
    required this.format,
    this.showText = true,
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'barcode',
      'position': position.toJson(),
      'size': size.toJson(),
      'data': data,
      'format': format,
      'show_text': showText,
    };
  }

  @override
  String toString() => 'BadgeBarcodeElement(data: $data, format: $format)';
}

/// QR code element for badge
@immutable
class BadgeQRCodeElement extends BadgeElement {
  final String data;
  final String errorCorrection;

  const BadgeQRCodeElement({
    required super.position,
    required super.size,
    required this.data,
    this.errorCorrection = 'medium',
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'qr_code',
      'position': position.toJson(),
      'size': size.toJson(),
      'data': data,
      'error_correction': errorCorrection,
    };
  }

  @override
  String toString() => 'BadgeQRCodeElement(data: $data)';
}

/// Card layout for badge printing
@immutable
class CardLayout {
  final CardSize cardSize;
  final CardOrientation orientation;
  final RGBColor? backgroundColor;
  final List<BadgeElement> elements;

  const CardLayout({
    required this.cardSize,
    this.orientation = CardOrientation.landscape,
    this.backgroundColor,
    required this.elements,
  });

  Map<String, dynamic> toJson() {
    return {
      'card_size': cardSize.name.toUpperCase(),
      'orientation': orientation.name,
      if (backgroundColor != null) 'background_color': backgroundColor!.toJson(),
      'elements': elements.map((e) => e.toJson()).toList(),
    };
  }

  @override
  String toString() =>
      'CardLayout(cardSize: $cardSize, elements: ${elements.length})';
}

/// Badge print job
@immutable
class BadgePrintJob {
  final CardLayout? frontLayout;
  final CardLayout? backLayout;
  final PrintSide printSide;
  final PrintQuality quality;
  final MagneticStripeData? magneticData;
  final RFIDEncodingData? rfidData;
  final bool applyOverlay;
  final int copies;

  const BadgePrintJob({
    this.frontLayout,
    this.backLayout,
    this.printSide = PrintSide.front,
    this.quality = PrintQuality.normal,
    this.magneticData,
    this.rfidData,
    this.applyOverlay = true,
    this.copies = 1,
  });

  Map<String, dynamic> toJson() {
    return {
      if (frontLayout != null) 'front_layout': frontLayout!.toJson(),
      if (backLayout != null) 'back_layout': backLayout!.toJson(),
      'print_side': printSide.name.toUpperCase(),
      'quality': quality.name.toUpperCase(),
      if (magneticData != null) 'magnetic_data': magneticData!.toJson(),
      if (rfidData != null) 'rfid_data': rfidData!.toJson(),
      'apply_overlay': applyOverlay,
      'copies': copies,
    };
  }

  @override
  String toString() => 'BadgePrintJob(printSide: $printSide, copies: $copies)';
}

/// Badge printer capabilities
@immutable
class BadgePrinterCapabilities {
  final List<CardSize> supportedCardSizes;
  final bool hasMagneticStripe;
  final bool hasRFIDEncoder;
  final bool supportsColorPrinting;
  final bool supportsDuplexPrinting;
  final int maxPrintResolution;
  final List<RibbonType> supportedRibbons;

  const BadgePrinterCapabilities({
    required this.supportedCardSizes,
    required this.hasMagneticStripe,
    required this.hasRFIDEncoder,
    required this.supportsColorPrinting,
    required this.supportsDuplexPrinting,
    required this.maxPrintResolution,
    required this.supportedRibbons,
  });

  factory BadgePrinterCapabilities.fromJson(Map<String, dynamic> json) {
    return BadgePrinterCapabilities(
      supportedCardSizes: (json['supported_card_sizes'] as List<dynamic>?)
              ?.map((e) => CardSize.fromString(e as String))
              .toList() ??
          [],
      hasMagneticStripe: json['has_magnetic_stripe'] as bool? ?? false,
      hasRFIDEncoder: json['has_rfid_encoder'] as bool? ?? false,
      supportsColorPrinting: json['supports_color_printing'] as bool? ?? false,
      supportsDuplexPrinting:
          json['supports_duplex_printing'] as bool? ?? false,
      maxPrintResolution: json['max_print_resolution'] as int? ?? 300,
      supportedRibbons: (json['supported_ribbons'] as List<dynamic>?)
              ?.map((e) => RibbonType.fromString(e as String))
              .toList() ??
          [],
    );
  }

  @override
  String toString() =>
      'BadgePrinterCapabilities(colorPrinting: $supportsColorPrinting, hasMagStripe: $hasMagneticStripe)';
}

/// Badge printer status
@immutable
class BadgePrinterStatus {
  final String deviceId;
  final String name;
  final DeviceStatus status;
  final int cardsRemaining;
  final int ribbonRemaining;
  final RibbonType? ribbonType;
  final bool coverOpen;
  final String? errorMessage;
  final BadgePrinterCapabilities? capabilities;

  const BadgePrinterStatus({
    required this.deviceId,
    required this.name,
    required this.status,
    required this.cardsRemaining,
    required this.ribbonRemaining,
    this.ribbonType,
    required this.coverOpen,
    this.errorMessage,
    this.capabilities,
  });

  factory BadgePrinterStatus.fromJson(Map<String, dynamic> json) {
    return BadgePrinterStatus(
      deviceId: json['device_id'] as String,
      name: json['name'] as String,
      status: DeviceStatus.fromString(json['status'] as String),
      cardsRemaining: json['cards_remaining'] as int? ?? 0,
      ribbonRemaining: json['ribbon_remaining'] as int? ?? 0,
      ribbonType: json['ribbon_type'] != null
          ? RibbonType.fromString(json['ribbon_type'] as String)
          : null,
      coverOpen: json['cover_open'] as bool? ?? false,
      errorMessage: json['error_message'] as String?,
      capabilities: json['capabilities'] != null
          ? BadgePrinterCapabilities.fromJson(
              json['capabilities'] as Map<String, dynamic>)
          : null,
    );
  }

  bool get isReady =>
      status == DeviceStatus.ready &&
      cardsRemaining > 0 &&
      ribbonRemaining > 0 &&
      !coverOpen;
  bool get hasError => status == DeviceStatus.error || errorMessage != null;
  bool get isCardsLow => cardsRemaining < 10;
  bool get isRibbonLow => ribbonRemaining < 10;

  @override
  String toString() =>
      'BadgePrinterStatus(deviceId: $deviceId, status: $status, cardsRemaining: $cardsRemaining)';
}

/// Badge print result
@immutable
class BadgePrintResult {
  final String jobId;
  final bool success;
  final Timestamp timestamp;
  final String? errorMessage;
  final bool? magneticStripeVerified;
  final bool? rfidVerified;

  const BadgePrintResult({
    required this.jobId,
    required this.success,
    required this.timestamp,
    this.errorMessage,
    this.magneticStripeVerified,
    this.rfidVerified,
  });

  factory BadgePrintResult.fromJson(Map<String, dynamic> json) {
    return BadgePrintResult(
      jobId: json['job_id'] as String,
      success: json['success'] as bool,
      timestamp: Timestamp.fromJson(json['timestamp']),
      errorMessage: json['error_message'] as String?,
      magneticStripeVerified: json['magnetic_stripe_verified'] as bool?,
      rfidVerified: json['rfid_verified'] as bool?,
    );
  }

  @override
  String toString() =>
      'BadgePrintResult(jobId: $jobId, success: $success, error: $errorMessage)';
}

/// Badge template for common badge designs
@immutable
class BadgeTemplate {
  final String templateId;
  final String name;
  final String description;
  final CardSize cardSize;
  final CardOrientation orientation;
  final Map<String, String> fields; // Field name -> placeholder mapping

  const BadgeTemplate({
    required this.templateId,
    required this.name,
    required this.description,
    required this.cardSize,
    required this.orientation,
    required this.fields,
  });

  factory BadgeTemplate.fromJson(Map<String, dynamic> json) {
    return BadgeTemplate(
      templateId: json['template_id'] as String,
      name: json['name'] as String,
      description: json['description'] as String,
      cardSize: CardSize.fromString(json['card_size'] as String),
      orientation: CardOrientation.fromString(json['orientation'] as String),
      fields: Map<String, String>.from(json['fields'] as Map),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'template_id': templateId,
      'name': name,
      'description': description,
      'card_size': cardSize.name.toUpperCase(),
      'orientation': orientation.name,
      'fields': fields,
    };
  }

  @override
  String toString() => 'BadgeTemplate(name: $name, fields: ${fields.length})';
}
