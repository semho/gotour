package main

import (
	"chat/internal/storage"
	"chat/pkg/logger"
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
	logger.Init()

	cfg := config.NewConfig()

	var storage storage.Storage
	switch cfg.StorageType {
	case "memory":
		storage = memory.NewMemoryStorage(cfg.MaxChatSize, cfg.MaxChatsCount)
	case "redis":
		// TODO: Implement Redis storage
		logger.Log.Error("Redis storage not implemented yet")
		os.Exit(1)
	case "postgres":
		// TODO: Implement Postgres storage
		logger.Log.Error("Postgres storage not implemented yet")
		os.Exit(1)
	default:
		logger.Log.Error("Unknown storage type", "type", cfg.StorageType)
		os.Exit(1)
	}

	chatService := service.NewChatService(storage)
	grpcServer := grpc.NewServer(chatService, cfg.GRPCPort)
	httpServer := http.NewServer(cfg.HTTPPort, cfg.GRPCPort)

	logger.Log.Info("Starting servers")

	go func() {
		if err := grpcServer.Start(); err != nil {
			logger.Log.Error("Failed to start gRPC server", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Log.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down servers")

	grpcServer.Stop()
	httpServer.Stop()
	logger.Log.Info("Servers stopped")
}
