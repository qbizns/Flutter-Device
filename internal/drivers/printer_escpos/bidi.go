package printer_escpos

import (
	"unicode"
)

// BiDi implements a simplified version of the Unicode Bidirectional Algorithm
// Reference: Unicode Standard Annex #9 (UAX#9)
// https://www.unicode.org/reports/tr9/

// BiDiClass represents the bidirectional character class
type BiDiClass int

const (
	// Strong types
	BiDiL   BiDiClass = iota // Left-to-Right (Latin, Cyrillic, etc.)
	BiDiR                    // Right-to-Left (Hebrew)
	BiDiAL                   // Right-to-Left Arabic

	// Weak types
	BiDiEN                   // European Number
	BiDiES                   // European Separator (+ -)
	BiDiET                   // European Terminator (currency)
	BiDiAN                   // Arabic Number
	BiDiCS                   // Common Separator (, .)
	BiDiNSM                  // Non-Spacing Mark (diacritics)
	BiDiBN                   // Boundary Neutral (format controls)

	// Neutral types
	BiDiB                    // Paragraph Separator
	BiDiS                    // Segment Separator
	BiDiWS                   // Whitespace
	BiDiON                   // Other Neutrals (punctuation, symbols)

	// Explicit formatting
	BiDiLRE                  // Left-to-Right Embedding
	BiDiLRO                  // Left-to-Right Override
	BiDiRLE                  // Right-to-Left Embedding
	BiDiRLO                  // Right-to-Left Override
	BiDiPDF                  // Pop Directional Format
	BiDiLRI                  // Left-to-Right Isolate
	BiDiRLI                  // Right-to-Left Isolate
	BiDiFSI                  // First Strong Isolate
	BiDiPDI                  // Pop Directional Isolate
)

// String returns the name of the BiDi class
func (c BiDiClass) String() string {
	names := map[BiDiClass]string{
		BiDiL: "L", BiDiR: "R", BiDiAL: "AL",
		BiDiEN: "EN", BiDiES: "ES", BiDiET: "ET", BiDiAN: "AN",
		BiDiCS: "CS", BiDiNSM: "NSM", BiDiBN: "BN",
		BiDiB: "B", BiDiS: "S", BiDiWS: "WS", BiDiON: "ON",
		BiDiLRE: "LRE", BiDiLRO: "LRO", BiDiRLE: "RLE", BiDiRLO: "RLO",
		BiDiPDF: "PDF", BiDiLRI: "LRI", BiDiRLI: "RLI", BiDiFSI: "FSI", BiDiPDI: "PDI",
	}
	if name, ok := names[c]; ok {
		return name
	}
	return "Unknown"
}

// getBiDiClass returns the bidirectional class for a rune
func getBiDiClass(r rune) BiDiClass {
	// Check specific character ranges first before general Arabic range

	// Paragraph separator (check before whitespace)
	if r == '\n' || r == '\r' || r == 0x2029 {
		return BiDiB
	}

	// Arabic-Indic digits (٠-٩) - check before general Arabic range
	if r >= 0x0660 && r <= 0x0669 {
		return BiDiAN
	}

	// Extended Arabic-Indic digits
	if r >= 0x06F0 && r <= 0x06F9 {
		return BiDiAN
	}

	// European digits (0-9)
	if r >= '0' && r <= '9' {
		return BiDiEN
	}

	// Arabic characters (U+0600-U+06FF, U+0750-U+077F, U+08A0-U+08FF)
	// NOTE: This must come AFTER Arabic-Indic digit checks
	if (r >= 0x0600 && r <= 0x06FF) ||
		(r >= 0x0750 && r <= 0x077F) ||
		(r >= 0x08A0 && r <= 0x08FF) ||
		(r >= 0xFB50 && r <= 0xFDFF) ||
		(r >= 0xFE70 && r <= 0xFEFF) {
		return BiDiAL
	}

	// Hebrew characters (U+0590-U+05FF)
	if r >= 0x0590 && r <= 0x05FF {
		return BiDiR
	}

	// Plus/minus signs
	if r == '+' || r == '-' {
		return BiDiES
	}

	// Currency symbols
	if r == '$' || r == '€' || r == '£' || r == '¥' {
		return BiDiET
	}

	// Common separators
	if r == ',' || r == '.' || r == ':' || r == ';' {
		return BiDiCS
	}

	// Whitespace (but not newlines, which are paragraph separators)
	if unicode.IsSpace(r) {
		return BiDiWS
	}

	// Non-spacing marks (diacritics)
	if unicode.Is(unicode.Mn, r) {
		return BiDiNSM
	}

	// Explicit directional marks
	switch r {
	case 0x202A: // LEFT-TO-RIGHT EMBEDDING
		return BiDiLRE
	case 0x202B: // RIGHT-TO-LEFT EMBEDDING
		return BiDiRLE
	case 0x202C: // POP DIRECTIONAL FORMATTING
		return BiDiPDF
	case 0x202D: // LEFT-TO-RIGHT OVERRIDE
		return BiDiLRO
	case 0x202E: // RIGHT-TO-LEFT OVERRIDE
		return BiDiRLO
	case 0x2066: // LEFT-TO-RIGHT ISOLATE
		return BiDiLRI
	case 0x2067: // RIGHT-TO-LEFT ISOLATE
		return BiDiRLI
	case 0x2068: // FIRST STRONG ISOLATE
		return BiDiFSI
	case 0x2069: // POP DIRECTIONAL ISOLATE
		return BiDiPDI
	case 0x200E: // LEFT-TO-RIGHT MARK
		return BiDiL
	case 0x200F: // RIGHT-TO-LEFT MARK
		return BiDiR
	}

	// Latin and other LTR scripts
	if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
		return BiDiL
	}

	// Cyrillic
	if r >= 0x0400 && r <= 0x04FF {
		return BiDiL
	}

	// Greek
	if r >= 0x0370 && r <= 0x03FF {
		return BiDiL
	}

	// Default to Other Neutral for punctuation and symbols
	return BiDiON
}

// BiDiRun represents a run of characters with the same direction
type BiDiRun struct {
	Text      string        // The text content
	Level     int           // Embedding level (0=LTR, 1=RTL, 2+=nested)
	Direction TextDirection // Resolved direction
	Start     int           // Start position in original text
	End       int           // End position in original text
}

// BiDiAnalyzer analyzes text and splits it into bidirectional runs
type BiDiAnalyzer struct {
	baseLevel     int          // Base paragraph level (0=LTR, 1=RTL)
	text          []rune       // Input text as runes
	classes       []BiDiClass  // BiDi class for each character
	levels        []int        // Embedding level for each character
	isolateStatus []bool       // Whether char is in isolate
}

// NewBiDiAnalyzer creates a new BiDi analyzer
func NewBiDiAnalyzer(text string, baseDirection TextDirection) *BiDiAnalyzer {
	runes := []rune(text)
	analyzer := &BiDiAnalyzer{
		text:          runes,
		classes:       make([]BiDiClass, len(runes)),
		levels:        make([]int, len(runes)),
		isolateStatus: make([]bool, len(runes)),
	}

	// Determine base level
	if baseDirection == DirectionRTL {
		analyzer.baseLevel = 1
	} else if baseDirection == DirectionLTR {
		analyzer.baseLevel = 0
	} else {
		// Auto-detect: use first strong character
		analyzer.baseLevel = analyzer.detectBaseLevel()
	}

	// Classify each character
	for i, r := range runes {
		analyzer.classes[i] = getBiDiClass(r)
		analyzer.levels[i] = analyzer.baseLevel
	}

	return analyzer
}

// detectBaseLevel detects the base paragraph level from first strong character
func (a *BiDiAnalyzer) detectBaseLevel() int {
	for _, r := range a.text {
		class := getBiDiClass(r)
		if class == BiDiL {
			return 0 // LTR
		}
		if class == BiDiR || class == BiDiAL {
			return 1 // RTL
		}
	}
	return 0 // Default to LTR
}

// Analyze performs bidirectional analysis and returns runs
func (a *BiDiAnalyzer) Analyze() []BiDiRun {
	// Simplified BiDi algorithm implementation
	// Full UAX#9 is complex; this handles common cases

	// Step 1: Resolve explicit embeddings and overrides
	a.resolveExplicitLevels()

	// Step 2: Resolve weak types
	a.resolveWeakTypes()

	// Step 3: Resolve neutral types
	a.resolveNeutralTypes()

	// Step 4: Resolve implicit levels
	a.resolveImplicitLevels()

	// Step 5: Split into runs based on levels
	return a.splitIntoRuns()
}

// resolveExplicitLevels handles explicit directional formatting characters
func (a *BiDiAnalyzer) resolveExplicitLevels() {
	// Simplified: just track explicit marks
	// Full implementation would handle embedding stack
	for i, class := range a.classes {
		switch class {
		case BiDiLRE, BiDiLRO, BiDiLRI:
			// Start LTR section
			a.levels[i] = 0
		case BiDiRLE, BiDiRLO, BiDiRLI:
			// Start RTL section
			a.levels[i] = 1
		case BiDiPDF, BiDiPDI:
			// Pop level - reset to base
			a.levels[i] = a.baseLevel
		}
	}
}

// resolveWeakTypes resolves weak character types
func (a *BiDiAnalyzer) resolveWeakTypes() {
	// Rule W1: NSM (non-spacing marks) take the type of the preceding character
	for i := 1; i < len(a.classes); i++ {
		if a.classes[i] == BiDiNSM {
			a.classes[i] = a.classes[i-1]
		}
	}

	// Rule W2: EN (European numbers) after AL (Arabic) become AN (Arabic numbers)
	prevAL := false
	for i := 0; i < len(a.classes); i++ {
		if a.classes[i] == BiDiAL {
			prevAL = true
		} else if a.classes[i] == BiDiEN && prevAL {
			a.classes[i] = BiDiAN
		} else if a.classes[i] != BiDiNSM {
			prevAL = false
		}
	}

	// Rule W4: Single separators between numbers take the type of the numbers
	for i := 1; i < len(a.classes)-1; i++ {
		if (a.classes[i] == BiDiES || a.classes[i] == BiDiCS) &&
			a.classes[i-1] == BiDiEN && a.classes[i+1] == BiDiEN {
			a.classes[i] = BiDiEN
		}
	}
}

// resolveNeutralTypes resolves neutral character types
func (a *BiDiAnalyzer) resolveNeutralTypes() {
	// Simplified: neutrals take direction from surrounding strong types
	for i := 0; i < len(a.classes); i++ {
		if a.classes[i] == BiDiWS || a.classes[i] == BiDiON || a.classes[i] == BiDiCS {
			// Look at surrounding context
			prevStrong := a.findPrevStrongType(i)
			nextStrong := a.findNextStrongType(i)

			if prevStrong == nextStrong {
				// Same direction on both sides
				if prevStrong == BiDiL {
					a.levels[i] = 0
				} else if prevStrong == BiDiR || prevStrong == BiDiAL {
					a.levels[i] = 1
				}
			} else {
				// Different directions - use base level
				a.levels[i] = a.baseLevel
			}
		}
	}
}

// findPrevStrongType finds the previous strong directional type
func (a *BiDiAnalyzer) findPrevStrongType(pos int) BiDiClass {
	for i := pos - 1; i >= 0; i-- {
		if a.classes[i] == BiDiL || a.classes[i] == BiDiR || a.classes[i] == BiDiAL {
			return a.classes[i]
		}
	}
	if a.baseLevel == 0 {
		return BiDiL
	}
	return BiDiR
}

// findNextStrongType finds the next strong directional type
func (a *BiDiAnalyzer) findNextStrongType(pos int) BiDiClass {
	for i := pos + 1; i < len(a.classes); i++ {
		if a.classes[i] == BiDiL || a.classes[i] == BiDiR || a.classes[i] == BiDiAL {
			return a.classes[i]
		}
	}
	if a.baseLevel == 0 {
		return BiDiL
	}
	return BiDiR
}

// resolveImplicitLevels assigns final embedding levels
func (a *BiDiAnalyzer) resolveImplicitLevels() {
	for i, class := range a.classes {
		switch class {
		case BiDiL:
			// LTR characters get even level
			if a.levels[i]%2 == 1 {
				a.levels[i]++
			}
		case BiDiR, BiDiAL:
			// RTL characters get odd level
			if a.levels[i]%2 == 0 {
				a.levels[i]++
			}
		case BiDiEN, BiDiAN:
			// Numbers in RTL context get RTL level
			if a.levels[i]%2 == 1 {
				a.levels[i]++
			}
		}
	}
}

// splitIntoRuns splits text into runs based on embedding levels
func (a *BiDiAnalyzer) splitIntoRuns() []BiDiRun {
	if len(a.text) == 0 {
		return nil
	}

	runs := []BiDiRun{}
	runStart := 0
	currentLevel := a.levels[0]

	for i := 1; i < len(a.text); i++ {
		if a.levels[i] != currentLevel {
			// Level changed - create run
			run := a.createRun(runStart, i, currentLevel)
			runs = append(runs, run)
			runStart = i
			currentLevel = a.levels[i]
		}
	}

	// Add final run
	run := a.createRun(runStart, len(a.text), currentLevel)
	runs = append(runs, run)

	return runs
}

// createRun creates a BiDiRun from a range of characters
func (a *BiDiAnalyzer) createRun(start, end, level int) BiDiRun {
	text := string(a.text[start:end])
	direction := DirectionLTR
	if level%2 == 1 {
		direction = DirectionRTL
	}

	return BiDiRun{
		Text:      text,
		Level:     level,
		Direction: direction,
		Start:     start,
		End:       end,
	}
}

// ReorderVisual reorders runs for visual display (RTL runs are reversed)
func ReorderVisual(runs []BiDiRun) []BiDiRun {
	// For each RTL run, reverse the character order
	reordered := make([]BiDiRun, len(runs))
	copy(reordered, runs)

	for i := range reordered {
		if reordered[i].Direction == DirectionRTL {
			// Reverse the text in RTL runs
			runes := []rune(reordered[i].Text)
			for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
				runes[left], runes[right] = runes[right], runes[left]
			}
			reordered[i].Text = string(runes)
		}
	}

	return reordered
}

// AnalyzeBiDi is a convenience function that analyzes text and returns runs
func AnalyzeBiDi(text string, baseDirection TextDirection) []BiDiRun {
	analyzer := NewBiDiAnalyzer(text, baseDirection)
	return analyzer.Analyze()
}

// GetParagraphDirection determines the overall direction of a paragraph
func GetParagraphDirection(text string) TextDirection {
	analyzer := NewBiDiAnalyzer(text, DirectionMixed)
	if analyzer.baseLevel == 0 {
		return DirectionLTR
	}
	return DirectionRTL
}
