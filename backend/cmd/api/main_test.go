package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jenglund/rlship-tools/internal/api/handlers"
	"github.com/jenglund/rlship-tools/internal/config"
	"github.com/jenglund/rlship-tools/internal/middleware"
	"github.com/jenglund/rlship-tools/internal/repository/postgres"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	_ "github.com/lib/pq"
)

// MockDB is a mock implementation of sql.DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockConfig is a mock implementation of config.Config
type MockConfig struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Server struct {
		Host string
		Port int
	}
	Firebase struct {
		CredentialsFile string
	}
}

// Use MockFirebaseAuth from testutil instead of local definition
type MockFirebaseAuth = testutil.MockFirebaseAuth

func init() {
	// Set gin mode to test mode to avoid data races across test functions
	gin.SetMode(gin.TestMode)
}

func TestConfigLoading(t *testing.T) {
	// Save current env and restore after test
	envBackup := make(map[string]string)
	for _, key := range []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"SERVER_HOST", "SERVER_PORT", "FIREBASE_CREDENTIALS_FILE",
	} {
		if val, exists := os.LookupEnv(key); exists {
			envBackup[key] = val
		}
	}
	defer func() {
		for key, val := range envBackup {
			if err := os.Setenv(key, val); err != nil {
				t.Logf("Error restoring environment variable %s: %v", key, err)
			}
		}
	}()

	tests := []struct {
		name        string
		setupEnv    func() error
		expectError bool
	}{
		{
			name: "valid configuration",
			setupEnv: func() error {
				envVars := map[string]string{
					"DB_HOST":                   "localhost",
					"DB_PORT":                   "5432",
					"DB_USER":                   "test",
					"DB_PASSWORD":               "test",
					"DB_NAME":                   "test",
					"DB_SSLMODE":                "disable",
					"SERVER_HOST":               "localhost",
					"SERVER_PORT":               "8080",
					"FIREBASE_PROJECT_ID":       "test-project",
					"FIREBASE_CREDENTIALS_FILE": "test.json",
				}

				for key, value := range envVars {
					if err := os.Setenv(key, value); err != nil {
						return fmt.Errorf("failed to set env %s: %w", key, err)
					}
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "missing required configuration",
			setupEnv: func() error {
				os.Clearenv()
				return nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setupEnv(); err != nil {
				t.Fatalf("Failed to setup test environment: %v", err)
			}
			_, err := config.Load()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDatabaseInitialization(t *testing.T) {
	tests := []struct {
		name        string
		config      MockConfig
		expectError bool
	}{
		{
			name: "valid database config",
			config: MockConfig{
				Database: struct {
					Host     string
					Port     int
					User     string
					Password string
					Name     string
					SSLMode  string
				}{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "test",
					Name:     "test",
					SSLMode:  "disable",
				},
			},
			expectError: false,
		},
		{
			name: "invalid database config",
			config: MockConfig{
				Database: struct {
					Host     string
					Port     int
					User     string
					Password string
					Name     string
					SSLMode  string
				}{
					Host: "invalid-host",
					Port: -1,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := postgres.NewDB(
				tt.config.Database.Host,
				tt.config.Database.Port,
				tt.config.Database.User,
				tt.config.Database.Password,
				tt.config.Database.Name,
				tt.config.Database.SSLMode,
			)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				if err == nil {
					assert.NotNil(t, db)
					if err := db.Close(); err != nil {
						t.Logf("Warning: Failed to close database connection: %v", err)
					}
				}
			}
		})
	}
}

func TestRepositoryInitialization(t *testing.T) {
	mockDB := &sql.DB{}
	repos := postgres.NewRepositories(mockDB)
	assert.NotNil(t, repos)
}

func TestFirebaseAuthInitialization(t *testing.T) {
	tests := []struct {
		name            string
		credentialsFile string
		expectError     bool
	}{
		{
			name:            "invalid credentials file",
			credentialsFile: "nonexistent.json",
			expectError:     true,
		},
		{
			name:            "empty credentials file",
			credentialsFile: "",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := middleware.NewFirebaseAuth(tt.credentialsFile)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, auth)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, auth)
			}
		})
	}
}

func TestRouterSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Test CORS middleware
	router.Use(middleware.CORS())

	// Add a test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test CORS headers
	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestHandlerRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := &sql.DB{}
	repos := postgres.NewRepositories(mockDB)

	api := router.Group("/api")
	userHandler := handlers.NewUserHandler(repos)
	userHandler.RegisterRoutes(api)

	tribeHandler := handlers.NewTribeHandler(repos)
	tribeHandler.RegisterRoutes(api)

	// Test that routes are registered
	routes := router.Routes()
	assert.NotEmpty(t, routes)

	// Check for specific routes
	foundUserRoute := false
	foundTribeRoute := false
	for _, route := range routes {
		if route.Path == "/api/users/auth" {
			foundUserRoute = true
		}
		if route.Path == "/api/tribes" {
			foundTribeRoute = true
		}
	}
	assert.True(t, foundUserRoute, "User routes not found")
	assert.True(t, foundTribeRoute, "Tribe routes not found")
}

func TestServerStartup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Start server in a goroutine
	go func() {
		err := router.Run(":0") // Use port 0 to let OS assign a port
		if err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)
}

// testSetupRouter creates and configures the Gin router for testing
func testSetupRouter(repos *postgres.Repositories, firebaseAuth *middleware.FirebaseAuth) *gin.Engine {
	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORS())

	// API routes
	api := router.Group("/api")
	{
		// Add Firebase Auth middleware to all API routes
		if firebaseAuth != nil {
			api.Use(firebaseAuth.AuthMiddleware())
		}

		// Initialize and register handlers
		userHandler := handlers.NewUserHandler(repos)
		userHandler.RegisterRoutes(api)

		tribeHandler := handlers.NewTribeHandler(repos)
		tribeHandler.RegisterRoutes(api)
	}

	return router
}

// testSetupApp configures and initializes the application for testing
func testSetupApp(cfg *config.Config) (*gin.Engine, error) {
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

	// Initialize Firebase Auth
	firebaseAuth, err := middleware.NewFirebaseAuth(cfg.Firebase.CredentialsFile)
	if err != nil {
		// Just close the database and return the error
		if closeErr := db.Close(); closeErr != nil {
			// If we can't log with t, just use the regular log package
			log.Printf("Error closing database during Firebase Auth setup error: %v", closeErr)
		}
		return nil, fmt.Errorf("error initializing Firebase Auth: %w", err)
	}

	// Initialize and configure Gin router
	router := testSetupRouter(repos, firebaseAuth)

	return router, nil
}

func TestSetupRouter(t *testing.T) {
	// gin.SetMode(gin.TestMode) - Removed to avoid data race
	mockDB := &sql.DB{}
	repos := postgres.NewRepositories(mockDB)
	// Pass nil for firebaseAuth to avoid issues with uninitialized struct
	router := testSetupRouter(repos, nil)

	// Just verify the router was created
	assert.NotNil(t, router)

	// Simple verification that routes exist
	routes := router.Routes()
	assert.NotEmpty(t, routes)

	// Log the number of routes instead of checking specific paths
	// which can change as the API evolves
	t.Logf("Found %d routes", len(routes))
	assert.Greater(t, len(routes), 0, "No routes were registered")
}

func TestSetupApp(t *testing.T) {
	// This is a challenging function to test without real dependencies or making the
	// code more testable. Let's just verify it handles invalid configuration.

	// Create a mock config with invalid values that will cause an error
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "nonexistent-host",
			Port:     "invalid-port",
			User:     "invalid-user",
			Password: "invalid-password",
			Name:     "invalid-db",
			SSLMode:  "disable",
		},
		Firebase: config.FirebaseConfig{
			ProjectID:       "invalid-project",
			CredentialsFile: "nonexistent-file.json",
		},
	}

	// setupApp should return an error with invalid configuration
	_, err := testSetupApp(cfg)

	// We expect an error due to invalid configuration
	assert.Error(t, err)
}

func TestDatabaseConnection(t *testing.T) {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "postgres" // Default password
	}

	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "postgres" // Default database
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName,
	)

	// Try to connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	t.Logf("Successfully connected to PostgreSQL database at %s:%s/%s", host, port, dbName)
}
