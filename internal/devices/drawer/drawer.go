package drawer

import (
	"context"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
)

// Drawer interface extends Device with drawer control
type Drawer interface {
	devices.Device

	// Open opens the cash drawer
	Open(ctx context.Context) error

	// GetStatus returns drawer status (if supported)
	GetStatus(ctx context.Context) (*DrawerStatus, error)
}

// DrawerStatus represents drawer state
type DrawerStatus struct {
	Open       bool
	LastOpened time.Time
}
