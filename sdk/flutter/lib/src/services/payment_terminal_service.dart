import '../client.dart';
import '../exceptions.dart';
import '../models/payment_models.dart';

/// Service for interacting with payment terminals
///
/// Provides methods for processing payments, refunds, and managing
/// payment transactions with support for various payment methods.
///
/// Example:
/// ```dart
/// final terminal = client.paymentTerminal('terminal-01');
///
/// // Check terminal status
/// final status = await terminal.getStatus();
/// if (status.isReady) {
///   // Process a payment
///   final request = PaymentRequest(
///     amount: 2550, // $25.50 in cents
///     currency: 'USD',
///     type: TransactionType.sale,
///     orderId: 'ORDER-123',
///     description: 'Coffee and pastry',
///   );
///
///   final result = await terminal.processPayment(request);
///   if (result.payment.isApproved) {
///     print('Payment approved: ${result.payment.approvalCode}');
///   }
/// }
/// ```
class PaymentTerminalService {
  final DeviceBridgeClient _client;
  final String deviceId;

  PaymentTerminalService(this._client, this.deviceId);

  /// Gets the current status of the payment terminal
  ///
  /// Returns comprehensive terminal status including readiness,
  /// battery level, and capabilities
  ///
  /// Throws [DeviceNotFoundException] if terminal not found
  Future<PaymentTerminalStatus> getStatus() async {
    final response =
        await _client.get('/v1/devices/$deviceId/payment-terminal/status');
    return PaymentTerminalStatus.fromJson(response);
  }

  /// Processes a payment transaction
  ///
  /// [request] - The payment request with amount and transaction details
  ///
  /// Returns the complete transaction result including payment response,
  /// receipt data, and EMV information
  ///
  /// Throws [DeviceNotFoundException] if terminal not found
  /// Throws [DeviceBusyException] if terminal is processing another transaction
  /// Throws [DeviceTimeoutException] if payment times out
  /// Throws [DeviceErrorException] if payment is declined or encounters an error
  ///
  /// Example:
  /// ```dart
  /// final request = PaymentRequest(
  ///   amount: 1050, // $10.50
  ///   currency: 'USD',
  ///   type: TransactionType.sale,
  ///   allowedMethods: [
  ///     PaymentMethod.chip,
  ///     PaymentMethod.contactless,
  ///   ],
  ///   orderId: 'ORD-12345',
  ///   timeout: Duration(seconds: 60),
  /// );
  ///
  /// try {
  ///   final result = await terminal.processPayment(request);
  ///   if (result.payment.isApproved) {
  ///     print('Payment successful');
  ///     print('Transaction ID: ${result.payment.transactionId}');
  ///     print('Card: ${result.payment.cardData?.brand}');
  ///   } else {
  ///     print('Payment declined: ${result.payment.responseMessage}');
  ///   }
  /// } on DeviceTimeoutException {
  ///   print('Payment timed out');
  /// }
  /// ```
  Future<TransactionResult> processPayment(PaymentRequest request) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/payment-terminal/process-payment',
      body: request.toJson(),
    );

    return TransactionResult.fromJson(response);
  }

  /// Processes a refund transaction
  ///
  /// [request] - The refund request with original transaction details
  ///
  /// Returns the refund transaction result
  ///
  /// Example:
  /// ```dart
  /// // Full refund
  /// final refund = await terminal.processRefund(
  ///   RefundRequest(
  ///     originalTransactionId: 'TXN-12345',
  ///     reason: 'Customer returned item',
  ///   ),
  /// );
  ///
  /// // Partial refund
  /// final partialRefund = await terminal.processRefund(
  ///   RefundRequest(
  ///     originalTransactionId: 'TXN-12345',
  ///     amount: 500, // Refund $5.00 of original transaction
  ///     reason: 'Damaged item',
  ///   ),
  /// );
  /// ```
  Future<TransactionResult> processRefund(RefundRequest request) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/payment-terminal/process-refund',
      body: request.toJson(),
    );

    return TransactionResult.fromJson(response);
  }

  /// Voids a previous transaction
  ///
  /// [transactionId] - The transaction ID to void
  /// [reason] - Optional reason for void
  ///
  /// Returns the void transaction result
  ///
  /// Note: Void is typically only available for same-day transactions
  /// before batch settlement
  Future<TransactionResult> voidTransaction(
    String transactionId, {
    String? reason,
  }) async {
    final body = {
      'transaction_id': transactionId,
      if (reason != null) 'reason': reason,
    };

    final response = await _client.post(
      '/v1/devices/$deviceId/payment-terminal/void-transaction',
      body: body,
    );

    return TransactionResult.fromJson(response);
  }

  /// Adjusts tip on a previous transaction
  ///
  /// [request] - The tip adjustment request
  ///
  /// Returns the updated transaction result
  ///
  /// This is typically used in restaurant scenarios where tip is
  /// added after the initial authorization
  ///
  /// Example:
  /// ```dart
  /// final adjusted = await terminal.adjustTip(
  ///   TipAdjustmentRequest(
  ///     transactionId: 'TXN-12345',
  ///     tipAmount: 500, // Add $5.00 tip
  ///   ),
  /// );
  /// ```
  Future<TransactionResult> adjustTip(TipAdjustmentRequest request) async {
    final response = await _client.post(
      '/v1/devices/$deviceId/payment-terminal/adjust-tip',
      body: request.toJson(),
    );

    return TransactionResult.fromJson(response);
  }

  /// Cancels the current payment in progress
  ///
  /// This attempts to cancel a payment that is currently being processed
  ///
  /// Throws [DeviceNotFoundException] if no payment is in progress
  Future<void> cancelPayment() async {
    await _client.post('/v1/devices/$deviceId/payment-terminal/cancel-payment');
  }

  /// Gets transaction details
  ///
  /// [transactionId] - The transaction ID to retrieve
  ///
  /// Returns the transaction result
  ///
  /// Throws [DeviceNotFoundException] if transaction not found
  Future<TransactionResult> getTransaction(String transactionId) async {
    final response = await _client.get(
      '/v1/devices/$deviceId/payment-terminal/transactions/$transactionId',
    );

    return TransactionResult.fromJson(response);
  }

  /// Gets recent transactions
  ///
  /// [limit] - Maximum number of transactions to return (default: 100)
  /// [transactionType] - Optional filter for transaction type
  /// [startDate] - Optional start date for filtering
  /// [endDate] - Optional end date for filtering
  ///
  /// Returns a list of payment responses
  Future<List<PaymentResponse>> getRecentTransactions({
    int limit = 100,
    TransactionType? transactionType,
    DateTime? startDate,
    DateTime? endDate,
  }) async {
    final queryParams = <String, dynamic>{
      'limit': limit.toString(),
      if (transactionType != null) 'type': transactionType.apiValue,
      if (startDate != null) 'start_date': startDate.toIso8601String(),
      if (endDate != null) 'end_date': endDate.toIso8601String(),
    };

    final response = await _client.get(
      '/v1/devices/$deviceId/payment-terminal/transactions',
      queryParameters: queryParams,
    );

    final transactions = response['transactions'] as List<dynamic>;
    return transactions
        .map((e) => PaymentResponse.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// Closes the current batch and initiates settlement
  ///
  /// Returns batch summary with transaction totals
  ///
  /// This is typically done at end of day to settle all transactions
  /// with the payment processor
  ///
  /// Example:
  /// ```dart
  /// final summary = await terminal.closeBatch();
  /// print('Settled ${summary.transactionCount} transactions');
  /// print('Total amount: \$${summary.totalAmount / 100}');
  /// ```
  Future<BatchSummary> closeBatch() async {
    final response =
        await _client.post('/v1/devices/$deviceId/payment-terminal/close-batch');

    return BatchSummary.fromJson(response);
  }

  /// Gets the current batch summary
  ///
  /// Returns batch information without closing the batch
  Future<BatchSummary> getBatchSummary() async {
    final response =
        await _client.get('/v1/devices/$deviceId/payment-terminal/batch-summary');

    return BatchSummary.fromJson(response);
  }

  /// Gets terminal capabilities
  ///
  /// Returns detailed information about supported payment methods,
  /// card brands, and features
  Future<PaymentTerminalCapabilities> getCapabilities() async {
    final response =
        await _client.get('/v1/devices/$deviceId/payment-terminal/capabilities');

    return PaymentTerminalCapabilities.fromJson(response);
  }

  /// Displays a message on the terminal screen
  ///
  /// [message] - The message to display
  /// [duration] - Optional duration to show message
  ///
  /// This is useful for showing custom prompts to customers
  ///
  /// Example:
  /// ```dart
  /// await terminal.displayMessage(
  ///   'Please insert card',
  ///   duration: Duration(seconds: 5),
  /// );
  /// ```
  Future<void> displayMessage(
    String message, {
    Duration? duration,
  }) async {
    final body = {
      'message': message,
      if (duration != null) 'duration_seconds': duration.inSeconds,
    };

    await _client.post(
      '/v1/devices/$deviceId/payment-terminal/display-message',
      body: body,
    );
  }

  /// Clears the terminal display
  ///
  /// Returns the terminal to idle/ready screen
  Future<void> clearDisplay() async {
    await _client.post('/v1/devices/$deviceId/payment-terminal/clear-display');
  }

  /// Resets the payment terminal
  ///
  /// This clears any pending transactions and resets terminal state
  Future<void> reset() async {
    await _client.post('/v1/devices/$deviceId/payment-terminal/reset');
  }

  /// Tests terminal connectivity
  ///
  /// Returns true if terminal can communicate with payment processor
  Future<bool> testConnection() async {
    try {
      final response = await _client.post(
        '/v1/devices/$deviceId/payment-terminal/test-connection',
      );

      return response['success'] as bool? ?? false;
    } catch (e) {
      return false;
    }
  }

  /// Prints a receipt for a transaction
  ///
  /// [transactionId] - The transaction ID
  /// [receiptType] - Type of receipt ('customer' or 'merchant')
  ///
  /// Note: Requires terminal with built-in printer
  Future<void> printReceipt(
    String transactionId, {
    String receiptType = 'customer',
  }) async {
    final body = {
      'transaction_id': transactionId,
      'receipt_type': receiptType,
    };

    await _client.post(
      '/v1/devices/$deviceId/payment-terminal/print-receipt',
      body: body,
    );
  }

  /// Gets receipt data for a transaction
  ///
  /// [transactionId] - The transaction ID
  /// [format] - Receipt format ('text' or 'html')
  ///
  /// Returns the formatted receipt data
  Future<String> getReceiptData(
    String transactionId, {
    String format = 'text',
  }) async {
    final response = await _client.get(
      '/v1/devices/$deviceId/payment-terminal/transactions/$transactionId/receipt',
      queryParameters: {'format': format},
    );

    return response['receipt_data'] as String;
  }

  /// Enables or disables specific payment methods
  ///
  /// [methods] - List of payment methods to enable
  ///
  /// This restricts which payment methods are accepted
  Future<void> setAllowedPaymentMethods(List<PaymentMethod> methods) async {
    final body = {
      'methods': methods.map((e) => e.name.toUpperCase()).toList(),
    };

    await _client.post(
      '/v1/devices/$deviceId/payment-terminal/set-allowed-methods',
      body: body,
    );
  }

  /// Configures terminal settings
  ///
  /// [settings] - Map of setting keys to values
  ///
  /// Available settings vary by terminal type
  Future<void> configure(Map<String, dynamic> settings) async {
    await _client.post(
      '/v1/devices/$deviceId/payment-terminal/configure',
      body: settings,
    );
  }
}
