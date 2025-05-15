package testutil

import (
	"database/sql"
	"fmt"
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

// SetupTestDB creates a test database and runs migrations
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	fmt.Printf("DEBUG: Setting up test database\n")

	// Create a unique test database name
	currentTestDBName = fmt.Sprintf("test_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))
	fmt.Printf("DEBUG: Test database name: %s\n", currentTestDBName)

	// Connect to postgres to create test database
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres sslmode=disable")
	if err != nil {
		fmt.Printf("DEBUG: Error connecting to postgres: %v\n", err)
		t.Fatalf("Error connecting to postgres: %v", err)
	}

	// Create test database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(currentTestDBName)))
	if err != nil {
		fmt.Printf("DEBUG: Error creating test database: %v\n", err)
		t.Fatalf("Error creating test database: %v", err)
	}
	db.Close()

	// Connect to test database
	testDB, err := sql.Open("postgres", fmt.Sprintf("host=localhost port=5432 user=postgres dbname=%s sslmode=disable", currentTestDBName))
	if err != nil {
		fmt.Printf("DEBUG: Error connecting to test database: %v\n", err)
		t.Fatalf("Error connecting to test database: %v", err)
	}

	// Run migrations
	fmt.Printf("DEBUG: Running migrations\n")

	// Get the absolute path to the migrations directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get current file path")
	}
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	fmt.Printf("DEBUG: Migrations path: %s\n", migrationsPath)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		fmt.Sprintf("postgres://postgres@localhost:5432/%s?sslmode=disable", currentTestDBName),
	)
	if err != nil {
		fmt.Printf("DEBUG: Error creating migrator: %v\n", err)
		t.Fatalf("Error creating migrator: %v", err)
	}

	// Print current migration version before running migrations
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		fmt.Printf("DEBUG: Error getting migration version: %v\n", err)
		t.Fatalf("Error getting migration version: %v", err)
	}
	fmt.Printf("DEBUG: Current migration version: %d, dirty: %v\n", version, dirty)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Printf("DEBUG: Error running migrations: %v\n", err)
		t.Fatalf("Error running migrations: %v", err)
	}

	// Print final migration version
	version, dirty, err = m.Version()
	if err != nil {
		fmt.Printf("DEBUG: Error getting final migration version: %v\n", err)
		t.Fatalf("Error getting final migration version: %v", err)
	}
	fmt.Printf("DEBUG: Final migration version: %d, dirty: %v\n", version, dirty)

	// Add test database to cleanup list
	testDBs = append(testDBs, currentTestDBName)

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

		// Connect to default postgres database
		defaultDB, err := sql.Open("postgres", "host=localhost port=5432 user=postgres sslmode=disable")
		if err != nil {
			t.Errorf("Error connecting to default database: %v", err)
			return
		}
		defer defaultDB.Close()

		// Terminate any remaining connections
		_, err = defaultDB.Exec(fmt.Sprintf(`
			SELECT pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = %s AND pid != pg_backend_pid()
		`, pq.QuoteLiteral(currentTestDBName)))
		if err != nil {
			t.Errorf("Error terminating connections: %v", err)
		}

		// Drop the test database
		_, err = defaultDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(currentTestDBName)))
		if err != nil {
			t.Errorf("Error dropping test database: %v", err)
		}

		// Clear the current test database name
		currentTestDBName = ""
	}
}
