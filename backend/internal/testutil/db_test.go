package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTestDB(t *testing.T) {
	// Test successful database setup
	t.Run("successful setup", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)

		// Verify database connection works
		err := db.Ping()
		require.NoError(t, err)

		// Verify migrations were applied by checking for a known table
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = $1
				AND table_name = 'users'
			)
		`, currentTestDBName).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "users table should exist after migrations")

		// Clean up
		TeardownTestDB(t, db)
	})

	// Test multiple database setups
	t.Run("multiple setups", func(t *testing.T) {
		db1 := SetupTestDB(t)
		require.NotNil(t, db1)
		defer TeardownTestDB(t, db1)

		db2 := SetupTestDB(t)
		require.NotNil(t, db2)
		defer TeardownTestDB(t, db2)

		// Verify both databases are different
		var db1Schema, db2Schema string
		err := db1.QueryRow("SELECT current_schema").Scan(&db1Schema)
		require.NoError(t, err)
		err = db2.QueryRow("SELECT current_schema").Scan(&db2Schema)
		require.NoError(t, err)

		assert.NotEqual(t, db1Schema, db2Schema, "each setup should create a unique schema")
	})
}

func TestTeardownTestDB(t *testing.T) {
	// Test successful teardown
	t.Run("successful teardown", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)

		// Get the schema name before teardown
		var schemaName string
		err := db.QueryRow("SELECT current_schema").Scan(&schemaName)
		require.NoError(t, err)

		// Perform teardown
		TeardownTestDB(t, db)

		// Verify schema no longer exists by checking information_schema
		pgdb, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer safeClose(pgdb)

		var exists bool
		err = pgdb.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.schemata 
				WHERE schema_name = $1
			)
		`, schemaName).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "schema should not exist after teardown")
	})

	// Test teardown with active connections
	t.Run("teardown with active connections", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)

		// Get the schema name
		var schemaName string
		err := db.QueryRow("SELECT current_schema").Scan(&schemaName)
		require.NoError(t, err)

		// Create additional connections
		db2, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer safeClose(db2)

		// Set search path for additional connections
		_, err = db2.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schemaName)))
		require.NoError(t, err)

		db3, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer safeClose(db3)

		// Set search path for additional connections
		_, err = db3.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schemaName)))
		require.NoError(t, err)

		// Create a test table
		_, err = db.Exec("CREATE TABLE teardown_test (id SERIAL PRIMARY KEY, name TEXT)")
		require.NoError(t, err)

		// Perform some queries to ensure connections are active
		var count int
		err = db2.QueryRow("SELECT 1").Scan(&count)
		require.NoError(t, err)
		err = db3.QueryRow("SELECT 1").Scan(&count)
		require.NoError(t, err)

		// Perform teardown
		TeardownTestDB(t, db)

		// Verify connections are closed by attempting queries to the dropped schema's table
		err = db2.QueryRow("SELECT * FROM teardown_test LIMIT 1").Scan(&count)
		assert.Error(t, err, "query should fail after teardown")
		err = db3.QueryRow("SELECT * FROM teardown_test LIMIT 1").Scan(&count)
		assert.Error(t, err, "query should fail after teardown")
	})

	// Test teardown with nil database
	t.Run("teardown with nil database", func(t *testing.T) {
		assert.NotPanics(t, func() {
			TeardownTestDB(t, nil)
		})
	})

	// Test teardown with empty database name
	t.Run("teardown with empty database name", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		currentTestDBName = ""
		assert.NotPanics(t, func() {
			TeardownTestDB(t, db)
		})
	})
}

func TestDatabaseOperations(t *testing.T) {
	// Test basic database operations
	t.Run("basic operations", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Test table creation
		_, err := db.Exec(`
			CREATE TABLE test_table (
				id SERIAL PRIMARY KEY,
				name TEXT NOT NULL
			)
		`)
		require.NoError(t, err)

		// Test insertion
		_, err = db.Exec("INSERT INTO test_table (name) VALUES ($1)", "test")
		require.NoError(t, err)

		// Test query
		var name string
		err = db.QueryRow("SELECT name FROM test_table WHERE id = 1").Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, "test", name)
	})

	// Test concurrent operations
	t.Run("concurrent operations", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Get the current schema name
		var schemaName string
		err := db.QueryRow("SELECT current_schema").Scan(&schemaName)
		require.NoError(t, err)
		require.Equal(t, currentTestDBName, schemaName, "Schema should be set to the test schema")

		// Create test table with fully qualified name
		tableName := fmt.Sprintf("%s.concurrent_test", pq.QuoteIdentifier(schemaName))
		_, err = db.Exec(fmt.Sprintf(`
			CREATE TABLE %s (
				id SERIAL PRIMARY KEY,
				value INTEGER
			)
		`, tableName))
		require.NoError(t, err)

		// Run concurrent insertions
		var wg sync.WaitGroup
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(val int) {
				defer wg.Done()
				_, execErr := db.Exec(fmt.Sprintf("INSERT INTO %s (value) VALUES ($1)", tableName), val)
				assert.NoError(t, execErr)
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Verify all insertions
		var count int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 10, count)
	})

	// Test transaction rollback
	t.Run("transaction rollback", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Create test table
		_, err := db.Exec(`
			CREATE TABLE rollback_test (
				id SERIAL PRIMARY KEY,
				value TEXT NOT NULL
			)
		`)
		require.NoError(t, err)

		// Start transaction
		tx, err := db.Begin()
		require.NoError(t, err)

		// Insert data
		_, err = tx.Exec("INSERT INTO rollback_test (value) VALUES ($1)", "test")
		require.NoError(t, err)

		// Rollback transaction
		safeClose(tx)

		// Verify no data was inserted
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM rollback_test").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	// Test connection timeout
	t.Run("connection timeout", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Set a very short timeout
		db.SetConnMaxLifetime(1 * time.Millisecond)

		// Wait for connection to timeout
		time.Sleep(10 * time.Millisecond)

		// Verify connection still works after timeout
		err := db.Ping()
		assert.NoError(t, err)
	})
}

func TestDatabaseErrors(t *testing.T) {
	// Test invalid connection string
	t.Run("invalid connection", func(t *testing.T) {
		// Use a non-existent host to cause a connection error
		origHost := os.Getenv("POSTGRES_HOST")
		os.Setenv("POSTGRES_HOST", "non-existent-host")
		defer os.Setenv("POSTGRES_HOST", origHost)

		// The function should now panic with a connection error
		assert.Panics(t, func() {
			db := SetupTestDB(t)
			if db != nil {
				safeClose(db)
			}
		})
	})

	// Test migration failures
	t.Run("migration failure", func(t *testing.T) {
		// Save current test DB name and restore after test
		oldName := currentTestDBName
		testDBName := fmt.Sprintf("test_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))

		// Connect to postgres to create test database
		db, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer safeClose(db)

		// Create test database
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(testDBName)))
		require.NoError(t, err)

		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(testDBName)))
		require.NoError(t, err)

		// Create an invalid migration state
		testDB, err := sql.Open("postgres", getPostgresConnection(testDBName))
		require.NoError(t, err)
		defer safeClose(testDB)

		// Create schema_migrations table with a dirty state
		_, err = testDB.Exec(`
			CREATE TABLE schema_migrations (
				version bigint NOT NULL,
				dirty boolean NOT NULL
			);
			INSERT INTO schema_migrations (version, dirty) VALUES (9999, true);
		`)
		require.NoError(t, err)

		// Close all connections to the test database
		_, err = db.Exec(fmt.Sprintf(`
			SELECT pg_terminate_backend(pid) 
			FROM pg_stat_activity 
			WHERE datname = %s AND pid != pg_backend_pid()
		`, pq.QuoteLiteral(testDBName)))
		require.NoError(t, err)

		// Set the current test database name
		currentTestDBName = testDBName

		// Ensure cleanup happens after we're done
		defer func() {
			cleanupDatabase(testDBName)
			currentTestDBName = oldName
		}()

		// Verify that SetupTestDB handles the dirty state gracefully
		var setupDB *sql.DB
		assert.NotPanics(t, func() {
			setupDB = SetupTestDB(t)
		})

		// Verify that either:
		// 1. The setup failed and returned nil
		// 2. The setup succeeded and cleaned up the dirty state
		if setupDB != nil {
			defer safeClose(setupDB)

			// Verify the schema_migrations table is in a clean state
			var version int64
			var dirty bool
			err = setupDB.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
			require.NoError(t, err)
			assert.False(t, dirty, "Database should be in a clean state after setup")
			assert.NotEqual(t, int64(9999), version, "Migration version should be reset")
		}
	})

	// Test database cleanup
	t.Run("cleanup after failed setup", func(t *testing.T) {
		// Save current test DB name and restore after test
		oldName := currentTestDBName
		testDBName := fmt.Sprintf("test_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))

		// Connect to postgres to create test database
		db, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer safeClose(db)

		// Create test database
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(testDBName)))
		require.NoError(t, err)

		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(testDBName)))
		require.NoError(t, err)

		// Create an invalid table to cause migration failure
		testDB, err := sql.Open("postgres", getPostgresConnection(testDBName))
		require.NoError(t, err)
		defer safeClose(testDB)

		// Create an invalid table that will conflict with migrations
		_, err = testDB.Exec(`
			CREATE TABLE users (
				id TEXT PRIMARY KEY -- This conflicts with the UUID type in migrations
			);
		`)
		require.NoError(t, err)

		// Close all connections to the test database
		_, err = db.Exec(fmt.Sprintf(`
			SELECT pg_terminate_backend(pid) 
			FROM pg_stat_activity 
			WHERE datname = %s AND pid != pg_backend_pid()
		`, pq.QuoteLiteral(testDBName)))
		require.NoError(t, err)

		// Set the current test database name
		currentTestDBName = testDBName

		// Ensure cleanup happens after we're done
		defer func() {
			cleanupDatabase(testDBName)
			currentTestDBName = oldName
		}()

		// Verify that SetupTestDB handles the migration failure gracefully
		var setupDB *sql.DB
		assert.NotPanics(t, func() {
			setupDB = SetupTestDB(t)
		})

		// Verify that either:
		// 1. The setup failed and returned nil
		// 2. The setup succeeded and fixed the schema
		if setupDB != nil {
			defer safeClose(setupDB)

			// Verify the users table has the correct schema
			var dataType string
			err = setupDB.QueryRow(`
				SELECT data_type 
				FROM information_schema.columns 
				WHERE table_name = 'users' AND column_name = 'id'
			`).Scan(&dataType)
			require.NoError(t, err)
			assert.Equal(t, "uuid", strings.ToLower(dataType), "Users table should have UUID type for id column")
		}
	})

	// Test concurrent database operations under contention
	t.Run("concurrent database operations", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Get the current schema
		var schemaName string
		err := db.QueryRow("SELECT current_schema").Scan(&schemaName)
		require.NoError(t, err)

		// Create test table with a unique constraint
		tableName := fmt.Sprintf("%s.concurrent_error_test", pq.QuoteIdentifier(schemaName))
		_, err = db.Exec(fmt.Sprintf(`
			CREATE TABLE %s (
				id SERIAL PRIMARY KEY,
				value TEXT UNIQUE,
				version INTEGER DEFAULT 1
			)
		`, tableName))
		require.NoError(t, err)

		// Insert a base record
		_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (value) VALUES ('base-value')", tableName))
		require.NoError(t, err)

		// Run concurrent operations with optimistic locking
		var wg sync.WaitGroup
		var mutex sync.Mutex
		successCount := 0

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// Use a separate connection for each goroutine
				conn, err := sql.Open("postgres", getPostgresConnection(""))
				if err != nil {
					t.Logf("Error creating connection: %v", err)
					return
				}
				defer safeClose(conn)

				// Set search path for the connection
				_, err = conn.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schemaName)))
				if err != nil {
					t.Logf("Error setting search path: %v", err)
					return
				}

				// Try to update with optimistic locking
				for retry := 0; retry < 3; retry++ {
					// Start a transaction
					tx, err := conn.Begin()
					if err != nil {
						continue
					}

					// Get current version
					var version int
					err = tx.QueryRow(fmt.Sprintf("SELECT version FROM %s WHERE value = 'base-value' FOR UPDATE", tableName)).Scan(&version)
					if err != nil {
						tx.Rollback()
						continue
					}

					// Try to update with version check
					result, err := tx.Exec(
						fmt.Sprintf("UPDATE %s SET value = $1, version = version + 1 WHERE value = 'base-value' AND version = $2",
							tableName),
						fmt.Sprintf("value-%d", i),
						version,
					)

					if err != nil {
						tx.Rollback()
						time.Sleep(10 * time.Millisecond)
						continue
					}

					rows, err := result.RowsAffected()
					if err != nil || rows != 1 {
						tx.Rollback()
						continue
					}

					// Commit the transaction
					if err := tx.Commit(); err == nil {
						mutex.Lock()
						successCount++
						mutex.Unlock()
						return
					}
				}
				t.Logf("Failed to update value-%d after retries", i)
			}(i)
		}

		wg.Wait()

		// Only one transaction should succeed due to optimistic locking
		assert.Equal(t, 1, successCount, "Only one update should succeed due to contention")
	})
}
