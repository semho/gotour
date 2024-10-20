package main

import (
	"log"

	"messenger/internal/app"
	"messenger/internal/config"
)

func main() {
	log.Println("Starting messenger application")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	log.Println("Application initialized successfully")

	if err := application.Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}
