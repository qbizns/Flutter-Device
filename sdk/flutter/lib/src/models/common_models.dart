import 'package:meta/meta.dart';

/// Device status enumeration
enum DeviceStatus {
  unknown,
  ready,
  busy,
  error,
  disconnected,
  offline;

  static DeviceStatus fromString(String value) {
    return DeviceStatus.values.firstWhere(
      (e) => e.name.toUpperCase() == value.toUpperCase().replaceAll('DEVICE_STATUS_', ''),
      orElse: () => DeviceStatus.unknown,
    );
  }
}

/// Base device information
@immutable
class DeviceInfo {
  final String deviceId;
  final String name;
  final String kind;
  final DeviceStatus status;
  final Map<String, dynamic>? metadata;

  const DeviceInfo({
    required this.deviceId,
    required this.name,
    required this.kind,
    required this.status,
    this.metadata,
  });

  factory DeviceInfo.fromJson(Map<String, dynamic> json) {
    return DeviceInfo(
      deviceId: json['device_id'] as String,
      name: json['name'] as String,
      kind: json['kind'] as String,
      status: DeviceStatus.fromString(json['status'] as String),
      metadata: json['metadata'] as Map<String, dynamic>?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'device_id': deviceId,
      'name': name,
      'kind': kind,
      'status': status.name.toUpperCase(),
      if (metadata != null) 'metadata': metadata,
    };
  }

  @override
  String toString() => 'DeviceInfo(deviceId: $deviceId, name: $name, status: $status)';
}

/// Position for elements (x, y coordinates)
@immutable
class Position {
  final int x;
  final int y;

  const Position({required this.x, required this.y});

  factory Position.fromJson(Map<String, dynamic> json) {
    return Position(
      x: json['x'] as int,
      y: json['y'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {'x': x, 'y': y};
  }

  @override
  String toString() => 'Position(x: $x, y: $y)';

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is Position && other.x == x && other.y == y;
  }

  @override
  int get hashCode => Object.hash(x, y);
}

/// Size dimensions (width, height)
@immutable
class Size {
  final int width;
  final int height;

  const Size({required this.width, required this.height});

  factory Size.fromJson(Map<String, dynamic> json) {
    return Size(
      width: json['width'] as int,
      height: json['height'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {'width': width, 'height': height};
  }

  @override
  String toString() => 'Size(width: $width, height: $height)';

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is Size && other.width == width && other.height == height;
  }

  @override
  int get hashCode => Object.hash(width, height);
}

/// Color representation (RGB)
@immutable
class RGBColor {
  final int red;
  final int green;
  final int blue;

  const RGBColor({
    required this.red,
    required this.green,
    required this.blue,
  });

  factory RGBColor.fromHex(String hex) {
    final hexColor = hex.replaceAll('#', '');
    return RGBColor(
      red: int.parse(hexColor.substring(0, 2), radix: 16),
      green: int.parse(hexColor.substring(2, 4), radix: 16),
      blue: int.parse(hexColor.substring(4, 6), radix: 16),
    );
  }

  factory RGBColor.fromJson(Map<String, dynamic> json) {
    return RGBColor(
      red: json['red'] as int,
      green: json['green'] as int,
      blue: json['blue'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'red': red,
      'green': green,
      'blue': blue,
    };
  }

  String toHex() {
    return '#${red.toRadixString(16).padLeft(2, '0')}'
        '${green.toRadixString(16).padLeft(2, '0')}'
        '${blue.toRadixString(16).padLeft(2, '0')}';
  }

  @override
  String toString() => 'RGBColor(r: $red, g: $green, b: $blue)';

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is RGBColor &&
        other.red == red &&
        other.green == green &&
        other.blue == blue;
  }

  @override
  int get hashCode => Object.hash(red, green, blue);
}

/// Timestamp wrapper for consistent date handling
@immutable
class Timestamp {
  final DateTime dateTime;

  const Timestamp(this.dateTime);

  factory Timestamp.now() => Timestamp(DateTime.now());

  factory Timestamp.fromJson(dynamic json) {
    if (json is String) {
      return Timestamp(DateTime.parse(json));
    } else if (json is Map<String, dynamic>) {
      // Handle protobuf timestamp format
      final seconds = json['seconds'] as int?;
      final nanos = json['nanos'] as int?;
      if (seconds != null) {
        return Timestamp(
          DateTime.fromMillisecondsSinceEpoch(
            seconds * 1000 + (nanos ?? 0) ~/ 1000000,
          ),
        );
      }
    }
    throw FormatException('Invalid timestamp format: $json');
  }

  Map<String, dynamic> toJson() {
    final millis = dateTime.millisecondsSinceEpoch;
    return {
      'seconds': millis ~/ 1000,
      'nanos': (millis % 1000) * 1000000,
    };
  }

  String toIso8601() => dateTime.toIso8601String();

  @override
  String toString() => dateTime.toString();

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is Timestamp && other.dateTime == dateTime;
  }

  @override
  int get hashCode => dateTime.hashCode;
}

/// Pagination information
@immutable
class PaginationInfo {
  final int pageSize;
  final String? pageToken;
  final String? nextPageToken;
  final int? totalCount;

  const PaginationInfo({
    required this.pageSize,
    this.pageToken,
    this.nextPageToken,
    this.totalCount,
  });

  factory PaginationInfo.fromJson(Map<String, dynamic> json) {
    return PaginationInfo(
      pageSize: json['page_size'] as int,
      pageToken: json['page_token'] as String?,
      nextPageToken: json['next_page_token'] as String?,
      totalCount: json['total_count'] as int?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'page_size': pageSize,
      if (pageToken != null) 'page_token': pageToken,
      if (nextPageToken != null) 'next_page_token': nextPageToken,
      if (totalCount != null) 'total_count': totalCount,
    };
  }

  bool get hasMore => nextPageToken != null && nextPageToken!.isNotEmpty;

  @override
  String toString() =>
      'PaginationInfo(pageSize: $pageSize, hasMore: $hasMore, totalCount: $totalCount)';
}

/// Generic result wrapper
@immutable
class Result<T> {
  final T? data;
  final String? error;
  final bool success;

  const Result.success(this.data)
      : error = null,
        success = true;

  const Result.failure(this.error)
      : data = null,
        success = false;

  bool get isSuccess => success;
  bool get isFailure => !success;

  R when<R>({
    required R Function(T data) success,
    required R Function(String error) failure,
  }) {
    if (isSuccess) {
      return success(data as T);
    } else {
      return failure(error!);
    }
  }

  @override
  String toString() {
    return success ? 'Result.success($data)' : 'Result.failure($error)';
  }
}
