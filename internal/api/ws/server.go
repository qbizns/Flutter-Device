package ws

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/Macber-eg/Flutter-Device/internal/app"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Server is the WebSocket server
type Server struct {
	hub      *Hub
	registry *app.Registry
	eventBus *events.Bus
	logger   *telemetry.Logger
	upgrader websocket.Upgrader
}

// Config holds WebSocket server configuration
type Config struct {
	Registry *app.Registry
	EventBus *events.Bus
	Logger   *telemetry.Logger
}

// NewServer creates a new WebSocket server
func NewServer(cfg Config) *Server {
	hub := NewHub(cfg.EventBus, cfg.Logger)

	return &Server{
		hub:      hub,
		registry: cfg.Registry,
		eventBus: cfg.EventBus,
		logger:   cfg.Logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now
				// In production, you should validate origins
				return true
			},
		},
	}
}

// Start starts the WebSocket server
func (s *Server) Start(ctx context.Context) {
	go s.hub.Run()
	s.logger.Info("websocket hub started")
}

// Stop stops the WebSocket server
func (s *Server) Stop() {
	s.hub.Stop()
	s.logger.Info("websocket hub stopped")
}

// RegisterRoutes registers WebSocket routes on the given mux
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ws/scanner/", s.handleScanner)
	mux.HandleFunc("/ws/payment/", s.handlePayment)
	mux.HandleFunc("/ws/devices", s.handleDevices)
}

// handleScanner handles scanner event WebSocket connections
func (s *Server) handleScanner(w http.ResponseWriter, r *http.Request) {
	// Extract device ID from path: /ws/scanner/:device_id
	path := strings.TrimPrefix(r.URL.Path, "/ws/scanner/")
	deviceID := strings.Trim(path, "/")

	if deviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	// Check if device exists
	_, err := s.registry.Get(deviceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("device not found: %s", deviceID), http.StatusNotFound)
		return
	}

	// Upgrade connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("failed to upgrade connection",
			telemetry.String("device_id", deviceID),
			telemetry.Error(err),
		)
		return
	}

	// Create client
	client := NewClient(conn, deviceID, "scanner", s.logger)

	// Register client
	s.hub.Register <- client

	// Start read/write pumps
	ctx := r.Context()
	go client.WritePump(ctx)
	go client.ReadPump(ctx, s.hub)

	s.logger.Info("scanner websocket connected",
		telemetry.String("device_id", deviceID),
		telemetry.String("client_id", client.ID),
	)
}

// handlePayment handles payment event WebSocket connections
func (s *Server) handlePayment(w http.ResponseWriter, r *http.Request) {
	// Extract device ID from path: /ws/payment/:device_id
	path := strings.TrimPrefix(r.URL.Path, "/ws/payment/")
	deviceID := strings.Trim(path, "/")

	if deviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	// Check if device exists
	_, err := s.registry.Get(deviceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("device not found: %s", deviceID), http.StatusNotFound)
		return
	}

	// Upgrade connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("failed to upgrade connection",
			telemetry.String("device_id", deviceID),
			telemetry.Error(err),
		)
		return
	}

	// Create client
	client := NewClient(conn, deviceID, "payment", s.logger)

	// Register client
	s.hub.Register <- client

	// Start read/write pumps
	ctx := r.Context()
	go client.WritePump(ctx)
	go client.ReadPump(ctx, s.hub)

	s.logger.Info("payment websocket connected",
		telemetry.String("device_id", deviceID),
		telemetry.String("client_id", client.ID),
	)
}

// handleDevices handles device status WebSocket connections
func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	// This endpoint streams all device status changes
	deviceID := "*" // Special device ID for all devices

	// Upgrade connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("failed to upgrade connection", telemetry.Error(err))
		return
	}

	// Create client
	client := NewClient(conn, deviceID, "devices", s.logger)

	// Register client
	s.hub.Register <- client

	// Start read/write pumps
	ctx := r.Context()
	go client.WritePump(ctx)
	go client.ReadPump(ctx, s.hub)

	s.logger.Info("devices websocket connected",
		telemetry.String("client_id", client.ID),
	)
}

// GetConnectedClients returns the number of connected clients
func (s *Server) GetConnectedClients() int {
	return s.hub.GetClientCount()
}
