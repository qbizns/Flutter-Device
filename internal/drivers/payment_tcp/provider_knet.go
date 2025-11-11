package payment_tcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// KNETProvider provides KNET-specific payment processing
// KNET is the Kuwait National Electronic Transfer payment network
type KNETProvider struct {
	driver *Driver
}

// NewKNETProvider creates a new KNET provider
func NewKNETProvider(driver *Driver) *KNETProvider {
	return &KNETProvider{
		driver: driver,
	}
}

// ProcessTransaction processes a KNET transaction
func (p *KNETProvider) ProcessTransaction(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	// Validate KNET-specific requirements
	if err := p.validateKNETTransaction(req); err != nil {
		return nil, fmt.Errorf("knet validation failed: %w", err)
	}

	// Set default currency to KWD if not specified
	if req.Currency == "" {
		req.Currency = "KWD"
	}

	// Process transaction through driver
	response, err := p.driver.ProcessTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	// Enhance response with KNET-specific information
	p.enhanceKNETResponse(response)

	return response, nil
}

// Settlement processes a KNET settlement
func (p *KNETProvider) Settlement(ctx context.Context, req SettlementRequest) (*SettlementResponse, error) {
	return p.driver.Settlement(ctx, req)
}

// validateKNETTransaction validates KNET-specific transaction requirements
func (p *KNETProvider) validateKNETTransaction(req TransactionRequest) error {
	// Currency validation - KNET only supports KWD
	if req.Currency != "" && req.Currency != "KWD" {
		return fmt.Errorf("knet only supports KWD currency, got: %s", req.Currency)
	}

	// Amount validation for KWD (fils are the smallest unit, 1 KWD = 1000 fils)
	if req.Type == TransactionSale || req.Type == TransactionRefund {
		if req.Amount < 100 {
			return fmt.Errorf("minimum transaction amount is 0.100 KWD (100 fils)")
		}

		// Maximum transaction limit: 5,000 KWD
		if req.Amount > 5000000 {
			return fmt.Errorf("maximum transaction amount is 5,000 KWD")
		}
	}

	// Transaction type support
	switch req.Type {
	case TransactionSale, TransactionVoid, TransactionRefund:
		// Supported
	case TransactionBalanceInquiry:
		return fmt.Errorf("balance inquiry not supported by KNET")
	case TransactionPreAuth:
		return fmt.Errorf("pre-authorization not widely supported by KNET")
	}

	return nil
}

// enhanceKNETResponse adds KNET-specific information to response
func (p *KNETProvider) enhanceKNETResponse(response *TransactionResponse) {
	// Detect KNET card
	if p.isKNETCard(response.CardNumberMasked) {
		response.CardType = CardTypeKNET
	}

	// Add KNET-specific receipt formatting
	if response.Success {
		response.ReceiptLines = p.BuildKNETReceipt(response)
	}
}

// KNET card BIN ranges (first 6 digits)
// These are the common KNET BIN ranges used in Kuwait
var knetBins = []string{
	"420644", // KNET Visa
	"420132", // KNET Visa
	"410834", // KNET Visa
	"432328", // KNET Visa
	"529415", // KNET MasterCard
	"537767", // KNET MasterCard
	"535825", // KNET MasterCard
	"513213", // KNET MasterCard
	"401462", // KNET Debit
	"445564", // KNET Debit
	"496274", // KNET Debit
	"489318", // KNET Debit
}

// isKNETCard checks if a card is a KNET card based on BIN
func (p *KNETProvider) isKNETCard(maskedNumber string) bool {
	if len(maskedNumber) < 6 {
		return false
	}

	// Extract BIN (first 6 digits)
	bin := strings.ReplaceAll(maskedNumber[:6], "*", "")
	if len(bin) < 6 {
		// Try to get more digits if masked
		for i := 0; i < len(maskedNumber) && len(bin) < 6; i++ {
			if maskedNumber[i] >= '0' && maskedNumber[i] <= '9' {
				bin += string(maskedNumber[i])
			}
		}
	}

	// Check against known KNET BINs
	for _, knetBin := range knetBins {
		if strings.HasPrefix(bin, knetBin) || strings.HasPrefix(knetBin, bin) {
			return true
		}
	}

	return false
}

// BuildKNETReceipt builds a KNET-compliant receipt
func (p *KNETProvider) BuildKNETReceipt(response *TransactionResponse) []string {
	lines := []string{
		"================================",
		"       KNET TRANSACTION         ",
		"  Kuwait National E-Transfer   ",
		"================================",
		"",
	}

	// Merchant info
	if response.MerchantName != "" {
		lines = append(lines, response.MerchantName)
	}
	if response.MerchantID != "" {
		lines = append(lines, fmt.Sprintf("Merchant ID: %s", response.MerchantID))
	}
	if response.TerminalID != "" {
		lines = append(lines, fmt.Sprintf("Terminal ID: %s", response.TerminalID))
	}
	lines = append(lines, "")

	// Transaction details
	lines = append(lines, fmt.Sprintf("Type: %s", strings.ToUpper(string(response.Status))))
	lines = append(lines, fmt.Sprintf("Card: %s", response.CardNumberMasked))
	if response.CardType != "" {
		lines = append(lines, fmt.Sprintf("Card Type: %s", response.CardType))
	}
	lines = append(lines, "")

	// Amount in KWD
	lines = append(lines, fmt.Sprintf("Amount: %s", p.FormatAmount(response.Amount)))
	lines = append(lines, "")

	// Auth info
	if response.AuthCode != "" {
		lines = append(lines, fmt.Sprintf("Auth Code: %s", response.AuthCode))
	}
	if response.RRN != "" {
		lines = append(lines, fmt.Sprintf("RRN: %s", response.RRN))
	}
	if response.STAN != "" {
		lines = append(lines, fmt.Sprintf("STAN: %s", response.STAN))
	}
	if response.TransactionID != "" {
		lines = append(lines, fmt.Sprintf("Trans ID: %s", response.TransactionID))
	}
	lines = append(lines, "")

	// Status
	if response.Success {
		lines = append(lines, "*** APPROVED ***")
	} else {
		lines = append(lines, "*** DECLINED ***")
		if response.ResponseMessage != "" {
			lines = append(lines, response.ResponseMessage)
		}
	}
	lines = append(lines, "")

	// Timestamp
	lines = append(lines, response.Timestamp.Format("2006-01-02 15:04:05"))
	lines = append(lines, "")
	lines = append(lines, "================================")
	lines = append(lines, "  Customer Copy / نسخة العميل  ")
	lines = append(lines, "================================")

	return lines
}

// FormatAmount formats fils to KWD string
func (p *KNETProvider) FormatAmount(fils int64) string {
	// Convert fils to KWD (1 KWD = 1000 fils)
	kwd := float64(fils) / 1000.0
	return fmt.Sprintf("%.3f KWD", kwd)
}

// ParseAmount parses KWD string to fils
func (p *KNETProvider) ParseAmount(amountStr string) (int64, error) {
	// Remove currency suffix if present
	amountStr = strings.TrimSuffix(amountStr, " KWD")
	amountStr = strings.TrimSpace(amountStr)

	// Parse as float
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	// Convert to fils
	fils := int64(amount * 1000)

	return fils, nil
}

// GetKNETTransactionLimits returns KNET transaction limits
func (p *KNETProvider) GetKNETTransactionLimits() map[string]interface{} {
	return map[string]interface{}{
		"min_amount_fils":    100,      // 0.100 KWD
		"max_amount_fils":    5000000,  // 5,000 KWD
		"currency":           "KWD",
		"supported_types": []string{
			"sale",
			"void",
			"refund",
		},
		"decimal_places": 3, // KWD uses 3 decimal places
	}
}

// GetKNETCardType returns the card type for a KNET card
func (p *KNETProvider) GetKNETCardType(maskedNumber string) string {
	if len(maskedNumber) < 6 {
		return "unknown"
	}

	bin := maskedNumber[:6]

	// Check if it's a KNET card
	if !p.isKNETCard(maskedNumber) {
		return "non-knet"
	}

	// Determine card type based on BIN ranges
	// Note: These are simplified classifications
	switch {
	case strings.HasPrefix(bin, "42"), strings.HasPrefix(bin, "41"):
		return "knet-visa"
	case strings.HasPrefix(bin, "52"), strings.HasPrefix(bin, "53"), strings.HasPrefix(bin, "51"):
		return "knet-mastercard"
	case strings.HasPrefix(bin, "44"), strings.HasPrefix(bin, "48"), strings.HasPrefix(bin, "49"):
		return "knet-debit"
	default:
		return "knet"
	}
}

// BuildKNETSettlementReceipt builds a settlement receipt for KNET
func (p *KNETProvider) BuildKNETSettlementReceipt(response *SettlementResponse) []string {
	lines := []string{
		"================================",
		"      KNET SETTLEMENT           ",
		"  Kuwait National E-Transfer   ",
		"================================",
		"",
		fmt.Sprintf("Batch: %s", response.BatchNumber),
		fmt.Sprintf("Date: %s", response.Timestamp.Format("2006-01-02")),
		fmt.Sprintf("Time: %s", response.Timestamp.Format("15:04:05")),
		"",
		"--------------------------------",
		"         TRANSACTION SUMMARY     ",
		"--------------------------------",
		"",
		fmt.Sprintf("Total Transactions: %d", response.TotalCount),
		fmt.Sprintf("Approved: %d", response.ApprovedCount),
		fmt.Sprintf("Declined: %d", response.TotalCount-response.ApprovedCount),
		"",
		fmt.Sprintf("Total Amount: %s", p.FormatAmount(response.TotalAmount)),
		"",
	}

	// Status
	if response.Success {
		lines = append(lines, "*** SETTLEMENT APPROVED ***")
	} else {
		lines = append(lines, "*** SETTLEMENT FAILED ***")
		if response.ResponseMessage != "" {
			lines = append(lines, response.ResponseMessage)
		}
	}

	lines = append(lines, "")
	lines = append(lines, "================================")
	lines = append(lines, "  Merchant Copy / نسخة التاجر  ")
	lines = append(lines, "================================")

	return lines
}

// ValidateKNETBIN validates if a BIN is a valid KNET BIN
func (p *KNETProvider) ValidateKNETBIN(bin string) bool {
	if len(bin) < 6 {
		return false
	}

	bin = bin[:6]
	for _, knetBin := range knetBins {
		if bin == knetBin {
			return true
		}
	}

	return false
}

// GetSupportedCurrency returns the currency supported by KNET
func (p *KNETProvider) GetSupportedCurrency() string {
	return "KWD"
}

// GetMinAmount returns the minimum transaction amount in fils
func (p *KNETProvider) GetMinAmount() int64 {
	return 100 // 0.100 KWD
}

// GetMaxAmount returns the maximum transaction amount in fils
func (p *KNETProvider) GetMaxAmount() int64 {
	return 5000000 // 5,000 KWD
}

// IsTransactionTypeSupported checks if a transaction type is supported
func (p *KNETProvider) IsTransactionTypeSupported(txType TransactionType) bool {
	switch txType {
	case TransactionSale, TransactionVoid, TransactionRefund:
		return true
	default:
		return false
	}
}
