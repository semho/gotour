package http

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"

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

	// health check endpoint
	s.httpMux.HandleFunc("/health", s.healthCheckHandler)

	// metrics endpoint
	s.httpMux.Handle("/metrics", promhttp.Handler())

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.httpMux,
	}

	fmt.Printf("Starting HTTP server on port %d\n", s.port)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *Server) Stop() {
	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
	}
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
