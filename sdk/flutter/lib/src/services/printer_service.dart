import '../client.dart';
import '../exceptions.dart';
import '../models/printer_models.dart';
import '../logging/device_log_entry.dart';
import '../logging/device_interaction_logger.dart';

/// Service for interacting with ESC/POS thermal printers
///
/// Provides methods for printing receipts, labels, and other documents
/// on thermal printers.
///
/// Example:
/// ```dart
/// final printer = client.printer('printer-01');
///
/// // Check printer status
/// final status = await printer.getStatus();
/// if (status.isReady) {
///   // Build and print receipt
///   final receipt = ReceiptBuilder()
///     .header('My Store')
///     .divider()
///     .keyValue('Item', 'Price')
///     .keyValue('Coffee', '\$3.50')
///     .divider()
///     .keyValue('Total', '\$3.50')
///     .lineFeed(lines: 2)
///     .qrCode('https://example.com/receipt/123')
///     .cut()
///     .build();
///
///   final result = await printer.print(receipt);
///   print('Print job: ${result.jobId}');
/// }
/// ```
class PrinterService {
  final DeviceBridgeClient _client;
  final String deviceId;

  PrinterService(this._client, this.deviceId);

  /// Gets the current status of the printer
  ///
  /// Returns comprehensive printer status including paper level,
  /// cover state, temperature, and capabilities
  ///
  /// Throws [DeviceNotFoundException] if printer not found
  Future<PrinterStatus> getStatus() async {
    final response = await _client.get('/v1/devices/$deviceId/printer/status');
    return PrinterStatus.fromJson(response);
  }

  /// Prints a job to the printer
  ///
  /// [job] - The print job containing elements to print
  ///
  /// Returns the print result with job ID
  ///
  /// Throws [DeviceNotFoundException] if printer not found
  /// Throws [DeviceBusyException] if printer is busy
  /// Throws [DeviceErrorException] if printer has an error (paper out, cover open, etc.)
  ///
  /// Example:
  /// ```dart
  /// final job = PrintJob(
  ///   elements: [
  ///     TextElement(
  ///       text: 'Hello World',
  ///       alignment: TextAlignment.center,
  ///       size: TextSize.large,
  ///       bold: true,
  ///     ),
  ///     LineFeedElement(lines: 2),
  ///     QRCodeElement(data: 'https://example.com'),
  ///     CutElement(),
  ///   ],
  ///   options: PrintOptions(copies: 2, autoCut: true),
  /// );
  ///
  /// final result = await printer.print(job);
  /// ```
  Future<PrintResult> print(PrintJob job) async {
    String? errorMessage;
    PrintResult? result;

    try {
      final response = await _client.post(
        '/v1/devices/$deviceId/printer/print',
        body: job.toJson(),
      );

      result = PrintResult.fromJson(response);

      // Log the print interaction
      await _logPrintInteraction(
        result,
        _extractPrintContent(job),
        copies: job.options?.copies,
      );

      return result;
    } catch (e) {
      errorMessage = e.toString();

      // Log failed print interaction
      if (result != null) {
        await _logPrintInteraction(
          result,
          _extractPrintContent(job),
          copies: job.options?.copies,
          errorMessage: errorMessage,
        );
      }

      rethrow;
    }
  }

  /// Extract print content from job for logging
  String _extractPrintContent(PrintJob job) {
    final buffer = StringBuffer();
    for (final element in job.elements) {
      if (element is TextElement) {
        buffer.writeln(element.text);
      } else if (element is BarcodeElement) {
        buffer.writeln('[BARCODE: ${element.data}]');
      } else if (element is QRCodeElement) {
        buffer.writeln('[QR CODE: ${element.data}]');
      } else if (element is ImageElement) {
        buffer.writeln('[IMAGE]');
      } else if (element is LineFeedElement) {
        // Skip line feeds in content extraction
      } else if (element is CutElement) {
        buffer.writeln('[CUT]');
      } else if (element is OpenDrawerElement) {
        buffer.writeln('[OPEN DRAWER]');
      }
    }
    return buffer.toString();
  }

  /// Prints raw text with minimal formatting
  ///
  /// [text] - The text to print
  /// [options] - Optional print options
  ///
  /// Returns the print result
  ///
  /// This is a convenience method for simple text printing
  Future<PrintResult> printText(
    String text, {
    PrintOptions? options,
  }) async {
    final job = PrintJob(
      elements: [
        TextElement(text: text),
        LineFeedElement(lines: 2),
        CutElement(),
      ],
      options: options,
    );

    return print(job);
  }

  /// Log a print interaction
  Future<void> _logPrintInteraction(
    PrintResult result,
    String content, {
    int? copies,
    String? errorMessage,
  }) async {
    try {
      final logger = _client.deviceLogger;
      if (logger != null && logger.isEnabled) {
        final logEntry = PrinterLogEntry(
          id: DeviceInteractionLogger.generateId(),
          timestamp: DateTime.now(),
          deviceId: deviceId,
          jobId: result.jobId ?? 'unknown',
          content: content,
          copies: copies,
          success: errorMessage == null,
          errorMessage: errorMessage,
        );
        await logger.log(logEntry);
      }
    } catch (e) {
      // Silently fail logging to not disrupt main operations
    }
  }

  /// Prints a receipt using the builder pattern
  ///
  /// [builder] - Function that builds the receipt
  /// [options] - Optional print options
  ///
  /// Returns the print result
  ///
  /// Example:
  /// ```dart
  /// final result = await printer.printReceipt(
  ///   (b) => b
  ///     .header('My Store')
  ///     .text('123 Main St')
  ///     .text('City, ST 12345')
  ///     .divider()
  ///     .keyValue('Item', 'Price')
  ///     .keyValue('Coffee', '\$3.50')
  ///     .keyValue('Tax', '\$0.30')
  ///     .divider()
  ///     .keyValue('Total', '\$3.80')
  ///     .lineFeed(lines: 2)
  ///     .text('Thank you!', alignment: TextAlignment.center)
  ///     .cut(),
  ///   options: PrintOptions(copies: 1, openDrawer: true),
  /// );
  /// ```
  Future<PrintResult> printReceipt(
    ReceiptBuilder Function(ReceiptBuilder) builder, {
    PrintOptions? options,
  }) async {
    final receiptBuilder = ReceiptBuilder();
    builder(receiptBuilder);
    final job = receiptBuilder.build(options: options);
    return print(job);
  }

  /// Prints a barcode
  ///
  /// [data] - The barcode data
  /// [format] - The barcode format
  /// [height] - The barcode height in dots (default: 50)
  /// [showText] - Whether to show the text below barcode (default: true)
  /// [options] - Optional print options
  ///
  /// Returns the print result
  Future<PrintResult> printBarcode(
    String data,
    BarcodeFormat format, {
    int height = 50,
    bool showText = true,
    PrintOptions? options,
  }) async {
    final job = PrintJob(
      elements: [
        BarcodeElement(
          data: data,
          format: format,
          height: height,
          showText: showText,
          alignment: TextAlignment.center,
        ),
        LineFeedElement(lines: 2),
        CutElement(),
      ],
      options: options,
    );

    return print(job);
  }

  /// Prints a QR code
  ///
  /// [data] - The QR code data
  /// [size] - The QR code size (1-16, default: 6)
  /// [errorCorrection] - Error correction level (default: medium)
  /// [options] - Optional print options
  ///
  /// Returns the print result
  Future<PrintResult> printQRCode(
    String data, {
    int size = 6,
    QRErrorCorrection errorCorrection = QRErrorCorrection.medium,
    PrintOptions? options,
  }) async {
    final job = PrintJob(
      elements: [
        QRCodeElement(
          data: data,
          size: size,
          errorCorrection: errorCorrection,
          alignment: TextAlignment.center,
        ),
        LineFeedElement(lines: 2),
        CutElement(),
      ],
      options: options,
    );

    return print(job);
  }

  /// Prints an image
  ///
  /// [imageData] - Base64 encoded image data
  /// [width] - Optional width constraint
  /// [height] - Optional height constraint
  /// [options] - Optional print options
  ///
  /// Returns the print result
  ///
  /// Supported formats: PNG, JPG, BMP
  Future<PrintResult> printImage(
    String imageData, {
    int? width,
    int? height,
    PrintOptions? options,
  }) async {
    final job = PrintJob(
      elements: [
        ImageElement(
          imageData: imageData,
          width: width,
          height: height,
          alignment: TextAlignment.center,
        ),
        LineFeedElement(lines: 2),
        CutElement(),
      ],
      options: options,
    );

    return print(job);
  }

  /// Feeds paper by the specified number of lines
  ///
  /// [lines] - Number of lines to feed (default: 1)
  ///
  /// This is useful for aligning paper before cutting
  Future<PrintResult> feedPaper({int lines = 1}) async {
    final job = PrintJob(
      elements: [
        LineFeedElement(lines: lines),
      ],
    );

    return print(job);
  }

  /// Cuts the paper
  ///
  /// [partial] - Whether to perform a partial cut (default: false)
  ///
  /// Returns the print result
  ///
  /// Note: Requires printer with cutter capability
  Future<PrintResult> cutPaper({bool partial = false}) async {
    final job = PrintJob(
      elements: [
        CutElement(partial: partial),
      ],
    );

    return print(job);
  }

  /// Opens the cash drawer
  ///
  /// [pulseOnTime] - Pulse on time in milliseconds (default: 60)
  /// [pulseOffTime] - Pulse off time in milliseconds (default: 120)
  ///
  /// Returns the print result
  ///
  /// Note: Requires printer with cash drawer port
  Future<PrintResult> openCashDrawer({
    int pulseOnTime = 60,
    int pulseOffTime = 120,
  }) async {
    final job = PrintJob(
      elements: [
        OpenDrawerElement(
          pulseOnTime: pulseOnTime,
          pulseOffTime: pulseOffTime,
        ),
      ],
    );

    return print(job);
  }

  /// Resets the printer to default settings
  ///
  /// This clears any pending jobs and resets printer state
  ///
  /// Throws [DeviceNotFoundException] if printer not found
  Future<void> reset() async {
    await _client.post('/v1/devices/$deviceId/printer/reset');
  }

  /// Gets printer capabilities
  ///
  /// Returns detailed information about printer features and supported formats
  Future<PrinterCapabilities> getCapabilities() async {
    final response =
        await _client.get('/v1/devices/$deviceId/printer/capabilities');
    return PrinterCapabilities.fromJson(response);
  }

  /// Tests the printer by printing a test page
  ///
  /// Returns the print result
  ///
  /// The test page includes printer information, current settings,
  /// and various formatting examples
  Future<PrintResult> printTestPage() async {
    final response =
        await _client.post('/v1/devices/$deviceId/printer/test-print');
    return PrintResult.fromJson(response);
  }

  /// Gets the status of a print job
  ///
  /// [jobId] - The job ID from a previous print operation
  ///
  /// Returns job status information
  Future<Map<String, dynamic>> getJobStatus(String jobId) async {
    return await _client.get('/v1/devices/$deviceId/printer/jobs/$jobId');
  }

  /// Cancels a pending print job
  ///
  /// [jobId] - The job ID to cancel
  ///
  /// Throws [DeviceNotFoundException] if job not found
  Future<void> cancelJob(String jobId) async {
    await _client.post('/v1/devices/$deviceId/printer/jobs/$jobId/cancel');
  }
}
