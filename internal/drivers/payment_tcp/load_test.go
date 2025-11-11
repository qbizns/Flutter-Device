package payment_tcp

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestLoad100Transactions tests processing 100 transactions
func TestLoad100Transactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	driver := NewSimulatorDriver("load-100", "Load Test 100")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	startTime := time.Now()
	successCount := 0

	for i := 0; i < 100; i++ {
		req := TransactionRequest{
			Type:     TransactionSale,
			Amount:   10000,
			Currency: "SAR",
		}

		resp, err := driver.ProcessTransaction(ctx, req)
		if err != nil {
			t.Errorf("Transaction %d failed: %v", i, err)
			continue
		}

		if resp.Success {
			successCount++
		}
	}

	duration := time.Since(startTime)
	tps := float64(100) / duration.Seconds()

	t.Logf("Processed 100 transactions in %v (%.2f TPS)", duration, tps)
	t.Logf("Success rate: %d/100 (%.1f%%)", successCount, float64(successCount))

	if successCount < 95 {
		t.Errorf("Success rate too low: %d/100", successCount)
	}

	// Target: At least 10 TPS with simulator
	if tps < 10 {
		t.Logf("Warning: TPS below target (%.2f < 10)", tps)
	}
}

// TestLoadConcurrent50 tests 50 concurrent transactions
func TestLoadConcurrent50(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	driver := NewSimulatorDriver("load-concurrent-50", "Load Concurrent 50")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	const numTransactions = 50
	var wg sync.WaitGroup
	var successCount int32
	var errorCount int32

	startTime := time.Now()

	for i := 0; i < numTransactions; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000 + int64(index%10), // Vary amounts
				Currency: "SAR",
			}

			resp, err := driver.ProcessTransaction(ctx, req)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return
			}

			if resp.Success {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)
	tps := float64(numTransactions) / duration.Seconds()

	t.Logf("Processed %d concurrent transactions in %v (%.2f TPS)", numTransactions, duration, tps)
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		t.Errorf("Got %d errors during concurrent load test", errorCount)
	}

	if successCount < numTransactions-5 {
		t.Errorf("Success rate too low: %d/%d", successCount, numTransactions)
	}
}

// TestLoadMixedTransactions tests mixed transaction types under load
func TestLoadMixedTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	driver := NewSimulatorDriver("load-mixed", "Load Mixed")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// First, create some approved sales to void/refund
	approvedTxIDs := make([]string, 0, 20)
	for i := 0; i < 20; i++ {
		req := TransactionRequest{
			Type:     TransactionSale,
			Amount:   10000,
			Currency: "SAR",
		}

		resp, err := driver.ProcessTransaction(ctx, req)
		if err != nil {
			t.Fatalf("Initial sale %d failed: %v", i, err)
		}

		if resp.Success {
			approvedTxIDs = append(approvedTxIDs, resp.TransactionID)
		}
	}

	// Now test mixed operations
	var wg sync.WaitGroup
	var successCount int32

	startTime := time.Now()

	// 30 sales
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "SAR",
			}
			resp, _ := driver.ProcessTransaction(ctx, req)
			if resp != nil && resp.Success {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	// 10 voids
	for i := 0; i < 10 && i < len(approvedTxIDs); i++ {
		wg.Add(1)
		go func(txID string) {
			defer wg.Done()
			req := TransactionRequest{
				Type:              TransactionVoid,
				Amount:            10000,
				Currency:          "SAR",
				OriginalReference: txID,
			}
			resp, _ := driver.ProcessTransaction(ctx, req)
			if resp != nil && resp.Success {
				atomic.AddInt32(&successCount, 1)
			}
		}(approvedTxIDs[i])
	}

	// 10 refunds
	for i := 10; i < 20 && i < len(approvedTxIDs); i++ {
		wg.Add(1)
		go func(txID string) {
			defer wg.Done()
			req := TransactionRequest{
				Type:              TransactionRefund,
				Amount:            10000,
				Currency:          "SAR",
				OriginalReference: txID,
			}
			resp, _ := driver.ProcessTransaction(ctx, req)
			if resp != nil && resp.Success {
				atomic.AddInt32(&successCount, 1)
			}
		}(approvedTxIDs[i])
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Processed mixed transactions in %v", duration)
	t.Logf("Success count: %d", successCount)

	// At least 40 should succeed (30 sales + 10 voids)
	if successCount < 40 {
		t.Errorf("Success count too low: %d", successCount)
	}
}

// TestLoadSettlementCycle tests multiple settlement cycles
func TestLoadSettlementCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	driver := NewSimulatorDriver("load-settlement", "Load Settlement")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Test 3 settlement cycles
	for cycle := 0; cycle < 3; cycle++ {
		t.Logf("Settlement cycle %d", cycle+1)

		// Process 50 transactions
		approvedCount := 0
		for i := 0; i < 50; i++ {
			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "SAR",
			}

			resp, err := driver.ProcessTransaction(ctx, req)
			if err != nil {
				t.Errorf("Transaction failed: %v", err)
				continue
			}

			if resp.Success {
				approvedCount++
			}
		}

		t.Logf("Cycle %d: %d approved transactions", cycle+1, approvedCount)

		// Settle batch
		settlementReq := SettlementRequest{
			BatchNumber: fmt.Sprintf("BATCH%03d", cycle+1),
		}

		settlementResp, err := driver.Settlement(ctx, settlementReq)
		if err != nil {
			t.Fatalf("Settlement %d failed: %v", cycle+1, err)
		}

		if !settlementResp.Success {
			t.Errorf("Settlement %d should succeed", cycle+1)
		}

		t.Logf("Cycle %d settled: %d transactions", cycle+1, settlementResp.TotalCount)

		// Verify status reset
		status := driver.GetStatus()
		if status.TransactionCount != 0 {
			t.Errorf("Transaction count should be 0 after settlement, got %d", status.TransactionCount)
		}
	}
}

// TestLoadSustained tests sustained load over time
func TestLoadSustained(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	driver := NewSimulatorDriver("load-sustained", "Load Sustained")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Run for 5 seconds with continuous load
	endTime := time.Now().Add(5 * time.Second)
	var totalCount int32
	var successCount int32
	var wg sync.WaitGroup

	// Start 10 workers
	for worker := 0; worker < 10; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for time.Now().Before(endTime) {
				req := TransactionRequest{
					Type:     TransactionSale,
					Amount:   10000 + int64(workerID),
					Currency: "SAR",
				}

				resp, err := driver.ProcessTransaction(ctx, req)
				atomic.AddInt32(&totalCount, 1)

				if err == nil && resp.Success {
					atomic.AddInt32(&successCount, 1)
				}

				// Small delay to avoid overwhelming the system
				time.Sleep(10 * time.Millisecond)
			}
		}(worker)
	}

	wg.Wait()

	successRate := float64(successCount) / float64(totalCount) * 100

	t.Logf("Sustained load test: %d total, %d success (%.1f%%)", totalCount, successCount, successRate)

	if successRate < 95 {
		t.Errorf("Success rate too low: %.1f%%", successRate)
	}
}

// BenchmarkTransactionThroughput benchmarks transaction processing throughput
func BenchmarkTransactionThroughput(b *testing.B) {
	driver := NewSimulatorDriver("bench-throughput", "Benchmark Throughput")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		b.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	req := TransactionRequest{
		Type:     TransactionSale,
		Amount:   10000,
		Currency: "SAR",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := driver.ProcessTransaction(ctx, req)
		if err != nil {
			b.Fatalf("Transaction failed: %v", err)
		}
	}
}

// BenchmarkConcurrentTransactions benchmarks concurrent transaction processing
func BenchmarkConcurrentTransactions(b *testing.B) {
	driver := NewSimulatorDriver("bench-concurrent", "Benchmark Concurrent")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		b.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "SAR",
			}

			_, err := driver.ProcessTransaction(ctx, req)
			if err != nil {
				b.Fatalf("Transaction failed: %v", err)
			}
		}
	})
}

// BenchmarkSettlement benchmarks settlement processing
func BenchmarkSettlement(b *testing.B) {
	driver := NewSimulatorDriver("bench-settlement", "Benchmark Settlement")
	ctx := context.Background()

	if err := driver.Connect(ctx); err != nil {
		b.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(ctx)

	// Process some transactions before each settlement
	for i := 0; i < 10; i++ {
		req := TransactionRequest{
			Type:     TransactionSale,
			Amount:   10000,
			Currency: "SAR",
		}
		driver.ProcessTransaction(ctx, req)
	}

	settlementReq := SettlementRequest{
		BatchNumber: time.Now().Format("20060102"),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := driver.Settlement(ctx, settlementReq)
		if err != nil {
			b.Fatalf("Settlement failed: %v", err)
		}

		// Re-add transactions for next settlement
		for j := 0; j < 10; j++ {
			req := TransactionRequest{
				Type:     TransactionSale,
				Amount:   10000,
				Currency: "SAR",
			}
			driver.ProcessTransaction(ctx, req)
		}
	}
}
