package testutil

import (
	"database/sql"
	"fmt"
	"strings"
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
				WHERE table_schema = 'public' 
				AND table_name = 'users'
			)
		`).Scan(&exists)
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
		var db1Name, db2Name string
		err := db1.QueryRow("SELECT current_database()").Scan(&db1Name)
		require.NoError(t, err)
		err = db2.QueryRow("SELECT current_database()").Scan(&db2Name)
		require.NoError(t, err)

		assert.NotEqual(t, db1Name, db2Name, "each setup should create a unique database")
	})
}

func TestTeardownTestDB(t *testing.T) {
	// Test successful teardown
	t.Run("successful teardown", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)

		// Get the database name before teardown
		var dbName string
		err := db.QueryRow("SELECT current_database()").Scan(&dbName)
		require.NoError(t, err)

		// Perform teardown
		TeardownTestDB(t, db)

		// Verify database no longer exists by connecting to postgres and checking
		pgdb, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer pgdb.Close()

		var exists bool
		err = pgdb.QueryRow(`
			SELECT EXISTS (
				SELECT FROM pg_database 
				WHERE datname = $1
			)
		`, dbName).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "database should not exist after teardown")
	})

	// Test teardown with active connections
	t.Run("teardown with active connections", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)

		// Create additional connections
		db2, err := sql.Open("postgres", getPostgresConnection(currentTestDBName))
		require.NoError(t, err)
		defer db2.Close()

		db3, err := sql.Open("postgres", getPostgresConnection(currentTestDBName))
		require.NoError(t, err)
		defer db3.Close()

		// Perform some queries to ensure connections are active
		var count int
		err = db2.QueryRow("SELECT 1").Scan(&count)
		require.NoError(t, err)
		err = db3.QueryRow("SELECT 1").Scan(&count)
		require.NoError(t, err)

		// Perform teardown
		TeardownTestDB(t, db)

		// Verify connections are closed by attempting queries
		err = db2.QueryRow("SELECT 1").Scan(&count)
		assert.Error(t, err, "query should fail after teardown")
		err = db3.QueryRow("SELECT 1").Scan(&count)
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

		// Create test table
		_, err := db.Exec(`
			CREATE TABLE concurrent_test (
				id SERIAL PRIMARY KEY,
				value INTEGER
			)
		`)
		require.NoError(t, err)

		// Run concurrent insertions
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(val int) {
				_, err := db.Exec("INSERT INTO concurrent_test (value) VALUES ($1)", val)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all insertions
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM concurrent_test").Scan(&count)
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
		err = tx.Rollback()
		require.NoError(t, err)

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
		// Save current test DB name and restore after test
		oldName := currentTestDBName
		defer func() { currentTestDBName = oldName }()

		// Use an invalid database name that should cause connection failure
		currentTestDBName = "invalid/db"

		// Should panic when trying to create database with invalid name
		assert.Panics(t, func() {
			SetupTestDB(t)
		})

		// Verify database was cleaned up
		db, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer db.Close()

		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", currentTestDBName).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "Database should have been cleaned up after failed setup")
	})

	// Test migration failures
	t.Run("migration failure", func(t *testing.T) {
		// Save current test DB name and restore after test
		oldName := currentTestDBName
		testDBName := fmt.Sprintf("test_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))

		// Connect to postgres to create test database
		db, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer db.Close()

		// Create test database
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(testDBName)))
		require.NoError(t, err)

		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(testDBName)))
		require.NoError(t, err)

		// Create an invalid migration state
		testDB, err := sql.Open("postgres", getPostgresConnection(testDBName))
		require.NoError(t, err)
		defer testDB.Close()

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
			defer setupDB.Close()

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
		defer db.Close()

		// Create test database
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(testDBName)))
		require.NoError(t, err)

		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(testDBName)))
		require.NoError(t, err)

		// Create an invalid table to cause migration failure
		testDB, err := sql.Open("postgres", getPostgresConnection(testDBName))
		require.NoError(t, err)
		defer testDB.Close()

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
			defer setupDB.Close()

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

	// Test concurrent database operations
	t.Run("concurrent database operations", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Create test table
		_, err := db.Exec(`
			CREATE TABLE concurrent_write_test (
				id SERIAL PRIMARY KEY,
				value TEXT
			)
		`)
		require.NoError(t, err)

		// Run concurrent writes with transaction retries
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(val int) {
				defer func() { done <- true }()

				maxRetries := 3
				for retry := 0; retry < maxRetries; retry++ {
					tx, err := db.Begin()
					require.NoError(t, err)

					_, err = tx.Exec("INSERT INTO concurrent_write_test (value) VALUES ($1)", fmt.Sprintf("value-%d", val))
					if err != nil {
						tx.Rollback()
						if retry < maxRetries-1 {
							time.Sleep(10 * time.Millisecond)
							continue
						}
						t.Errorf("Failed to insert after %d retries: %v", maxRetries, err)
						return
					}

					err = tx.Commit()
					if err != nil {
						if retry < maxRetries-1 {
							time.Sleep(10 * time.Millisecond)
							continue
						}
						t.Errorf("Failed to commit after %d retries: %v", maxRetries, err)
						return
					}
					break
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify all insertions
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM concurrent_write_test").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, numGoroutines, count)
	})
}
