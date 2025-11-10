package discovery

import (
	"context"
	"fmt"
	"net"
	"time"
)

// TCPScanner scans for TCP devices (e.g., network printers)
type TCPScanner struct {
	subnets []string
	ports   []int
}

// NewTCPScanner creates a new TCP scanner
func NewTCPScanner(subnets []string, ports []int) *TCPScanner {
	if len(ports) == 0 {
		ports = []int{9100} // Default: Raw printing port
	}

	return &TCPScanner{
		subnets: subnets,
		ports:   ports,
	}
}

// Scan scans for TCP devices
func (s *TCPScanner) Scan(ctx context.Context) ([]DiscoveredDevice, error) {
	devices := []DiscoveredDevice{}

	// For each subnet
	for _, subnet := range s.subnets {
		// For each port
		for _, port := range s.ports {
			// Try to connect
			address := fmt.Sprintf("%s:%d", subnet, port)

			// Set a short timeout
			dialer := &net.Dialer{
				Timeout: 2 * time.Second,
			}

			conn, err := dialer.DialContext(ctx, "tcp", address)
			if err != nil {
				// Device not responding
				continue
			}
			conn.Close()

			// Device found!
			device := DiscoveredDevice{
				ID:        fmt.Sprintf("tcp-%s-%d", subnet, port),
				Name:      fmt.Sprintf("Network Device at %s:%d", subnet, port),
				Kind:      "printer.escpos", // Assume ESC/POS printer on port 9100
				Transport: "tcp",
				Address:   subnet,
				Port:      port,
				Metadata: map[string]string{
					"discovered_via": "tcp_scan",
				},
			}

			devices = append(devices, device)
		}
	}

	return devices, nil
}

// Transport returns the transport name
func (s *TCPScanner) Transport() string {
	return "tcp"
}
