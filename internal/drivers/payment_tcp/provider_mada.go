package payment_tcp

import (
	"context"
	"fmt"
	"strings"
)

// MadaProvider implements Mada-specific payment processing
type MadaProvider struct {
	driver *Driver
}

// NewMadaProvider creates a new Mada payment provider
func NewMadaProvider(driver *Driver) *MadaProvider {
	return &MadaProvider{
		driver: driver,
	}
}

// ProcessTransaction processes a Mada transaction with provider-specific logic
func (p *MadaProvider) ProcessTransaction(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	// Mada-specific validation
	if err := p.validateMadaTransaction(req); err != nil {
		return nil, fmt.Errorf("mada validation failed: %w", err)
	}

	// Set default currency to SAR if not specified
	if req.Currency == "" {
		req.Currency = "SAR"
	}

	// Process through driver
	response, err := p.driver.ProcessTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	// Enhance response with Mada-specific information
	p.enhanceMadaResponse(response)

	return response, nil
}

// validateMadaTransaction validates Mada-specific requirements
func (p *MadaProvider) validateMadaTransaction(req TransactionRequest) error {
	// Validate currency
	if req.Currency != "" && req.Currency != "SAR" {
		return fmt.Errorf("mada only supports SAR currency, got %s", req.Currency)
	}

	// Validate amount (minimum 1 SAR = 100 halalas)
	if req.Amount < 100 {
		return fmt.Errorf("minimum transaction amount is 1.00 SAR (100 halalas), got %d", req.Amount)
	}

	// Validate amount (maximum 100,000 SAR)
	if req.Amount > 10000000 {
		return fmt.Errorf("maximum transaction amount is 100,000 SAR, got %d halalas", req.Amount)
	}

	// Mada doesn't support certain transaction types
	switch req.Type {
	case TransactionSale, TransactionVoid, TransactionRefund, TransactionPreAuth, TransactionCompletion:
		// Supported
	case TransactionBalanceInquiry:
		// Balance inquiry not typically supported by Mada
		return fmt.Errorf("balance inquiry not supported by Mada")
	default:
		return fmt.Errorf("unsupported transaction type for Mada: %s", req.Type)
	}

	return nil
}

// enhanceMadaResponse adds Mada-specific information to the response
func (p *MadaProvider) enhanceMadaResponse(response *TransactionResponse) {
	// Detect Mada card from BIN ranges
	if response.CardNumberMasked != "" && p.isMadaCard(response.CardNumberMasked) {
		response.CardType = CardTypeMada
	}

	// Add Mada-specific metadata
	if response.Metadata == nil {
		response.Metadata = make(map[string]string)
	}
	response.Metadata["provider"] = "mada"
	response.Metadata["network"] = "Saudi Payments Network (SPAN)"
}

// isMadaCard checks if a card number is a Mada card based on BIN ranges
func (p *MadaProvider) isMadaCard(maskedNumber string) bool {
	// Mada BIN ranges (first 6 digits)
	// This is a simplified check - real implementation would have complete BIN database
	madaBins := []string{
		"400861", "401757", "407197", "407395", "409201",
		"410685", "417633", "418548", "422817", "422818",
		"422819", "428331", "428671", "428672", "428673",
		"440533", "440647", "440795", "445564", "446404",
		"446672", "457865", "968201", "968202", "968203",
		"968204", "968205", "968206", "968207", "968208",
		"968209", "968210", "968211",
	}

	// Extract BIN from masked number (e.g., "400861****1234" -> "400861")
	if len(maskedNumber) < 6 {
		return false
	}

	bin := maskedNumber[:6]
	for _, madaBin := range madaBins {
		if strings.HasPrefix(bin, madaBin) {
			return true
		}
	}

	return false
}

// Settlement processes end-of-day settlement for Mada
func (p *MadaProvider) Settlement(ctx context.Context, req SettlementRequest) (*SettlementResponse, error) {
	// Mada settlement validation
	if err := p.validateMadaSettlement(req); err != nil {
		return nil, fmt.Errorf("mada settlement validation failed: %w", err)
	}

	// Process through driver
	return p.driver.Settlement(ctx, req)
}

// validateMadaSettlement validates Mada settlement requirements
func (p *MadaProvider) validateMadaSettlement(req SettlementRequest) error {
	// Mada requires settlement at end of business day
	// No specific validation needed for basic implementation
	return nil
}

// GetTerminalStatus returns the terminal status
func (p *MadaProvider) GetTerminalStatus() TerminalStatus {
	status := p.driver.GetStatus()

	// Add Mada-specific status information
	// In a real implementation, this might query Mada-specific terminal states

	return status
}

// FormatAmount formats an amount in SAR for display
func (p *MadaProvider) FormatAmount(amount int64) string {
	// Convert halalas to SAR
	sar := float64(amount) / 100.0
	return fmt.Sprintf("%.2f SAR", sar)
}

// ParseAmount parses an amount string to halalas
func (p *MadaProvider) ParseAmount(amountStr string) (int64, error) {
	var sar float64

	// Try parsing with currency suffix
	_, err := fmt.Sscanf(amountStr, "%f SAR", &sar)
	if err != nil {
		// Try parsing as plain number
		_, err = fmt.Sscanf(amountStr, "%f", &sar)
		if err != nil {
			return 0, fmt.Errorf("invalid amount format: %s", amountStr)
		}
	}

	// Convert to halalas (smallest unit)
	halalas := int64(sar * 100)

	return halalas, nil
}

// BuildMadaReceipt builds a Mada-compliant receipt
func (p *MadaProvider) BuildMadaReceipt(response *TransactionResponse) []string {
	lines := []string{
		"================================",
		"       MADA TRANSACTION         ",
		"   Saudi Payments Network       ",
		"================================",
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
	lines = append(lines, "--------------------------------")

	// Transaction type
	txType := strings.ToUpper(string(response.TransactionID))
	if len(txType) > 20 {
		txType = txType[:20]
	}
	lines = append(lines, fmt.Sprintf("Type: PURCHASE"))

	if response.CardNumberMasked != "" {
		lines = append(lines, fmt.Sprintf("Card: %s (MADA)", response.CardNumberMasked))
	}

	if response.EntryMode != "" {
		lines = append(lines, fmt.Sprintf("Entry: %s", response.EntryMode))
	}

	lines = append(lines, "--------------------------------")
	lines = append(lines, "")

	// Amount
	lines = append(lines, fmt.Sprintf("Amount: %s", p.FormatAmount(response.Amount)))

	lines = append(lines, "")
	lines = append(lines, "--------------------------------")

	// Authorization
	if response.AuthCode != "" {
		lines = append(lines, fmt.Sprintf("Auth Code: %s", response.AuthCode))
	}

	if response.RRN != "" {
		lines = append(lines, fmt.Sprintf("RRN: %s", response.RRN))
	}

	if response.STAN != "" {
		lines = append(lines, fmt.Sprintf("Trace: %s", response.STAN))
	}

	lines = append(lines, "")
	lines = append(lines, response.Timestamp.Format("2006-01-02 15:04:05"))
	lines = append(lines, "")

	// Status
	if response.Success {
		lines = append(lines, "    *** APPROVED ***")
	} else {
		lines = append(lines, fmt.Sprintf("    *** %s ***", response.ResponseMessage))
	}

	lines = append(lines, "")
	lines = append(lines, "================================")
	lines = append(lines, "")
	lines = append(lines, " CUSTOMER COPY")
	lines = append(lines, "")
	lines = append(lines, "================================")

	return lines
}

// ValidateCardNumber validates a Mada card number (basic Luhn check)
func (p *MadaProvider) ValidateCardNumber(cardNumber string) bool {
	// Remove spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// Check length (Mada cards are 16 digits)
	if len(cardNumber) != 16 {
		return false
	}

	// Check if it's a Mada BIN
	if !p.isMadaCard(cardNumber) {
		return false
	}

	// Luhn algorithm check
	return luhnCheck(cardNumber)
}

// luhnCheck performs Luhn algorithm validation
func luhnCheck(cardNumber string) bool {
	var sum int
	parity := len(cardNumber) % 2

	for i, digit := range cardNumber {
		if digit < '0' || digit > '9' {
			return false
		}

		d := int(digit - '0')

		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
	}

	return sum%10 == 0
}

// GetMadaTransactionLimits returns Mada transaction limits
func (p *MadaProvider) GetMadaTransactionLimits() map[string]interface{} {
	return map[string]interface{}{
		"min_amount_halalas": 100,        // 1.00 SAR
		"max_amount_halalas": 10000000,   // 100,000 SAR
		"currency":           "SAR",
		"supported_types": []TransactionType{
			TransactionSale,
			TransactionVoid,
			TransactionRefund,
			TransactionPreAuth,
			TransactionCompletion,
		},
	}
}
