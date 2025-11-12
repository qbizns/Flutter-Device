package payment_tcp

import (
	"time"
)

// TransactionType defines the type of payment transaction
type TransactionType string

const (
	TransactionSale          TransactionType = "sale"           // Purchase transaction
	TransactionVoid          TransactionType = "void"           // Void previous transaction
	TransactionRefund        TransactionType = "refund"         // Refund transaction
	TransactionPreAuth       TransactionType = "preauth"        // Pre-authorization
	TransactionCompletion    TransactionType = "completion"     // Complete pre-auth
	TransactionBalanceInquiry TransactionType = "balance"       // Balance inquiry
	TransactionSettlement    TransactionType = "settlement"     // Batch settlement
	TransactionReversal      TransactionType = "reversal"       // Transaction reversal
)

// TransactionStatus represents the current status of a transaction
type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"   // Transaction initiated
	StatusApproved  TransactionStatus = "approved"  // Transaction approved
	StatusDeclined  TransactionStatus = "declined"  // Transaction declined
	StatusCancelled TransactionStatus = "cancelled" // Transaction cancelled
	StatusTimeout   TransactionStatus = "timeout"   // Transaction timed out
	StatusError     TransactionStatus = "error"     // Transaction error
	StatusReversed  TransactionStatus = "reversed"  // Transaction reversed
)

// CardType represents the type of payment card
type CardType string

const (
	CardTypeVisa       CardType = "visa"
	CardTypeMastercard CardType = "mastercard"
	CardTypeMada       CardType = "mada"        // Saudi domestic card
	CardTypeKNET       CardType = "knet"        // Kuwait domestic card
	CardTypeBenefit    CardType = "benefit"     // Bahrain domestic card
	CardTypeAmex       CardType = "amex"
	CardTypeDiscover   CardType = "discover"
	CardTypeUnknown    CardType = "unknown"
)

// EntryMode represents how the card was presented
type EntryMode string

const (
	EntryModeManual      EntryMode = "manual"       // Manually keyed
	EntryModeSwipe       EntryMode = "swipe"        // Magnetic stripe
	EntryModeChip        EntryMode = "chip"         // EMV chip
	EntryModeContactless EntryMode = "contactless"  // NFC/contactless
	EntryModeQR          EntryMode = "qr"           // QR code payment
)

// TransactionRequest represents a payment transaction request
type TransactionRequest struct {
	Type               TransactionType `json:"type"`                          // Transaction type
	Amount             int64           `json:"amount"`                        // Amount in smallest currency unit (e.g., cents, fils)
	Currency           string          `json:"currency"`                      // ISO 4217 currency code (e.g., "SAR", "KWD")
	Reference          string          `json:"reference,omitempty"`           // Merchant reference number
	OriginalReference  string          `json:"original_reference,omitempty"`  // Original transaction reference (for void/refund)
	InvoiceNumber      string          `json:"invoice_number,omitempty"`      // Invoice/receipt number
	Description        string          `json:"description,omitempty"`         // Transaction description
	Timeout            time.Duration   `json:"timeout,omitempty"`             // Transaction timeout (default: 60s)

	// Customer identification (optional)
	CustomerID         string          `json:"customer_id,omitempty"`
	CustomerName       string          `json:"customer_name,omitempty"`

	// Additional metadata
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// TransactionResponse represents a payment transaction response
type TransactionResponse struct {
	Success            bool              `json:"success"`                       // Transaction successful
	Status             TransactionStatus `json:"status"`                        // Transaction status
	TransactionID      string            `json:"transaction_id"`                // Terminal transaction ID
	Reference          string            `json:"reference,omitempty"`           // Merchant reference
	AuthCode           string            `json:"auth_code,omitempty"`           // Authorization code
	ResponseCode       string            `json:"response_code"`                 // ISO 8583 response code
	ResponseMessage    string            `json:"response_message"`              // Human-readable message

	// Card information (masked for PCI compliance)
	CardType           CardType          `json:"card_type,omitempty"`
	CardNumberMasked   string            `json:"card_number_masked,omitempty"`  // e.g., "****1234"
	CardholderName     string            `json:"cardholder_name,omitempty"`
	ExpiryDate         string            `json:"expiry_date,omitempty"`         // YYMM format
	EntryMode          EntryMode         `json:"entry_mode,omitempty"`

	// Transaction details
	Amount             int64             `json:"amount"`                        // Transaction amount
	Currency           string            `json:"currency"`                      // Currency code
	Timestamp          time.Time         `json:"timestamp"`                     // Transaction timestamp

	// Receipt data
	MerchantName       string            `json:"merchant_name,omitempty"`
	MerchantID         string            `json:"merchant_id,omitempty"`
	TerminalID         string            `json:"terminal_id,omitempty"`
	AcquirerName       string            `json:"acquirer_name,omitempty"`

	// Additional fields
	RRN                string            `json:"rrn,omitempty"`                 // Retrieval Reference Number
	STAN               string            `json:"stan,omitempty"`                // System Trace Audit Number
	BatchNumber        string            `json:"batch_number,omitempty"`

	// Receipt lines (for printing)
	ReceiptLines       []string          `json:"receipt_lines,omitempty"`

	// Metadata
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// ISO 8583 Response Codes (common subset)
var ResponseCodes = map[string]string{
	"00": "Approved",
	"01": "Refer to card issuer",
	"02": "Refer to card issuer, special condition",
	"03": "Invalid merchant",
	"04": "Pick up card",
	"05": "Do not honor",
	"06": "Error",
	"07": "Pick up card, special condition",
	"08": "Honor with identification",
	"09": "Request in progress",
	"10": "Approved, partial",
	"11": "Approved, VIP",
	"12": "Invalid transaction",
	"13": "Invalid amount",
	"14": "Invalid card number",
	"15": "No such issuer",
	"19": "Re-enter transaction",
	"25": "Unable to locate record",
	"28": "File temporarily unavailable",
	"30": "Format error",
	"41": "Lost card, pick up",
	"43": "Stolen card, pick up",
	"51": "Insufficient funds",
	"54": "Expired card",
	"55": "Incorrect PIN",
	"57": "Transaction not permitted to cardholder",
	"58": "Transaction not permitted to terminal",
	"61": "Exceeds withdrawal amount limit",
	"62": "Restricted card",
	"63": "Security violation",
	"65": "Exceeds withdrawal frequency limit",
	"75": "Allowable PIN tries exceeded",
	"76": "Unable to locate previous message",
	"77": "Inconsistent with previous message",
	"78": "Blocked, first used",
	"79": "Already reversed",
	"80": "Invalid date",
	"81": "PIN cryptographic error",
	"82": "Incorrect CVV",
	"83": "Unable to verify PIN",
	"84": "Invalid authorization life cycle",
	"85": "Not declined",
	"86": "Cannot verify PIN",
	"87": "Purchase amount only, no cashback allowed",
	"88": "Cryptographic failure",
	"89": "Authentication failure",
	"90": "Cut-off in progress",
	"91": "Issuer or switch inoperative",
	"92": "Routing error",
	"93": "Violation of law",
	"94": "Duplicate transmission",
	"95": "Reconcile error",
	"96": "System malfunction",
}

// GetResponseMessage returns a human-readable message for a response code
func GetResponseMessage(code string) string {
	if msg, ok := ResponseCodes[code]; ok {
		return msg
	}
	return "Unknown response code"
}

// SettlementRequest represents a batch settlement request
type SettlementRequest struct {
	BatchNumber string `json:"batch_number,omitempty"` // Optional batch number
	Force       bool   `json:"force,omitempty"`        // Force settlement even if errors
}

// SettlementResponse represents a batch settlement response
type SettlementResponse struct {
	Success            bool              `json:"success"`
	BatchNumber        string            `json:"batch_number"`
	TotalCount         int               `json:"total_count"`          // Total transactions
	ApprovedCount      int               `json:"approved_count"`       // Approved transactions
	TotalAmount        int64             `json:"total_amount"`         // Total amount
	Currency           string            `json:"currency"`
	Timestamp          time.Time         `json:"timestamp"`
	ResponseCode       string            `json:"response_code"`
	ResponseMessage    string            `json:"response_message"`
	ReceiptLines       []string          `json:"receipt_lines,omitempty"`
}

// TerminalStatus represents the current status of the payment terminal
type TerminalStatus struct {
	Connected          bool              `json:"connected"`
	TerminalID         string            `json:"terminal_id,omitempty"`
	MerchantID         string            `json:"merchant_id,omitempty"`
	LastTransaction    time.Time         `json:"last_transaction,omitempty"`
	CurrentBatch       string            `json:"current_batch,omitempty"`
	TransactionCount   int               `json:"transaction_count"`
	LastError          string            `json:"last_error,omitempty"`
}

// ConnectionConfig represents the payment terminal connection configuration
type ConnectionConfig struct {
	Host               string            `json:"host"`                          // Terminal hostname/IP
	Port               int               `json:"port"`                          // Terminal port
	Timeout            time.Duration     `json:"timeout"`                       // Connection timeout
	ReadTimeout        time.Duration     `json:"read_timeout"`                  // Read timeout
	WriteTimeout       time.Duration     `json:"write_timeout"`                 // Write timeout
	UseTLS             bool              `json:"use_tls"`                       // Use TLS/SSL
	TLSSkipVerify      bool              `json:"tls_skip_verify,omitempty"`     // Skip TLS cert verification (dev only)

	// Terminal identification
	TerminalID         string            `json:"terminal_id,omitempty"`
	MerchantID         string            `json:"merchant_id,omitempty"`

	// Protocol settings
	Provider           string            `json:"provider"`                      // Provider name (mada, knet, simulator)
	MessageFormat      string            `json:"message_format,omitempty"`      // ISO 8583 format (default: "1987")

	// Retry settings
	MaxRetries         int               `json:"max_retries,omitempty"`         // Max retry attempts
	RetryDelay         time.Duration     `json:"retry_delay,omitempty"`         // Delay between retries

	// Metadata
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// Default configuration values
const (
	DefaultPort           = 3000
	DefaultTimeout        = 30 * time.Second
	DefaultReadTimeout    = 60 * time.Second
	DefaultWriteTimeout   = 10 * time.Second
	DefaultMaxRetries     = 3
	DefaultRetryDelay     = 2 * time.Second
	DefaultMessageFormat  = "1987"
)
