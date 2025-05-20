package testutil

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

// TestUser represents a test user
type TestUser struct {
	ID          uuid.UUID
	FirebaseUID string
	Provider    string
	Email       string
	Name        string
	AvatarURL   string
}

// TestTribe represents a test tribe
type TestTribe struct {
	ID      uuid.UUID
	Name    string
	Members []TestUser
}

// TestList represents a test list
type TestList struct {
	ID      uuid.UUID
	TribeID uuid.UUID
	Name    string
	Type    models.ListType
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db interface{}) TestUser {
	t.Helper()

	// Handle different DB types
	var sqlDB *sql.DB
	switch d := db.(type) {
	case *sql.DB:
		sqlDB = d
	case *SchemaDB:
		sqlDB = d.DB
	default:
		t.Fatalf("Unsupported DB type: %T", db)
	}

	// Start a transaction
	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}
	defer safeClose(tx)

	user := TestUser{
		ID:          uuid.New(),
		FirebaseUID: fmt.Sprintf("test-firebase-%s", uuid.New().String()[:8]),
		Provider:    "google",
		Email:       fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		Name:        fmt.Sprintf("Test User %s", uuid.New().String()[:8]),
		AvatarURL:   fmt.Sprintf("https://example.com/avatars/%s.jpg", uuid.New().String()),
	}

	_, err = tx.Exec(`
		INSERT INTO users (id, firebase_uid, provider, email, name, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.FirebaseUID, user.Provider, user.Email, user.Name, user.AvatarURL)

	if err != nil {
		t.Fatalf("Error creating test user: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		t.Fatalf("Error committing transaction: %v", err)
	}

	return user
}

// CreateTestTribe creates a test tribe with members in the database
func CreateTestTribe(t *testing.T, db interface{}, members []TestUser) TestTribe {
	t.Helper()

	// Handle different DB types
	var sqlDB *sql.DB
	switch d := db.(type) {
	case *sql.DB:
		sqlDB = d
	case *SchemaDB:
		sqlDB = d.DB
	default:
		t.Fatalf("Unsupported DB type: %T", db)
	}

	// Start a transaction
	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}
	defer safeClose(tx)

	// Set constraints to deferred
	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		t.Fatalf("Failed to defer constraints: %v", err)
	}

	tribe := TestTribe{
		ID:      uuid.New(),
		Name:    fmt.Sprintf("Test Tribe %s", uuid.New().String()),
		Members: members,
	}

	// Insert the tribe with all required fields
	_, err = tx.Exec(`
		INSERT INTO tribes (id, name, description, type, visibility, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tribe.ID, tribe.Name, "Test Description", models.TribeTypeCouple, models.VisibilityPrivate, `{"test":"fixture"}`)
	if err != nil {
		t.Fatalf("error creating tribe: %v", err)
	}

	// Add members to the tribe with all required fields
	for _, member := range members {
		_, err = tx.Exec(`
			INSERT INTO tribe_members (tribe_id, user_id, membership_type, metadata)
			VALUES ($1, $2, $3, $4)
		`, tribe.ID, member.ID, models.MembershipFull, `{"test":"member_fixture"}`)
		if err != nil {
			t.Fatalf("error adding member to tribe: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	return tribe
}

// CreateTestList creates a test list in the database
func CreateTestList(t *testing.T, db interface{}, tribe TestTribe) TestList {
	t.Helper()

	// Handle different DB types
	var sqlDB *sql.DB
	switch d := db.(type) {
	case *sql.DB:
		sqlDB = d
	case *SchemaDB:
		sqlDB = d.DB
	default:
		t.Fatalf("Unsupported DB type: %T", db)
	}

	// Start a transaction
	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}
	defer safeClose(tx)

	// Set constraints to deferred
	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		t.Fatalf("Failed to defer constraints: %v", err)
	}

	list := TestList{
		ID:      uuid.New(),
		TribeID: tribe.ID,
		Name:    fmt.Sprintf("Test List %s", uuid.New().String()),
		Type:    "wishlist",
	}

	// Insert the list
	_, err = tx.Exec(`
		INSERT INTO lists (id, name, type, owner_id, owner_type, visibility, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, list.ID, list.Name, list.Type, tribe.ID, models.OwnerTypeTribe, models.VisibilityPrivate, `{"test":"list_fixture"}`)
	if err != nil {
		t.Fatalf("Failed to create test list: %v", err)
	}

	// Create list owner entry
	_, err = tx.Exec(`
		INSERT INTO list_owners (list_id, owner_id, owner_type)
		VALUES ($1, $2, $3)
	`, list.ID, tribe.ID, models.OwnerTypeTribe)
	if err != nil {
		t.Fatalf("Failed to create list owner: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	return list
}

// CreateTestActivity creates a test activity in the database
func CreateTestActivity(t *testing.T, db interface{}, user TestUser) *models.Activity {
	t.Helper()

	// Handle different DB types
	var sqlDB *sql.DB
	switch d := db.(type) {
	case *sql.DB:
		sqlDB = d
	case *SchemaDB:
		sqlDB = d.DB
	default:
		t.Fatalf("Unsupported DB type: %T", db)
	}

	// Start a transaction
	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}
	defer safeClose(tx)

	// Set constraints to deferred
	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		t.Fatalf("Failed to defer constraints: %v", err)
	}

	activity := &models.Activity{
		ID:          uuid.New(),
		UserID:      user.ID,
		Type:        models.ActivityTypeLocation,
		Name:        fmt.Sprintf("Test Activity %s", uuid.New().String()),
		Description: "Test Description",
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{},
	}

	// Insert the activity
	_, err = tx.Exec(`
		INSERT INTO activities (id, user_id, type, name, description, visibility, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, activity.ID, activity.UserID, activity.Type, activity.Name, activity.Description, activity.Visibility, "{}")
	if err != nil {
		t.Fatalf("Failed to create test activity: %v", err)
	}

	// Create activity owner entry
	_, err = tx.Exec(`
		INSERT INTO activity_owners (activity_id, owner_id, owner_type)
		VALUES ($1, $2, $3)
	`, activity.ID, user.ID, models.OwnerTypeUser)
	if err != nil {
		t.Fatalf("Failed to create activity owner: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	return activity
}

// CleanupTestData removes all test data from the database
func CleanupTestData(t *testing.T, db interface{}) {
	t.Helper()

	// Handle different DB types
	var sqlDB *sql.DB
	switch d := db.(type) {
	case *sql.DB:
		sqlDB = d
	case *SchemaDB:
		sqlDB = d.DB
	default:
		t.Fatalf("Unsupported DB type: %T", db)
	}

	// Start a transaction
	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}
	defer safeClose(tx)

	// Set constraints to deferred for this transaction
	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		t.Fatalf("Error deferring constraints: %v", err)
	}

	tables := []string{
		"list_sharing",
		"list_items",
		"list_owners",
		"lists",
		"tribe_members",
		"tribes",
		"activity_shares",
		"activity_photos",
		"activity_owners",
		"activities",
		"users",
	}

	for _, table := range tables {
		_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Error clearing table %s: %v", table, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Error committing transaction: %v", err)
	}
}
