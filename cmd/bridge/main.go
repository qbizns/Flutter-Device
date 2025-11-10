package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	grpcapi "github.com/Macber-eg/Flutter-Device/internal/api/grpc"
	"github.com/Macber-eg/Flutter-Device/internal/api/rest"
	"github.com/Macber-eg/Flutter-Device/internal/api/ws"
	"github.com/Macber-eg/Flutter-Device/internal/app"
	"github.com/Macber-eg/Flutter-Device/internal/config"
	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/drivers/printer_escpos"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/jobs"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	"github.com/Macber-eg/Flutter-Device/test/virtual_devices"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
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

		var device devices.Device

		switch devCfg.Kind {
		case "printer.escpos":
			// Create ESC/POS printer
			device = printer_escpos.NewDriver(printer_escpos.Config{
				ID:       devCfg.ID,
				Name:     devCfg.Name,
				Address:  devCfg.Address,
				Port:     devCfg.Port,
				Timeout:  5 * time.Second,
				Metadata: devCfg.Metadata,
			}, logger)

		case "printer.virtual":
			// Create virtual printer
			device = virtual_devices.NewVirtualPrinter(devCfg.ID, devCfg.Name, logger)

		case "scanner.virtual":
			// Create virtual scanner
			device = virtual_devices.NewVirtualScanner(devCfg.ID, devCfg.Name, logger, eventBus)

		default:
			logger.Warn("unsupported device kind",
				telemetry.String("device_id", devCfg.ID),
				telemetry.String("kind", devCfg.Kind),
			)
			continue
		}

		// Register device
		if err := registry.Register(device); err != nil {
			logger.Error("failed to register device",
				telemetry.String("device_id", devCfg.ID),
				telemetry.Error(err),
			)
		} else {
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
	if err := registry.StartAll(ctx); err != nil {
		logger.Error("error starting devices", telemetry.Error(err))
	}

	// Start health monitoring
	go registry.MonitorHealth(ctx, 30*time.Second)

	// Create gRPC API server
	grpcAPIServer := grpcapi.NewServer(registry, jobQueue, eventBus, logger)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		logger.Fatal("failed to listen", telemetry.Error(err))
	}

	s := grpc.NewServer()
	pb.RegisterDeviceBridgeServer(s, grpcAPIServer)

	go func() {
		logger.Info("starting gRPC server", telemetry.Int("port", cfg.Server.GRPCPort))
		if err := s.Serve(lis); err != nil {
			logger.Error("gRPC server failed", telemetry.Error(err))
		}
	}()

	// Create REST gateway server
	restServer := rest.NewServer(rest.Config{
		GRPCAddress: fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort),
		HTTPPort:    cfg.Server.HTTPPort,
	})

	// Start REST server
	if err := restServer.Start(ctx); err != nil {
		logger.Fatal("failed to start REST server", telemetry.Error(err))
	}
	logger.Info("starting REST gateway", telemetry.Int("port", cfg.Server.HTTPPort))

	// Create WebSocket server
	wsServer := ws.NewServer(ws.Config{
		Registry: registry,
		EventBus: eventBus,
		Logger:   logger,
	})
	wsServer.Start(ctx)
	logger.Info("starting WebSocket server", telemetry.Int("port", cfg.Server.HTTPPort))

	// Serve Swagger UI and WebSocket endpoints
	mux := http.NewServeMux()
	rest.AddSwaggerRoutes(mux, "docs/api/devicebridge.swagger.json")
	wsServer.RegisterRoutes(mux)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
		logger.Info("swagger UI available", telemetry.String("url", fmt.Sprintf("http://localhost:%d/swagger/", cfg.Server.HTTPPort)))
		logger.Info("websocket endpoints available",
			telemetry.String("scanner", fmt.Sprintf("ws://localhost:%d/ws/scanner/:device_id", cfg.Server.HTTPPort)),
			telemetry.String("payment", fmt.Sprintf("ws://localhost:%d/ws/payment/:device_id", cfg.Server.HTTPPort)),
			telemetry.String("devices", fmt.Sprintf("ws://localhost:%d/ws/devices", cfg.Server.HTTPPort)),
		)
		// Note: The REST server handles /v1/* routes, Swagger handles /swagger/*, WebSocket handles /ws/*
		if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", telemetry.Error(err))
		}
	}()

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
	fmt.Printf("║  gRPC:      localhost:%d                               ║\n", cfg.Server.GRPCPort)
	fmt.Printf("║  REST:      http://localhost:%d/v1                     ║\n", cfg.Server.HTTPPort)
	fmt.Printf("║  WebSocket: ws://localhost:%d/ws                       ║\n", cfg.Server.HTTPPort)
	fmt.Printf("║  Swagger:   http://localhost:%d/swagger                ║\n", cfg.Server.HTTPPort)
	fmt.Printf("║  Metrics:   http://localhost:%d/metrics                ║\n", cfg.Server.MetricsPort)
	fmt.Printf("║                                                           ║\n")
	fmt.Printf("║  Devices: %d registered                                    ║\n", registry.Count())
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

	// Stop WebSocket server
	logger.Info("stopping WebSocket server")
	wsServer.Stop()

	// Stop REST server
	logger.Info("stopping REST server")
	if err := restServer.Stop(shutdownCtx); err != nil {
		logger.Error("error stopping REST server", telemetry.Error(err))
	}

	// Stop gRPC server
	logger.Info("stopping gRPC server")
	s.GracefulStop()

	// Stop job queue
	logger.Info("stopping job queue")
	jobQueue.Stop()

	// Stop devices
	logger.Info("stopping devices")
	registry.StopAll()

	logger.Info("Device Bridge stopped")
	fmt.Println("\nDevice Bridge stopped gracefully")
}
