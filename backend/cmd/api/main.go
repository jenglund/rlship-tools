package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jenglund/rlship-tools/internal/api/handlers"
	"github.com/jenglund/rlship-tools/internal/api/service"
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
	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORS())

	// Debug endpoint - doesn't require auth
	debug := router.Group("/api/debug")
	{
		debug.GET("/routes", func(c *gin.Context) {
			var routes []map[string]string
			for _, r := range router.Routes() {
				routes = append(routes, map[string]string{
					"method": r.Method,
					"path":   r.Path,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"routes": routes,
				"total":  len(routes),
			})
		})

		debug.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
				"status":  "ok",
				"time":    time.Now().Format(time.RFC3339),
			})
		})
	}

	// API routes
	api := router.Group("/api")
	{
		// Add Auth middleware to all API routes
		api.Use(authMiddleware.AuthMiddleware())

		// Add middleware to extract Firebase UID and convert to user ID
		api.Use(func(c *gin.Context) {
			// Get Firebase UID from context
			firebaseUID, exists := c.Get("firebase_uid")
			if !exists {
				log.Printf("Authentication failure: Firebase UID not found in context")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
				return
			}
			log.Printf("Found Firebase UID: %v", firebaseUID)

			// For dev users, extract the user ID directly if it's already set
			if userID, userExists := c.Get("user_id"); userExists {
				log.Printf("User ID already set in context: %v", userID)
				c.Next()
				return
			}

			// Otherwise, look up the user by Firebase UID
			log.Printf("Looking up user by Firebase UID: %s", firebaseUID.(string))
			user, err := repos.Users.GetByFirebaseUID(firebaseUID.(string))
			if err != nil {
				log.Printf("Failed to find user for Firebase UID %s: %v", firebaseUID, err)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found: " + err.Error()})
				return
			}

			// Set user ID in context
			log.Printf("Found user ID %s for Firebase UID %s", user.ID, firebaseUID)
			c.Set("user_id", user.ID)
			c.Next()
		})

		// Initialize and register handlers
		userHandler := handlers.NewUserHandler(repos)
		userHandler.RegisterRoutes(api)

		tribeHandler := handlers.NewTribeHandler(repos)
		tribeHandler.RegisterRoutes(api)

		// Initialize and register v1 group
		v1 := api.Group("/v1")

		// Initialize list service using the repository
		listService := service.NewListService(repos.Lists)

		// Initialize the list handler
		listHandler := handlers.NewListHandler(listService)

		// Register list routes directly with Gin
		lists := v1.Group("/lists")

		// Basic list operations
		lists.POST("", wrapHandler(listHandler.CreateList))
		lists.GET("", wrapHandler(listHandler.ListLists))
		lists.GET("/:listID", wrapHandler(listHandler.GetList))
		lists.PUT("/:listID", wrapHandler(listHandler.UpdateList))
		lists.DELETE("/:listID", wrapHandler(listHandler.DeleteList))

		// List items
		lists.POST("/:listID/items", wrapHandler(listHandler.AddListItem))
		lists.GET("/:listID/items", wrapHandler(listHandler.GetListItems))
		lists.PUT("/:listID/items/:itemID", wrapHandler(listHandler.UpdateListItem))
		lists.DELETE("/:listID/items/:itemID", wrapHandler(listHandler.RemoveListItem))

		// User and tribe specific lists
		lists.GET("/user/:userID", wrapHandler(listHandler.GetUserLists))
		lists.GET("/tribe/:tribeID", wrapHandler(listHandler.GetTribeLists))

		// Sharing
		// Generic share endpoint (accepts tribe ID in the request body)
		lists.POST("/:listID/share-with-body", wrapHandler(listHandler.ShareList))
		lists.GET("/shared/:tribeID", wrapHandler(listHandler.GetSharedLists))
		lists.GET("/:listID/shares", wrapHandler(listHandler.GetListShares))
		// Specific share endpoint (with tribe ID in the URL)
		lists.POST("/:listID/share/:tribeID", wrapHandler(listHandler.ShareListWithTribe))
		lists.DELETE("/:listID/share/:tribeID", wrapHandler(listHandler.UnshareListWithTribe))
	}

	// Log all registered routes
	logRoutes(router)

	return router
}

// logRoutes prints all registered routes for debugging
func logRoutes(router *gin.Engine) {
	routes := router.Routes()
	log.Printf("Registered routes (%d total):", len(routes))
	for _, route := range routes {
		log.Printf("  %s %s", route.Method, route.Path)
	}
}

// wrapHandler converts a standard http handler to a Gin handler
func wrapHandler(h func(http.ResponseWriter, *http.Request)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from Gin context
		userID, exists := c.Get("user_id")
		if exists {
			log.Printf("Wrapping handler: Setting user ID %v in request context", userID)
			// Create new context with user ID
			ctx := context.WithValue(c.Request.Context(), "user_id", userID)
			// Use the new context in the request
			c.Request = c.Request.WithContext(ctx)
		} else {
			log.Printf("Wrapping handler: No user ID found in Gin context")
		}
		h(c.Writer, c.Request)
	}
}
