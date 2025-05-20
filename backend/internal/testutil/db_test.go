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

		// Get the schema name through a simple, direct query
		var schemaName string
		err = db.QueryRow("SELECT current_schema()").Scan(&schemaName)
		require.NoError(t, err)
		require.Equal(t, db.GetSchemaName(), schemaName, "Schema should match expected value")

		// Verify migrations were applied by checking for a known table
		// Use a separate query to avoid multiple statements in a prepared statement
		var exists bool
		tableCheckQuery := "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = $1 AND table_name = 'users')"
		err = db.QueryRow(tableCheckQuery, schemaName).Scan(&exists)
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

		// Verify both databases are different - using simple, direct queries
		var db1Schema, db2Schema string
		err := db1.QueryRow("SELECT current_schema()").Scan(&db1Schema)
		require.NoError(t, err)

		err = db2.QueryRow("SELECT current_schema()").Scan(&db2Schema)
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
		schemaName := db.GetSchemaName()
		require.NotEmpty(t, schemaName)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test table creation with schema qualification
		_, err := db.SafeExecWithSchemaContext(ctx,
			"CREATE TABLE test_table (id SERIAL PRIMARY KEY, name TEXT NOT NULL)")
		require.NoError(t, err)

		// Test insertion with schema qualification
		_, err = db.SafeExecWithSchemaContext(ctx,
			"INSERT INTO test_table (name) VALUES ($1)", "test")
		require.NoError(t, err)

		// Test query with schema qualification
		var name string
		err = db.SafeQueryRowWithSchemaContext(ctx,
			"SELECT name FROM test_table WHERE id = 1").Scan(&name)
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
		err := db.QueryRow("SELECT current_schema()").Scan(&schemaName)
		require.NoError(t, err)
		assert.Equal(t, db.GetSchemaName(), schemaName, "Schema should be set to the test schema")

		// Create test table
		_, err = db.Exec("CREATE TABLE concurrent_test (id SERIAL PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		// Get the schema context for this test
		schemaCtx := GetSchemaContext(schemaName)
		require.NotNil(t, schemaCtx, "Schema context should exist")

		// Run concurrent insertions using a more controlled approach
		var wg sync.WaitGroup
		errCh := make(chan error, 10) // Channel to collect errors

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()

				// Create a connection that's fully aware of the schema
				insertDB, err := sql.Open("postgres", getPostgresConnection("postgres"))
				if err != nil {
					errCh <- err
					return
				}
				defer safeClose(insertDB)

				// Set the search path explicitly
				_, err = insertDB.Exec(fmt.Sprintf("SET search_path TO %s, public", pq.QuoteIdentifier(schemaName)))
				if err != nil {
					errCh <- err
					return
				}

				// Now insert using the schema-qualified table name to be extra safe
				_, execErr := insertDB.Exec(fmt.Sprintf("INSERT INTO %s.concurrent_test (value) VALUES ($1)",
					pq.QuoteIdentifier(schemaName)), val)
				if execErr != nil {
					errCh <- execErr
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(errCh)

		// Check for any errors
		for err := range errCh {
			require.NoError(t, err, "Concurrent insertions should succeed")
		}

		// Verify all insertions
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM concurrent_test").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 10, count)
	})

	// Test transaction rollback
	t.Run("transaction rollback", func(t *testing.T) {
		// Create a fresh database connection for this test
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Get the schema name from the DB connection
		schemaName := db.GetSchemaName()
		require.NotEmpty(t, schemaName)

		// Create a test table
		_, err := db.Exec("CREATE TABLE tx_test (id SERIAL PRIMARY KEY, value TEXT)")
		require.NoError(t, err)

		// Ensure we have a good connection for the transaction
		err = db.Ping()
		require.NoError(t, err, "Connection should be valid before starting transaction")

		// Begin a transaction with explicit timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		// Always clean up the transaction
		defer func() {
			if tx != nil {
				_ = tx.Rollback() // No-op if already committed or rolled back
			}
		}()

		// Insert test data in transaction with explicit schema
		_, err = tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s.tx_test (value) VALUES ($1)",
			pq.QuoteIdentifier(schemaName)), "test value")
		require.NoError(t, err)

		// Verify data exists within the transaction
		var txCount int
		err = tx.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s.tx_test",
			pq.QuoteIdentifier(schemaName))).Scan(&txCount)
		require.NoError(t, err)
		assert.Equal(t, 1, txCount, "Data should be visible within the transaction")

		// Rollback transaction
		err = tx.Rollback()
		require.NoError(t, err)
		tx = nil // Mark as handled so defer doesn't try to rollback again

		// Verify data wasn't inserted (using a fresh query to avoid connection reuse issues)
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tx_test").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "No rows should exist after rollback")
	})

	// Test connection timeout
	t.Run("connection timeout", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)
		defer TeardownTestDB(t, db)

		// Create a very short timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Sleep to ensure timeout
		time.Sleep(1 * time.Millisecond)

		// This should fail with timeout
		_, err := db.ExecContext(ctx, "SELECT pg_sleep(1)")
		assert.Error(t, err, "Query should timeout")
		assert.Contains(t, err.Error(), "context", "Error should be context-related")
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

// TestSchemaHandling tests schema handling in transactions and DB operations
func TestSchemaHandling(t *testing.T) {
	// Create a test database
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer TeardownTestDB(t, db)

	// Get the schema name for verification
	schemaName := db.GetSchemaName()
	require.NotEmpty(t, schemaName)

	// Verify the schema was set correctly
	var currentSchema string
	err := db.QueryRow("SELECT current_schema()").Scan(&currentSchema)
	require.NoError(t, err)
	assert.Equal(t, schemaName, currentSchema, "Schema should be correctly set")

	// Test transaction schema handling
	t.Run("transaction schema handling", func(t *testing.T) {
		// Ensure we have a valid connection first
		err = db.Ping()
		require.NoError(t, err, "Connection should be valid before starting transaction")

		// Create a context with timeout to prevent test hangs
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Begin a transaction with the context
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		// Ensure tx is always either committed or rolled back
		defer func() {
			if tx != nil {
				_ = tx.Rollback() // This is a no-op if already committed
			}
		}()

		// Explicitly set the schema in the transaction
		_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL search_path TO %s, public",
			pq.QuoteIdentifier(schemaName)))
		require.NoError(t, err)

		// Verify transaction schema context
		var txSchema string
		err = tx.QueryRowContext(ctx, "SELECT current_schema()").Scan(&txSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, txSchema, "Transaction should have correct schema")

		// Execute a simple query in the transaction with schema qualification
		_, err = tx.ExecContext(ctx, "CREATE TEMPORARY TABLE test_tx (id INT)")
		require.NoError(t, err)

		// Commit the transaction
		err = tx.Commit()
		require.NoError(t, err)
		tx = nil // Mark as handled
	})

	// Test transaction rollback schema handling
	t.Run("transaction rollback schema handling", func(t *testing.T) {
		// Ensure we have a valid connection first
		err = db.Ping()
		require.NoError(t, err, "Connection should be valid before starting transaction")

		// Create a context with timeout to prevent test hangs
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Begin a transaction with the context
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		// Ensure tx is always rolled back if not committed
		defer func() {
			if tx != nil {
				_ = tx.Rollback() // This is a no-op if already rolled back
			}
		}()

		// Explicitly set the schema in the transaction
		_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL search_path TO %s, public",
			pq.QuoteIdentifier(schemaName)))
		require.NoError(t, err)

		// Verify transaction schema context
		var txSchema string
		err = tx.QueryRowContext(ctx, "SELECT current_schema()").Scan(&txSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, txSchema, "Transaction should have correct schema")

		// Execute a query that will be rolled back, using schema qualification
		_, err = tx.ExecContext(ctx, "CREATE TEMPORARY TABLE test_rollback (id INT)")
		require.NoError(t, err)

		// Explicitly roll back the transaction
		err = tx.Rollback()
		require.NoError(t, err)
		tx = nil // Mark as handled

		// Verify connection still works after rollback
		var postRollbackSchema string
		err = db.QueryRowContext(ctx, "SELECT current_schema()").Scan(&postRollbackSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, postRollbackSchema, "Schema should still be correct after rollback")
	})

	// Test schema persistence across multiple queries
	t.Run("schema persistence", func(t *testing.T) {
		// Create a table in the schema
		_, err := db.Exec("CREATE TABLE test_persistence (id SERIAL PRIMARY KEY, name TEXT)")
		require.NoError(t, err)

		// Insert some data
		_, err = db.Exec("INSERT INTO test_persistence (name) VALUES ($1)", "test")
		require.NoError(t, err)

		// Query the data
		var name string
		err = db.QueryRow("SELECT name FROM test_persistence WHERE id = 1").Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, "test", name)

		// Verify schema is still correct
		var querySchema string
		err = db.QueryRow("SELECT current_schema()").Scan(&querySchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, querySchema, "Schema should be maintained across operations")
	})
}

// VerifySchemaContext ensures the correct schema is set for test operations
func VerifySchemaContext(t *testing.T, db interface{}) string {
	t.Helper()

	// Get appropriate DB based on type
	var sqlDB *sql.DB
	var expectedSchema string

	switch d := db.(type) {
	case *sql.DB:
		sqlDB = d
		expectedSchema = GetCurrentTestSchema()
	case *SchemaDB:
		sqlDB = d.DB
		expectedSchema = d.GetSchemaName()
	default:
		t.Fatalf("Unsupported DB type: %T", db)
	}

	if expectedSchema == "" {
		t.Fatal("No expected schema found for verification")
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Query for current schema
	var currentSchema string
	err := sqlDB.QueryRowContext(ctx, "SELECT current_schema()").Scan(&currentSchema)
	if err != nil {
		t.Fatalf("Error verifying schema context: %v", err)
	}

	// Verify and print for debugging
	if currentSchema != expectedSchema {
		t.Fatalf("Incorrect schema: got %s, expected %s", currentSchema, expectedSchema)
	}

	t.Logf("Schema verification successful: %s", currentSchema)
	return currentSchema
}
