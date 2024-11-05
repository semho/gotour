package main

import (
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func main() {
	var dsn string
	var migrationsPath string
	var down bool

	flag.StringVar(&dsn, "dsn", "", "Database connection string")
	flag.StringVar(&migrationsPath, "migrations", "migrations", "Path to migrations directory")
	flag.BoolVar(&down, "down", false, "Run downgrade migrations")
	flag.Parse()

	if dsn == "" {
		log.Fatal("Database connection string is required")
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	if down {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run down migrations: %v", err)
		}
		log.Println("Successfully ran down migrations")
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run up migrations: %v", err)
	}
	log.Println("Successfully ran up migrations")
}
