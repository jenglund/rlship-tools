package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jenglund/rlship-tools/internal/config"
)

// Migrator defines the interface for database migrations
type Migrator interface {
	Up() error
	Down() error
	Force(version int) error
	Version() (uint, bool, error)
}

// MigrateFactory is a function type that creates a new migrate instance
type MigrateFactory func(sourceURL, databaseURL string) (Migrator, error)

// ConfigLoader is a function type that loads configuration
type ConfigLoader func() (*config.Config, error)

// defaultMigrateFactory is the default implementation that uses migrate.New
func defaultMigrateFactory(sourceURL, databaseURL string) (Migrator, error) {
	return migrate.New(sourceURL, databaseURL)
}

// defaultConfigLoader is the default implementation that uses config.Load
func defaultConfigLoader() (*config.Config, error) {
	return config.Load()
}

// runMigrations handles the migration logic
func runMigrations(args []string, factory MigrateFactory, configLoader ConfigLoader) error {
	if len(args) < 2 {
		return fmt.Errorf("Command required: up, down, or force")
	}

	command := args[1]
	// Validate command before doing anything else
	switch command {
	case "up", "down":
		// Valid commands, continue
	case "force":
		if len(args) != 3 {
			return fmt.Errorf("Version number required for force command")
		}
		if _, err := strconv.ParseUint(args[2], 10, 64); err != nil {
			return fmt.Errorf("Invalid version number: %v", err)
		}
	default:
		return fmt.Errorf("Invalid command. Use 'up', 'down', or 'force'")
	}

	if configLoader == nil {
		configLoader = defaultConfigLoader
	}

	cfg, err := configLoader()
	if err != nil {
		return fmt.Errorf("Error loading config: %v", err)
	}

	if factory == nil {
		factory = defaultMigrateFactory
	}

	m, err := factory("file://migrations", cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("Error creating migrate instance: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("Error getting version: %v", err)
	}
	log.Printf("Current migration version: %d, Dirty: %v", version, dirty)

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("Error running migrations: %v", err)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("Error running migrations: %v", err)
		}
	case "force":
		version, _ := strconv.ParseUint(args[2], 10, 64) // Already validated above
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("Error forcing version: %v", err)
		}
		log.Printf("Successfully forced version to %d", version)
	}

	return nil
}

func main() {
	if err := runMigrations(os.Args, defaultMigrateFactory, defaultConfigLoader); err != nil {
		log.Fatal(err)
	}
}
