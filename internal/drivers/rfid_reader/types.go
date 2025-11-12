package rfid_reader

import (
	"context"
	"sync"
	"time"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// Config holds RFID reader configuration
type Config struct {
	// Reader identification
	ID   string
	Name string
	Kind string

	// Connection settings
	Transport string // "usb", "tcp", "serial"
	Address   string // For TCP readers
	Port      int    // For TCP readers
	VendorID  uint16 // For USB readers
	ProductID uint16 // For USB readers
	SerialDev string // For serial readers (e.g., /dev/ttyUSB0)
	BaudRate  int    // For serial readers

	// Reader capabilities
	SupportedProtocols []string // "prox", "mifare", "iclass", "ntag", "desfire", "felica"
	CanRead            bool
	CanWrite           bool
	CanAuthenticate    bool
	HasBuzzer          bool
	HasLED             bool
	SupportsNDEF       bool

	// Reader settings
	Mode            pb.ReaderMode // Default mode
	BeepOnDetect    bool
	FlashOnDetect   bool
	AutoRead        bool
	ReadIntervalMs  int32
	ReadTimeoutSec  int32
	WriteTimeoutSec int32

	// Authentication settings
	DefaultMifareKeyA []byte // Default Mifare key A (6 bytes)
	DefaultMifareKeyB []byte // Default Mifare key B (6 bytes)
	DefaultIClassKey  []byte // Default iClass key (8 bytes)

	// Model/manufacturer info
	Model        string
	Manufacturer string
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Transport:      "usb",
		CanRead:        true,
		CanWrite:       false,
		CanAuthenticate: false,
		HasBuzzer:      true,
		HasLED:         true,
		SupportsNDEF:   false,
		Mode:           pb.ReaderMode_READER_MODE_CONTINUOUS,
		BeepOnDetect:   true,
		FlashOnDetect:  true,
		AutoRead:       true,
		ReadIntervalMs: 500,
		ReadTimeoutSec: 30,
		WriteTimeoutSec: 10,
		// Default Mifare key (factory default: FF FF FF FF FF FF)
		DefaultMifareKeyA: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		DefaultMifareKeyB: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}
}

// Connection interface for reader communication
type Connection interface {
	// Write sends data to the reader
	Write(data []byte) (int, error)

	// Read reads data from the reader
	Read(buf []byte) (int, error)

	// Close closes the connection
	Close() error

	// IsConnected checks if the connection is active
	IsConnected() bool

	// SetTimeout sets read/write timeout
	SetTimeout(timeout time.Duration) error
}

// CardProtocolHandler handles specific card protocols
type CardProtocolHandler interface {
	// SupportsCard checks if this handler supports the card
	SupportsCard(uid []byte, atr []byte) bool

	// ReadCard reads full card data
	ReadCard(conn Connection, uid []byte, options *pb.ReadOptions) (*pb.CardData, error)

	// WriteCard writes data to card
	WriteCard(conn Connection, uid []byte, blocks []*pb.CardBlock, credentials *pb.AuthenticationCredentials) error

	// AuthenticateCard authenticates the card
	AuthenticateCard(conn Connection, uid []byte, credentials *pb.AuthenticationCredentials) (bool, error)

	// GetCardType returns the card type
	GetCardType() pb.CardType

	// GetProtocolName returns the protocol name
	GetProtocolName() string
}

// ReadRequest represents a card read request
type ReadRequest struct {
	ID              string
	Timeout         time.Duration
	ExpectedTypes   []pb.CardType
	Options         *pb.ReadOptions
	ResultChan      chan *ReadResult
	CancelFunc      context.CancelFunc
	Timestamp       time.Time
}

// ReadResult represents the result of a card read
type ReadResult struct {
	Card           *pb.CardData
	SignalStrength int32
	ReadTime       time.Duration
	Error          error
	Timestamp      time.Time
}

// WriteRequest represents a card write request
type WriteRequest struct {
	ID              string
	CardUID         []byte
	Blocks          []*pb.CardBlock
	NDEFRecords     []*pb.NDEFRecord
	Credentials     *pb.AuthenticationCredentials
	Options         *pb.WriteOptions
	ResultChan      chan *WriteResult
	Timestamp       time.Time
}

// WriteResult represents the result of a card write
type WriteResult struct {
	Success       bool
	BlocksWritten int32
	Error         error
	Timestamp     time.Time
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	ID              string
	CardUID         []byte
	Method          pb.AuthenticationMethod
	Credentials     *pb.AuthenticationCredentials
	Challenge       []byte
	ResultChan      chan *AuthResult
	Timestamp       time.Time
}

// AuthResult represents the result of authentication
type AuthResult struct {
	Authenticated  bool
	Reason         string
	CredentialInfo *pb.CredentialInfo
	Response       []byte
	Error          error
	Timestamp      time.Time
}

// ReaderStatistics tracks reader statistics
type ReaderStatistics struct {
	TotalReads        int64
	SuccessfulReads   int64
	FailedReads       int64
	TotalWrites       int64
	SuccessfulAuths   int64
	FailedAuths       int64
	StartedAt         time.Time
	TotalReadTimeMs   int64
	mu                sync.RWMutex
}

// IncrementReads increments read counters
func (s *ReaderStatistics) IncrementReads(success bool, readTime time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalReads++
	if success {
		s.SuccessfulReads++
	} else {
		s.FailedReads++
	}
	s.TotalReadTimeMs += readTime.Milliseconds()
}

// IncrementWrites increments write counter
func (s *ReaderStatistics) IncrementWrites() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalWrites++
}

// IncrementAuths increments authentication counters
func (s *ReaderStatistics) IncrementAuths(success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if success {
		s.SuccessfulAuths++
	} else {
		s.FailedAuths++
	}
}

// GetAvgReadTime returns average read time in milliseconds
func (s *ReaderStatistics) GetAvgReadTime() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.TotalReads == 0 {
		return 0
	}
	return int32(s.TotalReadTimeMs / s.TotalReads)
}

// ToProto converts statistics to protobuf message
func (s *ReaderStatistics) ToProto() *pb.ReaderStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &pb.ReaderStatistics{
		TotalReads:       s.TotalReads,
		SuccessfulReads:  s.SuccessfulReads,
		FailedReads:      s.FailedReads,
		TotalWrites:      s.TotalWrites,
		SuccessfulAuths:  s.SuccessfulAuths,
		FailedAuths:      s.FailedAuths,
		AvgReadTimeMs:    s.GetAvgReadTime(),
	}
}

// CardPresenceTracker tracks current card presence
type CardPresenceTracker struct {
	currentUID    []byte
	currentType   pb.CardType
	presentSince  time.Time
	isPresent     bool
	mu            sync.RWMutex
}

// SetCardPresent marks a card as present
func (t *CardPresenceTracker) SetCardPresent(uid []byte, cardType pb.CardType) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.currentUID = uid
	t.currentType = cardType
	t.presentSince = time.Now()
	t.isPresent = true
}

// SetCardRemoved marks the card as removed
func (t *CardPresenceTracker) SetCardRemoved() ([]byte, time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	uid := t.currentUID
	duration := time.Since(t.presentSince)
	t.isPresent = false
	t.currentUID = nil
	t.currentType = pb.CardType_CARD_TYPE_UNKNOWN
	return uid, duration
}

// IsCardPresent checks if a card is currently present
func (t *CardPresenceTracker) IsCardPresent() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isPresent
}

// GetCurrentCard returns current card info
func (t *CardPresenceTracker) GetCurrentCard() ([]byte, pb.CardType) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentUID, t.currentType
}
