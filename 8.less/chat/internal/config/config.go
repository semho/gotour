package config

import (
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"
)

type Config struct {
	GRPCPort      int
	HTTPPort      int
	MaxChatSize   int
	MaxChatsCount int
	StorageType   string
	StorageDSN    string
	DefaultTTL    time.Duration
	KafkaBrokers  string
	KafkaTopic    string
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func NewConfig() *Config {
	cfg := &Config{}

	pflag.IntVar(
		&cfg.GRPCPort, "grpc-port",
		getEnvIntOrDefault("GRPC_PORT", 50051),
		"gRPC server port",
	)

	pflag.IntVar(
		&cfg.HTTPPort, "http-port",
		getEnvIntOrDefault("HTTP_PORT", 8080),
		"HTTP server port",
	)

	pflag.IntVar(
		&cfg.MaxChatSize, "max-chat-size",
		getEnvIntOrDefault("MAX_CHAT_SIZE", 1000),
		"Maximum number of messages per chat",
	)

	pflag.IntVar(
		&cfg.MaxChatsCount, "max-chats-count",
		getEnvIntOrDefault("MAX_CHATS_COUNT", 1000),
		"Maximum number of chats",
	)

	pflag.StringVar(
		&cfg.StorageType, "storage",
		getEnvOrDefault("STORAGE", "memory"),
		"Storage type (memory, redis, postgres)",
	)

	pflag.StringVar(
		&cfg.StorageDSN, "storage-dsn",
		getEnvOrDefault("STORAGE_DSN", ""),
		"Storage DSN (for redis or postgres)",
	)

	pflag.StringVar(
		&cfg.KafkaBrokers, "kafka-brokers",
		getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"),
		"Kafka brokers list (comma-separated)",
	)

	pflag.StringVar(
		&cfg.KafkaTopic, "kafka-topic",
		getEnvOrDefault("KAFKA_TOPIC", "chat.messages"),
		"Kafka topic for chat messages",
	)

	defaultTTL := 24 * time.Hour
	if ttlStr := os.Getenv("DEFAULT_TTL"); ttlStr != "" {
		if parsed, err := time.ParseDuration(ttlStr); err == nil {
			defaultTTL = parsed
		}
	}
	pflag.DurationVar(&cfg.DefaultTTL, "default-ttl", defaultTTL, "Default TTL for chats")

	pflag.Parse()
	return cfg
}
