package rest

import (
	"fmt"
	"net/http"
	"os"
)

// SwaggerHandler serves the Swagger UI and OpenAPI spec
type SwaggerHandler struct {
	specPath string
}

// NewSwaggerHandler creates a new Swagger UI handler
func NewSwaggerHandler(specPath string) *SwaggerHandler {
	return &SwaggerHandler{
		specPath: specPath,
	}
}

// ServeHTTP serves the Swagger UI or OpenAPI spec
func (h *SwaggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/swagger/", "/swagger":
		// Serve Swagger UI HTML
		h.serveSwaggerUI(w, r)
	case "/swagger/devicebridge.swagger.json":
		// Serve OpenAPI spec
		h.serveSpec(w, r)
	default:
		http.NotFound(w, r)
	}
}

// serveSpec serves the OpenAPI specification file
func (h *SwaggerHandler) serveSpec(w http.ResponseWriter, r *http.Request) {
	// Read the swagger.json file
	data, err := os.ReadFile(h.specPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read spec file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(data)
}

// serveSwaggerUI serves a simple Swagger UI HTML page
func (h *SwaggerHandler) serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Device Bridge API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        body {
            margin: 0;
            padding: 0;
        }
        .topbar {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/devicebridge.swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
            window.ui = ui;
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// AddSwaggerRoutes adds Swagger UI routes to an HTTP mux
func AddSwaggerRoutes(mux *http.ServeMux, specPath string) {
	handler := NewSwaggerHandler(specPath)
	mux.HandleFunc("/swagger/", handler.ServeHTTP)
	mux.HandleFunc("/swagger/devicebridge.swagger.json", handler.ServeHTTP)
}
