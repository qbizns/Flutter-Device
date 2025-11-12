package access_control

import (
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// Simple decision engine implementation
type simpleDecisionEngine struct {
	rules  map[string]*pb.AccessRule
	rulesMu sync.RWMutex
	logger *telemetry.Logger
}

// NewSimpleDecisionEngine creates a new simple decision engine
func NewSimpleDecisionEngine(logger *telemetry.Logger) AccessDecisionEngine {
	return &simpleDecisionEngine{
		rules:  make(map[string]*pb.AccessRule),
		logger: logger,
	}
}

// CheckAccess checks if access should be granted
func (e *simpleDecisionEngine) CheckAccess(credential *pb.Credential, door string, context *pb.AccessContext) (*AccessDecisionResult, error) {
	e.logger.Debug("checking access",
		telemetry.String("door", door),
	)

	// Basic credential validation
	if credential == nil {
		return &AccessDecisionResult{
			Decision:  pb.AccessDecision_ACCESS_DECISION_DENIED,
			Reason:    "No credential provided",
			Timestamp: time.Now(),
		}, nil
	}

	// In a real implementation, this would:
	// 1. Look up credential in database
	// 2. Check if credential is valid (not expired, not revoked)
	// 3. Check access rules
	// 4. Check schedules
	// 5. Check zone transitions (anti-passback)
	// 6. Apply rule priorities

	// For now, create a simple mock implementation
	credentialInfo := e.mockCredentialLookup(credential)

	if credentialInfo == nil {
		return &AccessDecisionResult{
			Decision:  pb.AccessDecision_ACCESS_DECISION_DENIED,
			Reason:    "Invalid credential",
			Timestamp: time.Now(),
		}, nil
	}

	// Check if credential is expired
	now := time.Now()
	if credentialInfo.ValidFrom != nil && now.Before(credentialInfo.ValidFrom.AsTime()) {
		return &AccessDecisionResult{
			Decision:       pb.AccessDecision_ACCESS_DECISION_DENIED,
			Reason:         "Credential not yet valid",
			CredentialInfo: credentialInfo,
			Timestamp:      now,
		}, nil
	}

	if credentialInfo.ValidUntil != nil && now.After(credentialInfo.ValidUntil.AsTime()) {
		return &AccessDecisionResult{
			Decision:       pb.AccessDecision_ACCESS_DECISION_DENIED,
			Reason:         "Credential expired",
			CredentialInfo: credentialInfo,
			Timestamp:      now,
		}, nil
	}

	// Check if door is in allowed doors
	if len(credentialInfo.AllowedDoors) > 0 {
		allowed := false
		for _, allowedDoor := range credentialInfo.AllowedDoors {
			if allowedDoor == door || allowedDoor == "*" {
				allowed = true
				break
			}
		}
		if !allowed {
			return &AccessDecisionResult{
				Decision:       pb.AccessDecision_ACCESS_DECISION_DENIED,
				Reason:         "Door not in allowed list",
				CredentialInfo: credentialInfo,
				Timestamp:      now,
			}, nil
		}
	}

	// Check rules
	e.rulesMu.RLock()
	defer e.rulesMu.RUnlock()

	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		if e.ruleMatches(rule, credentialInfo, door, context) {
			decision := rule.Action.Decision
			if decision == pb.AccessDecision_ACCESS_DECISION_GRANTED {
				return &AccessDecisionResult{
					Decision:       pb.AccessDecision_ACCESS_DECISION_GRANTED,
					Reason:         fmt.Sprintf("Granted by rule: %s", rule.Name),
					CredentialInfo: credentialInfo,
					RuleID:         rule.RuleId,
					Timestamp:      now,
				}, nil
			} else {
				return &AccessDecisionResult{
					Decision:       pb.AccessDecision_ACCESS_DECISION_DENIED,
					Reason:         fmt.Sprintf("Denied by rule: %s", rule.Name),
					CredentialInfo: credentialInfo,
					RuleID:         rule.RuleId,
					Timestamp:      now,
				}, nil
			}
		}
	}

	// Default: grant access (in production, this would be configurable)
	return &AccessDecisionResult{
		Decision:       pb.AccessDecision_ACCESS_DECISION_GRANTED,
		Reason:         "Default access granted",
		CredentialInfo: credentialInfo,
		Timestamp:      now,
	}, nil
}

// ruleMatches checks if a rule matches the current access attempt
func (e *simpleDecisionEngine) ruleMatches(rule *pb.AccessRule, credInfo *pb.CredentialInfo, door string, context *pb.AccessContext) bool {
	if rule.Conditions == nil {
		return false
	}

	// Check user types
	if len(rule.Conditions.UserTypes) > 0 {
		match := false
		for _, userType := range rule.Conditions.UserTypes {
			if userType == credInfo.UserType {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// Check access levels
	if len(rule.Conditions.AccessLevels) > 0 {
		match := false
		for _, level := range rule.Conditions.AccessLevels {
			if level == credInfo.AccessLevel {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// Check doors
	if len(rule.Conditions.DoorIds) > 0 {
		match := false
		for _, doorID := range rule.Conditions.DoorIds {
			if doorID == door || doorID == "*" {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// Check time window
	now := time.Now()
	if rule.Conditions.ValidFrom != nil && now.Before(rule.Conditions.ValidFrom.AsTime()) {
		return false
	}
	if rule.Conditions.ValidUntil != nil && now.After(rule.Conditions.ValidUntil.AsTime()) {
		return false
	}

	return true
}

// mockCredentialLookup simulates looking up a credential
func (e *simpleDecisionEngine) mockCredentialLookup(credential *pb.Credential) *pb.CredentialInfo {
	// In a real implementation, this would query a database
	// For now, accept any credential and create mock info

	userID := "user_unknown"
	userName := "Unknown User"
	userType := pb.UserType_USER_TYPE_EMPLOYEE

	// Extract info from credential if possible
	switch credential.Type {
	case pb.CredentialType_CREDENTIAL_TYPE_CARD:
		if credential.GetCardUid() != nil {
			userID = fmt.Sprintf("card_%x", credential.GetCardUid())
			userName = fmt.Sprintf("Card User %x", credential.GetCardUid())
		}
	case pb.CredentialType_CREDENTIAL_TYPE_PIN:
		if credential.GetPin() != "" {
			userID = fmt.Sprintf("pin_%s", credential.GetPin())
			userName = fmt.Sprintf("PIN User %s", credential.GetPin())
		}
	}

	return &pb.CredentialInfo{
		UserId:       userID,
		UserName:     userName,
		UserType:     userType,
		AccessLevel:  "standard",
		AllowedDoors: []string{"*"}, // Allow all doors
	}
}

// AddRule adds an access rule
func (e *simpleDecisionEngine) AddRule(rule *pb.AccessRule) error {
	if rule.RuleId == "" {
		return fmt.Errorf("rule ID is required")
	}

	e.rulesMu.Lock()
	defer e.rulesMu.Unlock()

	e.rules[rule.RuleId] = rule
	e.logger.Info("access rule added",
		telemetry.String("rule_id", rule.RuleId),
		telemetry.String("rule_name", rule.Name),
	)

	return nil
}

// RemoveRule removes an access rule
func (e *simpleDecisionEngine) RemoveRule(ruleID string) error {
	e.rulesMu.Lock()
	defer e.rulesMu.Unlock()

	if _, exists := e.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	delete(e.rules, ruleID)
	e.logger.Info("access rule removed",
		telemetry.String("rule_id", ruleID),
	)

	return nil
}

// GetRules returns all rules
func (e *simpleDecisionEngine) GetRules() []*pb.AccessRule {
	e.rulesMu.RLock()
	defer e.rulesMu.RUnlock()

	rules := make([]*pb.AccessRule, 0, len(e.rules))
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}

	return rules
}
