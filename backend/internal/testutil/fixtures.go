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
func CreateTestUser(t *testing.T, db *sql.DB) TestUser {
	t.Helper()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				t.Logf("Error rolling back transaction: %v", rbErr)
			}
		}
	}()

	user := TestUser{
		ID:          uuid.New(),
		FirebaseUID: fmt.Sprintf("test-firebase-%s", uuid.New().String()[:8]),
		Provider:    "google",
		Email:       fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		Name:        fmt.Sprintf("Test User %s", uuid.New().String()[:8]),
		AvatarURL:   fmt.Sprintf("https://example.com/avatars/%s.jpg", uuid.New().String()),
	}

	fmt.Printf("DEBUG: Creating test user: %+v\n", user)
	_, err = tx.Exec(`
		INSERT INTO users (id, firebase_uid, provider, email, name, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.FirebaseUID, user.Provider, user.Email, user.Name, user.AvatarURL)

	if err != nil {
		fmt.Printf("DEBUG: Error creating test user: %v\n", err)
		t.Fatalf("Error creating test user: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		t.Fatalf("Error committing transaction: %v", err)
	}

	return user
}

// CreateTestTribe creates a test tribe with members in the database
func CreateTestTribe(t *testing.T, db *sql.DB, members []TestUser) TestTribe {
	t.Helper()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}
	defer tx.Rollback()

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

	// Insert the tribe
	_, err = tx.Exec(`
		INSERT INTO tribes (id, name)
		VALUES ($1, $2)
	`, tribe.ID, tribe.Name)
	if err != nil {
		t.Fatalf("Failed to create test tribe: %v", err)
	}

	// Add members to the tribe
	for _, member := range members {
		_, err = tx.Exec(`
			INSERT INTO tribe_members (tribe_id, user_id)
			VALUES ($1, $2)
		`, tribe.ID, member.ID)
		if err != nil {
			t.Fatalf("Failed to add member to tribe: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	return tribe
}

// CreateTestList creates a test list in the database
func CreateTestList(t *testing.T, db *sql.DB, tribe TestTribe) TestList {
	t.Helper()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}
	defer tx.Rollback()

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
		INSERT INTO lists (id, name, type)
		VALUES ($1, $2, $3)
	`, list.ID, list.Name, list.Type)
	if err != nil {
		t.Fatalf("Failed to create test list: %v", err)
	}

	// Create the list owner relationship
	_, err = tx.Exec(`
		INSERT INTO list_owners (list_id, owner_id, owner_type)
		VALUES ($1, $2, $3)
	`, list.ID, tribe.ID, "tribe")
	if err != nil {
		t.Fatalf("Failed to create list owner relationship: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	return list
}

// CleanupTestData removes all test data from the database
func CleanupTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				t.Logf("Error rolling back transaction: %v", rbErr)
			}
		}
	}()

	// Set constraints to deferred for this transaction
	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		t.Fatalf("Error deferring constraints: %v", err)
	}

	tables := []string{
		"list_shares",
		"list_items",
		"list_owners",
		"lists",
		"tribe_members",
		"tribes",
		"users",
	}

	for _, table := range tables {
		_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Fatalf("Error cleaning up %s table: %v", table, err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		t.Fatalf("Error committing transaction: %v", err)
	}
}
