package main

import (
	"context"
	"database/sql"
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
	"github.com/jenglund/rlship-tools/internal/worker"
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

	// Initialize services
	listService := service.NewListService(repos.Lists)

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
	router := setupRouter(repos, authMiddleware, listService)

	// Initialize and start background workers
	// Share cleanup worker runs every hour
	cleanupWorker := worker.NewShareCleanupWorker(listService, 1*time.Hour)
	cleanupWorker.Start()
	log.Println("Share cleanup worker started")

	// Start a background goroutine to monitor database health
	go monitorDatabaseHealth(repos.DB())

	return router, nil
}

// monitorDatabaseHealth periodically checks database health and attempts reconnection if needed
func monitorDatabaseHealth(db *sql.DB) {
	log.Println("Starting database health monitoring...")
	ticker := time.NewTicker(15 * time.Second) // Check more frequently (every 15 seconds)
	defer ticker.Stop()

	consecutiveFailures := 0
	maxConsecutiveFailures := 3

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := db.PingContext(ctx)
		cancel()

		if err != nil {
			consecutiveFailures++
			log.Printf("Database connection unhealthy (%d/%d): %v",
				consecutiveFailures, maxConsecutiveFailures, err)

			// Get connection pool stats
			stats := db.Stats()
			log.Printf("Connection pool stats: open=%d, in-use=%d, idle=%d, max-open=%d",
				stats.OpenConnections, stats.InUse, stats.Idle,
				stats.MaxOpenConnections)

			// If we've had multiple consecutive failures, perform more aggressive recovery
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Printf("Performing aggressive connection recovery after %d consecutive failures",
					consecutiveFailures)

				// Force close idle connections
				db.SetMaxIdleConns(0)
				time.Sleep(500 * time.Millisecond)
				db.SetMaxIdleConns(10)

				// Wait a moment and try to reconnect
				time.Sleep(1 * time.Second)

				// Try to reconnect with a fresh context
				reconnectCtx, reconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
				reconnectErr := db.PingContext(reconnectCtx)
				reconnectCancel()

				if reconnectErr != nil {
					log.Printf("Failed to reconnect to database: %v", reconnectErr)
				} else {
					log.Println("Successfully reconnected to database after aggressive recovery")
					consecutiveFailures = 0
				}
			}
		} else {
			// Reset counter on successful ping
			if consecutiveFailures > 0 {
				log.Println("Database connection restored")
				consecutiveFailures = 0
			}
		}
	}
}

// setupRouter creates and configures the Gin router with all routes and middlewares
func setupRouter(repos *postgres.Repositories, authMiddleware middleware.AuthMiddleware, listService service.ListService) *gin.Engine {
	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORS())

	// Add database health check middleware
	dbHealthChecker := middleware.NewDatabaseHealthChecker(repos.DB(), 30*time.Second)
	router.Use(dbHealthChecker.Middleware())

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

		debug.GET("/db-health", func(c *gin.Context) {
			// Test database connection
			err := repos.DB().Ping()

			// Check connection pool stats
			stats := repos.DB().Stats()

			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Database connection error: %v", err),
					"time":    time.Now().Format(time.RFC3339),
					"pool_stats": gin.H{
						"open_connections":     stats.OpenConnections,
						"in_use":               stats.InUse,
						"idle":                 stats.Idle,
						"max_open_connections": stats.MaxOpenConnections,
					},
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"message": "Database connection healthy",
				"time":    time.Now().Format(time.RFC3339),
				"pool_stats": gin.H{
					"open_connections":     stats.OpenConnections,
					"in_use":               stats.InUse,
					"idle":                 stats.Idle,
					"max_open_connections": stats.MaxOpenConnections,
				},
			})
		})
	}

	// API routes
	api := router.Group("/api")

	// Create a public API group that doesn't require authentication
	publicAPI := api.Group("")
	{
		// Initialize and register user handler for public routes
		userHandler := handlers.NewUserHandler(repos)
		userHandler.RegisterRoutes(publicAPI)
	}

	// Protected API routes
	protectedAPI := api.Group("")
	{
		// Add Auth middleware to protected API routes
		protectedAPI.Use(authMiddleware.AuthMiddleware())

		// Add middleware to extract Firebase UID and convert to user ID
		protectedAPI.Use(func(c *gin.Context) {
			// Get Firebase UID from context
			firebaseUID, exists := c.Get(string(middleware.ContextFirebaseUIDKey))
			if !exists {
				log.Printf("Authentication failure: Firebase UID not found in context")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
				return
			}
			log.Printf("Found Firebase UID: %v", firebaseUID)

			// For dev users, extract the user ID directly if it's already set
			if userID, userExists := c.Get(string(middleware.ContextUserIDKey)); userExists {
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
			c.Set(string(middleware.ContextUserIDKey), user.ID)
			c.Next()
		})

		// Initialize and register tribe handler
		tribeHandler := handlers.NewTribeHandler(repos)
		tribeHandler.RegisterRoutes(protectedAPI)

		// Initialize and register v1 group
		v1 := protectedAPI.Group("/v1")

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

		// Admin endpoints
		admin := lists.Group("/admin")
		admin.POST("/cleanup-expired-shares", wrapHandler(listHandler.CleanupExpiredShares))
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
		userID, exists := c.Get(string(middleware.ContextUserIDKey))
		if exists {
			log.Printf("Wrapping handler: Setting user ID %v in request context", userID)
			// Create new context with user ID
			ctx := context.WithValue(c.Request.Context(), middleware.ContextUserIDKey, userID)
			// Use the new context in the request
			c.Request = c.Request.WithContext(ctx)
		} else {
			log.Printf("Wrapping handler: No user ID found in Gin context")
		}
		h(c.Writer, c.Request)
	}
}
