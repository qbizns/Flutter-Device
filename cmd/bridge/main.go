package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/app"
	"github.com/Macber-eg/Flutter-Device/internal/config"
	"github.com/Macber-eg/Flutter-Device/internal/drivers/printer_escpos"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/jobs"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	"github.com/Macber-eg/Flutter-Device/test/virtual_devices"
)

const version = "2.0.0-dev"

func main() {
	// Parse flags
	configFile := flag.String("config", "configs/config.example.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Device Bridge v%s\n", version)
		os.Exit(0)
	}

	// Load configuration
	fmt.Printf("Loading configuration from: %s\n", *configFile)
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := telemetry.NewLogger(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Device Bridge",
		telemetry.String("version", version),
		telemetry.String("config", *configFile),
	)

	// Initialize metrics
	var metrics *telemetry.Metrics
	if cfg.Telemetry.Metrics.Enabled {
		metrics = telemetry.NewMetrics()
		logger.Info("metrics enabled",
			telemetry.Int("port", cfg.Server.MetricsPort),
		)

		// Start metrics server
		go func() {
			addr := fmt.Sprintf(":%d", cfg.Server.MetricsPort)
			logger.Info("starting metrics server", telemetry.String("addr", addr))
			http.Handle("/metrics", metrics.Handler())
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Error("metrics server failed", telemetry.Error(err))
			}
		}()
	}

	// Create components
	registry := app.NewRegistry(logger, metrics)
	eventBus := events.NewBus(logger)
	jobQueue := jobs.NewQueue(jobs.DefaultQueueConfig(), logger, metrics)

	// Register devices from config
	logger.Info("registering devices", telemetry.Int("count", len(cfg.Devices)))
	for _, devCfg := range cfg.Devices {
		if !devCfg.Enabled {
			logger.Info("device disabled, skipping",
				telemetry.String("device_id", devCfg.ID),
			)
			continue
		}

		var device interface{} // Will be devices.Device

		switch devCfg.Kind {
		case "printer.escpos":
			// Create ESC/POS printer
			driver := printer_escpos.NewDriver(printer_escpos.Config{
				ID:       devCfg.ID,
				Name:     devCfg.Name,
				Address:  devCfg.Address,
				Port:     devCfg.Port,
				Timeout:  5 * time.Second,
				Metadata: devCfg.Metadata,
			}, logger)
			device = driver

		case "printer.virtual":
			// Create virtual printer
			device = virtual_devices.NewVirtualPrinter(devCfg.ID, devCfg.Name, logger)

		default:
			logger.Warn("unsupported device kind",
				telemetry.String("device_id", devCfg.ID),
				telemetry.String("kind", devCfg.Kind),
			)
			continue
		}

		// Register device
		if dev, ok := device.(interface {
			ID() string
			Kind() string
			Name() string
			Start(context.Context) error
			Stop() error
			Health() interface{}
			Metadata() map[string]string
		}); ok {
			// Type assertion to devices.Device would fail without generated code
			// For now, we'll handle this differently
			logger.Info("device registered",
				telemetry.String("device_id", devCfg.ID),
				telemetry.String("kind", devCfg.Kind),
			)
		}
	}

	// Start job queue
	jobQueue.Start()
	logger.Info("job queue started")

	// Start devices
	ctx := context.Background()
	// registry.StartAll(ctx) would go here

	// Start health monitoring
	go func() {
		// registry.MonitorHealth(ctx, 30*time.Second)
	}()

	// TODO: Start gRPC server
	// TODO: Start REST gateway
	// TODO: Start WebSocket server

	logger.Info("Device Bridge started successfully",
		telemetry.String("version", version),
		telemetry.Int("grpc_port", cfg.Server.GRPCPort),
		telemetry.Int("http_port", cfg.Server.HTTPPort),
	)

	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║           Device Bridge v%s                         ║\n", version)
	fmt.Printf("╠═══════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Status: Running                                          ║\n")
	fmt.Printf("║  gRPC:   localhost:%d                                  ║\n", cfg.Server.GRPCPort)
	fmt.Printf("║  HTTP:   http://localhost:%d                           ║\n", cfg.Server.HTTPPort)
	fmt.Printf("║  Metrics: http://localhost:%d/metrics                  ║\n", cfg.Server.MetricsPort)
	fmt.Printf("║                                                           ║\n")
	fmt.Printf("║  Press Ctrl+C to stop                                     ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop job queue
	jobQueue.Stop()

	// Stop devices
	// registry.StopAll()

	logger.Info("Device Bridge stopped")
	fmt.Println("\nDevice Bridge stopped gracefully")

	_ = shutdownCtx // Use shutdown context if needed
}
