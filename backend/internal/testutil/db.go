package testutil

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var (
	testDB *sql.DB
)

// SetupTestDB creates a test database and runs migrations
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Get current user
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Error getting current user: %v", err)
	}

	dbName := fmt.Sprintf("rlship_test_%d", os.Getpid())
	connStr := fmt.Sprintf("host=localhost port=5432 user=%s dbname=postgres sslmode=disable", currentUser.Username)

	// Connect to default postgres database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Error connecting to database: %v", err)
	}

	// Create test database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		t.Fatalf("Error dropping test database: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Fatalf("Error creating test database: %v", err)
	}
	db.Close()

	// Connect to test database
	testConnStr := fmt.Sprintf("host=localhost port=5432 user=%s dbname=%s sslmode=disable", currentUser.Username, dbName)
	testDB, err = sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Error connecting to test database: %v", err)
	}

	// Run migrations
	driver, err := postgres.WithInstance(testDB, &postgres.Config{})
	if err != nil {
		t.Fatalf("Error creating postgres driver: %v", err)
	}

	// Get the path to the migrations directory
	_, b, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(b), "..", "..", "migrations")

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		t.Fatalf("Error creating migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Error running migrations: %v", err)
	}

	return testDB
}

// TeardownTestDB cleans up the test database
func TeardownTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	if db != nil {
		dbName := fmt.Sprintf("rlship_test_%d", os.Getpid())
		db.Close()

		// Get current user
		currentUser, err := user.Current()
		if err != nil {
			t.Fatalf("Error getting current user: %v", err)
		}

		// Connect to default postgres database to drop the test database
		connStr := fmt.Sprintf("host=localhost port=5432 user=%s dbname=postgres sslmode=disable", currentUser.Username)
		defaultDB, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Error connecting to default database during cleanup: %v", err)
			return
		}
		defer defaultDB.Close()

		_, err = defaultDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
		if err != nil {
			log.Printf("Error dropping test database during cleanup: %v", err)
		}
	}
}
