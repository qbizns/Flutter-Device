package payment_tcp

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Driver represents a payment terminal driver
type Driver struct {
	id           string
	name         string
	config       ConnectionConfig
	connection   *Connection
	stan         int32 // System Trace Audit Number (atomic counter)
	currentBatch string
	txCount      int
	mu           sync.Mutex
	lastError    error
}

// NewDriver creates a new payment terminal driver
func NewDriver(id, name string, config ConnectionConfig) *Driver {
	return &Driver{
		id:           id,
		name:         name,
		config:       config,
		connection:   NewConnection(config),
		stan:         0, // Start at 0, nextSTAN() will increment to 1
		currentBatch: generateBatchNumber(),
		txCount:      0,
	}
}

// Connect establishes a connection to the payment terminal
func (d *Driver) Connect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.connection.Connect(); err != nil {
		d.lastError = err
		return fmt.Errorf("connection failed: %w", err)
	}

	// Test connectivity with ping
	if err := d.connection.Ping(); err != nil {
		d.lastError = err
		// Close connection on ping failure
		_ = d.connection.Disconnect()
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// Disconnect closes the connection to the payment terminal
func (d *Driver) Disconnect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.connection.Disconnect(); err != nil {
		d.lastError = err
		return fmt.Errorf("disconnect failed: %w", err)
	}

	return nil
}

// IsConnected returns true if the driver is connected to the terminal
func (d *Driver) IsConnected() bool {
	return d.connection.IsConnected()
}

// ProcessTransaction processes a payment transaction
func (d *Driver) ProcessTransaction(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	// Check connection
	if !d.connection.IsConnected() {
		return nil, fmt.Errorf("not connected to terminal")
	}

	// Set default timeout if not provided
	if req.Timeout == 0 {
		req.Timeout = 60 * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	// Route to appropriate transaction handler
	switch req.Type {
	case TransactionSale:
		return d.processSale(ctx, req)
	case TransactionVoid:
		return d.processVoid(ctx, req)
	case TransactionRefund:
		return d.processRefund(ctx, req)
	case TransactionPreAuth:
		return d.processPreAuth(ctx, req)
	case TransactionBalanceInquiry:
		return d.processBalanceInquiry(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported transaction type: %s", req.Type)
	}
}

// processSale processes a sale (purchase) transaction
func (d *Driver) processSale(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	// Generate STAN
	stan := d.nextSTAN()

	// Build ISO 8583 message
	msg := BuildAuthorizationRequest(
		d.config.TerminalID,
		d.config.MerchantID,
		fmt.Sprintf("%d", req.Amount),
		req.Currency,
		stan,
	)

	// Add transaction-specific fields
	if req.Reference != "" {
		msg.SetField(Field37_RRN, req.Reference)
	}

	// Pack message
	msgBytes, err := msg.Pack()
	if err != nil {
		d.mu.Lock()
		d.lastError = err
		d.mu.Unlock()
		return nil, fmt.Errorf("error packing message: %w", err)
	}

	// Send message and get response
	responseBytes, err := d.connection.Send(msgBytes)
	if err != nil {
		d.mu.Lock()
		d.lastError = err
		d.mu.Unlock()
		return nil, fmt.Errorf("error sending message: %w", err)
	}

	// Parse response
	responseMsg, err := ParseResponse(responseBytes)
	if err != nil {
		d.mu.Lock()
		d.lastError = err
		d.mu.Unlock()
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Convert to TransactionResponse
	response := d.convertToTransactionResponse(responseMsg, req)

	// Update counters
	d.mu.Lock()
	d.txCount++
	d.lastError = nil
	d.mu.Unlock()

	return response, nil
}

// processVoid processes a void transaction
func (d *Driver) processVoid(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	if req.OriginalReference == "" {
		return nil, fmt.Errorf("original transaction reference required for void")
	}

	// Generate STAN
	stan := d.nextSTAN()

	// Build reversal message
	msg := NewISO8583Message(MTIReversalRequest)
	msg.SetField(Field3_ProcessingCode, ProcessingCodeReversal)
	msg.SetField(Field4_Amount, fmt.Sprintf("%012d", req.Amount))
	msg.SetField(Field7_TransmissionDateTime, time.Now().Format("0102150405"))
	msg.SetField(Field11_STAN, fmt.Sprintf("%06d", stan))
	msg.SetField(Field12_LocalTime, time.Now().Format("150405"))
	msg.SetField(Field13_LocalDate, time.Now().Format("0102"))
	msg.SetField(Field37_RRN, req.OriginalReference)
	msg.SetField(Field41_TerminalID, d.config.TerminalID)
	msg.SetField(Field42_MerchantID, d.config.MerchantID)
	msg.SetField(Field49_Currency, req.Currency)

	// Pack and send
	msgBytes, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("error packing void message: %w", err)
	}

	responseBytes, err := d.connection.Send(msgBytes)
	if err != nil {
		d.mu.Lock()
		d.lastError = err
		d.mu.Unlock()
		return nil, fmt.Errorf("error sending void message: %w", err)
	}

	// Parse response
	responseMsg, err := ParseResponse(responseBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing void response: %w", err)
	}

	return d.convertToTransactionResponse(responseMsg, req), nil
}

// processRefund processes a refund transaction
func (d *Driver) processRefund(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	// Similar to sale but with refund processing code
	stan := d.nextSTAN()

	msg := BuildAuthorizationRequest(
		d.config.TerminalID,
		d.config.MerchantID,
		fmt.Sprintf("%d", req.Amount),
		req.Currency,
		stan,
	)

	// Change processing code to refund
	msg.SetField(Field3_ProcessingCode, ProcessingCodeRefund)

	if req.OriginalReference != "" {
		msg.SetField(Field37_RRN, req.OriginalReference)
	}

	msgBytes, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("error packing refund message: %w", err)
	}

	responseBytes, err := d.connection.Send(msgBytes)
	if err != nil {
		d.mu.Lock()
		d.lastError = err
		d.mu.Unlock()
		return nil, fmt.Errorf("error sending refund message: %w", err)
	}

	responseMsg, err := ParseResponse(responseBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing refund response: %w", err)
	}

	return d.convertToTransactionResponse(responseMsg, req), nil
}

// processPreAuth processes a pre-authorization transaction
func (d *Driver) processPreAuth(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	// Use authorization MTI instead of financial
	stan := d.nextSTAN()

	msg := NewISO8583Message(MTIAuthorizationRequest)
	msg.SetField(Field3_ProcessingCode, ProcessingCodePurchase)
	msg.SetField(Field4_Amount, fmt.Sprintf("%012d", req.Amount))
	msg.SetField(Field7_TransmissionDateTime, time.Now().Format("0102150405"))
	msg.SetField(Field11_STAN, fmt.Sprintf("%06d", stan))
	msg.SetField(Field12_LocalTime, time.Now().Format("150405"))
	msg.SetField(Field13_LocalDate, time.Now().Format("0102"))
	msg.SetField(Field41_TerminalID, d.config.TerminalID)
	msg.SetField(Field42_MerchantID, d.config.MerchantID)
	msg.SetField(Field49_Currency, req.Currency)

	msgBytes, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("error packing preauth message: %w", err)
	}

	responseBytes, err := d.connection.Send(msgBytes)
	if err != nil {
		d.mu.Lock()
		d.lastError = err
		d.mu.Unlock()
		return nil, fmt.Errorf("error sending preauth message: %w", err)
	}

	responseMsg, err := ParseResponse(responseBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing preauth response: %w", err)
	}

	return d.convertToTransactionResponse(responseMsg, req), nil
}

// processBalanceInquiry processes a balance inquiry
func (d *Driver) processBalanceInquiry(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	stan := d.nextSTAN()

	msg := BuildAuthorizationRequest(
		d.config.TerminalID,
		d.config.MerchantID,
		"0", // No amount for balance inquiry
		req.Currency,
		stan,
	)

	// Change processing code to balance inquiry
	msg.SetField(Field3_ProcessingCode, ProcessingCodeBalanceInquiry)

	msgBytes, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("error packing balance inquiry message: %w", err)
	}

	responseBytes, err := d.connection.Send(msgBytes)
	if err != nil {
		d.mu.Lock()
		d.lastError = err
		d.mu.Unlock()
		return nil, fmt.Errorf("error sending balance inquiry message: %w", err)
	}

	responseMsg, err := ParseResponse(responseBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing balance inquiry response: %w", err)
	}

	return d.convertToTransactionResponse(responseMsg, req), nil
}

// Settlement processes a batch settlement
func (d *Driver) Settlement(ctx context.Context, req SettlementRequest) (*SettlementResponse, error) {
	// TODO: Implement settlement in Week 8
	return nil, fmt.Errorf("settlement not yet implemented")
}

// GetStatus returns the current terminal status
func (d *Driver) GetStatus() TerminalStatus {
	d.mu.Lock()
	defer d.mu.Unlock()

	status := TerminalStatus{
		Connected:        d.connection.IsConnected(),
		TerminalID:       d.config.TerminalID,
		MerchantID:       d.config.MerchantID,
		LastTransaction:  d.connection.GetLastActivity(),
		CurrentBatch:     d.currentBatch,
		TransactionCount: d.txCount,
	}

	if d.lastError != nil {
		status.LastError = d.lastError.Error()
	}

	return status
}

// convertToTransactionResponse converts an ISO 8583 response to TransactionResponse
func (d *Driver) convertToTransactionResponse(msg *ISO8583Message, req TransactionRequest) *TransactionResponse {
	response := &TransactionResponse{
		Amount:    req.Amount,
		Currency:  req.Currency,
		Timestamp: time.Now(),
	}

	// Response code (Field 39)
	if responseCode, ok := msg.GetField(Field39_ResponseCode); ok {
		response.ResponseCode = responseCode
		response.ResponseMessage = GetResponseMessage(responseCode)
		response.Success = (responseCode == "00")

		// Set status based on response code
		if responseCode == "00" {
			response.Status = StatusApproved
		} else {
			response.Status = StatusDeclined
		}
	}

	// Authorization code (Field 38)
	if authCode, ok := msg.GetField(Field38_AuthCode); ok {
		response.AuthCode = authCode
	}

	// RRN (Field 37)
	if rrn, ok := msg.GetField(Field37_RRN); ok {
		response.RRN = rrn
		response.TransactionID = rrn
	}

	// STAN (Field 11)
	if stan, ok := msg.GetField(Field11_STAN); ok {
		response.STAN = stan
	}

	// Card number (Field 2) - mask it
	if pan, ok := msg.GetField(Field2_PAN); ok {
		if len(pan) > 10 {
			response.CardNumberMasked = "****" + pan[len(pan)-4:]
		}
	}

	// Expiration date (Field 14)
	if expiry, ok := msg.GetField(Field14_ExpirationDate); ok {
		response.ExpiryDate = expiry
	}

	// Terminal ID (Field 41)
	if terminalID, ok := msg.GetField(Field41_TerminalID); ok {
		response.TerminalID = terminalID
	}

	// Merchant ID (Field 42)
	if merchantID, ok := msg.GetField(Field42_MerchantID); ok {
		response.MerchantID = merchantID
	}

	// Merchant name (Field 43)
	if merchantName, ok := msg.GetField(Field43_MerchantName); ok {
		response.MerchantName = merchantName
	}

	// Build receipt lines
	response.ReceiptLines = d.buildReceipt(response)

	return response
}

// buildReceipt builds receipt lines for printing
func (d *Driver) buildReceipt(response *TransactionResponse) []string {
	lines := []string{
		"========== RECEIPT ==========",
		"",
	}

	if response.MerchantName != "" {
		lines = append(lines, response.MerchantName)
	}

	if response.MerchantID != "" {
		lines = append(lines, fmt.Sprintf("Merchant: %s", response.MerchantID))
	}

	if response.TerminalID != "" {
		lines = append(lines, fmt.Sprintf("Terminal: %s", response.TerminalID))
	}

	lines = append(lines, "")

	if response.CardNumberMasked != "" {
		lines = append(lines, fmt.Sprintf("Card: %s", response.CardNumberMasked))
	}

	if response.ExpiryDate != "" {
		lines = append(lines, fmt.Sprintf("Expiry: %s", response.ExpiryDate))
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Amount: %s %.2f", response.Currency, float64(response.Amount)/100.0))
	lines = append(lines, "")

	if response.AuthCode != "" {
		lines = append(lines, fmt.Sprintf("Auth Code: %s", response.AuthCode))
	}

	if response.RRN != "" {
		lines = append(lines, fmt.Sprintf("RRN: %s", response.RRN))
	}

	lines = append(lines, "")
	lines = append(lines, response.Timestamp.Format("2006-01-02 15:04:05"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Status: %s", response.ResponseMessage))
	lines = append(lines, "")
	lines = append(lines, "=============================")

	return lines
}

// nextSTAN generates the next System Trace Audit Number
func (d *Driver) nextSTAN() int {
	stan := atomic.AddInt32(&d.stan, 1)
	if stan > 999999 {
		atomic.StoreInt32(&d.stan, 1)
		return 1
	}
	return int(stan)
}

// generateBatchNumber generates a batch number
func generateBatchNumber() string {
	return time.Now().Format("20060102")
}
