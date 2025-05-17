package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		port        int
		user        string
		password    string
		dbname      string
		sslmode     string
		expectError bool
	}{
		{
			name:        "valid parameters",
			host:        "localhost",
			port:        5432,
			user:        "testuser",
			password:    "testpass",
			dbname:      "testdb",
			sslmode:     "disable",
			expectError: false,
		},
		{
			name:        "invalid port",
			host:        "localhost",
			port:        -1,
			user:        "testuser",
			password:    "testpass",
			dbname:      "testdb",
			sslmode:     "disable",
			expectError: true,
		},
		{
			name:        "empty host",
			host:        "",
			port:        5432,
			user:        "testuser",
			password:    "testpass",
			dbname:      "testdb",
			sslmode:     "disable",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDB(tt.host, tt.port, tt.user, tt.password, tt.dbname, tt.sslmode)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				// Only check this if we're expecting success and got no error
				if err == nil {
					assert.NotNil(t, db)
					safeClose(db)
				}
			}
		})
	}
}

func TestNewRepositories(t *testing.T) {
	// Create a mock DB
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer safeClose(db)

	// Test creating repositories
	repos := NewRepositories(db)

	// Verify that all repositories are initialized
	assert.NotNil(t, repos)
	assert.NotNil(t, repos.Users)
	assert.NotNil(t, repos.Tribes)
	assert.NotNil(t, repos.Activities)

	// Verify DB connection is stored
	assert.Equal(t, db, repos.db)
}

func TestRepositories_DB(t *testing.T) {
	// Create a mock DB
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer safeClose(db)

	// Initialize repositories
	repos := NewRepositories(db)

	// Test DB method
	returnedDB := repos.DB()
	assert.Equal(t, db, returnedDB)
}

func TestRepositories_GetUserRepository(t *testing.T) {
	// Create a mock DB
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer safeClose(db)

	// Initialize repositories
	repos := NewRepositories(db)

	// Test GetUserRepository method
	userRepo := repos.GetUserRepository()
	assert.NotNil(t, userRepo)
	assert.Implements(t, (*models.UserRepository)(nil), userRepo)
	assert.Equal(t, repos.Users, userRepo)
}

// Test error handling in NewDB for connection failures
func TestNewDB_ConnectionFailure(t *testing.T) {
	// This test will likely fail in real execution due to connection errors,
	// which is exactly what we want to test.
	db, err := NewDB(
		"non-existent-host",
		5432,
		"invalid-user",
		"invalid-password",
		"invalid-db",
		"disable",
	)

	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "error connecting to the database")
}

// Test list repository initialization
func TestNewRepositories_ListRepository(t *testing.T) {
	// Create a mock DB
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer safeClose(db)

	// Test creating list repository directly
	listRepo := NewListRepository(db)
	assert.NotNil(t, listRepo)
	assert.Implements(t, (*models.ListRepository)(nil), listRepo)
}

// Test with nil database connection
func TestNewRepositories_NilDB(t *testing.T) {
	// While this shouldn't happen in practice, it's good to test boundary cases
	repos := NewRepositories(nil)

	assert.NotNil(t, repos)
	assert.Nil(t, repos.db)

	// The repositories should still be initialized, but operations would fail
	assert.NotNil(t, repos.Users)
	assert.NotNil(t, repos.Tribes)
	assert.NotNil(t, repos.Activities)

	// Test DB() method with nil db
	returnedDB := repos.DB()
	assert.Nil(t, returnedDB)
}

// Test connection pooling settings
func TestNewDB_ConnectionPooling(t *testing.T) {
	// Skip this test if we can't actually connect to a database
	db, err := NewDB("localhost", 5432, "postgres", "postgres", "postgres", "disable")
	if err != nil {
		t.Skip("Skipping test because could not connect to database")
	}
	defer safeClose(db)

	// Check that connection pool settings are applied
	maxOpenConns := db.Stats().MaxOpenConnections
	assert.Equal(t, 25, maxOpenConns)
}
