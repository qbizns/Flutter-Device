//go:build !linux && !darwin && !windows

package printer_usb

import (
	"fmt"
)

// findUSBPrinter is a stub for unsupported platforms
func findUSBPrinter(config Config) (USBDevice, error) {
	return nil, fmt.Errorf("USB printer not supported on this platform")
}
