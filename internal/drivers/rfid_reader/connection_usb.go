package rfid_reader

import (
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	"github.com/google/gousb"
)

// usbConnection implements Connection for USB RFID readers
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

	// Read timeout
	readTimeout time.Duration
}

// connectUSB creates a USB connection to the RFID reader
func connectUSB(vendorID, productID uint16, logger *telemetry.Logger) (Connection, error) {
	logger.Debug("connecting to RFID reader via USB",
		telemetry.Int("vendor_id", int(vendorID)),
		telemetry.Int("product_id", int(productID)),
	)

	conn := &usbConnection{
		vendorID:    vendorID,
		productID:   productID,
		logger:      logger,
		readTimeout: 5 * time.Second,
	}

	if err := conn.open(); err != nil {
		return nil, fmt.Errorf("USB connection failed: %w", err)
	}

	return conn, nil
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
		return fmt.Errorf("no matching USB RFID reader found (VID:0x%04x PID:0x%04x)", c.vendorID, c.productID)
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

	// Find appropriate interface
	// RFID readers can be:
	// - HID class (0x03) for some readers
	// - Vendor-specific class (0xFF) for others
	// - CCID class (0x0B) for smart card readers
	var readerIntf *gousb.Interface
	var readerSetting gousb.InterfaceSetting

	configDesc := desc.Configs[0]
	for _, intf := range configDesc.Interfaces {
		for _, setting := range intf.AltSettings {
			// Accept HID, CCID, or vendor-specific interfaces
			if setting.Class == gousb.ClassHID ||
				setting.Class == gousb.ClassSmart ||
				setting.Class == gousb.ClassVendorSpec {
				readerIntf, err = cfg.Interface(intf.Number, setting.Alternate)
				if err != nil {
					continue
				}
				readerSetting = setting
				break
			}
		}
		if readerIntf != nil {
			break
		}
	}

	if readerIntf == nil {
		c.Close()
		return fmt.Errorf("no suitable interface found")
	}
	c.intf = readerIntf

	// Find endpoints
	for _, endpoint := range readerSetting.Endpoints {
		if endpoint.Direction == gousb.EndpointDirectionOut {
			if endpoint.TransferType == gousb.TransferTypeBulk ||
				endpoint.TransferType == gousb.TransferTypeInterrupt {
				c.outEp, err = readerIntf.OutEndpoint(endpoint.Number)
				if err != nil {
					continue
				}
			}
		} else if endpoint.Direction == gousb.EndpointDirectionIn {
			if endpoint.TransferType == gousb.TransferTypeBulk ||
				endpoint.TransferType == gousb.TransferTypeInterrupt {
				c.inEp, err = readerIntf.InEndpoint(endpoint.Number)
				if err != nil {
					continue
				}
			}
		}
	}

	if c.outEp == nil || c.inEp == nil {
		c.Close()
		return fmt.Errorf("required endpoints not found (OUT: %v, IN: %v)", c.outEp != nil, c.inEp != nil)
	}

	c.logger.Info("USB RFID reader connected",
		telemetry.Int("vendor_id", int(c.vendorID)),
		telemetry.Int("product_id", int(c.productID)),
	)

	return nil
}

// Write sends data to the reader
func (c *usbConnection) Write(data []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.outEp == nil {
		return 0, fmt.Errorf("not connected")
	}

	// Write in chunks if needed
	totalWritten := 0
	chunkSize := 64 // Most USB HID/interrupt endpoints use 64-byte packets

	// Determine max packet size from endpoint
	if c.outEp.Desc.MaxPacketSize > 0 {
		chunkSize = c.outEp.Desc.MaxPacketSize
	}

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

// Read reads data from the reader
func (c *usbConnection) Read(buf []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.inEp == nil {
		return 0, fmt.Errorf("not connected")
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
	return c.device != nil && c.outEp != nil && c.inEp != nil
}

// SetTimeout sets the read/write timeout
func (c *usbConnection) SetTimeout(timeout time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readTimeout = timeout
	return nil
}
