package display

import (
	"context"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
)

// Display interface extends Device with display methods
type Display interface {
	devices.Device

	// ShowLines displays text on the customer display
	ShowLines(ctx context.Context, lines []string, duration time.Duration) error

	// Clear clears the display
	Clear(ctx context.Context) error

	// SetBrightness adjusts display brightness (0-100)
	SetBrightness(ctx context.Context, percent int) error
}
