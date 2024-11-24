package main

import (
	"chat/internal/api/worker"
	"chat/pkg/logger"
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	logger.Init()

	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	dbDSN := os.Getenv("DB_DSN")

	work, err := worker.NewMessageWorker(kafkaBrokers, kafkaGroupID, kafkaTopic, dbDSN)
	if err != nil {
		logger.Log.Error("Failed to create worker", "error", err)
		os.Exit(1)
	}
	defer work.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := work.Start(ctx); err != nil {
			logger.Log.Error("Worker failed", "error", err)
			cancel()
		}
	}()

	<-sigChan
	logger.Log.Info("Shutting down worker...")
}
