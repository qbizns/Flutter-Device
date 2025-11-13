# Arabic ESC/POS Printer Testing Guide

**Phase 4: Real Hardware Validation**

This guide provides comprehensive testing procedures for validating Arabic text rendering on ESC/POS printers.

## Table of Contents

1. [Hardware Requirements](#hardware-requirements)
2. [Pre-Testing Setup](#pre-testing-setup)
3. [Test Scenarios](#test-scenarios)
4. [Testing Checklist](#testing-checklist)
5. [Troubleshooting](#troubleshooting)
6. [Results Documentation](#results-documentation)

---

## Hardware Requirements

### Recommended Printers for Testing

Based on HARDWARE_MATRIX.md, test with these ESC/POS printers:

#### **Tier 1: Primary Test Devices (Essential)**

| Printer | Model | Code Page 864 | RTL Support | Price | Notes |
|---------|-------|---------------|-------------|-------|-------|
| **Epson** | TM-T88VI | ✅ Yes | ✅ Yes | $350 | Industry standard, excellent Arabic support |
| **Star** | TSP143III | ✅ Yes | ✅ Yes | $220 | Widely used in retail |
| **Chinese Generic** | XP-N160II | ⚠️ Varies | ⚠️ Varies | $80 | Test common budget printer |

#### **Tier 2: Secondary Test Devices (Recommended)**

| Printer | Model | Code Page 864 | Notes |
|---------|-------|---------------|-------|
| **Bixolon** | SRP-350III | ✅ Yes | Popular in Middle East |
| **Citizen** | CT-S310II | ✅ Yes | Good Arabic support |
| **Zebra** | ZD220 | ⚠️ Limited | Label printer variant |

### Hardware Budget

- **Minimum**: $650 (Epson + Star + Chinese)
- **Recommended**: $1,200 (All Tier 1 + 1-2 Tier 2)
- **Comprehensive**: $2,000+ (All devices + variants)

### Required Accessories

- USB cables (Type-A to Type-B)
- Network cables (if testing TCP/IP)
- Serial cables (RS-232) for legacy testing
- Power adapters
- Thermal receipt paper rolls (80mm width)

### Software Requirements

- Flutter-Device Bridge (this project)
- gRPC client tools (grpcurl, Bloom RPC)
- Network scanner (nmap) for IP discovery
- Serial port tools (minicom, screen) for debugging

---

## Pre-Testing Setup

### 1. Printer Configuration

#### Code Page Setup

**Critical**: Ensure the printer firmware supports Code Page 864 (Arabic).

**Verification Steps:**

1. **Print self-test page**:
   - Most printers: Hold FEED button while powering on
   - Look for "Code Page 864" in supported character sets

2. **Check manufacturer documentation**:
   - Download ESC/POS command manual
   - Verify `ESC t 30` (0x1B 0x74 0x1E) is supported

3. **Test Code Page command**:
   ```bash
   # Send test via USB/Serial
   echo -ne "\x1B\x40\x1B\x74\x1E\xC7\xC8\xC9\x0A" > /dev/usb/lp0
   ```
   Expected: Arabic characters ا أ إ should print

#### Right-to-Left Mode

**Verification Steps:**

1. **Check RTL support**:
   - ESC/POS command: `ESC { n` (0x1B 0x7B n)
   - n = 1: RTL enabled
   - n = 0: LTR (default)

2. **Test RTL command**:
   ```bash
   # Print "ABC" in RTL (should appear as "CBA")
   echo -ne "\x1B\x40\x1B\x7B\x01ABC\x1B\x7B\x00\x0A" > /dev/usb/lp0
   ```

### 2. Device Bridge Configuration

#### Connect Printer

**USB Connection:**
```yaml
# config.yaml
printers:
  - id: epson-tm-t88vi
    driver: escpos
    connection:
      type: usb
      device: /dev/usb/lp0
    metadata:
      model: "Epson TM-T88VI"
      location: "Test Lab"
```

**Network (TCP) Connection:**
```yaml
printers:
  - id: star-tsp143iii
    driver: escpos
    connection:
      type: tcp
      address: "192.168.1.100:9100"
    metadata:
      model: "Star TSP143III"
```

#### Verify Connection

```bash
# List devices
grpcurl -plaintext localhost:50051 device.DeviceService/ListDevices

# Get device health
grpcurl -plaintext -d '{"device_id":"epson-tm-t88vi"}' \
  localhost:50051 device.DeviceService/GetDeviceHealth
```

---

## Test Scenarios

### Test Suite 1: Character Encoding

**Objective**: Verify Code Page 864 character mapping

#### Test 1.1: Basic Arabic Letters

**Test Data:**
```
Input: "مرحبا"  (Hello)
Expected: Prints as "مرحبا" (readable Arabic text)
```

**Test Script:**
```go
doc := &printer.PrintDocument{
	Sections: []printer.Section{
		{
			Lines: []printer.Line{
				{
					Runs: []printer.Run{
						{Text: "مرحبا", Style: printer.TextStyle{}},
					},
				},
			},
		},
	},
	Options: printer.PrintOptions{Cut: true},
}

err := printerClient.Print(ctx, doc)
```

**Pass Criteria:**
- ✅ Text prints without errors
- ✅ All 5 Arabic letters are visible
- ✅ Text is readable (not garbled)
- ✅ No placeholder symbols (□, ?)

#### Test 1.2: Full Arabic Alphabet

**Test Data:**
```
ا ب ت ث ج ح خ د ذ ر ز س ش ص ض ط ظ ع غ ف ق ك ل م ن ه و ي
```

**Pass Criteria:**
- ✅ All 28 letters print correctly
- ✅ No missing characters
- ✅ Visual distinction between similar letters (ب ت ث)

#### Test 1.3: Arabic Diacritics

**Test Data:**
```
مَرْحَباً بِكُمْ
(Marhaban bikum - Welcome [with diacritics])
```

**Pass Criteria:**
- ✅ Diacritics visible (َ ُ ِ ً ْ)
- ✅ Diacritics positioned correctly (above/below letters)

#### Test 1.4: Arabic-Indic Numerals

**Test Data:**
```
Western: 0 1 2 3 4 5 6 7 8 9
Eastern: ٠ ١ ٢ ٣ ٤ ٥ ٦ ٧ ٨ ٩
```

**Pass Criteria:**
- ✅ Both number systems print correctly
- ✅ Distinguishable from each other

#### Test 1.5: Arabic Punctuation

**Test Data:**
```
؟ (Arabic question mark)
، (Arabic comma)
؛ (Arabic semicolon)
```

**Pass Criteria:**
- ✅ All punctuation marks visible
- ✅ Correctly oriented for RTL text

### Test Suite 2: Text Direction

**Objective**: Verify right-to-left text rendering

#### Test 2.1: Pure RTL Text

**Test Data:**
```
Input: "مرحبا بكم"
Expected: Text flows right-to-left
Visual: م ر ح ب ا ␣ ب ك م  →  م ك ب ␣ ا ب ح ر م
```

**Validation:**
- ✅ First character (م) appears on the right
- ✅ Last character (م) appears on the left
- ✅ Reading direction is natural for Arabic speakers

#### Test 2.2: Pure LTR Text

**Test Data:**
```
Input: "Hello World"
Expected: Text flows left-to-right (normal)
```

**Pass Criteria:**
- ✅ English text prints normally
- ✅ No interference from RTL mode

#### Test 2.3: Mixed BiDi Text

**Test Data:**
```
Input: "Coca-Cola كولا"
Expected: "Coca-Cola" (LTR) + " " + "كولا" (RTL)
```

**Validation:**
- ✅ English portion reads left-to-right
- ✅ Arabic portion reads right-to-left
- ✅ Proper separation between runs
- ✅ No character mirroring

#### Test 2.4: Numbers in RTL Context

**Test Data:**
```
Input: "السعر: 25.50 ر.س"
Expected: Price label with Western digits
```

**Pass Criteria:**
- ✅ "السعر:" appears on the right
- ✅ Numbers "25.50" read left-to-right (not reversed)
- ✅ Currency "ر.س" appears after numbers
- ✅ Overall flow is RTL, but numbers maintain LTR

### Test Suite 3: Real-World Scenarios

**Objective**: Test actual receipt content

#### Test 3.1: Receipt Header

**Test Data:**
```
مطعم البرج
Tower Restaurant
شارع الملك فهد، الرياض
King Fahd St, Riyadh
================================
```

**Pass Criteria:**
- ✅ Arabic business name readable
- ✅ English name readable
- ✅ Mixed address lines formatted correctly
- ✅ Separator line renders properly

#### Test 3.2: Product List

**Test Data:**
```
المنتج                    السعر
Product                   Price
--------------------------------
Coca-Cola كولا           5.00 ر.س
Big Mac بيج ماك         25.00 ر.س
French Fries بطاطس      8.50 ر.س
```

**Pass Criteria:**
- ✅ Column alignment maintained
- ✅ Mixed-script product names readable
- ✅ Prices aligned correctly
- ✅ Currency symbols display properly

#### Test 3.3: Receipt Footer

**Test Data:**
```
المجموع الفرعي:        38.50 ر.س
Subtotal:               38.50 SAR

الضريبة (15%):          5.78 ر.س
Tax (15%):              5.78 SAR

الإجمالي:              44.28 ر.س
Total:                 44.28 SAR

شكراً لزيارتكم!
Thank you for visiting!
```

**Pass Criteria:**
- ✅ Line labels in both languages
- ✅ Numbers aligned
- ✅ Thank you message centered

#### Test 3.4: Date and Time

**Test Data:**
```
التاريخ: ٢٠٢٤/٠١/١٥
Date: 2024/01/15

الوقت: ١٤:٣٠:٤٥
Time: 14:30:45
```

**Pass Criteria:**
- ✅ Eastern Arabic numerals in dates
- ✅ Western numerals in dates
- ✅ Time format correct in both

#### Test 3.5: Barcode with Arabic Label

**Test Data:**
```
رقم الفاتورة: INV-2024-001
Invoice No: INV-2024-001

[BARCODE: 1234567890128]

للاستفسارات: ٩٢٠٠١٢٣٤٥
Inquiries: 920012345
```

**Pass Criteria:**
- ✅ Barcode prints correctly
- ✅ Arabic label above barcode
- ✅ Eastern Arabic phone number readable

### Test Suite 4: Text Styling

**Objective**: Verify styled Arabic text rendering

#### Test 4.1: Bold Arabic

**Test Data:**
```
Normal: مرحبا
Bold: **مرحبا**
```

**Pass Criteria:**
- ✅ Bold text visually thicker
- ✅ Characters remain readable
- ✅ No glyph corruption

#### Test 4.2: Underlined Arabic

**Test Data:**
```
Normal: السعر
Underlined: السعر
            _____
```

**Pass Criteria:**
- ✅ Underline appears below text
- ✅ Underline spans full width
- ✅ No interference with diacritics

#### Test 4.3: Double-Width Arabic

**Test Data:**
```
Normal: مرحبا
Double-Width: م ر ح ب ا  (wider)
```

**Pass Criteria:**
- ✅ Characters appear wider
- ✅ Proportions maintained
- ✅ Still readable

#### Test 4.4: Combined Styles

**Test Data:**
```
Bold + Underline: **__مرحبا__**
Double-Width + Bold: **م ر ح ب ا**
```

**Pass Criteria:**
- ✅ Multiple styles apply correctly
- ✅ No visual conflicts
- ✅ Readable

### Test Suite 5: Edge Cases

**Objective**: Test boundary conditions and error handling

#### Test 5.1: Very Long Arabic Text

**Test Data:**
```
محل البقالة الكبير للمواد الغذائية والمشروبات والمنتجات الطازجة والمجمدة والمعلبة
(Very long Arabic store name)
```

**Pass Criteria:**
- ✅ Text wraps at word boundaries
- ✅ No character truncation
- ✅ RTL maintained across lines

#### Test 5.2: Empty Strings

**Test Data:**
```
""  (empty string)
```

**Pass Criteria:**
- ✅ No crash
- ✅ Blank line printed
- ✅ No error messages

#### Test 5.3: Special Characters

**Test Data:**
```
أسعار\tخاصة!  (tab character)
خصم %50 اليوم  (percent sign)
$25 + ر.س 50  (mixed currencies)
```

**Pass Criteria:**
- ✅ Tab rendered correctly
- ✅ Special characters don't break layout
- ✅ Mixed currencies readable

#### Test 5.4: Rapid Printing

**Test Data:**
Print 10 receipts in rapid succession (< 1 second apart)

**Pass Criteria:**
- ✅ All receipts print successfully
- ✅ No garbled output
- ✅ No printer buffer overflow

---

## Testing Checklist

### Pre-Test

- [ ] Printer powered on and connected
- [ ] Paper loaded (80mm thermal paper)
- [ ] Device Bridge running and printer registered
- [ ] Printer self-test passed
- [ ] Code Page 864 confirmed supported
- [ ] RTL mode confirmed supported

### Test Execution

**Basic Tests**
- [ ] Test 1.1: Basic Arabic letters
- [ ] Test 1.2: Full Arabic alphabet
- [ ] Test 1.3: Arabic diacritics
- [ ] Test 1.4: Arabic-Indic numerals
- [ ] Test 1.5: Arabic punctuation

**Direction Tests**
- [ ] Test 2.1: Pure RTL text
- [ ] Test 2.2: Pure LTR text
- [ ] Test 2.3: Mixed BiDi text
- [ ] Test 2.4: Numbers in RTL context

**Real-World Tests**
- [ ] Test 3.1: Receipt header
- [ ] Test 3.2: Product list
- [ ] Test 3.3: Receipt footer
- [ ] Test 3.4: Date and time
- [ ] Test 3.5: Barcode with Arabic label

**Styling Tests**
- [ ] Test 4.1: Bold Arabic
- [ ] Test 4.2: Underlined Arabic
- [ ] Test 4.3: Double-width Arabic
- [ ] Test 4.4: Combined styles

**Edge Cases**
- [ ] Test 5.1: Very long Arabic text
- [ ] Test 5.2: Empty strings
- [ ] Test 5.3: Special characters
- [ ] Test 5.4: Rapid printing

### Post-Test

- [ ] All receipts collected and labeled
- [ ] Photos taken of each test result
- [ ] Results documented in test report
- [ ] Issues logged in GitHub
- [ ] Printer cleaned and powered off

---

## Troubleshooting

### Issue: Garbled Arabic Text

**Symptoms:**
- Arabic characters appear as squares (□)
- Random symbols instead of Arabic letters
- Western characters instead of Arabic

**Possible Causes:**
1. Code Page not set correctly
2. Printer doesn't support CP864
3. Firmware outdated

**Solutions:**
```bash
# 1. Verify CP864 support
# Print self-test and check supported code pages

# 2. Force Code Page 864
echo -ne "\x1B\x40\x1B\x74\x1E" > /dev/usb/lp0

# 3. Update printer firmware
# Check manufacturer website for updates

# 4. Try alternative code page
# ESC t 0x11 (Code Page 852 - Arabic)
echo -ne "\x1B\x40\x1B\x74\x11" > /dev/usb/lp0
```

### Issue: Text Direction Incorrect

**Symptoms:**
- Arabic text reads left-to-right (backwards)
- Mixed text has wrong order
- Characters mirrored

**Solutions:**
```bash
# 1. Enable RTL mode
echo -ne "\x1B\x7B\x01" > /dev/usb/lp0

# 2. Check printer RTL support
# Some printers auto-detect; others need explicit command

# 3. Verify BiDi algorithm
# Check logs for run splitting
tail -f /var/log/flutter-device/printer.log | grep -i "bidi"
```

### Issue: Printer Not Responding

**Symptoms:**
- Connection timeout
- "Device offline" error
- No output

**Solutions:**
```bash
# 1. Check physical connection
lsusb  # For USB printers
netstat -an | grep 9100  # For network printers

# 2. Test direct printing
echo "Test" > /dev/usb/lp0

# 3. Restart Device Bridge
systemctl restart flutter-device

# 4. Check printer status
grpcurl -plaintext -d '{"device_id":"printer-id"}' \
  localhost:50051 device.DeviceService/GetDeviceHealth
```

### Issue: Diacritics Missing

**Symptoms:**
- Diacritics don't appear
- Text looks plain without vowel marks

**Possible Causes:**
1. Code Page doesn't include diacritics
2. Printer font limitations
3. Diacritics stripped during processing

**Solutions:**
```bash
# 1. Verify diacritic character codes in CP864
# Fatha (َ): 0xEB, Damma (ُ): 0xEC, Kasra (ِ): 0xED

# 2. Test direct diacritic printing
echo -ne "\x1B\x40\x1B\x74\x1E\xC7\xEB\x0A" > /dev/usb/lp0
# Should print: ALEF with FATHA (اَ)

# 3. Check font capabilities
# Some printer fonts don't support combining diacritics
```

### Issue: Text Cut Off

**Symptoms:**
- Lines truncated mid-character
- Missing words at end of line
- Incomplete receipts

**Solutions:**
```go
// 1. Enable text wrapping
wrapped := WrapArabicText(text, maxWidth)

// 2. Reduce font size
style := TextStyle{DoubleWidth: false}

// 3. Check paper width setting
// Verify printer is configured for 80mm paper
```

---

## Results Documentation

### Test Report Template

```markdown
# Arabic ESC/POS Printer Test Report

**Date**: YYYY-MM-DD
**Tester**: [Name]
**Printer Model**: [e.g., Epson TM-T88VI]
**Firmware Version**: [e.g., v5.2]
**Connection**: [USB / TCP / Serial]

## Test Results Summary

| Test Suite | Tests Passed | Tests Failed | Pass Rate |
|------------|-------------|--------------|-----------|
| Character Encoding | X/5 | X/5 | XX% |
| Text Direction | X/4 | X/4 | XX% |
| Real-World Scenarios | X/5 | X/5 | XX% |
| Text Styling | X/4 | X/4 | XX% |
| Edge Cases | X/4 | X/4 | XX% |
| **TOTAL** | **X/22** | **X/22** | **XX%** |

## Detailed Results

### Test 1.1: Basic Arabic Letters
- **Result**: ✅ PASS / ❌ FAIL
- **Notes**: [Observations]
- **Photo**: `test-1-1-basic-arabic.jpg`

[... continue for all tests ...]

## Issues Found

### Issue #1: [Title]
- **Severity**: Critical / High / Medium / Low
- **Description**: [Detailed description]
- **Steps to Reproduce**: [Steps]
- **Expected**: [Expected behavior]
- **Actual**: [Actual behavior]
- **Workaround**: [If available]

## Recommendations

[List recommendations for code changes, printer settings, documentation updates]

## Appendix

### Photos
[Attach photos of test receipts]

### Logs
[Attach relevant log excerpts]

### Raw Data
[Attach raw test data if applicable]
```

### Photo Documentation

**Requirements:**
- High resolution (at least 1920x1080)
- Good lighting
- Entire receipt visible
- Clear text (not blurry)
- Labeled with test number

**Naming Convention:**
```
test-[suite]-[number]-[description]-[result].jpg

Examples:
test-1-1-basic-arabic-pass.jpg
test-2-3-mixed-bidi-fail.jpg
test-3-2-product-list-pass.jpg
```

### Results Archive

**Structure:**
```
test-results/
├── YYYY-MM-DD-[printer-model]/
│   ├── report.md
│   ├── photos/
│   │   ├── test-1-1-basic-arabic-pass.jpg
│   │   ├── test-1-2-full-alphabet-pass.jpg
│   │   └── ...
│   ├── logs/
│   │   ├── device-bridge.log
│   │   ├── printer-driver.log
│   │   └── grpc-traces.log
│   └── raw-data/
│       ├── test-suite-1.json
│       └── ...
```

---

## Next Steps

After completing testing:

1. **Compile results** from all printer models
2. **Update GAPS.md** with remaining issues
3. **Create GitHub issues** for bugs found
4. **Update printer compatibility matrix** in HARDWARE_MATRIX.md
5. **Document workarounds** for known limitations
6. **Proceed to Phase 5**: Native Arabic speaker validation

---

## References

- **ESC/POS Command Reference**: [Epson ESC/POS Manual](https://reference.epson-biz.com/modules/ref_escpos/)
- **Code Page 864**: [Unicode CP864 Mapping](https://www.unicode.org/Public/MAPPINGS/VENDORS/MICSFT/PC/CP864.TXT)
- **BiDi Algorithm**: [Unicode UAX#9](https://www.unicode.org/reports/tr9/)
- **Hardware Matrix**: `docs/HARDWARE_MATRIX.md`
- **Integration Design**: `docs/ARABIC_INTEGRATION_DESIGN.md`
