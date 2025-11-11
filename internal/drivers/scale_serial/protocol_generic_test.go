package scale_serial

import (
	"testing"
)

func TestGenericProtocol_Name(t *testing.T) {
	p := NewGenericProtocol()
	if p.Name() != "Generic" {
		t.Errorf("expected 'Generic', got '%s'", p.Name())
	}
}

func TestGenericProtocol_DefaultConfig(t *testing.T) {
	config := DefaultGenericConfig()

	if config.WeightCommand != "W\r\n" {
		t.Errorf("expected weight command 'W\\r\\n', got '%s'", config.WeightCommand)
	}

	if config.ResponsePattern == "" {
		t.Error("expected non-empty response pattern")
	}

	if config.DefaultUnit != UnitGram {
		t.Errorf("expected default unit gram, got %v", config.DefaultUnit)
	}
}

func TestGenericProtocol_Commands(t *testing.T) {
	p := NewGenericProtocol()

	tests := []struct {
		name     string
		method   func() ([]byte, error)
		expected string
	}{
		{"SendWeight", p.SendWeight, "W\r\n"},
		{"SendStableWeight", p.SendStableWeight, "S\r\n"},
		{"SendZero", p.SendZero, "Z\r\n"},
		{"SendTare", p.SendTare, "T\r\n"},
		{"SendReset", p.SendReset, "R\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := tt.method()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(cmd) != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, string(cmd))
			}
		})
	}
}

func TestGenericProtocol_ParseResponse_MTSICSFormat(t *testing.T) {
	// Test with MT-SICS-like format
	config := DefaultGenericConfig()
	config.ResponsePattern = `^([SD])\s+([\d.+-]+)\s+(\w+)`
	config.StatusGroup = 1
	config.WeightGroup = 2
	config.UnitGroup = 3

	p := NewGenericProtocolWithConfig(config)

	tests := []struct {
		name       string
		response   string
		wantValue  float64
		wantUnit   WeightUnit
		wantStable bool
		wantStatus WeightStatus
		wantError  bool
	}{
		{
			name:       "stable grams",
			response:   "S       123.4 g",
			wantValue:  123.4,
			wantUnit:   UnitGram,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
		{
			name:       "unstable kilograms",
			response:   "D     5.678 kg",
			wantValue:  5.678,
			wantUnit:   UnitKilogram,
			wantStable: false,
			wantStatus: StatusUnstable,
			wantError:  false,
		},
		{
			name:       "stable pounds",
			response:   "S      10.5 lb",
			wantValue:  10.5,
			wantUnit:   UnitPound,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, status, err := p.ParseResponse([]byte(tt.response))

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reading.Value != tt.wantValue {
				t.Errorf("value: expected %f, got %f", tt.wantValue, reading.Value)
			}

			if reading.Unit != tt.wantUnit {
				t.Errorf("unit: expected %v, got %v", tt.wantUnit, reading.Unit)
			}

			if reading.Stable != tt.wantStable {
				t.Errorf("stable: expected %v, got %v", tt.wantStable, reading.Stable)
			}

			if status != tt.wantStatus {
				t.Errorf("status: expected %v, got %v", tt.wantStatus, status)
			}
		})
	}
}

func TestGenericProtocol_ParseResponse_SimpleFormat(t *testing.T) {
	// Test with simple weight-only format
	config := GenericProtocolConfig{
		WeightCommand:    "R\n",
		ResponsePattern:  `([\d.]+)`,
		StatusGroup:      0, // No status
		WeightGroup:      1,
		UnitGroup:        0, // No unit in response
		LineTerminator:   "\n",
		RequiresCommands: true,
		DefaultUnit:      UnitKilogram,
		DecimalPlaces:    -1, // Parse as float
		TrimWhitespace:   true,
	}

	p := NewGenericProtocolWithConfig(config)

	tests := []struct {
		name      string
		response  string
		wantValue float64
		wantUnit  WeightUnit
	}{
		{
			name:      "simple weight",
			response:  "123.45",
			wantValue: 123.45,
			wantUnit:  UnitKilogram,
		},
		{
			name:      "weight with whitespace",
			response:  "  567.89  ",
			wantValue: 567.89,
			wantUnit:  UnitKilogram,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, _, err := p.ParseResponse([]byte(tt.response))

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reading.Value != tt.wantValue {
				t.Errorf("value: expected %f, got %f", tt.wantValue, reading.Value)
			}

			if reading.Unit != tt.wantUnit {
				t.Errorf("unit: expected %v, got %v", tt.wantUnit, reading.Unit)
			}
		})
	}
}

func TestGenericProtocol_ParseResponse_IntegerFormat(t *testing.T) {
	// Test with integer format and implied decimal places
	config := GenericProtocolConfig{
		WeightCommand:    "W\n",
		ResponsePattern:  `(\d{6})`,
		StatusGroup:      0,
		WeightGroup:      1,
		UnitGroup:        0,
		LineTerminator:   "\n",
		RequiresCommands: true,
		DefaultUnit:      UnitGram,
		DecimalPlaces:    1, // 001234 = 123.4
		TrimWhitespace:   true,
	}

	p := NewGenericProtocolWithConfig(config)

	tests := []struct {
		name      string
		response  string
		wantValue float64
	}{
		{
			name:      "integer with 1 decimal",
			response:  "001234",
			wantValue: 123.4,
		},
		{
			name:      "zero",
			response:  "000000",
			wantValue: 0.0,
		},
		{
			name:      "large value",
			response:  "123456",
			wantValue: 12345.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, _, err := p.ParseResponse([]byte(tt.response))

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reading.Value != tt.wantValue {
				t.Errorf("value: expected %f, got %f", tt.wantValue, reading.Value)
			}
		})
	}
}

func TestGenericProtocol_ParseResponse_CustomUnitMapping(t *testing.T) {
	// Test with custom unit mapping
	config := GenericProtocolConfig{
		WeightCommand:   "W\r\n",
		ResponsePattern: `([\d.]+)\s+(\w+)`,
		StatusGroup:     0,
		WeightGroup:     1,
		UnitGroup:       2,
		LineTerminator:  "\r\n",
		DefaultUnit:     UnitGram,
		DecimalPlaces:   -1, // Parse as float
		UnitMapping: map[string]string{
			"G":  "g",
			"KG": "kg",
			"LB": "lb",
		},
		TrimWhitespace: true,
	}

	p := NewGenericProtocolWithConfig(config)

	tests := []struct {
		name      string
		response  string
		wantValue float64
		wantUnit  WeightUnit
	}{
		{
			name:      "uppercase G",
			response:  "123.4 G",
			wantValue: 123.4,
			wantUnit:  UnitGram,
		},
		{
			name:      "uppercase KG",
			response:  "5.678 KG",
			wantValue: 5.678,
			wantUnit:  UnitKilogram,
		},
		{
			name:      "uppercase LB",
			response:  "10.5 LB",
			wantValue: 10.5,
			wantUnit:  UnitPound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, _, err := p.ParseResponse([]byte(tt.response))

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reading.Value != tt.wantValue {
				t.Errorf("value: expected %f, got %f", tt.wantValue, reading.Value)
			}

			if reading.Unit != tt.wantUnit {
				t.Errorf("unit: expected %v, got %v", tt.wantUnit, reading.Unit)
			}
		})
	}
}

func TestGenericProtocol_ParseResponse_Errors(t *testing.T) {
	p := NewGenericProtocol()

	tests := []struct {
		name     string
		response string
	}{
		{
			name:     "empty response",
			response: "",
		},
		{
			name:     "no match",
			response: "INVALID FORMAT",
		},
		{
			name:     "error status",
			response: "E 0.0 g",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, _, err := p.ParseResponse([]byte(tt.response))

			if err == nil {
				t.Error("expected error, got nil")
			}

			if reading != nil {
				t.Errorf("expected nil reading, got %+v", reading)
			}
		})
	}
}

func TestGenericProtocol_LineTerminator(t *testing.T) {
	tests := []struct {
		name       string
		terminator string
		expected   []byte
	}{
		{
			name:       "CR+LF",
			terminator: "\\r\\n",
			expected:   []byte{'\r', '\n'},
		},
		{
			name:       "LF only",
			terminator: "\\n",
			expected:   []byte{'\n'},
		},
		{
			name:       "CR only",
			terminator: "\\r",
			expected:   []byte{'\r'},
		},
		{
			name:       "ETX (hex)",
			terminator: "\\x03",
			expected:   []byte{0x03},
		},
		{
			name:       "STX+ETX",
			terminator: "\\x02\\x03",
			expected:   []byte{0x02, 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultGenericConfig()
			config.LineTerminator = tt.terminator
			p := NewGenericProtocolWithConfig(config)

			term := p.LineTerminator()

			if len(term) != len(tt.expected) {
				t.Errorf("length: expected %d, got %d", len(tt.expected), len(term))
			}

			for i := range term {
				if term[i] != tt.expected[i] {
					t.Errorf("byte %d: expected 0x%02x, got 0x%02x", i, tt.expected[i], term[i])
				}
			}
		})
	}
}

func TestGenericProtocol_RequiresCommand(t *testing.T) {
	tests := []struct {
		name     string
		requires bool
	}{
		{"requires commands", true},
		{"continuous mode", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultGenericConfig()
			config.RequiresCommands = tt.requires
			p := NewGenericProtocolWithConfig(config)

			if p.RequiresCommand() != tt.requires {
				t.Errorf("expected %v, got %v", tt.requires, p.RequiresCommand())
			}
		})
	}
}

func TestGenericProtocol_ValidateResponse(t *testing.T) {
	p := NewGenericProtocol()

	tests := []struct {
		name     string
		response string
		want     bool
	}{
		{
			name:     "valid stable",
			response: "S 123.4 g",
			want:     true,
		},
		{
			name:     "valid unstable",
			response: "D 567.8 kg",
			want:     true,
		},
		{
			name:     "invalid format",
			response: "INVALID",
			want:     false,
		},
		{
			name:     "empty",
			response: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ValidateResponse([]byte(tt.response))
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestGenericProtocol_CaseInsensitive(t *testing.T) {
	config := DefaultGenericConfig()
	config.CaseSensitive = false
	config.ResponsePattern = `^([a-zA-Z])\s+([\d.]+)\s+(\w+)`
	config.StatusGroup = 1
	config.WeightGroup = 2
	config.UnitGroup = 3
	config.StableStatusValues = []string{"S", "ST"}
	config.UnstableStatusValues = []string{"D", "US"}

	p := NewGenericProtocolWithConfig(config)

	tests := []struct {
		name       string
		response   string
		wantStable bool
	}{
		{
			name:       "uppercase S",
			response:   "S 123.4 g",
			wantStable: true,
		},
		{
			name:       "lowercase s",
			response:   "s 123.4 g",
			wantStable: true,
		},
		{
			name:       "uppercase D",
			response:   "D 123.4 g",
			wantStable: false,
		},
		{
			name:       "lowercase d",
			response:   "d 123.4 g",
			wantStable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, _, err := p.ParseResponse([]byte(tt.response))

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reading.Stable != tt.wantStable {
				t.Errorf("stable: expected %v, got %v", tt.wantStable, reading.Stable)
			}
		})
	}
}

func TestGenericProtocol_CustomCommands(t *testing.T) {
	config := DefaultGenericConfig()
	config.WeightCommand = "READ\n"
	config.ZeroCommand = "ZERO\n"
	config.TareCommand = "TARE\n"

	p := NewGenericProtocolWithConfig(config)

	tests := []struct {
		name     string
		method   func() ([]byte, error)
		expected string
	}{
		{"weight", p.SendWeight, "READ\n"},
		{"zero", p.SendZero, "ZERO\n"},
		{"tare", p.SendTare, "TARE\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := tt.method()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(cmd) != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, string(cmd))
			}
		})
	}
}

func BenchmarkGenericProtocol_ParseResponse(b *testing.B) {
	p := NewGenericProtocol()
	response := []byte("S       123.4 g")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseResponse(response)
	}
}

func BenchmarkGenericProtocol_ParseResponse_SimpleFormat(b *testing.B) {
	config := GenericProtocolConfig{
		WeightCommand:    "R\n",
		ResponsePattern:  `([\d.]+)`,
		WeightGroup:      1,
		LineTerminator:   "\n",
		RequiresCommands: true,
		DefaultUnit:      UnitKilogram,
	}
	p := NewGenericProtocolWithConfig(config)
	response := []byte("123.45")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseResponse(response)
	}
}
