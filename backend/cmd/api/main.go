package main

import (
	"fmt"
	"log"

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

	// Initialize database connection
	db, err := postgres.NewDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	repos := postgres.NewRepositories(db)

	// Initialize Firebase Auth
	firebaseAuth, err := middleware.NewFirebaseAuth(cfg.Firebase.CredentialsFile)
	if err != nil {
		log.Fatalf("Error initializing Firebase Auth: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORS())

	// API routes
	api := router.Group("/api")
	{
		// Add Firebase Auth middleware to all API routes
		api.Use(firebaseAuth.AuthMiddleware())

		// Initialize and register handlers
		userHandler := handlers.NewUserHandler(repos)
		userHandler.RegisterRoutes(api)

		tribeHandler := handlers.NewTribeHandler(repos)
		tribeHandler.RegisterRoutes(api)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
