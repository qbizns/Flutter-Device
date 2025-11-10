package security

import (
	"testing"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

func TestNewACL(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	if acl == nil {
		t.Fatal("NewACL returned nil")
	}

	if acl.rules == nil {
		t.Error("rules map not initialized")
	}
}

func TestAddRule(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"printer-*"},
		AllowedOperations: []string{"print"},
	}

	acl.AddRule(rule)

	// Verify rule was added
	retrievedRule, err := acl.GetRule("client-1")
	if err != nil {
		t.Fatalf("GetRule failed: %v", err)
	}

	if retrievedRule.ClientID != "client-1" {
		t.Errorf("Expected client_id 'client-1', got '%s'", retrievedRule.ClientID)
	}
}

func TestCheckAccess_Allowed(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"printer-*"},
		AllowedOperations: []string{"print", "status"},
	}
	acl.AddRule(rule)

	// Should allow
	err := acl.CheckAccess("client-1", "printer-1", "print")
	if err != nil {
		t.Errorf("Access should be allowed: %v", err)
	}

	err = acl.CheckAccess("client-1", "printer-kitchen", "status")
	if err != nil {
		t.Errorf("Access should be allowed: %v", err)
	}
}

func TestCheckAccess_DeviceNotAllowed(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"printer-*"},
		AllowedOperations: []string{"print"},
	}
	acl.AddRule(rule)

	// Should deny - wrong device
	err := acl.CheckAccess("client-1", "scanner-1", "print")
	if err == nil {
		t.Error("Access should be denied for non-printer device")
	}
}

func TestCheckAccess_OperationNotAllowed(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"printer-*"},
		AllowedOperations: []string{"print"},
	}
	acl.AddRule(rule)

	// Should deny - wrong operation
	err := acl.CheckAccess("client-1", "printer-1", "delete")
	if err == nil {
		t.Error("Access should be denied for non-allowed operation")
	}
}

func TestCheckAccess_ExplicitDeny(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"*"}, // Allow all
		AllowedOperations: []string{"*"}, // Allow all
		DeniedDevices:     []string{"printer-secret"},
	}
	acl.AddRule(rule)

	// Should allow normal printer
	err := acl.CheckAccess("client-1", "printer-1", "print")
	if err != nil {
		t.Errorf("Access should be allowed: %v", err)
	}

	// Should deny explicitly denied printer
	err = acl.CheckAccess("client-1", "printer-secret", "print")
	if err == nil {
		t.Error("Access should be denied for explicitly denied device")
	}
}

func TestCheckAccess_WildcardOperations(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"printer-*"},
		AllowedOperations: []string{"*"}, // All operations
	}
	acl.AddRule(rule)

	// Should allow any operation
	operations := []string{"print", "status", "calibrate", "anything"}
	for _, op := range operations {
		err := acl.CheckAccess("client-1", "printer-1", op)
		if err != nil {
			t.Errorf("Operation '%s' should be allowed: %v", op, err)
		}
	}
}

func TestCheckAccess_WildcardDevices(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"*"}, // All devices
		AllowedOperations: []string{"status"},
	}
	acl.AddRule(rule)

	// Should allow any device
	devices := []string{"printer-1", "scanner-1", "scale-1", "anything"}
	for _, device := range devices {
		err := acl.CheckAccess("client-1", device, "status")
		if err != nil {
			t.Errorf("Device '%s' should be allowed: %v", device, err)
		}
	}
}

func TestCheckAccess_NoRule(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	// No rules added
	err := acl.CheckAccess("unknown-client", "printer-1", "print")
	if err == nil {
		t.Error("Access should be denied when no rule exists")
	}
}

func TestRemoveRule(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"*"},
		AllowedOperations: []string{"*"},
	}
	acl.AddRule(rule)

	// Verify it exists
	_, err := acl.GetRule("client-1")
	if err != nil {
		t.Fatalf("Rule should exist: %v", err)
	}

	// Remove it
	acl.RemoveRule("client-1")

	// Verify it's gone
	_, err = acl.GetRule("client-1")
	if err == nil {
		t.Error("Rule should not exist after removal")
	}
}

func TestListRules(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	// Add multiple rules
	for i := 1; i <= 3; i++ {
		rule := &ACLRule{
			ClientID:          "client-" + string(rune('0'+i)),
			AllowedDevices:    []string{"*"},
			AllowedOperations: []string{"*"},
		}
		acl.AddRule(rule)
	}

	rules := acl.ListRules()
	if len(rules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(rules))
	}
}

func TestMatchPattern_Exact(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"printer-1"},
		AllowedOperations: []string{"print"},
	}
	acl.AddRule(rule)

	// Exact match should work
	err := acl.CheckAccess("client-1", "printer-1", "print")
	if err != nil {
		t.Errorf("Exact match should work: %v", err)
	}

	// Different device should fail
	err = acl.CheckAccess("client-1", "printer-2", "print")
	if err == nil {
		t.Error("Different device should be denied")
	}
}

func TestMatchPattern_PrefixWildcard(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"printer-*"},
		AllowedOperations: []string{"print"},
	}
	acl.AddRule(rule)

	// Should match
	devices := []string{"printer-1", "printer-kitchen", "printer-abc"}
	for _, device := range devices {
		err := acl.CheckAccess("client-1", device, "print")
		if err != nil {
			t.Errorf("Device '%s' should match pattern 'printer-*': %v", device, err)
		}
	}

	// Should not match
	err := acl.CheckAccess("client-1", "scanner-1", "print")
	if err == nil {
		t.Error("'scanner-1' should not match pattern 'printer-*'")
	}
}

func TestMatchPattern_SuffixWildcard(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"*-1"},
		AllowedOperations: []string{"print"},
	}
	acl.AddRule(rule)

	// Should match
	devices := []string{"printer-1", "scanner-1", "scale-1"}
	for _, device := range devices {
		err := acl.CheckAccess("client-1", device, "print")
		if err != nil {
			t.Errorf("Device '%s' should match pattern '*-1': %v", device, err)
		}
	}

	// Should not match
	err := acl.CheckAccess("client-1", "printer-2", "print")
	if err == nil {
		t.Error("'printer-2' should not match pattern '*-1'")
	}
}

func BenchmarkCheckAccess(b *testing.B) {
	logger, _ := telemetry.NewLogger("info", "json")
	acl := NewACL(logger)

	rule := &ACLRule{
		ClientID:          "client-1",
		AllowedDevices:    []string{"printer-*", "scanner-*"},
		AllowedOperations: []string{"print", "scan", "status"},
	}
	acl.AddRule(rule)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acl.CheckAccess("client-1", "printer-1", "print")
	}
}
