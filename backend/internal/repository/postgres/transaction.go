package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

const (
	// Error codes from: https://www.postgresql.org/docs/current/errcodes-appendix.html
	errDeadlockDetected = "40P01"
	errLockTimeout      = "55P03"
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
}

// DefaultTransactionOptions returns default options
func DefaultTransactionOptions() TransactionOptions {
	return TransactionOptions{
		LockTimeout:     time.Second * 3,
		IsolationLevel:  sql.LevelReadCommitted,
		RetryOnDeadlock: true,
		MaxRetries:      3,
	}
}

// WithTransaction executes a function within a transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, opts TransactionOptions, fn func(*sql.Tx) error) error {
	var err error
	var tx *sql.Tx

	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		// Start transaction with specified isolation level
		tx, err = tm.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: opts.IsolationLevel,
		})
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}

		// Set lock timeout
		_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL lock_timeout = '%dms'", opts.LockTimeout.Milliseconds()))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error setting lock timeout: %w", err)
		}

		// Execute the transaction function
		err = fn(tx)

		// Check for deadlock or lock timeout
		if err != nil {
			tx.Rollback()

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

		// Commit the transaction
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("error committing transaction: %w", err)
		}

		return nil
	}

	return err
}

// OrderedTransaction executes multiple operations in a consistent order to prevent deadlocks
func (tm *TransactionManager) OrderedTransaction(ctx context.Context, opts TransactionOptions, operations []func(*sql.Tx) error) error {
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
			var deadlockCount int
			err := tm.db.QueryRowContext(ctx, `
				SELECT count(*) 
				FROM pg_stat_activity 
				WHERE wait_event_type = 'Lock' 
				AND wait_event = 'deadlock'
			`).Scan(&deadlockCount)

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
