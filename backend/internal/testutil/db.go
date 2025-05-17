package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	testDB  *sql.DB
	testDBs []string
)

// Track current test database name
var currentTestDBName string

// getPostgresConnection returns the connection string for the Postgres database
func getPostgresConnection(dbName string) string {
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
	sslMode := "disable"

	if dbName == "" {
		// Connect to default database
		defaultDB := os.Getenv("POSTGRES_DB")
		if defaultDB == "" {
			defaultDB = "postgres"
		}
		if password == "" {
			return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
				host, port, user, defaultDB, sslMode)
		}
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, defaultDB, sslMode)
	}

	// Connect to specific database
	if password == "" {
		return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
			host, port, user, dbName, sslMode)
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode)
}

// cleanupDatabase ensures the test database is dropped even if setup fails
func cleanupDatabase(dbName string) {
	if dbName == "" {
		return
	}

	// Connect to default postgres database
	db, err := sql.Open("postgres", getPostgresConnection(""))
	if err != nil {
		fmt.Printf("DEBUG: Error connecting to postgres for cleanup: %v\n", err)
		return // Can't clean up if we can't connect
	}
	defer db.Close()

	// Terminate any remaining connections with retries
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		_, err = db.Exec(fmt.Sprintf(`
			SELECT pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = %s AND pid != pg_backend_pid()
		`, pq.QuoteLiteral(dbName)))

		if err == nil {
			break
		}

		if i < maxRetries-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Drop the test database with retries
	for i := 0; i < maxRetries; i++ {
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(dbName)))
		if err == nil {
			fmt.Printf("DEBUG: Successfully dropped database %s\n", dbName)
			break
		}

		if i < maxRetries-1 {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("DEBUG: Retrying database drop for %s, attempt %d\n", dbName, i+2)
		} else {
			fmt.Printf("DEBUG: Failed to drop database %s after %d attempts: %v\n", dbName, maxRetries, err)
		}
	}
}

// SetupTestDB creates a test database and runs migrations
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	fmt.Printf("DEBUG: Setting up test database\n")

	// Create a unique test database name
	dbName := fmt.Sprintf("test_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))
	fmt.Printf("DEBUG: Test database name: %s\n", dbName)

	// Validate database name
	if strings.Contains(dbName, "/") || strings.Contains(dbName, "\\") || strings.Contains(currentTestDBName, "/") || strings.Contains(currentTestDBName, "\\") {
		cleanupDatabase(dbName)
		panic(fmt.Sprintf("Invalid database name: %s", dbName))
	}

	// Connect to postgres to create test database
	db, err := sql.Open("postgres", getPostgresConnection(""))
	if err != nil {
		cleanupDatabase(dbName)
		panic(fmt.Sprintf("Error connecting to postgres: %v", err))
	}
	defer db.Close()

	// Create test database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(dbName)))
	if err != nil {
		cleanupDatabase(dbName)
		panic(fmt.Sprintf("Error dropping existing database: %v", err))
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbName)))
	if err != nil {
		cleanupDatabase(dbName)
		panic(fmt.Sprintf("Error creating test database: %v", err))
	}

	// Connect to test database
	testDB, err := sql.Open("postgres", getPostgresConnection(dbName))
	if err != nil {
		cleanupDatabase(dbName)
		panic(fmt.Sprintf("Error connecting to test database: %v", err))
	}

	// Ensure we clean up if something goes wrong
	setupSuccess := false
	defer func() {
		if !setupSuccess {
			testDB.Close()
			cleanupDatabase(dbName)
		}
	}()

	// Run migrations
	fmt.Printf("DEBUG: Running migrations\n")

	// Get the absolute path to the migrations directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get current file path")
	}
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	fmt.Printf("DEBUG: Migrations path: %s\n", migrationsPath)

	// Build postgres connection URL for migrations
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
	passwordPart := ""
	if password != "" {
		passwordPart = ":" + password
	}

	postgresURL := fmt.Sprintf("postgres://%s%s@%s:%s/%s?sslmode=disable",
		user, passwordPart, host, port, dbName)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		postgresURL,
	)
	if err != nil {
		panic(fmt.Sprintf("Error creating migrate instance: %v", err))
	}
	defer m.Close()

	// Check for dirty state before running migrations
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		panic(fmt.Sprintf("Error checking migration version: %v", err))
	}
	if dirty {
		panic(fmt.Sprintf("Database is in a dirty state before migrations (version %d)", version))
	}

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		panic(fmt.Sprintf("Error running migrations: %v", err))
	}

	// Verify final migration state
	version, dirty, err = m.Version()
	if err != nil {
		panic(fmt.Sprintf("Error getting final migration version: %v", err))
	}
	if dirty {
		panic(fmt.Sprintf("Database is in a dirty state after migrations (version %d)", version))
	}

	fmt.Printf("DEBUG: Final migration version: %d, dirty: %v\n", version, dirty)

	// Mark setup as successful
	setupSuccess = true

	// Set current test database name only after successful setup
	currentTestDBName = dbName

	// Add test database to cleanup list
	testDBs = append(testDBs, dbName)

	return testDB
}

// TeardownTestDB cleans up the test database
func TeardownTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	if db != nil && currentTestDBName != "" {
		// Close the test database connection
		if err := db.Close(); err != nil {
			t.Errorf("Error closing test database connection: %v", err)
		}

		// Wait a bit to ensure all connections are closed
		time.Sleep(100 * time.Millisecond)

		cleanupDatabase(currentTestDBName)

		// Clear the current test database name
		currentTestDBName = ""
	}
}
