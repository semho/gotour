package http

import (
	"chat/pkg/logger"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "chat/pkg/chat/v1"
)

type Server struct {
	httpServer *http.Server
	grpcMux    *runtime.ServeMux
	httpMux    *http.ServeMux
	port       int
	grpcPort   int
}

func NewServer(port, grpcPort int) *Server {
	return &Server{
		grpcMux:  runtime.NewServeMux(),
		httpMux:  http.NewServeMux(),
		port:     port,
		grpcPort: grpcPort,
	}
}

func (s *Server) Start() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterChatServiceHandlerFromEndpoint(ctx, s.grpcMux, fmt.Sprintf("localhost:%d", s.grpcPort), opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	// Обработчик gRPC-шлюза
	s.httpMux.Handle("/", s.grpcMux)

	// Swagger JSON
	s.httpMux.HandleFunc("/swagger.json", s.serveSwaggerJSON)

	// Swagger UI
	s.httpMux.HandleFunc("/swagger/", s.serveSwaggerUI)

	// health check endpoint
	s.httpMux.HandleFunc("/health", s.healthCheckHandler)

	// metrics endpoint
	s.httpMux.Handle("/metrics", promhttp.Handler())

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.httpMux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Printf("Starting HTTP server on port %d\n", s.port)
	fmt.Printf("Swagger UI available at http://localhost:%d/swagger/\n", s.port)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *Server) Stop() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			logger.Log.Error("Error shutting down server", "error", err)
		}
	}
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		logger.Log.Error("Error writing response", "error", err)
	}
}

func (s *Server) serveSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	swaggerPath := filepath.Join("pkg", "chat", "v1", "chat.swagger.json")
	http.ServeFile(w, r, swaggerPath)
}

func (s *Server) serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	swaggerTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.52.0/swagger-ui.css" >
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.52.0/swagger-ui-bundle.js"> </script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.52.0/swagger-ui-standalone-preset.js"> </script>
</head>
<body>
<div id="swagger-ui"></div>
<script>
    window.onload = function() {
        SwaggerUIBundle({
            url: "/swagger.json",
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            layout: "BaseLayout"
        })
    }
</script>
</body>
</html>
`
	tmpl, err := template.New("swagger").Parse(swaggerTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
