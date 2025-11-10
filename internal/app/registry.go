package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Registry manages all devices
type Registry struct {
	devices map[string]devices.Device
	mu      sync.RWMutex
	logger  *telemetry.Logger
	metrics *telemetry.Metrics
}

// NewRegistry creates a new device registry
func NewRegistry(logger *telemetry.Logger, metrics *telemetry.Metrics) *Registry {
	return &Registry{
		devices: make(map[string]devices.Device),
		logger:  logger,
		metrics: metrics,
	}
}

// Register adds a device to the registry
func (r *Registry) Register(device devices.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.devices[device.ID()]; exists {
		return fmt.Errorf("device already registered: %s", device.ID())
	}

	r.devices[device.ID()] = device

	r.logger.Info("device registered",
		telemetry.String("device_id", device.ID()),
		telemetry.String("kind", device.Kind()),
		telemetry.String("name", device.Name()),
	)

	// Update metrics
	if r.metrics != nil {
		health := device.Health()
		r.metrics.SetDeviceStatus(device.ID(), device.Kind(), health.Status.ToProtoStatus())
	}

	return nil
}

// Unregister removes a device from the registry
func (r *Registry) Unregister(deviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	device, exists := r.devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	// Stop the device
	if err := device.Stop(); err != nil {
		r.logger.Error("error stopping device during unregister",
			telemetry.String("device_id", deviceID),
			telemetry.Error(err),
		)
	}

	delete(r.devices, deviceID)

	r.logger.Info("device unregistered",
		telemetry.String("device_id", deviceID),
	)

	return nil
}

// Get retrieves a device by ID
func (r *Registry) Get(deviceID string) (devices.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	device, exists := r.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	return device, nil
}

// List returns all devices
func (r *Registry) List() []devices.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]devices.Device, 0, len(r.devices))
	for _, device := range r.devices {
		list = append(list, device)
	}

	return list
}

// ListByKind returns devices of a specific kind
func (r *Registry) ListByKind(kind string) []devices.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []devices.Device
	for _, device := range r.devices {
		if device.Kind() == kind {
			list = append(list, device)
		}
	}

	return list
}

// ListByTag returns devices with a specific tag
func (r *Registry) ListByTag(tag string) []devices.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []devices.Device
	for _, device := range r.devices {
		metadata := device.Metadata()
		if tags, exists := metadata["tags"]; exists {
			// Simple tag matching (could be improved)
			if tags == tag {
				list = append(list, device)
			}
		}
	}

	return list
}

// StartAll starts all registered devices
func (r *Registry) StartAll(ctx context.Context) error {
	r.mu.RLock()
	devices := make([]devices.Device, 0, len(r.devices))
	for _, device := range r.devices {
		devices = append(devices, device)
	}
	r.mu.RUnlock()

	r.logger.Info("starting all devices", telemetry.Int("count", len(devices)))

	for _, device := range devices {
		if err := device.Start(ctx); err != nil {
			r.logger.Error("failed to start device",
				telemetry.String("device_id", device.ID()),
				telemetry.String("kind", device.Kind()),
				telemetry.Error(err),
			)
			// Continue starting other devices
		} else {
			r.logger.Info("device started",
				telemetry.String("device_id", device.ID()),
				telemetry.String("kind", device.Kind()),
			)
		}
	}

	return nil
}

// StopAll stops all registered devices
func (r *Registry) StopAll() error {
	r.mu.RLock()
	devices := make([]devices.Device, 0, len(r.devices))
	for _, device := range r.devices {
		devices = append(devices, device)
	}
	r.mu.RUnlock()

	r.logger.Info("stopping all devices", telemetry.Int("count", len(devices)))

	for _, device := range devices {
		if err := device.Stop(); err != nil {
			r.logger.Error("failed to stop device",
				telemetry.String("device_id", device.ID()),
				telemetry.Error(err),
			)
		}
	}

	return nil
}

// MonitorHealth periodically checks device health
func (r *Registry) MonitorHealth(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	r.logger.Info("starting health monitor", telemetry.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("health monitor stopped")
			return
		case <-ticker.C:
			r.checkHealth()
		}
	}
}

// checkHealth checks health of all devices
func (r *Registry) checkHealth() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, device := range r.devices {
		health := device.Health()

		// Update metrics
		if r.metrics != nil {
			r.metrics.SetDeviceStatus(device.ID(), device.Kind(), health.Status.ToProtoStatus())
		}

		// Log if degraded or offline
		if health.Status == devices.HealthDegraded || health.Status == devices.HealthOffline || health.Status == devices.HealthError {
			r.logger.Warn("device health issue",
				telemetry.String("device_id", device.ID()),
				telemetry.String("status", health.Status.String()),
				telemetry.String("message", health.Message),
			)
		}
	}
}

// Count returns the number of registered devices
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.devices)
}

// Exists checks if a device exists
func (r *Registry) Exists(deviceID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.devices[deviceID]
	return exists
}
