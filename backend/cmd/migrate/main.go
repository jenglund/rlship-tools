package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

// migrator is a wrapper for migrate.Migrate that implements Migrator
type migrator struct {
	*migrate.Migrate
}

// MigrateFactory is a function type that creates a new migrate instance
type MigrateFactory func(sourceURL, databaseURL string) (Migrator, error)

// ConfigLoader is a function type that loads configuration
type ConfigLoader func() (*config.Config, error)

// defaultMigrateFactory is the default implementation that uses migrate.New
func defaultMigrateFactory(sourceURL, databaseURL string) (Migrator, error) {
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return nil, err
	}
	return &migrator{Migrate: m}, nil
}

// defaultConfigLoader is the default implementation that uses config.Load
func defaultConfigLoader() (*config.Config, error) {
	return config.Load()
}

// findMigrationsPath finds the migrations directory by trying several possible locations
func findMigrationsPath() (string, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory: %v", err)
	}
	log.Printf("Current working directory: %s", wd)

	// Try various possible locations
	possiblePaths := []string{
		filepath.Join(wd, "migrations"),         // If we're in backend/
		filepath.Join(wd, "../../migrations"),   // If we're in backend/cmd/migrate/
		filepath.Join(wd, "../migrations"),      // If we're in backend/cmd/
		filepath.Join(wd, "backend/migrations"), // If we're in project root
	}

	for _, path := range possiblePaths {
		// Check if directory exists
		if _, err := os.Stat(path); err == nil {
			// Found the directory
			log.Printf("Found migrations at: %s", path)
			return fmt.Sprintf("file://%s", path), nil
		}
	}

	return "", fmt.Errorf("could not find migrations directory in any expected location")
}

// runMigrations handles the migration logic
func runMigrations(args []string, factory MigrateFactory, configLoader ConfigLoader) error {
	if len(args) < 2 {
		return fmt.Errorf("command required: up, down, or force")
	}

	command := args[1]
	// Validate command before doing anything else
	switch command {
	case "up", "down":
		// Valid commands, continue
	case "force":
		if len(args) != 3 {
			return fmt.Errorf("version number required for force command")
		}
		if _, err := strconv.ParseUint(args[2], 10, 64); err != nil {
			return fmt.Errorf("invalid version number: %v", err)
		}
	default:
		return fmt.Errorf("invalid command; use 'up', 'down', or 'force'")
	}

	if configLoader == nil {
		configLoader = defaultConfigLoader
	}

	cfg, err := configLoader()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	if factory == nil {
		factory = defaultMigrateFactory
	}

	// Find the migrations path
	migrationsPath, err := findMigrationsPath()
	if err != nil {
		return fmt.Errorf("error finding migrations path: %v", err)
	}
	log.Printf("Using migrations path: %s", migrationsPath)

	m, err := factory(migrationsPath, cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("error getting version: %v", err)
	}
	log.Printf("Current migration version: %d, Dirty: %v", version, dirty)

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error running migrations: %v", err)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error running migrations: %v", err)
		}
	case "force":
		version, _ := strconv.ParseUint(args[2], 10, 64) // Already validated above
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("error forcing version: %v", err)
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
