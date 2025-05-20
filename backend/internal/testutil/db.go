package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Add a global mutex to protect schema name access
var (
	testDBs           []string
	testDBsMutex      sync.Mutex // Mutex to protect testDBs
	currentTestDBName string
	currentDBMutex    sync.Mutex // Mutex to protect currentTestDBName

	// Registry of active schema contexts for test isolation
	SchemaContexts     map[string]*SchemaContext // Exported for direct access
	SchemaContextMutex sync.Mutex                // Exported for direct access
)

// Initialize the schema contexts map
func init() {
	SchemaContexts = make(map[string]*SchemaContext)
}

// Default timeout for database operations to prevent test hangs
var defaultQueryTimeout = 10 * time.Second

// DB connection pool settings for tests
var (
	// Tests tend to have many short-lived transactions, so we need a decent pool size
	maxOpenConns = 10              // How many connections can be open at once
	maxIdleConns = 5               // How many idle connections to maintain in the pool
	connLifetime = 5 * time.Minute // How long a connection can live before being recycled
)

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
	if password == "" {
		password = "postgres"
	}
	sslMode := "disable"

	if dbName == "" || dbName == "test" {
		// Connect to default database, ensuring we don't try to connect to a non-existent 'test' database
		dbName = "postgres"
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

	// Create a context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	// Drop the schema
	_, err = db.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(schemaName)))
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

// runMigrations runs all database migrations for the test schema
func runMigrations(t *testing.T, dbName, schemaName, migrationsPath string) error {
	// List files in migrations directory for debugging
	migrationDirFiles, err := os.ReadDir(migrationsPath)
	if err != nil {
		t.Logf("ERROR: Failed to read migrations directory: %v", err)
		return err
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
	if password == "" {
		password = "postgres"
	}
	passwordPart := ":" + password

	// Add schema to connection URL
	postgresURL := fmt.Sprintf("postgres://%s%s@%s:%s/%s?sslmode=disable&search_path=%s",
		user, passwordPart, host, port, dbName, schemaName)

	// Debug logging - use the URL to prevent unused variable warning
	t.Logf("Using Postgres URL: %s (with password redacted)",
		strings.ReplaceAll(postgresURL, passwordPart, ":***"))

	// Connect to database to execute migrations manually
	db, err := sql.Open("postgres", getPostgresConnection(dbName))
	if err != nil {
		return fmt.Errorf("error connecting to database for migrations: %w", err)
	}
	defer safeClose(db)

	// Set search path to test schema
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	_, err = db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schemaName)))
	if err != nil {
		return fmt.Errorf("error setting search path for migrations: %w", err)
	}

	// Read and execute all migration files
	migrationFiles, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %w", err)
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

	// Execute each migration file
	for _, fileName := range upMigrations {
		t.Logf("Executing migration file: %s", fileName)

		migrationSQL, err := os.ReadFile(filepath.Join(migrationsPath, fileName))
		if err != nil {
			t.Logf("Error reading migration file %s: %v", fileName, err)
			continue
		}

		// Create a transaction for each migration
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Logf("Error starting transaction for migration %s: %v", fileName, err)
			continue
		}

		// Set the search path in the transaction
		_, err = tx.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schemaName)))
		if err != nil {
			t.Logf("Error setting search path in transaction for migration %s: %v", fileName, err)
			tx.Rollback()
			continue
		}

		// Execute the migration within the transaction
		_, err = tx.ExecContext(ctx, string(migrationSQL))
		if err != nil {
			t.Logf("Error executing migration %s: %v", fileName, err)
			tx.Rollback()
			continue
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			t.Logf("Error committing transaction for migration %s: %v", fileName, err)
			continue
		}

		t.Logf("Successfully executed migration file: %s", fileName)
	}

	// Verify essential tables exist
	tables := []string{"users", "tribes", "tribe_members", "lists", "list_items", "list_owners", "list_sharing", "list_conflicts", "activities", "activity_owners", "activity_photos", "activity_shares"}
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
			if !tableExists {
				t.Logf("WARNING: Essential table %s is missing", table)
			}
		}
	}

	t.Logf("Migration execution completed")
	return nil
}

// NewConnectionFromUrl creates a new database connection from a URL string
func NewConnectionFromUrl(dbUrl string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for better stability in tests
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// SetupTestDB creates a test schema and runs migrations
func SetupTestDB(t *testing.T) *SchemaDB {
	t.Helper()

	// Create a unique schema name for this test
	schemaName := fmt.Sprintf("test_%s", strings.Replace(uuid.New().String(), "-", "_", -1))
	t.Logf("SetupTestDB: Setting current test schema to %s", schemaName)

	// Connect to the database
	connString := GetTestDBConnectionString()
	t.Logf("Connecting to Postgres with DB: %s", getDatabaseNameFromConnString(connString))

	db, err := NewConnectionFromUrl(connString)
	if err != nil {
		t.Fatalf("Error connecting to Postgres: %v", err)
	}
	t.Logf("Successfully connected to Postgres, creating schema: %s", schemaName)

	// Set schema search path and ensure it's applied
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	// Create a schema context for this test
	schemaCtx := NewSchemaContext(schemaName)

	// Register the schema context - with mutex protection
	SchemaContextMutex.Lock()
	SchemaContexts[schemaName] = schemaCtx
	SchemaContextMutex.Unlock()

	// For backward compatibility, also set the global variable
	currentDBMutex.Lock()
	currentTestDBName = schemaName
	fmt.Printf("SetupTestDB: Setting current test schema to %s\n", schemaName)
	currentDBMutex.Unlock()

	// Create test schema
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", pq.QuoteIdentifier(schemaName)))
	if err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error creating test schema: %v", err))
	}

	// Set search path to our test schema
	_, err = db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", pq.QuoteIdentifier(schemaName)))
	if err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error setting search path: %v", err))
	}

	// Verify search path was properly set
	var searchPath string
	err = db.QueryRowContext(ctx, "SHOW search_path").Scan(&searchPath)
	if err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error verifying search path: %v", err))
	}
	t.Logf("Initial search_path set to: %s", searchPath)
	if !strings.Contains(searchPath, schemaName) {
		safeClose(db)
		panic(fmt.Sprintf("Failed to set correct search path: expected %s in %s", schemaName, searchPath))
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

	// Run migrations with improved error handling
	if err := runMigrations(t, getDatabaseNameFromConnString(connString), schemaName, migrationsPath); err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error running migrations: %v", err))
	}

	// After migrations, verify schema is still correctly set
	err = db.QueryRowContext(ctx, "SHOW search_path").Scan(&searchPath)
	if err != nil {
		safeClose(db)
		panic(fmt.Sprintf("Error verifying search path after migrations: %v", err))
	}
	t.Logf("Post-migration search_path: %s", searchPath)
	if !strings.Contains(searchPath, schemaName) {
		safeClose(db)
		panic(fmt.Sprintf("Search path lost after migrations: expected %s in %s", schemaName, searchPath))
	}

	// Create wrapped DB with schema handling
	wrappedDB := &SchemaDB{
		DB:            db,
		schemaName:    schemaName,
		schemaContext: schemaCtx,
		timeout:       defaultQueryTimeout,
		mu:            sync.Mutex{}, // Initialize mutex
	}

	// Add test schema to list for cleanup - with mutex protection
	testDBsMutex.Lock()
	testDBs = append(testDBs, schemaName)
	testDBsMutex.Unlock()

	return wrappedDB
}

// SchemaDB wraps a *sql.DB to ensure the search path is set for all operations
type SchemaDB struct {
	*sql.DB
	schemaName    string
	schemaContext *SchemaContext // Added schema context reference
	timeout       time.Duration
	mu            sync.Mutex // Add a mutex for thread safety
}

// SetSearchPath sets the search path for the current connection
func (db *SchemaDB) SetSearchPath() error {
	// Create a new context each time to avoid data races
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	// Lock to prevent concurrent schema changes
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.DB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(db.schemaName)))
	if err != nil {
		return fmt.Errorf("error setting search path: %w", err)
	}

	// Verify search path was set correctly
	var searchPath string
	err = db.DB.QueryRowContext(ctx, "SHOW search_path").Scan(&searchPath)
	if err != nil {
		return fmt.Errorf("error verifying search path: %w", err)
	}

	if !strings.Contains(searchPath, db.schemaName) {
		return fmt.Errorf("schema not in search path: expected %s in %s", db.schemaName, searchPath)
	}

	return nil
}

// Exec overrides the default Exec method to set the search path
func (db *SchemaDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	// Create a new context each time to avoid data races
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	return db.ExecContext(ctx, query, args...)
}

// ExecContext overrides the default ExecContext method to set the search path
func (db *SchemaDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// Create a child context with a separate timeout to avoid affecting the parent
	childCtx, cancel := context.WithTimeout(ctx, db.timeout)
	defer cancel()

	// First set the search path
	err := db.SetSearchPath()
	if err != nil {
		return nil, fmt.Errorf("error setting search path: %w", err)
	}

	// Execute the query as a separate operation
	return db.DB.ExecContext(childCtx, query, args...)
}

// Query overrides the default Query method to set the search path
func (db *SchemaDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// Create a new context each time to avoid data races
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	return db.QueryContext(ctx, query, args...)
}

// QueryContext overrides the default QueryContext method to set the search path
func (db *SchemaDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Create a child context with a separate timeout to avoid affecting the parent
	childCtx, cancel := context.WithTimeout(ctx, db.timeout)
	defer cancel()

	// First set the search path
	err := db.SetSearchPath()
	if err != nil {
		return nil, fmt.Errorf("error setting search path: %w", err)
	}

	// Execute the query as a separate operation
	return db.DB.QueryContext(childCtx, query, args...)
}

// QueryRow overrides the default QueryRow method to set the search path
func (db *SchemaDB) QueryRow(query string, args ...interface{}) *sql.Row {
	// Create a new context each time to avoid data races
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	return db.QueryRowContext(ctx, query, args...)
}

// QueryRowContext overrides the default QueryRowContext method to set the search path
func (db *SchemaDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Create a child context with a separate timeout to avoid affecting the parent
	childCtx, cancel := context.WithTimeout(ctx, db.timeout)
	defer cancel()

	// Set the search path first
	err := db.SetSearchPath()
	if err != nil {
		// Since QueryRow doesn't return an error, log it
		log.Printf("Error setting search path in QueryRowContext: %v", err)
	}

	// Execute the query as a separate operation
	return db.DB.QueryRowContext(childCtx, query, args...)
}

// Begin overrides the default Begin method to set the search path on the transaction
func (db *SchemaDB) Begin() (*sql.Tx, error) {
	// Create a new context each time to avoid data races
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	return db.BeginTx(ctx, nil)
}

// BeginTx overrides the default BeginTx method to set the search path on the transaction
func (db *SchemaDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// Create a child context with a separate timeout to avoid affecting the parent
	childCtx, cancel := context.WithTimeout(ctx, db.timeout)
	defer cancel()

	// Lock to ensure consistent schema setting
	db.mu.Lock()
	defer db.mu.Unlock()

	// Start transaction
	tx, err := db.DB.BeginTx(childCtx, opts)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	// Set LOCAL search path in the transaction
	// Use LOCAL to avoid affecting other sessions
	_, err = tx.ExecContext(childCtx, fmt.Sprintf("SET LOCAL search_path TO %s", pq.QuoteIdentifier(db.schemaName)))
	if err != nil {
		// In case of error, safely rollback
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Printf("Error rolling back transaction: %v", rollbackErr)
		}
		return nil, fmt.Errorf("error setting search path in transaction: %w", err)
	}

	// Verify search path was set correctly (without consuming any results)
	var searchPath string
	err = tx.QueryRowContext(childCtx, "SELECT current_schema").Scan(&searchPath)
	if err != nil {
		// In case of error, safely rollback
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Printf("Error rolling back transaction: %v", rollbackErr)
		}
		return nil, fmt.Errorf("error verifying search path in transaction: %w", err)
	}

	// Verify the schema was set correctly
	if searchPath != db.schemaName {
		// In case of mismatch, safely rollback
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Printf("Error rolling back transaction: %v", rollbackErr)
		}
		return nil, fmt.Errorf("schema mismatch in transaction: got %s, expected %s",
			searchPath, db.schemaName)
	}

	// Debug output
	fmt.Printf("Transaction search_path: %s (expected test schema: %s)\n",
		searchPath, db.schemaName)

	return tx, nil
}

func (db *SchemaDB) Close() error {
	return db.DB.Close()
}

// TeardownTestDB cleans up the test schema
func TeardownTestDB(t *testing.T, db *SchemaDB) {
	t.Helper()

	if db == nil {
		return
	}

	// Get schema name from the SchemaDB
	schemaName := db.GetSchemaName()

	if schemaName == "" {
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	// Lock to prevent concurrent schema changes
	db.mu.Lock()
	// Clean up schema
	_, err := db.DB.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(schemaName)))
	db.mu.Unlock()

	if err != nil {
		t.Logf("Error dropping test schema: %v", err)
	}

	// Close the database connection
	safeClose(db.DB)

	// Remove the schema context from the registry
	SchemaContextMutex.Lock()
	delete(SchemaContexts, schemaName)
	SchemaContextMutex.Unlock()

	// For backward compatibility, clear the global variable
	currentDBMutex.Lock()
	if currentTestDBName == schemaName {
		currentTestDBName = ""
	}
	currentDBMutex.Unlock()
}

// GetCurrentTestSchema returns the name of the current test schema
// This is an enhanced version that can look up schema contexts from registry
func GetCurrentTestSchema() string {
	// First, try to look up from the global variable (for backward compatibility)
	currentDBMutex.Lock()
	schema := currentTestDBName
	currentDBMutex.Unlock()

	if schema != "" {
		return schema
	}

	// If not found, see if we can determine schema from active contexts
	// This is more robust for parallel tests
	SchemaContextMutex.Lock()
	defer SchemaContextMutex.Unlock()

	// If there's only one active schema context, use that
	if len(SchemaContexts) == 1 {
		for _, ctx := range SchemaContexts {
			return ctx.GetSchemaName()
		}
	}

	// Otherwise return empty - caller will need to provide context explicitly
	return ""
}

// GetSchemaContext returns the schema context for the given schema name
func GetSchemaContext(schemaName string) *SchemaContext {
	SchemaContextMutex.Lock()
	defer SchemaContextMutex.Unlock()

	return SchemaContexts[schemaName]
}

// UnwrapDB sets the search path and returns a raw DB connection
// This function is for backward compatibility
func UnwrapDB(db *SchemaDB) *sql.DB {
	if db == nil {
		return nil
	}

	// Create a context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	// Lock to prevent concurrent schema changes
	db.mu.Lock()

	// Store the schema information
	schemaName := db.schemaName

	// Update schema registry with this schema
	SchemaContextMutex.Lock()
	if db.schemaContext == nil {
		db.schemaContext = NewSchemaContext(schemaName)
		SchemaContexts[schemaName] = db.schemaContext
	}
	SchemaContextMutex.Unlock()

	// For backward compatibility
	currentDBMutex.Lock()
	currentTestDBName = schemaName
	currentDBMutex.Unlock()

	// Set search path on the DB connection
	_, err := db.DB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schemaName)))
	db.mu.Unlock()

	if err != nil {
		// Log the error but continue
		log.Printf("Warning: Failed to set search path in UnwrapDB: %v", err)
	} else {
		log.Printf("UnwrapDB: Set current schema to %s", schemaName)

		// Debug verify the search_path was set correctly
		verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer verifyCancel()

		var actualPath string
		err = db.DB.QueryRowContext(verifyCtx, "SHOW search_path").Scan(&actualPath)
		if err != nil {
			log.Printf("Warning: Failed to verify search_path in UnwrapDB: %v", err)
		} else if !strings.Contains(actualPath, schemaName) {
			log.Printf("Warning: Schema mismatch after UnwrapDB: got %s, expected %s", actualPath, schemaName)
		} else {
			log.Printf("UnwrapDB: Verified search_path contains %s: %s", schemaName, actualPath)
		}
	}

	return db.DB
}

// GetSchemaName returns the schema name for this database connection
func (db *SchemaDB) GetSchemaName() string {
	return db.schemaName
}

// UnwrapDB returns the underlying *sql.DB with the search path properly set
func (db *SchemaDB) UnwrapDB() *sql.DB {
	return UnwrapDB(db)
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

// GetTestDBConnectionString returns the connection string for the test database
func GetTestDBConnectionString() string {
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" || dbName == "test" {
		dbName = "postgres"
	}
	return getPostgresConnection(dbName)
}

// getDatabaseNameFromConnString extracts the database name from a connection string
func getDatabaseNameFromConnString(connString string) string {
	parts := strings.Split(connString, "/")
	if len(parts) < 4 {
		return "postgres" // Default if we can't parse
	}
	dbNameWithParams := parts[3]
	dbNameParts := strings.Split(dbNameWithParams, "?")
	return dbNameParts[0]
}
