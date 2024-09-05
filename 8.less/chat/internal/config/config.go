package config

import (
	"github.com/spf13/pflag"
	"time"
)

type Config struct {
	GRPCPort      int
	HTTPPort      int
	MaxChatSize   int
	MaxChatsCount int
	StorageType   string
	StorageDSN    string
	DefaultTTL    time.Duration
}

func NewConfig() *Config {
	cfg := &Config{}

	pflag.IntVar(&cfg.GRPCPort, "grpc-port", 50051, "gRPC server port")
	pflag.IntVar(&cfg.HTTPPort, "http-port", 8080, "HTTP server port")
	pflag.IntVar(&cfg.MaxChatSize, "max-chat-size", 1000, "Maximum number of messages per chat")
	pflag.IntVar(&cfg.MaxChatsCount, "max-chats-count", 1000, "Maximum number of chats")
	pflag.StringVar(&cfg.StorageType, "storage", "memory", "Storage type (memory, redis, postgres)")
	pflag.StringVar(&cfg.StorageDSN, "storage-dsn", "", "Storage DSN (for redis or postgres)")
	pflag.DurationVar(&cfg.DefaultTTL, "default-ttl", 24*time.Hour, "Default TTL for chats")

	pflag.Parse()

	return cfg
}
