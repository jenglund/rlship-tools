package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestTransactionManager_WithTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer safeClose(db)

	tm := NewTransactionManager(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		fn      func(*sql.Tx) error
		wantErr bool
	}{
		{
			name: "successful transaction",
			fn: func(tx *sql.Tx) error {
				_, err := tx.Exec("SELECT 1")
				return err
			},
			wantErr: false,
		},
		{
			name: "transaction error",
			fn: func(tx *sql.Tx) error {
				return errors.New("test error")
			},
			wantErr: true,
		},
		{
			name: "deadlock error with retry",
			fn: func(tx *sql.Tx) error {
				if tx == nil {
					return &pq.Error{Code: errDeadlockDetected}
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "lock timeout error with retry",
			fn: func(tx *sql.Tx) error {
				if tx == nil {
					return &pq.Error{Code: errLockTimeout}
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultTransactionOptions()
			err := tm.WithTransaction(ctx, opts, tt.fn)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransactionManager_OrderedTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer safeClose(db)

	tm := NewTransactionManager(db)
	ctx := context.Background()

	tests := []struct {
		name       string
		operations []func(*sql.Tx) error
		wantErr    bool
	}{
		{
			name: "successful ordered operations",
			operations: []func(*sql.Tx) error{
				func(tx *sql.Tx) error {
					_, err := tx.Exec("SELECT 1")
					return err
				},
				func(tx *sql.Tx) error {
					_, err := tx.Exec("SELECT 2")
					return err
				},
			},
			wantErr: false,
		},
		{
			name: "error in operation",
			operations: []func(*sql.Tx) error{
				func(tx *sql.Tx) error {
					return errors.New("test error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultTransactionOptions()
			err := tm.OrderedTransaction(ctx, opts, tt.operations)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransactionManager_MonitorDeadlocks(t *testing.T) {
	db := setupTestDB(t)
	defer safeClose(db)

	tm := NewTransactionManager(db)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := tm.MonitorDeadlocks(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestDefaultTransactionOptions(t *testing.T) {
	opts := DefaultTransactionOptions()

	assert.Equal(t, time.Second*5, opts.LockTimeout)
	assert.Equal(t, sql.LevelReadCommitted, opts.IsolationLevel)
	assert.True(t, opts.RetryOnDeadlock)
	assert.Equal(t, 5, opts.MaxRetries)
	assert.Equal(t, time.Second*15, opts.StatementTimeout)
}
