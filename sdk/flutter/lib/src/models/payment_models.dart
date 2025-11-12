import 'package:meta/meta.dart';
import 'common_models.dart';

/// Payment method enumeration
enum PaymentMethod {
  unknown,
  swipe,
  chip,
  contactless,
  manual,
  qr;

  static PaymentMethod fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('PAYMENT_METHOD_', '');
    return PaymentMethod.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => PaymentMethod.unknown,
    );
  }
}

/// Transaction type enumeration
enum TransactionType {
  unknown,
  sale,
  refund,
  void_,
  authorization,
  capture,
  preauth;

  static TransactionType fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('TRANSACTION_TYPE_', '');
    if (normalized == 'VOID') return void_;
    return TransactionType.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized || (e.name == 'void_' && normalized == 'VOID'),
      orElse: () => TransactionType.unknown,
    );
  }

  String get apiValue => this == void_ ? 'VOID' : name.toUpperCase();
}

/// Payment card type enumeration
enum PaymentCardType {
  unknown,
  credit,
  debit,
  prepaid,
  gift;

  static PaymentCardType fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('CARD_TYPE_', '');
    return PaymentCardType.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => PaymentCardType.unknown,
    );
  }
}

/// Card brand enumeration
enum CardBrand {
  unknown,
  visa,
  mastercard,
  amex,
  discover,
  jcb,
  unionpay,
  dinersClub;

  static CardBrand fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('BRAND_', '');
    final map = {
      'VISA': visa,
      'MASTERCARD': mastercard,
      'AMEX': amex,
      'DISCOVER': discover,
      'JCB': jcb,
      'UNIONPAY': unionpay,
      'DINERS_CLUB': dinersClub,
      'DINERSCLUB': dinersClub,
    };
    return map[normalized] ?? CardBrand.unknown;
  }
}

/// Payment status enumeration
enum PaymentStatus {
  unknown,
  pending,
  processing,
  approved,
  declined,
  cancelled,
  error;

  static PaymentStatus fromString(String value) {
    final normalized = value.toUpperCase().replaceAll('PAYMENT_STATUS_', '');
    return PaymentStatus.values.firstWhere(
      (e) => e.name.toUpperCase() == normalized,
      orElse: () => PaymentStatus.unknown,
    );
  }
}

/// Payment request
@immutable
class PaymentRequest {
  final int amount; // Amount in cents
  final String currency;
  final TransactionType type;
  final List<PaymentMethod>? allowedMethods;
  final String? orderId;
  final String? description;
  final Map<String, dynamic>? metadata;
  final Duration? timeout;

  const PaymentRequest({
    required this.amount,
    this.currency = 'USD',
    this.type = TransactionType.sale,
    this.allowedMethods,
    this.orderId,
    this.description,
    this.metadata,
    this.timeout,
  });

  Map<String, dynamic> toJson() {
    return {
      'amount': amount,
      'currency': currency,
      'type': type.apiValue,
      if (allowedMethods != null)
        'allowed_methods': allowedMethods!.map((e) => e.name.toUpperCase()).toList(),
      if (orderId != null) 'order_id': orderId,
      if (description != null) 'description': description,
      if (metadata != null) 'metadata': metadata,
      if (timeout != null) 'timeout_seconds': timeout!.inSeconds,
    };
  }

  @override
  String toString() =>
      'PaymentRequest(amount: \$${amount / 100}, type: $type, currency: $currency)';
}

/// Card data from payment
@immutable
class CardData {
  final String? maskedPan;
  final CardBrand brand;
  final PaymentCardType cardType;
  final String? cardholderName;
  final String? expiryMonth;
  final String? expiryYear;
  final String? last4;

  const CardData({
    this.maskedPan,
    required this.brand,
    required this.cardType,
    this.cardholderName,
    this.expiryMonth,
    this.expiryYear,
    this.last4,
  });

  factory CardData.fromJson(Map<String, dynamic> json) {
    return CardData(
      maskedPan: json['masked_pan'] as String?,
      brand: CardBrand.fromString(json['brand'] as String? ?? 'UNKNOWN'),
      cardType: PaymentCardType.fromString(json['card_type'] as String? ?? 'UNKNOWN'),
      cardholderName: json['cardholder_name'] as String?,
      expiryMonth: json['expiry_month'] as String?,
      expiryYear: json['expiry_year'] as String?,
      last4: json['last4'] as String?,
    );
  }

  @override
  String toString() => 'CardData(brand: $brand, last4: $last4, type: $cardType)';
}

/// Payment response
@immutable
class PaymentResponse {
  final String transactionId;
  final PaymentStatus status;
  final int amount;
  final String currency;
  final TransactionType type;
  final PaymentMethod method;
  final CardData? cardData;
  final String? approvalCode;
  final String? responseMessage;
  final Timestamp timestamp;
  final Map<String, dynamic>? metadata;

  const PaymentResponse({
    required this.transactionId,
    required this.status,
    required this.amount,
    required this.currency,
    required this.type,
    required this.method,
    this.cardData,
    this.approvalCode,
    this.responseMessage,
    required this.timestamp,
    this.metadata,
  });

  factory PaymentResponse.fromJson(Map<String, dynamic> json) {
    return PaymentResponse(
      transactionId: json['transaction_id'] as String,
      status: PaymentStatus.fromString(json['status'] as String),
      amount: json['amount'] as int,
      currency: json['currency'] as String? ?? 'USD',
      type: TransactionType.fromString(json['type'] as String),
      method: PaymentMethod.fromString(json['method'] as String),
      cardData: json['card_data'] != null
          ? CardData.fromJson(json['card_data'] as Map<String, dynamic>)
          : null,
      approvalCode: json['approval_code'] as String?,
      responseMessage: json['response_message'] as String?,
      timestamp: Timestamp.fromJson(json['timestamp']),
      metadata: json['metadata'] as Map<String, dynamic>?,
    );
  }

  bool get isApproved => status == PaymentStatus.approved;
  bool get isDeclined => status == PaymentStatus.declined;
  bool get isPending => status == PaymentStatus.pending;

  @override
  String toString() =>
      'PaymentResponse(transactionId: $transactionId, status: $status, amount: \$${amount / 100})';
}

/// Transaction result (more detailed than PaymentResponse)
@immutable
class TransactionResult {
  final PaymentResponse payment;
  final String? receiptData;
  final String? signatureRequired;
  final Map<String, dynamic>? emvData;

  const TransactionResult({
    required this.payment,
    this.receiptData,
    this.signatureRequired,
    this.emvData,
  });

  factory TransactionResult.fromJson(Map<String, dynamic> json) {
    return TransactionResult(
      payment: PaymentResponse.fromJson(json['payment'] as Map<String, dynamic>),
      receiptData: json['receipt_data'] as String?,
      signatureRequired: json['signature_required'] as String?,
      emvData: json['emv_data'] as Map<String, dynamic>?,
    );
  }

  @override
  String toString() => 'TransactionResult(payment: $payment)';
}

/// Payment terminal capabilities
@immutable
class PaymentTerminalCapabilities {
  final List<PaymentMethod> supportedMethods;
  final List<CardBrand> supportedCardBrands;
  final bool supportsSwipe;
  final bool supportsChip;
  final bool supportsContactless;
  final bool supportsManualEntry;
  final bool supportsQRPayments;
  final bool supportsPinEntry;
  final bool supportsSignature;
  final bool supportsTipping;
  final bool supportsRefunds;

  const PaymentTerminalCapabilities({
    required this.supportedMethods,
    required this.supportedCardBrands,
    required this.supportsSwipe,
    required this.supportsChip,
    required this.supportsContactless,
    required this.supportsManualEntry,
    required this.supportsQRPayments,
    required this.supportsPinEntry,
    required this.supportsSignature,
    required this.supportsTipping,
    required this.supportsRefunds,
  });

  factory PaymentTerminalCapabilities.fromJson(Map<String, dynamic> json) {
    return PaymentTerminalCapabilities(
      supportedMethods: (json['supported_methods'] as List<dynamic>?)
              ?.map((e) => PaymentMethod.fromString(e as String))
              .toList() ??
          [],
      supportedCardBrands: (json['supported_card_brands'] as List<dynamic>?)
              ?.map((e) => CardBrand.fromString(e as String))
              .toList() ??
          [],
      supportsSwipe: json['supports_swipe'] as bool? ?? false,
      supportsChip: json['supports_chip'] as bool? ?? false,
      supportsContactless: json['supports_contactless'] as bool? ?? false,
      supportsManualEntry: json['supports_manual_entry'] as bool? ?? false,
      supportsQRPayments: json['supports_qr_payments'] as bool? ?? false,
      supportsPinEntry: json['supports_pin_entry'] as bool? ?? false,
      supportsSignature: json['supports_signature'] as bool? ?? false,
      supportsTipping: json['supports_tipping'] as bool? ?? false,
      supportsRefunds: json['supports_refunds'] as bool? ?? false,
    );
  }

  @override
  String toString() =>
      'PaymentTerminalCapabilities(swipe: $supportsSwipe, chip: $supportsChip, contactless: $supportsContactless)';
}

/// Payment terminal status
@immutable
class PaymentTerminalStatus {
  final String deviceId;
  final String name;
  final DeviceStatus status;
  final bool cardPresent;
  final bool busy;
  final int? batteryLevel;
  final PaymentTerminalCapabilities? capabilities;
  final String? errorMessage;

  const PaymentTerminalStatus({
    required this.deviceId,
    required this.name,
    required this.status,
    required this.cardPresent,
    required this.busy,
    this.batteryLevel,
    this.capabilities,
    this.errorMessage,
  });

  factory PaymentTerminalStatus.fromJson(Map<String, dynamic> json) {
    return PaymentTerminalStatus(
      deviceId: json['device_id'] as String,
      name: json['name'] as String,
      status: DeviceStatus.fromString(json['status'] as String),
      cardPresent: json['card_present'] as bool? ?? false,
      busy: json['busy'] as bool? ?? false,
      batteryLevel: json['battery_level'] as int?,
      capabilities: json['capabilities'] != null
          ? PaymentTerminalCapabilities.fromJson(
              json['capabilities'] as Map<String, dynamic>)
          : null,
      errorMessage: json['error_message'] as String?,
    );
  }

  bool get isReady => status == DeviceStatus.ready && !busy;
  bool get hasError => status == DeviceStatus.error || errorMessage != null;
  bool get isBatteryLow => batteryLevel != null && batteryLevel! < 20;

  @override
  String toString() =>
      'PaymentTerminalStatus(deviceId: $deviceId, status: $status, busy: $busy)';
}

/// Refund request
@immutable
class RefundRequest {
  final String originalTransactionId;
  final int? amount; // If null, refund full amount
  final String? reason;
  final Map<String, dynamic>? metadata;

  const RefundRequest({
    required this.originalTransactionId,
    this.amount,
    this.reason,
    this.metadata,
  });

  Map<String, dynamic> toJson() {
    return {
      'original_transaction_id': originalTransactionId,
      if (amount != null) 'amount': amount,
      if (reason != null) 'reason': reason,
      if (metadata != null) 'metadata': metadata,
    };
  }

  @override
  String toString() => 'RefundRequest(transactionId: $originalTransactionId)';
}

/// Tip adjustment request
@immutable
class TipAdjustmentRequest {
  final String transactionId;
  final int tipAmount; // Amount in cents
  final Map<String, dynamic>? metadata;

  const TipAdjustmentRequest({
    required this.transactionId,
    required this.tipAmount,
    this.metadata,
  });

  Map<String, dynamic> toJson() {
    return {
      'transaction_id': transactionId,
      'tip_amount': tipAmount,
      if (metadata != null) 'metadata': metadata,
    };
  }

  @override
  String toString() =>
      'TipAdjustmentRequest(transactionId: $transactionId, tip: \$${tipAmount / 100})';
}

/// Batch summary for end-of-day settlement
@immutable
class BatchSummary {
  final String batchId;
  final int transactionCount;
  final int totalAmount;
  final int refundCount;
  final int refundAmount;
  final Timestamp openedAt;
  final Timestamp? closedAt;
  final Map<CardBrand, int>? amountsByBrand;

  const BatchSummary({
    required this.batchId,
    required this.transactionCount,
    required this.totalAmount,
    required this.refundCount,
    required this.refundAmount,
    required this.openedAt,
    this.closedAt,
    this.amountsByBrand,
  });

  factory BatchSummary.fromJson(Map<String, dynamic> json) {
    return BatchSummary(
      batchId: json['batch_id'] as String,
      transactionCount: json['transaction_count'] as int,
      totalAmount: json['total_amount'] as int,
      refundCount: json['refund_count'] as int? ?? 0,
      refundAmount: json['refund_amount'] as int? ?? 0,
      openedAt: Timestamp.fromJson(json['opened_at']),
      closedAt: json['closed_at'] != null
          ? Timestamp.fromJson(json['closed_at'])
          : null,
      amountsByBrand: json['amounts_by_brand'] != null
          ? (json['amounts_by_brand'] as Map<String, dynamic>).map(
              (k, v) => MapEntry(
                CardBrand.fromString(k),
                v as int,
              ),
            )
          : null,
    );
  }

  bool get isClosed => closedAt != null;

  @override
  String toString() =>
      'BatchSummary(batchId: $batchId, transactions: $transactionCount, total: \$${totalAmount / 100})';
}
