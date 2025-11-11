// +build windows

package scanner_hid

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/gousb"
)

// windowsUSBDevice implements USBDevice for Windows using gousb
type windowsUSBDevice struct {
	device   *gousb.Device
	intf     *gousb.Interface
	endpoint *gousb.InEndpoint
	config   *gousb.Config

	// Device info
	vendorID  uint16
	productID uint16
	serial    string

	// Reading state
	mu       sync.Mutex
	timeout  time.Duration
	closed   bool
}

// findUSBDevice finds and opens a USB HID scanner on Windows
func findUSBDevice(config Config) (USBDevice, error) {
	// Create USB context
	ctx := gousb.NewContext()

	// Find devices matching vendor/product ID
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// If specific vendor/product specified, match exactly
		if config.VendorID != 0 {
			return desc.Vendor == gousb.ID(config.VendorID) &&
				desc.Product == gousb.ID(config.ProductID)
		}

		// Otherwise, match any HID device (class 0x03)
		// Check if device has HID interface
		for _, cfg := range desc.Configs {
			for _, intf := range cfg.Interfaces {
				for _, setting := range intf.AltSettings {
					if setting.Class == gousb.ClassHID {
						return true
					}
				}
			}
		}
		return false
	})

	if err != nil {
		ctx.Close()
		return nil, fmt.Errorf("failed to enumerate USB devices: %w", err)
	}

	if len(devs) == 0 {
		ctx.Close()
		return nil, fmt.Errorf("no matching USB HID scanner found")
	}

	// Use first matching device
	device := devs[0]

	// Close other devices
	for i := 1; i < len(devs); i++ {
		devs[i].Close()
	}

	// Get device info
	desc, err := device.Desc()
	if err != nil {
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("failed to get device descriptor: %w", err)
	}

	vendorID := uint16(desc.Vendor)
	productID := uint16(desc.Product)

	// Get serial number
	serial, err := device.SerialNumber()
	if err != nil {
		serial = "" // Serial number is optional
	}

	// Try to match with serial if specified
	if config.Serial != "" && serial != config.Serial {
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("device serial %s does not match requested %s", serial, config.Serial)
	}

	// Note: Windows uses WinUSB driver or libusb-win32
	// Driver must be installed using Zadig or similar tool
	// SetAutoDetach is not applicable on Windows

	// Open default configuration (usually config 1)
	cfg, err := device.Config(1)
	if err != nil {
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("failed to set configuration: %w", err)
	}

	// Find HID interface
	var hidInterface *gousb.Interface
	var hidSetting gousb.InterfaceSetting

	configDesc, err := device.ActiveConfigNum()
	if err != nil {
		cfg.Close()
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("failed to get active config: %w", err)
	}

	cfgDesc := desc.Configs[configDesc-1]

	for _, intf := range cfgDesc.Interfaces {
		for _, setting := range intf.AltSettings {
			if setting.Class == gousb.ClassHID {
				hidInterface, err = cfg.Interface(intf.Number, setting.Alternate)
				if err != nil {
					continue
				}
				hidSetting = setting
				break
			}
		}
		if hidInterface != nil {
			break
		}
	}

	if hidInterface == nil {
		cfg.Close()
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("no HID interface found")
	}

	// Find interrupt IN endpoint
	var inEndpoint *gousb.InEndpoint
	for _, endpoint := range hidSetting.Endpoints {
		if endpoint.Direction == gousb.EndpointDirectionIn &&
			endpoint.TransferType == gousb.TransferTypeInterrupt {
			inEndpoint, err = hidInterface.InEndpoint(endpoint.Number)
			if err != nil {
				continue
			}
			break
		}
	}

	if inEndpoint == nil {
		hidInterface.Close()
		cfg.Close()
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("no interrupt IN endpoint found")
	}

	timeout := config.ReadTimeout
	if timeout == 0 {
		timeout = 1 * time.Second
	}

	return &windowsUSBDevice{
		device:    device,
		intf:      hidInterface,
		endpoint:  inEndpoint,
		config:    cfg,
		vendorID:  vendorID,
		productID: productID,
		serial:    serial,
		timeout:   timeout,
	}, nil
}

// Read reads a HID report from the USB device
func (d *windowsUSBDevice) Read(buffer []byte) (int, error) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return 0, fmt.Errorf("device closed")
	}
	d.mu.Unlock()

	// Read with timeout
	n, err := d.endpoint.ReadTimeout(buffer, d.timeout)
	if err != nil {
		return 0, fmt.Errorf("USB read error: %w", err)
	}

	return n, nil
}

// Close closes the USB device
func (d *windowsUSBDevice) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true

	// Close in reverse order
	if d.endpoint != nil {
		// Endpoint doesn't need explicit close
		d.endpoint = nil
	}

	if d.intf != nil {
		d.intf.Close()
		d.intf = nil
	}

	if d.config != nil {
		d.config.Close()
		d.config = nil
	}

	if d.device != nil {
		d.device.Close()
		d.device = nil
	}

	return nil
}

// GetInfo returns device information
func (d *windowsUSBDevice) GetInfo() (vendorID, productID uint16, serial string, err error) {
	return d.vendorID, d.productID, d.serial, nil
}

// EnumerateScanners lists all connected USB HID scanners
//
// This is a helper function for device discovery.
// It can be used by the auto-discovery system to find available scanners.
func EnumerateScanners() ([]ScannerInfo, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	var scanners []ScannerInfo

	// Open all devices for inspection
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// Look for HID devices
		for _, cfg := range desc.Configs {
			for _, intf := range cfg.Interfaces {
				for _, setting := range intf.AltSettings {
					if setting.Class == gousb.ClassHID {
						return true
					}
				}
			}
		}
		return false
	})

	if err != nil {
		return nil, fmt.Errorf("failed to enumerate devices: %w", err)
	}

	for _, dev := range devs {
		desc, err := dev.Desc()
		if err != nil {
			dev.Close()
			continue
		}

		// Get device strings
		vendor, _ := dev.Manufacturer()
		product, _ := dev.Product()
		serial, _ := dev.SerialNumber()

		// Get bus and address
		busNum, devAddr, err := dev.BusNumber()
		if err != nil {
			dev.Close()
			continue
		}

		info := ScannerInfo{
			VendorID:   uint16(desc.Vendor),
			ProductID:  uint16(desc.Product),
			Vendor:     vendor,
			Product:    product,
			Serial:     serial,
			BusNumber:  busNum,
			DeviceAddr: devAddr,
		}

		scanners = append(scanners, info)
		dev.Close()
	}

	return scanners, nil
}

// ScannerInfo contains information about a discovered scanner
type ScannerInfo struct {
	VendorID   uint16
	ProductID  uint16
	Vendor     string
	Product    string
	Serial     string
	BusNumber  int
	DeviceAddr int
}

// String returns a human-readable description
func (s ScannerInfo) String() string {
	vendor := s.Vendor
	if vendor == "" {
		vendor = fmt.Sprintf("VID:0x%04x", s.VendorID)
	}

	product := s.Product
	if product == "" {
		product = fmt.Sprintf("PID:0x%04x", s.ProductID)
	}

	serial := s.Serial
	if serial == "" {
		serial = "(no serial)"
	}

	return fmt.Sprintf("%s %s S/N:%s [bus %d addr %d]",
		vendor, product, serial, s.BusNumber, s.DeviceAddr)
}

// IsKnownScanner checks if this device is in our known scanner database
func (s ScannerInfo) IsKnownScanner() bool {
	for _, known := range KnownScanners {
		if s.VendorID == known.VendorID && s.ProductID == known.ProductID {
			return true
		}
	}
	return false
}

// GetKnownModel returns the model name if this is a known scanner
func (s ScannerInfo) GetKnownModel() string {
	for _, known := range KnownScanners {
		if s.VendorID == known.VendorID && s.ProductID == known.ProductID {
			return fmt.Sprintf("%s %s", known.Vendor, known.Model)
		}
	}
	return ""
}

// Common USB HID scanner vendor/product IDs
//
// This list can be used for auto-detection in the discovery system.
var KnownScanners = []struct {
	VendorID  uint16
	ProductID uint16
	Vendor    string
	Model     string
}{
	// Symbol/Zebra
	{0x05e0, 0x1200, "Symbol", "LS2208"},
	{0x05e0, 0x1300, "Symbol", "DS2208"},
	{0x05e0, 0x1900, "Symbol", "DS9208"},
	{0x05e0, 0x1203, "Symbol", "LS2208 (Keyboard Wedge)"},

	// Honeywell
	{0x0c2e, 0x0200, "Honeywell", "Voyager 1200g"},
	{0x0c2e, 0x0720, "Honeywell", "Voyager 1400g"},
	{0x0c2e, 0x0b61, "Honeywell", "Xenon 1900"},
	{0x0c2e, 0x0b00, "Honeywell", "Xenon 1902"},

	// Datalogic
	{0x05f9, 0x4204, "Datalogic", "QuickScan QD2430"},
	{0x05f9, 0x4206, "Datalogic", "QuickScan QBT2430"},
	{0x05f9, 0x2206, "Datalogic", "Gryphon GD4430"},
	{0x05f9, 0x2232, "Datalogic", "Gryphon I GD4400"},

	// Code Corporation
	{0x065a, 0x0001, "Code", "CR2600"},
	{0x065a, 0x0002, "Code", "CR2700"},

	// Opticon
	{0x065a, 0x0009, "Opticon", "OPN-2001"},
	{0x065a, 0x0011, "Opticon", "OPN-2002"},

	// Metrologic (now Honeywell)
	{0x0c2e, 0x0007, "Metrologic", "MS7120 Orbit"},
	{0x0c2e, 0x0009, "Metrologic", "MS9520 Voyager"},

	// Unitech
	{0x1ec8, 0x0101, "Unitech", "MS840"},
	{0x1ec8, 0x0102, "Unitech", "MS842"},

	// Generic/Unknown (common PIDs for HID POS)
	{0x1234, 0x5678, "Generic", "HID POS Scanner"},
}
