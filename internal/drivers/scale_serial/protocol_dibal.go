package scale_serial

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DibalProtocol implements the Dibal serial protocol
//
// Dibal is a Spanish manufacturer of retail scales, common in Europe and Latin America.
// The protocol typically runs in continuous mode, sending weight automatically.
//
// Protocol Characteristics:
//   - Continuous output (no commands required)
//   - STX/ETX framing
//   - Status indicators for stable/unstable
//   - Supports multiple units (kg, g, lb)
type DibalProtocol struct{}

// NewDibalProtocol creates a new Dibal protocol handler
func NewDibalProtocol() *DibalProtocol {
	return &DibalProtocol{}
}

// Name returns the protocol name
func (p *DibalProtocol) Name() string {
	return "Dibal"
}

// SendWeight sends a command to request the current weight
//
// Note: Most Dibal scales run in continuous mode and don't require commands.
// However, some models support on-demand weight requests.
func (p *DibalProtocol) SendWeight() ([]byte, error) {
	// Dibal: <STX>W<ETX> (some models)
	// Many Dibal scales auto-send, so this may not be needed
	return []byte{0x02, 'W', 0x03}, nil
}

// SendStableWeight sends a command to request stable weight only
func (p *DibalProtocol) SendStableWeight() ([]byte, error) {
	// Dibal doesn't distinguish in command mode
	return p.SendWeight()
}

// SendZero sends a command to zero the scale
func (p *DibalProtocol) SendZero() ([]byte, error) {
	// Dibal: <STX>Z<ETX>
	return []byte{0x02, 'Z', 0x03}, nil
}

// SendTare sends a command to tare the scale
func (p *DibalProtocol) SendTare() ([]byte, error) {
	// Dibal: <STX>T<ETX>
	return []byte{0x02, 'T', 0x03}, nil
}

// SendReset sends a command to reset the scale
func (p *DibalProtocol) SendReset() ([]byte, error) {
	// Dibal: <STX>R<ETX> (reset)
	return []byte{0x02, 'R', 0x03}, nil
}

// RequiresCommand returns false since Dibal typically runs in continuous mode
func (p *DibalProtocol) RequiresCommand() bool {
	// Most Dibal scales send weight continuously
	// Set to false for continuous mode
	return false
}

// LineTerminator returns the line terminator for Dibal
func (p *DibalProtocol) LineTerminator() []byte {
	return []byte{0x03} // ETX
}

// ParseResponse parses a response from the scale
//
// Dibal Response Format (Continuous Mode):
//   <STX><Status><Weight><Unit><ETX>
//
// Status indicators:
//   +  = Stable weight
//   -  = Unstable weight
//   O  = Overload
//   U  = Underload
//   E  = Error
//
// Weight: Variable length, typically 5-7 digits with implied decimal
// Unit: kg, g, lb, oz (2-3 characters)
//
// Examples:
//   <STX>+00123kg<ETX>     # Stable, 1.23 kg (implied 2 decimals)
//   <STX>-01500g<ETX>      # Unstable, 1500 g
//   <STX>+00055lb<ETX>     # Stable, 0.55 lb
//   <STX>O99999kg<ETX>     # Overload
//   <STX>E00000<ETX>       # Error
//
// Dibal D-900 Series Format:
//   <STX><Status><Weight><Unit><Checksum><ETX>
//   Checksum: XOR of all bytes between STX and ETX
func (p *DibalProtocol) ParseResponse(data []byte) (*WeightReading, WeightStatus, error) {
	// Check minimum length: STX + Status + Weight + Unit + ETX = at least 9 bytes
	if len(data) < 9 {
		return nil, StatusError, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// Check for STX at start
	if data[0] != 0x02 {
		return nil, StatusError, fmt.Errorf("invalid start byte: expected STX (0x02), got 0x%02x", data[0])
	}

	// Find ETX position (may have checksum before ETX in some models)
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

	// Extract content between STX and ETX
	content := string(data[1:etxPos])

	// Parse status byte
	if len(content) < 1 {
		return nil, StatusError, fmt.Errorf("empty content")
	}

	statusByte := content[0]
	var status WeightStatus
	var stable bool

	switch statusByte {
	case '+':
		status = StatusStable
		stable = true
	case '-':
		status = StatusUnstable
		stable = false
	case 'O', 'o':
		return nil, StatusOverload, fmt.Errorf("scale overload")
	case 'U', 'u':
		return nil, StatusUnderload, fmt.Errorf("scale underload")
	case 'E', 'e':
		return nil, StatusError, fmt.Errorf("scale error")
	default:
		return nil, StatusError, fmt.Errorf("unknown status byte: %c", statusByte)
	}

	// Extract weight and unit (everything after status byte)
	// Format: <Status><Weight><Unit>
	// Example: +00123kg -> 00123kg
	if len(content) < 3 {
		return nil, StatusError, fmt.Errorf("content too short for weight data")
	}

	weightData := content[1:]

	// Find where unit starts (first non-digit character after weight)
	unitStart := -1
	for i, ch := range weightData {
		if !isDigit(ch) && ch != '.' && ch != '-' && ch != '+' {
			unitStart = i
			break
		}
	}

	if unitStart == -1 {
		return nil, StatusError, fmt.Errorf("unit not found in weight data: %s", weightData)
	}

	// Extract weight and unit
	weightStr := strings.TrimSpace(weightData[:unitStart])
	unitStr := strings.TrimSpace(weightData[unitStart:])

	// Parse weight value
	weightInt, err := strconv.ParseInt(weightStr, 10, 64)
	if err != nil {
		return nil, StatusError, fmt.Errorf("failed to parse weight '%s': %w", weightStr, err)
	}

	// Dibal typically uses 2 decimal places (divide by 100)
	// So 00123 = 1.23 kg
	// However, some models use 3 decimals for grams (001500 = 1.500 kg)
	weight := float64(weightInt) / 100.0

	// Parse unit
	unit, err := ParseWeightUnit(unitStr)
	if err != nil {
		return nil, StatusError, fmt.Errorf("failed to parse unit '%s': %w", unitStr, err)
	}

	// Adjust decimal places based on unit
	// Grams typically don't need decimal adjustment
	if unit == UnitGram && weight < 10 {
		// If weight is very small for grams, likely needs no division
		weight = float64(weightInt)
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

// ParseWithChecksum parses a response with checksum validation
//
// Dibal D-900 series includes a checksum byte before ETX:
// Format: <STX><Status><Weight><Unit><Checksum><ETX>
//
// Checksum: XOR of all bytes between STX and ETX (exclusive)
func (p *DibalProtocol) ParseWithChecksum(data []byte) (*WeightReading, WeightStatus, error) {
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

	// Check if there's a checksum byte before ETX
	if etxPos < 2 {
		return nil, StatusError, fmt.Errorf("response too short for checksum")
	}

	// Extract checksum (byte before ETX)
	checksumByte := data[etxPos-1]

	// Calculate expected checksum (XOR of all bytes between STX and checksum)
	calculated := byte(0)
	for i := 1; i < etxPos-1; i++ {
		calculated ^= data[i]
	}

	if checksumByte != calculated {
		return nil, StatusError, fmt.Errorf("checksum mismatch: expected 0x%02x, got 0x%02x", calculated, checksumByte)
	}

	// Parse response without checksum byte
	// Create new slice: STX + content + ETX
	dataWithoutChecksum := append([]byte{data[0]}, data[1:etxPos-1]...)
	dataWithoutChecksum = append(dataWithoutChecksum, 0x03)

	return p.ParseResponse(dataWithoutChecksum)
}

// DibalProtocolConfig holds configuration for Dibal protocol variants
type DibalProtocolConfig struct {
	DecimalPlaces int  // Number of decimal places (0, 1, 2, or 3)
	UseChecksum   bool // Whether to validate checksum (D-900 series)
	Continuous    bool // Continuous output mode (most common)
}

// NewDibalProtocolWithConfig creates a Dibal protocol with custom configuration
func NewDibalProtocolWithConfig(config DibalProtocolConfig) *DibalProtocolConfigurable {
	return &DibalProtocolConfigurable{
		DibalProtocol: &DibalProtocol{},
		decimalPlaces: config.DecimalPlaces,
		useChecksum:   config.UseChecksum,
		continuous:    config.Continuous,
	}
}

// DibalProtocolConfigurable extends DibalProtocol with configurable options
type DibalProtocolConfigurable struct {
	*DibalProtocol
	decimalPlaces int
	useChecksum   bool
	continuous    bool
}

// RequiresCommand returns the configured mode
func (p *DibalProtocolConfigurable) RequiresCommand() bool {
	return !p.continuous
}

// ParseResponse overrides the base implementation with configurable options
func (p *DibalProtocolConfigurable) ParseResponse(data []byte) (*WeightReading, WeightStatus, error) {
	// Use checksum parsing if enabled
	if p.useChecksum {
		return p.ParseWithChecksum(data)
	}

	// Use base parsing but adjust decimal places
	reading, status, err := p.DibalProtocol.ParseResponse(data)
	if err != nil {
		return nil, status, err
	}

	// Adjust weight based on decimal places configuration
	if p.decimalPlaces != 2 {
		// Base implementation uses 2 decimal places (divide by 100)
		// Reverse the base division
		rawValue := reading.Value * 100.0

		// Apply correct decimal divisor
		divisor := 1.0
		for i := 0; i < p.decimalPlaces; i++ {
			divisor *= 10.0
		}

		if divisor > 0 {
			reading.Value = rawValue / divisor
		} else {
			reading.Value = rawValue
		}
	}

	return reading, status, nil
}

// isDigit checks if a rune is a digit
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
