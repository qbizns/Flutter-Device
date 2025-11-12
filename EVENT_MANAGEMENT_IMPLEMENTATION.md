# Event Management Hardware - Implementation Guide
## Phase 1: Badge Printing & Access Control (Month 1)

**Goal:** Transform Device Bridge into the go-to solution for event management hardware integration, starting with the most critical components: badge printing, RFID access control, and attendee tracking.

---

## 🎯 Month 1 Priorities

### Week 1-2: Zebra Card/Badge Printers ⭐ HIGHEST PRIORITY
### Week 3-4: RFID/NFC Readers for Badge Scanning ⭐ HIGH PRIORITY
### Week 5-6: Access Control Systems (Doors, Gates, Turnstiles) ⭐ HIGH PRIORITY

---

## 📋 Week 1-2: Zebra Badge Printer Implementation

### Overview
Zebra card printers (ZC100, ZC300, ZC350, ZXP Series) are the industry standard for event badge printing. They support:
- PVC card printing (CR80 size: 85.6mm x 53.98mm)
- Magnetic stripe encoding
- RFID/NFC encoding (Mifare, HID iClass)
- Dual-sided printing
- Multiple connectivity options (USB, Ethernet, WiFi)

### Technical Requirements

#### Hardware Support
**Primary Models:**
- Zebra ZC100 (entry-level, single-side)
- Zebra ZC300 (mid-range, dual-side, encoding)
- Zebra ZC350 (high-volume, dual-side, multiple encoders)
- Zebra ZXP Series 7 (retransfer, high-quality)

**Communication Protocols:**
- **USB**: Virtual COM port or Direct USB
- **Network**: TCP/IP on port 9100 (like network printers)
- **Driver SDK**: Zebra Card Studio SDK (optional for advanced features)

### File Structure

```
internal/drivers/badge_printer_zebra/
├── driver.go              # Main driver implementation
├── zpl_card.go            # ZPL commands for card printing
├── card_designer.go       # Badge layout and design engine
├── encoder_magnetic.go    # Magnetic stripe encoding
├── encoder_rfid.go        # RFID/NFC encoding
├── templates.go           # Pre-built badge templates
├── status.go              # Printer status monitoring
├── driver_test.go         # Unit tests
├── integration_test.go    # Integration tests
└── mock_printer.go        # Mock for testing

proto/devicebridge/v1/
└── badge.proto            # Protocol definitions

docs/
├── BADGE_PRINTER_SETUP.md
├── BADGE_DESIGN_GUIDE.md
└── ZEBRA_TROUBLESHOOTING.md

configs/
└── config.badge-printer.example.yaml

examples/badge/
├── simple_badge.go
├── event_registration.go
└── badge_templates/
    ├── conference_attendee.json
    ├── vip_badge.json
    ├── staff_badge.json
    └── visitor_badge.json
```

### Protocol Definitions (badge.proto)

```protobuf
syntax = "proto3";

package devicebridge.v1;

import "google/protobuf/timestamp.proto";
import "devicebridge/v1/common.proto";

// Badge printer service
service BadgePrinterService {
  // Print a badge
  rpc PrintBadge(PrintBadgeRequest) returns (PrintBadgeResponse);

  // Design and preview a badge
  rpc DesignBadge(DesignBadgeRequest) returns (BadgeDesign);

  // Get predefined templates
  rpc GetBadgeTemplates(GetBadgeTemplatesRequest) returns (GetBadgeTemplatesResponse);

  // Encode card (magnetic stripe or RFID)
  rpc EncodeCard(EncodeCardRequest) returns (EncodeCardResponse);

  // Get printer status (ribbon level, cards remaining, errors)
  rpc GetPrinterStatus(GetPrinterStatusRequest) returns (PrinterStatus);

  // Subscribe to printer events
  rpc SubscribePrinterEvents(SubscribePrinterEventsRequest) returns (stream PrinterEvent);
}

message PrintBadgeRequest {
  string device_id = 1;
  BadgeDesign design = 2;
  EncodingData encoding = 3;
  PrintOptions options = 4;
}

message BadgeDesign {
  // Template-based or custom
  oneof design_type {
    string template_id = 1;  // Use predefined template
    CustomDesign custom = 2;  // Custom design
  }

  // Variable data to fill in
  map<string, string> variables = 3;

  // Front side elements
  repeated BadgeElement front_elements = 4;

  // Back side elements (if dual-sided)
  repeated BadgeElement back_elements = 5;
}

message CustomDesign {
  CardSize size = 1;
  Orientation orientation = 2;
  BackgroundColor background = 3;
  repeated BadgeElement elements = 4;
}

message BadgeElement {
  ElementType type = 1;
  Position position = 2;
  Size size = 3;
  string content = 4;
  ElementStyle style = 5;

  enum ElementType {
    TEXT = 0;
    IMAGE = 1;
    PHOTO = 2;
    LOGO = 3;
    BARCODE = 4;
    QR_CODE = 5;
    LINE = 6;
    RECTANGLE = 7;
  }
}

message Position {
  int32 x = 1;  // Pixels from left (0-1016 for 300dpi CR80)
  int32 y = 2;  // Pixels from top (0-648 for 300dpi CR80)
}

message Size {
  int32 width = 1;   // Pixels
  int32 height = 2;  // Pixels
}

message ElementStyle {
  // Text style
  string font_family = 1;
  int32 font_size = 2;
  bool bold = 3;
  bool italic = 4;
  string color = 5;         // Hex color "#RRGGBB"
  string background = 6;    // Hex color
  TextAlign alignment = 7;

  // Border
  bool has_border = 8;
  int32 border_width = 9;
  string border_color = 10;

  // Image style
  ImageFit image_fit = 11;  // "fill", "contain", "cover"

  enum TextAlign {
    LEFT = 0;
    CENTER = 1;
    RIGHT = 2;
    JUSTIFY = 3;
  }

  enum ImageFit {
    FILL = 0;
    CONTAIN = 1;
    COVER = 2;
  }
}

message EncodingData {
  MagneticStripeData magnetic_stripe = 1;
  RFIDData rfid = 2;
}

message MagneticStripeData {
  bool enabled = 1;
  string track1_data = 2;  // Up to 79 chars
  string track2_data = 3;  // Up to 40 chars
  string track3_data = 4;  // Up to 107 chars
  CoercivityType coercivity = 5;

  enum CoercivityType {
    LOCO = 0;  // 300 Oe (low coercivity)
    HICO = 1;  // 2750 Oe (high coercivity)
  }
}

message RFIDData {
  bool enabled = 1;
  RFIDType type = 2;
  bytes uid = 3;              // Card UID (4 or 7 bytes)
  repeated RFIDBlock blocks = 4;  // Data blocks to write

  enum RFIDType {
    MIFARE_CLASSIC_1K = 0;
    MIFARE_CLASSIC_4K = 1;
    MIFARE_ULTRALIGHT = 2;
    MIFARE_DESFIRE = 3;
    HID_ICLASS = 4;
    HID_PROX = 5;
  }
}

message RFIDBlock {
  int32 block_number = 1;
  bytes data = 2;  // 16 bytes for Mifare
}

message PrintOptions {
  PrintSide side = 1;
  PrintQuality quality = 2;
  int32 copies = 3;
  bool laminate = 4;

  enum PrintSide {
    FRONT_ONLY = 0;
    BACK_ONLY = 1;
    BOTH_SIDES = 2;
  }

  enum PrintQuality {
    DRAFT = 0;
    NORMAL = 1;
    HIGH = 2;
  }
}

message PrintBadgeResponse {
  string job_id = 1;
  JobStatus status = 2;
  string message = 3;
  int32 cards_printed = 4;
}

message PrinterStatus {
  string device_id = 1;
  DeviceStatus status = 2;

  // Consumables
  RibbonStatus ribbon = 3;
  int32 cards_remaining = 4;

  // Current job
  string current_job_id = 5;
  int32 cards_in_job = 6;
  int32 cards_completed = 7;

  // Errors/Warnings
  repeated PrinterError errors = 8;
  repeated PrinterWarning warnings = 9;
}

message RibbonStatus {
  string type = 1;           // "YMCKO", "KO", "Monochrome"
  int32 panels_remaining = 2;
  int32 panels_capacity = 3;
  bool low_ribbon = 4;
}

message PrinterError {
  ErrorCode code = 1;
  string message = 2;
  google.protobuf.Timestamp timestamp = 3;

  enum ErrorCode {
    NO_ERROR = 0;
    NO_RIBBON = 1;
    NO_CARDS = 2;
    CARD_JAM = 3;
    RIBBON_JAM = 4;
    PRINT_HEAD_ERROR = 5;
    ENCODER_ERROR = 6;
    COMMUNICATION_ERROR = 7;
  }
}

message PrinterWarning {
  WarningCode code = 1;
  string message = 2;

  enum WarningCode {
    LOW_RIBBON = 0;
    LOW_CARDS = 1;
    CLEANING_REQUIRED = 2;
  }
}

message PrinterEvent {
  string device_id = 1;
  EventType type = 2;
  google.protobuf.Timestamp timestamp = 3;
  string job_id = 4;
  string details = 5;

  enum EventType {
    JOB_STARTED = 0;
    JOB_COMPLETED = 1;
    JOB_FAILED = 2;
    CARD_PRINTED = 3;
    RIBBON_LOW = 4;
    CARDS_LOW = 5;
    ERROR = 6;
  }
}

message CardSize {
  SizeType type = 1;

  enum SizeType {
    CR80 = 0;  // Standard credit card (85.6 x 53.98 mm)
    CR79 = 1;  // Slightly smaller
    CUSTOM = 2;
  }
}

enum Orientation {
  PORTRAIT = 0;
  LANDSCAPE = 1;
}

message BackgroundColor {
  string hex_color = 1;  // "#RRGGBB"
}
```

### Driver Implementation (driver.go)

```go
package badge_printer_zebra

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/Macber-eg/Flutter-Device/internal/devices"
    "github.com/Macber-eg/Flutter-Device/internal/telemetry"
    pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// Driver implements badge printer for Zebra card printers
type Driver struct {
    id          string
    name        string
    config      Config
    logger      *telemetry.Logger

    // Connection
    conn        Connection
    connMu      sync.RWMutex

    // State
    status      devices.DeviceStatus
    statusMu    sync.RWMutex

    // Job queue
    jobQueue    chan *PrintJob
    jobResults  map[string]*JobResult
    jobsMu      sync.RWMutex

    // Events
    eventChan   chan *pb.PrinterEvent

    // Components
    designer    *CardDesigner
    encoder     *Encoder
    statusMon   *StatusMonitor

    // Lifecycle
    ctx         context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup
}

// Config contains driver configuration
type Config struct {
    // Connection
    Transport   string  // "usb", "tcp"
    Address     string  // For TCP
    Port        int     // For TCP
    VendorID    uint16  // For USB
    ProductID   uint16  // For USB

    // Capabilities
    DualSided   bool
    MagStripe   bool
    RFID        bool
    RFIDTypes   []string  // ["mifare_1k", "iclass"]
    Lamination  bool

    // Defaults
    DPI         int     // 300 or 600
    PrintSpeed  string  // "fast", "normal", "quality"

    // Timeouts
    ConnectTimeout  time.Duration
    PrintTimeout    time.Duration
}

// PrintJob represents a badge print job
type PrintJob struct {
    ID          string
    Design      *pb.BadgeDesign
    Encoding    *pb.EncodingData
    Options     *pb.PrintOptions
    Timestamp   time.Time
    ResultChan  chan *JobResult
}

// JobResult contains the result of a print job
type JobResult struct {
    JobID       string
    Success     bool
    Error       error
    CardsPrinted int
    Duration    time.Duration
}

// NewDriver creates a new Zebra badge printer driver
func NewDriver(id, name string, config Config, logger *telemetry.Logger) (*Driver, error) {
    ctx, cancel := context.WithCancel(context.Background())

    d := &Driver{
        id:          id,
        name:        name,
        config:      config,
        logger:      logger,
        status:      devices.StatusDisconnected,
        jobQueue:    make(chan *PrintJob, 100),
        jobResults:  make(map[string]*JobResult),
        eventChan:   make(chan *pb.PrinterEvent, 50),
        ctx:         ctx,
        cancel:      cancel,
    }

    // Initialize components
    d.designer = NewCardDesigner(config.DPI)
    d.encoder = NewEncoder(logger)
    d.statusMon = NewStatusMonitor(d, logger)

    return d, nil
}

// Start initializes the printer connection and workers
func (d *Driver) Start(ctx context.Context) error {
    d.logger.Info("starting zebra badge printer driver",
        telemetry.String("device_id", d.id),
    )

    // Connect to printer
    if err := d.connect(); err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }

    // Start job processor
    d.wg.Add(1)
    go d.processJobs()

    // Start status monitor
    d.wg.Add(1)
    go d.statusMon.Run(d.ctx)

    d.setStatus(devices.StatusReady)

    d.logger.Info("zebra badge printer started",
        telemetry.String("device_id", d.id),
    )

    return nil
}

// Stop cleanly shuts down the driver
func (d *Driver) Stop(ctx context.Context) error {
    d.logger.Info("stopping zebra badge printer driver",
        telemetry.String("device_id", d.id),
    )

    d.cancel()
    d.wg.Wait()

    if err := d.disconnect(); err != nil {
        d.logger.Error("error disconnecting printer",
            telemetry.String("device_id", d.id),
            telemetry.Error(err),
        )
    }

    close(d.jobQueue)
    close(d.eventChan)

    d.setStatus(devices.StatusDisconnected)

    return nil
}

// PrintBadge prints a badge with the given design and encoding
func (d *Driver) PrintBadge(ctx context.Context, req *pb.PrintBadgeRequest) (*pb.PrintBadgeResponse, error) {
    // Validate request
    if err := d.validatePrintRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }

    // Create job
    job := &PrintJob{
        ID:         generateJobID(),
        Design:     req.Design,
        Encoding:   req.Encoding,
        Options:    req.Options,
        Timestamp:  time.Now(),
        ResultChan: make(chan *JobResult, 1),
    }

    // Queue job
    select {
    case d.jobQueue <- job:
        d.logger.Debug("job queued",
            telemetry.String("job_id", job.ID),
            telemetry.String("device_id", d.id),
        )
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    // Wait for result
    select {
    case result := <-job.ResultChan:
        if !result.Success {
            return &pb.PrintBadgeResponse{
                JobId:  result.JobID,
                Status: pb.JobStatus_FAILED,
                Message: result.Error.Error(),
            }, nil
        }

        return &pb.PrintBadgeResponse{
            JobId:        result.JobID,
            Status:       pb.JobStatus_COMPLETED,
            Message:      "Badge printed successfully",
            CardsPrinted: int32(result.CardsPrinted),
        }, nil

    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// processJobs processes print jobs from the queue
func (d *Driver) processJobs() {
    defer d.wg.Done()

    for {
        select {
        case job := <-d.jobQueue:
            result := d.executeJob(job)
            job.ResultChan <- result
            close(job.ResultChan)

        case <-d.ctx.Done():
            return
        }
    }
}

// executeJob executes a single print job
func (d *Driver) executeJob(job *PrintJob) *JobResult {
    start := time.Now()

    d.logger.Info("executing print job",
        telemetry.String("job_id", job.ID),
        telemetry.String("device_id", d.id),
    )

    result := &JobResult{
        JobID: job.ID,
    }

    // Emit job started event
    d.emitEvent(&pb.PrinterEvent{
        DeviceId: d.id,
        Type:     pb.PrinterEvent_JOB_STARTED,
        Timestamp: timestampProto(job.Timestamp),
        JobId:    job.ID,
    })

    // 1. Render badge design to ZPL
    zpl, err := d.designer.RenderToZPL(job.Design, job.Options)
    if err != nil {
        result.Error = fmt.Errorf("render failed: %w", err)
        d.emitJobFailedEvent(job.ID, err)
        return result
    }

    // 2. Encode magnetic stripe or RFID (if requested)
    if job.Encoding != nil {
        if err := d.encoder.Encode(d.conn, job.Encoding); err != nil {
            result.Error = fmt.Errorf("encoding failed: %w", err)
            d.emitJobFailedEvent(job.ID, err)
            return result
        }
    }

    // 3. Send ZPL to printer
    copies := int(job.Options.Copies)
    if copies <= 0 {
        copies = 1
    }

    for i := 0; i < copies; i++ {
        if err := d.sendZPL(zpl); err != nil {
            result.Error = fmt.Errorf("print failed on copy %d: %w", i+1, err)
            d.emitJobFailedEvent(job.ID, err)
            return result
        }
        result.CardsPrinted++

        // Emit card printed event
        d.emitEvent(&pb.PrinterEvent{
            DeviceId: d.id,
            Type:     pb.PrinterEvent_CARD_PRINTED,
            Timestamp: timestampProto(time.Now()),
            JobId:    job.ID,
            Details:  fmt.Sprintf("Card %d of %d", i+1, copies),
        })
    }

    result.Success = true
    result.Duration = time.Since(start)

    d.logger.Info("job completed",
        telemetry.String("job_id", job.ID),
        telemetry.Int("cards_printed", result.CardsPrinted),
        telemetry.Duration("duration", result.Duration),
    )

    // Emit job completed event
    d.emitEvent(&pb.PrinterEvent{
        DeviceId: d.id,
        Type:     pb.PrinterEvent_JOB_COMPLETED,
        Timestamp: timestampProto(time.Now()),
        JobId:    job.ID,
        Details:  fmt.Sprintf("%d cards printed", result.CardsPrinted),
    })

    return result
}

// GetStatus returns the current printer status
func (d *Driver) GetStatus(ctx context.Context) (*pb.PrinterStatus, error) {
    status := d.statusMon.GetCurrentStatus()
    return status, nil
}

// Events returns a channel of printer events
func (d *Driver) Events(ctx context.Context) (<-chan *pb.PrinterEvent, error) {
    return d.eventChan, nil
}

// Helper methods...

func (d *Driver) connect() error {
    var conn Connection
    var err error

    switch d.config.Transport {
    case "usb":
        conn, err = connectUSB(d.config.VendorID, d.config.ProductID)
    case "tcp":
        conn, err = connectTCP(d.config.Address, d.config.Port)
    default:
        return fmt.Errorf("unsupported transport: %s", d.config.Transport)
    }

    if err != nil {
        return err
    }

    d.connMu.Lock()
    d.conn = conn
    d.connMu.Unlock()

    return nil
}

func (d *Driver) sendZPL(zpl string) error {
    d.connMu.RLock()
    defer d.connMu.RUnlock()

    if d.conn == nil {
        return fmt.Errorf("not connected")
    }

    _, err := d.conn.Write([]byte(zpl))
    return err
}

func (d *Driver) emitEvent(event *pb.PrinterEvent) {
    select {
    case d.eventChan <- event:
    default:
        d.logger.Warn("event channel full, dropping event")
    }
}
```

### Card Designer (card_designer.go)

```go
package badge_printer_zebra

import (
    "fmt"
    "strings"

    pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// CardDesigner creates ZPL commands for badge printing
type CardDesigner struct {
    dpi int  // 300 or 600
}

// NewCardDesigner creates a new card designer
func NewCardDesigner(dpi int) *CardDesigner {
    if dpi != 300 && dpi != 600 {
        dpi = 300  // Default
    }

    return &CardDesigner{
        dpi: dpi,
    }
}

// RenderToZPL converts a badge design to ZPL commands
func (cd *CardDesigner) RenderToZPL(design *pb.BadgeDesign, options *pb.PrintOptions) (string, error) {
    var zpl strings.Builder

    // Start ZPL
    zpl.WriteString("^XA\n")

    // Set print origin
    zpl.WriteString("^LH0,0\n")

    // Set field orientation to portrait
    zpl.WriteString("^FWN\n")

    // Render front side
    if err := cd.renderSide(&zpl, design.FrontElements, "front"); err != nil {
        return "", fmt.Errorf("front side render error: %w", err)
    }

    // If dual-sided and back elements exist
    if options.Side == pb.PrintOptions_BOTH_SIDES && len(design.BackElements) > 0 {
        zpl.WriteString("^POI\n")  // Print orientation inverted (back side)
        if err := cd.renderSide(&zpl, design.BackElements, "back"); err != nil {
            return "", fmt.Errorf("back side render error: %w", err)
        }
    }

    // Print quantity
    if options.Copies > 0 {
        zpl.WriteString(fmt.Sprintf("^PQ%d\n", options.Copies))
    }

    // End ZPL
    zpl.WriteString("^XZ\n")

    return zpl.String(), nil
}

// renderSide renders one side of the card
func (cd *CardDesigner) renderSide(zpl *strings.Builder, elements []*pb.BadgeElement, side string) error {
    for _, elem := range elements {
        switch elem.Type {
        case pb.BadgeElement_TEXT:
            cd.renderText(zpl, elem)
        case pb.BadgeElement_IMAGE:
            cd.renderImage(zpl, elem)
        case pb.BadgeElement_BARCODE:
            cd.renderBarcode(zpl, elem)
        case pb.BadgeElement_QR_CODE:
            cd.renderQRCode(zpl, elem)
        case pb.BadgeElement_RECTANGLE:
            cd.renderRectangle(zpl, elem)
        case pb.BadgeElement_LINE:
            cd.renderLine(zpl, elem)
        default:
            return fmt.Errorf("unsupported element type: %v", elem.Type)
        }
    }
    return nil
}

// renderText renders a text element
func (cd *CardDesigner) renderText(zpl *strings.Builder, elem *pb.BadgeElement) {
    x := elem.Position.X
    y := elem.Position.Y

    // Field origin
    zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

    // Font selection
    fontSize := elem.Style.FontSize
    if fontSize == 0 {
        fontSize = 24
    }

    // ZPL font command (A-Z, 0-9 fonts)
    // ^A0N,height,width
    zpl.WriteString(fmt.Sprintf("^A0N,%d,%d\n", fontSize, fontSize))

    // Field data
    content := elem.Content
    if content == "" {
        content = "N/A"
    }

    zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", content))
}

// renderImage renders an image element
func (cd *CardDesigner) renderImage(zpl *strings.Builder, elem *pb.BadgeElement) {
    // For images, we'd need to convert to GRF format
    // This is simplified - in production, use Zebra SDK
    x := elem.Position.X
    y := elem.Position.Y

    // Assuming elem.Content contains base64 encoded image or path
    // In production, convert image to ZPL ^GF command

    zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))
    // ^GFA command for graphic field
    // This would require image processing - placeholder for now
    zpl.WriteString("^GFA,<bytes>,<bytes>,<rowBytes>,<data>^FS\n")
}

// renderBarcode renders a barcode
func (cd *CardDesigner) renderBarcode(zpl *strings.Builder, elem *pb.BadgeElement) {
    x := elem.Position.X
    y := elem.Position.Y

    zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

    // Code 128 barcode (common for badges)
    // ^BC: Code 128 barcode
    // ^BCN,height,printInterpretationLine,printInterpretationLineAbove
    height := elem.Size.Height
    if height == 0 {
        height = 100
    }

    zpl.WriteString(fmt.Sprintf("^BCN,%d,Y,N,N\n", height))
    zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", elem.Content))
}

// renderQRCode renders a QR code
func (cd *CardDesigner) renderQRCode(zpl *strings.Builder, elem *pb.BadgeElement) {
    x := elem.Position.X
    y := elem.Position.Y

    zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

    // ^BQN: QR code
    // Model 2, magnification 3, error correction M
    zpl.WriteString("^BQN,2,3,M,7\n")
    zpl.WriteString(fmt.Sprintf("^FDMA,%s^FS\n", elem.Content))
}

// renderRectangle renders a rectangle
func (cd *CardDesigner) renderRectangle(zpl *strings.Builder, elem *pb.BadgeElement) {
    x := elem.Position.X
    y := elem.Position.Y
    w := elem.Size.Width
    h := elem.Size.Height

    zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

    thickness := 1
    if elem.Style != nil && elem.Style.BorderWidth > 0 {
        thickness = int(elem.Style.BorderWidth)
    }

    // ^GB: Graphic Box
    zpl.WriteString(fmt.Sprintf("^GB%d,%d,%d^FS\n", w, h, thickness))
}

// renderLine renders a line
func (cd *CardDesigner) renderLine(zpl *strings.Builder, elem *pb.BadgeElement) {
    x := elem.Position.X
    y := elem.Position.Y
    w := elem.Size.Width
    h := elem.Size.Height

    thickness := 2
    if elem.Style != nil && elem.Style.BorderWidth > 0 {
        thickness = int(elem.Style.BorderWidth)
    }

    zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

    // Horizontal or vertical line
    if h > w {
        // Vertical line
        zpl.WriteString(fmt.Sprintf("^GB%d,%d,%d^FS\n", thickness, h, thickness))
    } else {
        // Horizontal line
        zpl.WriteString(fmt.Sprintf("^GB%d,%d,%d^FS\n", w, thickness, thickness))
    }
}
```

### Badge Templates (templates.go)

```go
package badge_printer_zebra

import (
    pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// PredefinedTemplates contains common badge templates
var PredefinedTemplates = map[string]*BadgeTemplate{
    "conference_attendee": ConferenceAttendeeTemplate(),
    "vip_badge":           VIPBadgeTemplate(),
    "staff_badge":         StaffBadgeTemplate(),
    "visitor_badge":       VisitorBadgeTemplate(),
}

// BadgeTemplate represents a reusable badge template
type BadgeTemplate struct {
    ID          string
    Name        string
    Description string
    Variables   []string          // Required variables
    Design      *pb.BadgeDesign
}

// ConferenceAttendeeTemplate returns a standard conference attendee badge
func ConferenceAttendeeTemplate() *BadgeTemplate {
    return &BadgeTemplate{
        ID:          "conference_attendee",
        Name:        "Conference Attendee",
        Description: "Standard attendee badge with name, company, and QR code",
        Variables:   []string{"name", "company", "badge_id", "qr_data"},
        Design: &pb.BadgeDesign{
            FrontElements: []*pb.BadgeElement{
                // Logo at top
                {
                    Type: pb.BadgeElement_IMAGE,
                    Position: &pb.Position{X: 150, Y: 30},
                    Size:     &pb.Size{Width: 200, Height: 80},
                    Content:  "{{logo}}",
                },
                // Name (large)
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 150},
                    Size:     &pb.Size{Width: 400, Height: 60},
                    Content:  "{{name}}",
                    Style: &pb.ElementStyle{
                        FontSize:   36,
                        Bold:       true,
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Company (smaller)
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 220},
                    Size:     &pb.Size{Width: 400, Height: 40},
                    Content:  "{{company}}",
                    Style: &pb.ElementStyle{
                        FontSize:   24,
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // QR Code
                {
                    Type: pb.BadgeElement_QR_CODE,
                    Position: &pb.Position{X: 150, Y: 300},
                    Size:     &pb.Size{Width: 200, Height: 200},
                    Content:  "{{qr_data}}",
                },
                // Badge ID barcode
                {
                    Type: pb.BadgeElement_BARCODE,
                    Position: &pb.Position{X: 80, Y: 520},
                    Size:     &pb.Size{Width: 340, Height: 80},
                    Content:  "{{badge_id}}",
                },
            },
        },
    }
}

// VIPBadgeTemplate returns a VIP badge with gold border
func VIPBadgeTemplate() *BadgeTemplate {
    return &BadgeTemplate{
        ID:          "vip_badge",
        Name:        "VIP Badge",
        Description: "VIP badge with gold border and special designation",
        Variables:   []string{"name", "title", "access_level"},
        Design: &pb.BadgeDesign{
            FrontElements: []*pb.BadgeElement{
                // Gold border
                {
                    Type: pb.BadgeElement_RECTANGLE,
                    Position: &pb.Position{X: 10, Y: 10},
                    Size:     &pb.Size{Width: 980, Height: 628},
                    Style: &pb.ElementStyle{
                        BorderWidth: 5,
                        BorderColor: "#FFD700",  // Gold
                    },
                },
                // "VIP" text at top
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 50},
                    Size:     &pb.Size{Width: 400, Height: 80},
                    Content:  "VIP",
                    Style: &pb.ElementStyle{
                        FontSize:   48,
                        Bold:       true,
                        Color:      "#FFD700",
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Name
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 180},
                    Size:     &pb.Size{Width: 400, Height: 60},
                    Content:  "{{name}}",
                    Style: &pb.ElementStyle{
                        FontSize:   32,
                        Bold:       true,
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Title
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 260},
                    Size:     &pb.Size{Width: 400, Height: 40},
                    Content:  "{{title}}",
                    Style: &pb.ElementStyle{
                        FontSize:   24,
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
            },
        },
    }
}

// StaffBadgeTemplate returns a staff badge
func StaffBadgeTemplate() *BadgeTemplate {
    return &BadgeTemplate{
        ID:          "staff_badge",
        Name:        "Staff Badge",
        Description: "Staff identification badge with department",
        Variables:   []string{"name", "department", "employee_id", "photo"},
        Design: &pb.BadgeDesign{
            FrontElements: []*pb.BadgeElement{
                // "STAFF" banner
                {
                    Type: pb.BadgeElement_RECTANGLE,
                    Position: &pb.Position{X: 0, Y: 0},
                    Size:     &pb.Size{Width: 1000, Height: 80},
                    Style: &pb.ElementStyle{
                        Background: "#0000FF",  // Blue
                    },
                },
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 20},
                    Size:     &pb.Size{Width: 400, Height: 40},
                    Content:  "STAFF",
                    Style: &pb.ElementStyle{
                        FontSize:   32,
                        Bold:       true,
                        Color:      "#FFFFFF",
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Photo
                {
                    Type: pb.BadgeElement_PHOTO,
                    Position: &pb.Position{X: 80, Y: 120},
                    Size:     &pb.Size{Width: 200, Height: 250},
                    Content:  "{{photo}}",
                    Style: &pb.ElementStyle{
                        HasBorder:   true,
                        BorderWidth: 2,
                    },
                },
                // Name
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 300, Y: 140},
                    Size:     &pb.Size{Width: 600, Height: 50},
                    Content:  "{{name}}",
                    Style: &pb.ElementStyle{
                        FontSize: 28,
                        Bold:     true,
                    },
                },
                // Department
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 300, Y: 200},
                    Size:     &pb.Size{Width: 600, Height: 40},
                    Content:  "{{department}}",
                    Style: &pb.ElementStyle{
                        FontSize: 24,
                    },
                },
                // Employee ID barcode
                {
                    Type: pb.BadgeElement_BARCODE,
                    Position: &pb.Position{X: 80, Y: 450},
                    Size:     &pb.Size{Width: 340, Height: 80},
                    Content:  "{{employee_id}}",
                },
            },
        },
    }
}

// VisitorBadgeTemplate returns a visitor badge
func VisitorBadgeTemplate() *BadgeTemplate {
    return &BadgeTemplate{
        ID:          "visitor_badge",
        Name:        "Visitor Badge",
        Description: "Temporary visitor badge with date and host",
        Variables:   []string{"name", "company", "host", "date", "expires"},
        Design: &pb.BadgeDesign{
            FrontElements: []*pb.BadgeElement{
                // "VISITOR" text
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 50},
                    Size:     &pb.Size{Width: 400, Height: 60},
                    Content:  "VISITOR",
                    Style: &pb.ElementStyle{
                        FontSize:   40,
                        Bold:       true,
                        Color:      "#FF0000",  // Red
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Name
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 150},
                    Size:     &pb.Size{Width: 400, Height: 50},
                    Content:  "{{name}}",
                    Style: &pb.ElementStyle{
                        FontSize:   30,
                        Bold:       true,
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Company
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 210},
                    Size:     &pb.Size{Width: 400, Height: 40},
                    Content:  "{{company}}",
                    Style: &pb.ElementStyle{
                        FontSize:   24,
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Host
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 280},
                    Size:     &pb.Size{Width: 400, Height: 35},
                    Content:  "Visiting: {{host}}",
                    Style: &pb.ElementStyle{
                        FontSize:   20,
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Date
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 330},
                    Size:     &pb.Size{Width: 400, Height: 35},
                    Content:  "Date: {{date}}",
                    Style: &pb.ElementStyle{
                        FontSize:   20,
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
                // Expiry notice
                {
                    Type: pb.BadgeElement_TEXT,
                    Position: &pb.Position{X: 50, Y: 500},
                    Size:     &pb.Size{Width: 400, Height: 60},
                    Content:  "EXPIRES {{expires}}",
                    Style: &pb.ElementStyle{
                        FontSize:   24,
                        Bold:       true,
                        Color:      "#FF0000",
                        Alignment:  pb.ElementStyle_CENTER,
                    },
                },
            },
        },
    }
}
```

---

## Configuration Example

```yaml
# config.badge-printer.example.yaml
devices:
  - id: "badge-printer-registration"
    name: "Registration Desk Badge Printer"
    kind: "badge_printer.zebra"
    enabled: true
    transport: "usb"
    vendor_id: 0x0a5f    # Zebra Technologies
    product_id: 0x0176   # ZC300

    settings:
      # Printer capabilities
      dual_sided: true
      magnetic_stripe: true
      rfid: true
      rfid_types:
        - "mifare_1k"
        - "iclass"
      lamination: false

      # Print settings
      dpi: 300             # 300 or 600
      print_speed: "normal"  # "fast", "normal", "quality"

      # Timeouts
      connect_timeout: "30s"
      print_timeout: "60s"

    tags:
      - "registration"
      - "attendee-badges"
      - "main-desk"

  - id: "badge-printer-vip"
    name: "VIP Lounge Badge Printer"
    kind: "badge_printer.zebra"
    enabled: true
    transport: "tcp"
    address: "192.168.1.100"
    port: 9100

    settings:
      dual_sided: true
      rfid: true
      rfid_types: ["iclass"]
      print_speed: "quality"

    tags:
      - "vip"
      - "special-access"
```

---

## Testing Plan

### Unit Tests
```go
// driver_test.go
func TestDriver_PrintBadge_Success(t *testing.T) {}
func TestDriver_PrintBadge_InvalidDesign(t *testing.T) {}
func TestDriver_JobQueue_Concurrent(t *testing.T) {}
func TestDriver_Reconnection(t *testing.T) {}

// card_designer_test.go
func TestCardDesigner_RenderText(t *testing.T) {}
func TestCardDesigner_RenderQRCode(t *testing.T) {}
func TestCardDesigner_UseTemplate(t *testing.T) {}

// encoder_test.go
func TestEncoder_MagneticStripe(t *testing.T) {}
func TestEncoder_RFID_Mifare(t *testing.T) {}
```

### Integration Tests
```go
// integration_test.go
func TestIntegration_FullPrintFlow(t *testing.T) {
    // Requires mock printer or simulator
}

func TestIntegration_StatusMonitoring(t *testing.T) {}
func TestIntegration_EventStreaming(t *testing.T) {}
```

---

## Documentation Requirements

1. **BADGE_PRINTER_SETUP.md**
   - Hardware requirements
   - Driver installation (Windows/Linux/macOS)
   - USB configuration
   - Network configuration
   - Troubleshooting

2. **BADGE_DESIGN_GUIDE.md**
   - Template usage
   - Custom design creation
   - ZPL reference
   - Best practices
   - Design tips

3. **ZEBRA_TROUBLESHOOTING.md**
   - Common errors
   - Ribbon issues
   - Card jams
   - Encoding failures
   - Network connectivity

---

## Timeline

**Week 1:**
- [ ] Day 1-2: Protocol research and ZPL commands
- [ ] Day 3-4: Basic driver implementation
- [ ] Day 5-7: Card designer implementation

**Week 2:**
- [ ] Day 8-10: Encoder implementation (magnetic + RFID)
- [ ] Day 11-12: Status monitoring
- [ ] Day 13-14: Testing, documentation, examples

---

## Success Criteria

✅ Print basic badge (text + logo)
✅ Print badge with QR code
✅ Print badge with barcode
✅ Use predefined templates
✅ Encode magnetic stripe
✅ Encode RFID tag
✅ Monitor printer status
✅ Handle ribbon low warnings
✅ Handle card jam errors
✅ Stream printer events via WebSocket
✅ Support USB and network printers
✅ 80%+ test coverage
✅ Complete documentation

---

**Ready to start implementation?** Let me know and I'll proceed with creating the first driver files!
