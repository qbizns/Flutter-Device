package badge_printer_zebra

import (
	"fmt"
	"sync"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	"github.com/google/gousb"
)

// usbConnection implements Connection for USB printers
type usbConnection struct {
	vendorID  uint16
	productID uint16
	logger    *telemetry.Logger

	ctx      *gousb.Context
	device   *gousb.Device
	intf     *gousb.Interface
	config   *gousb.Config
	outEp    *gousb.OutEndpoint
	inEp     *gousb.InEndpoint
	mu       sync.Mutex
}

// open opens the USB connection
func (c *usbConnection) open() error {
	// Create USB context
	c.ctx = gousb.NewContext()

	// Find matching device
	devs, err := c.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(c.vendorID) &&
			desc.Product == gousb.ID(c.productID)
	})

	if err != nil {
		c.ctx.Close()
		return fmt.Errorf("failed to enumerate USB devices: %w", err)
	}

	if len(devs) == 0 {
		c.ctx.Close()
		return fmt.Errorf("no matching USB printer found (VID:0x%04x PID:0x%04x)", c.vendorID, c.productID)
	}

	// Use first matching device
	c.device = devs[0]

	// Close other devices
	for i := 1; i < len(devs); i++ {
		devs[i].Close()
	}

	// Set auto-detach kernel driver
	c.device.SetAutoDetach(true)

	// Open default configuration
	cfg, err := c.device.Config(1)
	if err != nil {
		c.device.Close()
		c.ctx.Close()
		return fmt.Errorf("failed to set configuration: %w", err)
	}
	c.config = cfg

	// Get device descriptor
	desc := c.device.Desc
	if desc == nil {
		c.Close()
		return fmt.Errorf("failed to get device descriptor")
	}

	// Find printer interface (class 0x07)
	var printerIntf *gousb.Interface
	var printerSetting gousb.InterfaceSetting

	configDesc := desc.Configs[0]
	for _, intf := range configDesc.Interfaces {
		for _, setting := range intf.AltSettings {
			if setting.Class == gousb.ClassPrinter {
				printerIntf, err = cfg.Interface(intf.Number, setting.Alternate)
				if err != nil {
					continue
				}
				printerSetting = setting
				break
			}
		}
		if printerIntf != nil {
			break
		}
	}

	if printerIntf == nil {
		c.Close()
		return fmt.Errorf("no printer interface found")
	}
	c.intf = printerIntf

	// Find OUT endpoint (for sending data to printer)
	for _, endpoint := range printerSetting.Endpoints {
		if endpoint.Direction == gousb.EndpointDirectionOut &&
			endpoint.TransferType == gousb.TransferTypeBulk {
			c.outEp, err = printerIntf.OutEndpoint(endpoint.Number)
			if err != nil {
				continue
			}
			break
		}
	}

	if c.outEp == nil {
		c.Close()
		return fmt.Errorf("no OUT endpoint found")
	}

	// Find IN endpoint (for reading status - optional)
	for _, endpoint := range printerSetting.Endpoints {
		if endpoint.Direction == gousb.EndpointDirectionIn &&
			endpoint.TransferType == gousb.TransferTypeBulk {
			c.inEp, err = printerIntf.InEndpoint(endpoint.Number)
			if err != nil {
				continue
			}
			break
		}
	}

	c.logger.Info("USB printer connected",
		telemetry.Int("vendor_id", int(c.vendorID)),
		telemetry.Int("product_id", int(c.productID)),
	)

	return nil
}

// Write sends data to the printer
func (c *usbConnection) Write(data []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.outEp == nil {
		return 0, fmt.Errorf("not connected")
	}

	// Write in chunks if needed (USB bulk transfer limit is typically 512 bytes or 64KB)
	totalWritten := 0
	chunkSize := 4096

	for totalWritten < len(data) {
		end := totalWritten + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[totalWritten:end]
		n, err := c.outEp.Write(chunk)
		if err != nil {
			return totalWritten, fmt.Errorf("USB write error: %w", err)
		}

		totalWritten += n
	}

	return totalWritten, nil
}

// Read reads data from the printer (status responses)
func (c *usbConnection) Read(buf []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.inEp == nil {
		return 0, fmt.Errorf("printer does not support status read")
	}

	n, err := c.inEp.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("USB read error: %w", err)
	}

	return n, nil
}

// Close closes the USB connection
func (c *usbConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.intf != nil {
		c.intf.Close()
		c.intf = nil
	}

	if c.config != nil {
		c.config.Close()
		c.config = nil
	}

	if c.device != nil {
		c.device.Close()
		c.device = nil
	}

	if c.ctx != nil {
		c.ctx.Close()
		c.ctx = nil
	}

	return nil
}

// IsConnected checks if the connection is active
func (c *usbConnection) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.device != nil && c.outEp != nil
}
