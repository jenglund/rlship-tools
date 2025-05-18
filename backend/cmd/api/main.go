package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jenglund/rlship-tools/internal/api/handlers"
	"github.com/jenglund/rlship-tools/internal/config"
	"github.com/jenglund/rlship-tools/internal/middleware"
	"github.com/jenglund/rlship-tools/internal/repository/postgres"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Setup application
	router, err := setupApp(cfg)
	if err != nil {
		log.Fatalf("Error setting up application: %v", err)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// setupApp configures and initializes the application
func setupApp(cfg *config.Config) (*gin.Engine, error) {
	// Initialize database connection
	port := 5432 // Default port
	if cfg.Database.Port != "" {
		if p, err := strconv.Atoi(cfg.Database.Port); err == nil {
			port = p
		}
	}

	db, err := postgres.NewDB(
		cfg.Database.Host,
		port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Initialize repositories
	repos := postgres.NewRepositories(db)

	// Check for development mode
	environment := os.Getenv("ENVIRONMENT")
	isDevelopment := environment == "development"

	// Check if Firebase credentials file exists
	_, err = os.Stat(cfg.Firebase.CredentialsFile)
	credentialsFileExists := err == nil

	// Initialize Authentication middleware
	var authMiddleware middleware.AuthMiddleware

	if isDevelopment || !credentialsFileExists {
		if !credentialsFileExists {
			log.Printf("Warning: Firebase credentials file not found at %s", cfg.Firebase.CredentialsFile)
		}
		log.Printf("Using development authentication mode with dev user email pattern")
		devAuth := middleware.NewDevFirebaseAuth()
		devAuth.SetRepositoryProvider(repos)
		authMiddleware = devAuth
	} else {
		// Initialize Firebase Auth for production
		firebaseAuth, err := middleware.NewFirebaseAuth(cfg.Firebase.CredentialsFile)
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				log.Printf("Error closing database during Firebase Auth setup error: %v", closeErr)
			}
			return nil, fmt.Errorf("error initializing Firebase Auth: %w", err)
		}
		authMiddleware = firebaseAuth
	}

	// Initialize and configure Gin router
	router := setupRouter(repos, authMiddleware)

	return router, nil
}

// setupRouter creates and configures the Gin router with all routes and middlewares
func setupRouter(repos *postgres.Repositories, authMiddleware middleware.AuthMiddleware) *gin.Engine {
	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORS())

	// API routes
	api := router.Group("/api")
	{
		// Add Auth middleware to all API routes
		api.Use(authMiddleware.AuthMiddleware())

		// Initialize and register handlers
		userHandler := handlers.NewUserHandler(repos)
		userHandler.RegisterRoutes(api)

		tribeHandler := handlers.NewTribeHandler(repos)
		tribeHandler.RegisterRoutes(api)
	}

	return router
}
