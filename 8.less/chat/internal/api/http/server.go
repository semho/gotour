package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "chat/pkg/chat/v1"
)

type Server struct {
	httpServer *http.Server
	mux        *runtime.ServeMux
	port       int
	grpcPort   int
}

func NewServer(port, grpcPort int) *Server {
	return &Server{
		mux:      runtime.NewServeMux(),
		port:     port,
		grpcPort: grpcPort,
	}
}

func (s *Server) Start() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterChatServiceHandlerFromEndpoint(ctx, s.mux, fmt.Sprintf("localhost:%d", s.grpcPort), opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	// health check endpoint
	s.mux.HandlePath("GET", "/healthz", s.healthCheckHandler)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.mux,
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

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
