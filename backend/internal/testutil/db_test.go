package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

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

		// Get the schema name for explicit schema referencing
		var schemaName string
		err = db.QueryRow("SELECT current_schema").Scan(&schemaName)
		require.NoError(t, err)

		// Verify migrations were applied by checking for a known table
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = $1
				AND table_name = 'users'
			)
		`, schemaName).Scan(&exists)
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

		// Get the schema name for explicit schema referencing
		var schemaName string
		err := db.QueryRow("SELECT current_schema").Scan(&schemaName)
		require.NoError(t, err)

		// Use fully qualified table name to avoid schema issues
		tableName := fmt.Sprintf("%s.test_table", pq.QuoteIdentifier(schemaName))

		// Test table creation - avoid SET commands
		_, err = db.Exec(fmt.Sprintf(`
			CREATE TABLE %s (
				id SERIAL PRIMARY KEY,
				name TEXT NOT NULL
			)
		`, tableName))
		require.NoError(t, err)

		// Test insertion - avoid SET commands
		_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", tableName), "test")
		require.NoError(t, err)

		// Test query - avoid SET commands
		var name string
		err = db.QueryRow(fmt.Sprintf("SELECT name FROM %s WHERE id = 1", tableName)).Scan(&name)
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
		// Create a fresh database connection for this test
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Get the schema name for explicit schema referencing
		var schemaName string
		err := db.QueryRow("SELECT current_schema").Scan(&schemaName)
		require.NoError(t, err)

		// Use fully qualified table name to avoid schema issues
		tableName := fmt.Sprintf("%s.rollback_test", pq.QuoteIdentifier(schemaName))

		// Create test table
		_, err = db.Exec(fmt.Sprintf(`
			CREATE TABLE %s (
				id SERIAL PRIMARY KEY,
				value TEXT NOT NULL
			)
		`, tableName))
		require.NoError(t, err)

		// Verify table is empty
		var initialCount int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&initialCount)
		require.NoError(t, err)
		assert.Equal(t, 0, initialCount, "Table should start empty")

		// Use a single operation with ExecContext to do the insert in a transaction and then rollback
		ctx := context.Background()
		sqlDB := UnwrapDB(db)

		// Start an explicit transaction
		tx, err := sqlDB.BeginTx(ctx, nil)
		require.NoError(t, err)

		// Insert test data
		_, err = tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (value) VALUES ($1)", tableName), "test")
		require.NoError(t, err)

		// Explicitly rollback the transaction
		err = tx.Rollback()
		require.NoError(t, err)

		// Verify no data is in the table after rollback
		// Use a completely fresh query to avoid connection issues
		var finalCount int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&finalCount)
		require.NoError(t, err)
		assert.Equal(t, 0, finalCount, "No data should be inserted after rollback")
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
	t.Run("database operation errors", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		sqlDB := GetDB(db)

		// Nonexistent table
		_, err := sqlDB.Exec("SELECT * FROM nonexistent_table")
		assert.Error(t, err)

		// Invalid SQL syntax
		_, err = sqlDB.Exec("SELECT FROM WHERE")
		assert.Error(t, err)

		// Invalid column name
		_, err = sqlDB.Exec("CREATE TABLE error_test (id SERIAL PRIMARY KEY)")
		assert.NoError(t, err)

		_, err = sqlDB.Exec("INSERT INTO error_test (nonexistent_column) VALUES (1)")
		assert.Error(t, err)
	})
}
