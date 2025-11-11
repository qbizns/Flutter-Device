package scale_serial

import (
	"fmt"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices/scale"
)

// WeightUnit represents the unit of weight measurement
type WeightUnit int

const (
	UnitGram WeightUnit = iota
	UnitKilogram
	UnitPound
	UnitOunce
	UnitMetricTon
	UnitTroyOunce
	UnitPennyweight
	UnitCarat
)

// String returns the string representation of the weight unit
func (u WeightUnit) String() string {
	switch u {
	case UnitGram:
		return "g"
	case UnitKilogram:
		return "kg"
	case UnitPound:
		return "lb"
	case UnitOunce:
		return "oz"
	case UnitMetricTon:
		return "t"
	case UnitTroyOunce:
		return "ozt"
	case UnitPennyweight:
		return "dwt"
	case UnitCarat:
		return "ct"
	default:
		return "unknown"
	}
}

// ToScaleUnit converts internal WeightUnit to scale.WeightUnit
func (u WeightUnit) ToScaleUnit() scale.WeightUnit {
	switch u {
	case UnitGram:
		return scale.UnitGram
	case UnitKilogram:
		return scale.UnitKilogram
	case UnitPound:
		return scale.UnitPound
	case UnitOunce:
		return scale.UnitOunce
	default:
		return scale.UnitKilogram // Default fallback
	}
}

// ParseWeightUnit parses a string into a WeightUnit
func ParseWeightUnit(s string) (WeightUnit, error) {
	switch s {
	case "g", "G", "gram", "grams":
		return UnitGram, nil
	case "kg", "KG", "kilogram", "kilograms":
		return UnitKilogram, nil
	case "lb", "LB", "pound", "pounds", "lbs":
		return UnitPound, nil
	case "oz", "OZ", "ounce", "ounces":
		return UnitOunce, nil
	case "t", "T", "ton", "tons":
		return UnitMetricTon, nil
	case "ozt", "OZT":
		return UnitTroyOunce, nil
	case "dwt", "DWT":
		return UnitPennyweight, nil
	case "ct", "CT", "carat":
		return UnitCarat, nil
	default:
		return 0, fmt.Errorf("unknown weight unit: %s", s)
	}
}

// WeightReading represents a weight reading from a scale
type WeightReading struct {
	Value     float64
	Unit      WeightUnit
	Stable    bool
	Timestamp time.Time
	Raw       string // Original response from scale
}

// WeightStatus represents the status of a weight reading
type WeightStatus int

const (
	StatusStable WeightStatus = iota
	StatusUnstable
	StatusOverload
	StatusUnderload
	StatusError
	StatusZeroRequired
	StatusTareRequired
)

// String returns the string representation of the weight status
func (s WeightStatus) String() string {
	switch s {
	case StatusStable:
		return "stable"
	case StatusUnstable:
		return "unstable"
	case StatusOverload:
		return "overload"
	case StatusUnderload:
		return "underload"
	case StatusError:
		return "error"
	case StatusZeroRequired:
		return "zero_required"
	case StatusTareRequired:
		return "tare_required"
	default:
		return "unknown"
	}
}

// Protocol defines the interface for serial scale communication protocols
type Protocol interface {
	// Name returns the protocol name (e.g., "MT-SICS", "CAS", "Dibal")
	Name() string

	// SendWeight sends a command to request the current weight
	SendWeight() ([]byte, error)

	// SendStableWeight sends a command to request stable weight only
	SendStableWeight() ([]byte, error)

	// SendZero sends a command to zero the scale
	SendZero() ([]byte, error)

	// SendTare sends a command to tare the scale
	SendTare() ([]byte, error)

	// SendReset sends a command to reset the scale
	SendReset() ([]byte, error)

	// ParseResponse parses a response from the scale
	ParseResponse(data []byte) (*WeightReading, WeightStatus, error)

	// RequiresCommand returns true if the protocol requires commands to be sent
	// (as opposed to continuous output protocols like Toledo 8217)
	RequiresCommand() bool

	// LineTerminator returns the line terminator for this protocol
	LineTerminator() []byte
}

// Config represents the configuration for a serial scale
type Config struct {
	// Port name (e.g., "/dev/ttyUSB0", "COM1")
	Port string

	// Baud rate (e.g., 9600, 19200)
	BaudRate int

	// Data bits (7 or 8)
	DataBits int

	// Parity ("none", "even", "odd")
	Parity string

	// Stop bits (1 or 2)
	StopBits int

	// Protocol name ("mtsics", "cas", "dibal", "toledo8217", "generic")
	Protocol string

	// Read timeout
	ReadTimeout time.Duration

	// Continuous mode (scale sends weight automatically)
	Continuous bool

	// Weight unit (for conversion if needed)
	PreferredUnit WeightUnit

	// Retry attempts for failed reads
	RetryAttempts int

	// Delay between retries
	RetryDelay time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Port:          "/dev/ttyUSB0",
		BaudRate:      9600,
		DataBits:      8,
		Parity:        "none",
		StopBits:      1,
		Protocol:      "mtsics",
		ReadTimeout:   1 * time.Second,
		Continuous:    false,
		PreferredUnit: UnitKilogram,
		RetryAttempts: 3,
		RetryDelay:    100 * time.Millisecond,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	if c.BaudRate <= 0 {
		return fmt.Errorf("baud rate must be positive")
	}

	if c.DataBits != 7 && c.DataBits != 8 {
		return fmt.Errorf("data bits must be 7 or 8")
	}

	if c.Parity != "none" && c.Parity != "even" && c.Parity != "odd" {
		return fmt.Errorf("parity must be 'none', 'even', or 'odd'")
	}

	if c.StopBits != 1 && c.StopBits != 2 {
		return fmt.Errorf("stop bits must be 1 or 2")
	}

	if c.Protocol == "" {
		return fmt.Errorf("protocol cannot be empty")
	}

	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 1 * time.Second
	}

	if c.RetryAttempts < 0 {
		c.RetryAttempts = 0
	}

	if c.RetryDelay < 0 {
		c.RetryDelay = 100 * time.Millisecond
	}

	return nil
}

// ConvertWeight converts a weight value from one unit to another
func ConvertWeight(value float64, from, to WeightUnit) float64 {
	if from == to {
		return value
	}

	// Convert to grams first
	grams := value
	switch from {
	case UnitGram:
		grams = value
	case UnitKilogram:
		grams = value * 1000
	case UnitPound:
		grams = value * 453.59237
	case UnitOunce:
		grams = value * 28.349523125
	case UnitMetricTon:
		grams = value * 1000000
	case UnitTroyOunce:
		grams = value * 31.1034768
	case UnitPennyweight:
		grams = value * 1.55517384
	case UnitCarat:
		grams = value * 0.2
	}

	// Convert from grams to target unit
	switch to {
	case UnitGram:
		return grams
	case UnitKilogram:
		return grams / 1000
	case UnitPound:
		return grams / 453.59237
	case UnitOunce:
		return grams / 28.349523125
	case UnitMetricTon:
		return grams / 1000000
	case UnitTroyOunce:
		return grams / 31.1034768
	case UnitPennyweight:
		return grams / 1.55517384
	case UnitCarat:
		return grams / 0.2
	default:
		return grams
	}
}
