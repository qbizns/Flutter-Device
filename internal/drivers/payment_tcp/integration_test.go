package payment_tcp

import (
	"context"
	"testing"
	"time"
)

// TestIntegrationFullPaymentFlow tests the complete payment flow
func TestIntegrationFullPaymentFlow(t *testing.T) {
	// Create simulator driver
	driver := NewSimulatorDriver("int-test-01", "Integration Test")
	ctx := context.Background()

	// Test connection
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	if !driver.IsConnected() {
		t.Fatal("Driver should be connected")
	}

	// Test successful sale
	saleReq := TransactionRequest{
		Type:      TransactionSale,
		Amount:    10000, // 100.00 SAR
		Currency:  "SAR",
		Reference: "TEST001",
		Timeout:   30 * time.Second,
	}

	saleResp, err := driver.ProcessTransaction(ctx, saleReq)
	if err != nil {
		t.Fatalf("Sale transaction failed: %v", err)
	}

	if !saleResp.Success {
		t.Errorf("Sale should be approved: %s", saleResp.ResponseMessage)
	}

	if saleResp.ResponseCode != "00" {
		t.Errorf("Expected response code 00, got %s", saleResp.ResponseCode)
	}

	if saleResp.Amount != 10000 {
		t.Errorf("Expected amount 10000, got %d", saleResp.Amount)
	}

	if saleResp.AuthCode == "" {
		t.Error("Auth code should not be empty for approved transaction")
	}

	// Test void of the sale
	voidReq := TransactionRequest{
		Type:              TransactionVoid,
		Amount:            10000,
		Currency:          "SAR",
		OriginalReference: saleResp.TransactionID,
		Timeout:           30 * time.Second,
	}

	voidResp, err := driver.ProcessTransaction(ctx, voidReq)
	if err != nil {
		t.Fatalf("Void transaction failed: %v", err)
	}

	if !voidResp.Success {
		t.Errorf("Void should be approved: %s", voidResp.ResponseMessage)
	}

	// Test settlement
	settlementReq := SettlementRequest{
		BatchNumber: time.Now().Format("20060102"),
	}

	settlementResp, err := driver.Settlement(ctx, settlementReq)
	if err != nil {
		t.Fatalf("Settlement failed: %v", err)
	}

	if !settlementResp.Success {
		t.Error("Settlement should succeed")
	}

	// After settlement, transaction count should be reset
	status := driver.GetStatus()
	if status.TransactionCount != 0 {
		t.Errorf("Transaction count should be 0 after settlement, got %d", status.TransactionCount)
	}
}

// TestIntegrationMadaProvider tests Mada provider integration
func TestIntegrationMadaProvider(t *testing.T) {
	// Create simulator driver
	simDriver := NewSimulatorDriver("mada-int-test", "Mada Integration Test")
	ctx := context.Background()

	if err := simDriver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer simDriver.Disconnect(ctx)

	// Create Mada provider wrapping the simulator
	// Note: In real use, this would wrap a real TCP driver
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "MADA001",
		MerchantID: "MADA_MERCH",
		Provider:   "mada",
	}

	driver := NewDriver("mada-01", "Mada Test", config)
	provider := NewMadaProvider(driver)

	// Test Mada-specific validation
	tests := []struct {
		name    string
		req     TransactionRequest
		wantErr bool
	}{
		{
			name: "Valid Mada transaction",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "SAR",
			},
			wantErr: false,
		},
		{
			name: "Invalid currency for Mada",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "USD",
			},
			wantErr: true,
		},
		{
			name: "Amount below Mada minimum",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   50,
				Currency: "SAR",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.validateMadaTransaction(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMadaTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIntegrationKNETProvider tests KNET provider integration
func TestIntegrationKNETProvider(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
		Provider:   "knet",
	}

	driver := NewDriver("knet-01", "KNET Test", config)
	provider := NewKNETProvider(driver)

	// Test KNET-specific validation
	tests := []struct {
		name    string
		req     TransactionRequest
		wantErr bool
	}{
		{
			name: "Valid KNET transaction",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   1000,
				Currency: "KWD",
			},
			wantErr: false,
		},
		{
			name: "Invalid currency for KNET",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   1000,
				Currency: "SAR",
			},
			wantErr: true,
		},
		{
			name: "Pre-auth not supported by KNET",
			req: TransactionRequest{
				Type:     TransactionPreAuth,
				Amount:   1000,
				Currency: "KWD",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.validateKNETTransaction(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateKNETTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIntegrationAuditTrail tests audit logging integration
func TestIntegrationAuditTrail(t *testing.T) {
	// Create memory audit writer
	auditWriter := NewMemoryAuditReader()
	auditLogger := NewAuditLogger(auditWriter, 10)

	// Create simulator with audit logging
	driver := NewSimulatorDriver("audit-test", "Audit Test")
	driver.SetAuditLogger(auditLogger)

	ctx := context.Background()

	// Connect and verify connection audit
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Flush to get connection event
	if err := auditLogger.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	events := auditWriter.GetAllEvents()
	if len(events) < 1 {
		t.Fatal("Expected at least 1 audit event (connection)")
	}

	// Verify connection event
	connEvent := events[0]
	if connEvent.EventType != "connection" {
		t.Errorf("Expected connection event, got %s", connEvent.EventType)
	}
	if !connEvent.Success {
		t.Error("Connection event should be successful")
	}

	// Clear events for transaction test
	auditWriter.Clear()

	// Process transaction and verify transaction audit
	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	}

	resp, err := driver.ProcessTransaction(ctx, req)
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Flush to get transaction events
	if err := auditLogger.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	events = auditWriter.GetAllEvents()
	if len(events) == 0 {
		t.Fatal("Expected transaction audit events")
	}

	// Find transaction complete event
	var txEvent *AuditEvent
	for i := range events {
		if events[i].EventType == "transaction" && events[i].Action == "complete" {
			txEvent = &events[i]
			break
		}
	}

	if txEvent == nil {
		t.Fatal("Expected transaction complete event")
	}

	if txEvent.Amount != 10000 {
		t.Errorf("Expected amount 10000, got %d", txEvent.Amount)
	}

	if txEvent.TransactionID != resp.TransactionID {
		t.Errorf("Transaction ID mismatch: %s != %s", txEvent.TransactionID, resp.TransactionID)
	}

	if !txEvent.Success {
		t.Error("Transaction event should be successful")
	}

	if txEvent.Duration == 0 {
		t.Error("Transaction duration should be recorded")
	}
}

// TestIntegrationMultipleTransactions tests multiple transactions in sequence
func TestIntegrationMultipleTransactions(t *testing.T) {
	driver := NewSimulatorDriver("multi-test", "Multi Transaction Test")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Process multiple approved transactions
	approvedCount := 0
	for i := 0; i < 10; i++ {
		req := TransactionRequest{
			Type:     TransactionSale,
			Amount:   10000, // Ends in 00, will be approved
			Currency: "SAR",
		}

		resp, err := driver.ProcessTransaction(ctx, req)
		if err != nil {
			t.Fatalf("Transaction %d failed: %v", i, err)
		}

		if resp.Success {
			approvedCount++
		}
	}

	if approvedCount != 10 {
		t.Errorf("Expected 10 approved transactions, got %d", approvedCount)
	}

	// Verify status shows correct transaction count
	status := driver.GetStatus()
	if status.TransactionCount < 10 {
		t.Errorf("Expected at least 10 transactions, got %d", status.TransactionCount)
	}
}

// TestIntegrationConcurrentTransactions tests concurrent transaction processing
func TestIntegrationConcurrentTransactions(t *testing.T) {
	driver := NewSimulatorDriver("concurrent-test", "Concurrent Test")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Process transactions concurrently
	const numTransactions = 5
	results := make(chan *TransactionResponse, numTransactions)
	errors := make(chan error, numTransactions)

	for i := 0; i < numTransactions; i++ {
		go func(index int) {
			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000 + int64(index), // Vary amounts slightly
				Currency: "SAR",
			}

			resp, err := driver.ProcessTransaction(ctx, req)
			if err != nil {
				errors <- err
				return
			}
			results <- resp
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numTransactions; i++ {
		select {
		case resp := <-results:
			if resp.Success {
				successCount++
			}
		case err := <-errors:
			t.Errorf("Transaction failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for transactions")
		}
	}

	if successCount < numTransactions {
		t.Errorf("Expected %d successful transactions, got %d", numTransactions, successCount)
	}
}

// TestIntegrationErrorScenarios tests various error scenarios
func TestIntegrationErrorScenarios(t *testing.T) {
	driver := NewSimulatorDriver("error-test", "Error Test")
	ctx := context.Background()

	// Test transaction before connection
	_, err := driver.ProcessTransaction(ctx, TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	})
	if err == nil {
		t.Error("Expected error when processing transaction before connection")
	}

	// Connect
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Test declined transactions
	declineTests := []struct {
		name         string
		amount       int64
		expectedCode string
	}{
		{"Insufficient funds", 10005, "51"},
		{"Expired card", 10054, "54"},
		{"Incorrect PIN", 10055, "55"},
		{"Issuer unavailable", 10091, "91"},
	}

	for _, tt := range declineTests {
		t.Run(tt.name, func(t *testing.T) {
			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   tt.amount,
				Currency: "SAR",
			}

			resp, err := driver.ProcessTransaction(ctx, req)
			if err != nil {
				t.Fatalf("Transaction failed: %v", err)
			}

			if resp.Success {
				t.Error("Expected declined transaction")
			}

			if resp.ResponseCode != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, resp.ResponseCode)
			}
		})
	}

	// Test void without original reference (simulator declines, doesn't error)
	voidResp, err := driver.ProcessTransaction(ctx, TransactionRequest{
		Type:     TransactionVoid,
		Amount:   10000,
		Currency: "SAR",
	})
	if err != nil {
		t.Logf("Void without reference error: %v", err)
	}
	if voidResp != nil && voidResp.Success {
		t.Error("Void without original reference should be declined")
	}

	// Test refund without original reference (simulator declines, doesn't error)
	refundResp, err := driver.ProcessTransaction(ctx, TransactionRequest{
		Type:     TransactionRefund,
		Amount:   10000,
		Currency: "SAR",
	})
	if err != nil {
		t.Logf("Refund without reference error: %v", err)
	}
	if refundResp != nil && refundResp.Success {
		t.Error("Refund without original reference should be declined")
	}
}

// TestIntegrationDisconnectReconnect tests disconnect and reconnect scenarios
func TestIntegrationDisconnectReconnect(t *testing.T) {
	driver := NewSimulatorDriver("reconnect-test", "Reconnect Test")
	ctx := context.Background()

	// Initial connection
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Initial connect failed: %v", err)
	}

	// Verify connected
	if !driver.IsConnected() {
		t.Fatal("Driver should be connected")
	}

	// Process a transaction
	_, err := driver.ProcessTransaction(ctx, TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Disconnect
	if err := driver.Disconnect(ctx); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	// Verify disconnected
	if driver.IsConnected() {
		t.Fatal("Driver should be disconnected")
	}

	// Reconnect
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Reconnect failed: %v", err)
	}

	// Verify reconnected
	if !driver.IsConnected() {
		t.Fatal("Driver should be connected after reconnect")
	}

	// Process another transaction
	resp, err := driver.ProcessTransaction(ctx, TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	})
	if err != nil {
		t.Fatalf("Transaction after reconnect failed: %v", err)
	}

	if !resp.Success {
		t.Error("Transaction after reconnect should succeed")
	}

	// Cleanup
	driver.Disconnect(ctx)
}
