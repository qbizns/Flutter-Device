package discovery

import (
	"context"
	"fmt"

	"github.com/Macber-eg/Flutter-Device/internal/drivers/printer_usb"
	"github.com/Macber-eg/Flutter-Device/internal/drivers/scanner_hid"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// USBScanner scans for USB devices (printers, scanners)
type USBScanner struct {
	logger             *telemetry.Logger
	discoverPrinters   bool
	discoverScanners   bool
	knownPrinterVIDs   map[uint16]string // VID -> Manufacturer name
	knownScannerVIDs   map[uint16]string
	knownPrinterModels map[string]string // VID:PID -> Model name
	knownScannerModels map[string]string
}

// USBScannerConfig configures USB discovery
type USBScannerConfig struct {
	DiscoverPrinters bool
	DiscoverScanners bool
}

// NewUSBScanner creates a new USB scanner
func NewUSBScanner(cfg USBScannerConfig, logger *telemetry.Logger) *USBScanner {
	scanner := &USBScanner{
		logger:             logger,
		discoverPrinters:   cfg.DiscoverPrinters,
		discoverScanners:   cfg.DiscoverScanners,
		knownPrinterVIDs:   make(map[uint16]string),
		knownScannerVIDs:   make(map[uint16]string),
		knownPrinterModels: make(map[string]string),
		knownScannerModels: make(map[string]string),
	}

	// Initialize known printer vendors
	scanner.knownPrinterVIDs[0x04b8] = "Epson"
	scanner.knownPrinterVIDs[0x0519] = "Star Micronics"
	scanner.knownPrinterVIDs[0x2730] = "Citizen"
	scanner.knownPrinterVIDs[0x1504] = "Bixolon"
	scanner.knownPrinterVIDs[0x0dd4] = "Custom"
	scanner.knownPrinterVIDs[0x0425] = "Posiflex"

	// Initialize known scanner vendors
	scanner.knownScannerVIDs[0x05e0] = "Symbol/Zebra"
	scanner.knownScannerVIDs[0x0c2e] = "Honeywell"
	scanner.knownScannerVIDs[0x05f9] = "Datalogic"
	scanner.knownScannerVIDs[0x065a] = "Code/Opticon"
	scanner.knownScannerVIDs[0x1ec8] = "Unitech"

	// Initialize known printer models
	scanner.knownPrinterModels["04b8:0202"] = "TM-T88II"
	scanner.knownPrinterModels["04b8:0005"] = "TM-T88III"
	scanner.knownPrinterModels["04b8:0e03"] = "TM-T20"
	scanner.knownPrinterModels["04b8:0e15"] = "TM-T88V"
	scanner.knownPrinterModels["04b8:0e28"] = "TM-T88VI"
	scanner.knownPrinterModels["04b8:0e01"] = "TM-U220"
	scanner.knownPrinterModels["0519:0001"] = "TSP100"
	scanner.knownPrinterModels["0519:0003"] = "TSP600"
	scanner.knownPrinterModels["0519:0011"] = "TSP650"
	scanner.knownPrinterModels["0519:0017"] = "TSP700II"
	scanner.knownPrinterModels["0519:001f"] = "TSP800II"
	scanner.knownPrinterModels["2730:0fff"] = "CT-S310"
	scanner.knownPrinterModels["2730:2002"] = "CT-S601"
	scanner.knownPrinterModels["2730:200f"] = "CT-S801"
	scanner.knownPrinterModels["1504:0006"] = "SRP-350"
	scanner.knownPrinterModels["1504:001d"] = "SRP-275"
	scanner.knownPrinterModels["1504:0011"] = "SRP-350plus"
	scanner.knownPrinterModels["1504:00a0"] = "SRP-Q300"

	// Initialize known scanner models
	scanner.knownScannerModels["05e0:1200"] = "LS2208"
	scanner.knownScannerModels["05e0:1300"] = "DS2208"
	scanner.knownScannerModels["05e0:1900"] = "DS9208"
	scanner.knownScannerModels["0c2e:0200"] = "Voyager 1200g"
	scanner.knownScannerModels["0c2e:0720"] = "Voyager 1400g"
	scanner.knownScannerModels["0c2e:0b61"] = "Xenon 1900"
	scanner.knownScannerModels["0c2e:0b00"] = "Xenon 1902"
	scanner.knownScannerModels["05f9:4204"] = "QuickScan QD2430"
	scanner.knownScannerModels["05f9:4206"] = "QuickScan QBT2430"
	scanner.knownScannerModels["05f9:2206"] = "Gryphon GD4430"

	return scanner
}

// Scan scans for USB devices
func (s *USBScanner) Scan(ctx context.Context) ([]DiscoveredDevice, error) {
	var devices []DiscoveredDevice

	// Scan for USB printers
	if s.discoverPrinters {
		printers, err := s.scanPrinters(ctx)
		if err != nil {
			s.logger.Warn("USB printer scan failed", telemetry.Error(err))
		} else {
			devices = append(devices, printers...)
		}
	}

	// Scan for USB scanners
	if s.discoverScanners {
		scanners, err := s.scanScanners(ctx)
		if err != nil {
			s.logger.Warn("USB scanner scan failed", telemetry.Error(err))
		} else {
			devices = append(devices, scanners...)
		}
	}

	s.logger.Debug("USB scan completed",
		telemetry.Int("printers", len(devices)),
	)

	return devices, nil
}

// scanPrinters scans for USB printers (Class 0x07)
func (s *USBScanner) scanPrinters(ctx context.Context) ([]DiscoveredDevice, error) {
	// Use the platform-specific printer enumeration
	printers, err := printer_usb.EnumeratePrinters()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate USB printers: %w", err)
	}

	devices := make([]DiscoveredDevice, 0, len(printers))

	for _, printer := range printers {
		// Generate device ID
		deviceID := fmt.Sprintf("usb-printer-%04x-%04x", printer.VendorID, printer.ProductID)
		if printer.Serial != "" {
			deviceID = fmt.Sprintf("%s-%s", deviceID, printer.Serial)
		}

		// Get vendor name
		vendor := printer.Vendor
		if vendor == "" {
			if name, ok := s.knownPrinterVIDs[printer.VendorID]; ok {
				vendor = name
			} else {
				vendor = fmt.Sprintf("Unknown (0x%04x)", printer.VendorID)
			}
		}

		// Get model name
		model := printer.Product
		if model == "" {
			vidpid := fmt.Sprintf("%04x:%04x", printer.VendorID, printer.ProductID)
			if modelName, ok := s.knownPrinterModels[vidpid]; ok {
				model = modelName
			} else {
				model = fmt.Sprintf("USB Printer (0x%04x)", printer.ProductID)
			}
		}

		// Create discovered device
		device := DiscoveredDevice{
			ID:        deviceID,
			Name:      fmt.Sprintf("%s %s", vendor, model),
			Kind:      "printer.usb",
			Transport: "usb",
			VendorID:  int(printer.VendorID),
			ProductID: int(printer.ProductID),
			Serial:    printer.Serial,
			Metadata: map[string]string{
				"vendor":        vendor,
				"model":         model,
				"bus":           fmt.Sprintf("%d", printer.BusNumber),
				"address":       fmt.Sprintf("%d", printer.DeviceAddr),
				"discovered_via": "usb_enumeration",
			},
		}

		devices = append(devices, device)

		s.logger.Debug("USB printer discovered",
			telemetry.String("vendor", vendor),
			telemetry.String("model", model),
			telemetry.String("serial", printer.Serial),
		)
	}

	return devices, nil
}

// scanScanners scans for USB HID scanners
func (s *USBScanner) scanScanners(ctx context.Context) ([]DiscoveredDevice, error) {
	// Use the platform-specific scanner enumeration
	scanners, err := scanner_hid.EnumerateScanners()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate USB scanners: %w", err)
	}

	devices := make([]DiscoveredDevice, 0, len(scanners))

	for _, scanner := range scanners {
		// Generate device ID
		deviceID := fmt.Sprintf("usb-scanner-%04x-%04x", scanner.VendorID, scanner.ProductID)
		if scanner.Serial != "" {
			deviceID = fmt.Sprintf("%s-%s", deviceID, scanner.Serial)
		}

		// Get vendor name
		vendor := scanner.Vendor
		if vendor == "" {
			if name, ok := s.knownScannerVIDs[scanner.VendorID]; ok {
				vendor = name
			} else {
				vendor = fmt.Sprintf("Unknown (0x%04x)", scanner.VendorID)
			}
		}

		// Get model name
		model := scanner.Product
		if model == "" {
			vidpid := fmt.Sprintf("%04x:%04x", scanner.VendorID, scanner.ProductID)
			if modelName, ok := s.knownScannerModels[vidpid]; ok {
				model = modelName
			} else {
				model = fmt.Sprintf("USB Scanner (0x%04x)", scanner.ProductID)
			}
		}

		// Create discovered device
		device := DiscoveredDevice{
			ID:        deviceID,
			Name:      fmt.Sprintf("%s %s", vendor, model),
			Kind:      "scanner.hid",
			Transport: "usb",
			VendorID:  int(scanner.VendorID),
			ProductID: int(scanner.ProductID),
			Serial:    scanner.Serial,
			Metadata: map[string]string{
				"vendor":         vendor,
				"model":          model,
				"bus":            fmt.Sprintf("%d", scanner.BusNumber),
				"address":        fmt.Sprintf("%d", scanner.DeviceAddr),
				"discovered_via": "usb_enumeration",
			},
		}

		devices = append(devices, device)

		s.logger.Debug("USB scanner discovered",
			telemetry.String("vendor", vendor),
			telemetry.String("model", model),
			telemetry.String("serial", scanner.Serial),
		)
	}

	return devices, nil
}

// Transport returns the transport name
func (s *USBScanner) Transport() string {
	return "usb"
}
