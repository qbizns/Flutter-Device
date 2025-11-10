package payment

import (
	"context"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
)

// PaymentProvider interface extends Device with payment processing
type PaymentProvider interface {
	devices.Device

	// StartPayment initiates a payment transaction
	StartPayment(ctx context.Context, req PaymentRequest) (PaymentID, error)

	// CancelPayment cancels an in-progress payment
	CancelPayment(ctx context.Context, id PaymentID) error

	// Status returns the current payment status
	Status(ctx context.Context, id PaymentID) (PaymentStatus, error)

	// Events streams payment events
	Events(ctx context.Context, id PaymentID) (<-chan PaymentEvent, error)
}

// PaymentID is a unique payment identifier
type PaymentID string

// PaymentRequest represents a payment initiation
type PaymentRequest struct {
	Amount            float64
	Currency          string
	MerchantReference string
	Metadata          map[string]string
	Timeout           time.Duration
}

// PaymentStatus represents payment state
type PaymentStatus struct {
	ID           PaymentID
	Status       PaymentState
	Amount       float64
	Currency     string
	CardType     string
	Last4        string
	ApprovalCode string
	Error        string
	InitiatedAt  time.Time
	CompletedAt  *time.Time
}

// PaymentState defines payment states
type PaymentState int

const (
	PaymentInitiated PaymentState = iota
	PaymentInProgress
	PaymentApproved
	PaymentDeclined
	PaymentCancelled
	PaymentTimeout
	PaymentError
)

// String returns string representation
func (p PaymentState) String() string {
	switch p {
	case PaymentInitiated:
		return "initiated"
	case PaymentInProgress:
		return "in_progress"
	case PaymentApproved:
		return "approved"
	case PaymentDeclined:
		return "declined"
	case PaymentCancelled:
		return "cancelled"
	case PaymentTimeout:
		return "timeout"
	case PaymentError:
		return "error"
	default:
		return "unknown"
	}
}

// PaymentEvent represents a payment state change
type PaymentEvent struct {
	PaymentID PaymentID
	Event     string
	Status    PaymentState
	Message   string
	Timestamp time.Time
}
