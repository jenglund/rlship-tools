package testutil

import (
	"context"
	"database/sql"
	"fmt"
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

		// Get schema name from first database
		db1Schema := db1.GetSchemaName()
		require.NotEmpty(t, db1Schema)

		db2 := SetupTestDB(t)
		require.NotNil(t, db2)
		defer TeardownTestDB(t, db2)

		// Get schema name from second database
		db2Schema := db2.GetSchemaName()
		require.NotEmpty(t, db2Schema)

		// Verify both databases have different schema names
		assert.NotEqual(t, db1Schema, db2Schema, "each setup should create a unique schema")

		// Verify both schemas are set correctly in their respective connections
		var currentSchema1 string
		err := db1.QueryRow("SELECT current_schema()").Scan(&currentSchema1)
		require.NoError(t, err)
		assert.Equal(t, db1Schema, currentSchema1, "Schema should be set correctly for db1")

		var currentSchema2 string
		err = db2.QueryRow("SELECT current_schema()").Scan(&currentSchema2)
		require.NoError(t, err)
		assert.Equal(t, db2Schema, currentSchema2, "Schema should be set correctly for db2")
	})
}

func TestTeardownTestDB(t *testing.T) {
	// Test successful teardown
	t.Run("successful teardown", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)

		// Get the schema name before teardown
		schemaName := db.GetSchemaName()
		require.NotEmpty(t, schemaName)

		// Verify schema exists before teardown
		pgdb, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer safeClose(pgdb)

		var existsBefore bool
		err = pgdb.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.schemata 
				WHERE schema_name = $1
			)
		`, schemaName).Scan(&existsBefore)
		require.NoError(t, err)
		assert.True(t, existsBefore, "schema should exist before teardown")

		// Perform teardown
		TeardownTestDB(t, db)

		// Verify schema no longer exists by checking information_schema
		pgdb2, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer safeClose(pgdb2)

		var exists bool
		err = pgdb2.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.schemata 
				WHERE schema_name = $1
			)
		`, schemaName).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "schema should not exist after teardown")

		// Verify schema context was removed
		SchemaContextMutex.Lock()
		_, contextExists := SchemaContexts[schemaName]
		SchemaContextMutex.Unlock()
		assert.False(t, contextExists, "schema context should be removed after teardown")
	})

	// Test teardown with active connections
	t.Run("teardown with active connections", func(t *testing.T) {
		db := SetupTestDB(t)
		require.NotNil(t, db)

		// Get the schema name
		schemaName := db.GetSchemaName()
		require.NotEmpty(t, schemaName)

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

		// Verify the table exists before teardown
		var tableExists bool
		err = db.QueryRow(fmt.Sprintf(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = '%s'
				AND table_name = 'teardown_test'
			)
		`, schemaName)).Scan(&tableExists)
		require.NoError(t, err)
		assert.True(t, tableExists, "table should exist before teardown")

		// Perform teardown
		TeardownTestDB(t, db)

		// Verify schema no longer exists
		pgdb, err := sql.Open("postgres", getPostgresConnection(""))
		require.NoError(t, err)
		defer safeClose(pgdb)

		var schemaExists bool
		err = pgdb.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.schemata 
				WHERE schema_name = $1
			)
		`, schemaName).Scan(&schemaExists)
		require.NoError(t, err)
		assert.False(t, schemaExists, "schema should not exist after teardown")

		// Verify connections are closed by attempting queries to the dropped schema's table
		err = db2.QueryRow(fmt.Sprintf("SELECT * FROM %s.teardown_test LIMIT 1", pq.QuoteIdentifier(schemaName))).Scan(&count)
		assert.Error(t, err, "query should fail after teardown")
		err = db3.QueryRow(fmt.Sprintf("SELECT * FROM %s.teardown_test LIMIT 1", pq.QuoteIdentifier(schemaName))).Scan(&count)
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

		// Get the schema name from the connection
		schemaName := db.GetSchemaName()
		require.NotEmpty(t, schemaName)

		// Verify schema is set correctly
		var currentSchema string
		err := db.QueryRow("SELECT current_schema()").Scan(&currentSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, currentSchema, "Schema should be set to the test schema")

		// Create test table with explicit schema
		_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s.concurrent_test (id SERIAL PRIMARY KEY, value INTEGER)",
			pq.QuoteIdentifier(schemaName)))
		require.NoError(t, err)

		// Run concurrent insertions with better synchronization
		var wg sync.WaitGroup
		errCh := make(chan error, 10)  // Channel to collect errors
		resultCh := make(chan int, 10) // Channel to collect successful insertions

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()

				// Create a new connection for this goroutine
				connString := getPostgresConnection("")
				insertDB, err := sql.Open("postgres", connString)
				if err != nil {
					errCh <- fmt.Errorf("connection error (value %d): %w", val, err)
					return
				}
				defer safeClose(insertDB)

				// Set the search path explicitly for this connection
				_, err = insertDB.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schemaName)))
				if err != nil {
					errCh <- fmt.Errorf("search path error (value %d): %w", val, err)
					return
				}

				// Do a read to verify connection is set up properly
				var testSchema string
				err = insertDB.QueryRow("SELECT current_schema()").Scan(&testSchema)
				if err != nil {
					errCh <- fmt.Errorf("schema verification error (value %d): %w", val, err)
					return
				}

				if testSchema != schemaName {
					errCh <- fmt.Errorf("schema mismatch (value %d): got %s, expected %s", val, testSchema, schemaName)
					return
				}

				// Now insert with explicit schema qualification
				insertQuery := fmt.Sprintf("INSERT INTO %s.concurrent_test (value) VALUES ($1) RETURNING id",
					pq.QuoteIdentifier(schemaName))
				var insertedID int
				err = insertDB.QueryRow(insertQuery, val).Scan(&insertedID)
				if err != nil {
					errCh <- fmt.Errorf("insert error (value %d): %w", val, err)
					return
				}

				// Record successful insertion
				resultCh <- val
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(errCh)
		close(resultCh)

		// Check for any errors
		errors := make([]error, 0)
		for err := range errCh {
			errors = append(errors, err)
		}

		// Show detailed errors if any
		for _, err := range errors {
			t.Logf("Concurrent operation error: %v", err)
		}
		require.Empty(t, errors, "There should be no errors in concurrent operations")

		// Count successful insertions
		successCount := 0
		successfulValues := make([]int, 0, 10)
		for val := range resultCh {
			successCount++
			successfulValues = append(successfulValues, val)
		}

		// Verify all insertions from database
		var count int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s.concurrent_test",
			pq.QuoteIdentifier(schemaName))).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 10, count, "All 10 insertions should succeed")
		assert.Equal(t, 10, successCount, "All 10 goroutines should have reported success")
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

		// Create a context with a longer timeout to prevent test hangs
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

		// Create a unique table name to ensure tests don't conflict
		tableName := fmt.Sprintf("test_tx_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))

		// Explicitly set the schema in the transaction
		_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL search_path TO %s",
			pq.QuoteIdentifier(schemaName)))
		require.NoError(t, err)

		// Verify transaction schema context
		var txSchema string
		err = tx.QueryRowContext(ctx, "SELECT current_schema()").Scan(&txSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, txSchema, "Transaction should have correct schema")

		// Execute a simple query in the transaction with schema qualification
		_, err = tx.ExecContext(ctx, fmt.Sprintf("CREATE TABLE %s.%s (id INT)",
			pq.QuoteIdentifier(schemaName), pq.QuoteIdentifier(tableName)))
		require.NoError(t, err)

		// Verify table exists within transaction
		var tableExists bool
		err = tx.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = '%s' AND table_name = '%s'
			)`, schemaName, tableName)).Scan(&tableExists)
		require.NoError(t, err)
		assert.True(t, tableExists, "Table should exist within transaction")

		// Commit the transaction
		err = tx.Commit()
		require.NoError(t, err)
		tx = nil // Mark as handled

		// Verify schema is still set after transaction
		var postTxSchema string
		err = db.QueryRowContext(ctx, "SELECT current_schema()").Scan(&postTxSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, postTxSchema, "Schema should remain correct after transaction")

		// Verify table exists after transaction is committed
		var tableExistsAfterCommit bool
		err = db.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = '%s' AND table_name = '%s'
			)`, schemaName, tableName)).Scan(&tableExistsAfterCommit)
		require.NoError(t, err)
		assert.True(t, tableExistsAfterCommit, "Table should exist after transaction commit")
	})

	// Test transaction rollback schema handling
	t.Run("transaction rollback schema handling", func(t *testing.T) {
		// Ensure we have a valid connection first
		err = db.Ping()
		require.NoError(t, err, "Connection should be valid before starting transaction")

		// Create a context with timeout to prevent test hangs
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Verify schema is set correctly before starting transaction
		var preSchema string
		err = db.QueryRowContext(ctx, "SELECT current_schema()").Scan(&preSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, preSchema, "Schema should be correct before transaction")

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
		_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL search_path TO %s",
			pq.QuoteIdentifier(schemaName)))
		require.NoError(t, err)

		// Verify transaction schema context
		var txSchema string
		err = tx.QueryRowContext(ctx, "SELECT current_schema()").Scan(&txSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, txSchema, "Transaction should have correct schema")

		// Create a unique table name to ensure we're testing the right transaction
		tableName := fmt.Sprintf("test_rollback_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))

		// Execute a query that will be rolled back, using schema qualification
		_, err = tx.ExecContext(ctx, fmt.Sprintf(
			"CREATE TABLE %s.%s (id INT)",
			pq.QuoteIdentifier(schemaName),
			pq.QuoteIdentifier(tableName)))
		require.NoError(t, err)

		// Verify table exists within transaction
		var tableExistsInTx bool
		err = tx.QueryRowContext(ctx, fmt.Sprintf(`
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_schema = '%s' AND table_name = '%s'
            )`, schemaName, tableName)).Scan(&tableExistsInTx)
		require.NoError(t, err)
		assert.True(t, tableExistsInTx, "Table should exist within transaction")

		// Explicitly roll back the transaction
		err = tx.Rollback()
		require.NoError(t, err)
		tx = nil // Mark as handled

		// Verify table doesn't exist outside transaction after rollback
		var tableExistsAfterRollback bool
		err = db.QueryRowContext(ctx, fmt.Sprintf(`
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_schema = '%s' AND table_name = '%s'
            )`, schemaName, tableName)).Scan(&tableExistsAfterRollback)
		require.NoError(t, err)
		assert.False(t, tableExistsAfterRollback, "Table should not exist after rollback")

		// Verify schema is still correct after rollback
		var postRollbackSchema string
		err = db.QueryRowContext(ctx, "SELECT current_schema()").Scan(&postRollbackSchema)
		require.NoError(t, err)
		assert.Equal(t, schemaName, postRollbackSchema, "Schema should still be correct after rollback")
	})

	// Test schema persistence across multiple queries
	t.Run("schema persistence", func(t *testing.T) {
		// Create a unique table name for this test
		tableName := fmt.Sprintf("test_persistence_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))

		// Create a table in the schema with explicit schema qualification
		_, err := db.Exec(fmt.Sprintf("CREATE TABLE %s.%s (id SERIAL PRIMARY KEY, name TEXT)",
			pq.QuoteIdentifier(schemaName), pq.QuoteIdentifier(tableName)))
		require.NoError(t, err)

		// Verify table was created
		var tableExists bool
		err = db.QueryRow(fmt.Sprintf(`
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_schema = '%s' AND table_name = '%s'
            )`, schemaName, tableName)).Scan(&tableExists)
		require.NoError(t, err)
		assert.True(t, tableExists, "Table should have been created")

		// Insert some data with schema qualification
		_, err = db.Exec(fmt.Sprintf("INSERT INTO %s.%s (name) VALUES ($1)",
			pq.QuoteIdentifier(schemaName), pq.QuoteIdentifier(tableName)), "test")
		require.NoError(t, err)

		// Query the data with schema qualification
		var name string
		err = db.QueryRow(fmt.Sprintf("SELECT name FROM %s.%s WHERE id = 1",
			pq.QuoteIdentifier(schemaName), pq.QuoteIdentifier(tableName))).Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, "test", name)

		// Verify schema is still correct after operations
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
