package http

import (
	"chat/internal/api/docs"
	"chat/internal/api/health"
	"chat/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"google.golang.org/grpc/metadata"

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
		grpcMux:  runtime.NewServeMux(runtime.WithMetadata(metadataAnnotator)),
		httpMux:  http.NewServeMux(),
		port:     port,
		grpcPort: grpcPort,
	}
}

func metadataAnnotator(_ context.Context, req *http.Request) metadata.MD {
	md := metadata.New(nil)

	if sessionID := req.Header.Get("Session-Id"); sessionID != "" {
		md.Set("session_id", sessionID)
		logger.Log.Info("Session ID set in metadata", "session_id", sessionID)
	} else {
		logger.Log.Warn("Session ID not found in request headers")
	}

	for key, values := range req.Header {
		// Игнорируем заголовок "Connection", почему-то с ним не работает, пишет "ошибка протокола"
		if strings.ToLower(key) != "connection" {
			for _, value := range values {
				md.Append(strings.ToLower(key), value)
			}
		}
	}

	return md
}

func (s *Server) Start() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10 * 1024 * 1024)),
	}
	err := pb.RegisterChatServiceHandlerFromEndpoint(ctx, s.grpcMux, fmt.Sprintf("localhost:%d", s.grpcPort), opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	// Обработчик gRPC-шлюза
	s.httpMux.Handle("/", s.grpcMux)
	// Swagger JSON
	s.httpMux.HandleFunc("/swagger.json", docs.ServeSwaggerJSON)
	// Swagger UI
	s.httpMux.HandleFunc("/swagger/", docs.ServeSwaggerUI)
	// health check endpoint
	s.httpMux.HandleFunc("/health", health.CheckHandler)
	// metrics endpoint
	s.httpMux.Handle("/metrics", promhttp.Handler())
	// CORS middleware
	corsMiddleware := cors.New(
		cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{
				"Accept",
				"Content-Type",
				"Content-Length",
				"Accept-Encoding",
				"Authorization",
				"Session-Id",
			},
		},
	)

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           corsMiddleware.Handler(logger.LogRequest(s.httpMux)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Printf("Starting HTTP server on port %d\n", s.port)
	fmt.Printf("Swagger UI available at http://localhost:%d/swagger/\n", s.port)
	if err = s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
