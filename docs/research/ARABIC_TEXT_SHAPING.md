# Arabic Text Shaping Library Evaluation

**Date**: 2025-11-13
**Task**: Phase 1.1.1 - Evaluate Arabic text shaping libraries for ESC/POS printer
**Status**: Research Complete - Recommendation Made

---

## Executive Summary

**Recommendation**: Use **`github.com/go-text/typesetting/harfbuzz`** for Arabic text shaping in ESC/POS printer driver.

**Rationale**:
- Pure Go implementation (no C dependencies)
- Active development (last updated Feb 2025)
- Used by multiple production GUI toolkits (Fyne, Gio, Ebitengine)
- Dual-licensed (BSD-3-Clause / Unlicense)
- Comprehensive text layout with HarfBuzz and BiDi support
- Successor to `benoitkugler/textlayout` (now merged)

**Alternative**: Supplement with **`github.com/01walid/goarabic`** (MIT) for Arabic-specific utilities if needed.

---

## Problem Statement

Current ESC/POS printer driver cannot render Arabic text correctly because:

1. **No Text Shaping**: Arabic characters must change shape based on position (initial, medial, final, isolated)
2. **No RTL Support**: Arabic is written right-to-left, not left-to-right
3. **No BiDi Algorithm**: Mixed Arabic/English text needs bidirectional text handling
4. **No Glyph Joining**: Arabic letters connect to each other with specific joining rules

**Example Issue**:
- Input: `"مرحباً"` (Hello in Arabic)
- Current output: م ر ح ب ا ً (disconnected, left-to-right - **wrong**)
- Correct output: مرحباً (connected, right-to-left shapes)

---

## Library Options Evaluated

### Option 1: `github.com/go-text/typesetting` ⭐ **RECOMMENDED**

**Source**: https://pkg.go.dev/github.com/go-text/typesetting
**License**: Dual (BSD-3-Clause / Unlicense) ✅
**Last Updated**: February 21, 2025 (active)
**Version**: v0.2.1 (unstable, but production-ready per authors)

**What it provides**:
- **HarfBuzz**: Advanced text layout engine (text shaping for all scripts including Arabic)
- **FriBidi**: Unicode Bidirectional Algorithm (for RTL and mixed LTR/RTL)
- **Segmenter**: Unicode text segmentation
- **Font**: OpenType font property access
- Pure Go implementation (no C dependencies or CGO required)

**Pros**:
- ✅ Actively maintained (updated Feb 2025)
- ✅ Used in production by major Go GUI toolkits
- ✅ Pure Go (easier to cross-compile for Linux/Windows/embedded)
- ✅ Comprehensive: handles text shaping, BiDi, and font features
- ✅ Permissive dual license
- ✅ Single dependency for all text layout needs

**Cons**:
- ⚠️ API still evolving (v0.x - breaking changes possible)
- ⚠️ No specific Arabic examples in docs (need to test)
- ⚠️ May be more complex than needed for simple ESC/POS rendering

**Integration Effort**: Medium
- Need to understand HarfBuzz shaping API
- Need to integrate with ESC/POS command generation
- Requires font loading (ESC/POS printers have built-in fonts - need to map)

**Example packages**:
```go
import (
    "github.com/go-text/typesetting/harfbuzz"  // Text shaping
    "github.com/go-text/typesetting/font"      // Font loading
    "github.com/go-text/typesetting/segmenter" // Text segmentation
)
```

---

### Option 2: `github.com/benoitkugler/textlayout` ❌ **DEPRECATED**

**Source**: https://pkg.go.dev/github.com/benoitkugler/textlayout
**License**: Unlicense
**Last Updated**: November 14, 2024
**Status**: ⚠️ **Merged into go-text/typesetting - No longer maintained**

**What it was**:
- Port of HarfBuzz and FriBidi from C to Go
- Included both text shaping and BiDi algorithm

**Recommendation**: **Do NOT use** - Authors state it will not be maintained. Use `go-text/typesetting` instead (the successor).

---

### Option 3: `github.com/01walid/goarabic` ⚠️ SUPPLEMENTARY ONLY

**Source**: https://github.com/01walid/goarabic
**License**: MIT ✅
**Last Updated**: Unknown (21 commits total)
**Version**: No releases, use master branch

**What it provides**:
- Glyph representation for image/PDF generation
- Strip Tashkeel (Arabic vowel marks)
- SmartLength (length excluding diacritics)
- Strip Tatweel (Arabic elongation character)
- UTF-8 reversal (reverse Arabic but preserve Latin)

**Pros**:
- ✅ Simple, focused on Arabic
- ✅ MIT licensed
- ✅ Has `ToGlyph()` function for Arabic text shaping
- ✅ Handles diacritics and special characters

**Cons**:
- ⚠️ Limited documentation
- ⚠️ No versioning/releases
- ⚠️ Unclear if actively maintained
- ⚠️ Does not handle BiDi algorithm (only reversal)
- ⚠️ May not handle all joining rules correctly

**Use Case**:
- Could be used as a supplementary library for Arabic-specific utilities
- `RemoveTashkeel()` could be useful for simplified printing
- `ToGlyph()` could be used if we need simpler Arabic handling

**Integration Effort**: Low
```go
import "github.com/01walid/goarabic"

shaped := goarabic.ToGlyph("نص عربي")
```

**Recommendation**: Use as **supplement** to go-text/typesetting for Arabic-specific utilities, but NOT as primary shaping engine.

---

### Option 4: CGO with C HarfBuzz/FriBidi ❌ NOT RECOMMENDED

**What it would be**:
- Use `cgo` to call native C libraries
- HarfBuzz: `libharfbuzz-dev`
- FriBidi: `libfribidi-dev`

**Pros**:
- ✅ Most mature/battle-tested implementations
- ✅ Used by browsers, Pango, etc.

**Cons**:
- ❌ Requires CGO (complicates cross-compilation)
- ❌ Requires C libraries on target system
- ❌ Harder to deploy (system dependencies)
- ❌ Cross-platform challenges (Windows, embedded Linux)

**Recommendation**: **Avoid** - Pure Go solution is preferable for device bridge deployment

---

## Recommended Solution

### Primary: `go-text/typesetting`

```go
import (
    "github.com/go-text/typesetting/harfbuzz"
    "github.com/go-text/typesetting/font"
    "github.com/go-text/typesetting/language"
)

func shapeArabicText(text string, fontFace *font.Face) []Glyph {
    // Create HarfBuzz buffer
    buffer := harfbuzz.NewBuffer()
    buffer.AddRunes([]rune(text), 0, -1)

    // Set language and direction
    buffer.Props.Language = language.NewLanguage("ar") // Arabic
    buffer.Props.Direction = harfbuzz.RightToLeft

    // Shape the text
    buffer.Shape(fontFace, nil)

    // Get positioned glyphs
    info := buffer.Info
    positions := buffer.Pos

    // Convert to ESC/POS commands...
    return convertToGlyphs(info, positions)
}
```

### Supplementary: `goarabic` (optional)

```go
import "github.com/01walid/goarabic"

// For stripping diacritics if printer doesn't support them
simplified := goarabic.RemoveTashkeel("نًصٌ عَربيُّ")
// Output: "نص عربي"
```

---

## Implementation Approach

### Phase 1: Prototype (Week 1)
1. Add dependency: `go get github.com/go-text/typesetting`
2. Create test file: `internal/drivers/printer_escpos/arabic_test.go`
3. Test cases:
   - Simple Arabic word: `"مرحباً"` (Hello)
   - Mixed Arabic/English: `"Coca-Cola كولا"`
   - Arabic numbers: `"١٢٣٤٥"`
   - Long text with wrapping

### Phase 2: Integration (Week 2)
1. Modify `internal/drivers/printer_escpos/renderer.go`
2. Detect Arabic text (check if any RTL characters)
3. Route Arabic text through shaping engine
4. Generate ESC/POS commands for shaped glyphs
5. Handle mixed LTR/RTL runs

### Phase 3: Testing (Week 3)
1. Unit tests with various Arabic strings
2. Test with real printers (Epson, Star, Chinese brands)
3. Test alignment (right-align for Arabic amounts)
4. Test line wrapping
5. Validate with native Arabic speakers

---

## Key Challenges

### Challenge 1: Font Mapping
**Problem**: ESC/POS printers have built-in fonts, not external font files
**Solution**:
- Create a virtual font representation for HarfBuzz
- Map shaped glyphs back to ESC/POS character codes
- May need to use ESC/POS "international character set" commands

### Challenge 2: Glyph Positioning
**Problem**: ESC/POS is character-cell based, not pixel-based
**Solution**:
- HarfBuzz returns glyph positions (x, y offsets)
- Round to nearest character position
- May lose some precision but acceptable for receipts

### Challenge 3: Mixed Scripts
**Problem**: "Coca-Cola كولا 12.50" has LTR and RTL runs
**Solution**:
- Use BiDi algorithm to split into runs
- Shape each run separately
- Reorder runs for final output

### Challenge 4: Printer Compatibility
**Problem**: Not all ESC/POS printers support Arabic character sets
**Solution**:
- Test with multiple printer models
- Document required printer capabilities (Code Page 864 for Arabic)
- Fall back to transliteration if not supported

---

## Dependencies to Add

```go
// go.mod
require (
    github.com/go-text/typesetting v0.2.1
    // Optional:
    github.com/01walid/goarabic v0.0.0-latest
)
```

---

## Testing Plan

### Test Strings
1. **Simple Words**:
   - `"مرحباً"` - Hello
   - `"شكراً"` - Thank you
   - `"السعر"` - Price

2. **Sentences**:
   - `"إجمالي الفاتورة"` - Invoice total
   - `"مع تمنياتنا بزيارة سعيدة"` - Have a nice visit

3. **Mixed Scripts**:
   - `"Coca-Cola كولا"`
   - `"السعر: 25.50 ر.س"` - Price: 25.50 SAR

4. **Numbers**:
   - Western: `"123.45"`
   - Eastern Arabic: `"١٢٣٫٤٥"`

5. **Long Text** (word wrapping):
   - `"محل البقالة الكبير للمواد الغذائية والمشروبات والمنتجات الطازجة"`

### Success Criteria
- ✅ Arabic characters render in correct connected forms
- ✅ Text flows right-to-left
- ✅ Mixed Arabic/English lines render correctly
- ✅ Numbers align properly (right-align for amounts)
- ✅ Works on 3+ printer models (Epson, Star, Chinese)
- ✅ Validated by native Arabic speakers

---

## Estimated Effort

- **Research & Setup**: 2 days (DONE)
- **Prototype**: 3 days
- **Integration**: 3-4 days
- **Testing**: 2-3 days
- **Documentation**: 1 day
- **Total**: 3 weeks (as estimated in IMPLEMENTATION_PLAN.md)

---

## Next Steps

1. ✅ Research complete - Decision: use `go-text/typesetting`
2. ⏭️ Add dependency to `go.mod`
3. ⏭️ Create prototype in `arabic_test.go`
4. ⏭️ Test with sample Arabic strings
5. ⏭️ Integrate into ESC/POS renderer
6. ⏭️ Test with real printers

---

## References

- go-text/typesetting: https://pkg.go.dev/github.com/go-text/typesetting
- Unicode BiDi Algorithm: https://unicode.org/reports/tr9/
- HarfBuzz Manual: https://harfbuzz.github.io/
- ESC/POS Command Reference: https://reference.epson-biz.com/modules/ref_escpos/
- Arabic Unicode Range: U+0600 to U+06FF
- goarabic: https://github.com/01walid/goarabic

---

## Decision Log

- **2025-11-13**: Evaluated 4 options, selected `go-text/typesetting`
- **Rationale**: Pure Go, actively maintained, comprehensive, production-proven
- **Risk**: API may change (v0.x), but benefits outweigh risks
- **Fallback**: If go-text proves too complex, consider goarabic.ToGlyph() as simpler alternative

---

**Status**: ✅ Research Complete
**Next**: Proceed to Phase 1.1.2 (Create Prototype)
