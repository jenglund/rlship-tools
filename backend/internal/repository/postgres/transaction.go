package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/lib/pq"
)

const (
	// Error codes from: https://www.postgresql.org/docs/current/errcodes-appendix.html
	errDeadlockDetected = "40P01"
	errLockTimeout      = "55P03"

	// Default timeout to prevent hanging operations
	defaultOperationTimeout = 10 * time.Second
)

// TransactionManager handles transaction management with deadlock detection and prevention
type TransactionManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// TransactionOptions configures transaction behavior
type TransactionOptions struct {
	// Timeout for acquiring locks
	LockTimeout time.Duration
	// Isolation level for the transaction
	IsolationLevel sql.IsolationLevel
	// Whether to retry on deadlock
	RetryOnDeadlock bool
	// Maximum number of retries
	MaxRetries int
	// Statement timeout to prevent query hangs
	StatementTimeout time.Duration
}

// DefaultTransactionOptions returns default options
func DefaultTransactionOptions() TransactionOptions {
	return TransactionOptions{
		LockTimeout:      time.Second * 5,
		IsolationLevel:   sql.LevelReadCommitted,
		RetryOnDeadlock:  true,
		MaxRetries:       5,
		StatementTimeout: time.Second * 15,
	}
}

// WithTransaction executes a function within a transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, opts TransactionOptions, fn func(*sql.Tx) error) error {
	var tx *sql.Tx
	var err error

	// Get test schema from context or fallback to global value
	testSchema := ""
	if ctx != nil {
		if schemaVal := ctx.Value("test_schema"); schemaVal != nil {
			if schema, ok := schemaVal.(string); ok && schema != "" {
				testSchema = schema
			}
		}
	}

	if testSchema == "" {
		testSchema = testutil.GetCurrentTestSchema()
	}

	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		// Verify DB connection is valid before starting transaction
		if err := tm.db.PingContext(ctx); err != nil {
			log.Printf("Database connection error on attempt %d/%d: %v",
				attempt+1, opts.MaxRetries+1, err)

			if attempt < opts.MaxRetries {
				retryDelay := time.Millisecond * time.Duration(100*(1<<attempt))
				log.Printf("Database connection retry in %d ms", retryDelay.Milliseconds())
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("database connection failed after %d retries: %w",
				opts.MaxRetries, err)
		}

		// Start transaction with specified isolation level
		tx, err = tm.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: opts.IsolationLevel,
		})
		if err != nil {
			// Check if it's a connection error
			if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
				log.Printf("Failed to start transaction (bad connection) on attempt %d/%d",
					attempt+1, opts.MaxRetries+1)
				retryDelay := time.Millisecond * time.Duration(100*(1<<attempt))
				log.Printf("Transaction retry in %d ms", retryDelay.Milliseconds())
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("error starting transaction: %w", err)
		}

		// Create a new context specifically for configuration queries
		configCtx, configCancel := context.WithTimeout(ctx, 2*time.Second)

		// Set lock timeout
		_, err = tx.ExecContext(configCtx, fmt.Sprintf("SET LOCAL lock_timeout = '%dms'", opts.LockTimeout.Milliseconds()))
		if err != nil {
			configCancel()
			safeClose(tx)
			// Check if it's a connection error
			if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
				log.Printf("Failed to set lock timeout (bad connection) on attempt %d/%d",
					attempt+1, opts.MaxRetries+1)
				retryDelay := time.Millisecond * time.Duration(100*(1<<attempt))
				log.Printf("Transaction retry in %d ms", retryDelay.Milliseconds())
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("error setting lock timeout: %w", err)
		}

		// Set statement timeout to prevent query hangs
		_, err = tx.ExecContext(configCtx, fmt.Sprintf("SET LOCAL statement_timeout = '%dms'", opts.StatementTimeout.Milliseconds()))
		if err != nil {
			configCancel()
			safeClose(tx)
			// Check if it's a connection error
			if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
				log.Printf("Failed to set statement timeout (bad connection) on attempt %d/%d",
					attempt+1, opts.MaxRetries+1)
				retryDelay := time.Millisecond * time.Duration(100*(1<<attempt))
				log.Printf("Transaction retry in %d ms", retryDelay.Milliseconds())
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("error setting statement timeout: %w", err)
		}

		// Priority for search_path:
		// 1. If we have a test schema (either from context or global), use that
		// 2. Otherwise, get the current search_path from the connection
		if testSchema != "" {
			// We're running in a test, set the test schema
			// Be more explicit with search_path to avoid public schema being included first
			_, err = tx.ExecContext(configCtx, fmt.Sprintf("SET LOCAL search_path TO %s, public", pq.QuoteIdentifier(testSchema)))
			if err != nil {
				configCancel()
				safeClose(tx)
				// Check if it's a connection error
				if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
					log.Printf("Failed to set test schema search_path (bad connection) on attempt %d/%d",
						attempt+1, opts.MaxRetries+1)
					retryDelay := time.Millisecond * time.Duration(100*(1<<attempt))
					log.Printf("Transaction retry in %d ms", retryDelay.Milliseconds())
					time.Sleep(retryDelay)
					continue
				}
				return fmt.Errorf("error setting test schema search_path: %w", err)
			}
		} else {
			// Get the current search_path and ensure it's correctly set in the transaction
			var searchPath string
			err = tm.db.QueryRowContext(configCtx, "SHOW search_path").Scan(&searchPath)
			if err != nil {
				configCancel()
				safeClose(tx)
				// Check if it's a connection error
				if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
					log.Printf("Failed to get current search_path (bad connection), retrying in %d ms",
						100*(1<<attempt))
					time.Sleep(time.Millisecond * time.Duration(100*(1<<attempt)))
					continue
				}
				return fmt.Errorf("error getting current search_path: %w", err)
			}

			// Set the same search_path in the transaction with LOCAL so it only affects this transaction
			_, err = tx.ExecContext(configCtx, fmt.Sprintf("SET LOCAL search_path TO %s", searchPath))
			if err != nil {
				configCancel()
				safeClose(tx)
				// Check if it's a connection error
				if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
					log.Printf("Failed to set search_path in transaction (bad connection), retrying in %d ms",
						100*(1<<attempt))
					time.Sleep(time.Millisecond * time.Duration(100*(1<<attempt)))
					continue
				}
				return fmt.Errorf("error setting search_path in transaction: %w", err)
			}
		}

		// Verify search_path was properly set
		var txSearchPath string
		err = tx.QueryRowContext(configCtx, "SHOW search_path").Scan(&txSearchPath)
		if err != nil {
			configCancel()
			safeClose(tx)
			// Check if it's a connection error
			if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
				log.Printf("Failed to verify search_path in transaction (bad connection), retrying in %d ms",
					100*(1<<attempt))
				time.Sleep(time.Millisecond * time.Duration(100*(1<<attempt)))
				continue
			}
			return fmt.Errorf("error verifying search_path in transaction: %w", err)
		}

		// Log the transaction search path for debugging
		if testSchema != "" {
			fmt.Printf("Transaction search_path: %s (expected test schema: %s)\n", txSearchPath, testSchema)
			if !strings.Contains(txSearchPath, testSchema) {
				configCancel()
				safeClose(tx)
				return fmt.Errorf("test schema not in transaction search_path: expected %s in %s",
					testSchema, txSearchPath)
			}
		}

		// Configuration is done, release the config context
		configCancel()

		// Execute the transaction function
		err = fn(tx)

		if err != nil {
			safeClose(tx)

			// Check for specific error types
			if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
				log.Printf("Transaction execution failed (bad connection), retrying in %d ms",
					100*(1<<attempt))
				time.Sleep(time.Millisecond * time.Duration(100*(1<<attempt)))
				continue
			}

			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case errDeadlockDetected, errLockTimeout:
					if opts.RetryOnDeadlock && attempt < opts.MaxRetries {
						// Wait before retrying (exponential backoff)
						time.Sleep(time.Millisecond * time.Duration(100*(1<<attempt)))
						continue
					}
				}
			}
			return err
		}

		// Commit the transaction with a specific timeout
		// Note: We create a timeout context but tx.Commit() doesn't accept a context
		// This at least gives us some protection via the parent context
		commitCtx, commitCancel := context.WithTimeout(ctx, 5*time.Second)
		defer commitCancel()

		// Verify search_path is still correctly set before committing
		var finalSearchPath string
		err = tx.QueryRowContext(commitCtx, "SHOW search_path").Scan(&finalSearchPath)
		if err != nil {
			safeClose(tx)
			// Check if it's a connection error
			if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
				log.Printf("Failed to verify search_path before commit (bad connection), retrying in %d ms",
					100*(1<<attempt))
				time.Sleep(time.Millisecond * time.Duration(100*(1<<attempt)))
				continue
			}
			return fmt.Errorf("error verifying search_path before commit: %w", err)
		}

		// For test schemas, verify it's still in the search path
		if testSchema != "" && !strings.Contains(finalSearchPath, testSchema) {
			safeClose(tx)
			return fmt.Errorf("test schema lost in transaction before commit: expected %s in %s",
				testSchema, finalSearchPath)
		}

		err = tx.Commit()

		if err != nil {
			// Check if it's a connection error
			if err.Error() == "driver: bad connection" && attempt < opts.MaxRetries {
				log.Printf("Failed to commit transaction (bad connection), retrying in %d ms",
					100*(1<<attempt))
				time.Sleep(time.Millisecond * time.Duration(100*(1<<attempt)))
				continue
			}
			return fmt.Errorf("error committing transaction: %w", err)
		}

		return nil
	}

	return err
}

// OrderedTransaction executes multiple operations in a consistent order to prevent deadlocks
func (tm *TransactionManager) OrderedTransaction(ctx context.Context, opts TransactionOptions, operations []func(*sql.Tx) error) error {
	// Create a timeout context if one wasn't provided
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.StatementTimeout)
		defer cancel()
	}

	return tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		for _, op := range operations {
			if err := op(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// MonitorDeadlocks starts monitoring for deadlocks
func (tm *TransactionManager) MonitorDeadlocks(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Create a timeout context for this specific query
			queryCtx, cancel := context.WithTimeout(context.Background(), defaultOperationTimeout)

			var deadlockCount int
			err := tm.db.QueryRowContext(queryCtx, `
				SELECT count(*) 
				FROM pg_stat_activity 
				WHERE wait_event_type = 'Lock' 
				AND wait_event = 'deadlock'
			`).Scan(&deadlockCount)

			cancel() // Always cancel the context to release resources

			if err != nil {
				return fmt.Errorf("error monitoring deadlocks: %w", err)
			}

			if deadlockCount > 0 {
				// TODO: Add proper logging and alerting here
				fmt.Printf("Warning: %d deadlocks detected\n", deadlockCount)
			}
		}
	}
}
