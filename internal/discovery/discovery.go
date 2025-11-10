package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Discovery manages automatic device discovery
type Discovery struct {
	logger      *telemetry.Logger
	scanners    map[string]Scanner
	devices     map[string]*DiscoveredDevice
	mu          sync.RWMutex
	scanInterval time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

// DiscoveredDevice represents a discovered device
type DiscoveredDevice struct {
	ID          string
	Name        string
	Kind        string
	Transport   string
	Address     string
	Port        int
	VendorID    int
	ProductID   int
	Serial      string
	LastSeen    time.Time
	Metadata    map[string]string
}

// Scanner interface for transport-specific discovery
type Scanner interface {
	// Scan scans for devices on this transport
	Scan(ctx context.Context) ([]DiscoveredDevice, error)

	// Transport returns the transport name
	Transport() string
}

// Config holds discovery configuration
type Config struct {
	Enabled      bool
	ScanInterval time.Duration
	Transports   []string // usb, serial, tcp, mdns
}

// NewDiscovery creates a new discovery manager
func NewDiscovery(cfg Config, logger *telemetry.Logger) *Discovery {
	return &Discovery{
		logger:       logger,
		scanners:     make(map[string]Scanner),
		devices:      make(map[string]*DiscoveredDevice),
		scanInterval: cfg.ScanInterval,
	}
}

// RegisterScanner registers a transport scanner
func (d *Discovery) RegisterScanner(scanner Scanner) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.scanners[scanner.Transport()] = scanner

	d.logger.Info("discovery scanner registered",
		telemetry.String("transport", scanner.Transport()),
	)
}

// Start starts the discovery process
func (d *Discovery) Start(ctx context.Context) {
	d.ctx, d.cancel = context.WithCancel(ctx)

	d.logger.Info("discovery started",
		telemetry.Duration("scan_interval", d.scanInterval),
	)

	// Run initial scan
	d.scan()

	// Start periodic scanning
	go d.scanLoop()
}

// Stop stops the discovery process
func (d *Discovery) Stop() {
	if d.cancel != nil {
		d.cancel()
	}

	d.logger.Info("discovery stopped")
}

// GetDevices returns all discovered devices
func (d *Discovery) GetDevices() []DiscoveredDevice {
	d.mu.RLock()
	defer d.mu.RUnlock()

	devices := make([]DiscoveredDevice, 0, len(d.devices))
	for _, dev := range d.devices {
		devices = append(devices, *dev)
	}

	return devices
}

// GetDevice returns a specific discovered device
func (d *Discovery) GetDevice(id string) (*DiscoveredDevice, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	dev, exists := d.devices[id]
	if !exists {
		return nil, fmt.Errorf("device not found: %s", id)
	}

	return dev, nil
}

// scanLoop runs periodic scans
func (d *Discovery) scanLoop() {
	ticker := time.NewTicker(d.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.scan()
		}
	}
}

// scan performs a discovery scan across all registered scanners
func (d *Discovery) scan() {
	d.mu.RLock()
	scanners := make([]Scanner, 0, len(d.scanners))
	for _, scanner := range d.scanners {
		scanners = append(scanners, scanner)
	}
	d.mu.RUnlock()

	d.logger.Debug("starting discovery scan",
		telemetry.Int("scanners", len(scanners)),
	)

	// Scan all transports in parallel
	var wg sync.WaitGroup
	devicesChan := make(chan []DiscoveredDevice, len(scanners))

	for _, scanner := range scanners {
		wg.Add(1)
		go func(s Scanner) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
			defer cancel()

			devices, err := s.Scan(ctx)
			if err != nil {
				d.logger.Warn("scan failed",
					telemetry.String("transport", s.Transport()),
					telemetry.Error(err),
				)
				return
			}

			if len(devices) > 0 {
				devicesChan <- devices
			}
		}(scanner)
	}

	// Wait for all scans to complete
	go func() {
		wg.Wait()
		close(devicesChan)
	}()

	// Collect discovered devices
	now := time.Now()
	newDevices := 0

	for devices := range devicesChan {
		for _, dev := range devices {
			d.mu.Lock()

			// Update or add device
			existing, exists := d.devices[dev.ID]
			if exists {
				existing.LastSeen = now
			} else {
				dev.LastSeen = now
				d.devices[dev.ID] = &dev
				newDevices++

				d.logger.Info("device discovered",
					telemetry.String("id", dev.ID),
					telemetry.String("name", dev.Name),
					telemetry.String("kind", dev.Kind),
					telemetry.String("transport", dev.Transport),
				)
			}

			d.mu.Unlock()
		}
	}

	// Remove stale devices (not seen for 5 minutes)
	d.mu.Lock()
	staleThreshold := now.Add(-5 * time.Minute)
	for id, dev := range d.devices {
		if dev.LastSeen.Before(staleThreshold) {
			delete(d.devices, id)

			d.logger.Info("device removed (stale)",
				telemetry.String("id", id),
			)
		}
	}
	d.mu.Unlock()

	d.logger.Debug("discovery scan completed",
		telemetry.Int("new_devices", newDevices),
		telemetry.Int("total_devices", len(d.devices)),
	)
}
