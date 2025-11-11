package scale_serial

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CASProtocol implements the CAS (Korean manufacturer) serial protocol
type CASProtocol struct{}

// NewCASProtocol creates a new CAS protocol handler
func NewCASProtocol() *CASProtocol {
	return &CASProtocol{}
}

// Name returns the protocol name
func (p *CASProtocol) Name() string {
	return "CAS"
}

// SendWeight sends a command to request the current weight
func (p *CASProtocol) SendWeight() ([]byte, error) {
	// CAS: <STX>W<ETX>
	return []byte{0x02, 'W', 0x03}, nil
}

// SendStableWeight sends a command to request stable weight only
func (p *CASProtocol) SendStableWeight() ([]byte, error) {
	// CAS doesn't distinguish between stable and immediate
	return p.SendWeight()
}

// SendZero sends a command to zero the scale
func (p *CASProtocol) SendZero() ([]byte, error) {
	// CAS: <STX>Z<ETX>
	return []byte{0x02, 'Z', 0x03}, nil
}

// SendTare sends a command to tare the scale
func (p *CASProtocol) SendTare() ([]byte, error) {
	// CAS: <STX>T<ETX>
	return []byte{0x02, 'T', 0x03}, nil
}

// SendReset sends a command to reset the scale
func (p *CASProtocol) SendReset() ([]byte, error) {
	// CAS doesn't have a standard reset command
	// Return zero command as a soft reset
	return p.SendZero()
}

// RequiresCommand returns true since CAS is command-based
func (p *CASProtocol) RequiresCommand() bool {
	return true
}

// LineTerminator returns the line terminator for CAS
func (p *CASProtocol) LineTerminator() []byte {
	return []byte{0x03} // ETX
}

// ParseResponse parses a response from the scale
//
// CAS Response Format:
//   <STX><Status><Weight><Unit><ETX>
//
// Status codes:
//   S = Stable
//   U = Unstable
//   E = Error
//
// Weight: 6 digits (e.g., 001234 = 123.4, implied decimal)
// Unit: kg, lb, g
//
// Examples:
//   <STX>S001234kg<ETX>   # Stable, 123.4 kg
//   <STX>U000500g<ETX>    # Unstable, 50.0 g
//   <STX>E000000<ETX>     # Error
func (p *CASProtocol) ParseResponse(data []byte) (*WeightReading, WeightStatus, error) {
	// Check minimum length: STX + Status + 6 digits + unit + ETX = at least 10 bytes
	if len(data) < 10 {
		return nil, StatusError, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// Check for STX at start
	if data[0] != 0x02 {
		return nil, StatusError, fmt.Errorf("invalid start byte: expected STX (0x02), got 0x%02x", data[0])
	}

	// Check for ETX at end
	if data[len(data)-1] != 0x03 {
		return nil, StatusError, fmt.Errorf("invalid end byte: expected ETX (0x03), got 0x%02x", data[len(data)-1])
	}

	// Extract content between STX and ETX
	content := string(data[1 : len(data)-1])

	// Parse status byte
	if len(content) < 1 {
		return nil, StatusError, fmt.Errorf("empty content")
	}

	statusByte := content[0]
	var status WeightStatus
	var stable bool

	switch statusByte {
	case 'S':
		status = StatusStable
		stable = true
	case 'U':
		status = StatusUnstable
		stable = false
	case 'E':
		return nil, StatusError, fmt.Errorf("scale error")
	default:
		return nil, StatusError, fmt.Errorf("unknown status byte: %c", statusByte)
	}

	// Extract weight and unit
	// Format after status: 6 digits + unit
	// Example: S001234kg -> 001234kg
	if len(content) < 7 {
		return nil, StatusError, fmt.Errorf("content too short for weight data")
	}

	// Extract 6-digit weight value
	weightStr := content[1:7]

	// Extract unit (everything after the 6 digits)
	unitStr := strings.TrimSpace(content[7:])

	// Parse weight value
	weightInt, err := strconv.ParseInt(weightStr, 10, 64)
	if err != nil {
		return nil, StatusError, fmt.Errorf("failed to parse weight '%s': %w", weightStr, err)
	}

	// CAS typically uses 1 decimal place (divide by 10)
	// So 001234 = 123.4
	weight := float64(weightInt) / 10.0

	// Parse unit
	unit, err := ParseWeightUnit(unitStr)
	if err != nil {
		return nil, StatusError, fmt.Errorf("failed to parse unit '%s': %w", unitStr, err)
	}

	reading := &WeightReading{
		Value:     weight,
		Unit:      unit,
		Stable:    stable,
		Timestamp: time.Now(),
		Raw:       content,
	}

	return reading, status, nil
}

// ValidateResponse checks if a response is valid CAS format
func (p *CASProtocol) ValidateResponse(data []byte) bool {
	// Check minimum length
	if len(data) < 10 {
		return false
	}

	// Check STX and ETX
	if data[0] != 0x02 || data[len(data)-1] != 0x03 {
		return false
	}

	// Check status byte
	if len(data) < 2 {
		return false
	}

	statusByte := data[1]
	validStatus := statusByte == 'S' || statusByte == 'U' || statusByte == 'E'

	return validStatus
}

// IsStableResponse checks if a response indicates stable weight
func (p *CASProtocol) IsStableResponse(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	return data[0] == 0x02 && data[1] == 'S'
}

// IsErrorResponse checks if a response indicates an error
func (p *CASProtocol) IsErrorResponse(data []byte) bool {
	if len(data) < 2 {
		return true
	}

	if data[0] != 0x02 {
		return true
	}

	return data[1] == 'E'
}

// ParseWithChecksum parses a response with optional checksum
//
// Some CAS models include a checksum byte after ETX.
// Format: <STX><Status><Weight><Unit><ETX><Checksum>
//
// The checksum is typically XOR of all bytes between STX and ETX (exclusive).
func (p *CASProtocol) ParseWithChecksum(data []byte) (*WeightReading, WeightStatus, error) {
	// Check if there's a checksum byte after ETX
	if len(data) < 11 {
		// No checksum, use standard parsing
		return p.ParseResponse(data)
	}

	// Find ETX position
	etxPos := -1
	for i, b := range data {
		if b == 0x03 {
			etxPos = i
			break
		}
	}

	if etxPos == -1 {
		return nil, StatusError, fmt.Errorf("ETX not found")
	}

	// Check if there's a byte after ETX (checksum)
	if etxPos < len(data)-1 {
		// Validate checksum
		checksum := data[etxPos+1]
		calculated := byte(0)

		// XOR all bytes between STX and ETX (exclusive)
		for i := 1; i < etxPos; i++ {
			calculated ^= data[i]
		}

		if checksum != calculated {
			return nil, StatusError, fmt.Errorf("checksum mismatch: expected 0x%02x, got 0x%02x", calculated, checksum)
		}
	}

	// Parse response up to ETX (including ETX)
	return p.ParseResponse(data[:etxPos+1])
}

// SetDecimalPlaces sets the number of decimal places for weight parsing
//
// Some CAS models use different decimal precision:
// - 0 decimal places: 001234 = 1234
// - 1 decimal place: 001234 = 123.4 (default)
// - 2 decimal places: 001234 = 12.34
//
// Note: This is not part of the Protocol interface, but provided as
// an extension for CAS-specific configuration.
type CASProtocolConfig struct {
	DecimalPlaces int  // Number of decimal places (0, 1, or 2)
	UseChecksum   bool // Whether to validate checksum
}

// NewCASProtocolWithConfig creates a CAS protocol with custom configuration
func NewCASProtocolWithConfig(config CASProtocolConfig) *CASProtocolConfigurable {
	return &CASProtocolConfigurable{
		CASProtocol:   &CASProtocol{},
		decimalPlaces: config.DecimalPlaces,
		useChecksum:   config.UseChecksum,
	}
}

// CASProtocolConfigurable extends CASProtocol with configurable options
type CASProtocolConfigurable struct {
	*CASProtocol
	decimalPlaces int
	useChecksum   bool
}

// ParseResponse overrides the base implementation with configurable decimal places
func (p *CASProtocolConfigurable) ParseResponse(data []byte) (*WeightReading, WeightStatus, error) {
	// Use checksum parsing if enabled
	if p.useChecksum {
		return p.ParseWithChecksum(data)
	}

	// Use base parsing but adjust decimal places
	reading, status, err := p.CASProtocol.ParseResponse(data)
	if err != nil {
		return nil, status, err
	}

	// Adjust weight based on decimal places configuration
	// Base implementation uses 1 decimal place (divide by 10)
	// We need to adjust if different decimal places are configured
	if p.decimalPlaces != 1 {
		// Reverse the base division
		rawValue := reading.Value * 10.0

		// Apply correct decimal divisor
		divisor := 1.0
		for i := 0; i < p.decimalPlaces; i++ {
			divisor *= 10.0
		}

		reading.Value = rawValue / divisor
	}

	return reading, status, nil
}
