import 'package:meta/meta.dart';

/// Configuration for Device Bridge client
@immutable
class DeviceBridgeConfig {
  /// Base URL for Device Bridge server (REST API)
  ///
  /// Example: 'http://localhost:8080' or 'https://device-bridge.company.com'
  final String baseUrl;

  /// gRPC host (optional, for gRPC connections)
  final String? grpcHost;

  /// gRPC port (optional, for gRPC connections)
  final int? grpcPort;

  /// Use TLS for connections
  final bool useTls;

  /// API key for authentication (optional)
  final String? apiKey;

  /// Request timeout duration
  final Duration timeout;

  /// Enable debug logging
  final bool debug;

  /// Custom headers for all requests
  final Map<String, String>? headers;

  /// Maximum number of retry attempts for failed requests
  final int maxRetries;

  /// Delay between retry attempts
  final Duration retryDelay;

  const DeviceBridgeConfig({
    required this.baseUrl,
    this.grpcHost,
    this.grpcPort,
    this.useTls = false,
    this.apiKey,
    this.timeout = const Duration(seconds: 30),
    this.debug = false,
    this.headers,
    this.maxRetries = 3,
    this.retryDelay = const Duration(seconds: 1),
  });

  /// Create configuration for local development
  factory DeviceBridgeConfig.local({
    int port = 8080,
    bool debug = true,
  }) {
    return DeviceBridgeConfig(
      baseUrl: 'http://localhost:$port',
      debug: debug,
      timeout: const Duration(seconds: 30),
    );
  }

  /// Create configuration for production
  factory DeviceBridgeConfig.production({
    required String host,
    String? apiKey,
    bool useTls = true,
  }) {
    final scheme = useTls ? 'https' : 'http';
    return DeviceBridgeConfig(
      baseUrl: '$scheme://$host',
      apiKey: apiKey,
      useTls: useTls,
      debug: false,
      timeout: const Duration(seconds: 30),
    );
  }

  /// Get full URL for an endpoint
  String getUrl(String path) {
    final cleanPath = path.startsWith('/') ? path : '/$path';
    return '$baseUrl$cleanPath';
  }

  /// Get headers for requests
  Map<String, String> getHeaders() {
    final requestHeaders = <String, String>{
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      ...?headers,
    };

    if (apiKey != null) {
      requestHeaders['Authorization'] = 'Bearer $apiKey';
    }

    return requestHeaders;
  }

  /// Copy with new values
  DeviceBridgeConfig copyWith({
    String? baseUrl,
    String? grpcHost,
    int? grpcPort,
    bool? useTls,
    String? apiKey,
    Duration? timeout,
    bool? debug,
    Map<String, String>? headers,
    int? maxRetries,
    Duration? retryDelay,
  }) {
    return DeviceBridgeConfig(
      baseUrl: baseUrl ?? this.baseUrl,
      grpcHost: grpcHost ?? this.grpcHost,
      grpcPort: grpcPort ?? this.grpcPort,
      useTls: useTls ?? this.useTls,
      apiKey: apiKey ?? this.apiKey,
      timeout: timeout ?? this.timeout,
      debug: debug ?? this.debug,
      headers: headers ?? this.headers,
      maxRetries: maxRetries ?? this.maxRetries,
      retryDelay: retryDelay ?? this.retryDelay,
    );
  }

  @override
  String toString() {
    return 'DeviceBridgeConfig('
        'baseUrl: $baseUrl, '
        'useTls: $useTls, '
        'timeout: $timeout, '
        'debug: $debug'
        ')';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;

    return other is DeviceBridgeConfig &&
        other.baseUrl == baseUrl &&
        other.grpcHost == grpcHost &&
        other.grpcPort == grpcPort &&
        other.useTls == useTls &&
        other.apiKey == apiKey &&
        other.timeout == timeout &&
        other.debug == debug;
  }

  @override
  int get hashCode {
    return Object.hash(
      baseUrl,
      grpcHost,
      grpcPort,
      useTls,
      apiKey,
      timeout,
      debug,
    );
  }
}
