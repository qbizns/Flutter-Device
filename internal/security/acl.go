package security

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// ACL manages access control lists
type ACL struct {
	rules  map[string]*ACLRule
	mu     sync.RWMutex
	logger *telemetry.Logger
}

// ACLRule defines permissions for a client
type ACLRule struct {
	ClientID          string
	AllowedDevices    []string // Device ID patterns (supports wildcards)
	AllowedOperations []string // Operation names
	DeniedDevices     []string // Explicit denies
	DeniedOperations  []string // Explicit denies
}

// NewACL creates a new ACL manager
func NewACL(logger *telemetry.Logger) *ACL {
	return &ACL{
		rules:  make(map[string]*ACLRule),
		logger: logger,
	}
}

// AddRule adds an ACL rule
func (a *ACL) AddRule(rule *ACLRule) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.rules[rule.ClientID] = rule

	a.logger.Info("ACL rule added",
		telemetry.String("client_id", rule.ClientID),
		telemetry.Int("allowed_devices", len(rule.AllowedDevices)),
		telemetry.Int("allowed_operations", len(rule.AllowedOperations)),
	)
}

// RemoveRule removes an ACL rule
func (a *ACL) RemoveRule(clientID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.rules, clientID)

	a.logger.Info("ACL rule removed",
		telemetry.String("client_id", clientID),
	)
}

// CheckAccess checks if a client has access to perform an operation on a device
func (a *ACL) CheckAccess(clientID, deviceID, operation string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	rule, exists := a.rules[clientID]
	if !exists {
		return fmt.Errorf("no ACL rule for client: %s", clientID)
	}

	// Check explicit denies first
	if a.matchesPattern(deviceID, rule.DeniedDevices) {
		return fmt.Errorf("device %s is explicitly denied for client %s", deviceID, clientID)
	}

	if a.containsString(operation, rule.DeniedOperations) {
		return fmt.Errorf("operation %s is explicitly denied for client %s", operation, clientID)
	}

	// Check allows
	if !a.matchesPattern(deviceID, rule.AllowedDevices) {
		return fmt.Errorf("device %s not allowed for client %s", deviceID, clientID)
	}

	if !a.containsString(operation, rule.AllowedOperations) {
		return fmt.Errorf("operation %s not allowed for client %s", operation, clientID)
	}

	return nil
}

// matchesPattern checks if a device ID matches any pattern in the list
func (a *ACL) matchesPattern(deviceID string, patterns []string) bool {
	for _, pattern := range patterns {
		if a.matchPattern(deviceID, pattern) {
			return true
		}
	}
	return false
}

// matchPattern checks if a string matches a pattern with wildcard support
func (a *ACL) matchPattern(s, pattern string) bool {
	// Simple wildcard matching (* means any characters)
	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(s, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(s, suffix)
	}

	return s == pattern
}

// containsString checks if a string is in a slice
func (a *ACL) containsString(s string, slice []string) bool {
	for _, item := range slice {
		if item == s || item == "*" {
			return true
		}
	}
	return false
}

// GetRule returns an ACL rule for a client
func (a *ACL) GetRule(clientID string) (*ACLRule, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	rule, exists := a.rules[clientID]
	if !exists {
		return nil, fmt.Errorf("no rule found for client: %s", clientID)
	}

	return rule, nil
}

// ListRules returns all ACL rules
func (a *ACL) ListRules() []*ACLRule {
	a.mu.RLock()
	defer a.mu.RUnlock()

	rules := make([]*ACLRule, 0, len(a.rules))
	for _, rule := range a.rules {
		rules = append(rules, rule)
	}

	return rules
}

// ACLInterceptor creates a gRPC interceptor for ACL enforcement
func (a *ACL) ACLInterceptor() func(context.Context, interface{}, interface{}, func(context.Context, interface{}) error) error {
	return func(ctx context.Context, req interface{}, info interface{}, handler func(context.Context, interface{}) error) error {
		// Extract client ID from context (set by auth interceptor)
		clientID, ok := ctx.Value("client_id").(string)
		if !ok {
			return fmt.Errorf("client_id not found in context")
		}

		// Extract device ID and operation from request
		// This would need to be customized based on your proto definitions
		// For now, we'll skip actual extraction and just log
		a.logger.Debug("ACL check",
			telemetry.String("client_id", clientID),
		)

		// Call the handler
		return handler(ctx, req)
	}
}
