package printer_escpos

import (
	"bytes"
	"testing"
)

// TestArabicShaperCreation tests creating an Arabic shaper
func TestArabicShaperCreation(t *testing.T) {
	shaper := NewArabicShaper()

	if shaper == nil {
		t.Fatal("NewArabicShaper() returned nil")
	}

	if shaper.codePage != 864 {
		t.Errorf("Expected code page 864, got %d", shaper.codePage)
	}

	if shaper.font == nil {
		t.Fatal("ArabicShaper.font is nil")
	}
}

// TestESCPOSFontCreation tests creating an ESC/POS Arabic font
func TestESCPOSFontCreation(t *testing.T) {
	font := NewESCPOSArabicFont()

	if font == nil {
		t.Fatal("NewESCPOSArabicFont() returned nil")
	}

	if font.codePage != 864 {
		t.Errorf("Expected code page 864, got %d", font.codePage)
	}

	if len(font.charMap) == 0 {
		t.Error("charMap is empty")
	}

	if len(font.reverseMap) == 0 {
		t.Error("reverseMap is empty")
	}

	if len(font.glyphMap) == 0 {
		t.Error("glyphMap is empty")
	}
}

// TestArabicCharacterMapping tests Unicode to CP864 character mapping
func TestArabicCharacterMapping(t *testing.T) {
	font := NewESCPOSArabicFont()

	testCases := []struct {
		name     string
		char     rune
		expected byte
	}{
		{"ALEF", 'ا', 0xC7},
		{"BEH", 'ب', 0xCB},
		{"JEEM", 'ج', 0xCE},
		{"DAL", 'د', 0xD1},
		{"REH", 'ر', 0xD3},
		{"SEEN", 'س', 0xD5},
		{"AIN", 'ع', 0xDB},
		{"LAM", 'ل', 0xE0},
		{"MEEM", 'م', 0xE1},
		{"YEH", 'ي', 0xE6},
		{"Arabic-Indic Zero", '٠', 0x80},
		{"Arabic-Indic Five", '٥', 0x85},
		{"Arabic Question Mark", '؟', 0xBF},
		{"Arabic Comma", '،', 0xAC},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if mapped, ok := font.charMap[tc.char]; !ok {
				t.Errorf("Character %c (U+%04X) not found in charMap", tc.char, tc.char)
			} else if mapped != tc.expected {
				t.Errorf("Character %c mapped to 0x%02X, expected 0x%02X", tc.char, mapped, tc.expected)
			}
		})
	}
}

// TestUnicodeToCP864Conversion tests converting Unicode strings to CP864
func TestUnicodeToCP864Conversion(t *testing.T) {
	font := NewESCPOSArabicFont()

	testCases := []struct {
		name     string
		input    string
		contains []byte // Bytes that should be in output
	}{
		{
			name:     "Simple Arabic Word",
			input:    "مرحبا",
			contains: []byte{0xE1, 0xD3, 0xCF, 0xCB, 0xC7}, // م ر ح ب ا
		},
		{
			name:     "Arabic Numbers",
			input:    "١٢٣",
			contains: []byte{0x81, 0x82, 0x83}, // ١ ٢ ٣
		},
		{
			name:     "English ASCII",
			input:    "ABC",
			contains: []byte{'A', 'B', 'C'},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := font.MapUnicodeToCP864(tc.input)
			if err != nil {
				t.Errorf("MapUnicodeToCP864() error: %v", err)
				return
			}

			// Check that all expected bytes are present
			for _, expected := range tc.contains {
				found := false
				for _, b := range result {
					if b == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected byte 0x%02X not found in result %v", expected, result)
				}
			}
		})
	}
}

// TestCharacterSupport tests validating character support
func TestCharacterSupport(t *testing.T) {
	font := NewESCPOSArabicFont()

	testCases := []struct {
		name            string
		input           string
		expectSupported bool
	}{
		{"Simple Arabic", "مرحبا", true},
		{"English ASCII", "Hello", true},
		{"Mixed Arabic-English", "Hello مرحبا", true},
		{"Arabic with numbers", "السعر: 123", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			supported, unsupported := font.ValidateCharacterSupport(tc.input)

			if supported != tc.expectSupported {
				t.Errorf("Expected supported=%v, got %v (unsupported chars: %v)",
					tc.expectSupported, supported, unsupported)
			}
		})
	}
}

// TestTextShaping tests the ShapeText function
func TestTextShaping(t *testing.T) {
	font := NewESCPOSArabicFont()

	testCases := []struct {
		name      string
		input     string
		direction TextDirection
	}{
		{"Simple Arabic RTL", "مرحبا", DirectionRTL},
		{"English LTR", "Hello", DirectionLTR},
		{"Mixed", "Hello مرحبا", DirectionMixed},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := font.ShapeText(tc.input, tc.direction)
			if err != nil {
				t.Errorf("ShapeText() error: %v", err)
				return
			}

			if len(result) == 0 {
				t.Error("ShapeText() returned empty result")
			}

			t.Logf("Input: %s (%d runes)", tc.input, len([]rune(tc.input)))
			t.Logf("Output: %d bytes", len(result))
		})
	}
}

// TestArabicRenderingCommands tests Arabic text rendering with ESC/POS commands
func TestArabicRenderingCommands(t *testing.T) {
	shaper := NewArabicShaper()

	testCases := []struct {
		name  string
		input string
		style TextStyle
	}{
		{
			name:  "Simple Arabic",
			input: "مرحبا",
			style: TextStyle{Bold: false, Underline: false},
		},
		{
			name:  "Bold Arabic",
			input: "مرحبا",
			style: TextStyle{Bold: true, Underline: false},
		},
		{
			name:  "Underlined Arabic",
			input: "السعر",
			style: TextStyle{Bold: false, Underline: true},
		},
		{
			name:  "Double Width Arabic",
			input: "إجمالي",
			style: TextStyle{DoubleWidth: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := shaper.RenderArabicText(tc.input, tc.style)

			if len(result) == 0 {
				t.Error("RenderArabicText() returned empty result")
				return
			}

			// Check for Code Page 864 command (ESC t 0x1E)
			if !bytes.Contains(result, []byte{0x1B, 't', 0x1E}) {
				t.Error("Result doesn't contain Code Page 864 command")
			}

			// Check for Code Page reset (ESC t 0x00)
			if !bytes.Contains(result, []byte{0x1B, 't', 0x00}) {
				t.Error("Result doesn't contain Code Page reset command")
			}

			t.Logf("Input: %s", tc.input)
			t.Logf("Output length: %d bytes", len(result))
			t.Logf("Contains CP864 set: %v", bytes.Contains(result, []byte{0x1B, 't', 0x1E}))
			t.Logf("Contains RTL set: %v", bytes.Contains(result, []byte{0x1B, 0x7B, 0x01}))
		})
	}
}

// TestMixedTextRendering tests rendering mixed LTR/RTL text
func TestMixedTextRendering(t *testing.T) {
	shaper := NewArabicShaper()

	testCases := []struct {
		name  string
		input string
	}{
		{"Product Name", "Coca-Cola كولا"},
		{"Price with currency", "السعر: 25.50 ر.س"},
		{"Receipt header", "مطعم البرج - Tower Restaurant"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			style := TextStyle{}
			result := shaper.RenderArabicText(tc.input, style)

			if len(result) == 0 {
				t.Error("RenderArabicText() returned empty result")
				return
			}

			t.Logf("Input: %s", tc.input)
			t.Logf("Direction: %v", DetectTextDirection(tc.input))
			t.Logf("Output: %d bytes", len(result))
		})
	}
}

// TestTextRunSplitting tests splitting text into LTR and RTL runs
func TestTextRunSplitting(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedRuns  int
		checkFunction func([]TextRun) error
	}{
		{
			name:         "Pure English",
			input:        "Hello World",
			expectedRuns: 1,
			checkFunction: func(runs []TextRun) error {
				if runs[0].Direction != DirectionLTR {
					return testError("expected LTR direction")
				}
				return nil
			},
		},
		{
			name:         "Pure Arabic",
			input:        "مرحباً",
			expectedRuns: 1,
			checkFunction: func(runs []TextRun) error {
				if runs[0].Direction != DirectionRTL {
					return testError("expected RTL direction")
				}
				return nil
			},
		},
		{
			name:         "Mixed English-Arabic",
			input:        "Hello مرحبا",
			expectedRuns: 2,
			checkFunction: func(runs []TextRun) error {
				if runs[0].Direction == runs[1].Direction {
					return testError("expected different directions for runs")
				}
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runs := SplitIntoRuns(tc.input)

			if len(runs) != tc.expectedRuns {
				t.Errorf("Expected %d runs, got %d", tc.expectedRuns, len(runs))
			}

			if tc.checkFunction != nil {
				if err := tc.checkFunction(runs); err != nil {
					t.Error(err)
				}
			}

			// Log run details
			for i, run := range runs {
				t.Logf("Run %d: '%s' (%v, pos %d-%d)",
					i, run.Text, run.Direction, run.Start, run.End)
			}
		})
	}
}

// testError is a helper to create test errors
func testError(msg string) error {
	return &testErr{msg: msg}
}

type testErr struct {
	msg string
}

func (e *testErr) Error() string {
	return e.msg
}

// TestArabicTextWrapping tests wrapping Arabic text at word boundaries
func TestArabicTextWrapping(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		maxWidth   int
		expectLines int
	}{
		{
			name:       "Short text",
			input:      "مرحبا",
			maxWidth:   40,
			expectLines: 1,
		},
		{
			name:       "Long text needs wrapping",
			input:      "محل البقالة الكبير للمواد الغذائية والمشروبات",
			maxWidth:   20,
			expectLines: 3, // Approximate
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lines := WrapArabicText(tc.input, tc.maxWidth)

			if len(lines) < 1 {
				t.Error("WrapArabicText returned no lines")
				return
			}

			// Check that each line is within max width
			for i, line := range lines {
				width := len([]rune(line))
				if width > tc.maxWidth {
					t.Errorf("Line %d exceeds max width: %d > %d", i, width, tc.maxWidth)
				}
				t.Logf("Line %d: %s (%d chars)", i, line, width)
			}
		})
	}
}

// TestArabicExamples tests the GetArabicExamples function
func TestArabicExamples(t *testing.T) {
	examples := GetArabicExamples()

	if len(examples) == 0 {
		t.Error("GetArabicExamples() returned empty map")
		return
	}

	expectedKeys := []string{"hello", "thank_you", "total", "invoice", "price", "sar"}

	for _, key := range expectedKeys {
		if _, ok := examples[key]; !ok {
			t.Errorf("Expected key %q not found in examples", key)
		}
	}

	t.Logf("Found %d Arabic examples", len(examples))
	for key, value := range examples {
		t.Logf("  %s: %s", key, value)
	}
}

// TestFontMetrics tests getting font metrics
func TestFontMetrics(t *testing.T) {
	font := NewESCPOSArabicFont()
	metrics := font.GetFontMetrics()

	if metrics.CharWidth == 0 {
		t.Error("CharWidth is 0")
	}

	if metrics.CharHeight == 0 {
		t.Error("CharHeight is 0")
	}

	if !metrics.Monospaced {
		t.Error("Expected monospaced font for ESC/POS")
	}

	t.Logf("Font metrics: %dx%d, spacing: %d", metrics.CharWidth, metrics.CharHeight, metrics.LineSpacing)
}

// BenchmarkArabicShaping benchmarks Arabic text shaping
func BenchmarkArabicShaping(b *testing.B) {
	shaper := NewArabicShaper()
	text := "مرحباً بكم في مطعمنا"
	style := TextStyle{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = shaper.RenderArabicText(text, style)
	}
}

// BenchmarkUnicodeToCP864 benchmarks Unicode to CP864 mapping
func BenchmarkUnicodeToCP864(b *testing.B) {
	font := NewESCPOSArabicFont()
	text := "السعر: 25.50 ر.س"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = font.MapUnicodeToCP864(text)
	}
}

// BenchmarkRunSplitting benchmarks text run splitting
func BenchmarkRunSplitting(b *testing.B) {
	text := "Coca-Cola كولا Price: 25.50 SAR"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SplitIntoRuns(text)
	}
}
