package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jenglund/rlship-tools/internal/api/handlers"
	"github.com/jenglund/rlship-tools/internal/config"
	"github.com/jenglund/rlship-tools/internal/middleware"
	"github.com/jenglund/rlship-tools/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// MockFirebaseAuth is a mock implementation of FirebaseAuth
type MockFirebaseAuth struct {
	mock.Mock
}

func (m *MockFirebaseAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
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
			os.Setenv(key, val)
		}
	}()

	tests := []struct {
		name        string
		setupEnv    func()
		expectError bool
	}{
		{
			name: "valid configuration",
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_PASSWORD", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("DB_SSLMODE", "disable")
				os.Setenv("SERVER_HOST", "localhost")
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("FIREBASE_PROJECT_ID", "test-project")
				os.Setenv("FIREBASE_CREDENTIALS_FILE", "test.json")
			},
			expectError: false,
		},
		{
			name: "missing required configuration",
			setupEnv: func() {
				os.Clearenv()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
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
					db.Close()
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
