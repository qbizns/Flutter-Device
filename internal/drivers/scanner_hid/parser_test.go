package scanner_hid

import (
	"testing"
)

func TestParser_SimpleBarcode(t *testing.T) {
	parser := NewParser()

	// Simulate scanning "123" followed by Enter
	tests := []struct {
		name       string
		report     []byte
		wantCode   string
		wantSymbol string
		wantOK     bool
	}{
		{
			name:       "digit 1",
			report:     []byte{0x00, 0x00, 0x1E, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantCode:   "",
			wantSymbol: "",
			wantOK:     false,
		},
		{
			name:       "digit 2",
			report:     []byte{0x00, 0x00, 0x1F, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantCode:   "",
			wantSymbol: "",
			wantOK:     false,
		},
		{
			name:       "digit 3",
			report:     []byte{0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantCode:   "",
			wantSymbol: "",
			wantOK:     false,
		},
		{
			name:       "enter",
			report:     []byte{0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantCode:   "123",
			wantSymbol: "UNKNOWN",
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			barcode, symbology, ok := parser.Parse(tt.report)
			if ok != tt.wantOK {
				t.Errorf("Parse() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && barcode != tt.wantCode {
				t.Errorf("Parse() barcode = %v, want %v", barcode, tt.wantCode)
			}
			if ok && symbology != tt.wantSymbol {
				t.Errorf("Parse() symbology = %v, want %v", symbology, tt.wantSymbol)
			}
		})
	}
}

func TestParser_EAN13(t *testing.T) {
	parser := NewParser()

	input := "5901234123457"
	reports := makeHIDReports(input)

	var barcode, symbology string
	var ok bool

	for i, report := range reports {
		barcode, symbology, ok = parser.Parse(report)
		if i < len(reports)-1 && ok {
			t.Errorf("Parse completed early at position %d", i)
		}
	}

	if !ok {
		t.Fatal("Parse did not complete")
	}

	if barcode != input {
		t.Errorf("barcode = %q, want %q", barcode, input)
	}

	if symbology != "EAN13" {
		t.Errorf("symbology = %q, want EAN13", symbology)
	}
}

func TestParser_UPCA(t *testing.T) {
	parser := NewParser()

	input := "012345678905"
	reports := makeHIDReports(input)

	var barcode, symbology string
	var ok bool

	for _, report := range reports {
		barcode, symbology, ok = parser.Parse(report)
	}

	if !ok {
		t.Fatal("Parse did not complete")
	}

	if barcode != input {
		t.Errorf("barcode = %q, want %q", barcode, input)
	}

	if symbology != "UPCA" {
		t.Errorf("symbology = %q, want UPCA", symbology)
	}
}

func TestParser_Code39(t *testing.T) {
	parser := NewParser()

	input := "ABC-12345"
	reports := makeHIDReports(input)

	var barcode, symbology string
	var ok bool

	for _, report := range reports {
		barcode, symbology, ok = parser.Parse(report)
	}

	if !ok {
		t.Fatal("Parse did not complete")
	}

	if barcode != input {
		t.Errorf("barcode = %q, want %q", barcode, input)
	}

	if symbology != "CODE39" {
		t.Errorf("symbology = %q, want CODE39", symbology)
	}
}

func TestParser_Reset(t *testing.T) {
	parser := NewParser()

	// Start parsing but don't finish
	parser.Parse([]byte{0x00, 0x00, 0x1E, 0x00, 0x00, 0x00, 0x00, 0x00})
	parser.Parse([]byte{0x00, 0x00, 0x1F, 0x00, 0x00, 0x00, 0x00, 0x00})

	parser.Reset()

	// Parse new barcode
	parser.Parse([]byte{0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00})
	barcode, _, ok := parser.Parse([]byte{0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00})

	if !ok {
		t.Fatal("Parse did not complete")
	}

	if barcode != "3" {
		t.Errorf("barcode = %q, want '3' (reset failed)", barcode)
	}
}

func TestDetectSymbology(t *testing.T) {
	tests := []struct {
		barcode       string
		wantSymbology string
	}{
		{"5901234123457", "EAN13"},
		{"12345678", "EAN8"},
		{"012345678905", "UPCA"},
		{"123456", "UPCE"},
		{"12345678901234", "ITF14"},
		{"ABC-12345", "CODE39"},
		{"Test123", "CODE128"},
		{"XYZ", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.barcode, func(t *testing.T) {
			got := detectSymbology(tt.barcode)
			if got != tt.wantSymbology {
				t.Errorf("detectSymbology(%q) = %q, want %q", tt.barcode, got, tt.wantSymbology)
			}
		})
	}
}

func TestParseWithAIM(t *testing.T) {
	tests := []struct {
		input         string
		wantBarcode   string
		wantSymbology string
	}{
		{"]E05901234123457", "5901234123457", "EAN13"},
		{"]E412345678", "12345678", "EAN8"},
		{"]A0ABC123", "0ABC123", "CODE39"},
		{"]C0Test123", "Test123", "CODE128"},
		{"]Q3https://example.com", "https://example.com", "QRCODE"},
		{"123456789", "123456789", "CODE39"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			barcode, symbology := ParseWithAIM(tt.input)
			if barcode != tt.wantBarcode {
				t.Errorf("barcode = %q, want %q", barcode, tt.wantBarcode)
			}
			if symbology != tt.wantSymbology {
				t.Errorf("symbology = %q, want %q", symbology, tt.wantSymbology)
			}
		})
	}
}

func BenchmarkParser_Parse(b *testing.B) {
	parser := NewParser()
	report := []byte{0x00, 0x00, 0x1E, 0x00, 0x00, 0x00, 0x00, 0x00}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse(report)
	}
}

func BenchmarkParser_CompleteBarcode(b *testing.B) {
	input := "5901234123457"
	reports := makeHIDReports(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		for _, report := range reports {
			parser.Parse(report)
		}
	}
}

// Helper functions for tests

func makeHIDReports(input string) [][]byte {
	reports := make([][]byte, 0, len(input)+1)

	for _, char := range input {
		keyCode, modifier := charToKeyCode(char)
		report := []byte{modifier, 0x00, keyCode, 0x00, 0x00, 0x00, 0x00, 0x00}
		reports = append(reports, report)
	}

	// Add Enter key
	reports = append(reports, []byte{0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00})

	return reports
}

func charToKeyCode(char rune) (keyCode, modifier byte) {
	switch {
	case char >= '1' && char <= '9':
		return byte(0x1E + (char - '1')), 0x00
	case char == '0':
		return 0x27, 0x00
	case char >= 'a' && char <= 'z':
		return byte(0x04 + (char - 'a')), 0x00
	case char >= 'A' && char <= 'Z':
		return byte(0x04 + (char - 'A')), 0x02
	case char == '-':
		return 0x2D, 0x00
	case char == '.':
		return 0x37, 0x00
	case char == ' ':
		return 0x2C, 0x00
	default:
		return 0x00, 0x00
	}
}
