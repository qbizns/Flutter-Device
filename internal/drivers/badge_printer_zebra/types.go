package badge_printer_zebra

import (
	"time"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// Config contains badge printer configuration
type Config struct {
	// Connection
	Transport string // "usb", "tcp"
	Address   string // For TCP
	Port      int    // For TCP
	VendorID  uint16 // For USB (0x0a5f for Zebra)
	ProductID uint16 // For USB

	// Capabilities
	DualSided  bool
	MagStripe  bool
	RFID       bool
	RFIDTypes  []string // ["mifare_1k", "iclass", etc.]
	Lamination bool

	// Printer settings
	DPI        int    // 300 or 600
	PrintSpeed string // "fast", "normal", "quality"
	Model      string // "ZC100", "ZC300", "ZC350", "ZXP7"

	// Timeouts
	ConnectTimeout time.Duration
	PrintTimeout   time.Duration
	EncodeTimeout  time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Transport:      "usb",
		DPI:            300,
		PrintSpeed:     "normal",
		ConnectTimeout: 30 * time.Second,
		PrintTimeout:   60 * time.Second,
		EncodeTimeout:  30 * time.Second,
	}
}

// PrintJob represents a badge print job
type PrintJob struct {
	ID         string
	Design     *pb.BadgeDesign
	Encoding   *pb.EncodingData
	Options    *pb.PrintOptions
	Timestamp  time.Time
	ResultChan chan *JobResult
}

// JobResult contains the result of a print job
type JobResult struct {
	JobID        string
	Success      bool
	Error        error
	CardsPrinted int
	Duration     time.Duration
	StartTime    time.Time
	EndTime      time.Time
}

// Connection represents a connection to the printer
type Connection interface {
	Write(data []byte) (int, error)
	Read(buf []byte) (int, error)
	Close() error
	IsConnected() bool
}

// PrinterInfo contains printer hardware information
type PrinterInfo struct {
	Model           string
	SerialNumber    string
	FirmwareVersion string
	Capabilities    *PrinterCapabilities
}

// PrinterCapabilities describes what the printer can do
type PrinterCapabilities struct {
	DualSided        bool
	MagneticStripe   bool
	RFID             bool
	RFIDTypes        []string
	Lamination       bool
	MaxDPI           int
	RibbonTypes      []string
	CardFeederCapacity int
}

// RibbonInfo contains ribbon status information
type RibbonInfo struct {
	Type             string
	PanelsRemaining  int
	PanelsCapacity   int
	PercentRemaining int
	LowRibbon        bool
	RibbonInstalled  bool
}

// CardFeederInfo contains card feeder status
type CardFeederInfo struct {
	CardsRemaining int
	CardsCapacity  int
	LowCards       bool
	NoCards        bool
}

// CleaningInfo contains maintenance information
type CleaningInfo struct {
	PrintsSinceCleaning int
	PrintsUntilCleaning int
	CleaningRequired    bool
	LastCleaned         time.Time
}

// ZPLCommand represents a ZPL command string
type ZPLCommand string

// Common ZPL constants
const (
	// Card dimensions at 300 DPI (CR80 standard card)
	CardWidthMM  = 85.6
	CardHeightMM = 53.98
	CardWidth300 = 1016 // pixels at 300 DPI
	CardHeight300 = 648 // pixels at 300 DPI
	CardWidth600 = 2032 // pixels at 600 DPI
	CardHeight600 = 1296 // pixels at 600 DPI

	// Zebra vendor ID
	ZebraVendorID = 0x0a5f
)

// Zebra printer product IDs
const (
	ProductIDZC100 = 0x0165
	ProductIDZC300 = 0x0176
	ProductIDZC350 = 0x0177
	ProductIDZXP7  = 0x0179
)

// ZPL command prefixes
const (
	ZPLStart        = "^XA"  // Start format
	ZPLEnd          = "^XZ"  // End format
	ZPLFieldOrigin  = "^FO"  // Field origin
	ZPLFieldData    = "^FD"  // Field data
	ZPLFieldSep     = "^FS"  // Field separator
	ZPLGraphicBox   = "^GB"  // Graphic box
	ZPLGraphicField = "^GF"  // Graphic field
	ZPLBarcode128   = "^BC"  // Code 128 barcode
	ZPLQRCode       = "^BQ"  // QR code
	ZPLFont         = "^A"   // Font selection
	ZPLPrintQty     = "^PQ"  // Print quantity
	ZPLLabelHome    = "^LH"  // Label home position
	ZPLFieldOrientation = "^FW" // Field orientation
	ZPLPrintOrientation = "^PO" // Print orientation
)
