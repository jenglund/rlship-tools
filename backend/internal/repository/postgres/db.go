package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/lib/pq"
)

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	db interface{}
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db interface{}) BaseRepository {
	return BaseRepository{db: db}
}

// GetDB returns the database connection with appropriate type
func (r *BaseRepository) GetDB() interface{} {
	return r.db
}

// GetQueryDB returns a SQL DB that can be used for queries
func (r *BaseRepository) GetQueryDB() *sql.DB {
	switch db := r.db.(type) {
	case *sql.DB:
		// Check if we're in a test context and should use a specific schema
		if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
			fmt.Printf("GetQueryDB for *sql.DB: Setting search path to test schema: %s\n", testSchema)
			// Try to set the search path - but don't fail if it doesn't work
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, err := db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(testSchema)))
			if err != nil {
				fmt.Printf("Warning: Failed to set search path in GetQueryDB: %v\n", err)
			}
		}
		return db
	case *testutil.SchemaDB:
		// Always use the UnwrapDB method to ensure search path is set
		sqlDB := testutil.UnwrapDB(db)

		// Double-check that search path is set correctly
		if sqlDB != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			var searchPath string
			err := sqlDB.QueryRowContext(ctx, "SHOW search_path").Scan(&searchPath)
			if err != nil {
				fmt.Printf("Warning: Error checking search path in GetQueryDB: %v\n", err)
			} else {
				schemaName := db.GetSchemaName()
				if !strings.Contains(searchPath, schemaName) {
					fmt.Printf("Warning: Schema %s not in search path (%s), attempting to fix\n",
						schemaName, searchPath)

					if err := db.SetSearchPath(); err != nil {
						fmt.Printf("Error setting search path: %v\n", err)
					}
				}
			}
		}

		return sqlDB
	default:
		if r.db == nil {
			return nil
		}
		panic(fmt.Sprintf("Unsupported DB type: %T", r.db))
	}
}

// GetSchemaDB returns a SchemaDB if available, or nil
func (r *BaseRepository) GetSchemaDB() *testutil.SchemaDB {
	if db, ok := r.db.(*testutil.SchemaDB); ok {
		return db
	}
	return nil
}

// GetSchemaName returns the current schema name, if available
func (r *BaseRepository) GetSchemaName() string {
	if db, ok := r.db.(*testutil.SchemaDB); ok {
		return db.GetSchemaName()
	}

	// Check for current test schema
	if schema := testutil.GetCurrentTestSchema(); schema != "" {
		return schema
	}

	// Default to public schema
	return "public"
}

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

	// Set improved connection pool settings
	db.SetMaxOpenConns(50)                  // Increased from 25 to handle more concurrent requests
	db.SetMaxIdleConns(10)                  // Increased from 5 to maintain more idle connections
	db.SetConnMaxLifetime(10 * time.Minute) // Increased from 5 minutes to reduce connection cycling
	db.SetConnMaxIdleTime(3 * time.Minute)  // Increased from 1 minute to reduce connection cycling

	// Log connection pool settings
	log.Printf("Database connection pool configured: MaxOpenConns=%d, MaxIdleConns=%d, ConnMaxLifetime=%v, ConnMaxIdleTime=%v",
		50, 10, 10*time.Minute, 3*time.Minute)

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
		sqlDB = d.DB
	default:
		if db == nil {
			sqlDB = nil
		} else {
			panic(fmt.Sprintf("Unsupported DB type: %T", db))
		}
	}

	return &Repositories{
		Users:          NewUserRepository(db),
		Tribes:         NewTribeRepository(db),
		Activities:     NewActivityRepository(db),
		ActivityPhotos: NewActivityPhotosRepository(db),
		Lists:          NewListRepository(db),
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
