package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using defaults and environment variables")
		} else {
			log.Fatalf("Error reading config file: %s", err)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Command required: up or down")
	}

	command := os.Args[1]

	dbHost := viper.GetString("database.host")
	dbPort := viper.GetString("database.port")
	dbName := viper.GetString("database.name")
	dbUser := viper.GetString("database.user")
	dbPass := viper.GetString("database.password")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Fatal("Error creating migrate instance:", err)
	}
	defer m.Close()

	// Get current version before migration
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatal("Error getting current version:", err)
	}
	log.Printf("Current migration version: %d, Dirty: %v", version, dirty)

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Error running migrations:", err)
		}
		newVersion, _, _ := m.Version()
		if newVersion > version {
			log.Printf("Successfully migrated from version %d to %d", version, newVersion)
		} else {
			log.Println("No new migrations to apply")
		}

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Error rolling back migrations:", err)
		}
		newVersion, _, _ := m.Version()
		if newVersion < version {
			log.Printf("Successfully rolled back from version %d to %d", version, newVersion)
		} else {
			log.Println("No migrations to roll back")
		}

	default:
		log.Fatal("Invalid command. Use 'up' or 'down'")
	}
}
