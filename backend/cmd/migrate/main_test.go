package main

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/jenglund/rlship-tools/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockSource implements source.Driver
// It is kept for reference even though currently unused
//
//nolint:unused
type mockSource struct {
	version uint
	err     error
}

//nolint:unused
func (m *mockSource) Open(url string) (source.Driver, error) { return m, nil }

//nolint:unused
func (m *mockSource) Close() error { return nil }

//nolint:unused
func (m *mockSource) First() (uint, error) { return m.version, m.err }

//nolint:unused
func (m *mockSource) Prev(version uint) (uint, error) { return version - 1, m.err }

//nolint:unused
func (m *mockSource) Next(version uint) (uint, error) { return version + 1, m.err }

//nolint:unused
func (m *mockSource) ReadUp(version uint) (io.ReadCloser, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	// Return a string reader with some mock SQL content
	return io.NopCloser(strings.NewReader("CREATE TABLE mock (id int);")), "mock.sql", nil
}

//nolint:unused
func (m *mockSource) ReadDown(version uint) (io.ReadCloser, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	// Return a string reader with some mock SQL content
	return io.NopCloser(strings.NewReader("DROP TABLE mock;")), "mock.sql", nil
}

// mockDB implements database.Driver
// It is kept for reference even though currently unused
//
//nolint:unused
type mockDB struct {
	mock.Mock
	version int
	dirty   bool
	err     error
}

//nolint:unused
func (m *mockDB) Open(url string) (database.Driver, error) { return m, nil }

//nolint:unused
func (m *mockDB) Close() error { return nil }

//nolint:unused
func (m *mockDB) Lock() error { return m.err }

//nolint:unused
func (m *mockDB) Unlock() error { return m.err }

//nolint:unused
func (m *mockDB) Run(migration io.Reader) error { return m.err }

//nolint:unused
func (m *mockDB) SetVersion(version int, dirty bool) error {
	m.version = version
	m.dirty = dirty
	return m.err
}

//nolint:unused
func (m *mockDB) Version() (int, bool, error) { return m.version, m.dirty, m.err }

//nolint:unused
func (m *mockDB) Drop() error { return m.err }

// mockMigrator implements the Migrator interface for testing
type mockMigrator struct {
	version     uint
	dirty       bool
	versionErr  error
	upErr       error
	downErr     error
	forceErr    error
	upCalled    bool
	downCalled  bool
	forceCalled bool
}

func (m *mockMigrator) Up() error {
	m.upCalled = true
	return m.upErr
}

func (m *mockMigrator) Down() error {
	m.downCalled = true
	return m.downErr
}

func (m *mockMigrator) Force(version int) error {
	m.forceCalled = true
	return m.forceErr
}

func (m *mockMigrator) Version() (uint, bool, error) {
	return m.version, m.dirty, m.versionErr
}

func mockFactory(sourceErr, dbErr error) MigrateFactory {
	return func(sourceURL, databaseURL string) (Migrator, error) {
		if sourceErr != nil {
			return nil, sourceErr
		}
		if dbErr != nil {
			return nil, dbErr
		}

		return &mockMigrator{
			version: 1,
			dirty:   false,
		}, nil
	}
}

func mockConfigLoader() ConfigLoader {
	return func() (*config.Config, error) {
		return &config.Config{
			Database: config.DatabaseConfig{
				URL: "postgres://mock:mock@localhost:5432/mock?sslmode=disable",
			},
		}, nil
	}
}

func mockConfigLoaderWithError() ConfigLoader {
	return func() (*config.Config, error) {
		return nil, fmt.Errorf("config error")
	}
}

func TestMigrateCommand_NoArgs(t *testing.T) {
	args := []string{"migrate"}
	err := runMigrations(args, nil, mockConfigLoader())
	assert.EqualError(t, err, "Command required: up, down, or force")
}

func TestMigrateCommand_InvalidCommand(t *testing.T) {
	args := []string{"migrate", "invalid"}
	err := runMigrations(args, nil, mockConfigLoader())
	assert.EqualError(t, err, "Invalid command. Use 'up', 'down', or 'force'")
}

func TestMigrateCommand_Up(t *testing.T) {
	args := []string{"migrate", "up"}
	err := runMigrations(args, mockFactory(nil, nil), mockConfigLoader())
	assert.NoError(t, err)
}

func TestMigrateCommand_Up_SourceError(t *testing.T) {
	args := []string{"migrate", "up"}
	err := runMigrations(args, mockFactory(fmt.Errorf("source error"), nil), mockConfigLoader())
	assert.EqualError(t, err, "Error creating migrate instance: source error")
}

func TestMigrateCommand_Up_DBError(t *testing.T) {
	args := []string{"migrate", "up"}
	err := runMigrations(args, mockFactory(nil, fmt.Errorf("db error")), mockConfigLoader())
	assert.EqualError(t, err, "Error creating migrate instance: db error")
}

func TestMigrateCommand_Up_MigrationError(t *testing.T) {
	args := []string{"migrate", "up"}
	factory := func(sourceURL, databaseURL string) (Migrator, error) {
		return &mockMigrator{
			version: 1,
			dirty:   false,
			upErr:   fmt.Errorf("migration error"),
		}, nil
	}
	err := runMigrations(args, factory, mockConfigLoader())
	assert.EqualError(t, err, "Error running migrations: migration error")
}

func TestMigrateCommand_Up_NoChange(t *testing.T) {
	args := []string{"migrate", "up"}
	factory := func(sourceURL, databaseURL string) (Migrator, error) {
		return &mockMigrator{
			version: 1,
			dirty:   false,
			upErr:   migrate.ErrNoChange,
		}, nil
	}
	err := runMigrations(args, factory, mockConfigLoader())
	assert.NoError(t, err)
}

func TestMigrateCommand_Down(t *testing.T) {
	args := []string{"migrate", "down"}
	err := runMigrations(args, mockFactory(nil, nil), mockConfigLoader())
	assert.NoError(t, err)
}

func TestMigrateCommand_Down_SourceError(t *testing.T) {
	args := []string{"migrate", "down"}
	err := runMigrations(args, mockFactory(fmt.Errorf("source error"), nil), mockConfigLoader())
	assert.EqualError(t, err, "Error creating migrate instance: source error")
}

func TestMigrateCommand_Down_DBError(t *testing.T) {
	args := []string{"migrate", "down"}
	err := runMigrations(args, mockFactory(nil, fmt.Errorf("db error")), mockConfigLoader())
	assert.EqualError(t, err, "Error creating migrate instance: db error")
}

func TestMigrateCommand_Down_MigrationError(t *testing.T) {
	args := []string{"migrate", "down"}
	factory := func(sourceURL, databaseURL string) (Migrator, error) {
		return &mockMigrator{
			version: 1,
			dirty:   false,
			downErr: fmt.Errorf("migration error"),
		}, nil
	}
	err := runMigrations(args, factory, mockConfigLoader())
	assert.EqualError(t, err, "Error running migrations: migration error")
}

func TestMigrateCommand_Down_NoChange(t *testing.T) {
	args := []string{"migrate", "down"}
	factory := func(sourceURL, databaseURL string) (Migrator, error) {
		return &mockMigrator{
			version: 1,
			dirty:   false,
			downErr: migrate.ErrNoChange,
		}, nil
	}
	err := runMigrations(args, factory, mockConfigLoader())
	assert.NoError(t, err)
}

func TestMigrateCommand_Force(t *testing.T) {
	args := []string{"migrate", "force", "2"}
	err := runMigrations(args, mockFactory(nil, nil), mockConfigLoader())
	assert.NoError(t, err)
}

func TestMigrateCommand_Force_InvalidVersion(t *testing.T) {
	args := []string{"migrate", "force", "invalid"}
	err := runMigrations(args, nil, mockConfigLoader())
	assert.Contains(t, err.Error(), "Invalid version number")
}

func TestMigrateCommand_Force_MissingVersion(t *testing.T) {
	args := []string{"migrate", "force"}
	err := runMigrations(args, nil, mockConfigLoader())
	assert.EqualError(t, err, "Version number required for force command")
}

func TestMigrateCommand_Force_SourceError(t *testing.T) {
	args := []string{"migrate", "force", "2"}
	err := runMigrations(args, mockFactory(fmt.Errorf("source error"), nil), mockConfigLoader())
	assert.EqualError(t, err, "Error creating migrate instance: source error")
}

func TestMigrateCommand_Force_DBError(t *testing.T) {
	args := []string{"migrate", "force", "2"}
	err := runMigrations(args, mockFactory(nil, fmt.Errorf("db error")), mockConfigLoader())
	assert.EqualError(t, err, "Error creating migrate instance: db error")
}

func TestMigrateCommand_Force_MigrationError(t *testing.T) {
	args := []string{"migrate", "force", "2"}
	factory := func(sourceURL, databaseURL string) (Migrator, error) {
		return &mockMigrator{
			version:  1,
			dirty:    false,
			forceErr: fmt.Errorf("migration error"),
		}, nil
	}
	err := runMigrations(args, factory, mockConfigLoader())
	assert.EqualError(t, err, "Error forcing version: migration error")
}

func TestMigrateCommand_ConfigError(t *testing.T) {
	args := []string{"migrate", "up"}
	err := runMigrations(args, mockFactory(nil, nil), mockConfigLoaderWithError())
	assert.EqualError(t, err, "Error loading config: config error")
}

func TestMigrateCommand_VersionError(t *testing.T) {
	args := []string{"migrate", "up"}
	factory := func(sourceURL, databaseURL string) (Migrator, error) {
		return &mockMigrator{
			version:    1,
			dirty:      false,
			versionErr: fmt.Errorf("version error"),
		}, nil
	}
	err := runMigrations(args, factory, mockConfigLoader())
	assert.EqualError(t, err, "Error getting version: version error")
}

func TestMigrateCommand_NilVersion(t *testing.T) {
	args := []string{"migrate", "up"}
	factory := func(sourceURL, databaseURL string) (Migrator, error) {
		return &mockMigrator{
			version:    1,
			dirty:      false,
			versionErr: migrate.ErrNilVersion,
		}, nil
	}
	err := runMigrations(args, factory, mockConfigLoader())
	assert.NoError(t, err)
}
