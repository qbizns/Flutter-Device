package payment_tcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Simulator provides a payment terminal simulator for testing
type Simulator struct {
	terminalID string
	merchantID string
	stan       int
	approved   map[string]bool // Map of transaction IDs to approval status
}

// NewSimulator creates a new payment simulator
func NewSimulator(terminalID, merchantID string) *Simulator {
	return &Simulator{
		terminalID: terminalID,
		merchantID: merchantID,
		stan:       1,
		approved:   make(map[string]bool),
	}
}

// SimulateTransaction simulates a payment transaction
func (s *Simulator) SimulateTransaction(req TransactionRequest) *TransactionResponse {
	response := &TransactionResponse{
		Timestamp:    time.Now(),
		Amount:       req.Amount,
		Currency:     req.Currency,
		TerminalID:   s.terminalID,
		MerchantID:   s.merchantID,
		MerchantName: "Test Merchant",
	}

	// Generate transaction ID
	s.stan++
	response.STAN = fmt.Sprintf("%06d", s.stan)
	// Use nanoseconds for unique transaction IDs
	response.TransactionID = fmt.Sprintf("%019d", time.Now().UnixNano())
	response.RRN = response.TransactionID

	// Simulate card details
	response.CardNumberMasked = "****1234"
	response.CardType = CardTypeMada
	response.ExpiryDate = "2512" // Dec 2025
	response.EntryMode = EntryModeChip
	response.CardholderName = "TEST CARD"

	// Determine approval based on amount
	// Amount ending in 00 = approved
	// Amount ending in 05 = declined (insufficient funds)
	// Amount ending in 54 = declined (expired card)
	// Amount ending in 55 = declined (incorrect PIN)
	// Amount ending in 91 = declined (issuer unavailable)
	// All others = approved

	lastTwoDigits := req.Amount % 100

	switch lastTwoDigits {
	case 5:
		// Declined - insufficient funds
		response.Success = false
		response.Status = StatusDeclined
		response.ResponseCode = "51"
		response.ResponseMessage = "Insufficient funds"

	case 54:
		// Declined - expired card
		response.Success = false
		response.Status = StatusDeclined
		response.ResponseCode = "54"
		response.ResponseMessage = "Expired card"

	case 55:
		// Declined - incorrect PIN
		response.Success = false
		response.Status = StatusDeclined
		response.ResponseCode = "55"
		response.ResponseMessage = "Incorrect PIN"

	case 91:
		// Declined - issuer unavailable
		response.Success = false
		response.Status = StatusDeclined
		response.ResponseCode = "91"
		response.ResponseMessage = "Issuer unavailable"

	default:
		// Approved
		response.Success = true
		response.Status = StatusApproved
		response.ResponseCode = "00"
		response.ResponseMessage = "Approved"
		response.AuthCode = s.generateAuthCode()
	}

	// For void/refund, check if original transaction exists and was approved
	if req.Type == TransactionVoid || req.Type == TransactionRefund {
		if req.OriginalReference == "" {
			response.Success = false
			response.ResponseCode = "12"
			response.ResponseMessage = "Invalid transaction"
			return response
		}

		// Check if original was approved
		if !s.approved[req.OriginalReference] {
			response.Success = false
			response.ResponseCode = "12"
			response.ResponseMessage = "Original transaction not found"
			return response
		}

		// Void/refund approved
		response.Success = true
		response.ResponseCode = "00"
		response.ResponseMessage = "Approved"
		response.AuthCode = s.generateAuthCode()

		// Remove from approved list
		delete(s.approved, req.OriginalReference)
	}

	// Store approved transactions for void/refund
	if response.Success && (req.Type == TransactionSale || req.Type == TransactionPreAuth) {
		s.approved[response.TransactionID] = true
	}

	// Build receipt
	response.ReceiptLines = s.buildSimulatedReceipt(response, req)

	return response
}

// SimulateSettlement simulates a batch settlement
func (s *Simulator) SimulateSettlement(req SettlementRequest) *SettlementResponse {
	response := &SettlementResponse{
		Success:       true,
		BatchNumber:   req.BatchNumber,
		TotalCount:    len(s.approved),
		ApprovedCount: len(s.approved),
		TotalAmount:   0,
		Currency:      "SAR",
		Timestamp:     time.Now(),
		ResponseCode:  "00",
		ResponseMessage: "Approved",
	}

	// Calculate total from approved transactions
	// In a real system, we'd track amounts
	response.TotalAmount = int64(len(s.approved)) * 1000 // Placeholder

	// Build receipt
	response.ReceiptLines = []string{
		"======= SETTLEMENT =======",
		"",
		fmt.Sprintf("Batch: %s", response.BatchNumber),
		fmt.Sprintf("Terminal: %s", s.terminalID),
		fmt.Sprintf("Merchant: %s", s.merchantID),
		"",
		fmt.Sprintf("Total Transactions: %d", response.TotalCount),
		fmt.Sprintf("Approved: %d", response.ApprovedCount),
		"",
		fmt.Sprintf("Status: %s", response.ResponseMessage),
		"",
		response.Timestamp.Format("2006-01-02 15:04:05"),
		"",
		"==========================",
	}

	// Clear approved transactions after settlement
	s.approved = make(map[string]bool)

	return response
}

// buildSimulatedReceipt builds a receipt for simulated transaction
func (s *Simulator) buildSimulatedReceipt(response *TransactionResponse, req TransactionRequest) []string {
	lines := []string{
		"======= SIMULATED RECEIPT =======",
		"",
		"*** TEST MODE - NOT FOR PRODUCTION ***",
		"",
		fmt.Sprintf("Merchant: %s", response.MerchantName),
		fmt.Sprintf("Terminal: %s", response.TerminalID),
		"",
		fmt.Sprintf("Type: %s", strings.ToUpper(string(req.Type))),
		fmt.Sprintf("Card: %s (%s)", response.CardNumberMasked, response.CardType),
		"",
		fmt.Sprintf("Amount: %.2f %s", float64(response.Amount)/100.0, response.Currency),
		"",
	}

	if response.Success {
		lines = append(lines, fmt.Sprintf("Auth Code: %s", response.AuthCode))
		lines = append(lines, fmt.Sprintf("RRN: %s", response.RRN))
		lines = append(lines, "")
		lines = append(lines, "*** APPROVED ***")
	} else {
		lines = append(lines, fmt.Sprintf("Response: %s", response.ResponseMessage))
		lines = append(lines, fmt.Sprintf("Code: %s", response.ResponseCode))
		lines = append(lines, "")
		lines = append(lines, "*** DECLINED ***")
	}

	lines = append(lines, "")
	lines = append(lines, response.Timestamp.Format("2006-01-02 15:04:05"))
	lines = append(lines, "")
	lines = append(lines, "=================================")

	return lines
}

// generateAuthCode generates a simulated authorization code
func (s *Simulator) generateAuthCode() string {
	// Generate 6-digit auth code based on timestamp
	timestamp := time.Now().Unix()
	code := timestamp % 1000000
	return fmt.Sprintf("%06d", code)
}

// SimulatorDriver wraps a driver with simulator functionality
type SimulatorDriver struct {
	driver    *Driver
	simulator *Simulator
	connected bool
}

// NewSimulatorDriver creates a new simulator-backed driver
func NewSimulatorDriver(id, name string) *SimulatorDriver {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "SIM00001",
		MerchantID: "SIM_MERCHANT",
		Provider:   "simulator",
	}

	driver := NewDriver(id, name, config)
	simulator := NewSimulator(config.TerminalID, config.MerchantID)

	return &SimulatorDriver{
		driver:    driver,
		simulator: simulator,
		connected: false,
	}
}

// Connect simulates connection (always succeeds)
func (sd *SimulatorDriver) Connect(ctx context.Context) error {
	sd.connected = true
	return nil
}

// Disconnect simulates disconnection
func (sd *SimulatorDriver) Disconnect(ctx context.Context) error {
	sd.connected = false
	return nil
}

// IsConnected returns connection status
func (sd *SimulatorDriver) IsConnected() bool {
	return sd.connected
}

// ProcessTransaction processes a transaction using the simulator
func (sd *SimulatorDriver) ProcessTransaction(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	if !sd.connected {
		return nil, fmt.Errorf("not connected")
	}

	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	// Process with simulator
	response := sd.simulator.SimulateTransaction(req)

	return response, nil
}

// Settlement processes settlement using the simulator
func (sd *SimulatorDriver) Settlement(ctx context.Context, req SettlementRequest) (*SettlementResponse, error) {
	if !sd.connected {
		return nil, fmt.Errorf("not connected")
	}

	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	// Process with simulator
	response := sd.simulator.SimulateSettlement(req)

	return response, nil
}

// GetStatus returns simulated terminal status
func (sd *SimulatorDriver) GetStatus() TerminalStatus {
	return TerminalStatus{
		Connected:        sd.connected,
		TerminalID:       sd.simulator.terminalID,
		MerchantID:       sd.simulator.merchantID,
		LastTransaction:  time.Now(),
		CurrentBatch:     time.Now().Format("20060102"),
		TransactionCount: sd.simulator.stan,
	}
}

// SetAuditLogger sets the audit logger
func (sd *SimulatorDriver) SetAuditLogger(logger *AuditLogger) {
	sd.driver.SetAuditLogger(logger)
}

// SimulatorTestHelper provides helper functions for testing
type SimulatorTestHelper struct {
	driver *SimulatorDriver
}

// NewSimulatorTestHelper creates a test helper
func NewSimulatorTestHelper() *SimulatorTestHelper {
	driver := NewSimulatorDriver("test-sim-01", "Test Simulator")
	return &SimulatorTestHelper{
		driver: driver,
	}
}

// CreateApprovedTransaction creates a transaction that will be approved
func (h *SimulatorTestHelper) CreateApprovedTransaction(amount int64) TransactionRequest {
	// Ensure amount ends in 00 for approval
	amount = (amount / 100) * 100

	return TransactionRequest{
		Type:     TransactionSale,
		Amount:   amount,
		Currency: "SAR",
		Reference: fmt.Sprintf("TEST-%d", time.Now().Unix()),
	}
}

// CreateDeclinedTransaction creates a transaction that will be declined
func (h *SimulatorTestHelper) CreateDeclinedTransaction(amount int64, reason string) TransactionRequest {
	// Adjust amount to trigger specific decline reason
	switch reason {
	case "insufficient_funds":
		amount = (amount / 100) * 100 + 5
	case "expired_card":
		amount = (amount / 100) * 100 + 54
	case "incorrect_pin":
		amount = (amount / 100) * 100 + 55
	case "issuer_unavailable":
		amount = (amount / 100) * 100 + 91
	default:
		amount = (amount / 100) * 100 + 5 // Default to insufficient funds
	}

	return TransactionRequest{
		Type:     TransactionSale,
		Amount:   amount,
		Currency: "SAR",
		Reference: fmt.Sprintf("TEST-%d", time.Now().Unix()),
	}
}

// ProcessTestTransaction processes a test transaction
func (h *SimulatorTestHelper) ProcessTestTransaction(req TransactionRequest) (*TransactionResponse, error) {
	ctx := context.Background()

	// Connect if not connected
	if !h.driver.IsConnected() {
		if err := h.driver.Connect(ctx); err != nil {
			return nil, err
		}
	}

	// Process transaction
	return h.driver.ProcessTransaction(ctx, req)
}

// GetDriver returns the simulator driver
func (h *SimulatorTestHelper) GetDriver() *SimulatorDriver {
	return h.driver
}

// ParseSimulatorAmount parses an amount string for simulator
func ParseSimulatorAmount(amountStr string) (int64, error) {
	// Remove currency suffix if present
	amountStr = strings.TrimSuffix(amountStr, " SAR")
	amountStr = strings.TrimSpace(amountStr)

	// Parse as float
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %s", amountStr)
	}

	// Convert to smallest unit (halalas)
	halalas := int64(amount * 100)

	return halalas, nil
}

// FormatSimulatorAmount formats an amount for display
func FormatSimulatorAmount(amount int64, currency string) string {
	value := float64(amount) / 100.0
	return fmt.Sprintf("%.2f %s", value, currency)
}
