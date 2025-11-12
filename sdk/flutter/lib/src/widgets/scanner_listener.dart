import 'dart:async';
import 'package:flutter/material.dart';
import '../services/scanner_service.dart';
import '../models/scanner_models.dart';

/// A widget that listens to scanner events and invokes a callback
///
/// This widget manages the scanner subscription lifecycle and automatically
/// handles cleanup when the widget is disposed.
///
/// Example:
/// ```dart
/// ScannerListener(
///   scanner: client.scanner('scanner-01'),
///   onScan: (event) {
///     print('Scanned: ${event.scan.data}');
///     // Update UI or process scan
///   },
///   onError: (error) {
///     print('Scanner error: $error');
///   },
///   child: YourAppWidget(),
/// )
/// ```
class ScannerListener extends StatefulWidget {
  /// The scanner service to listen to
  final ScannerService scanner;

  /// Callback invoked when a code is scanned
  final void Function(ScanEvent event) onScan;

  /// Optional callback for errors
  final void Function(Object error)? onError;

  /// Optional filter for specific symbologies
  final List<BarcodeSymbology>? symbologyFilter;

  /// Whether to enable sound feedback (default: true)
  final bool enableSound;

  /// Whether to enable haptic feedback (default: true)
  final bool enableHaptic;

  /// The child widget
  final Widget child;

  const ScannerListener({
    super.key,
    required this.scanner,
    required this.onScan,
    this.onError,
    this.symbologyFilter,
    this.enableSound = true,
    this.enableHaptic = true,
    required this.child,
  });

  @override
  State<ScannerListener> createState() => _ScannerListenerState();
}

class _ScannerListenerState extends State<ScannerListener> {
  StreamSubscription<ScanEvent>? _subscription;

  @override
  void initState() {
    super.initState();
    _subscribe();
  }

  @override
  void didUpdateWidget(ScannerListener oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.scanner != widget.scanner ||
        oldWidget.symbologyFilter != widget.symbologyFilter) {
      _unsubscribe();
      _subscribe();
    }
  }

  @override
  void dispose() {
    _unsubscribe();
    super.dispose();
  }

  void _subscribe() {
    _subscription = widget.scanner
        .subscribeScans(symbologyFilter: widget.symbologyFilter)
        .listen(
      (event) {
        // Provide feedback
        if (widget.enableHaptic && mounted) {
          // Haptic feedback would go here if using services
        }

        // Invoke callback
        widget.onScan(event);
      },
      onError: (error) {
        if (widget.onError != null) {
          widget.onError!(error);
        }
      },
    );
  }

  void _unsubscribe() {
    _subscription?.cancel();
    _subscription = null;
  }

  @override
  Widget build(BuildContext context) {
    return widget.child;
  }
}

/// A widget that shows scan results with animation
class ScanResultDisplay extends StatefulWidget {
  /// The scan result to display
  final ScanResult? scanResult;

  /// Duration to show the result
  final Duration displayDuration;

  /// Custom builder for the result display
  final Widget Function(BuildContext context, ScanResult result)? builder;

  const ScanResultDisplay({
    super.key,
    this.scanResult,
    this.displayDuration = const Duration(seconds: 3),
    this.builder,
  });

  @override
  State<ScanResultDisplay> createState() => _ScanResultDisplayState();
}

class _ScanResultDisplayState extends State<ScanResultDisplay>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _fadeAnimation;
  Timer? _hideTimer;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
    );
    _fadeAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(_controller);
  }

  @override
  void didUpdateWidget(ScanResultDisplay oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.scanResult != oldWidget.scanResult &&
        widget.scanResult != null) {
      _showResult();
    }
  }

  @override
  void dispose() {
    _hideTimer?.cancel();
    _controller.dispose();
    super.dispose();
  }

  void _showResult() {
    _controller.forward();
    _hideTimer?.cancel();
    _hideTimer = Timer(widget.displayDuration, () {
      if (mounted) {
        _controller.reverse();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    if (widget.scanResult == null) {
      return const SizedBox.shrink();
    }

    return FadeTransition(
      opacity: _fadeAnimation,
      child: widget.builder != null
          ? widget.builder!(context, widget.scanResult!)
          : _buildDefaultDisplay(context, widget.scanResult!),
    );
  }

  Widget _buildDefaultDisplay(BuildContext context, ScanResult result) {
    return Container(
      padding: const EdgeInsets.all(16),
      margin: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.green.shade100,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.green),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.check_circle, color: Colors.green),
              const SizedBox(width: 8),
              Text(
                'Scanned: ${result.symbology.name}',
                style: const TextStyle(
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            result.data,
            style: const TextStyle(
              fontFamily: 'monospace',
              fontSize: 16,
            ),
          ),
        ],
      ),
    );
  }
}
