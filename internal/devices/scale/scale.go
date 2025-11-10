package scale

import (
	"context"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
)

// Scale interface extends Device with weight reading
type Scale interface {
	devices.Device

	// ReadWeight reads the current weight
	ReadWeight(ctx context.Context) (WeightReading, error)

	// Zero sets the current reading as zero point
	Zero(ctx context.Context) error

	// Tare subtracts container weight
	Tare(ctx context.Context) error
}

// WeightReading represents a weight measurement
type WeightReading struct {
	Weight    float64
	Unit      WeightUnit
	Stable    bool
	Timestamp time.Time
}

// WeightUnit defines weight units
type WeightUnit string

const (
	UnitKilogram WeightUnit = "kg"
	UnitGram     WeightUnit = "g"
	UnitPound    WeightUnit = "lb"
	UnitOunce    WeightUnit = "oz"
)
