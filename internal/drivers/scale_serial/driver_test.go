package scale_serial

import (
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

func TestNewDriver(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "text")
	config := DefaultConfig()
	config.Port = "/dev/ttyUSB0"

	driver, err := NewDriver("test-scale", "Test Scale", config, logger)
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}

	if driver.ID() != "test-scale" {
		t.Errorf("expected ID 'test-scale', got '%s'", driver.ID())
	}

	if driver.Name() != "Test Scale" {
		t.Errorf("expected name 'Test Scale', got '%s'", driver.Name())
	}

	if driver.Kind() != "scale.serial" {
		t.Errorf("expected kind 'scale.serial', got '%s'", driver.Kind())
	}
}

func TestNewDriver_InvalidConfig(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "text")
	config := DefaultConfig()
	config.Port = "" // Invalid: empty port

	_, err := NewDriver("test-scale", "Test Scale", config, logger)
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

func TestNewDriver_UnknownProtocol(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "text")
	config := DefaultConfig()
	config.Protocol = "unknown"

	_, err := NewDriver("test-scale", "Test Scale", config, logger)
	if err == nil {
		t.Error("expected error for unknown protocol, got nil")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name:      "valid config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "empty port",
			config: Config{
				Port:     "",
				BaudRate: 9600,
			},
			wantError: true,
		},
		{
			name: "invalid baud rate",
			config: Config{
				Port:     "/dev/ttyUSB0",
				BaudRate: -1,
			},
			wantError: true,
		},
		{
			name: "invalid data bits",
			config: Config{
				Port:     "/dev/ttyUSB0",
				BaudRate: 9600,
				DataBits: 6,
			},
			wantError: true,
		},
		{
			name: "invalid parity",
			config: Config{
				Port:     "/dev/ttyUSB0",
				BaudRate: 9600,
				DataBits: 8,
				Parity:   "invalid",
			},
			wantError: true,
		},
		{
			name: "invalid stop bits",
			config: Config{
				Port:     "/dev/ttyUSB0",
				BaudRate: 9600,
				DataBits: 8,
				Parity:   "none",
				StopBits: 3,
			},
			wantError: true,
		},
		{
			name: "empty protocol",
			config: Config{
				Port:     "/dev/ttyUSB0",
				BaudRate: 9600,
				DataBits: 8,
				Parity:   "none",
				StopBits: 1,
				Protocol: "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConvertWeight(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		from     WeightUnit
		to       WeightUnit
		expected float64
		delta    float64 // Tolerance for floating point comparison
	}{
		{
			name:     "gram to kilogram",
			value:    1000.0,
			from:     UnitGram,
			to:       UnitKilogram,
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "kilogram to gram",
			value:    1.5,
			from:     UnitKilogram,
			to:       UnitGram,
			expected: 1500.0,
			delta:    0.001,
		},
		{
			name:     "pound to kilogram",
			value:    2.2046,
			from:     UnitPound,
			to:       UnitKilogram,
			expected: 1.0,
			delta:    0.01,
		},
		{
			name:     "kilogram to pound",
			value:    1.0,
			from:     UnitKilogram,
			to:       UnitPound,
			expected: 2.2046,
			delta:    0.01,
		},
		{
			name:     "same unit",
			value:    100.0,
			from:     UnitGram,
			to:       UnitGram,
			expected: 100.0,
			delta:    0.001,
		},
		{
			name:     "ounce to gram",
			value:    1.0,
			from:     UnitOunce,
			to:       UnitGram,
			expected: 28.35,
			delta:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertWeight(tt.value, tt.from, tt.to)
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.delta {
				t.Errorf("expected %.4f, got %.4f (diff: %.4f > %.4f)",
					tt.expected, result, diff, tt.delta)
			}
		})
	}
}

func TestParseWeightUnit(t *testing.T) {
	tests := []struct {
		input    string
		expected WeightUnit
		wantErr  bool
	}{
		{"g", UnitGram, false},
		{"G", UnitGram, false},
		{"gram", UnitGram, false},
		{"grams", UnitGram, false},
		{"kg", UnitKilogram, false},
		{"KG", UnitKilogram, false},
		{"kilogram", UnitKilogram, false},
		{"lb", UnitPound, false},
		{"LB", UnitPound, false},
		{"pound", UnitPound, false},
		{"pounds", UnitPound, false},
		{"lbs", UnitPound, false},
		{"oz", UnitOunce, false},
		{"OZ", UnitOunce, false},
		{"ounce", UnitOunce, false},
		{"unknown", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseWeightUnit(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWeightUnit_String(t *testing.T) {
	tests := []struct {
		unit     WeightUnit
		expected string
	}{
		{UnitGram, "g"},
		{UnitKilogram, "kg"},
		{UnitPound, "lb"},
		{UnitOunce, "oz"},
		{UnitMetricTon, "t"},
		{UnitTroyOunce, "ozt"},
		{UnitPennyweight, "dwt"},
		{UnitCarat, "ct"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.unit.String()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestWeightStatus_String(t *testing.T) {
	tests := []struct {
		status   WeightStatus
		expected string
	}{
		{StatusStable, "stable"},
		{StatusUnstable, "unstable"},
		{StatusOverload, "overload"},
		{StatusUnderload, "underload"},
		{StatusError, "error"},
		{StatusZeroRequired, "zero_required"},
		{StatusTareRequired, "tare_required"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Port != "/dev/ttyUSB0" {
		t.Errorf("expected default port '/dev/ttyUSB0', got '%s'", config.Port)
	}

	if config.BaudRate != 9600 {
		t.Errorf("expected default baud rate 9600, got %d", config.BaudRate)
	}

	if config.DataBits != 8 {
		t.Errorf("expected default data bits 8, got %d", config.DataBits)
	}

	if config.Parity != "none" {
		t.Errorf("expected default parity 'none', got '%s'", config.Parity)
	}

	if config.StopBits != 1 {
		t.Errorf("expected default stop bits 1, got %d", config.StopBits)
	}

	if config.Protocol != "mtsics" {
		t.Errorf("expected default protocol 'mtsics', got '%s'", config.Protocol)
	}

	if config.ReadTimeout != 1*time.Second {
		t.Errorf("expected default read timeout 1s, got %v", config.ReadTimeout)
	}

	// Validate default config
	if err := config.Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
}

func TestCreateProtocol(t *testing.T) {
	tests := []struct {
		name      string
		protocol  string
		wantError bool
	}{
		{"mtsics", "mtsics", false},
		{"mt-sics", "mt-sics", false},
		{"mettler", "mettler", false},
		{"cas", "cas", false},
		{"generic", "generic", false},
		{"unknown", "unknown", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto, err := createProtocol(tt.protocol)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if proto != nil {
					t.Error("expected nil protocol, got non-nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if proto == nil {
				t.Error("expected non-nil protocol, got nil")
			}
		})
	}
}

func TestBytesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []byte
		b        []byte
		expected bool
	}{
		{"equal", []byte{1, 2, 3}, []byte{1, 2, 3}, true},
		{"different values", []byte{1, 2, 3}, []byte{1, 2, 4}, false},
		{"different lengths", []byte{1, 2}, []byte{1, 2, 3}, false},
		{"empty", []byte{}, []byte{}, true},
		{"one empty", []byte{1}, []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Note: Start, Stop, ReadWeight, Zero, and Tare tests require serial port mocking
// or physical hardware, so they are not included here. See hardware tests for
// integration testing with actual scales.

func BenchmarkConvertWeight(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConvertWeight(1000.0, UnitGram, UnitKilogram)
	}
}

func BenchmarkParseWeightUnit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseWeightUnit("kg")
	}
}
