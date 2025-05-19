// Package postgres provides repository implementations for PostgreSQL.
package postgres

import (
	"database/sql"
	"log"
)

// safeClose closes a database resource (rows, stmt, tx) and logs any error.
// This function should be used in defer statements for proper error handling.
func safeClose(c interface{}) {
	var err error
	switch v := c.(type) {
	case *sql.Rows:
		err = v.Close()
	case *sql.Stmt:
		err = v.Close()
	case *sql.Tx:
		// For transactions, Rollback is safe to call after Commit
		err = v.Rollback()
		// Ignore error if transaction was already committed
		if err == sql.ErrTxDone {
			return
		}
	case *sql.DB:
		err = v.Close()
	default:
		log.Printf("Unknown closer type: %T", v)
		return
	}

	if err != nil {
		// Log the error but don't return it as defer statements can't return values
		log.Printf("Error closing resource: %v", err)
	}
}
