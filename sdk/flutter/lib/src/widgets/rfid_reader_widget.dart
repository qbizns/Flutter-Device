import 'dart:async';
import 'package:flutter/material.dart';
import '../services/rfid_reader_service.dart';
import '../models/rfid_models.dart';
import '../exceptions.dart';

/// A widget that provides UI for reading RFID/NFC cards
///
/// Example:
/// ```dart
/// RFIDReaderWidget(
///   readerService: client.rfidReader('reader-01'),
///   onCardRead: (response) {
///     print('Card UID: ${response.card.uid}');
///   },
/// )
/// ```
class RFIDReaderWidget extends StatefulWidget {
  /// The RFID reader service
  final RFIDReaderService readerService;

  /// Callback when a card is successfully read
  final void Function(CardReadResponse response) onCardRead;

  /// Optional callback for errors
  final void Function(Object error)? onError;

  /// Timeout for card reading
  final Duration timeout;

  /// Expected card types (null = accept all)
  final List<CardType>? expectedTypes;

  /// Custom button text
  final String? buttonText;

  /// Custom reading text
  final String? readingText;

  const RFIDReaderWidget({
    super.key,
    required this.readerService,
    required this.onCardRead,
    this.onError,
    this.timeout = const Duration(seconds: 30),
    this.expectedTypes,
    this.buttonText,
    this.readingText,
  });

  @override
  State<RFIDReaderWidget> createState() => _RFIDReaderWidgetState();
}

class _RFIDReaderWidgetState extends State<RFIDReaderWidget> {
  bool _isReading = false;
  CardReadResponse? _lastCard;
  String? _error;

  Future<void> _startReading() async {
    setState(() {
      _isReading = true;
      _error = null;
    });

    try {
      final response = await widget.readerService.readCard(
        timeout: widget.timeout,
        expectedTypes: widget.expectedTypes,
      );

      if (mounted) {
        setState(() {
          _lastCard = response;
          _isReading = false;
        });

        widget.onCardRead(response);
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isReading = false;
          _error = e is DeviceBridgeException ? e.message : e.toString();
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
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (_isReading)
              Column(
                children: [
                  const CircularProgressIndicator(),
                  const SizedBox(height: 16),
                  Text(
                    widget.readingText ?? 'Present card to reader...',
                    style: const TextStyle(fontSize: 16),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () {
                      setState(() => _isReading = false);
                    },
                    child: const Text('Cancel'),
                  ),
                ],
              )
            else
              Column(
                children: [
                  if (_lastCard != null) _buildCardInfo(_lastCard!),
                  if (_error != null) _buildError(_error!),
                  const SizedBox(height: 16),
                  ElevatedButton.icon(
                    onPressed: _startReading,
                    icon: const Icon(Icons.nfc),
                    label: Text(widget.buttonText ?? 'Read Card'),
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildCardInfo(CardReadResponse response) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.green.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.green),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.check_circle, color: Colors.green),
              const SizedBox(width: 8),
              const Text(
                'Card Read Successfully',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          _buildInfoRow('UID', response.card.uid),
          _buildInfoRow('Type', response.card.type.name),
          _buildInfoRow('Technology', response.card.technology.name),
          if (response.card.serialNumber.isNotEmpty)
            _buildInfoRow('Serial', response.card.serialNumber),
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

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        children: [
          Text(
            '$label: ',
            style: TextStyle(
              color: Colors.grey.shade700,
              fontSize: 14,
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(
                fontFamily: 'monospace',
                fontSize: 14,
              ),
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}

/// A widget that continuously listens for RFID card reads
class RFIDCardListener extends StatefulWidget {
  /// The RFID reader service
  final RFIDReaderService readerService;

  /// Callback when a card is read
  final void Function(CardReadEvent event) onCardRead;

  /// Optional callback for errors
  final void Function(Object error)? onError;

  /// Optional filter for card types
  final List<CardType>? typeFilter;

  /// The child widget
  final Widget child;

  const RFIDCardListener({
    super.key,
    required this.readerService,
    required this.onCardRead,
    this.onError,
    this.typeFilter,
    required this.child,
  });

  @override
  State<RFIDCardListener> createState() => _RFIDCardListenerState();
}

class _RFIDCardListenerState extends State<RFIDCardListener> {
  StreamSubscription<CardReadEvent>? _subscription;

  @override
  void initState() {
    super.initState();
    _subscribe();
  }

  @override
  void didUpdateWidget(RFIDCardListener oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.readerService != widget.readerService ||
        oldWidget.typeFilter != widget.typeFilter) {
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
    _subscription =
        widget.readerService.subscribeCardReads(filterTypes: widget.typeFilter).listen(
      (event) {
        widget.onCardRead(event);
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
