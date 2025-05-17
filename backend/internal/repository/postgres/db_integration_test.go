package postgres

import (
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseIntegration_ConnectionLifecycle tests the database connection lifecycle
// This test requires a running postgres instance and is skipped if DB environment vars are not set
func TestDatabaseIntegration_ConnectionLifecycle(t *testing.T) {
	// Skip test if DB environment variables are not set
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		t.Skip("Skipping integration test because TEST_DB_HOST is not set")
	}

	// Get database connection parameters from environment variables
	port := 5432 // default
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_NAME")
	sslmode := os.Getenv("TEST_DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable" // default
	}

	// Create database connection
	db, err := NewDB(host, port, user, password, dbname, sslmode)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Test that connection is working
	err = db.Ping()
	assert.NoError(t, err)

	// Test querying
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	// Close connection
	err = db.Close()
	assert.NoError(t, err)

	// Verify connection is closed
	err = db.Ping()
	assert.Error(t, err)
}

// TestDatabaseIntegration_Reconnection tests that the application can recover from connection failures
// This test requires a running postgres instance and is skipped if DB environment vars are not set
func TestDatabaseIntegration_Reconnection(t *testing.T) {
	// Skip test if DB environment variables are not set
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		t.Skip("Skipping integration test because TEST_DB_HOST is not set")
	}

	// Get database connection parameters from environment variables
	port := 5432 // default
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_NAME")
	sslmode := os.Getenv("TEST_DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable" // default
	}

	// Create database connection
	db, err := NewDB(host, port, user, password, dbname, sslmode)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer safeClose(db)

	// Simulate idle connection timeout by waiting
	t.Log("Waiting for connection to potentially timeout...")
	time.Sleep(1 * time.Second) // This is just for demonstration, a real timeout would be much longer

	// Connection should auto-reconnect
	err = db.Ping()
	assert.NoError(t, err, "Database should auto-reconnect")

	// Verify we can still execute queries
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

// TestDatabaseIntegration_ConnectionPooling tests the connection pool behavior
// This test requires a running postgres instance and is skipped if DB environment vars are not set
func TestDatabaseIntegration_ConnectionPooling(t *testing.T) {
	// Skip test if DB environment variables are not set
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		t.Skip("Skipping integration test because TEST_DB_HOST is not set")
	}

	// Get database connection parameters from environment variables
	port := 5432 // default
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_NAME")
	sslmode := os.Getenv("TEST_DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable" // default
	}

	// Create database connection
	db, err := NewDB(host, port, user, password, dbname, sslmode)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer safeClose(db)

	// Verify connection pool settings
	assert.Equal(t, 25, db.Stats().MaxOpenConnections)

	// Execute multiple concurrent queries to test pool
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			var result int
			err := db.QueryRow("SELECT 1").Scan(&result)
			assert.NoError(t, err)
			assert.Equal(t, 1, result)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check connection pool stats
	stats := db.Stats()
	t.Logf("Connection pool stats: open=%d, in-use=%d, idle=%d",
		stats.OpenConnections, stats.InUse, stats.Idle)

	// The exact values will depend on timing, but we can check that values are reasonable
	assert.True(t, stats.OpenConnections > 0, "Should have some open connections")
	assert.True(t, stats.OpenConnections <= 25, "Should not exceed max open connections")
}

// TestDatabaseIntegration_ErrorHandling tests error handling for invalid queries
// This test requires a running postgres instance and is skipped if DB environment vars are not set
func TestDatabaseIntegration_ErrorHandling(t *testing.T) {
	// Skip test if DB environment variables are not set
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		t.Skip("Skipping integration test because TEST_DB_HOST is not set")
	}

	// Get database connection parameters from environment variables
	port := 5432 // default
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_NAME")
	sslmode := os.Getenv("TEST_DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable" // default
	}

	// Create database connection
	db, err := NewDB(host, port, user, password, dbname, sslmode)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer safeClose(db)

	// Test invalid SQL syntax
	_, err = db.Exec("INVALID SQL QUERY")
	assert.Error(t, err, "Invalid SQL should return an error")

	// Test query on non-existent table
	_, err = db.Query("SELECT * FROM nonexistent_table")
	assert.Error(t, err, "Query on non-existent table should return an error")

	// Test that connection is still usable after errors
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

// TestRepositoriesIntegration_Initialization tests repository initialization with a real database
// This test requires a running postgres instance and is skipped if DB environment vars are not set
func TestRepositoriesIntegration_Initialization(t *testing.T) {
	// Skip test if DB environment variables are not set
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		t.Skip("Skipping integration test because TEST_DB_HOST is not set")
	}

	// Get database connection parameters from environment variables
	port := 5432 // default
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_NAME")
	sslmode := os.Getenv("TEST_DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable" // default
	}

	// Create database connection
	db, err := NewDB(host, port, user, password, dbname, sslmode)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer safeClose(db)

	// Initialize repositories
	repos := NewRepositories(db)
	assert.NotNil(t, repos)
	assert.NotNil(t, repos.Users)
	assert.NotNil(t, repos.Tribes)
	assert.NotNil(t, repos.Activities)
	assert.Equal(t, db, repos.DB())

	// Verify user repository is initialized
	userRepo := repos.GetUserRepository()
	assert.NotNil(t, userRepo)
}
