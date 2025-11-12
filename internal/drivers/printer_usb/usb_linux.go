//go:build linux

package printer_usb

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/gousb"
)

// linuxUSBDevice implements USBDevice for Linux using gousb
type linuxUSBDevice struct {
	device   *gousb.Device
	intf     *gousb.Interface
	outEp    *gousb.OutEndpoint // For sending print data
	inEp     *gousb.InEndpoint  // For reading status (optional)
	config   *gousb.Config

	// Device info
	vendorID  uint16
	productID uint16
	serial    string

	// State
	mu      sync.Mutex
	timeout time.Duration
	closed  bool
}

// findUSBPrinter finds and opens a USB ESC/POS printer on Linux
func findUSBPrinter(config Config) (USBDevice, error) {
	// Create USB context
	ctx := gousb.NewContext()

	// Find devices matching vendor/product ID or printer class
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// If specific vendor/product specified, match exactly
		if config.VendorID != 0 {
			return desc.Vendor == gousb.ID(config.VendorID) &&
				desc.Product == gousb.ID(config.ProductID)
		}

		// Otherwise, match any USB Printer class device (0x07)
		for _, cfg := range desc.Configs {
			for _, intf := range cfg.Interfaces {
				for _, setting := range intf.AltSettings {
					if setting.Class == gousb.ClassPrinter {
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
		return nil, fmt.Errorf("no matching USB printer found")
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

	// Check serial if specified
	if config.Serial != "" && serial != config.Serial {
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("device serial %s does not match requested %s", serial, config.Serial)
	}

	// Set auto-detach kernel driver (for Linux)
	device.SetAutoDetach(true)

	// Open default configuration (usually config 1)
	cfg, err := device.Config(1)
	if err != nil {
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("failed to set configuration: %w", err)
	}

	// Find printer interface
	var printerInterface *gousb.Interface
	var printerSetting gousb.InterfaceSetting

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
			if setting.Class == gousb.ClassPrinter {
				printerInterface, err = cfg.Interface(intf.Number, setting.Alternate)
				if err != nil {
					continue
				}
				printerSetting = setting
				break
			}
		}
		if printerInterface != nil {
			break
		}
	}

	if printerInterface == nil {
		cfg.Close()
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("no printer interface found")
	}

	// Find bulk OUT endpoint (for sending print data)
	var outEndpoint *gousb.OutEndpoint
	var inEndpoint *gousb.InEndpoint

	for _, endpoint := range printerSetting.Endpoints {
		if endpoint.Direction == gousb.EndpointDirectionOut &&
			endpoint.TransferType == gousb.TransferTypeBulk {
			outEndpoint, err = printerInterface.OutEndpoint(endpoint.Number)
			if err != nil {
				continue
			}
		}
		// Optional: bulk IN endpoint for reading status
		if endpoint.Direction == gousb.EndpointDirectionIn &&
			endpoint.TransferType == gousb.TransferTypeBulk {
			inEndpoint, err = printerInterface.InEndpoint(endpoint.Number)
			if err != nil {
				// Status endpoint is optional
				inEndpoint = nil
			}
		}
	}

	if outEndpoint == nil {
		printerInterface.Close()
		cfg.Close()
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("no bulk OUT endpoint found")
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &linuxUSBDevice{
		device:    device,
		intf:      printerInterface,
		outEp:     outEndpoint,
		inEp:      inEndpoint,
		config:    cfg,
		vendorID:  vendorID,
		productID: productID,
		serial:    serial,
		timeout:   timeout,
	}, nil
}

// Write sends data to the printer
func (d *linuxUSBDevice) Write(data []byte) (int, error) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return 0, fmt.Errorf("device closed")
	}
	d.mu.Unlock()

	// Write with timeout
	n, err := d.outEp.WriteTimeout(data, d.timeout)
	if err != nil {
		return 0, fmt.Errorf("USB write error: %w", err)
	}

	return n, nil
}

// Read reads data from the printer (for status queries)
func (d *linuxUSBDevice) Read(buffer []byte) (int, error) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return 0, fmt.Errorf("device closed")
	}
	if d.inEp == nil {
		d.mu.Unlock()
		return 0, fmt.Errorf("printer does not support status read")
	}
	d.mu.Unlock()

	// Read with timeout
	n, err := d.inEp.ReadTimeout(buffer, d.timeout)
	if err != nil {
		return 0, fmt.Errorf("USB read error: %w", err)
	}

	return n, nil
}

// Close closes the USB device
func (d *linuxUSBDevice) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true

	// Close in reverse order
	if d.outEp != nil {
		d.outEp = nil
	}

	if d.inEp != nil {
		d.inEp = nil
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
func (d *linuxUSBDevice) GetInfo() (vendorID, productID uint16, serial string, err error) {
	return d.vendorID, d.productID, d.serial, nil
}

// Reset resets the USB device
func (d *linuxUSBDevice) Reset() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return fmt.Errorf("device closed")
	}

	return d.device.Reset()
}

// EnumeratePrinters lists all connected USB printers
//
// This is a helper function for device discovery.
func EnumeratePrinters() ([]PrinterInfo, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	var printers []PrinterInfo

	// Open all printer class devices
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// Look for printer class devices
		for _, cfg := range desc.Configs {
			for _, intf := range cfg.Interfaces {
				for _, setting := range intf.AltSettings {
					if setting.Class == gousb.ClassPrinter {
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

		info := PrinterInfo{
			VendorID:   uint16(desc.Vendor),
			ProductID:  uint16(desc.Product),
			Vendor:     vendor,
			Product:    product,
			Serial:     serial,
			BusNumber:  busNum,
			DeviceAddr: devAddr,
		}

		printers = append(printers, info)
		dev.Close()
	}

	return printers, nil
}

// PrinterInfo contains information about a discovered printer
type PrinterInfo struct {
	VendorID   uint16
	ProductID  uint16
	Vendor     string
	Product    string
	Serial     string
	BusNumber  int
	DeviceAddr int
}

// String returns a human-readable description
func (p PrinterInfo) String() string {
	vendor := p.Vendor
	if vendor == "" {
		vendor = fmt.Sprintf("VID:0x%04x", p.VendorID)
	}

	product := p.Product
	if product == "" {
		product = fmt.Sprintf("PID:0x%04x", p.ProductID)
	}

	serial := p.Serial
	if serial == "" {
		serial = "(no serial)"
	}

	return fmt.Sprintf("%s %s S/N:%s [bus %d addr %d]",
		vendor, product, serial, p.BusNumber, p.DeviceAddr)
}

// Common USB ESC/POS printer vendor/product IDs
var KnownPrinters = []struct {
	VendorID  uint16
	ProductID uint16
	Vendor    string
	Model     string
}{
	// Epson
	{0x04b8, 0x0202, "Epson", "TM-T88II"},
	{0x04b8, 0x0005, "Epson", "TM-T88III"},
	{0x04b8, 0x0e03, "Epson", "TM-T20"},
	{0x04b8, 0x0e15, "Epson", "TM-T88V"},
	{0x04b8, 0x0e28, "Epson", "TM-T88VI"},
	{0x04b8, 0x0e01, "Epson", "TM-U220"},

	// Star Micronics
	{0x0519, 0x0001, "Star", "TSP100"},
	{0x0519, 0x0003, "Star", "TSP600"},
	{0x0519, 0x0011, "Star", "TSP650"},
	{0x0519, 0x0017, "Star", "TSP700II"},
	{0x0519, 0x001f, "Star", "TSP800II"},

	// Citizen
	{0x2730, 0x0fff, "Citizen", "CT-S310"},
	{0x2730, 0x2002, "Citizen", "CT-S601"},
	{0x2730, 0x200f, "Citizen", "CT-S801"},

	// Bixolon
	{0x1504, 0x0006, "Bixolon", "SRP-350"},
	{0x1504, 0x001d, "Bixolon", "SRP-275"},
	{0x1504, 0x0011, "Bixolon", "SRP-350plus"},
	{0x1504, 0x00a0, "Bixolon", "SRP-Q300"},

	// Custom/Seiko
	{0x0dd4, 0x0101, "Custom", "VKP80"},

	// Posiflex
	{0x0425, 0x0101, "Posiflex", "PP-8000"},
}
