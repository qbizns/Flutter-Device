import 'package:meta/meta.dart';
import 'common_models.dart';

/// Paper width enumeration
enum PaperWidth {
  unknown,
  mm58,
  mm80,
  mm112;

  static PaperWidth fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('PAPER_WIDTH_', '');
    final map = {
      '58MM': mm58,
      '80MM': mm80,
      '112MM': mm112,
    };
    return map[normalized] ?? PaperWidth.unknown;
  }

  int get widthMm {
    switch (this) {
      case mm58:
        return 58;
      case mm80:
        return 80;
      case mm112:
        return 112;
      default:
        return 0;
    }
  }
}

/// Text alignment enumeration
enum TextAlignment {
  left,
  center,
  right;

  static TextAlignment fromString(String value) {
    return TextAlignment.values.firstWhere(
      (e) => e.name.toUpperCase() == value.toUpperCase(),
      orElse: () => TextAlignment.left,
    );
  }
}

/// Text size enumeration
enum TextSize {
  small,
  normal,
  large,
  extraLarge;

  static TextSize fromString(String value) {
    return TextSize.values.firstWhere(
      (e) => e.name.toUpperCase() == value.toUpperCase(),
      orElse: () => TextSize.normal,
    );
  }
}

/// Barcode format enumeration
enum BarcodeFormat {
  unknown,
  upca,
  upce,
  ean13,
  ean8,
  code39,
  code93,
  code128,
  itf,
  codabar;

  static BarcodeFormat fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('BARCODE_FORMAT_', '');
    return BarcodeFormat.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => BarcodeFormat.unknown,
    );
  }
}

/// QR code error correction level
enum QRErrorCorrection {
  low,
  medium,
  quartile,
  high;

  static QRErrorCorrection fromString(String value) {
    return QRErrorCorrection.values.firstWhere(
      (e) => e.name.toUpperCase() == value.toUpperCase(),
      orElse: () => QRErrorCorrection.medium,
    );
  }
}

/// Printer capabilities
@immutable
class PrinterCapabilities {
  final PaperWidth paperWidth;
  final bool hasColorSupport;
  final bool hasCutter;
  final bool hasCashDrawer;
  final int maxPrintSpeed;
  final List<BarcodeFormat> supportedBarcodes;
  final bool supportsQRCodes;
  final bool supportsImages;

  const PrinterCapabilities({
    required this.paperWidth,
    required this.hasColorSupport,
    required this.hasCutter,
    required this.hasCashDrawer,
    required this.maxPrintSpeed,
    required this.supportedBarcodes,
    required this.supportsQRCodes,
    required this.supportsImages,
  });

  factory PrinterCapabilities.fromJson(Map<String, dynamic> json) {
    return PrinterCapabilities(
      paperWidth: PaperWidth.fromString(json['paper_width'] as String),
      hasColorSupport: json['has_color_support'] as bool? ?? false,
      hasCutter: json['has_cutter'] as bool? ?? false,
      hasCashDrawer: json['has_cash_drawer'] as bool? ?? false,
      maxPrintSpeed: json['max_print_speed'] as int? ?? 0,
      supportedBarcodes: (json['supported_barcodes'] as List<dynamic>?)
              ?.map((e) => BarcodeFormat.fromString(e as String))
              .toList() ??
          [],
      supportsQRCodes: json['supports_qr_codes'] as bool? ?? false,
      supportsImages: json['supports_images'] as bool? ?? false,
    );
  }

  @override
  String toString() =>
      'PrinterCapabilities(paperWidth: $paperWidth, hasColorSupport: $hasColorSupport, hasCutter: $hasCutter)';
}

/// Printer status
@immutable
class PrinterStatus {
  final String deviceId;
  final String name;
  final DeviceStatus status;
  final bool paperPresent;
  final bool coverOpen;
  final int? paperPercentRemaining;
  final int? temperature;
  final PrinterCapabilities? capabilities;
  final String? errorMessage;

  const PrinterStatus({
    required this.deviceId,
    required this.name,
    required this.status,
    required this.paperPresent,
    required this.coverOpen,
    this.paperPercentRemaining,
    this.temperature,
    this.capabilities,
    this.errorMessage,
  });

  factory PrinterStatus.fromJson(Map<String, dynamic> json) {
    return PrinterStatus(
      deviceId: json['device_id'] as String,
      name: json['name'] as String,
      status: DeviceStatus.fromString(json['status'] as String),
      paperPresent: json['paper_present'] as bool,
      coverOpen: json['cover_open'] as bool,
      paperPercentRemaining: json['paper_percent_remaining'] as int?,
      temperature: json['temperature'] as int?,
      capabilities: json['capabilities'] != null
          ? PrinterCapabilities.fromJson(
              json['capabilities'] as Map<String, dynamic>)
          : null,
      errorMessage: json['error_message'] as String?,
    );
  }

  bool get isReady => status == DeviceStatus.ready && paperPresent && !coverOpen;
  bool get hasError => status == DeviceStatus.error || errorMessage != null;
  bool get isPaperLow =>
      paperPercentRemaining != null && paperPercentRemaining! < 20;

  @override
  String toString() =>
      'PrinterStatus(deviceId: $deviceId, status: $status, paperPresent: $paperPresent, coverOpen: $coverOpen)';
}

/// Base class for printable elements
@immutable
abstract class PrintElement {
  const PrintElement();

  Map<String, dynamic> toJson();
}

/// Text element for printing
@immutable
class TextElement extends PrintElement {
  final String text;
  final TextAlignment alignment;
  final TextSize size;
  final bool bold;
  final bool underline;
  final bool invert;

  const TextElement({
    required this.text,
    this.alignment = TextAlignment.left,
    this.size = TextSize.normal,
    this.bold = false,
    this.underline = false,
    this.invert = false,
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'text',
      'text': text,
      'alignment': alignment.name,
      'size': size.name,
      'bold': bold,
      'underline': underline,
      'invert': invert,
    };
  }

  @override
  String toString() => 'TextElement(text: $text, size: $size, bold: $bold)';
}

/// Line feed element
@immutable
class LineFeedElement extends PrintElement {
  final int lines;

  const LineFeedElement({this.lines = 1});

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'line_feed',
      'lines': lines,
    };
  }

  @override
  String toString() => 'LineFeedElement(lines: $lines)';
}

/// Barcode element for printing
@immutable
class BarcodeElement extends PrintElement {
  final String data;
  final BarcodeFormat format;
  final int height;
  final bool showText;
  final TextAlignment alignment;

  const BarcodeElement({
    required this.data,
    required this.format,
    this.height = 50,
    this.showText = true,
    this.alignment = TextAlignment.center,
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'barcode',
      'data': data,
      'format': format.name.toUpperCase(),
      'height': height,
      'show_text': showText,
      'alignment': alignment.name,
    };
  }

  @override
  String toString() => 'BarcodeElement(data: $data, format: $format)';
}

/// QR code element for printing
@immutable
class QRCodeElement extends PrintElement {
  final String data;
  final int size;
  final QRErrorCorrection errorCorrection;
  final TextAlignment alignment;

  const QRCodeElement({
    required this.data,
    this.size = 6,
    this.errorCorrection = QRErrorCorrection.medium,
    this.alignment = TextAlignment.center,
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'qr_code',
      'data': data,
      'size': size,
      'error_correction': errorCorrection.name.toUpperCase(),
      'alignment': alignment.name,
    };
  }

  @override
  String toString() => 'QRCodeElement(data: $data, size: $size)';
}

/// Image element for printing
@immutable
class ImageElement extends PrintElement {
  final String imageData; // Base64 encoded image
  final int? width;
  final int? height;
  final TextAlignment alignment;

  const ImageElement({
    required this.imageData,
    this.width,
    this.height,
    this.alignment = TextAlignment.center,
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'image',
      'image_data': imageData,
      if (width != null) 'width': width,
      if (height != null) 'height': height,
      'alignment': alignment.name,
    };
  }

  @override
  String toString() => 'ImageElement(width: $width, height: $height)';
}

/// Cut paper element
@immutable
class CutElement extends PrintElement {
  final bool partial;

  const CutElement({this.partial = false});

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'cut',
      'partial': partial,
    };
  }

  @override
  String toString() => 'CutElement(partial: $partial)';
}

/// Open cash drawer element
@immutable
class OpenDrawerElement extends PrintElement {
  final int pulseOnTime;
  final int pulseOffTime;

  const OpenDrawerElement({
    this.pulseOnTime = 60,
    this.pulseOffTime = 120,
  });

  @override
  Map<String, dynamic> toJson() {
    return {
      'type': 'open_drawer',
      'pulse_on_time': pulseOnTime,
      'pulse_off_time': pulseOffTime,
    };
  }

  @override
  String toString() => 'OpenDrawerElement()';
}

/// Print options
@immutable
class PrintOptions {
  final int copies;
  final bool autoCut;
  final bool openDrawer;
  final int? feedLinesAfterCut;

  const PrintOptions({
    this.copies = 1,
    this.autoCut = true,
    this.openDrawer = false,
    this.feedLinesAfterCut,
  });

  Map<String, dynamic> toJson() {
    return {
      'copies': copies,
      'auto_cut': autoCut,
      'open_drawer': openDrawer,
      if (feedLinesAfterCut != null) 'feed_lines_after_cut': feedLinesAfterCut,
    };
  }

  @override
  String toString() =>
      'PrintOptions(copies: $copies, autoCut: $autoCut, openDrawer: $openDrawer)';
}

/// Print job
@immutable
class PrintJob {
  final List<PrintElement> elements;
  final PrintOptions? options;

  const PrintJob({
    required this.elements,
    this.options,
  });

  Map<String, dynamic> toJson() {
    return {
      'elements': elements.map((e) => e.toJson()).toList(),
      if (options != null) 'options': options!.toJson(),
    };
  }

  @override
  String toString() => 'PrintJob(elements: ${elements.length})';
}

/// Print result
@immutable
class PrintResult {
  final String jobId;
  final bool success;
  final Timestamp timestamp;
  final String? errorMessage;

  const PrintResult({
    required this.jobId,
    required this.success,
    required this.timestamp,
    this.errorMessage,
  });

  factory PrintResult.fromJson(Map<String, dynamic> json) {
    return PrintResult(
      jobId: json['job_id'] as String,
      success: json['success'] as bool,
      timestamp: Timestamp.fromJson(json['timestamp']),
      errorMessage: json['error_message'] as String?,
    );
  }

  @override
  String toString() =>
      'PrintResult(jobId: $jobId, success: $success, error: $errorMessage)';
}

/// Receipt builder helper class
class ReceiptBuilder {
  final List<PrintElement> _elements = [];

  /// Adds text to the receipt
  ReceiptBuilder text(
    String text, {
    TextAlignment alignment = TextAlignment.left,
    TextSize size = TextSize.normal,
    bool bold = false,
    bool underline = false,
    bool invert = false,
  }) {
    _elements.add(TextElement(
      text: text,
      alignment: alignment,
      size: size,
      bold: bold,
      underline: underline,
      invert: invert,
    ));
    return this;
  }

  /// Adds a line feed
  ReceiptBuilder lineFeed({int lines = 1}) {
    _elements.add(LineFeedElement(lines: lines));
    return this;
  }

  /// Adds a barcode
  ReceiptBuilder barcode(
    String data,
    BarcodeFormat format, {
    int height = 50,
    bool showText = true,
    TextAlignment alignment = TextAlignment.center,
  }) {
    _elements.add(BarcodeElement(
      data: data,
      format: format,
      height: height,
      showText: showText,
      alignment: alignment,
    ));
    return this;
  }

  /// Adds a QR code
  ReceiptBuilder qrCode(
    String data, {
    int size = 6,
    QRErrorCorrection errorCorrection = QRErrorCorrection.medium,
    TextAlignment alignment = TextAlignment.center,
  }) {
    _elements.add(QRCodeElement(
      data: data,
      size: size,
      errorCorrection: errorCorrection,
      alignment: alignment,
    ));
    return this;
  }

  /// Adds an image
  ReceiptBuilder image(
    String imageData, {
    int? width,
    int? height,
    TextAlignment alignment = TextAlignment.center,
  }) {
    _elements.add(ImageElement(
      imageData: imageData,
      width: width,
      height: height,
      alignment: alignment,
    ));
    return this;
  }

  /// Adds a cut command
  ReceiptBuilder cut({bool partial = false}) {
    _elements.add(CutElement(partial: partial));
    return this;
  }

  /// Adds an open drawer command
  ReceiptBuilder openDrawer({int pulseOnTime = 60, int pulseOffTime = 120}) {
    _elements.add(OpenDrawerElement(
      pulseOnTime: pulseOnTime,
      pulseOffTime: pulseOffTime,
    ));
    return this;
  }

  /// Adds a divider line
  ReceiptBuilder divider({String char = '-'}) {
    _elements.add(TextElement(
      text: char * 42, // 42 chars for 80mm paper
      alignment: TextAlignment.center,
    ));
    return this;
  }

  /// Adds a header with title
  ReceiptBuilder header(String title) {
    return text(
      title,
      alignment: TextAlignment.center,
      size: TextSize.large,
      bold: true,
    ).lineFeed();
  }

  /// Adds a key-value pair
  ReceiptBuilder keyValue(String key, String value) {
    final spaces = 42 - key.length - value.length;
    return text('$key${' ' * (spaces > 0 ? spaces : 1)}$value');
  }

  /// Builds the print job
  PrintJob build({PrintOptions? options}) {
    return PrintJob(
      elements: List.unmodifiable(_elements),
      options: options,
    );
  }
}
