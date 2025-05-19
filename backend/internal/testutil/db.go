package testutil

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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

	// Debug logging
	t.Logf("Connecting to Postgres with DB: %s", dbName)

	db, err := sql.Open("postgres", getPostgresConnection(dbName))
	if err != nil {
		panic(fmt.Sprintf("Error connecting to postgres: %v", err))
	}

	// Ensure we can connect
	if err = db.Ping(); err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error pinging database: %v", err))
	}

	// Debug logging
	t.Logf("Successfully connected to Postgres, creating schema: %s", schemaName)

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

	// Debug logging
	t.Logf("Migrations path: %s", migrationsPath)

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

	// Debug logging
	t.Logf("Using Postgres URL: %s (with password redacted)",
		fmt.Sprintf("postgres://%s:***@%s:%s/%s?sslmode=disable&search_path=%s",
			user, host, port, dbName, schemaName))

	// Try to run migrations
	migrationSuccess := false
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		postgresURL,
	)

	if err != nil {
		t.Logf("Warning: Error creating migrate instance: %v", err)
		t.Logf("Will try to run SQL directly from file...")
	} else {
		defer safeClose(m)

		// Run migrations
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			t.Logf("Warning: Error running migrations: %v", err)
			t.Logf("Will try to run SQL directly from file...")
		} else {
			migrationSuccess = true
			t.Logf("Migrations applied successfully")
		}
	}

	// If migrations failed, try running the SQL directly
	if !migrationSuccess {
		t.Logf("Falling back to direct SQL execution")

		// Read and execute all migration files
		migrationFiles, err := os.ReadDir(migrationsPath)
		if err != nil {
			safeClose(db)
			panic(fmt.Sprintf("Error reading migrations directory: %v", err))
		}

		// Sort migrations by name to ensure they're executed in order
		var upMigrations []string
		for _, file := range migrationFiles {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
				upMigrations = append(upMigrations, file.Name())
			}
		}

		// Sort migration files
		sort.Strings(upMigrations)

		// Execute each migration file
		for _, fileName := range upMigrations {
			t.Logf("Executing migration file: %s", fileName)

			migrationSQL, err := os.ReadFile(filepath.Join(migrationsPath, fileName))
			if err != nil {
				safeClose(db)
				panic(fmt.Sprintf("Error reading migration file %s: %v", fileName, err))
			}

			// Execute the migration with schema path explicitly set
			_, err = db.Exec(fmt.Sprintf("SET search_path TO %s; %s",
				pq.QuoteIdentifier(schemaName), string(migrationSQL)))
			if err != nil {
				t.Logf("Warning: Error executing migration %s: %v", fileName, err)
				// Continue with other migrations, don't fail completely
			}
		}
		t.Logf("Direct SQL execution completed")
	}

	// Double check essential tables
	tables := []string{"users", "lists", "list_items", "list_owners", "list_sharing", "list_conflicts"}
	for _, table := range tables {
		var tableExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = $1
				AND table_name = $2
			)`, schemaName, table).Scan(&tableExists)

		if err != nil {
			t.Logf("Error checking if %s table exists: %v", table, err)
		} else {
			t.Logf("%s table exists: %v", table, tableExists)

			// If table doesn't exist, attempt to create it using SQL from the schema file
			if !tableExists {
				t.Logf("Table %s doesn't exist, attempting to create it directly", table)
				schemaSQL, err := os.ReadFile(filepath.Join(migrationsPath, "000001_schema.up.sql"))
				if err != nil {
					t.Logf("Error reading schema SQL file: %v", err)
					continue
				}

				// Execute schema SQL with explicit schema name
				_, err = db.Exec(fmt.Sprintf("SET search_path TO %s; %s",
					pq.QuoteIdentifier(schemaName), string(schemaSQL)))
				if err != nil {
					t.Logf("Error executing schema SQL for table %s: %v", table, err)
				}
			}
		}
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

// GetCurrentTestSchema returns the name of the current test schema
func GetCurrentTestSchema() string {
	return currentTestDBName
}
