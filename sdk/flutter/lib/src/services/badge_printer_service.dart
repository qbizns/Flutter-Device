import '../client.dart';
import '../exceptions.dart';
import '../models/badge_printer_models.dart';

/// Service for interacting with badge/card printers
///
/// Provides methods for printing ID cards, badges, and access cards
/// with support for magnetic stripe and RFID encoding.
///
/// Example:
/// ```dart
/// final badgePrinter = client.badgePrinter('zebra-zxp-01');
///
/// // Check printer status
/// final status = await badgePrinter.getStatus();
/// if (status.isReady) {
///   // Create badge layout
///   final layout = CardLayout(
///     cardSize: CardSize.cr80,
///     orientation: CardOrientation.landscape,
///     elements: [
///       BadgeTextElement(
///         position: Position(x: 50, y: 100),
///         size: Size(width: 300, height: 40),
///         text: 'John Doe',
///         fontSize: 24,
///         bold: true,
///       ),
///       BadgeImageElement(
///         position: Position(x: 50, y: 50),
///         size: Size(width: 100, height: 100),
///         imageData: photoBase64,
///       ),
///     ],
///   );
///
///   // Print badge
///   final result = await badgePrinter.printBadge(
///     BadgePrintJob(
///       frontLayout: layout,
///       quality: PrintQuality.high,
///     ),
///   );
/// }
/// ```
class BadgePrinterService {
  final DeviceBridgeClient _client;
  final String deviceId;

  BadgePrinterService(this._client, this.deviceId);

  /// Gets the current status of the badge printer
  ///
  /// Returns comprehensive printer status including card/ribbon levels,
  /// cover state, and capabilities
  ///
  /// Throws [DeviceNotFoundException] if printer not found
  Future<BadgePrinterStatus> getStatus() async {
    final response =
        await _client.get('/v1/devices/$deviceId/badge-printer/status');
    return BadgePrinterStatus.fromJson(response);
  }

  /// Prints a badge card
  ///
  /// [job] - The badge print job containing layout and encoding data
  ///
  /// Returns the print result with job ID and verification status
  ///
  /// Throws [DeviceNotFoundException] if printer not found
  /// Throws [DeviceBusyException] if printer is busy
  /// Throws [DeviceErrorException] if printer has an error
  ///
  /// Example:
  /// ```dart
  /// final job = BadgePrintJob(
  ///   frontLayout: CardLayout(
  ///     cardSize: CardSize.cr80,
  ///     elements: [
  ///       BadgeTextElement(
  ///         position: Position(x: 100, y: 150),
  ///         size: Size(width: 400, height: 50),
  ///         text: 'Employee Name',
  ///         fontSize: 20,
  ///         bold: true,
  ///       ),
  ///     ],
  ///   ),
  ///   quality: PrintQuality.high,
  ///   applyOverlay: true,
  /// );
  ///
  /// final result = await badgePrinter.printBadge(job);
  /// ```
  Future<BadgePrintResult> printBadge(BadgePrintJob job) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/badge-printer/print',
      body: job.toJson(),
    );

    return BadgePrintResult.fromJson(response);
  }

  /// Prints a simple badge from a template
  ///
  /// [templateId] - The template identifier
  /// [fieldValues] - Map of field names to values
  /// [quality] - Print quality (default: normal)
  /// [copies] - Number of copies (default: 1)
  ///
  /// Returns the print result
  ///
  /// This is a convenience method for template-based printing
  ///
  /// Example:
  /// ```dart
  /// final result = await badgePrinter.printFromTemplate(
  ///   'employee-badge-v1',
  ///   {
  ///     'name': 'John Doe',
  ///     'employee_id': 'EMP-12345',
  ///     'department': 'Engineering',
  ///     'photo': photoBase64,
  ///   },
  ///   quality: PrintQuality.high,
  /// );
  /// ```
  Future<BadgePrintResult> printFromTemplate(
    String templateId,
    Map<String, String> fieldValues, {
    PrintQuality quality = PrintQuality.normal,
    int copies = 1,
  }) async {
    final body = {
      'template_id': templateId,
      'field_values': fieldValues,
      'quality': quality.name.toUpperCase(),
      'copies': copies,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/badge-printer/print-template',
      body: body,
    );

    return BadgePrintResult.fromJson(response);
  }

  /// Encodes a magnetic stripe on a card
  ///
  /// [magneticData] - The magnetic stripe data to encode
  ///
  /// Returns the encoding result with verification status
  ///
  /// Note: Requires printer with magnetic stripe encoder
  ///
  /// Example:
  /// ```dart
  /// final result = await badgePrinter.encodeMagneticStripe(
  ///   MagneticStripeData(
  ///     track1: '%B1234567890123456^DOE/JOHN^2512101123456789?',
  ///     track2: ';1234567890123456=25121011234567890?',
  ///     verify: true,
  ///   ),
  /// );
  /// ```
  Future<BadgePrintResult> encodeMagneticStripe(
    MagneticStripeData magneticData,
  ) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/badge-printer/encode-magnetic',
      body: magneticData.toJson(),
    );

    return BadgePrintResult.fromJson(response);
  }

  /// Encodes RFID data on a card
  ///
  /// [rfidData] - The RFID data to encode
  ///
  /// Returns the encoding result with verification status
  ///
  /// Note: Requires printer with RFID encoder
  ///
  /// Example:
  /// ```dart
  /// final result = await badgePrinter.encodeRFID(
  ///   RFIDEncodingData(
  ///     uid: '01020304',
  ///     blockData: {
  ///       4: 'Employee12345',
  ///       5: 'Engineering',
  ///     },
  ///     verify: true,
  ///   ),
  /// );
  /// ```
  Future<BadgePrintResult> encodeRFID(RFIDEncodingData rfidData) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/badge-printer/encode-rfid',
      body: rfidData.toJson(),
    );

    return BadgePrintResult.fromJson(response);
  }

  /// Prints a badge with both magnetic stripe and RFID encoding
  ///
  /// [job] - The complete badge print job
  ///
  /// This is a convenience method that handles printing and encoding in one operation
  Future<BadgePrintResult> printAndEncode(BadgePrintJob job) async {
    if (job.magneticData == null && job.rfidData == null) {
      throw ValidationException(
        'Either magneticData or rfidData must be provided for encoding',
      );
    }

    return printBadge(job);
  }

  /// Gets printer capabilities
  ///
  /// Returns detailed information about printer features,
  /// supported card sizes, and encoding capabilities
  Future<BadgePrinterCapabilities> getCapabilities() async {
    final response =
        await _client.get('/v1/devices/$deviceId/badge-printer/capabilities');
    return BadgePrinterCapabilities.fromJson(response);
  }

  /// Ejects a card from the printer
  ///
  /// This is useful for retrieving a printed card or clearing a jam
  Future<void> ejectCard() async {
    await _client.post('/v1/devices/$deviceId/badge-printer/eject');
  }

  /// Cleans the printer
  ///
  /// [useCleaningCard] - Whether to use a cleaning card (default: true)
  ///
  /// Performs a cleaning cycle to maintain print quality
  Future<void> clean({bool useCleaningCard = true}) async {
    await _client.post(
      '/v1/devices/$deviceId/badge-printer/clean',
      body: {'use_cleaning_card': useCleaningCard},
    );
  }

  /// Resets the printer to default settings
  ///
  /// This clears any pending jobs and resets printer state
  Future<void> reset() async {
    await _client.post('/v1/devices/$deviceId/badge-printer/reset');
  }

  /// Gets the status of a print job
  ///
  /// [jobId] - The job ID from a previous print operation
  ///
  /// Returns job status information
  Future<Map<String, dynamic>> getJobStatus(String jobId) async {
    return await _client
        .get('/v1/devices/$deviceId/badge-printer/jobs/$jobId');
  }

  /// Cancels a pending print job
  ///
  /// [jobId] - The job ID to cancel
  ///
  /// Throws [DeviceNotFoundException] if job not found
  Future<void> cancelJob(String jobId) async {
    await _client
        .post('/v1/devices/$deviceId/badge-printer/jobs/$jobId/cancel');
  }

  /// Lists available badge templates
  ///
  /// Returns a list of template definitions
  Future<List<BadgeTemplate>> listTemplates() async {
    final response =
        await _client.get('/v1/devices/$deviceId/badge-printer/templates');

    final templates = response['templates'] as List<dynamic>;
    return templates
        .map((e) => BadgeTemplate.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// Gets a specific badge template
  ///
  /// [templateId] - The template identifier
  ///
  /// Returns the template definition
  ///
  /// Throws [DeviceNotFoundException] if template not found
  Future<BadgeTemplate> getTemplate(String templateId) async {
    final response = await _client
        .get('/v1/devices/$deviceId/badge-printer/templates/$templateId');
    return BadgeTemplate.fromJson(response);
  }

  /// Creates or updates a badge template
  ///
  /// [template] - The template definition
  ///
  /// Returns the created/updated template
  Future<BadgeTemplate> saveTemplate(BadgeTemplate template) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/badge-printer/templates',
      body: template.toJson(),
    );
    return BadgeTemplate.fromJson(response);
  }

  /// Deletes a badge template
  ///
  /// [templateId] - The template identifier to delete
  Future<void> deleteTemplate(String templateId) async {
    await _client.delete(
      '/v1/devices/$deviceId/badge-printer/templates/$templateId',
    );
  }

  /// Tests the printer by printing a test card
  ///
  /// Returns the print result
  ///
  /// The test card includes color gradients, alignment marks,
  /// and sample text for quality verification
  Future<BadgePrintResult> printTestCard() async {
    final response =
        await _client.post('/v1/devices/$deviceId/badge-printer/test-print');
    return BadgePrintResult.fromJson(response);
  }

  /// Gets ribbon information
  ///
  /// Returns details about the current ribbon including type,
  /// remaining prints, and compatibility
  Future<Map<String, dynamic>> getRibbonInfo() async {
    return await _client
        .get('/v1/devices/$deviceId/badge-printer/ribbon-info');
  }

  /// Validates a card layout before printing
  ///
  /// [layout] - The card layout to validate
  ///
  /// Returns validation results with any errors or warnings
  ///
  /// This is useful for checking layouts before committing to print
  Future<Map<String, dynamic>> validateLayout(CardLayout layout) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/badge-printer/validate-layout',
      body: layout.toJson(),
    );
    return response;
  }
}
