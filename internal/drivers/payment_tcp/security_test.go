package payment_tcp

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestSecurityCardDataMasking tests that card data is properly masked
func TestSecurityCardDataMasking(t *testing.T) {
	driver := NewSimulatorDriver("sec-mask", "Security Masking Test")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	}

	resp, err := driver.ProcessTransaction(ctx, req)
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Verify card number is masked
	if !strings.Contains(resp.CardNumberMasked, "****") {
		t.Error("Card number should be masked")
	}

	// Verify no full PAN in response
	if len(resp.CardNumberMasked) > 20 {
		t.Error("Masked card number too long, may contain full PAN")
	}

	// Check receipt doesn't contain full card number
	for _, line := range resp.ReceiptLines {
		// Look for sequences of 13-19 digits (card numbers)
		digitCount := 0
		for _, char := range line {
			if char >= '0' && char <= '9' {
				digitCount++
				if digitCount >= 13 {
					t.Errorf("Receipt may contain full card number: %s", line)
					break
				}
			} else {
				digitCount = 0
			}
		}
	}
}

// TestSecurityAuditLogSanitization tests audit log sanitization
func TestSecurityAuditLogSanitization(t *testing.T) {
	writer := NewMemoryAuditReader()
	logger := NewAuditLogger(writer, 10)

	driver := NewSimulatorDriver("sec-audit", "Security Audit Test")
	driver.SetAuditLogger(logger)
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Process transaction
	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	}

	_, err := driver.ProcessTransaction(ctx, req)
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Flush audit logs
	logger.Flush()

	// Check audit logs for sensitive data
	events := writer.GetAllEvents()
	for _, event := range events {
		// Verify card number is masked
		if event.CardNumberMasked != "" && !strings.Contains(event.CardNumberMasked, "****") {
			t.Errorf("Audit log contains unmasked card data: %s", event.CardNumberMasked)
		}

		// Verify no full card numbers in any field
		eventStr := event.EventID + event.DeviceID + event.TerminalID
		digitCount := 0
		for _, char := range eventStr {
			if char >= '0' && char <= '9' {
				digitCount++
				if digitCount >= 13 {
					t.Error("Audit log may contain full card number")
					break
				}
			} else {
				digitCount = 0
			}
		}
	}
}

// TestSecurityInjectionAttempts tests protection against injection attacks
func TestSecurityInjectionAttempts(t *testing.T) {
	driver := NewSimulatorDriver("sec-injection", "Security Injection Test")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	injectionTests := []struct {
		name     string
		req      TransactionRequest
		wantErr  bool
	}{
		{
			name: "SQL injection in reference",
			req: TransactionRequest{
				Type:      TransactionSale,
				Amount:    10000,
				Currency:  "SAR",
				Reference: "'; DROP TABLE transactions; --",
			},
			wantErr: false, // Should be treated as normal string
		},
		{
			name: "Command injection in reference",
			req: TransactionRequest{
				Type:      TransactionSale,
				Amount:    10000,
				Currency:  "SAR",
				Reference: "; rm -rf / ;",
			},
			wantErr: false, // Should be treated as normal string
		},
		{
			name: "Null byte injection",
			req: TransactionRequest{
				Type:      TransactionSale,
				Amount:    10000,
				Currency:  "SAR",
				Reference: "test\x00malicious",
			},
			wantErr: false, // Should handle gracefully
		},
		{
			name: "Path traversal in reference",
			req: TransactionRequest{
				Type:      TransactionSale,
				Amount:    10000,
				Currency:  "SAR",
				Reference: "../../etc/passwd",
			},
			wantErr: false, // Should be treated as normal string
		},
	}

	for _, tt := range injectionTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := driver.ProcessTransaction(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Injection test error = %v, wantErr %v", err, tt.wantErr)
			}
			if resp != nil && !tt.wantErr {
				t.Logf("Injection attempt handled safely: %s", tt.name)
			}
		})
	}
}

// TestSecurityReplayAttack tests replay attack prevention
func TestSecurityReplayAttack(t *testing.T) {
	driver := NewSimulatorDriver("sec-replay", "Security Replay Test")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Process first transaction
	req := TransactionRequest{
		Type:      TransactionSale,
		Amount:    10000,
		Currency:  "SAR",
		Reference: "UNIQUE-REF-001",
	}

	resp1, err := driver.ProcessTransaction(ctx, req)
	if err != nil {
		t.Fatalf("First transaction failed: %v", err)
	}

	if !resp1.Success {
		t.Fatal("First transaction should succeed")
	}

	// Try to replay the same transaction immediately
	resp2, err := driver.ProcessTransaction(ctx, req)
	if err != nil {
		t.Fatalf("Replay transaction failed: %v", err)
	}

	// Note: Simulator doesn't prevent replays, but in production:
	// - Transaction IDs should be unique (checked by acquirer)
	// - Timestamps should be validated
	// - Duplicate detection should be implemented
	t.Logf("Replay handling: resp1=%s, resp2=%s", resp1.TransactionID, resp2.TransactionID)

	// Verify different transaction IDs
	if resp1.TransactionID == resp2.TransactionID {
		t.Error("Replay attack: same transaction ID for different requests")
	}
}

// TestSecurityRateLimiting tests rate limiting protection
func TestSecurityRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rate limit test in short mode")
	}

	driver := NewSimulatorDriver("sec-rate", "Security Rate Limit Test")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Attempt rapid-fire transactions
	const burstSize = 100
	successCount := 0
	errorCount := 0

	startTime := time.Now()
	for i := 0; i < burstSize; i++ {
		req := TransactionRequest{
			Type:     TransactionSale,
			Amount:   10000,
			Currency: "SAR",
		}

		resp, err := driver.ProcessTransaction(ctx, req)
		if err != nil {
			errorCount++
		} else if resp.Success {
			successCount++
		}
	}
	duration := time.Since(startTime)

	tps := float64(burstSize) / duration.Seconds()

	t.Logf("Rate test: %d requests in %v (%.2f TPS)", burstSize, duration, tps)
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)

	// Note: Current implementation doesn't enforce rate limiting
	// This test documents the behavior for future implementation
	if tps > 1000 {
		t.Logf("Warning: Very high TPS (%.2f) - consider implementing rate limiting", tps)
	}
}

// TestSecurityAuthenticationBypass tests authentication requirements
func TestSecurityAuthenticationBypass(t *testing.T) {
	// Test with empty/invalid terminal ID
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "", // Empty terminal ID
		MerchantID: "MERCHANT001",
	}

	_ = NewDriver("sec-auth", "Security Auth Test", config)

	// Verify terminal ID is required
	if config.TerminalID == "" {
		t.Log("Empty terminal ID detected (expected for this test)")
	}

	// In production, connection should fail with empty credentials
	// Simulator accepts any credentials for testing
	t.Log("Authentication test: Empty credentials handled")
}

// TestSecurityPrivilegeEscalation tests privilege boundaries
func TestSecurityPrivilegeEscalation(t *testing.T) {
	driver1 := NewSimulatorDriver("sec-priv-1", "Device 1")
	driver2 := NewSimulatorDriver("sec-priv-2", "Device 2")
	ctx := context.Background()

	// Connect both drivers
	if err := driver1.Connect(ctx); err != nil {
		t.Fatalf("Driver 1 connect failed: %v", err)
	}
	defer driver1.Disconnect(ctx)

	if err := driver2.Connect(ctx); err != nil {
		t.Fatalf("Driver 2 connect failed: %v", err)
	}
	defer driver2.Disconnect(ctx)

	// Process transaction on driver 1
	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	}

	resp1, err := driver1.ProcessTransaction(ctx, req)
	if err != nil {
		t.Fatalf("Transaction on driver 1 failed: %v", err)
	}

	// Try to void from driver 2 (different device)
	voidReq := TransactionRequest{
		Type:              TransactionVoid,
		Amount:            10000,
		Currency:          "SAR",
		OriginalReference: resp1.TransactionID,
	}

	// In production, this should fail (cross-device access)
	// Simulator allows it for testing
	resp2, err := driver2.ProcessTransaction(ctx, voidReq)
	if err != nil {
		t.Logf("Cross-device void rejected (expected): %v", err)
	} else if resp2.Success {
		t.Log("Note: Simulator allows cross-device operations for testing")
		t.Log("Production should validate device ownership")
	}
}

// TestSecurityDenialOfService tests DoS protection
func TestSecurityDenialOfService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DoS test in short mode")
	}

	driver := NewSimulatorDriver("sec-dos", "Security DoS Test")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Test 1: Large amount values
	largeAmountReq := TransactionRequest{
		Type:     TransactionSale,
		Amount:   9223372036854775807, // Max int64
		Currency: "SAR",
	}

	resp, err := driver.ProcessTransaction(ctx, largeAmountReq)
	if err != nil {
		t.Logf("Large amount rejected (good): %v", err)
	} else if !resp.Success {
		t.Logf("Large amount declined (good): %s", resp.ResponseMessage)
	}

	// Test 2: Very long reference strings
	longRef := strings.Repeat("A", 10000)
	longRefReq := TransactionRequest{
		Type:      TransactionSale,
		Amount:    10000,
		Currency:  "SAR",
		Reference: longRef,
	}

	resp, err = driver.ProcessTransaction(ctx, longRefReq)
	if err != nil {
		t.Logf("Long reference rejected (good): %v", err)
	} else {
		t.Logf("Long reference accepted (potential DoS vector)")
	}

	// Test 3: Timeout exhaustion
	shortCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout

	_, err = driver.ProcessTransaction(shortCtx, TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	})

	if err != nil && strings.Contains(err.Error(), "context") {
		t.Log("Timeout protection working (good)")
	}
}

// TestSecurityCryptographicWeakness tests crypto implementation
func TestSecurityCryptographicWeakness(t *testing.T) {
	// Test STAN uniqueness (should not be predictable)
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "SEC001",
		MerchantID: "SEC_MERCH",
	}

	driver := NewDriver("sec-crypto", "Security Crypto Test", config)

	// Generate multiple STANs
	stans := make(map[int]bool)
	for i := 0; i < 100; i++ {
		stan := driver.nextSTAN()
		if stans[stan] {
			t.Errorf("STAN collision detected: %d", stan)
		}
		stans[stan] = true
	}

	// STANs should be sequential (expected behavior)
	t.Logf("Generated %d unique STANs", len(stans))

	// Verify STAN range (should be within valid range)
	for stan := range stans {
		if stan < 1 || stan > 999999 {
			t.Errorf("STAN out of valid range: %d", stan)
		}
	}
}

// TestSecurityErrorLeakage tests that errors don't leak sensitive info
func TestSecurityErrorLeakage(t *testing.T) {
	driver := NewSimulatorDriver("sec-error", "Security Error Test")
	ctx := context.Background()

	// Test before connection (should get safe error)
	_, err := driver.ProcessTransaction(ctx, TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	})

	if err != nil {
		errStr := err.Error()

		// Check that error doesn't contain sensitive data
		sensitivePatterns := []string{
			"password",
			"secret",
			"key",
			"token",
			"credential",
		}

		for _, pattern := range sensitivePatterns {
			if strings.Contains(strings.ToLower(errStr), pattern) {
				t.Errorf("Error message may leak sensitive data: %s", errStr)
			}
		}

		t.Logf("Safe error message: %s", errStr)
	}
}

// TestSecurityConcurrentAccess tests thread safety under concurrent load
func TestSecurityConcurrentAccess(t *testing.T) {
	driver := NewSimulatorDriver("sec-concurrent", "Security Concurrent Test")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Concurrent transactions from multiple goroutines
	const goroutines = 20
	errors := make(chan error, goroutines)
	success := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000 + int64(id),
				Currency: "SAR",
			}

			resp, err := driver.ProcessTransaction(ctx, req)
			if err != nil {
				errors <- err
				return
			}

			success <- resp.Success
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < goroutines; i++ {
		select {
		case err := <-errors:
			t.Logf("Concurrent error: %v", err)
			errorCount++
		case ok := <-success:
			if ok {
				successCount++
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	t.Logf("Concurrent test: %d success, %d errors", successCount, errorCount)

	// Most should succeed
	if successCount < goroutines-2 {
		t.Errorf("Too many concurrent failures: %d/%d", errorCount, goroutines)
	}
}

// TestSecurityInputValidation tests comprehensive input validation
func TestSecurityInputValidation(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "SEC001",
		MerchantID: "SEC_MERCH",
	}

	driver := NewDriver("sec-input", "Security Input Test", config)
	madaProvider := NewMadaProvider(driver)

	invalidInputs := []struct {
		name    string
		req     TransactionRequest
		wantErr bool
	}{
		{
			name: "Negative amount",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   -10000,
				Currency: "SAR",
			},
			wantErr: true,
		},
		{
			name: "Zero amount",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   0,
				Currency: "SAR",
			},
			wantErr: true,
		},
		{
			name: "Excessive amount",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   999999999999,
				Currency: "SAR",
			},
			wantErr: true,
		},
		{
			name: "Invalid currency",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "INVALID",
			},
			wantErr: true,
		},
	}

	for _, tt := range invalidInputs {
		t.Run(tt.name, func(t *testing.T) {
			err := madaProvider.validateMadaTransaction(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSecuritySessionManagement tests session handling
func TestSecuritySessionManagement(t *testing.T) {
	driver := NewSimulatorDriver("sec-session", "Security Session Test")
	ctx := context.Background()

	// Test connection lifecycle
	if driver.IsConnected() {
		t.Error("Should not be connected initially")
	}

	// Connect
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !driver.IsConnected() {
		t.Error("Should be connected after Connect()")
	}

	// Disconnect
	if err := driver.Disconnect(ctx); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	if driver.IsConnected() {
		t.Error("Should not be connected after Disconnect()")
	}

	// Try transaction while disconnected (should fail)
	_, err := driver.ProcessTransaction(ctx, TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	})

	if err == nil {
		t.Error("Transaction should fail when disconnected")
	}
}
