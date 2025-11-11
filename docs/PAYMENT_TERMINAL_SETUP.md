# Payment Terminal Setup Guide

**Device Bridge v2 - Payment Terminal Driver**

This guide covers setup, configuration, and usage of the payment terminal driver for processing card transactions.

## Table of Contents

- [Overview](#overview)
- [Supported Networks](#supported-networks)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Transaction Types](#transaction-types)
- [Security & PCI DSS](#security--pci-dss)
- [API Reference](#api-reference)
- [Payment Providers](#payment-providers)
  - [Mada Provider (Saudi Arabia)](#mada-provider-saudi-arabia)
  - [KNET Provider (Kuwait)](#knet-provider-kuwait)
- [Testing](#testing)
  - [Payment Simulator](#payment-simulator)
- [Troubleshooting](#troubleshooting)
- [Production Deployment](#production-deployment)

## Overview

The payment terminal driver enables card payment processing through ISO 8583 protocol over TCP/IP. It supports:

- **Payment Networks:** Mada, KNET, Visa, Mastercard
- **Transaction Types:** Sale, Void, Refund, Pre-authorization, Balance Inquiry, Settlement
- **Security:** TLS/SSL encryption, PCI DSS compliant, card data masking
- **Reliability:** Automatic reconnection, retry logic, timeout handling

### Architecture

```
┌─────────────────────┐
│  Your Application   │
│   (POS/Kiosk/Web)   │
└──────────┬──────────┘
           │ gRPC/REST/WebSocket
           ↓
┌─────────────────────┐
│  Device Bridge v2   │
│  Payment Driver     │
└──────────┬──────────┘
           │ ISO 8583 / TCP
           ↓
┌─────────────────────┐
│ Payment Terminal    │
│  (Physical/Cloud)   │
└─────────────────────┘
           │
           ↓
┌─────────────────────┐
│  Payment Network    │
│ (Mada/KNET/Visa)   │
└─────────────────────┘
```

## Supported Networks

| Network | Region | Currency | Card Types |
|---------|--------|----------|------------|
| **Mada** | Saudi Arabia | SAR | Domestic debit |
| **KNET** | Kuwait | KWD | Domestic debit/credit |
| **Benefit** | Bahrain | BHD | Domestic cards |
| **Visa** | International | Multiple | Credit/debit |
| **Mastercard** | International | Multiple | Credit/debit |

## Quick Start

### Prerequisites

1. **Payment Terminal**
   - Physical terminal or cloud-based service
   - Terminal ID and Merchant ID from your acquirer
   - Network connectivity to terminal

2. **Device Bridge v2**
   - Version 2.0.0 or later
   - Go 1.21+ installed (for building from source)

3. **Credentials**
   - Terminal ID (8-15 digits)
   - Merchant ID (up to 15 digits)
   - Acquirer information

### Basic Configuration

Create `config.yaml`:

```yaml
server:
  grpc:
    address: ":50051"
  rest:
    address: ":8080"

devices:
  - id: mada-terminal-01
    name: "Mada Payment Terminal"
    kind: payment.tcp
    enabled: true
    metadata:
      host: "192.168.1.100"      # Terminal IP
      port: "3000"                # Terminal port
      terminal_id: "12345678"     # Your terminal ID
      merchant_id: "123456789012" # Your merchant ID
      provider: "mada"            # Payment provider
      use_tls: "true"             # Use TLS encryption
      timeout: "30s"
      read_timeout: "60s"
```

### Start Device Bridge

```bash
# Build Device Bridge
go build -o device-bridge ./cmd/bridge

# Run with config
./device-bridge -config config.yaml
```

### Process Your First Transaction

Using gRPC:

```bash
grpcurl -plaintext -d '{
  "device_id": "mada-terminal-01",
  "transaction": {
    "type": "sale",
    "amount": 10000,
    "currency": "SAR",
    "reference": "INV-001"
  }
}' localhost:50051 devicebridge.v1.DeviceBridge/ProcessPayment
```

Using REST API:

```bash
curl -X POST http://localhost:8080/v1/devices/mada-terminal-01/payment \
  -H "Content-Type: application/json" \
  -d '{
    "type": "sale",
    "amount": 10000,
    "currency": "SAR",
    "reference": "INV-001"
  }'
```

## Configuration

### Connection Settings

```yaml
metadata:
  # Terminal connection
  host: "192.168.1.100"      # Terminal IP address or hostname
  port: "3000"                # Terminal port (usually 3000 or 443)
  use_tls: "true"             # Enable TLS/SSL encryption
  tls_skip_verify: "false"    # Verify TLS certificates (always false in production)

  # Terminal identification
  terminal_id: "TERM0001"     # Terminal ID from acquirer
  merchant_id: "MERCH001"     # Merchant ID from acquirer

  # Provider
  provider: "mada"            # Payment provider: mada, knet, simulator

  # Timeouts
  timeout: "30s"              # Connection timeout
  read_timeout: "60s"         # Transaction read timeout (wait for customer)
  write_timeout: "10s"        # Write timeout

  # Retry settings
  max_retries: "3"            # Maximum retry attempts on failure
  retry_delay: "2s"           # Delay between retries
```

### Provider Configuration

#### Mada (Saudi Arabia)

```yaml
devices:
  - id: mada-terminal
    name: "Mada Terminal"
    kind: payment.tcp
    metadata:
      host: "192.168.1.100"
      port: "3000"
      terminal_id: "12345678"
      merchant_id: "123456789012345"
      provider: "mada"
      use_tls: "true"
      currency: "SAR"  # Saudi Riyal (ISO 4217: 682)
```

Currency: SAR (2 decimal places)
- Amount 10.50 SAR = 1050 (in fils/halalas)
- Amount 100.00 SAR = 10000

#### KNET (Kuwait)

```yaml
devices:
  - id: knet-terminal
    name: "KNET Terminal"
    kind: payment.tcp
    metadata:
      host: "10.0.1.50"
      port: "3000"
      terminal_id: "KNET0001"
      merchant_id: "KNETMERCHANT"
      provider: "knet"
      use_tls: "true"
      currency: "KWD"  # Kuwaiti Dinar (ISO 4217: 414)
      read_timeout: "90s"  # Longer for PIN entry
```

Currency: KWD (3 decimal places)
- Amount 10.500 KWD = 10500 (in fils)
- Amount 100.000 KWD = 100000

#### Payment Simulator (Development)

```yaml
devices:
  - id: simulator
    name: "Payment Simulator"
    kind: payment.tcp
    metadata:
      host: "localhost"
      port: "3000"
      terminal_id: "TEST0001"
      merchant_id: "TEST"
      provider: "simulator"
      use_tls: "false"  # No TLS for local testing
      timeout: "10s"
      read_timeout: "30s"
      max_retries: "0"  # Fail fast in testing
```

## Transaction Types

### 1. Sale Transaction

Process a purchase transaction.

**Request:**
```json
{
  "type": "sale",
  "amount": 10000,
  "currency": "SAR",
  "reference": "INV-12345",
  "invoice_number": "INV-12345",
  "description": "Product purchase"
}
```

**Response:**
```json
{
  "success": true,
  "status": "approved",
  "transaction_id": "123456789012",
  "auth_code": "ABC123",
  "response_code": "00",
  "response_message": "Approved",
  "card_number_masked": "****1234",
  "card_type": "mada",
  "amount": 10000,
  "currency": "SAR",
  "timestamp": "2025-11-11T10:30:00Z",
  "rrn": "123456789012",
  "terminal_id": "12345678",
  "merchant_id": "123456789012345"
}
```

### 2. Void Transaction

Cancel a previous transaction (same day).

**Request:**
```json
{
  "type": "void",
  "original_reference": "123456789012",
  "amount": 10000,
  "currency": "SAR"
}
```

### 3. Refund Transaction

Refund a previous transaction (any day).

**Request:**
```json
{
  "type": "refund",
  "original_reference": "123456789012",
  "amount": 10000,
  "currency": "SAR",
  "reference": "REFUND-001"
}
```

### 4. Pre-Authorization

Authorize an amount without capturing.

**Request:**
```json
{
  "type": "preauth",
  "amount": 10000,
  "currency": "SAR",
  "reference": "PREAUTH-001"
}
```

### 5. Balance Inquiry

Check card balance.

**Request:**
```json
{
  "type": "balance",
  "currency": "SAR"
}
```

### 6. Settlement

Settle batch transactions (end of day).

**Request:**
```json
{
  "batch_number": "20251111",
  "force": false
}
```

## Security & PCI DSS

### PCI DSS Compliance

The payment driver is designed for PCI DSS compliance:

#### ✅ Data Protection

1. **Card Data Masking**
   - PANs are automatically masked: `****1234`
   - Only first 6 and last 4 digits shown in logs
   - No full card numbers in logs or storage

2. **PIN Security**
   - PIN data never logged
   - PIN fields automatically filtered
   - Encrypted transmission only

3. **Secure Transmission**
   - TLS 1.2+ encryption required in production
   - Certificate validation enforced
   - No plaintext sensitive data

4. **No Data Storage**
   - Device Bridge does not store card data
   - All transaction data is ephemeral
   - No database persistence of sensitive info

#### 🔒 Security Best Practices

1. **Enable TLS**
   ```yaml
   use_tls: "true"
   tls_skip_verify: "false"  # Verify certificates in production
   ```

2. **Restrict Network Access**
   - Use firewall rules to limit terminal access
   - Isolate payment network from other networks
   - Use private VLANs for terminals

3. **Monitor Activity**
   - Enable audit logging
   - Monitor failed transactions
   - Alert on suspicious patterns

4. **Secure Configuration**
   - Restrict config file permissions: `chmod 600 config.yaml`
   - Store credentials securely
   - Rotate terminal IDs regularly

5. **Software Updates**
   - Keep Device Bridge updated
   - Apply security patches promptly
   - Monitor security advisories

### Sensitive Data Handling

**What is masked/filtered:**
- Primary Account Number (PAN) - only last 4 digits shown
- PIN data - never logged
- CVV/CVC - never transmitted or stored
- Track data - never logged

**What is logged:**
- Transaction ID (RRN)
- Authorization code
- Response codes
- Amounts and currencies
- Terminal/merchant IDs
- Timestamps

## API Reference

### gRPC API

```protobuf
service DeviceBridge {
  // Process a payment transaction
  rpc ProcessPayment(PaymentRequest) returns (PaymentResponse);

  // Get terminal status
  rpc GetTerminalStatus(StatusRequest) returns (TerminalStatus);

  // Settle batch transactions
  rpc SettleBatch(SettlementRequest) returns (SettlementResponse);
}

message PaymentRequest {
  string device_id = 1;
  Transaction transaction = 2;
}

message Transaction {
  string type = 1;  // sale, void, refund, preauth, balance
  int64 amount = 2;
  string currency = 3;
  string reference = 4;
  string original_reference = 5;  // For void/refund
  int32 timeout = 6;  // Seconds
}
```

### REST API

#### Process Payment

```http
POST /v1/devices/{device_id}/payment
Content-Type: application/json

{
  "type": "sale",
  "amount": 10000,
  "currency": "SAR",
  "reference": "INV-001"
}
```

#### Get Terminal Status

```http
GET /v1/devices/{device_id}/status
```

Response:
```json
{
  "connected": true,
  "terminal_id": "12345678",
  "merchant_id": "123456789012345",
  "last_transaction": "2025-11-11T10:30:00Z",
  "transaction_count": 42
}
```

#### Settle Batch

```http
POST /v1/devices/{device_id}/settlement
Content-Type: application/json

{
  "batch_number": "20251111",
  "force": false
}
```

### WebSocket API

Subscribe to transaction events:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.send(JSON.stringify({
  type: 'subscribe',
  device_id: 'mada-terminal-01',
  event_types: ['transaction.completed', 'transaction.failed']
}));

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Transaction event:', data);
};
```

## Payment Providers

Device Bridge v2 includes specialized payment providers for regional networks with network-specific validation, formatting, and receipt generation.

### Mada Provider (Saudi Arabia)

The Mada provider implements Saudi Arabia's domestic payment network requirements.

**Features:**
- Currency: SAR (Saudi Riyal) with halalas as smallest unit (1 SAR = 100 halalas)
- Transaction limits: 1.00 SAR (min) to 100,000 SAR (max)
- BIN detection for 23+ Mada card ranges
- Arabic-English bilingual receipts
- Supported transactions: Sale, Void, Refund

**Configuration:**

```yaml
devices:
  - id: mada-terminal-01
    name: "Mada Terminal"
    kind: payment.tcp
    metadata:
      host: "192.168.1.100"
      port: "3000"
      terminal_id: "MADA001"
      merchant_id: "MADA_MERCH_12345"
      provider: "mada"
      use_tls: "true"
```

**Usage Example:**

```go
// Create Mada provider
config := payment_tcp.ConnectionConfig{
    Host:       "192.168.1.100",
    Port:       3000,
    TerminalID: "MADA001",
    MerchantID: "MADA_MERCH_12345",
    Provider:   "mada",
}

driver := payment_tcp.NewDriver("mada-01", "Mada Terminal", config)
provider := payment_tcp.NewMadaProvider(driver)

// Process transaction (currency defaults to SAR)
req := payment_tcp.TransactionRequest{
    Type:     payment_tcp.TransactionSale,
    Amount:   10000, // 100.00 SAR
    Currency: "SAR", // Optional, defaults to SAR
}

resp, err := provider.ProcessTransaction(ctx, req)
```

**Mada-Specific Functions:**

```go
// Format amount in halalas to SAR string
formatted := provider.FormatAmount(10000)  // Returns "100.00 SAR"

// Parse SAR string to halalas
halalas, err := provider.ParseAmount("100.00 SAR")  // Returns 10000

// Check if card is Mada
isMada := provider.isMadaCard("400861****1234")  // Returns true

// Get transaction limits
limits := provider.GetMadaTransactionLimits()
// Returns: {"min_amount_halalas": 100, "max_amount_halalas": 10000000, "currency": "SAR"}
```

### KNET Provider (Kuwait)

The KNET provider implements Kuwait's National Electronic Transfer requirements.

**Features:**
- Currency: KWD (Kuwaiti Dinar) with fils as smallest unit (1 KWD = 1000 fils)
- Transaction limits: 0.100 KWD (min) to 5,000 KWD (max)
- BIN detection for KNET Visa, MasterCard, and debit cards
- Arabic-English bilingual receipts
- 3 decimal place precision (KWD uses 3 decimals)
- Supported transactions: Sale, Void, Refund

**Configuration:**

```yaml
devices:
  - id: knet-terminal-01
    name: "KNET Terminal"
    kind: payment.tcp
    metadata:
      host: "192.168.1.101"
      port: "3000"
      terminal_id: "KNET001"
      merchant_id: "KNET_MERCH_12345"
      provider: "knet"
      use_tls: "true"
```

**Usage Example:**

```go
// Create KNET provider
config := payment_tcp.ConnectionConfig{
    Host:       "192.168.1.101",
    Port:       3000,
    TerminalID: "KNET001",
    MerchantID: "KNET_MERCH_12345",
    Provider:   "knet",
}

driver := payment_tcp.NewDriver("knet-01", "KNET Terminal", config)
provider := payment_tcp.NewKNETProvider(driver)

// Process transaction (currency defaults to KWD)
req := payment_tcp.TransactionRequest{
    Type:     payment_tcp.TransactionSale,
    Amount:   10000, // 10.000 KWD
    Currency: "KWD", // Optional, defaults to KWD
}

resp, err := provider.ProcessTransaction(ctx, req)
```

**KNET-Specific Functions:**

```go
// Format amount in fils to KWD string
formatted := provider.FormatAmount(10000)  // Returns "10.000 KWD"

// Parse KWD string to fils
fils, err := provider.ParseAmount("10.000 KWD")  // Returns 10000

// Check if card is KNET
isKNET := provider.isKNETCard("420644****1234")  // Returns true

// Get card type
cardType := provider.GetKNETCardType("420644****1234")  // Returns "knet-visa"

// Get transaction limits
limits := provider.GetKNETTransactionLimits()
// Returns: {"min_amount_fils": 100, "max_amount_fils": 5000000, "currency": "KWD", "decimal_places": 3}
```

### Provider Comparison

| Feature | Mada | KNET |
|---------|------|------|
| **Country** | Saudi Arabia | Kuwait |
| **Currency** | SAR (100 halalas) | KWD (1000 fils) |
| **Min Amount** | 1.00 SAR | 0.100 KWD |
| **Max Amount** | 100,000 SAR | 5,000 KWD |
| **Decimals** | 2 places | 3 places |
| **Pre-auth** | Supported | Not supported |
| **Balance Inquiry** | Not supported | Not supported |

## Testing

### Payment Simulator

The built-in payment simulator enables development and testing without physical terminals or network connectivity.

**Features:**
- Simulates terminal responses based on transaction amount
- Supports all transaction types (Sale, Void, Refund, Settlement)
- Configurable approval/decline logic
- No network required
- Tracks transaction history

**Configuration:**

```yaml
devices:
  - id: simulator
    name: "Payment Simulator"
    kind: payment.tcp
    metadata:
      provider: "simulator"
      terminal_id: "SIM00001"
      merchant_id: "SIM_MERCHANT"
```

**Approval Logic:**

The simulator uses the last 2 digits of the transaction amount to determine the response:

| Amount Last 2 Digits | Response Code | Message |
|---------------------|---------------|---------|
| `00` | 00 | Approved |
| `05` | 51 | Insufficient funds |
| `54` | 54 | Expired card |
| `55` | 55 | Incorrect PIN |
| `91` | 91 | Issuer unavailable |
| Other | 00 | Approved |

**Usage Example:**

```go
// Create simulator driver
driver := payment_tcp.NewSimulatorDriver("sim-01", "Test Simulator")

// Connect (always succeeds)
err := driver.Connect(context.Background())

// Test approved transaction
approvedReq := payment_tcp.TransactionRequest{
    Type:     payment_tcp.TransactionSale,
    Amount:   10000, // Ends in 00, will be approved
    Currency: "SAR",
}
approvedResp, _ := driver.ProcessTransaction(ctx, approvedReq)
// approvedResp.Success == true, approvedResp.ResponseCode == "00"

// Test declined transaction
declinedReq := payment_tcp.TransactionRequest{
    Type:     payment_tcp.TransactionSale,
    Amount:   10005, // Ends in 05, will be declined (insufficient funds)
    Currency: "SAR",
}
declinedResp, _ := driver.ProcessTransaction(ctx, declinedReq)
// declinedResp.Success == false, declinedResp.ResponseCode == "51"
```

**Test Helper:**

Use the `SimulatorTestHelper` for easier test creation:

```go
func TestPaymentFlow(t *testing.T) {
    helper := payment_tcp.NewSimulatorTestHelper()

    // Create approved transaction
    approvedReq := helper.CreateApprovedTransaction(5000)
    approvedResp, err := helper.ProcessTestTransaction(approvedReq)
    assert.NoError(t, err)
    assert.True(t, approvedResp.Success)

    // Create declined transaction
    declinedReq := helper.CreateDeclinedTransaction(5000, "insufficient_funds")
    declinedResp, err := helper.ProcessTestTransaction(declinedReq)
    assert.NoError(t, err)
    assert.False(t, declinedResp.Success)
    assert.Equal(t, "51", declinedResp.ResponseCode)
}
```

### Local Simulator

For development without a physical terminal:

```yaml
devices:
  - id: simulator
    name: "Payment Simulator"
    kind: payment.tcp
    metadata:
      host: "localhost"
      port: "3000"
      provider: "simulator"
      terminal_id: "TEST0001"
      merchant_id: "TEST"
      use_tls: "false"
```

Start the simulator:

```bash
# Install payment simulator
go install github.com/payment-simulator/simulator@latest

# Start simulator
simulator -port 3000
```

### Test Transactions

1. **Successful Purchase**
   ```bash
   # Use amount ending in 00 for approval
   curl -X POST http://localhost:8080/v1/devices/simulator/payment \
     -d '{"type":"sale","amount":10000,"currency":"SAR"}'
   ```

2. **Declined Transaction**
   ```bash
   # Use amount ending in 05 for decline
   curl -X POST http://localhost:8080/v1/devices/simulator/payment \
     -d '{"type":"sale","amount":10005,"currency":"SAR"}'
   ```

3. **Timeout Test**
   ```bash
   # Use very short timeout
   curl -X POST http://localhost:8080/v1/devices/simulator/payment \
     -d '{"type":"sale","amount":10000,"currency":"SAR","timeout":1}'
   ```

### Integration Tests

```go
func TestPaymentTransaction(t *testing.T) {
    // Create driver
    config := payment_tcp.ConnectionConfig{
        Host:       "localhost",
        Port:       3000,
        TerminalID: "TEST0001",
        MerchantID: "TEST",
        Provider:   "simulator",
    }

    driver := payment_tcp.NewDriver("test", "Test Terminal", config)

    // Connect
    err := driver.Connect(context.Background())
    require.NoError(t, err)

    // Process sale
    req := payment_tcp.TransactionRequest{
        Type:     payment_tcp.TransactionSale,
        Amount:   10000,
        Currency: "SAR",
    }

    resp, err := driver.ProcessTransaction(context.Background(), req)
    require.NoError(t, err)
    assert.True(t, resp.Success)
    assert.Equal(t, "00", resp.ResponseCode)
}
```

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to terminal

```
error: connection failed: dial tcp 192.168.1.100:3000: i/o timeout
```

**Solutions:**
1. Verify terminal IP and port: `ping 192.168.1.100`
2. Check firewall rules
3. Verify terminal is powered on
4. Check network connectivity
5. Try disabling TLS temporarily: `use_tls: "false"`

### Transaction Failures

**Problem:** Transaction declined (response code 05)

**Solutions:**
1. Check card has sufficient funds
2. Verify merchant/terminal IDs are correct
3. Check if card is expired
4. Try a different card
5. Contact acquirer for terminal status

**Problem:** Transaction timeout

```
error: transaction timeout after 60s
```

**Solutions:**
1. Increase `read_timeout` in configuration
2. Check network latency
3. Verify terminal is responsive
4. Check for terminal software updates

### TLS Certificate Errors

**Problem:** Certificate verification failed

```
error: x509: certificate signed by unknown authority
```

**Solutions:**
1. Install CA certificate on the system
2. For testing only: `tls_skip_verify: "true"`
3. Verify certificate is valid and not expired
4. Check system time is synchronized

### Response Codes

| Code | Message | Action |
|------|---------|--------|
| 00 | Approved | Success |
| 05 | Do not honor | Card declined |
| 51 | Insufficient funds | Customer needs to use different card |
| 54 | Expired card | Card is expired |
| 55 | Incorrect PIN | Customer should re-enter PIN |
| 57 | Transaction not permitted | Check terminal configuration |
| 91 | Issuer unavailable | Retry later |

Full response code list in ISO 8583 specification.

## Production Deployment

### Pre-Production Checklist

- [ ] TLS enabled and certificate verified
- [ ] Correct terminal ID and merchant ID
- [ ] Network firewall rules configured
- [ ] Config file permissions restricted (chmod 600)
- [ ] Logging level set to 'info' (not 'debug')
- [ ] Metrics and monitoring enabled
- [ ] Backup terminal configured (if applicable)
- [ ] Settlement schedule configured
- [ ] Alert rules configured
- [ ] Test transactions completed successfully
- [ ] PCI DSS compliance verified

### Deployment Configuration

```yaml
# Production configuration
server:
  grpc:
    address: ":50051"
  rest:
    address: ":8080"

devices:
  - id: mada-prod-01
    name: "Mada Production Terminal"
    kind: payment.tcp
    enabled: true
    metadata:
      host: "payment.prod.example.com"
      port: "443"
      use_tls: "true"
      tls_skip_verify: "false"
      terminal_id: "12345678"
      merchant_id: "123456789012345"
      provider: "mada"
      timeout: "30s"
      read_timeout: "60s"
      max_retries: "3"

logging:
  level: info
  format: json

telemetry:
  metrics: true
  tracing: false
```

### Monitoring

#### Metrics to Monitor

1. **Transaction Metrics**
   - Total transactions per hour
   - Success rate (approved %)
   - Average transaction time
   - Decline rate by response code

2. **System Metrics**
   - Terminal connectivity status
   - Connection failures
   - Retry attempts
   - Timeout rate

3. **Security Metrics**
   - Failed authentication attempts
   - TLS handshake failures
   - Suspicious transaction patterns

#### Alerting Rules

```yaml
alerts:
  - name: terminal_disconnected
    condition: terminal_status == "disconnected"
    duration: 5m
    severity: critical

  - name: high_decline_rate
    condition: decline_rate > 20%
    duration: 10m
    severity: warning

  - name: transaction_timeout
    condition: timeout_rate > 10%
    duration: 15m
    severity: warning
```

### Backup and High Availability

Configure backup terminal:

```yaml
devices:
  # Primary terminal
  - id: primary-terminal
    name: "Primary Terminal"
    kind: payment.tcp
    enabled: true
    metadata:
      host: "192.168.1.100"
      port: "3000"
      terminal_id: "PRIMARY1"
      merchant_id: "MERCHANT01"
      priority: "1"  # Highest priority

  # Backup terminal (used if primary fails)
  - id: backup-terminal
    name: "Backup Terminal"
    kind: payment.tcp
    enabled: true
    metadata:
      host: "192.168.1.200"
      port: "3000"
      terminal_id: "BACKUP01"
      merchant_id: "MERCHANT01"
      priority: "2"  # Lower priority
```

Application should implement failover logic to use backup terminal when primary fails.

### Performance Optimization

1. **Connection Pooling**
   - Reuse connections when possible
   - Configure appropriate timeout values
   - Monitor connection pool metrics

2. **Retry Strategy**
   - Use exponential backoff
   - Set max_retries based on SLA
   - Implement circuit breaker pattern

3. **Caching**
   - Cache terminal status
   - Cache response code mappings
   - Monitor cache hit rates

### Security Hardening

1. **Network Security**
   ```bash
   # Firewall rules (example)
   iptables -A INPUT -p tcp --dport 3000 -s 192.168.1.0/24 -j ACCEPT
   iptables -A INPUT -p tcp --dport 3000 -j DROP
   ```

2. **File Permissions**
   ```bash
   chmod 600 config.yaml
   chown devicebridge:devicebridge config.yaml
   ```

3. **Audit Logging**
   ```yaml
   logging:
     level: info
     audit: true
     audit_file: /var/log/devicebridge/audit.log
   ```

## Related Documentation

- [ISO 8583 Specification](https://en.wikipedia.org/wiki/ISO_8583)
- [PCI DSS Requirements](https://www.pcisecuritystandards.org/)
- [Device Bridge API Reference](API.md)
- [Configuration Guide](CONFIGURATION.md)

## Support

For issues or questions:
- GitHub Issues: https://github.com/Macber-eg/Flutter-Device/issues
- Documentation: https://github.com/Macber-eg/Flutter-Device/docs
- Email: support@example.com

---

**Last Updated:** 2025-11-11
**Version:** 2.0.0 (Week 7 - Foundation)
