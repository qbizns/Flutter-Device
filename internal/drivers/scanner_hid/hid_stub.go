// +build !linux,!darwin,!windows

package scanner_hid

import (
	"fmt"
)

// findUSBDevice is a stub for unsupported platforms
func findUSBDevice(config Config) (USBDevice, error) {
	return nil, fmt.Errorf("USB HID scanner not supported on this platform")
}
