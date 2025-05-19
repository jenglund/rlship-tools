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
func SetupTestDB(t *testing.T) *SchemaDB {
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

	// Create a prepared statement that will set the search path for every new connection
	_, err = db.Exec(fmt.Sprintf("PREPARE set_search_path AS SET search_path TO %s",
		pq.QuoteIdentifier(schemaName)))
	if err != nil {
		t.Logf("Warning: Failed to create prepared statement for search path: %v", err)
	}

	// Run migrations
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get current file path")
	}
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	// Debug logging
	t.Logf("Migrations path: %s", migrationsPath)
	// Check if migrations directory exists and is accessible
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		t.Logf("ERROR: Migrations directory does not exist: %s", migrationsPath)
		panic(fmt.Sprintf("Migrations directory does not exist: %s", migrationsPath))
	}

	// List files in migrations directory for debugging
	migrationDirFiles, err := os.ReadDir(migrationsPath)
	if err != nil {
		t.Logf("ERROR: Failed to read migrations directory: %v", err)
	} else {
		t.Logf("Found %d files in migrations directory:", len(migrationDirFiles))
		for _, file := range migrationDirFiles {
			t.Logf("  - %s", file.Name())
		}
	}

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

		t.Logf("Found %d migration files to execute", len(upMigrations))

		// Direct approach: try to execute the schema file directly first
		schemaFilePath := filepath.Join(migrationsPath, "000001_schema.up.sql")
		if _, err := os.Stat(schemaFilePath); err == nil {
			t.Logf("Executing schema file directly: %s", schemaFilePath)
			schemaSQL, err := os.ReadFile(schemaFilePath)
			if err != nil {
				t.Logf("Error reading schema file: %v", err)
			} else {
				// Set search_path and execute the schema SQL
				_, err = db.Exec(fmt.Sprintf("SET search_path TO %s; %s",
					pq.QuoteIdentifier(schemaName), string(schemaSQL)))
				if err != nil {
					t.Logf("Error executing schema SQL: %v", err)
				} else {
					t.Logf("Successfully executed schema SQL directly")
					migrationSuccess = true
				}
			}
		}

		// If direct execution of schema file failed or wasn't available, try individual migrations
		if !migrationSuccess {
			// Execute each migration file
			for _, fileName := range upMigrations {
				t.Logf("Executing migration file: %s", fileName)

				migrationSQL, err := os.ReadFile(filepath.Join(migrationsPath, fileName))
				if err != nil {
					t.Logf("Error reading migration file %s: %v", fileName, err)
					continue
				}

				// Execute the migration with schema path explicitly set
				_, err = db.Exec(fmt.Sprintf("SET search_path TO %s; %s",
					pq.QuoteIdentifier(schemaName), string(migrationSQL)))
				if err != nil {
					t.Logf("Warning: Error executing migration %s: %v", fileName, err)
					// Continue with other migrations, don't fail completely
				} else {
					t.Logf("Successfully executed migration file: %s", fileName)
				}
			}
		}
		t.Logf("Direct SQL execution completed")
	}

	// Double check essential tables
	tables := []string{"users", "tribes", "tribe_members", "lists", "list_items", "list_owners", "list_sharing", "list_conflicts", "activities", "activity_owners", "activity_photos", "activity_shares"}
	tablesExist := true
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
			tablesExist = false
		} else {
			t.Logf("%s table exists: %v", table, tableExists)
			if !tableExists {
				tablesExist = false
			}
		}
	}

	if !tablesExist {
		// Last resort: try to create schema directly
		t.Logf("Some essential tables are missing, executing schema SQL one more time")
		schemaFilePath := filepath.Join(migrationsPath, "000001_schema.up.sql")
		schemaSQL, err := os.ReadFile(schemaFilePath)
		if err != nil {
			t.Logf("Error reading schema SQL file: %v", err)
			panic("Failed to create essential database tables")
		}

		// Execute schema SQL with explicit schema name
		_, err = db.Exec(fmt.Sprintf("SET search_path TO %s; %s",
			pq.QuoteIdentifier(schemaName), string(schemaSQL)))
		if err != nil {
			t.Logf("Error executing schema SQL: %v", err)
			panic("Failed to create essential database tables")
		}

		// Verify tables one more time
		for _, table := range tables {
			var tableExists bool
			err = db.QueryRow(`
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_schema = $1
					AND table_name = $2
				)`, schemaName, table).Scan(&tableExists)
			t.Logf("After final attempt, %s table exists: %v", table, tableExists)
			if !tableExists {
				panic(fmt.Sprintf("Failed to create table: %s", table))
			}
		}
	}

	// Create a wrapper DB that will set the search path for all operations
	wrappedDB := &SchemaDB{
		DB:         db,
		schemaName: schemaName,
	}

	// Add test schema to list for cleanup
	testDBs = append(testDBs, schemaName)

	return wrappedDB
}

// SchemaDB wraps a *sql.DB to ensure the search path is set for all operations
type SchemaDB struct {
	*sql.DB
	schemaName string
}

// Ensure common DB methods set the search path

func (db *SchemaDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(fmt.Sprintf("SET search_path TO %s; %s",
		pq.QuoteIdentifier(db.schemaName), query), args...)
}

func (db *SchemaDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(fmt.Sprintf("SET search_path TO %s; %s",
		pq.QuoteIdentifier(db.schemaName), query), args...)
}

func (db *SchemaDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRow(fmt.Sprintf("SET search_path TO %s; %s",
		pq.QuoteIdentifier(db.schemaName), query), args...)
}

func (db *SchemaDB) Begin() (*sql.Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(db.schemaName)))
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

func (db *SchemaDB) Close() error {
	return db.DB.Close()
}

// UnwrapDB returns the underlying *sql.DB for backward compatibility
// This allows existing code to use the *SchemaDB as a *sql.DB
func (db *SchemaDB) UnwrapDB() *sql.DB {
	return db.DB
}

// GetSchemaName returns the schema name for this database connection
func (db *SchemaDB) GetSchemaName() string {
	return db.schemaName
}

// TeardownTestDB cleans up the test schema
func TeardownTestDB(t *testing.T, db *SchemaDB) {
	t.Helper()

	if db != nil && currentTestDBName != "" {
		// Clean up schema
		_, err := db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(currentTestDBName)))
		if err != nil {
			t.Logf("Error dropping test schema: %v", err)
		}

		// Close the database connection
		safeClose(db.DB)

		// Clear the current test schema name
		currentTestDBName = ""
	}
}

// GetCurrentTestSchema returns the name of the current test schema
func GetCurrentTestSchema() string {
	return currentTestDBName
}

// GetDB returns a *sql.DB from any supported DB type (for backward compatibility)
func GetDB(db interface{}) *sql.DB {
	switch d := db.(type) {
	case *sql.DB:
		return d
	case *SchemaDB:
		return d.DB
	default:
		panic(fmt.Sprintf("Unsupported DB type: %T", db))
	}
}
