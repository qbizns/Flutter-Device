package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Auth manages authentication
type Auth struct {
	apiKeys map[string]*APIKey
	mu      sync.RWMutex
	logger  *telemetry.Logger
}

// APIKey represents an API key
type APIKey struct {
	Key       string
	ClientID  string
	Name      string
	CreatedAt time.Time
	ExpiresAt *time.Time
	Enabled   bool
}

// NewAuth creates a new auth manager
func NewAuth(logger *telemetry.Logger) *Auth {
	return &Auth{
		apiKeys: make(map[string]*APIKey),
		logger:  logger,
	}
}

// AddAPIKey adds an API key
func (a *Auth) AddAPIKey(key *APIKey) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Hash the key for storage
	hash := a.hashKey(key.Key)
	a.apiKeys[hash] = key

	a.logger.Info("API key added",
		telemetry.String("client_id", key.ClientID),
		telemetry.String("name", key.Name),
	)
}

// ValidateAPIKey validates an API key and returns the client ID
func (a *Auth) ValidateAPIKey(key string) (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	hash := a.hashKey(key)
	apiKey, exists := a.apiKeys[hash]
	if !exists {
		return "", fmt.Errorf("invalid API key")
	}

	if !apiKey.Enabled {
		return "", fmt.Errorf("API key disabled")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return "", fmt.Errorf("API key expired")
	}

	return apiKey.ClientID, nil
}

// RevokeAPIKey revokes an API key
func (a *Auth) RevokeAPIKey(key string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	hash := a.hashKey(key)
	if apiKey, exists := a.apiKeys[hash]; exists {
		apiKey.Enabled = false
		a.logger.Info("API key revoked",
			telemetry.String("client_id", apiKey.ClientID),
		)
		return nil
	}

	return fmt.Errorf("API key not found")
}

// hashKey creates a SHA-256 hash of the API key
func (a *Auth) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// UnaryInterceptor creates a gRPC unary interceptor for authentication
func (a *Auth) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for health check and ping
		if strings.HasSuffix(info.FullMethod, "/Ping") {
			return handler(ctx, req)
		}

		// Extract API key from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("missing metadata")
		}

		apiKeys := md.Get("x-api-key")
		if len(apiKeys) == 0 {
			return nil, fmt.Errorf("missing API key")
		}

		// Validate API key
		clientID, err := a.ValidateAPIKey(apiKeys[0])
		if err != nil {
			a.logger.Warn("authentication failed", telemetry.Error(err))
			return nil, fmt.Errorf("authentication failed: %w", err)
		}

		// Add client ID to context
		ctx = context.WithValue(ctx, "client_id", clientID)

		// Call handler
		return handler(ctx, req)
	}
}

// StreamInterceptor creates a gRPC stream interceptor for authentication
func (a *Auth) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Extract API key from metadata
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return fmt.Errorf("missing metadata")
		}

		apiKeys := md.Get("x-api-key")
		if len(apiKeys) == 0 {
			return fmt.Errorf("missing API key")
		}

		// Validate API key
		clientID, err := a.ValidateAPIKey(apiKeys[0])
		if err != nil {
			a.logger.Warn("authentication failed", telemetry.Error(err))
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Add client ID to context
		ctx := context.WithValue(ss.Context(), "client_id", clientID)

		// Wrap stream with new context
		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Call handler
		return handler(srv, wrapped)
	}
}

// wrappedStream wraps grpc.ServerStream with a new context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// GenerateAPIKey generates a new random API key
func GenerateAPIKey() string {
	// Generate a random key using timestamp and random bytes
	timestamp := time.Now().UnixNano()
	data := []byte(fmt.Sprintf("%d-%d", timestamp, time.Now().Unix()))
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
