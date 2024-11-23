package main

import (
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"path/filepath"
)

func main() {
	var dsn string
	var migrationsPath string
	var down bool
	var status bool

	flag.StringVar(&dsn, "dsn", "", "Database connection string")
	flag.StringVar(&migrationsPath, "migrations", "migrations", "Path to migrations directory")
	flag.BoolVar(&down, "down", false, "Run downgrade migrations")
	flag.BoolVar(&status, "status", false, "Show migrations status")
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

	if status {
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				log.Println("No migrations have been applied yet")
				return
			}
			log.Fatalf("Failed to get migration version: %v", err)
		}

		fmt.Printf("Current migration version: %d\n", version)
		if dirty {
			fmt.Println("WARNING: Migration state is dirty!")
		}

		fmt.Println("\nAvailable migration files:")
		files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
		if err != nil {
			log.Printf("Failed to read migration files: %v", err)
		} else {
			for _, file := range files {
				fmt.Printf("- %s\n", filepath.Base(file))
			}
		}
		return
	}

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
