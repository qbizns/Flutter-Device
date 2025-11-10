package virtual_devices

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/scale"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// VirtualScale is a virtual weighing scale for testing
type VirtualScale struct {
	*devices.BaseDevice
	logger       *telemetry.Logger
	mu           sync.RWMutex
	currentWeight float64
	zeroOffset    float64
	tareWeight    float64
	unit          scale.WeightUnit
	stable        bool
	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewVirtualScale creates a new virtual scale
func NewVirtualScale(id, name string, logger *telemetry.Logger) *VirtualScale {
	return &VirtualScale{
		BaseDevice: devices.NewBaseDevice(id, "scale.virtual", name, map[string]string{
			"virtual": "true",
			"type":    "scale",
		}),
		logger:        logger,
		currentWeight: 0.0,
		zeroOffset:    0.0,
		tareWeight:    0.0,
		unit:          scale.UnitKilogram,
		stable:        true,
	}
}

// Start initializes the virtual scale
func (s *VirtualScale) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("virtual scale already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	s.SetHealth(devices.HealthReady, "virtual scale ready", nil)

	s.logger.Info("virtual scale started",
		telemetry.String("device_id", s.ID()),
		telemetry.String("name", s.Name()),
	)

	// Start weight simulator
	go s.weightSimulator()

	return nil
}

// Stop stops the virtual scale
func (s *VirtualScale) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.running = false

	s.SetHealth(devices.HealthOffline, "virtual scale stopped", nil)

	s.logger.Info("virtual scale stopped",
		telemetry.String("device_id", s.ID()),
	)

	return nil
}

// ReadWeight reads the current weight
func (s *VirtualScale) ReadWeight(ctx context.Context) (scale.WeightReading, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return scale.WeightReading{}, fmt.Errorf("scale not running")
	}

	// Calculate displayed weight (current - zero - tare)
	displayWeight := s.currentWeight - s.zeroOffset - s.tareWeight

	// Clamp to zero if negative
	if displayWeight < 0 {
		displayWeight = 0
	}

	reading := scale.WeightReading{
		Weight:    displayWeight,
		Unit:      s.unit,
		Stable:    s.stable,
		Timestamp: time.Now(),
	}

	s.logger.Info("weight read",
		telemetry.String("device_id", s.ID()),
		telemetry.Float64("weight", reading.Weight),
		telemetry.String("unit", string(reading.Unit)),
		telemetry.Bool("stable", reading.Stable),
	)

	fmt.Printf("\n")
	fmt.Printf("===== VIRTUAL SCALE READING =====\n")
	fmt.Printf("Device:  %s\n", s.Name())
	fmt.Printf("Weight:  %.3f %s\n", reading.Weight, reading.Unit)
	fmt.Printf("Stable:  %v\n", reading.Stable)
	fmt.Printf("Time:    %s\n", reading.Timestamp.Format(time.RFC3339))
	fmt.Printf("=================================\n")
	fmt.Printf("\n")

	return reading, nil
}

// Zero sets the current reading as zero point
func (s *VirtualScale) Zero(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scale not running")
	}

	s.zeroOffset = s.currentWeight
	s.tareWeight = 0 // Reset tare when zeroing

	s.logger.Info("scale zeroed",
		telemetry.String("device_id", s.ID()),
		telemetry.Float64("zero_offset", s.zeroOffset),
	)

	fmt.Printf("\n")
	fmt.Printf("===== VIRTUAL SCALE ZEROED =====\n")
	fmt.Printf("Device:       %s\n", s.Name())
	fmt.Printf("Zero Offset:  %.3f %s\n", s.zeroOffset, s.unit)
	fmt.Printf("================================\n")
	fmt.Printf("\n")

	return nil
}

// Tare subtracts container weight
func (s *VirtualScale) Tare(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scale not running")
	}

	// Set tare to current displayed weight
	s.tareWeight = s.currentWeight - s.zeroOffset

	s.logger.Info("scale tared",
		telemetry.String("device_id", s.ID()),
		telemetry.Float64("tare_weight", s.tareWeight),
	)

	fmt.Printf("\n")
	fmt.Printf("===== VIRTUAL SCALE TARED =====\n")
	fmt.Printf("Device:      %s\n", s.Name())
	fmt.Printf("Tare Weight: %.3f %s\n", s.tareWeight, s.unit)
	fmt.Printf("===============================\n")
	fmt.Printf("\n")

	return nil
}

// SetWeight manually sets the weight (for testing)
func (s *VirtualScale) SetWeight(weight float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentWeight = weight
	s.stable = true

	s.logger.Info("weight set manually",
		telemetry.String("device_id", s.ID()),
		telemetry.Float64("weight", weight),
	)
}

// SetUnit changes the display unit
func (s *VirtualScale) SetUnit(unit scale.WeightUnit) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldUnit := s.unit
	s.unit = unit

	// Convert current weight to new unit
	switch {
	case oldUnit == scale.UnitKilogram && unit == scale.UnitGram:
		s.currentWeight *= 1000
		s.zeroOffset *= 1000
		s.tareWeight *= 1000
	case oldUnit == scale.UnitGram && unit == scale.UnitKilogram:
		s.currentWeight /= 1000
		s.zeroOffset /= 1000
		s.tareWeight /= 1000
	case oldUnit == scale.UnitKilogram && unit == scale.UnitPound:
		s.currentWeight *= 2.20462
		s.zeroOffset *= 2.20462
		s.tareWeight *= 2.20462
	case oldUnit == scale.UnitPound && unit == scale.UnitKilogram:
		s.currentWeight /= 2.20462
		s.zeroOffset /= 2.20462
		s.tareWeight /= 2.20462
	}

	s.logger.Info("unit changed",
		telemetry.String("device_id", s.ID()),
		telemetry.String("old_unit", string(oldUnit)),
		telemetry.String("new_unit", string(unit)),
	)
}

// weightSimulator simulates weight changes over time
func (s *VirtualScale) weightSimulator() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Predefined test weights (in kg)
	testWeights := []float64{
		0.0,    // Empty
		0.250,  // 250g
		0.500,  // 500g
		1.234,  // 1.234kg
		2.500,  // 2.5kg
		5.678,  // 5.678kg
		10.000, // 10kg
	}

	index := 0

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()

			// Cycle through test weights
			s.currentWeight = testWeights[index]
			index = (index + 1) % len(testWeights)

			// Simulate instability (10% chance)
			s.stable = rand.Float64() > 0.1

			s.mu.Unlock()

			s.logger.Debug("weight simulated",
				telemetry.String("device_id", s.ID()),
				telemetry.Float64("weight", s.currentWeight),
				telemetry.Bool("stable", s.stable),
			)
		}
	}
}

// SimulateWeighing simulates a weighing scenario
func (s *VirtualScale) SimulateWeighing(weights []float64, delay time.Duration) {
	for _, weight := range weights {
		s.SetWeight(weight)
		if delay > 0 {
			time.Sleep(delay)
		}
	}
}
