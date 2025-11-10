// +build linux

package scanner_hid

import (
	"fmt"
	"time"
)

// TODO: Implement USB HID using github.com/google/gousb or github.com/karalabe/usb
//
// Implementation plan:
// 1. Initialize USB context
// 2. Enumerate USB devices
// 3. Match by vendor/product ID
// 4. Open device and claim interface
// 5. Find HID interface and interrupt endpoint
// 6. Create interrupt transfer for reading

// linuxUSBDevice implements USBDevice for Linux using gousb
type linuxUSBDevice struct {
	// device *gousb.Device
	// intf   *gousb.Interface
	// endpoint *gousb.InEndpoint
	vendorID  uint16
	productID uint16
	serial    string
}

// findUSBDevice finds and opens a USB HID scanner on Linux
func findUSBDevice(config Config) (USBDevice, error) {
	// TODO: Implement using gousb
	//
	// Example implementation:
	// ctx := gousb.NewContext()
	// defer ctx.Close()
	//
	// devices, _ := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
	//     return desc.Vendor == gousb.ID(config.VendorID) &&
	//            desc.Product == gousb.ID(config.ProductID)
	// })
	//
	// if len(devices) == 0 {
	//     return nil, fmt.Errorf("no matching USB device found")
	// }
	//
	// device := devices[0]
	// // Claim HID interface
	// // Find interrupt IN endpoint
	// // Return linuxUSBDevice

	return nil, fmt.Errorf("USB HID scanner not yet implemented on Linux (requires gousb library)")
}

// Read reads a HID report from the USB device
func (d *linuxUSBDevice) Read(buffer []byte) (int, error) {
	// TODO: Implement USB interrupt transfer read
	// return d.endpoint.Read(buffer)

	time.Sleep(100 * time.Millisecond) // Prevent busy loop
	return 0, fmt.Errorf("not implemented")
}

// Close closes the USB device
func (d *linuxUSBDevice) Close() error {
	// TODO: Release interface and close device
	// d.intf.Close()
	// d.device.Close()
	return nil
}

// GetInfo returns device information
func (d *linuxUSBDevice) GetInfo() (vendorID, productID uint16, serial string, err error) {
	return d.vendorID, d.productID, d.serial, nil
}

// EnumerateScanners lists all connected USB HID scanners
//
// This is a helper function for device discovery.
// It can be used by the auto-discovery system to find available scanners.
func EnumerateScanners() ([]ScannerInfo, error) {
	// TODO: Implement device enumeration
	//
	// Pseudo-code:
	// ctx := gousb.NewContext()
	// devices, _ := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
	//     // Check if device class is HID (0x03)
	//     return desc.Class == gousb.ClassHID
	// })
	//
	// var scanners []ScannerInfo
	// for _, dev := range devices {
	//     // Read product string, serial, etc.
	//     scanners = append(scanners, ScannerInfo{...})
	// }
	// return scanners, nil

	return nil, fmt.Errorf("enumeration not yet implemented")
}

// ScannerInfo contains information about a discovered scanner
type ScannerInfo struct {
	VendorID    uint16
	ProductID   uint16
	Vendor      string
	Product     string
	Serial      string
	BusNumber   int
	DeviceAddr  int
}

// String returns a human-readable description
func (s ScannerInfo) String() string {
	return fmt.Sprintf("%s %s (VID:0x%04x PID:0x%04x S/N:%s) at bus %d addr %d",
		s.Vendor, s.Product, s.VendorID, s.ProductID, s.Serial, s.BusNumber, s.DeviceAddr)
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

	// Honeywell
	{0x0c2e, 0x0200, "Honeywell", "Voyager 1200g"},
	{0x0c2e, 0x0720, "Honeywell", "Voyager 1400g"},
	{0x0c2e, 0x0b61, "Honeywell", "Xenon 1900"},

	// Datalogic
	{0x05f9, 0x4204, "Datalogic", "QuickScan QD2430"},
	{0x05f9, 0x4206, "Datalogic", "QuickScan QBT2430"},
	{0x05f9, 0x2206, "Datalogic", "Gryphon GD4430"},

	// Add more as needed
}
