import 'dart:async';
import 'package:flutter/material.dart';
import '../services/access_control_service.dart';
import '../models/access_control_models.dart';
import '../exceptions.dart';

/// A widget that provides UI for access control operations
///
/// Example:
/// ```dart
/// AccessControlWidget(
///   accessControlService: client.accessControl('controller-01'),
///   doorId: 'main-entrance',
///   onAccessGranted: (response) {
///     print('Access granted to ${response.userInfo?.userName}');
///   },
/// )
/// ```
class AccessControlWidget extends StatefulWidget {
  /// The access control service
  final AccessControlService accessControlService;

  /// The door ID to control
  final String doorId;

  /// Callback when access is granted
  final void Function(GrantAccessResponse response)? onAccessGranted;

  /// Callback when access is denied
  final void Function(GrantAccessResponse response)? onAccessDenied;

  /// Optional callback for errors
  final void Function(Object error)? onError;

  const AccessControlWidget({
    super.key,
    required this.accessControlService,
    required this.doorId,
    this.onAccessGranted,
    this.onAccessDenied,
    this.onError,
  });

  @override
  State<AccessControlWidget> createState() => _AccessControlWidgetState();
}

class _AccessControlWidgetState extends State<AccessControlWidget> {
  DoorStatus? _doorStatus;
  bool _isLoading = false;
  String? _error;
  GrantAccessResponse? _lastResponse;

  @override
  void initState() {
    super.initState();
    _loadDoorStatus();
  }

  Future<void> _loadDoorStatus() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final status = await widget.accessControlService.getDoorStatus(widget.doorId);
      if (mounted) {
        setState(() {
          _doorStatus = status;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e is DeviceBridgeException ? e.message : e.toString();
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _unlockDoor() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final status = await widget.accessControlService.unlockDoor(
        doorId: widget.doorId,
        duration: const Duration(seconds: 5),
      );

      if (mounted) {
        setState(() {
          _doorStatus = status;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e is DeviceBridgeException ? e.message : e.toString();
          _isLoading = false;
        });

        if (widget.onError != null) {
          widget.onError!(e);
        }
      }
    }
  }

  Future<void> _lockDoor() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final status = await widget.accessControlService.lockDoor(
        doorId: widget.doorId,
      );

      if (mounted) {
        setState(() {
          _doorStatus = status;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e is DeviceBridgeException ? e.message : e.toString();
          _isLoading = false;
        });

        if (widget.onError != null) {
          widget.onError!(e);
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              _doorStatus?.doorName ?? widget.doorId,
              style: const TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            if (_isLoading)
              const Center(child: CircularProgressIndicator())
            else if (_error != null)
              _buildError(_error!)
            else if (_doorStatus != null)
              _buildDoorStatus(_doorStatus!)
            else
              const Center(child: Text('Loading...')),
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                ElevatedButton.icon(
                  onPressed: _isLoading ? null : _unlockDoor,
                  icon: const Icon(Icons.lock_open),
                  label: const Text('Unlock'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.green,
                    foregroundColor: Colors.white,
                  ),
                ),
                ElevatedButton.icon(
                  onPressed: _isLoading ? null : _lockDoor,
                  icon: const Icon(Icons.lock),
                  label: const Text('Lock'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.red,
                    foregroundColor: Colors.white,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            TextButton.icon(
              onPressed: _isLoading ? null : _loadDoorStatus,
              icon: const Icon(Icons.refresh),
              label: const Text('Refresh Status'),
            ),
            if (_lastResponse != null) ...[
              const SizedBox(height: 16),
              _buildLastResponse(_lastResponse!),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildDoorStatus(DoorStatus status) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: _getStatusColor(status.lockStatus).withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _getStatusColor(status.lockStatus)),
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Lock Status:'),
              Text(
                status.lockStatus.name.toUpperCase(),
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: _getStatusColor(status.lockStatus),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Door Position:'),
              Text(
                status.doorPosition.name.toUpperCase(),
                style: const TextStyle(fontWeight: FontWeight.w500),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Mode:'),
              Text(
                status.mode.name.toUpperCase(),
                style: const TextStyle(fontWeight: FontWeight.w500),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildError(String error) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.red.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.red),
      ),
      child: Row(
        children: [
          const Icon(Icons.error, color: Colors.red),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              error,
              style: const TextStyle(color: Colors.red),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLastResponse(GrantAccessResponse response) {
    final isGranted = response.granted;
    final color = isGranted ? Colors.green : Colors.red;

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                isGranted ? Icons.check_circle : Icons.cancel,
                color: color,
              ),
              const SizedBox(width: 8),
              Text(
                isGranted ? 'Access Granted' : 'Access Denied',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: color,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(response.reason),
          if (response.userInfo != null) ...[
            const SizedBox(height: 4),
            Text('User: ${response.userInfo!.userName}'),
          ],
        ],
      ),
    );
  }

  Color _getStatusColor(LockStatus status) {
    switch (status) {
      case LockStatus.locked:
        return Colors.red;
      case LockStatus.unlocked:
        return Colors.green;
      case LockStatus.error:
        return Colors.red.shade900;
      default:
        return Colors.grey;
    }
  }
}

/// A widget that listens to access control events
class AccessControlEventListener extends StatefulWidget {
  /// The access control service
  final AccessControlService accessControlService;

  /// Callback when an access event occurs
  final void Function(AccessEvent event) onEvent;

  /// Optional callback for errors
  final void Function(Object error)? onError;

  /// Optional filter for specific doors
  final List<String>? doorFilter;

  /// The child widget
  final Widget child;

  const AccessControlEventListener({
    super.key,
    required this.accessControlService,
    required this.onEvent,
    this.onError,
    this.doorFilter,
    required this.child,
  });

  @override
  State<AccessControlEventListener> createState() => _AccessControlEventListenerState();
}

class _AccessControlEventListenerState extends State<AccessControlEventListener> {
  StreamSubscription<AccessEvent>? _subscription;

  @override
  void initState() {
    super.initState();
    _subscribe();
  }

  @override
  void didUpdateWidget(AccessControlEventListener oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.accessControlService != widget.accessControlService ||
        oldWidget.doorFilter != widget.doorFilter) {
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
    _subscription = widget.accessControlService
        .subscribeAccessEvents(doorIds: widget.doorFilter)
        .listen(
      (event) {
        widget.onEvent(event);
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
