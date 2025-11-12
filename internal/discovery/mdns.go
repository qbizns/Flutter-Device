package discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	"github.com/hashicorp/mdns"
)

// MDNSScanner scans for network devices via mDNS/Bonjour
type MDNSScanner struct {
	logger   *telemetry.Logger
	services []string
	timeout  time.Duration
}

// MDNSScannerConfig configures mDNS discovery
type MDNSScannerConfig struct {
	Services []string      // Service types to discover (e.g., "_ipp._tcp", "_printer._tcp")
	Timeout  time.Duration // Discovery timeout per service
}

// NewMDNSScanner creates a new mDNS scanner
func NewMDNSScanner(cfg MDNSScannerConfig, logger *telemetry.Logger) *MDNSScanner {
	// Default services (network printers)
	services := cfg.Services
	if len(services) == 0 {
		services = []string{
			"_ipp._tcp",        // Internet Printing Protocol
			"_printer._tcp",    // Generic printer
			"_pdl-datastream._tcp", // HP printers
			"_http._tcp",       // Generic HTTP (for REST printers)
		}
	}

	// Default timeout
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	return &MDNSScanner{
		logger:   logger,
		services: services,
		timeout:  timeout,
	}
}

// Scan scans for mDNS services
func (s *MDNSScanner) Scan(ctx context.Context) ([]DiscoveredDevice, error) {
	var allDevices []DiscoveredDevice

	// Scan each service type
	for _, service := range s.services {
		select {
		case <-ctx.Done():
			return allDevices, ctx.Err()
		default:
		}

		devices, err := s.scanService(ctx, service)
		if err != nil {
			s.logger.Warn("mDNS service scan failed",
				telemetry.String("service", service),
				telemetry.Error(err),
			)
			continue
		}

		allDevices = append(allDevices, devices...)
	}

	s.logger.Debug("mDNS scan completed",
		telemetry.Int("devices", len(allDevices)),
	)

	return allDevices, nil
}

// scanService scans for a specific mDNS service type
func (s *MDNSScanner) scanService(ctx context.Context, service string) ([]DiscoveredDevice, error) {
	// Create channel for entries
	entriesCh := make(chan *mdns.ServiceEntry, 10)

	// Start mDNS query
	params := &mdns.QueryParam{
		Service: service,
		Domain:  "local",
		Timeout: s.timeout,
		Entries: entriesCh,
	}

	// Run query in goroutine
	go func() {
		if err := mdns.Query(params); err != nil {
			s.logger.Debug("mDNS query error",
				telemetry.String("service", service),
				telemetry.Error(err),
			)
		}
		close(entriesCh)
	}()

	// Collect discovered devices
	var devices []DiscoveredDevice
	seen := make(map[string]bool)

	// Read entries with timeout
	timeout := time.After(s.timeout + time.Second)

	for {
		select {
		case <-ctx.Done():
			return devices, ctx.Err()

		case <-timeout:
			return devices, nil

		case entry, ok := <-entriesCh:
			if !ok {
				// Channel closed, done
				return devices, nil
			}

			if entry == nil {
				continue
			}

			// Generate device ID
			deviceID := fmt.Sprintf("mdns-%s-%s", sanitizeName(entry.Name), entry.Host)
			if seen[deviceID] {
				continue
			}
			seen[deviceID] = true

			// Determine device kind from service type
			kind := s.serviceToKind(service)

			// Get primary address
			var address string
			var port int
			if len(entry.AddrV4) > 0 {
				address = entry.AddrV4[0].String()
				port = entry.Port
			} else if len(entry.AddrV6) > 0 {
				address = entry.AddrV6[0].String()
				port = entry.Port
			} else {
				// No IP address
				continue
			}

			// Create discovered device
			device := DiscoveredDevice{
				ID:        deviceID,
				Name:      s.formatDeviceName(entry.Name, entry.Host),
				Kind:      kind,
				Transport: "network",
				Address:   address,
				Port:      port,
				Metadata: map[string]string{
					"hostname":       entry.Host,
					"service":        service,
					"discovered_via": "mdns",
				},
			}

			// Add TXT record info
			for _, txt := range entry.Info {
				device.Metadata["txt_"+txt] = "true"
			}

			devices = append(devices, device)

			s.logger.Debug("mDNS device discovered",
				telemetry.String("name", entry.Name),
				telemetry.String("host", entry.Host),
				telemetry.String("address", address),
				telemetry.Int("port", port),
			)
		}
	}
}

// serviceToKind maps mDNS service type to device kind
func (s *MDNSScanner) serviceToKind(service string) string {
	service = strings.ToLower(service)

	switch {
	case strings.Contains(service, "ipp"):
		return "printer.ipp"
	case strings.Contains(service, "printer"):
		return "printer.network"
	case strings.Contains(service, "pdl"):
		return "printer.network"
	case strings.Contains(service, "http"):
		return "device.network"
	default:
		return "device.network"
	}
}

// formatDeviceName formats a friendly device name
func (s *MDNSScanner) formatDeviceName(name, host string) string {
	// Remove domain suffixes
	name = strings.TrimSuffix(name, ".local")
	name = strings.TrimSuffix(name, ".local.")

	// Clean up the name
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.TrimSpace(name)

	if name == "" {
		name = host
	}

	return name
}

// Transport returns the transport name
func (s *MDNSScanner) Transport() string {
	return "mdns"
}

// Helper functions

func sanitizeName(name string) string {
	// Replace special characters for use in IDs
	result := ""
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			result += string(ch)
		} else {
			result += "-"
		}
	}
	return result
}
