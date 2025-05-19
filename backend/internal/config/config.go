package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Firebase FirebaseConfig `mapstructure:"firebase"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	URL      string `mapstructure:"url"`
	SSLMode  string `mapstructure:"sslmode"`
}

type FirebaseConfig struct {
	ProjectID       string `mapstructure:"project_id"`
	CredentialsFile string `mapstructure:"credentials_file"`
}

type AuthConfig struct {
	FirebaseProjectID string `mapstructure:"firebase_project_id"`
}

// Load reads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	// Map environment variables
	viper.SetEnvPrefix("")
	if err := viper.BindEnv("database.host", "DB_HOST"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("database.port", "DB_PORT"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("database.user", "DB_USER"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("database.password", "DB_PASSWORD"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("database.name", "DB_NAME"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("database.sslmode", "DB_SSLMODE"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("server.host", "SERVER_HOST"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("server.port", "SERVER_PORT"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("firebase.project_id", "FIREBASE_PROJECT_ID"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := viper.BindEnv("firebase.credentials_file", "FIREBASE_CREDENTIALS_FILE"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}

	// Default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("database.sslmode", "disable")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Check for development mode
	isDevelopment := os.Getenv("ENVIRONMENT") == "development"

	// Validate required fields
	if config.Database.Host == "" {
		return nil, fmt.Errorf("database host is required")
	}
	if config.Database.User == "" {
		return nil, fmt.Errorf("database user is required")
	}
	if config.Database.Name == "" {
		return nil, fmt.Errorf("database name is required")
	}

	// Only validate Firebase configuration in non-development mode
	if !isDevelopment {
		if config.Firebase.ProjectID == "" {
			return nil, fmt.Errorf("firebase project ID is required")
		}
		if config.Firebase.CredentialsFile == "" {
			return nil, fmt.Errorf("firebase credentials file is required")
		}
	} else {
		// Set default values for development mode
		if config.Firebase.ProjectID == "" {
			config.Firebase.ProjectID = "dev-project"
		}
		if config.Firebase.CredentialsFile == "" {
			config.Firebase.CredentialsFile = "/tmp/firebase-credentials.json"
		}
	}

	// Override with environment variables if present
	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = viper.GetInt("PORT")
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		// Parse DATABASE_URL if present
		// For now, just log that we don't support this yet
		fmt.Println("WARNING: DATABASE_URL environment variable found but not implemented")
	}
	if projectID := os.Getenv("FIREBASE_PROJECT_ID"); projectID != "" {
		config.Firebase.ProjectID = projectID
	}
	if credsFile := os.Getenv("FIREBASE_CREDENTIALS_FILE"); credsFile != "" {
		config.Firebase.CredentialsFile = credsFile
	}

	// Construct database URL
	config.Database.URL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name)

	return &config, nil
}
