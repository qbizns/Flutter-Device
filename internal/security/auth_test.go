package security

import (
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

func TestNewAuth(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	auth := NewAuth(logger)

	if auth == nil {
		t.Fatal("NewAuth returned nil")
	}

	if auth.apiKeys == nil {
		t.Error("apiKeys map not initialized")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	key := GenerateAPIKey()

	if len(key) != 64 {
		t.Errorf("Expected key length 64, got %d", len(key))
	}

	// Generate another and ensure they're different
	key2 := GenerateAPIKey()
	if key == key2 {
		t.Error("Generated keys should be unique")
	}
}

func TestAddAPIKey(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	auth := NewAuth(logger)

	key := &APIKey{
		Key:       "test-key-123",
		ClientID:  "client-1",
		Name:      "Test Key",
		CreatedAt: time.Now(),
		Enabled:   true,
	}

	auth.AddAPIKey(key)

	// Validate key
	clientID, err := auth.ValidateAPIKey("test-key-123")
	if err != nil {
		t.Fatalf("ValidateAPIKey failed: %v", err)
	}

	if clientID != "client-1" {
		t.Errorf("Expected client_id 'client-1', got '%s'", clientID)
	}
}

func TestValidateAPIKey_Invalid(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	auth := NewAuth(logger)

	_, err := auth.ValidateAPIKey("invalid-key")
	if err == nil {
		t.Error("Expected error for invalid key")
	}
}

func TestValidateAPIKey_Disabled(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	auth := NewAuth(logger)

	key := &APIKey{
		Key:       "disabled-key",
		ClientID:  "client-1",
		Name:      "Disabled Key",
		CreatedAt: time.Now(),
		Enabled:   false,
	}

	auth.AddAPIKey(key)

	_, err := auth.ValidateAPIKey("disabled-key")
	if err == nil {
		t.Error("Expected error for disabled key")
	}
}

func TestValidateAPIKey_Expired(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	auth := NewAuth(logger)

	expiredTime := time.Now().Add(-1 * time.Hour)
	key := &APIKey{
		Key:       "expired-key",
		ClientID:  "client-1",
		Name:      "Expired Key",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: &expiredTime,
		Enabled:   true,
	}

	auth.AddAPIKey(key)

	_, err := auth.ValidateAPIKey("expired-key")
	if err == nil {
		t.Error("Expected error for expired key")
	}
}

func TestRevokeAPIKey(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	auth := NewAuth(logger)

	key := &APIKey{
		Key:       "revoke-key",
		ClientID:  "client-1",
		Name:      "Revoke Key",
		CreatedAt: time.Now(),
		Enabled:   true,
	}

	auth.AddAPIKey(key)

	// Verify it exists
	_, err := auth.ValidateAPIKey("revoke-key")
	if err != nil {
		t.Fatalf("Key should exist before revocation: %v", err)
	}

	// Revoke it
	err = auth.RevokeAPIKey("revoke-key")
	if err != nil {
		t.Fatalf("RevokeAPIKey failed: %v", err)
	}

	// Verify it's gone
	_, err = auth.ValidateAPIKey("revoke-key")
	if err == nil {
		t.Error("Key should not exist after revocation")
	}
}

// Note: UnaryInterceptor and StreamInterceptor tests require full gRPC setup
// These are tested in integration tests instead

func BenchmarkGenerateAPIKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateAPIKey()
	}
}

func BenchmarkValidateAPIKey(b *testing.B) {
	logger, _ := telemetry.NewLogger("info", "json")
	auth := NewAuth(logger)

	key := &APIKey{
		Key:       "benchmark-key",
		ClientID:  "client-1",
		Name:      "Benchmark Key",
		CreatedAt: time.Now(),
		Enabled:   true,
	}
	auth.AddAPIKey(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.ValidateAPIKey("benchmark-key")
	}
}
