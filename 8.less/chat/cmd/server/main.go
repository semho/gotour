package main

import (
	"chat/internal/storage"
	"log"
	"os"
	"os/signal"
	"syscall"

	"chat/internal/api/grpc"
	"chat/internal/api/http"
	"chat/internal/config"
	"chat/internal/service"
	"chat/internal/storage/memory"
	// TODO: import redis and postgres storage when implemented
)

func main() {
	cfg := config.NewConfig()

	var storage storage.Storage
	switch cfg.StorageType {
	case "memory":
		storage = memory.NewMemoryStorage(cfg.MaxChatSize, cfg.MaxChatsCount)
	case "redis":
		// TODO: Implement Redis storage
		log.Fatal("Redis storage not implemented yet")
	case "postgres":
		// TODO: Implement Postgres storage
		log.Fatal("Postgres storage not implemented yet")
	default:
		log.Fatalf("Unknown storage type: %s", cfg.StorageType)
	}

	chatService := service.NewChatService(storage)
	grpcServer := grpc.NewServer(chatService, cfg.GRPCPort)
	httpServer := http.NewServer(cfg.HTTPPort, cfg.GRPCPort)

	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	grpcServer.Stop()
	httpServer.Stop()
	log.Println("Servers stopped")
}
