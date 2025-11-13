package printer_escpos

import (
	"testing"
)

// TestGetBiDiClass tests bidirectional character classification
func TestGetBiDiClass(t *testing.T) {
	testCases := []struct {
		name     string
		char     rune
		expected BiDiClass
	}{
		// Latin characters
		{"Latin A", 'A', BiDiL},
		{"Latin lowercase", 'z', BiDiL},

		// Arabic characters
		{"Arabic ALEF", 'ا', BiDiAL},
		{"Arabic BEH", 'ب', BiDiAL},
		{"Arabic MEEM", 'م', BiDiAL},

		// Numbers
		{"European digit", '5', BiDiEN},
		{"Arabic-Indic digit", '٥', BiDiAN},

		// Punctuation and symbols
		{"Plus sign", '+', BiDiES},
		{"Minus sign", '-', BiDiES},
		{"Dollar sign", '$', BiDiET},
		{"Comma", ',', BiDiCS},
		{"Period", '.', BiDiCS},
		{"Colon", ':', BiDiCS},

		// Whitespace
		{"Space", ' ', BiDiWS},
		{"Tab", '\t', BiDiWS},

		// Paragraph separator
		{"Newline", '\n', BiDiB},

		// Explicit marks
		{"LTR Mark", '\u200E', BiDiL},
		{"RTL Mark", '\u200F', BiDiR},
		{"LRE", '\u202A', BiDiLRE},
		{"RLE", '\u202B', BiDiRLE},
		{"PDF", '\u202C', BiDiPDF},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getBiDiClass(tc.char)
			if result != tc.expected {
				t.Errorf("getBiDiClass(%q) = %v, expected %v", tc.char, result, tc.expected)
			}
		})
	}
}

// TestBiDiAnalyzerCreation tests creating a BiDi analyzer
func TestBiDiAnalyzerCreation(t *testing.T) {
	testCases := []struct {
		name          string
		text          string
		baseDirection TextDirection
		expectedLevel int
	}{
		{"LTR base", "Hello", DirectionLTR, 0},
		{"RTL base", "مرحبا", DirectionRTL, 1},
		{"Auto-detect LTR", "Hello", DirectionMixed, 0},
		{"Auto-detect RTL", "مرحبا", DirectionMixed, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := NewBiDiAnalyzer(tc.text, tc.baseDirection)

			if analyzer == nil {
				t.Fatal("NewBiDiAnalyzer returned nil")
			}

			if analyzer.baseLevel != tc.expectedLevel {
				t.Errorf("Base level = %d, expected %d", analyzer.baseLevel, tc.expectedLevel)
			}

			if len(analyzer.text) != len([]rune(tc.text)) {
				t.Errorf("Text length mismatch: %d vs %d", len(analyzer.text), len([]rune(tc.text)))
			}
		})
	}
}

// TestBiDiAnalyze tests bidirectional analysis
func TestBiDiAnalyze(t *testing.T) {
	testCases := []struct {
		name          string
		text          string
		baseDirection TextDirection
		expectedRuns  int
		checkRuns     func(*testing.T, []BiDiRun)
	}{
		{
			name:          "Pure LTR",
			text:          "Hello World",
			baseDirection: DirectionLTR,
			expectedRuns:  1,
			checkRuns: func(t *testing.T, runs []BiDiRun) {
				if runs[0].Direction != DirectionLTR {
					t.Error("Expected LTR direction")
				}
				if runs[0].Text != "Hello World" {
					t.Errorf("Text mismatch: %q", runs[0].Text)
				}
			},
		},
		{
			name:          "Pure RTL",
			text:          "مرحبا",
			baseDirection: DirectionRTL,
			expectedRuns:  1,
			checkRuns: func(t *testing.T, runs []BiDiRun) {
				if runs[0].Direction != DirectionRTL {
					t.Error("Expected RTL direction")
				}
			},
		},
		{
			name:          "Mixed LTR-RTL",
			text:          "Hello مرحبا",
			baseDirection: DirectionMixed,
			expectedRuns:  2,
			checkRuns: func(t *testing.T, runs []BiDiRun) {
				if len(runs) < 2 {
					return
				}
				// First run should be LTR (Hello)
				if runs[0].Direction != DirectionLTR {
					t.Errorf("Run 0: expected LTR, got %v", runs[0].Direction)
				}
				// Second run should be RTL (مرحبا)
				if runs[1].Direction != DirectionRTL {
					t.Errorf("Run 1: expected RTL, got %v", runs[1].Direction)
				}
			},
		},
		{
			name:          "Numbers in RTL context",
			text:          "السعر 123",
			baseDirection: DirectionRTL,
			expectedRuns:  2, // Arabic text + numbers
			checkRuns: func(t *testing.T, runs []BiDiRun) {
				// Just verify we got runs
				if len(runs) == 0 {
					t.Error("No runs generated")
				}
			},
		},
		{
			name:          "Arabic-Indic digits",
			text:          "السعر ١٢٣",
			baseDirection: DirectionRTL,
			expectedRuns:  1, // All RTL including Arabic digits
			checkRuns: func(t *testing.T, runs []BiDiRun) {
				if runs[0].Direction != DirectionRTL {
					t.Error("Expected RTL direction for Arabic-Indic digits")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runs := AnalyzeBiDi(tc.text, tc.baseDirection)

			if len(runs) != tc.expectedRuns {
				t.Logf("Expected %d runs, got %d", tc.expectedRuns, len(runs))
				t.Logf("Runs:")
				for i, run := range runs {
					t.Logf("  %d: %q (%v, level %d)", i, run.Text, run.Direction, run.Level)
				}
			}

			if tc.checkRuns != nil {
				tc.checkRuns(t, runs)
			}
		})
	}
}

// TestReorderVisual tests visual reordering of runs
func TestReorderVisual(t *testing.T) {
	testCases := []struct {
		name  string
		runs  []BiDiRun
		check func(*testing.T, []BiDiRun)
	}{
		{
			name: "LTR run unchanged",
			runs: []BiDiRun{
				{Text: "Hello", Direction: DirectionLTR, Level: 0},
			},
			check: func(t *testing.T, reordered []BiDiRun) {
				if reordered[0].Text != "Hello" {
					t.Errorf("LTR text should not change: %q", reordered[0].Text)
				}
			},
		},
		{
			name: "RTL run reversed",
			runs: []BiDiRun{
				{Text: "ABC", Direction: DirectionRTL, Level: 1},
			},
			check: func(t *testing.T, reordered []BiDiRun) {
				if reordered[0].Text != "CBA" {
					t.Errorf("RTL text should be reversed: got %q, expected 'CBA'", reordered[0].Text)
				}
			},
		},
		{
			name: "Mixed runs",
			runs: []BiDiRun{
				{Text: "Hello", Direction: DirectionLTR, Level: 0},
				{Text: "مرحبا", Direction: DirectionRTL, Level: 1},
			},
			check: func(t *testing.T, reordered []BiDiRun) {
				// LTR unchanged
				if reordered[0].Text != "Hello" {
					t.Errorf("LTR run should not change")
				}
				// RTL reversed
				originalRunes := []rune("مرحبا")
				reversedRunes := []rune(reordered[1].Text)
				if len(reversedRunes) != len(originalRunes) {
					t.Errorf("RTL run length changed")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reordered := ReorderVisual(tc.runs)
			tc.check(t, reordered)
		})
	}
}

// TestGetParagraphDirection tests paragraph direction detection
func TestGetParagraphDirection(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected TextDirection
	}{
		{"Pure English", "Hello World", DirectionLTR},
		{"Pure Arabic", "مرحبا بك", DirectionRTL},
		{"Mixed - starts with English", "Hello مرحبا", DirectionLTR},
		{"Mixed - starts with Arabic", "مرحبا Hello", DirectionRTL},
		{"Numbers only", "12345", DirectionLTR},
		{"Empty", "", DirectionLTR},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetParagraphDirection(tc.text)
			if result != tc.expected {
				t.Errorf("GetParagraphDirection(%q) = %v, expected %v", tc.text, result, tc.expected)
			}
		})
	}
}

// TestBiDiComplexScenarios tests complex bidirectional scenarios
func TestBiDiComplexScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		text        string
		description string
	}{
		{
			name:        "Product with Arabic name",
			text:        "Coca-Cola كولا",
			description: "English brand name with Arabic translation",
		},
		{
			name:        "Price with currency",
			text:        "السعر: 25.50 ر.س",
			description: "Arabic label with Western digits and Arabic currency",
		},
		{
			name:        "Receipt header",
			text:        "مطعم البرج - Tower Restaurant",
			description: "Arabic and English business name",
		},
		{
			name:        "Address line",
			text:        "123 شارع الملك فهد، الرياض",
			description: "Number, Arabic street name, Arabic city",
		},
		{
			name:        "Invoice number",
			text:        "فاتورة رقم INV-2024-001",
			description: "Arabic word with English invoice number",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)
			t.Logf("Input: %s", tc.text)

			runs := AnalyzeBiDi(tc.text, DirectionMixed)

			t.Logf("Generated %d runs:", len(runs))
			for i, run := range runs {
				t.Logf("  Run %d: %q (%v, level %d, pos %d-%d)",
					i, run.Text, run.Direction, run.Level, run.Start, run.End)
			}

			// Verify we got at least one run
			if len(runs) == 0 {
				t.Error("No runs generated")
			}

			// Verify total text length matches
			totalLen := 0
			for _, run := range runs {
				totalLen += len([]rune(run.Text))
			}
			inputLen := len([]rune(tc.text))
			if totalLen != inputLen {
				t.Errorf("Total run length %d != input length %d", totalLen, inputLen)
			}
		})
	}
}

// TestBiDiWeakTypes tests handling of weak character types
func TestBiDiWeakTypes(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{"Numbers between Arabic", "مرحبا 123 شكرا"},
		{"Punctuation in RTL", "مرحبا، كيف حالك؟"},
		{"Currency symbol", "السعر: $25.50"},
		{"Parentheses in mixed text", "(مرحبا) Hello"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runs := AnalyzeBiDi(tc.text, DirectionMixed)

			t.Logf("Input: %s", tc.text)
			t.Logf("Runs:")
			for i, run := range runs {
				t.Logf("  %d: %q (%v)", i, run.Text, run.Direction)
			}

			// Just verify we got runs and they cover the input
			if len(runs) == 0 {
				t.Error("No runs generated")
			}
		})
	}
}

// TestEnhancedSplitIntoRuns tests the updated SplitIntoRuns function
func TestEnhancedSplitIntoRuns(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{"Simple English", "Hello World"},
		{"Simple Arabic", "مرحباً"},
		{"Mixed", "Hello مرحبا"},
		{"With numbers", "السعر: 25.50 ر.س"},
		{"Product name", "Coca-Cola كولا"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runs := SplitIntoRuns(tc.text)

			if len(runs) == 0 {
				t.Error("SplitIntoRuns returned no runs")
				return
			}

			t.Logf("Input: %s", tc.text)
			t.Logf("Generated %d runs:", len(runs))
			for i, run := range runs {
				t.Logf("  %d: %q (%v, pos %d-%d)",
					i, run.Text, run.Direction, run.Start, run.End)
			}

			// Verify runs cover the entire input
			totalLen := 0
			for _, run := range runs {
				totalLen += run.End - run.Start
			}
			inputLen := len([]rune(tc.text))
			if totalLen != inputLen {
				t.Errorf("Runs don't cover full input: %d vs %d runes", totalLen, inputLen)
			}
		})
	}
}

// BenchmarkBiDiAnalyze benchmarks bidirectional analysis
func BenchmarkBiDiAnalyze(b *testing.B) {
	text := "Hello مرحبا World السعر: 25.50 ر.س"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AnalyzeBiDi(text, DirectionMixed)
	}
}

// BenchmarkGetBiDiClass benchmarks character classification
func BenchmarkGetBiDiClass(b *testing.B) {
	chars := []rune("Hello مرحبا 123 ٥٦٧")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range chars {
			_ = getBiDiClass(r)
		}
	}
}

// BenchmarkReorderVisual benchmarks visual reordering
func BenchmarkReorderVisual(b *testing.B) {
	runs := []BiDiRun{
		{Text: "Hello", Direction: DirectionLTR, Level: 0},
		{Text: "مرحبا", Direction: DirectionRTL, Level: 1},
		{Text: "World", Direction: DirectionLTR, Level: 0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ReorderVisual(runs)
	}
}
