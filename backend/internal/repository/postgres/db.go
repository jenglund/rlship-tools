package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
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
	Users          models.UserRepository
	Tribes         models.TribeRepository
	Activities     models.ActivityRepository
	ActivityPhotos models.ActivityPhotosRepository
	Lists          models.ListRepository
	db             *sql.DB
}

// NewRepositories creates new instances of all repositories
func NewRepositories(db interface{}) *Repositories {
	var sqlDB *sql.DB

	// Handle different DB types
	switch d := db.(type) {
	case *sql.DB:
		sqlDB = d
	case *testutil.SchemaDB:
		sqlDB = d.UnwrapDB()
	default:
		if db == nil {
			sqlDB = nil
		} else {
			panic(fmt.Sprintf("Unsupported DB type: %T", db))
		}
	}

	return &Repositories{
		Users:          NewUserRepository(sqlDB),
		Tribes:         NewTribeRepository(sqlDB),
		Activities:     NewActivityRepository(sqlDB),
		ActivityPhotos: NewActivityPhotosRepository(sqlDB),
		Lists:          NewListRepository(sqlDB),
		db:             sqlDB,
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
