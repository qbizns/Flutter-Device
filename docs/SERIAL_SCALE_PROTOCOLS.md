# Serial Scale Protocols Research

**Date:** 2025-11-11
**Phase:** 3, Week 4 (Stage 2: Serial Scale Driver)
**Status:** Research Complete

---

## Overview

This document provides comprehensive research on serial scale communication protocols for retail and industrial weighing scales. These protocols enable weight reading via RS-232/RS-485 serial connections.

---

## 1. Mettler Toledo MT-SICS Protocol

**Status:** ✅ Industry Standard (Recommended Primary Implementation)

### Overview
- **Full Name:** Mettler Toledo Standard Interface Command Set
- **Type:** Bidirectional ASCII protocol
- **Adoption:** Most widely used in retail and industrial
- **Documentation:** Publicly available
- **Complexity:** Medium (well-structured, documented)

### Communication Parameters
```
Baud Rate:    4800, 9600, 19200, 38400 (9600 most common)
Data Bits:    7 or 8
Parity:       Even, Odd, or None
Stop Bits:    1
Flow Control: None (typically)
```

### Command Structure
```
<CMD>[_<Param>]<CR><LF>

Where:
  CMD    = Command code (e.g., "S" for send weight)
  Param  = Optional parameter
  <CR>   = Carriage return (0x0D)
  <LF>   = Line feed (0x0A)
```

### Common Commands

| Command | Description | Example |
|---------|-------------|---------|
| `S` | Send stable weight immediately | `S<CR><LF>` |
| `SI` | Send weight immediately (stable or not) | `SI<CR><LF>` |
| `SR` | Send weight repeatedly | `SR<CR><LF>` |
| `SIR` | Send weight immediately and repeatedly | `SIR<CR><LF>` |
| `Z` | Zero the scale | `Z<CR><LF>` |
| `ZI` | Zero immediately | `ZI<CR><LF>` |
| `T` | Tare | `T<CR><LF>` |
| `TAC` | Clear tare | `TAC<CR><LF>` |
| `@` | Reset scale | `@<CR><LF>` |
| `I0` | Get scale information | `I0<CR><LF>` |

### Response Format
```
<Status><Space><Weight><Space><Unit><CR><LF>

Where:
  Status =
    S  = Stable weight
    D  = Dynamic (unstable) weight
    +  = Overload
    -  = Underload
    I  = Invalid command
    L  = Command not executable

  Weight = Weight value (decimal with sign)
  Unit   = Unit of measurement (g, kg, lb, oz, etc.)
```

### Response Examples
```
S       1234.5 g<CR><LF>      # Stable, 1234.5 grams
D        500.2 kg<CR><LF>     # Dynamic, 500.2 kilograms
S     -00012.3 lb<CR><LF>     # Stable, -12.3 pounds (tared)
+       9999.9 g<CR><LF>      # Overload
I<CR><LF>                     # Invalid command
```

### Continuous Mode
When using `SR` or `SIR`, scale sends weight continuously:
```
Send: SR<CR><LF>
Recv: S     1234.5 g<CR><LF>
Recv: S     1235.0 g<CR><LF>
Recv: S     1235.2 g<CR><LF>
...
```

Stop continuous mode by sending any command.

### Implementation Notes
- **Parser Complexity:** Low (fixed format)
- **Stability Detection:** Built-in (S/D status)
- **Error Handling:** Clear error codes
- **Recommended:** Start with this protocol

---

## 2. CAS Protocol

**Status:** ✅ Common in Retail (Secondary Implementation)

### Overview
- **Full Name:** CAS (Korean manufacturer) serial protocol
- **Type:** Bidirectional ASCII
- **Adoption:** Popular in Asia, retail POS systems
- **Documentation:** Limited public docs
- **Complexity:** Low-Medium

### Communication Parameters
```
Baud Rate:    9600 (standard)
Data Bits:    8
Parity:       None
Stop Bits:    1
Flow Control: None
```

### Command Structure
```
<STX><CMD><ETX>

Where:
  <STX> = Start of text (0x02)
  CMD   = Command character
  <ETX> = End of text (0x03)
```

### Common Commands

| Command | Code | Description |
|---------|------|-------------|
| Weight | `W` | Request current weight |
| Zero | `Z` | Zero the scale |
| Tare | `T` | Tare the scale |

### Response Format
```
<STX><Status><Weight><Unit><ETX>

Where:
  Status =
    S = Stable
    U = Unstable
    E = Error

  Weight = 6 digits (e.g., 001234 = 123.4)
  Unit   = kg, lb, g
```

### Response Examples
```
<STX>S001234kg<ETX>   # Stable, 123.4 kg
<STX>U000500g<ETX>    # Unstable, 50.0 g
<STX>E000000<ETX>     # Error
```

### Continuous Mode
Some CAS models support continuous output:
- Scale automatically sends weight at intervals (configurable)
- No command needed once enabled in scale settings
- Format same as response above

### Implementation Notes
- **Parser Complexity:** Low (binary delimiters)
- **Decimal Point:** Implied (usually 1 decimal place)
- **Checksum:** Some models include checksum byte
- **Variations:** Different CAS models may vary slightly

---

## 3. Dibal Protocol

**Status:** ✅ European Market (Tertiary Implementation)

### Overview
- **Full Name:** Dibal (Spanish manufacturer) serial protocol
- **Type:** Bidirectional ASCII/Binary
- **Adoption:** Common in Europe, especially Spain
- **Documentation:** Limited English docs
- **Complexity:** Medium

### Communication Parameters
```
Baud Rate:    9600, 19200 (9600 typical)
Data Bits:    8
Parity:       None
Stop Bits:    1
Flow Control: None
```

### Command Structure (ASCII Mode)
```
<STX><CMD><DATA><ETX><BCC>

Where:
  <STX>  = Start of text (0x02)
  CMD    = 2-character command
  DATA   = Variable length data
  <ETX>  = End of text (0x03)
  <BCC>  = Block check character (XOR checksum)
```

### Common Commands

| Command | Description |
|---------|-------------|
| `KR` | Request weight |
| `KZ` | Zero scale |
| `KT` | Tare scale |
| `KI` | Request scale info |

### Response Format
```
<STX><CMD><Status><Weight><Unit><ETX><BCC>

Status codes:
  00 = OK, stable
  01 = OK, unstable
  02 = Overload
  03 = Underload
  04 = Zero error
  05 = Tare error
```

### Response Examples
```
<STX>KR00001234kg<ETX>??   # Stable, 123.4 kg (BCC varies)
<STX>KR01000500g<ETX>??    # Unstable, 50.0 g
<STX>KR02999999<ETX>??     # Overload
```

### Implementation Notes
- **Parser Complexity:** Medium (checksum validation)
- **BCC Calculation:** XOR of all bytes between STX and ETX
- **Variations:** Multiple Dibal models with slight differences
- **Priority:** Lower (less common outside Europe)

---

## 4. Toledo 8217 Protocol

**Status:** ⚠️ Legacy (Optional Implementation)

### Overview
- **Full Name:** Toledo 8217 Serial Protocol
- **Type:** Unidirectional continuous output
- **Adoption:** Legacy systems (older POS)
- **Documentation:** Limited availability
- **Complexity:** Low (receive only)

### Communication Parameters
```
Baud Rate:    9600, 4800
Data Bits:    7
Parity:       Even
Stop Bits:    1
Flow Control: None
```

### Output Format
The scale continuously outputs weight without request:
```
<Status><Weight><Unit><CR>

Where:
  Status = 1 byte status
  Weight = 6 bytes ASCII
  Unit   = 2 bytes
  <CR>   = Carriage return
```

### Status Byte
```
Bit 0: Motion (1 = unstable)
Bit 1: Range (1 = over capacity)
Bit 2: Zero error
Bit 3: Reserved
Bit 4-7: Reserved
```

### Example Output
```
0001234lb<CR>   # Stable, 123.4 lb
1000500kg<CR>   # Unstable, 50.0 kg
2999999<CR>     # Overload
```

### Implementation Notes
- **Parser Complexity:** Very Low (continuous stream)
- **No Commands:** Receive-only protocol
- **Legacy:** Consider for backwards compatibility only
- **Priority:** Lowest

---

## 5. Generic ASCII Protocol

**Status:** ✅ Fallback Implementation (Configurable Parser)

### Overview
- **Type:** Configurable ASCII protocol
- **Use Case:** Unknown or custom scales
- **Complexity:** Low (user-configured)

### Approach
Allow user to configure:
- **Start delimiter:** None, STX, custom character
- **End delimiter:** CR, LF, CRLF, ETX, custom
- **Field separator:** Space, comma, tab, custom
- **Field order:** Status, weight, unit (configurable)
- **Weight format:** Decimal places, sign position
- **Unit mapping:** Custom unit strings

### Configuration Example
```yaml
scale:
  protocol: generic
  format:
    start: "\x02"          # STX
    end: "\x03"            # ETX
    separator: ","
    fields: [status, weight, unit]
    weight_decimals: 1
    units:
      kg: kilogram
      lb: pound
```

### Implementation Notes
- **Parser Complexity:** Medium (configurable parser)
- **Flexibility:** Supports many unknown scales
- **Testing:** Requires extensive validation
- **Priority:** After main protocols

---

## Protocol Selection Matrix

| Protocol | Priority | Complexity | Market | Use Case |
|----------|----------|------------|--------|----------|
| **MT-SICS** | 🥇 Primary | Medium | Global | Mettler Toledo scales (most common) |
| **CAS** | 🥈 Secondary | Low | Asia/Retail | CAS brand scales |
| **Dibal** | 🥉 Tertiary | Medium | Europe | Dibal brand scales |
| **Toledo 8217** | Optional | Low | Legacy | Old POS systems |
| **Generic** | Fallback | Medium | Unknown | Custom/unknown scales |

---

## Implementation Recommendations

### Phase 1: Core Protocol (Week 4)
1. **Implement MT-SICS First**
   - Most widely used
   - Best documentation
   - Good test coverage
   - Industry standard

### Phase 2: Secondary Protocols (Week 5)
2. **Add CAS Protocol**
   - Simpler than MT-SICS
   - Good for Asian markets
   - Different enough to validate architecture

3. **Add Generic Protocol**
   - Provides flexibility
   - Catches unknown scales
   - User-configurable

### Phase 3: Extended Support (Week 6)
4. **Add Dibal Protocol** (if time permits)
   - European market coverage
   - Tests checksum validation

5. **Add Toledo 8217** (if needed)
   - Legacy system support
   - Very simple receive-only

---

## Serial Port Requirements

### Hardware
- **RS-232 Serial Port:** Most common
- **RS-485:** Some industrial scales
- **USB-to-Serial Adapter:** For modern computers

### Port Configuration
Most scales support multiple configurations. Try in order:
1. **9600 baud, 8N1** (most common)
2. **9600 baud, 7E1** (older scales)
3. **4800 baud, 8N1** (legacy)
4. **19200 baud, 8N1** (high-speed)

### Auto-Detection Strategy
1. Try common baud rates (9600, 4800, 19200)
2. Send MT-SICS info command: `I0<CR><LF>`
3. Wait for response (timeout: 1 second)
4. Try other protocols if no response
5. Fall back to continuous listening (Toledo style)

---

## Weight Unit Standardization

### Standard Units
- **Metric:** g (gram), kg (kilogram), t (metric ton)
- **Imperial:** oz (ounce), lb (pound)
- **Other:** dwt (pennyweight), ozt (troy ounce), ct (carat)

### Internal Representation
Store weight as:
```go
type WeightReading struct {
    Value      float64      // Numeric value
    Unit       WeightUnit   // Standardized unit enum
    Stable     bool         // Stability indicator
    Timestamp  time.Time    // Reading time
    Raw        string       // Original response
}

type WeightUnit int

const (
    UnitGram WeightUnit = iota
    UnitKilogram
    UnitPound
    UnitOunce
    // ... more units
)
```

### Unit Conversion
Provide conversion utilities:
- Gram ↔ Kilogram ↔ Pound ↔ Ounce
- Maintain precision
- Optional auto-conversion based on config

---

## Testing Strategy

### Without Physical Scale
1. **Mock Serial Port:**
   - Create mock serial port that returns test data
   - Simulate various responses
   - Test error conditions

2. **Serial Port Loopback:**
   - Connect TX to RX (hardware loopback)
   - Send commands, receive echoed data
   - Tests serial communication

3. **Virtual Serial Ports:**
   - Use socat on Linux
   - Use com0com on Windows
   - Create virtual port pairs

### With Physical Scale
1. **Basic Reading:** Send command, verify response
2. **Stability:** Test stable vs. unstable detection
3. **Zero/Tare:** Test calibration commands
4. **Continuous Mode:** Test repeated readings
5. **Error Conditions:** Overload, underload, etc.

---

## Reference Links

### Mettler Toledo
- MT-SICS Level 0: Basic commands
- MT-SICS Level 1: Extended commands
- MT-SICS Level 2: Advanced features
- Documentation: Available from Mettler Toledo

### CAS
- CAS protocol varies by model series
- Check specific model documentation

### General Resources
- RS-232 Standard: EIA-232
- Serial Port Programming: POSIX termios
- Go Serial Libraries: go.bug.st/serial

---

## Next Steps

1. ✅ Research complete
2. ⏳ Evaluate Go serial libraries
3. ⏳ Create driver skeleton
4. ⏳ Implement MT-SICS protocol parser
5. ⏳ Create mock/virtual testing environment
6. ⏳ Add unit tests
7. ⏳ Test with physical scale (if available)

---

**Document Version:** 1.0
**Last Updated:** 2025-11-11
**Status:** Complete - Ready for implementation
