import 'package:flutter/material.dart';
import '../models/common_models.dart';

/// A widget that displays device status with icon and message
///
/// Example:
/// ```dart
/// DeviceStatusWidget(
///   deviceInfo: DeviceInfo(
///     deviceId: 'printer-01',
///     name: 'Receipt Printer',
///     kind: 'printer',
///     status: DeviceStatus.ready,
///   ),
/// )
/// ```
class DeviceStatusWidget extends StatelessWidget {
  /// The device information to display
  final DeviceInfo deviceInfo;

  /// Whether to show the device name
  final bool showName;

  /// Whether to show the device ID
  final bool showDeviceId;

  /// Custom status message override
  final String? statusMessage;

  /// Size of the status icon
  final double iconSize;

  const DeviceStatusWidget({
    super.key,
    required this.deviceInfo,
    this.showName = true,
    this.showDeviceId = false,
    this.statusMessage,
    this.iconSize = 24,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        _buildStatusIcon(),
        const SizedBox(width: 8),
        Flexible(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              if (showName)
                Text(
                  deviceInfo.name,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              Text(
                statusMessage ?? _getStatusMessage(),
                style: TextStyle(
                  color: _getStatusColor(),
                  fontSize: 14,
                ),
                overflow: TextOverflow.ellipsis,
              ),
              if (showDeviceId)
                Text(
                  deviceInfo.deviceId,
                  style: TextStyle(
                    color: Colors.grey.shade600,
                    fontSize: 12,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildStatusIcon() {
    IconData iconData;
    Color iconColor;

    switch (deviceInfo.status) {
      case DeviceStatus.ready:
        iconData = Icons.check_circle;
        iconColor = Colors.green;
        break;
      case DeviceStatus.busy:
        iconData = Icons.hourglass_empty;
        iconColor = Colors.orange;
        break;
      case DeviceStatus.error:
        iconData = Icons.error;
        iconColor = Colors.red;
        break;
      case DeviceStatus.disconnected:
      case DeviceStatus.offline:
        iconData = Icons.cloud_off;
        iconColor = Colors.grey;
        break;
      default:
        iconData = Icons.help_outline;
        iconColor = Colors.grey;
    }

    return Icon(iconData, size: iconSize, color: iconColor);
  }

  Color _getStatusColor() {
    switch (deviceInfo.status) {
      case DeviceStatus.ready:
        return Colors.green;
      case DeviceStatus.busy:
        return Colors.orange;
      case DeviceStatus.error:
        return Colors.red;
      case DeviceStatus.disconnected:
      case DeviceStatus.offline:
        return Colors.grey;
      default:
        return Colors.grey;
    }
  }

  String _getStatusMessage() {
    switch (deviceInfo.status) {
      case DeviceStatus.ready:
        return 'Ready';
      case DeviceStatus.busy:
        return 'Busy';
      case DeviceStatus.error:
        return 'Error';
      case DeviceStatus.disconnected:
        return 'Disconnected';
      case DeviceStatus.offline:
        return 'Offline';
      default:
        return 'Unknown';
    }
  }
}

/// A widget that shows a detailed device status card
class DeviceStatusCard extends StatelessWidget {
  /// The device information to display
  final DeviceInfo deviceInfo;

  /// Optional additional status details
  final Map<String, String>? details;

  /// Optional action buttons
  final List<Widget>? actions;

  const DeviceStatusCard({
    super.key,
    required this.deviceInfo,
    this.details,
    this.actions,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            DeviceStatusWidget(
              deviceInfo: deviceInfo,
              showName: true,
              showDeviceId: true,
            ),
            if (details != null && details!.isNotEmpty) ...[
              const SizedBox(height: 16),
              const Divider(),
              const SizedBox(height: 8),
              ...details!.entries.map((entry) => Padding(
                    padding: const EdgeInsets.symmetric(vertical: 4),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          entry.key,
                          style: TextStyle(
                            color: Colors.grey.shade700,
                            fontSize: 14,
                          ),
                        ),
                        Text(
                          entry.value,
                          style: const TextStyle(
                            fontWeight: FontWeight.w500,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  )),
            ],
            if (actions != null && actions!.isNotEmpty) ...[
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: actions!,
              ),
            ],
          ],
        ),
      ),
    );
  }
}

/// A widget that shows a loading indicator for device operations
class DeviceOperationLoader extends StatelessWidget {
  /// The message to show
  final String message;

  /// Whether to show a progress indicator
  final bool showProgress;

  const DeviceOperationLoader({
    super.key,
    required this.message,
    this.showProgress = true,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (showProgress) const CircularProgressIndicator(),
          if (showProgress) const SizedBox(height: 16),
          Text(
            message,
            style: const TextStyle(fontSize: 16),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}
