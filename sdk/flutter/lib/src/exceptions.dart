/// Base exception for Device Bridge SDK
class DeviceBridgeException implements Exception {
  /// Error message
  final String message;

  /// HTTP status code (if applicable)
  final int? statusCode;

  /// Original exception (if any)
  final Object? cause;

  /// Stack trace (if any)
  final StackTrace? stackTrace;

  const DeviceBridgeException(
    this.message, {
    this.statusCode,
    this.cause,
    this.stackTrace,
  });

  @override
  String toString() {
    final buffer = StringBuffer('DeviceBridgeException: $message');

    if (statusCode != null) {
      buffer.write(' (status code: $statusCode)');
    }

    if (cause != null) {
      buffer.write('\nCaused by: $cause');
    }

    return buffer.toString();
  }
}

/// Exception thrown when a device is not found
class DeviceNotFoundException extends DeviceBridgeException {
  final String deviceId;

  const DeviceNotFoundException(this.deviceId)
      : super('Device not found: $deviceId', statusCode: 404);
}

/// Exception thrown when a device operation times out
class DeviceTimeoutException extends DeviceBridgeException {
  final Duration timeout;

  const DeviceTimeoutException(String message, this.timeout)
      : super('$message (timeout: ${timeout.inSeconds}s)', statusCode: 408);
}

/// Exception thrown when a device is busy
class DeviceBusyException extends DeviceBridgeException {
  final String deviceId;

  const DeviceBusyException(this.deviceId)
      : super('Device is busy: $deviceId', statusCode: 503);
}

/// Exception thrown when a device returns an error
class DeviceErrorException extends DeviceBridgeException {
  final String deviceId;
  final String errorCode;

  const DeviceErrorException(
    this.deviceId,
    this.errorCode,
    String message,
  ) : super('Device error ($errorCode): $message');
}

/// Exception thrown for authentication errors
class AuthenticationException extends DeviceBridgeException {
  const AuthenticationException(String message)
      : super(message, statusCode: 401);
}

/// Exception thrown for authorization errors
class AuthorizationException extends DeviceBridgeException {
  const AuthorizationException(String message)
      : super(message, statusCode: 403);
}

/// Exception thrown for validation errors
class ValidationException extends DeviceBridgeException {
  final Map<String, List<String>> errors;

  const ValidationException(this.errors)
      : super('Validation failed', statusCode: 400);

  @override
  String toString() {
    final buffer = StringBuffer('ValidationException:\n');
    errors.forEach((field, messages) {
      buffer.write('  $field: ${messages.join(', ')}\n');
    });
    return buffer.toString();
  }
}

/// Exception thrown for connection errors
class ConnectionException extends DeviceBridgeException {
  const ConnectionException(String message, {Object? cause})
      : super(message, cause: cause);
}

/// Exception thrown when operation is not supported
class UnsupportedOperationException extends DeviceBridgeException {
  final String operation;
  final String deviceType;

  const UnsupportedOperationException(this.operation, this.deviceType)
      : super('Operation "$operation" is not supported for device type: $deviceType');
}
