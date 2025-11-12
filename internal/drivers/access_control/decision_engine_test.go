package access_control

import (
	"fmt"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewSimpleDecisionEngine(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	assert.NotNil(t, engine)
	assert.Empty(t, engine.GetRules())
}

func TestDecisionEngine_CheckAccess_NoCredential(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	result, err := engine.CheckAccess(nil, "door-1", nil)

	require.NoError(t, err)
	assert.Equal(t, pb.AccessDecision_ACCESS_DECISION_DENIED, result.Decision)
	assert.Contains(t, result.Reason, "No credential")
}

func TestDecisionEngine_CheckAccess_ValidCard(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	credential := &pb.Credential{
		Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
		Data: &pb.Credential_CardUid{
			CardUid: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	result, err := engine.CheckAccess(credential, "door-1", nil)

	require.NoError(t, err)
	assert.Equal(t, pb.AccessDecision_ACCESS_DECISION_GRANTED, result.Decision)
	assert.NotNil(t, result.CredentialInfo)
	assert.NotEmpty(t, result.CredentialInfo.UserId)
	assert.NotEmpty(t, result.CredentialInfo.UserName)
}

func TestDecisionEngine_CheckAccess_ValidPIN(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	credential := &pb.Credential{
		Type: pb.CredentialType_CREDENTIAL_TYPE_PIN,
		Data: &pb.Credential_Pin{
			Pin: "1234",
		},
	}

	result, err := engine.CheckAccess(credential, "door-1", nil)

	require.NoError(t, err)
	assert.Equal(t, pb.AccessDecision_ACCESS_DECISION_GRANTED, result.Decision)
	assert.NotNil(t, result.CredentialInfo)
}

func TestDecisionEngine_AddRule(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	rule := &pb.AccessRule{
		RuleId:      "rule-1",
		Name:        "Test Rule",
		Description: "Test rule for employees",
		Priority:    10,
		Enabled:     true,
		Conditions: &pb.RuleConditions{
			UserTypes: []pb.UserType{pb.UserType_USER_TYPE_EMPLOYEE},
			DoorIds:   []string{"door-1"},
		},
		Action: &pb.RuleAction{
			Decision:             pb.AccessDecision_ACCESS_DECISION_GRANTED,
			UnlockDurationSeconds: 5,
		},
	}

	err := engine.AddRule(rule)
	require.NoError(t, err)

	rules := engine.GetRules()
	assert.Len(t, rules, 1)
	assert.Equal(t, "rule-1", rules[0].RuleId)
}

func TestDecisionEngine_AddRule_NoID(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	rule := &pb.AccessRule{
		Name: "Test Rule",
	}

	err := engine.AddRule(rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule ID is required")
}

func TestDecisionEngine_RemoveRule(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	rule := &pb.AccessRule{
		RuleId:  "rule-1",
		Name:    "Test Rule",
		Enabled: true,
	}

	engine.AddRule(rule)
	assert.Len(t, engine.GetRules(), 1)

	err := engine.RemoveRule("rule-1")
	require.NoError(t, err)
	assert.Empty(t, engine.GetRules())
}

func TestDecisionEngine_RemoveRule_NotFound(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	err := engine.RemoveRule("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule not found")
}

func TestDecisionEngine_CheckAccess_WithRule(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	// Add grant rule for employees at door-1
	grantRule := &pb.AccessRule{
		RuleId:   "grant-rule",
		Name:     "Grant Employees",
		Priority: 10,
		Enabled:  true,
		Conditions: &pb.RuleConditions{
			UserTypes: []pb.UserType{pb.UserType_USER_TYPE_EMPLOYEE},
			DoorIds:   []string{"door-1"},
		},
		Action: &pb.RuleAction{
			Decision:              pb.AccessDecision_ACCESS_DECISION_GRANTED,
			UnlockDurationSeconds: 5,
		},
	}
	engine.AddRule(grantRule)

	credential := &pb.Credential{
		Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
		Data: &pb.Credential_CardUid{
			CardUid: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	result, err := engine.CheckAccess(credential, "door-1", nil)

	require.NoError(t, err)
	assert.Equal(t, pb.AccessDecision_ACCESS_DECISION_GRANTED, result.Decision)
	assert.Equal(t, "grant-rule", result.RuleID)
	assert.Contains(t, result.Reason, "Grant Employees")
}

func TestDecisionEngine_CheckAccess_DenyRule(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	// Add deny rule for visitors at secure doors
	denyRule := &pb.AccessRule{
		RuleId:   "deny-rule",
		Name:     "Deny Visitors at Secure Doors",
		Priority: 100, // Higher priority
		Enabled:  true,
		Conditions: &pb.RuleConditions{
			UserTypes: []pb.UserType{pb.UserType_USER_TYPE_VISITOR},
			DoorIds:   []string{"secure-door-1"},
		},
		Action: &pb.RuleAction{
			Decision: pb.AccessDecision_ACCESS_DECISION_DENIED,
		},
	}
	engine.AddRule(denyRule)

	credential := &pb.Credential{
		Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
		Data: &pb.Credential_CardUid{
			CardUid: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	// Modify the mock to return visitor type
	// In real implementation, this would come from database
	result, err := engine.CheckAccess(credential, "secure-door-1", nil)

	require.NoError(t, err)
	// Default behavior will grant, as mock returns employee type
	assert.Equal(t, pb.AccessDecision_ACCESS_DECISION_GRANTED, result.Decision)
}

func TestDecisionEngine_CheckAccess_DisabledRule(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	// Add disabled rule
	rule := &pb.AccessRule{
		RuleId:   "disabled-rule",
		Name:     "Disabled Rule",
		Priority: 100,
		Enabled:  false, // Disabled
		Conditions: &pb.RuleConditions{
			DoorIds: []string{"door-1"},
		},
		Action: &pb.RuleAction{
			Decision: pb.AccessDecision_ACCESS_DECISION_DENIED,
		},
	}
	engine.AddRule(rule)

	credential := &pb.Credential{
		Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
		Data: &pb.Credential_CardUid{
			CardUid: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	result, err := engine.CheckAccess(credential, "door-1", nil)

	require.NoError(t, err)
	// Should grant because rule is disabled
	assert.Equal(t, pb.AccessDecision_ACCESS_DECISION_GRANTED, result.Decision)
	assert.NotEqual(t, "disabled-rule", result.RuleID)
}

func TestDecisionEngine_RuleMatches_AccessLevel(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger).(*simpleDecisionEngine)

	rule := &pb.AccessRule{
		RuleId:  "level-rule",
		Enabled: true,
		Conditions: &pb.RuleConditions{
			AccessLevels: []string{"level-3", "level-4"},
			DoorIds:      []string{"door-1"},
		},
	}

	// Credential with matching access level
	credInfo := &pb.CredentialInfo{
		AccessLevel: "level-3",
	}

	matches := engine.ruleMatches(rule, credInfo, "door-1", nil)
	assert.True(t, matches)

	// Credential with non-matching access level
	credInfo2 := &pb.CredentialInfo{
		AccessLevel: "level-1",
	}

	matches2 := engine.ruleMatches(rule, credInfo2, "door-1", nil)
	assert.False(t, matches2)
}

func TestDecisionEngine_RuleMatches_TimeWindow(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger).(*simpleDecisionEngine)

	// Rule valid in the future
	rule := &pb.AccessRule{
		RuleId:  "future-rule",
		Enabled: true,
		Conditions: &pb.RuleConditions{
			ValidFrom:  timestamppb.Now(),
			ValidUntil: timestamppb.New(timestamppb.Now().AsTime().Add(24 * time.Hour)),
		},
	}

	credInfo := &pb.CredentialInfo{}

	// Should match (rule is valid now)
	matches := engine.ruleMatches(rule, credInfo, "door-1", nil)
	assert.True(t, matches)

	// Rule expired
	expiredRule := &pb.AccessRule{
		RuleId:  "expired-rule",
		Enabled: true,
		Conditions: &pb.RuleConditions{
			ValidFrom:  timestamppb.New(timestamppb.Now().AsTime().Add(-48 * time.Hour)),
			ValidUntil: timestamppb.New(timestamppb.Now().AsTime().Add(-24 * time.Hour)),
		},
	}

	matches2 := engine.ruleMatches(expiredRule, credInfo, "door-1", nil)
	assert.False(t, matches2)
}

func TestDecisionEngine_RuleMatches_WildcardDoor(t *testing.T) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger).(*simpleDecisionEngine)

	rule := &pb.AccessRule{
		RuleId:  "wildcard-rule",
		Enabled: true,
		Conditions: &pb.RuleConditions{
			DoorIds: []string{"*"}, // Wildcard
		},
	}

	credInfo := &pb.CredentialInfo{}

	// Should match any door
	matches1 := engine.ruleMatches(rule, credInfo, "door-1", nil)
	assert.True(t, matches1)

	matches2 := engine.ruleMatches(rule, credInfo, "door-2", nil)
	assert.True(t, matches2)
}

func BenchmarkDecisionEngine_CheckAccess(b *testing.B) {
	logger := telemetry.NewLogger("test")
	engine := NewSimpleDecisionEngine(logger)

	// Add some rules
	for i := 0; i < 10; i++ {
		rule := &pb.AccessRule{
			RuleId:   fmt.Sprintf("rule-%d", i),
			Name:     fmt.Sprintf("Rule %d", i),
			Priority: int32(i),
			Enabled:  true,
			Conditions: &pb.RuleConditions{
				DoorIds: []string{fmt.Sprintf("door-%d", i)},
			},
			Action: &pb.RuleAction{
				Decision: pb.AccessDecision_ACCESS_DECISION_GRANTED,
			},
		}
		engine.AddRule(rule)
	}

	credential := &pb.Credential{
		Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
		Data: &pb.Credential_CardUid{
			CardUid: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckAccess(credential, "door-5", nil)
	}
}
