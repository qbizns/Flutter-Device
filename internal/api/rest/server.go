package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// Server is the REST/JSON gateway server
type Server struct {
	grpcAddress string
	httpPort    int
	mux         *runtime.ServeMux
	httpServer  *http.Server
}

// Config holds REST server configuration
type Config struct {
	GRPCAddress string // Address of the gRPC server (e.g., "localhost:50051")
	HTTPPort    int    // Port for HTTP server (e.g., 8080)
}

// NewServer creates a new REST gateway server
func NewServer(cfg Config) *Server {
	// Create gateway mux with options
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: runtime.JSONPb{
				MarshalOptions: runtime.JSONPb{}.MarshalOptions,
			}.MarshalOptions,
		}),
	)

	return &Server{
		grpcAddress: cfg.GRPCAddress,
		httpPort:    cfg.HTTPPort,
		mux:         mux,
	}
}

// Start starts the REST gateway server
func (s *Server) Start(ctx context.Context) error {
	// Create gRPC client connection to local gRPC server
	conn, err := grpc.NewClient(
		s.grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	// Register DeviceBridge service
	if err := pb.RegisterDeviceBridgeHandler(ctx, s.mux, conn); err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Create HTTP server with CORS middleware
	handler := corsMiddleware(s.mux)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpPort),
		Handler: handler,
	}

	// Start HTTP server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("REST server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the REST gateway server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// corsMiddleware adds CORS headers to allow browser access
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}
