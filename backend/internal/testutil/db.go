package testutil

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
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

// cleanupDatabase is now cleanupSchema and removes a schema instead of a database
func cleanupDatabase(schemaName string) {
	if schemaName == "" {
		return
	}

	// Connect to database
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "postgres"
	}

	db, err := sql.Open("postgres", getPostgresConnection(dbName))
	if err != nil {
		return // Can't clean up if we can't connect
	}
	defer safeClose(db)

	// Drop the schema
	_, err = db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(schemaName)))
	if err != nil {
		log.Printf("Error dropping schema %s: %v", schemaName, err)
	}
}

// safeClose helper for testutil package
func safeClose(c interface{}) {
	var err error
	switch v := c.(type) {
	case *sql.Rows:
		err = v.Close()
	case *sql.Stmt:
		err = v.Close()
	case *sql.Tx:
		err = v.Rollback()
	case *sql.DB:
		err = v.Close()
	case *migrate.Migrate:
		// migrate.Close() returns (error, int)
		err, _ = v.Close()
	default:
		// Nothing to close for other types
		return
	}

	if err != nil {
		log.Printf("Error closing resource: %v", err)
	}
}

// SetupTestDB creates a test schema and runs migrations
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Create a unique test schema name
	schemaName := fmt.Sprintf("test_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))

	// Store current schema name for cleanup
	currentTestDBName = schemaName

	// Connect to existing database
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "postgres"
	}

	db, err := sql.Open("postgres", getPostgresConnection(dbName))
	if err != nil {
		panic(fmt.Sprintf("Error connecting to postgres: %v", err))
	}

	// Ensure we can connect
	if err = db.Ping(); err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error pinging database: %v", err))
	}

	// Create test schema
	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", pq.QuoteIdentifier(schemaName)))
	if err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error creating test schema: %v", err))
	}

	// Set search path to our test schema
	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schemaName)))
	if err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error setting search path: %v", err))
	}

	// Run migrations
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get current file path")
	}
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

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

	// Add schema to connection URL
	postgresURL := fmt.Sprintf("postgres://%s%s@%s:%s/%s?sslmode=disable&search_path=%s",
		user, passwordPart, host, port, dbName, schemaName)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		postgresURL,
	)
	if err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error creating migrate instance: %v", err))
	}
	defer safeClose(m)

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		safeClose(db)
		panic(fmt.Sprintf("Error running migrations: %v", err))
	}

	// Add test schema to list for cleanup
	testDBs = append(testDBs, schemaName)

	return db
}

// TeardownTestDB cleans up the test schema
func TeardownTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	if db != nil && currentTestDBName != "" {
		// Clean up schema
		_, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(currentTestDBName)))
		if err != nil {
			t.Logf("Error dropping test schema: %v", err)
		}

		// Close the database connection
		safeClose(db)

		// Clear the current test schema name
		currentTestDBName = ""
	}
}
