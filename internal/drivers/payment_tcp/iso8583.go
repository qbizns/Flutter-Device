package payment_tcp

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ISO8583Message represents an ISO 8583 message
type ISO8583Message struct {
	MTI    string            // Message Type Indicator (4 digits)
	Bitmap []byte            // Bitmap (8 or 16 bytes)
	Fields map[int]string    // Data fields (field number -> value)
}

// MessageTypeIndicator constants
const (
	MTIAuthorizationRequest      = "0100" // Authorization request
	MTIAuthorizationResponse     = "0110" // Authorization response
	MTIFinancialRequest          = "0200" // Financial request
	MTIFinancialResponse         = "0210" // Financial response
	MTIReversalRequest           = "0400" // Reversal request
	MTIReversalResponse          = "0410" // Reversal response
	MTINetworkManagementRequest  = "0800" // Network management request
	MTINetworkManagementResponse = "0810" // Network management response
)

// Common ISO 8583 field numbers
const (
	Field2_PAN                      = 2   // Primary Account Number (card number)
	Field3_ProcessingCode           = 3   // Processing code
	Field4_Amount                   = 4   // Transaction amount
	Field7_TransmissionDateTime     = 7   // Transmission date & time
	Field11_STAN                    = 11  // System Trace Audit Number
	Field12_LocalTime               = 12  // Local transaction time
	Field13_LocalDate               = 13  // Local transaction date
	Field14_ExpirationDate          = 14  // Card expiration date
	Field15_SettlementDate          = 15  // Settlement date
	Field18_MerchantType            = 18  // Merchant category code
	Field22_EntryMode               = 22  // POS entry mode
	Field23_CardSequence            = 23  // Card sequence number
	Field25_ConditionCode           = 25  // POS condition code
	Field32_AcquiringID             = 32  // Acquiring institution ID
	Field37_RRN                     = 37  // Retrieval Reference Number
	Field38_AuthCode                = 38  // Authorization code
	Field39_ResponseCode            = 39  // Response code
	Field41_TerminalID              = 41  // Terminal ID
	Field42_MerchantID              = 42  // Merchant ID
	Field43_MerchantName            = 43  // Card acceptor name/location
	Field49_Currency                = 49  // Transaction currency code
	Field52_PIN                     = 52  // PIN data
	Field53_SecurityInfo            = 53  // Security related control information
	Field54_AdditionalAmounts       = 54  // Additional amounts
	Field55_ICC                     = 55  // ICC/Chip data
	Field60_Reserved                = 60  // Reserved for national use
	Field61_Reserved                = 61  // Reserved for private use
	Field62_Reserved                = 62  // Reserved for private use
	Field63_Reserved                = 63  // Reserved for private use
	Field64_MAC                     = 64  // Message Authentication Code
	Field95_ReplacementAmounts      = 95  // Replacement amounts
	Field102_AccountID1             = 102 // Account identification 1
	Field103_AccountID2             = 103 // Account identification 2
	Field123_POSData                = 123 // POS data code
)

// Processing codes (Field 3)
const (
	ProcessingCodePurchase        = "000000" // Purchase
	ProcessingCodeCashWithdrawal  = "010000" // Cash withdrawal
	ProcessingCodeRefund          = "200000" // Refund
	ProcessingCodeBalanceInquiry  = "300000" // Balance inquiry
	ProcessingCodeReversal        = "020000" // Reversal
)

// NewISO8583Message creates a new ISO 8583 message
func NewISO8583Message(mti string) *ISO8583Message {
	return &ISO8583Message{
		MTI:    mti,
		Bitmap: make([]byte, 8), // Primary bitmap (64 fields)
		Fields: make(map[int]string),
	}
}

// SetField sets a field value
func (m *ISO8583Message) SetField(fieldNum int, value string) {
	if fieldNum < 1 || fieldNum > 128 {
		return
	}
	m.Fields[fieldNum] = value
	m.setBit(fieldNum)
}

// GetField gets a field value
func (m *ISO8583Message) GetField(fieldNum int) (string, bool) {
	value, ok := m.Fields[fieldNum]
	return value, ok
}

// setBit sets a bit in the bitmap
func (m *ISO8583Message) setBit(fieldNum int) {
	if fieldNum < 1 || fieldNum > 64 {
		return
	}

	bytePos := (fieldNum - 1) / 8
	bitPos := 7 - ((fieldNum - 1) % 8)
	m.Bitmap[bytePos] |= (1 << uint(bitPos))
}

// isBitSet checks if a bit is set in the bitmap
func (m *ISO8583Message) isBitSet(fieldNum int) bool {
	if fieldNum < 1 || fieldNum > 64 {
		return false
	}

	bytePos := (fieldNum - 1) / 8
	bitPos := 7 - ((fieldNum - 1) % 8)
	return (m.Bitmap[bytePos] & (1 << uint(bitPos))) != 0
}

// Pack converts the message to bytes
func (m *ISO8583Message) Pack() ([]byte, error) {
	var result []byte

	// Add MTI (4 bytes)
	result = append(result, []byte(m.MTI)...)

	// Add bitmap (8 or 16 bytes in hex)
	bitmapHex := hex.EncodeToString(m.Bitmap)
	bitmapHex = strings.ToUpper(bitmapHex)
	result = append(result, []byte(bitmapHex)...)

	// Add fields in order
	for fieldNum := 2; fieldNum <= 64; fieldNum++ {
		if !m.isBitSet(fieldNum) {
			continue
		}

		value, ok := m.Fields[fieldNum]
		if !ok {
			return nil, fmt.Errorf("field %d is in bitmap but not in fields", fieldNum)
		}

		// Get field format
		format := getFieldFormat(fieldNum)

		// Encode field based on format
		fieldBytes, err := encodeField(value, format)
		if err != nil {
			return nil, fmt.Errorf("error encoding field %d: %v", fieldNum, err)
		}

		result = append(result, fieldBytes...)
	}

	return result, nil
}

// Unpack parses bytes into a message
func (m *ISO8583Message) Unpack(data []byte) error {
	if len(data) < 20 { // MTI (4) + Bitmap hex (16)
		return fmt.Errorf("message too short: %d bytes", len(data))
	}

	offset := 0

	// Parse MTI
	m.MTI = string(data[offset : offset+4])
	offset += 4

	// Parse bitmap (16 hex chars = 8 bytes)
	bitmapHex := string(data[offset : offset+16])
	bitmap, err := hex.DecodeString(bitmapHex)
	if err != nil {
		return fmt.Errorf("invalid bitmap: %v", err)
	}
	m.Bitmap = bitmap
	offset += 16

	// Parse fields
	for fieldNum := 2; fieldNum <= 64; fieldNum++ {
		if !m.isBitSet(fieldNum) {
			continue
		}

		if offset >= len(data) {
			return fmt.Errorf("unexpected end of data at field %d", fieldNum)
		}

		// Get field format
		format := getFieldFormat(fieldNum)

		// Decode field
		value, bytesRead, err := decodeField(data[offset:], format)
		if err != nil {
			return fmt.Errorf("error decoding field %d: %v", fieldNum, err)
		}

		m.Fields[fieldNum] = value
		offset += bytesRead
	}

	return nil
}

// String returns a string representation of the message
func (m *ISO8583Message) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("MTI: %s\n", m.MTI))
	sb.WriteString(fmt.Sprintf("Bitmap: %s\n", hex.EncodeToString(m.Bitmap)))

	for fieldNum := 2; fieldNum <= 64; fieldNum++ {
		if value, ok := m.Fields[fieldNum]; ok {
			// Mask sensitive fields
			displayValue := value
			if fieldNum == Field2_PAN && len(value) > 6 {
				// Mask PAN: show first 6 and last 4
				displayValue = value[:6] + strings.Repeat("*", len(value)-10) + value[len(value)-4:]
			} else if fieldNum == Field52_PIN {
				displayValue = "****"
			}
			sb.WriteString(fmt.Sprintf("Field %d: %s\n", fieldNum, displayValue))
		}
	}

	return sb.String()
}

// FieldFormat represents the format of an ISO 8583 field
type FieldFormat struct {
	Type       string // "n" (numeric), "a" (alpha), "an" (alphanumeric), "b" (binary)
	Length     int    // Fixed length
	LengthType string // "" (fixed), "LLVAR" (2-digit length prefix), "LLLVAR" (3-digit length prefix)
}

// getFieldFormat returns the format for a field number
// This is a simplified subset - real implementations have all 128 fields
func getFieldFormat(fieldNum int) FieldFormat {
	formats := map[int]FieldFormat{
		2:   {Type: "n", LengthType: "LLVAR"},     // PAN
		3:   {Type: "n", Length: 6},               // Processing code
		4:   {Type: "n", Length: 12},              // Amount
		7:   {Type: "n", Length: 10},              // Transmission date/time
		11:  {Type: "n", Length: 6},               // STAN
		12:  {Type: "n", Length: 6},               // Local time
		13:  {Type: "n", Length: 4},               // Local date
		14:  {Type: "n", Length: 4},               // Expiration date
		18:  {Type: "n", Length: 4},               // Merchant type
		22:  {Type: "n", Length: 3},               // Entry mode
		25:  {Type: "n", Length: 2},               // Condition code
		32:  {Type: "n", LengthType: "LLVAR"},     // Acquiring ID
		37:  {Type: "an", Length: 12},             // RRN
		38:  {Type: "an", Length: 6},              // Auth code
		39:  {Type: "an", Length: 2},              // Response code
		41:  {Type: "ans", Length: 8},             // Terminal ID
		42:  {Type: "ans", Length: 15},            // Merchant ID
		43:  {Type: "ans", Length: 40},            // Merchant name/location
		49:  {Type: "n", Length: 3},               // Currency code
		52:  {Type: "b", Length: 8},               // PIN
		53:  {Type: "n", Length: 16},              // Security info
		55:  {Type: "b", LengthType: "LLLVAR"},    // ICC data
		60:  {Type: "ans", LengthType: "LLLVAR"},  // Reserved
		61:  {Type: "ans", LengthType: "LLLVAR"},  // Reserved
		62:  {Type: "ans", LengthType: "LLLVAR"},  // Reserved
		63:  {Type: "ans", LengthType: "LLLVAR"},  // Reserved
	}

	if format, ok := formats[fieldNum]; ok {
		return format
	}

	// Default to LLLVAR alphanumeric
	return FieldFormat{Type: "ans", LengthType: "LLLVAR"}
}

// encodeField encodes a field value according to its format
func encodeField(value string, format FieldFormat) ([]byte, error) {
	var result []byte

	// Add length prefix if variable length
	switch format.LengthType {
	case "LLVAR":
		lengthStr := fmt.Sprintf("%02d", len(value))
		result = append(result, []byte(lengthStr)...)
	case "LLLVAR":
		lengthStr := fmt.Sprintf("%03d", len(value))
		result = append(result, []byte(lengthStr)...)
	case "":
		// Fixed length - pad if needed
		if len(value) < format.Length {
			if format.Type == "n" {
				// Pad numeric fields with leading zeros
				value = strings.Repeat("0", format.Length-len(value)) + value
			} else {
				// Pad alphanumeric fields with trailing spaces
				value = value + strings.Repeat(" ", format.Length-len(value))
			}
		}
		if len(value) > format.Length {
			value = value[:format.Length]
		}
	}

	// Add value
	if format.Type == "b" {
		// Binary field - decode from hex
		binary, err := hex.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("invalid binary data: %v", err)
		}
		result = append(result, binary...)
	} else {
		result = append(result, []byte(value)...)
	}

	return result, nil
}

// decodeField decodes a field value from bytes
func decodeField(data []byte, format FieldFormat) (string, int, error) {
	offset := 0
	var length int

	// Parse length prefix if variable length
	switch format.LengthType {
	case "LLVAR":
		if len(data) < 2 {
			return "", 0, fmt.Errorf("insufficient data for LLVAR length")
		}
		var err error
		length, err = strconv.Atoi(string(data[0:2]))
		if err != nil {
			return "", 0, fmt.Errorf("invalid LLVAR length: %v", err)
		}
		offset = 2
	case "LLLVAR":
		if len(data) < 3 {
			return "", 0, fmt.Errorf("insufficient data for LLLVAR length")
		}
		var err error
		length, err = strconv.Atoi(string(data[0:3]))
		if err != nil {
			return "", 0, fmt.Errorf("invalid LLLVAR length: %v", err)
		}
		offset = 3
	case "":
		length = format.Length
	}

	// Check if we have enough data
	if offset+length > len(data) {
		return "", 0, fmt.Errorf("insufficient data: need %d bytes, have %d", length, len(data)-offset)
	}

	// Extract value
	var value string
	if format.Type == "b" {
		// Binary field - encode as hex
		value = hex.EncodeToString(data[offset : offset+length])
		value = strings.ToUpper(value)
	} else {
		value = string(data[offset : offset+length])
		// Trim trailing spaces from alphanumeric fields
		if format.Type != "n" {
			value = strings.TrimRight(value, " ")
		}
	}

	return value, offset + length, nil
}

// Helper functions for building common messages

// BuildAuthorizationRequest builds an authorization request message
func BuildAuthorizationRequest(terminalID, merchantID, amount, currency string, stan int) *ISO8583Message {
	msg := NewISO8583Message(MTIFinancialRequest)

	// Field 3: Processing code (purchase)
	msg.SetField(Field3_ProcessingCode, ProcessingCodePurchase)

	// Field 4: Amount (12 digits, zero-padded)
	msg.SetField(Field4_Amount, fmt.Sprintf("%012s", amount))

	// Field 7: Transmission date/time (MMDDhhmmss)
	now := time.Now()
	msg.SetField(Field7_TransmissionDateTime, now.Format("0102150405"))

	// Field 11: STAN (6 digits)
	msg.SetField(Field11_STAN, fmt.Sprintf("%06d", stan))

	// Field 12: Local time (hhmmss)
	msg.SetField(Field12_LocalTime, now.Format("150405"))

	// Field 13: Local date (MMDD)
	msg.SetField(Field13_LocalDate, now.Format("0102"))

	// Field 18: Merchant type (4 digits)
	msg.SetField(Field18_MerchantType, "5999") // Default: Miscellaneous

	// Field 22: Entry mode (3 digits)
	msg.SetField(Field22_EntryMode, "051") // Chip with PIN

	// Field 25: Condition code (2 digits)
	msg.SetField(Field25_ConditionCode, "00") // Normal presentment

	// Field 41: Terminal ID
	msg.SetField(Field41_TerminalID, terminalID)

	// Field 42: Merchant ID
	msg.SetField(Field42_MerchantID, merchantID)

	// Field 49: Currency code (3 digits)
	msg.SetField(Field49_Currency, currency)

	return msg
}

// ParseResponse parses a response message and extracts key fields
func ParseResponse(data []byte) (*ISO8583Message, error) {
	msg := NewISO8583Message("")
	if err := msg.Unpack(data); err != nil {
		return nil, err
	}
	return msg, nil
}
