package payment_tcp

import (
	"context"
	"testing"
	"time"
)

// TestMadaProvider tests Mada provider functionality
func TestMadaProvider(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "MADA001",
		MerchantID: "MADA_MERCH",
		Provider:   "mada",
	}

	driver := NewDriver("mada-01", "Mada Terminal", config)
	provider := NewMadaProvider(driver)

	// Test validation
	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000, // 100.00 SAR
		Currency: "SAR",
	}

	err := provider.validateMadaTransaction(req)
	if err != nil {
		t.Errorf("Valid Mada transaction failed validation: %v", err)
	}
}

// TestMadaValidation tests Mada-specific validation
func TestMadaValidation(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "MADA001",
		MerchantID: "MADA_MERCH",
	}

	driver := NewDriver("mada-01", "Mada Terminal", config)
	provider := NewMadaProvider(driver)

	tests := []struct {
		name    string
		req     TransactionRequest
		wantErr bool
	}{
		{
			name: "Valid SAR transaction",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "SAR",
			},
			wantErr: false,
		},
		{
			name: "Invalid currency",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "USD",
			},
			wantErr: true,
		},
		{
			name: "Amount too low",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   50, // Less than 1 SAR
				Currency: "SAR",
			},
			wantErr: true,
		},
		{
			name: "Amount too high",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   20000000, // More than 100,000 SAR
				Currency: "SAR",
			},
			wantErr: true,
		},
		{
			name: "Balance inquiry not supported",
			req: TransactionRequest{
				Type:     TransactionBalanceInquiry,
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

// TestMadaCardDetection tests Mada card BIN detection
func TestMadaCardDetection(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "MADA001",
		MerchantID: "MADA_MERCH",
	}

	driver := NewDriver("mada-01", "Mada Terminal", config)
	provider := NewMadaProvider(driver)

	tests := []struct {
		name         string
		maskedNumber string
		want         bool
	}{
		{"Mada BIN 400861", "400861****1234", true},
		{"Mada BIN 968201", "968201****5678", true},
		{"Non-Mada BIN", "411111****1111", false},
		{"Invalid format", "1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.isMadaCard(tt.maskedNumber)
			if result != tt.want {
				t.Errorf("isMadaCard(%s) = %v, want %v", tt.maskedNumber, result, tt.want)
			}
		})
	}
}

// TestMadaAmountFormatting tests amount formatting
func TestMadaAmountFormatting(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "MADA001",
		MerchantID: "MADA_MERCH",
	}

	driver := NewDriver("mada-01", "Mada Terminal", config)
	provider := NewMadaProvider(driver)

	tests := []struct {
		name     string
		halalas  int64
		expected string
	}{
		{"100 halalas", 100, "1.00 SAR"},
		{"1000 halalas", 1000, "10.00 SAR"},
		{"12345 halalas", 12345, "123.45 SAR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.FormatAmount(tt.halalas)
			if result != tt.expected {
				t.Errorf("FormatAmount(%d) = %s, want %s", tt.halalas, result, tt.expected)
			}
		})
	}
}

// TestMadaAmountParsing tests amount parsing
func TestMadaAmountParsing(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "MADA001",
		MerchantID: "MADA_MERCH",
	}

	driver := NewDriver("mada-01", "Mada Terminal", config)
	provider := NewMadaProvider(driver)

	tests := []struct {
		name      string
		amountStr string
		expected  int64
		wantErr   bool
	}{
		{"With SAR suffix", "10.00 SAR", 1000, false},
		{"Without suffix", "10.00", 1000, false},
		{"Decimal", "123.45", 12345, false},
		{"Invalid format", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.ParseAmount(tt.amountStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseAmount(%s) = %d, want %d", tt.amountStr, result, tt.expected)
			}
		})
	}
}

// TestSimulator tests the payment simulator
func TestSimulator(t *testing.T) {
	sim := NewSimulator("TEST001", "TEST_MERCH")

	// Test approved transaction
	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000, // Ends in 00, should be approved
		Currency: "SAR",
	}

	response := sim.SimulateTransaction(req)

	if !response.Success {
		t.Errorf("Expected approved transaction, got declined: %s", response.ResponseMessage)
	}

	if response.ResponseCode != "00" {
		t.Errorf("Expected response code 00, got %s", response.ResponseCode)
	}

	if response.AuthCode == "" {
		t.Error("Expected auth code for approved transaction")
	}
}

// TestSimulatorDeclines tests simulator decline scenarios
func TestSimulatorDeclines(t *testing.T) {
	sim := NewSimulator("TEST001", "TEST_MERCH")

	tests := []struct {
		name         string
		amount       int64
		expectedCode string
		expectedMsg  string
	}{
		{"Insufficient funds", 10005, "51", "Insufficient funds"},
		{"Expired card", 10054, "54", "Expired card"},
		{"Incorrect PIN", 10055, "55", "Incorrect PIN"},
		{"Issuer unavailable", 10091, "91", "Issuer unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   tt.amount,
				Currency: "SAR",
			}

			response := sim.SimulateTransaction(req)

			if response.Success {
				t.Error("Expected declined transaction, got approved")
			}

			if response.ResponseCode != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, response.ResponseCode)
			}

			if response.ResponseMessage != tt.expectedMsg {
				t.Errorf("Expected message %s, got %s", tt.expectedMsg, response.ResponseMessage)
			}
		})
	}
}

// TestSimulatorVoidRefund tests void and refund functionality
func TestSimulatorVoidRefund(t *testing.T) {
	sim := NewSimulator("TEST001", "TEST_MERCH")

	// First, create an approved transaction
	saleReq := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	}

	saleResp := sim.SimulateTransaction(saleReq)
	if !saleResp.Success {
		t.Fatal("Sale transaction should be approved")
	}

	// Now void it
	voidReq := TransactionRequest{
		Type:              TransactionVoid,
		Amount:            10000,
		Currency:          "SAR",
		OriginalReference: saleResp.TransactionID,
	}

	voidResp := sim.SimulateTransaction(voidReq)
	if !voidResp.Success {
		t.Errorf("Void should be approved, got: %s", voidResp.ResponseMessage)
	}

	// Try to void again (should fail)
	voidReq2 := TransactionRequest{
		Type:              TransactionVoid,
		Amount:            10000,
		Currency:          "SAR",
		OriginalReference: saleResp.TransactionID,
	}

	voidResp2 := sim.SimulateTransaction(voidReq2)
	if voidResp2.Success {
		t.Error("Second void should fail (transaction already voided)")
	}
}

// TestSimulatorDriver tests the simulator driver
func TestSimulatorDriver(t *testing.T) {
	driver := NewSimulatorDriver("sim-01", "Test Simulator")

	// Test connection
	ctx := context.Background()
	err := driver.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !driver.IsConnected() {
		t.Error("Driver should be connected")
	}

	// Test transaction
	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	}

	response, err := driver.ProcessTransaction(ctx, req)
	if err != nil {
		t.Fatalf("ProcessTransaction failed: %v", err)
	}

	if !response.Success {
		t.Errorf("Transaction should be approved, got: %s", response.ResponseMessage)
	}

	// Test disconnect
	err = driver.Disconnect(ctx)
	if err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	if driver.IsConnected() {
		t.Error("Driver should be disconnected")
	}
}

// TestSimulatorTestHelper tests the test helper
func TestSimulatorTestHelper(t *testing.T) {
	helper := NewSimulatorTestHelper()

	// Test approved transaction
	approvedReq := helper.CreateApprovedTransaction(10000)
	approvedResp, err := helper.ProcessTestTransaction(approvedReq)
	if err != nil {
		t.Fatalf("ProcessTestTransaction failed: %v", err)
	}

	if !approvedResp.Success {
		t.Error("Expected approved transaction")
	}

	// Test declined transaction
	declinedReq := helper.CreateDeclinedTransaction(10000, "insufficient_funds")
	declinedResp, err := helper.ProcessTestTransaction(declinedReq)
	if err != nil {
		t.Fatalf("ProcessTestTransaction failed: %v", err)
	}

	if declinedResp.Success {
		t.Error("Expected declined transaction")
	}

	if declinedResp.ResponseCode != "51" {
		t.Errorf("Expected code 51, got %s", declinedResp.ResponseCode)
	}
}

// TestSimulatorSettlement tests settlement simulation
func TestSimulatorSettlement(t *testing.T) {
	driver := NewSimulatorDriver("sim-01", "Test Simulator")
	ctx := context.Background()

	// Connect
	err := driver.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Process some approved transactions
	approvedCount := 0
	for i := 0; i < 5; i++ {
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

	// Settlement
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

	// Check that settlement count matches approved transactions
	if settlementResp.TotalCount != approvedCount {
		t.Errorf("Expected %d transactions, got %d", approvedCount, settlementResp.TotalCount)
	}
}

// TestLuhnCheck tests Luhn algorithm validation
func TestLuhnCheck(t *testing.T) {
	tests := []struct {
		name       string
		cardNumber string
		valid      bool
	}{
		{"Valid number", "4532015112830366", true},
		{"Invalid number", "4532015112830367", false},
		{"Non-numeric", "453201511283036A", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := luhnCheck(tt.cardNumber)
			if result != tt.valid {
				t.Errorf("luhnCheck(%s) = %v, want %v", tt.cardNumber, result, tt.valid)
			}
		})
	}
}

// TestMadaReceipt tests Mada receipt generation
func TestMadaReceipt(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "MADA001",
		MerchantID: "MADA_MERCH",
	}

	driver := NewDriver("mada-01", "Mada Terminal", config)
	provider := NewMadaProvider(driver)

	response := &TransactionResponse{
		Success:          true,
		Status:           StatusApproved,
		TransactionID:    "123456789012",
		AuthCode:         "ABC123",
		ResponseCode:     "00",
		ResponseMessage:  "Approved",
		CardNumberMasked: "400861****1234",
		CardType:         CardTypeMada,
		EntryMode:        EntryModeChip,
		Amount:           10000,
		Currency:         "SAR",
		Timestamp:        time.Now(),
		TerminalID:       "MADA001",
		MerchantID:       "MADA_MERCH",
		MerchantName:     "Test Merchant",
		RRN:              "123456789012",
		STAN:             "000001",
	}

	receipt := provider.BuildMadaReceipt(response)

	if len(receipt) == 0 {
		t.Error("Receipt should not be empty")
	}

	// Check for key elements
	found := false
	for _, line := range receipt {
		if line == "       MADA TRANSACTION         " {
			found = true
			break
		}
	}

	if !found {
		t.Error("Receipt should contain MADA TRANSACTION header")
	}
}

// TestMadaTransactionLimits tests Mada transaction limits
func TestMadaTransactionLimits(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "MADA001",
		MerchantID: "MADA_MERCH",
	}

	driver := NewDriver("mada-01", "Mada Terminal", config)
	provider := NewMadaProvider(driver)

	limits := provider.GetMadaTransactionLimits()

	if limits["min_amount_halalas"] != 100 {
		t.Errorf("Expected min amount 100, got %v", limits["min_amount_halalas"])
	}

	if limits["max_amount_halalas"] != 10000000 {
		t.Errorf("Expected max amount 10000000, got %v", limits["max_amount_halalas"])
	}

	if limits["currency"] != "SAR" {
		t.Errorf("Expected currency SAR, got %v", limits["currency"])
	}
}

// TestKNETProvider tests KNET provider functionality
func TestKNETProvider(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
		Provider:   "knet",
	}

	driver := NewDriver("knet-01", "KNET Terminal", config)
	provider := NewKNETProvider(driver)

	// Test validation
	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   1000, // 1.000 KWD
		Currency: "KWD",
	}

	err := provider.validateKNETTransaction(req)
	if err != nil {
		t.Errorf("Valid KNET transaction failed validation: %v", err)
	}
}

// TestKNETValidation tests KNET-specific validation
func TestKNETValidation(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
	}

	driver := NewDriver("knet-01", "KNET Terminal", config)
	provider := NewKNETProvider(driver)

	tests := []struct {
		name    string
		req     TransactionRequest
		wantErr bool
	}{
		{
			name: "Valid KWD transaction",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   1000,
				Currency: "KWD",
			},
			wantErr: false,
		},
		{
			name: "Invalid currency",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   1000,
				Currency: "SAR",
			},
			wantErr: true,
		},
		{
			name: "Amount too low",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   50, // Less than 0.100 KWD
				Currency: "KWD",
			},
			wantErr: true,
		},
		{
			name: "Amount too high",
			req: TransactionRequest{
				Type:     TransactionSale,
				Amount:   6000000, // More than 5,000 KWD
				Currency: "KWD",
			},
			wantErr: true,
		},
		{
			name: "Pre-auth not supported",
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

// TestKNETCardDetection tests KNET card BIN detection
func TestKNETCardDetection(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
	}

	driver := NewDriver("knet-01", "KNET Terminal", config)
	provider := NewKNETProvider(driver)

	tests := []struct {
		name         string
		maskedNumber string
		want         bool
	}{
		{"KNET BIN 420644", "420644****1234", true},
		{"KNET BIN 529415", "529415****5678", true},
		{"Non-KNET BIN", "411111****1111", false},
		{"Invalid format", "1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.isKNETCard(tt.maskedNumber)
			if result != tt.want {
				t.Errorf("isKNETCard(%s) = %v, want %v", tt.maskedNumber, result, tt.want)
			}
		})
	}
}

// TestKNETAmountFormatting tests amount formatting
func TestKNETAmountFormatting(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
	}

	driver := NewDriver("knet-01", "KNET Terminal", config)
	provider := NewKNETProvider(driver)

	tests := []struct {
		name     string
		fils     int64
		expected string
	}{
		{"100 fils", 100, "0.100 KWD"},
		{"1000 fils", 1000, "1.000 KWD"},
		{"12345 fils", 12345, "12.345 KWD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.FormatAmount(tt.fils)
			if result != tt.expected {
				t.Errorf("FormatAmount(%d) = %s, want %s", tt.fils, result, tt.expected)
			}
		})
	}
}

// TestKNETAmountParsing tests amount parsing
func TestKNETAmountParsing(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
	}

	driver := NewDriver("knet-01", "KNET Terminal", config)
	provider := NewKNETProvider(driver)

	tests := []struct {
		name      string
		amountStr string
		expected  int64
		wantErr   bool
	}{
		{"With KWD suffix", "1.000 KWD", 1000, false},
		{"Without suffix", "1.000", 1000, false},
		{"Decimal", "12.345", 12345, false},
		{"Invalid format", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.ParseAmount(tt.amountStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseAmount(%s) = %d, want %d", tt.amountStr, result, tt.expected)
			}
		})
	}
}

// TestKNETReceipt tests KNET receipt generation
func TestKNETReceipt(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
	}

	driver := NewDriver("knet-01", "KNET Terminal", config)
	provider := NewKNETProvider(driver)

	response := &TransactionResponse{
		Success:          true,
		Status:           StatusApproved,
		TransactionID:    "123456789012",
		AuthCode:         "ABC123",
		ResponseCode:     "00",
		ResponseMessage:  "Approved",
		CardNumberMasked: "420644****1234",
		CardType:         CardTypeKNET,
		EntryMode:        EntryModeChip,
		Amount:           1000,
		Currency:         "KWD",
		Timestamp:        time.Now(),
		TerminalID:       "KNET001",
		MerchantID:       "KNET_MERCH",
		MerchantName:     "Test Merchant",
		RRN:              "123456789012",
		STAN:             "000001",
	}

	receipt := provider.BuildKNETReceipt(response)

	if len(receipt) == 0 {
		t.Error("Receipt should not be empty")
	}

	// Check for key elements
	found := false
	for _, line := range receipt {
		if line == "       KNET TRANSACTION         " {
			found = true
			break
		}
	}

	if !found {
		t.Error("Receipt should contain KNET TRANSACTION header")
	}
}

// TestKNETTransactionLimits tests KNET transaction limits
func TestKNETTransactionLimits(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
	}

	driver := NewDriver("knet-01", "KNET Terminal", config)
	provider := NewKNETProvider(driver)

	limits := provider.GetKNETTransactionLimits()

	if limits["min_amount_fils"] != 100 {
		t.Errorf("Expected min amount 100, got %v", limits["min_amount_fils"])
	}

	if limits["max_amount_fils"] != 5000000 {
		t.Errorf("Expected max amount 5000000, got %v", limits["max_amount_fils"])
	}

	if limits["currency"] != "KWD" {
		t.Errorf("Expected currency KWD, got %v", limits["currency"])
	}

	if limits["decimal_places"] != 3 {
		t.Errorf("Expected 3 decimal places, got %v", limits["decimal_places"])
	}
}

// TestKNETCardType tests KNET card type detection
func TestKNETCardType(t *testing.T) {
	config := ConnectionConfig{
		Host:       "localhost",
		Port:       3000,
		TerminalID: "KNET001",
		MerchantID: "KNET_MERCH",
	}

	driver := NewDriver("knet-01", "KNET Terminal", config)
	provider := NewKNETProvider(driver)

	tests := []struct {
		name         string
		maskedNumber string
		expected     string
	}{
		{"KNET Visa", "420644****1234", "knet-visa"},
		{"KNET MasterCard", "529415****5678", "knet-mastercard"},
		{"KNET Debit", "445564****9012", "knet-debit"},
		{"Non-KNET", "411111****1111", "non-knet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.GetKNETCardType(tt.maskedNumber)
			if result != tt.expected {
				t.Errorf("GetKNETCardType(%s) = %s, want %s", tt.maskedNumber, result, tt.expected)
			}
		})
	}
}
