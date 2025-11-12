import 'package:flutter/material.dart';
import 'package:flutter_device_bridge/flutter_device_bridge.dart';

void main() {
  runApp(const DeviceBridgeExampleApp());
}

class DeviceBridgeExampleApp extends StatelessWidget {
  const DeviceBridgeExampleApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Device Bridge Example',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        useMaterial3: true,
      ),
      home: const HomePage(),
    );
  }
}

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  late DeviceBridgeClient client;

  @override
  void initState() {
    super.initState();
    // Initialize Device Bridge client
    client = DeviceBridgeClient(
      baseUrl: 'http://localhost:8080',
      debug: true,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Device Bridge Examples'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _buildExampleCard(
            context,
            title: 'Printer Example',
            description: 'Print receipts with text, barcodes, and QR codes',
            icon: Icons.print,
            color: Colors.blue,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => PrinterExamplePage(client: client),
              ),
            ),
          ),
          _buildExampleCard(
            context,
            title: 'Scanner Example',
            description: 'Scan barcodes and QR codes in real-time',
            icon: Icons.qr_code_scanner,
            color: Colors.green,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => ScannerExamplePage(client: client),
              ),
            ),
          ),
          _buildExampleCard(
            context,
            title: 'RFID Reader Example',
            description: 'Read RFID/NFC cards',
            icon: Icons.nfc,
            color: Colors.orange,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => RFIDExamplePage(client: client),
              ),
            ),
          ),
          _buildExampleCard(
            context,
            title: 'Access Control Example',
            description: 'Control doors and manage access',
            icon: Icons.door_front_door,
            color: Colors.red,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => AccessControlExamplePage(client: client),
              ),
            ),
          ),
          _buildExampleCard(
            context,
            title: 'Badge Printer Example',
            description: 'Print ID badges and access cards',
            icon: Icons.badge,
            color: Colors.purple,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => BadgePrinterExamplePage(client: client),
              ),
            ),
          ),
          _buildExampleCard(
            context,
            title: 'Payment Terminal Example',
            description: 'Process payments and refunds',
            icon: Icons.payment,
            color: Colors.teal,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => PaymentTerminalExamplePage(client: client),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildExampleCard(
    BuildContext context, {
    required String title,
    required String description,
    required IconData icon,
    required Color color,
    required VoidCallback onTap,
  }) {
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: color,
          child: Icon(icon, color: Colors.white),
        ),
        title: Text(
          title,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Text(description),
        trailing: const Icon(Icons.arrow_forward_ios),
        onTap: onTap,
      ),
    );
  }
}

// Printer Example Page
class PrinterExamplePage extends StatefulWidget {
  final DeviceBridgeClient client;

  const PrinterExamplePage({super.key, required this.client});

  @override
  State<PrinterExamplePage> createState() => _PrinterExamplePageState();
}

class _PrinterExamplePageState extends State<PrinterExamplePage> {
  bool _printing = false;
  String? _result;

  Future<void> _printReceipt() async {
    setState(() {
      _printing = true;
      _result = null;
    });

    try {
      final printer = widget.client.printer('printer-01');

      final receipt = ReceiptBuilder()
          .header('Example Store')
          .text('123 Main Street', alignment: TextAlignment.center)
          .text('City, ST 12345', alignment: TextAlignment.center)
          .lineFeed()
          .divider()
          .keyValue('Item', 'Price')
          .keyValue('Coffee', '\$3.50')
          .keyValue('Pastry', '\$2.50')
          .divider()
          .keyValue('Subtotal', '\$6.00')
          .keyValue('Tax', '\$0.48')
          .keyValue('Total', '\$6.48')
          .lineFeed(lines: 2)
          .qrCode('https://example.com/receipt/12345', size: 6)
          .lineFeed()
          .text('Thank you!', alignment: TextAlignment.center)
          .lineFeed(lines: 2)
          .cut()
          .build();

      final result = await printer.print(receipt);

      setState(() {
        _result = 'Receipt printed successfully!\nJob ID: ${result.jobId}';
        _printing = false;
      });
    } catch (e) {
      setState(() {
        _result = 'Error: $e';
        _printing = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Printer Example'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Print a sample receipt with text, dividers, key-value pairs, and a QR code.',
              style: TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 24),
            if (_result != null)
              Card(
                color: _result!.startsWith('Error') ? Colors.red.shade50 : Colors.green.shade50,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Text(_result!),
                ),
              ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: _printing ? null : _printReceipt,
              icon: _printing
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.print),
              label: Text(_printing ? 'Printing...' : 'Print Receipt'),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.all(16),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// Scanner Example Page
class ScannerExamplePage extends StatefulWidget {
  final DeviceBridgeClient client;

  const ScannerExamplePage({super.key, required this.client});

  @override
  State<ScannerExamplePage> createState() => _ScannerExamplePageState();
}

class _ScannerExamplePageState extends State<ScannerExamplePage> {
  ScanResult? _lastScan;

  @override
  Widget build(BuildContext context) {
    final scanner = widget.client.scanner('scanner-01');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Scanner Example'),
      ),
      body: ScannerListener(
        scanner: scanner,
        onScan: (event) {
          setState(() {
            _lastScan = event.scan;
          });
        },
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Scan a barcode or QR code to see the result below:',
                style: TextStyle(fontSize: 16),
              ),
              const SizedBox(height: 24),
              const Icon(
                Icons.qr_code_scanner,
                size: 100,
                color: Colors.blue,
              ),
              const SizedBox(height: 24),
              if (_lastScan != null)
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Last Scan:',
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 18,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text('Data: ${_lastScan!.data}'),
                        Text('Type: ${_lastScan!.symbology.name}'),
                        Text('Quality: ${_lastScan!.quality}%'),
                      ],
                    ),
                  ),
                )
              else
                const Card(
                  child: Padding(
                    padding: EdgeInsets.all(16),
                    child: Text(
                      'Waiting for scan...',
                      textAlign: TextAlign.center,
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

// RFID Example Page
class RFIDExamplePage extends StatelessWidget {
  final DeviceBridgeClient client;

  const RFIDExamplePage({super.key, required this.client});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('RFID Reader Example'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Present an RFID/NFC card to the reader:',
              style: TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 24),
            RFIDReaderWidget(
              readerService: client.rfidReader('reader-01'),
              onCardRead: (response) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Card read: ${response.card.uid}'),
                    backgroundColor: Colors.green,
                  ),
                );
              },
              onError: (error) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Error: $error'),
                    backgroundColor: Colors.red,
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

// Access Control Example Page
class AccessControlExamplePage extends StatelessWidget {
  final DeviceBridgeClient client;

  const AccessControlExamplePage({super.key, required this.client});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Access Control Example'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Control door access:',
              style: TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 24),
            AccessControlWidget(
              accessControlService: client.accessControl('controller-01'),
              doorId: 'main-entrance',
              onAccessGranted: (response) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Access granted'),
                    backgroundColor: Colors.green,
                  ),
                );
              },
              onAccessDenied: (response) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Access denied: ${response.reason}'),
                    backgroundColor: Colors.red,
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

// Badge Printer Example Page
class BadgePrinterExamplePage extends StatefulWidget {
  final DeviceBridgeClient client;

  const BadgePrinterExamplePage({super.key, required this.client});

  @override
  State<BadgePrinterExamplePage> createState() =>
      _BadgePrinterExamplePageState();
}

class _BadgePrinterExamplePageState extends State<BadgePrinterExamplePage> {
  bool _printing = false;
  String? _result;

  Future<void> _printBadge() async {
    setState(() {
      _printing = true;
      _result = null;
    });

    try {
      final badgePrinter = widget.client.badgePrinter('zebra-01');

      final layout = CardLayout(
        cardSize: CardSize.cr80,
        orientation: CardOrientation.landscape,
        backgroundColor: const RGBColor(red: 255, green: 255, blue: 255),
        elements: [
          BadgeTextElement(
            position: const Position(x: 200, y: 150),
            size: const Size(width: 400, height: 60),
            text: 'John Doe',
            fontSize: 28,
            bold: true,
          ),
          BadgeTextElement(
            position: const Position(x: 200, y: 220),
            size: const Size(width: 400, height: 40),
            text: 'Engineering Department',
            fontSize: 18,
          ),
          BadgeTextElement(
            position: const Position(x: 200, y: 270),
            size: const Size(width: 400, height: 40),
            text: 'ID: EMP-12345',
            fontSize: 16,
            color: const RGBColor(red: 128, green: 128, blue: 128),
          ),
        ],
      );

      final job = BadgePrintJob(
        frontLayout: layout,
        quality: PrintQuality.high,
      );

      final result = await badgePrinter.printBadge(job);

      setState(() {
        _result = 'Badge printed successfully!\nJob ID: ${result.jobId}';
        _printing = false;
      });
    } catch (e) {
      setState(() {
        _result = 'Error: $e';
        _printing = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Badge Printer Example'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Print a sample employee badge:',
              style: TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 24),
            if (_result != null)
              Card(
                color: _result!.startsWith('Error')
                    ? Colors.red.shade50
                    : Colors.green.shade50,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Text(_result!),
                ),
              ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: _printing ? null : _printBadge,
              icon: _printing
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.badge),
              label: Text(_printing ? 'Printing...' : 'Print Badge'),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.all(16),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// Payment Terminal Example Page
class PaymentTerminalExamplePage extends StatefulWidget {
  final DeviceBridgeClient client;

  const PaymentTerminalExamplePage({super.key, required this.client});

  @override
  State<PaymentTerminalExamplePage> createState() =>
      _PaymentTerminalExamplePageState();
}

class _PaymentTerminalExamplePageState
    extends State<PaymentTerminalExamplePage> {
  bool _processing = false;
  PaymentResponse? _result;

  Future<void> _processPayment() async {
    setState(() {
      _processing = true;
      _result = null;
    });

    try {
      final terminal = widget.client.paymentTerminal('terminal-01');

      final request = PaymentRequest(
        amount: 1050, // $10.50
        currency: 'USD',
        type: TransactionType.sale,
        orderId: 'ORDER-${DateTime.now().millisecondsSinceEpoch}',
      );

      final result = await terminal.processPayment(request);

      setState(() {
        _result = result.payment;
        _processing = false;
      });
    } catch (e) {
      setState(() {
        _processing = false;
      });

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Payment Terminal Example'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Process a \$10.50 payment:',
              style: TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 24),
            if (_result != null)
              Card(
                color: _result!.isApproved
                    ? Colors.green.shade50
                    : Colors.red.shade50,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _result!.isApproved
                            ? 'Payment Approved'
                            : 'Payment Declined',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 18,
                          color: _result!.isApproved ? Colors.green : Colors.red,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text('Transaction ID: ${_result!.transactionId}'),
                      Text('Amount: \$${_result!.amount / 100}'),
                      if (_result!.cardData != null)
                        Text('Card: ${_result!.cardData!.brand.name}'),
                      if (_result!.approvalCode != null)
                        Text('Approval Code: ${_result!.approvalCode}'),
                      if (_result!.responseMessage != null)
                        Text('Message: ${_result!.responseMessage}'),
                    ],
                  ),
                ),
              ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: _processing ? null : _processPayment,
              icon: _processing
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.payment),
              label: Text(_processing ? 'Processing...' : 'Process Payment'),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.all(16),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
