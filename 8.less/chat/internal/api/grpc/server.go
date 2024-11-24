package grpc

import (
	"chat/internal/middleware"
	"chat/internal/service"
	pb "chat/pkg/chat/v1"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	chatService *service.ChatService
	grpcServer  *grpc.Server
	port        int
}

func NewServer(chatService *service.ChatService, port int) *Server {
	return &Server{
		chatService: chatService,
		port:        port,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				middleware.AuthInterceptor,
				middleware.MetricsInterceptor,
			),
		),
	)
	pb.RegisterChatServiceServer(s.grpcServer, s.chatService)

	reflection.Register(s.grpcServer)

	fmt.Printf("Starting gRPC server on port %d\n", s.port)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

func (s *Server) Close() error {
	if s.chatService != nil {
		return s.chatService.Close()
	}
	return nil
}
