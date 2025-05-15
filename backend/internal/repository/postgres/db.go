package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jenglund/rlship-tools/internal/models"
	_ "github.com/lib/pq"
)

// NewDB creates a new database connection
func NewDB(host string, port int, user, password, dbname, sslmode string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	// Set reasonable defaults for connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// Repositories holds all repository implementations
type Repositories struct {
	Users      *UserRepository
	Tribes     *TribeRepository
	Activities *ActivityRepository
	db         *sql.DB
}

// NewRepositories creates new instances of all repositories
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Users:      NewUserRepository(db),
		Tribes:     NewTribeRepository(db),
		Activities: NewActivityRepository(db),
		db:         db,
	}
}

// DB returns the underlying database connection
func (r *Repositories) DB() *sql.DB {
	return r.db
}

// GetUserRepository returns the user repository
func (r *Repositories) GetUserRepository() models.UserRepository {
	return r.Users
}
