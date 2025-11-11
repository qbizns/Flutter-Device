package scale_serial

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GenericProtocolConfig defines configuration for generic ASCII protocol
type GenericProtocolConfig struct {
	// Command Configuration
	WeightCommand       string // Command to request weight (e.g., "W\r\n")
	StableWeightCommand string // Command to request stable weight (e.g., "S\r\n")
	ZeroCommand         string // Command to zero the scale (e.g., "Z\r\n")
	TareCommand         string // Command to tare the scale (e.g., "T\r\n")
	ResetCommand        string // Command to reset the scale (e.g., "R\r\n")

	// Response Configuration
	ResponsePattern  string // Regex pattern to parse response (e.g., `^([SD])\s+([\d.+-]+)\s+(\w+)`)
	StatusGroup      int    // Regex group index for status (0 = no status)
	WeightGroup      int    // Regex group index for weight value (required)
	UnitGroup        int    // Regex group index for unit (0 = use default unit)
	LineTerminator   string // Line terminator (e.g., "\r\n", "\n", "\x03")
	RequiresCommands bool   // True if scale requires commands, false for continuous output

	// Status Mapping
	StableStatusValues   []string // Status values that indicate stable weight (e.g., ["S", "ST"])
	UnstableStatusValues []string // Status values that indicate unstable weight (e.g., ["D", "US"])
	ErrorStatusValues    []string // Status values that indicate error (e.g., ["E", "ER"])

	// Unit Configuration
	DefaultUnit   WeightUnit        // Default unit if not specified in response
	UnitMapping   map[string]string // Custom unit mapping (e.g., {"g": "g", "kg": "kg", "KG": "kg"})
	DecimalPlaces int               // Number of decimal places (for integer formats)

	// Parsing Options
	TrimWhitespace bool // Trim whitespace from extracted values
	CaseSensitive  bool // Case-sensitive status/unit matching
}

// DefaultGenericConfig returns a sensible default configuration
// for generic ASCII protocols (similar to MT-SICS format)
func DefaultGenericConfig() GenericProtocolConfig {
	return GenericProtocolConfig{
		WeightCommand:       "W\r\n",
		StableWeightCommand: "S\r\n",
		ZeroCommand:         "Z\r\n",
		TareCommand:         "T\r\n",
		ResetCommand:        "R\r\n",

		ResponsePattern:  `^([A-Z])\s+([\d.+-]+)\s+(\w+)`,
		StatusGroup:      1,
		WeightGroup:      2,
		UnitGroup:        3,
		LineTerminator:   "\r\n",
		RequiresCommands: true,

		StableStatusValues:   []string{"S"},
		UnstableStatusValues: []string{"D", "U"},
		ErrorStatusValues:    []string{"E"},

		DefaultUnit:   UnitGram,
		UnitMapping:   nil,
		DecimalPlaces: -1, // -1 means parse as float

		TrimWhitespace: true,
		CaseSensitive:  false,
	}
}

// GenericProtocol implements the Protocol interface for configurable ASCII protocols
type GenericProtocol struct {
	config GenericProtocolConfig
	regex  *regexp.Regexp
}

// NewGenericProtocol creates a new generic protocol with default configuration
func NewGenericProtocol() *GenericProtocol {
	return NewGenericProtocolWithConfig(DefaultGenericConfig())
}

// NewGenericProtocolWithConfig creates a new generic protocol with custom configuration
func NewGenericProtocolWithConfig(config GenericProtocolConfig) *GenericProtocol {
	regex, err := regexp.Compile(config.ResponsePattern)
	if err != nil {
		// If regex compilation fails, use a simple pattern
		regex = regexp.MustCompile(`([\d.+-]+)`)
	}

	return &GenericProtocol{
		config: config,
		regex:  regex,
	}
}

func (p *GenericProtocol) Name() string {
	return "Generic"
}

func (p *GenericProtocol) SendWeight() ([]byte, error) {
	return []byte(p.config.WeightCommand), nil
}

func (p *GenericProtocol) SendStableWeight() ([]byte, error) {
	if p.config.StableWeightCommand != "" {
		return []byte(p.config.StableWeightCommand), nil
	}
	// Fall back to regular weight command
	return p.SendWeight()
}

func (p *GenericProtocol) SendZero() ([]byte, error) {
	if p.config.ZeroCommand == "" {
		return nil, fmt.Errorf("zero command not configured")
	}
	return []byte(p.config.ZeroCommand), nil
}

func (p *GenericProtocol) SendTare() ([]byte, error) {
	if p.config.TareCommand == "" {
		return nil, fmt.Errorf("tare command not configured")
	}
	return []byte(p.config.TareCommand), nil
}

func (p *GenericProtocol) SendReset() ([]byte, error) {
	if p.config.ResetCommand == "" {
		return nil, fmt.Errorf("reset command not configured")
	}
	return []byte(p.config.ResetCommand), nil
}

func (p *GenericProtocol) ParseResponse(data []byte) (*WeightReading, WeightStatus, error) {
	if len(data) == 0 {
		return nil, StatusError, fmt.Errorf("empty response")
	}

	response := string(data)

	// Match response pattern
	matches := p.regex.FindStringSubmatch(response)
	if matches == nil || len(matches) < 2 {
		return nil, StatusError, fmt.Errorf("response does not match pattern: %s", response)
	}

	// Extract status (if configured)
	var status WeightStatus = StatusUnstable
	var stable bool = false

	if p.config.StatusGroup > 0 && p.config.StatusGroup < len(matches) {
		statusStr := matches[p.config.StatusGroup]
		if p.config.TrimWhitespace {
			statusStr = strings.TrimSpace(statusStr)
		}
		if !p.config.CaseSensitive {
			statusStr = strings.ToUpper(statusStr)
		}

		// Check status values
		if p.contains(p.config.StableStatusValues, statusStr) {
			status = StatusStable
			stable = true
		} else if p.contains(p.config.UnstableStatusValues, statusStr) {
			status = StatusUnstable
			stable = false
		} else if p.contains(p.config.ErrorStatusValues, statusStr) {
			return nil, StatusError, fmt.Errorf("scale error status: %s", statusStr)
		}
	}

	// Extract weight (required)
	if p.config.WeightGroup >= len(matches) {
		return nil, StatusError, fmt.Errorf("weight group index out of range")
	}

	weightStr := matches[p.config.WeightGroup]
	if p.config.TrimWhitespace {
		weightStr = strings.TrimSpace(weightStr)
	}

	var weight float64
	var err error

	// Parse weight based on decimal places configuration
	if p.config.DecimalPlaces >= 0 {
		// Integer format with implied decimal places
		weightInt, err := strconv.ParseInt(weightStr, 10, 64)
		if err != nil {
			return nil, StatusError, fmt.Errorf("failed to parse weight '%s': %w", weightStr, err)
		}

		divisor := 1.0
		for i := 0; i < p.config.DecimalPlaces; i++ {
			divisor *= 10
		}
		weight = float64(weightInt) / divisor
	} else {
		// Floating point format
		weight, err = strconv.ParseFloat(weightStr, 64)
		if err != nil {
			return nil, StatusError, fmt.Errorf("failed to parse weight '%s': %w", weightStr, err)
		}
	}

	// Extract unit (if configured)
	var unit WeightUnit = p.config.DefaultUnit

	if p.config.UnitGroup > 0 && p.config.UnitGroup < len(matches) {
		unitStr := matches[p.config.UnitGroup]
		if p.config.TrimWhitespace {
			unitStr = strings.TrimSpace(unitStr)
		}

		// Apply custom unit mapping if configured
		if p.config.UnitMapping != nil {
			if mappedUnit, ok := p.config.UnitMapping[unitStr]; ok {
				unitStr = mappedUnit
			}
		}

		// Parse unit
		parsedUnit, err := ParseWeightUnit(unitStr)
		if err == nil {
			unit = parsedUnit
		}
		// If parsing fails, use default unit
	}

	reading := &WeightReading{
		Value:     weight,
		Unit:      unit,
		Stable:    stable,
		Timestamp: time.Now(),
		Raw:       response,
	}

	return reading, status, nil
}

func (p *GenericProtocol) RequiresCommand() bool {
	return p.config.RequiresCommands
}

func (p *GenericProtocol) LineTerminator() []byte {
	// Convert escape sequences
	terminator := p.config.LineTerminator
	terminator = strings.ReplaceAll(terminator, "\\r", "\r")
	terminator = strings.ReplaceAll(terminator, "\\n", "\n")
	terminator = strings.ReplaceAll(terminator, "\\t", "\t")

	// Handle hex escapes like \x03
	if strings.Contains(terminator, "\\x") {
		re := regexp.MustCompile(`\\x([0-9a-fA-F]{2})`)
		terminator = re.ReplaceAllStringFunc(terminator, func(s string) string {
			hex := s[2:]
			val, _ := strconv.ParseInt(hex, 16, 32)
			return string(rune(val))
		})
	}

	return []byte(terminator)
}

// contains checks if a string is in a slice (with optional case-insensitive matching)
func (p *GenericProtocol) contains(slice []string, str string) bool {
	compareStr := str
	if !p.config.CaseSensitive {
		compareStr = strings.ToUpper(str)
	}

	for _, item := range slice {
		compareItem := item
		if !p.config.CaseSensitive {
			compareItem = strings.ToUpper(item)
		}
		if compareItem == compareStr {
			return true
		}
	}
	return false
}

// ValidateResponse checks if the response matches the expected format
func (p *GenericProtocol) ValidateResponse(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	response := string(data)
	matches := p.regex.FindStringSubmatch(response)

	return matches != nil && len(matches) >= 2
}
