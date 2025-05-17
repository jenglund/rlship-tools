package main

import (
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jenglund/rlship-tools/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Command required: up, down, or force")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	m, err := migrate.New(
		"file://migrations",
		cfg.Database.URL,
	)
	if err != nil {
		log.Fatalf("Error creating migrate instance: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatalf("Error getting version: %v", err)
	}
	log.Printf("Current migration version: %d, Dirty: %v", version, dirty)

	command := os.Args[1]
	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error running migrations: %v", err)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error running migrations: %v", err)
		}
	case "force":
		if len(os.Args) != 3 {
			log.Fatal("Version number required for force command")
		}
		version, err := strconv.ParseUint(os.Args[2], 10, 64)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := m.Force(int(version)); err != nil {
			log.Fatalf("Error forcing version: %v", err)
		}
		log.Printf("Successfully forced version to %d", version)
	default:
		log.Fatalf("Invalid command. Use 'up', 'down', or 'force'")
	}
}
