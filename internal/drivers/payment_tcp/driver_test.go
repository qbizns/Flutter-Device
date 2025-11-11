package payment_tcp

import (
	"context"
	"encoding/hex"
	"testing"
	"time"
)

// TestISO8583MessagePacking tests ISO 8583 message packing
func TestISO8583MessagePacking(t *testing.T) {
	msg := NewISO8583Message(MTIFinancialRequest)

	// Set some fields
	msg.SetField(Field3_ProcessingCode, "000000")
	msg.SetField(Field4_Amount, "000000010000") // 100.00
	msg.SetField(Field11_STAN, "000001")
	msg.SetField(Field41_TerminalID, "TERM0001")
	msg.SetField(Field42_MerchantID, "MERCHANT001")

	// Pack message
	packed, err := msg.Pack()
	if err != nil {
		t.Fatalf("Pack failed: %v", err)
	}

	// Check MTI
	if string(packed[0:4]) != "0200" {
		t.Errorf("Expected MTI 0200, got %s", string(packed[0:4]))
	}

	// Unpack and verify
	msg2 := NewISO8583Message("")
	if err := msg2.Unpack(packed); err != nil {
		t.Fatalf("Unpack failed: %v", err)
	}

	if msg2.MTI != msg.MTI {
		t.Errorf("MTI mismatch: expected %s, got %s", msg.MTI, msg2.MTI)
	}

	// Verify fields
	if val, ok := msg2.GetField(Field3_ProcessingCode); !ok || val != "000000" {
		t.Errorf("Field 3 mismatch: expected 000000, got %s", val)
	}

	if val, ok := msg2.GetField(Field11_STAN); !ok || val != "000001" {
		t.Errorf("Field 11 mismatch: expected 000001, got %s", val)
	}
}

// TestISO8583FieldEncoding tests field encoding
func TestISO8583FieldEncoding(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		format   FieldFormat
		expected string
	}{
		{
			name:     "Fixed numeric",
			value:    "123",
			format:   FieldFormat{Type: "n", Length: 6},
			expected: "000123",
		},
		{
			name:     "LLVAR",
			value:    "12345",
			format:   FieldFormat{Type: "n", LengthType: "LLVAR"},
			expected: "0512345",
		},
		{
			name:     "LLLVAR",
			value:    "ABCDEF",
			format:   FieldFormat{Type: "an", LengthType: "LLLVAR"},
			expected: "006ABCDEF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := encodeField(tt.value, tt.format)
			if err != nil {
				t.Fatalf("encodeField failed: %v", err)
			}

			result := string(encoded)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestISO8583FieldDecoding tests field decoding
func TestISO8583FieldDecoding(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		format   FieldFormat
		expected string
		length   int
	}{
		{
			name:     "Fixed numeric",
			data:     "000123",
			format:   FieldFormat{Type: "n", Length: 6},
			expected: "000123",
			length:   6,
		},
		{
			name:     "LLVAR",
			data:     "0512345",
			format:   FieldFormat{Type: "n", LengthType: "LLVAR"},
			expected: "12345",
			length:   7,
		},
		{
			name:     "LLLVAR",
			data:     "006ABCDEF",
			format:   FieldFormat{Type: "an", LengthType: "LLLVAR"},
			expected: "ABCDEF",
			length:   9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead, err := decodeField([]byte(tt.data), tt.format)
			if err != nil {
				t.Fatalf("decodeField failed: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected value %s, got %s", tt.expected, value)
			}

			if bytesRead != tt.length {
				t.Errorf("Expected length %d, got %d", tt.length, bytesRead)
			}
		})
	}
}

// TestBuildAuthorizationRequest tests building authorization request
func TestBuildAuthorizationRequest(t *testing.T) {
	msg := BuildAuthorizationRequest("TERM001", "MERCH001", "10000", "784", 123)

	// Check MTI
	if msg.MTI != MTIFinancialRequest {
		t.Errorf("Expected MTI %s, got %s", MTIFinancialRequest, msg.MTI)
	}

	// Check fields
	if val, ok := msg.GetField(Field3_ProcessingCode); !ok || val != ProcessingCodePurchase {
		t.Errorf("Field 3 (processing code) mismatch: %s", val)
	}

	if _, ok := msg.GetField(Field4_Amount); !ok {
		t.Errorf("Field 4 (amount) not found")
	}

	if val, ok := msg.GetField(Field11_STAN); !ok || val != "000123" {
		t.Errorf("Field 11 (STAN) mismatch: %s", val)
	}

	if val, ok := msg.GetField(Field41_TerminalID); !ok || val != "TERM001" {
		t.Errorf("Field 41 (terminal ID) mismatch: %s", val)
	}

	if val, ok := msg.GetField(Field42_MerchantID); !ok || val != "MERCH001" {
		t.Errorf("Field 42 (merchant ID) mismatch: %s", val)
	}

	if val, ok := msg.GetField(Field49_Currency); !ok || val != "784" {
		t.Errorf("Field 49 (currency) mismatch: %s", val)
	}
}

// TestResponseCodes tests response code messages
func TestResponseCodes(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"00", "Approved"},
		{"05", "Do not honor"},
		{"51", "Insufficient funds"},
		{"54", "Expired card"},
		{"55", "Incorrect PIN"},
		{"99", "Unknown response code"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			msg := GetResponseMessage(tt.code)
			if msg != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, msg)
			}
		})
	}
}

// TestNewDriver tests driver creation
func TestNewDriver(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "TERM001",
		MerchantID: "MERCH001",
		Provider:   "simulator",
	}

	driver := NewDriver("pay-01", "Test Payment Terminal", config)

	if driver.id != "pay-01" {
		t.Errorf("Expected ID pay-01, got %s", driver.id)
	}

	if driver.name != "Test Payment Terminal" {
		t.Errorf("Expected name 'Test Payment Terminal', got %s", driver.name)
	}

	if driver.config.TerminalID != "TERM001" {
		t.Errorf("Expected terminal ID TERM001, got %s", driver.config.TerminalID)
	}

	if driver.IsConnected() {
		t.Error("Driver should not be connected initially")
	}
}

// TestDriverStatus tests getting driver status
func TestDriverStatus(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "TERM001",
		MerchantID: "MERCH001",
		Provider:   "simulator",
	}

	driver := NewDriver("pay-01", "Test Payment Terminal", config)

	status := driver.GetStatus()

	if status.Connected {
		t.Error("Status should show not connected")
	}

	if status.TerminalID != "TERM001" {
		t.Errorf("Expected terminal ID TERM001, got %s", status.TerminalID)
	}

	if status.MerchantID != "MERCH001" {
		t.Errorf("Expected merchant ID MERCH001, got %s", status.MerchantID)
	}

	if status.TransactionCount != 0 {
		t.Errorf("Expected transaction count 0, got %d", status.TransactionCount)
	}
}

// TestTransactionRequest tests transaction request structure
func TestTransactionRequest(t *testing.T) {
	req := TransactionRequest{
		Type:          TransactionSale,
		Amount:        10000, // 100.00
		Currency:      "SAR",
		Reference:     "INV-001",
		InvoiceNumber: "12345",
		Description:   "Test purchase",
		Timeout:       60 * time.Second,
	}

	if req.Type != TransactionSale {
		t.Errorf("Expected type sale, got %s", req.Type)
	}

	if req.Amount != 10000 {
		t.Errorf("Expected amount 10000, got %d", req.Amount)
	}

	if req.Currency != "SAR" {
		t.Errorf("Expected currency SAR, got %s", req.Currency)
	}
}

// TestTransactionResponse tests transaction response structure
func TestTransactionResponse(t *testing.T) {
	response := TransactionResponse{
		Success:          true,
		Status:           StatusApproved,
		TransactionID:    "123456789012",
		AuthCode:         "ABC123",
		ResponseCode:     "00",
		ResponseMessage:  "Approved",
		CardType:         CardTypeMada,
		CardNumberMasked: "****1234",
		Amount:           10000,
		Currency:         "SAR",
		Timestamp:        time.Now(),
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Status != StatusApproved {
		t.Errorf("Expected status approved, got %s", response.Status)
	}

	if response.ResponseCode != "00" {
		t.Errorf("Expected response code 00, got %s", response.ResponseCode)
	}
}

// TestBitmapOperations tests bitmap bit operations
func TestBitmapOperations(t *testing.T) {
	msg := NewISO8583Message(MTIFinancialRequest)

	// Set some fields
	msg.SetField(Field3_ProcessingCode, "000000")
	msg.SetField(Field11_STAN, "000001")
	msg.SetField(Field41_TerminalID, "TERM001")

	// Check bits are set
	if !msg.isBitSet(Field3_ProcessingCode) {
		t.Error("Bit 3 should be set")
	}

	if !msg.isBitSet(Field11_STAN) {
		t.Error("Bit 11 should be set")
	}

	if !msg.isBitSet(Field41_TerminalID) {
		t.Error("Bit 41 should be set")
	}

	// Check unset bits
	if msg.isBitSet(Field4_Amount) {
		t.Error("Bit 4 should not be set")
	}
}

// TestConnectionConfig tests connection configuration defaults
func TestConnectionConfig(t *testing.T) {
	config := ConnectionConfig{
		Host:     "localhost",
		Port:     3000,
		Provider: "simulator",
	}

	conn := NewConnection(config)

	// Check defaults are applied
	if conn.config.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, conn.config.Timeout)
	}

	if conn.config.ReadTimeout != DefaultReadTimeout {
		t.Errorf("Expected read timeout %v, got %v", DefaultReadTimeout, conn.config.ReadTimeout)
	}

	if conn.config.WriteTimeout != DefaultWriteTimeout {
		t.Errorf("Expected write timeout %v, got %v", DefaultWriteTimeout, conn.config.WriteTimeout)
	}

	if conn.config.MaxRetries != DefaultMaxRetries {
		t.Errorf("Expected max retries %d, got %d", DefaultMaxRetries, conn.config.MaxRetries)
	}
}

// TestSTANCounter tests STAN counter
func TestSTANCounter(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "TERM001",
		MerchantID: "MERCH001",
	}

	driver := NewDriver("pay-01", "Test Terminal", config)

	// Generate several STANs
	stan1 := driver.nextSTAN()
	stan2 := driver.nextSTAN()
	stan3 := driver.nextSTAN()

	if stan1 != 1 {
		t.Errorf("Expected first STAN to be 1, got %d", stan1)
	}

	if stan2 != 2 {
		t.Errorf("Expected second STAN to be 2, got %d", stan2)
	}

	if stan3 != 3 {
		t.Errorf("Expected third STAN to be 3, got %d", stan3)
	}
}

// TestCardTypes tests card type constants
func TestCardTypes(t *testing.T) {
	cards := []CardType{
		CardTypeVisa,
		CardTypeMastercard,
		CardTypeMada,
		CardTypeKNET,
		CardTypeBenefit,
		CardTypeAmex,
	}

	for _, card := range cards {
		if card == "" {
			t.Errorf("Card type should not be empty")
		}
	}
}

// TestTransactionTypes tests transaction type constants
func TestTransactionTypes(t *testing.T) {
	types := []TransactionType{
		TransactionSale,
		TransactionVoid,
		TransactionRefund,
		TransactionPreAuth,
		TransactionBalanceInquiry,
		TransactionSettlement,
	}

	for _, txType := range types {
		if txType == "" {
			t.Errorf("Transaction type should not be empty")
		}
	}
}

// BenchmarkISO8583Pack benchmarks message packing
func BenchmarkISO8583Pack(b *testing.B) {
	msg := BuildAuthorizationRequest("TERM001", "MERCH001", "10000", "784", 123)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.Pack()
	}
}

// BenchmarkISO8583Unpack benchmarks message unpacking
func BenchmarkISO8583Unpack(b *testing.B) {
	msg := BuildAuthorizationRequest("TERM001", "MERCH001", "10000", "784", 123)
	packed, _ := msg.Pack()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newMsg := NewISO8583Message("")
		_ = newMsg.Unpack(packed)
	}
}

// TestHexEncoding tests hex encoding for bitmap
func TestHexEncoding(t *testing.T) {
	bitmap := []byte{0xF2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	hexStr := hex.EncodeToString(bitmap)

	if len(hexStr) != 16 {
		t.Errorf("Expected hex string length 16, got %d", len(hexStr))
	}

	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("Hex decode failed: %v", err)
	}

	if len(decoded) != 8 {
		t.Errorf("Expected decoded length 8, got %d", len(decoded))
	}
}

// TestContextTimeout tests transaction with context timeout
func TestContextTimeout(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "TERM001",
		MerchantID: "MERCH001",
		Provider:   "simulator",
	}

	driver := NewDriver("pay-01", "Test Terminal", config)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
		Timeout:  1 * time.Millisecond,
	}

	// This should fail because we're not connected and timeout is very short
	_, err := driver.ProcessTransaction(ctx, req)
	if err == nil {
		t.Error("Expected error due to not connected, got nil")
	}
}
