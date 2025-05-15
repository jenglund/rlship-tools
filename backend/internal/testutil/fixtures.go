package testutil

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
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

// TestActivityList represents a test activity list
type TestActivityList struct {
	ID      uuid.UUID
	TribeID uuid.UUID
	Name    string
	Type    string
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *sql.DB) TestUser {
	t.Helper()

	user := TestUser{
		ID:          uuid.New(),
		FirebaseUID: fmt.Sprintf("test-firebase-%s", uuid.New().String()[:8]),
		Provider:    "google",
		Email:       fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		Name:        fmt.Sprintf("Test User %s", uuid.New().String()[:8]),
		AvatarURL:   fmt.Sprintf("https://example.com/avatars/%s.jpg", uuid.New().String()),
	}

	fmt.Printf("DEBUG: Creating test user: %+v\n", user)
	_, err := db.Exec(`
		INSERT INTO users (id, firebase_uid, provider, email, name, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.FirebaseUID, user.Provider, user.Email, user.Name, user.AvatarURL)

	if err != nil {
		fmt.Printf("DEBUG: Error creating test user: %v\n", err)
		t.Fatalf("Error creating test user: %v", err)
	}

	return user
}

// CreateTestTribe creates a test tribe with members in the database
func CreateTestTribe(t *testing.T, db *sql.DB, members []TestUser) TestTribe {
	t.Helper()

	tribe := TestTribe{
		ID:      uuid.New(),
		Name:    fmt.Sprintf("Test Tribe %s", uuid.New().String()[:8]),
		Members: members,
	}

	fmt.Printf("DEBUG: Creating test tribe: %+v\n", tribe)

	// Check if tribes table exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'tribes'
		)
	`).Scan(&exists)
	if err != nil {
		fmt.Printf("DEBUG: Error checking if tribes table exists: %v\n", err)
		t.Fatalf("Error checking if tribes table exists: %v", err)
	}
	fmt.Printf("DEBUG: Tribes table exists: %v\n", exists)

	// Check tribes table schema
	rows, err := db.Query(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = 'tribes'
	`)
	if err != nil {
		fmt.Printf("DEBUG: Error checking tribes table schema: %v\n", err)
		t.Fatalf("Error checking tribes table schema: %v", err)
	}
	defer rows.Close()

	fmt.Printf("DEBUG: Tribes table columns:\n")
	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&colName, &dataType); err != nil {
			fmt.Printf("DEBUG: Error scanning column info: %v\n", err)
			t.Fatalf("Error scanning column info: %v", err)
		}
		fmt.Printf("DEBUG: Column: %s (%s)\n", colName, dataType)
	}

	_, err = db.Exec(`
		INSERT INTO tribes (id, name)
		VALUES ($1, $2)
	`, tribe.ID, tribe.Name)

	if err != nil {
		fmt.Printf("DEBUG: Error creating test tribe: %v\n", err)
		t.Fatalf("Error creating test tribe: %v", err)
	}

	for _, member := range members {
		fmt.Printf("DEBUG: Adding member to tribe: %+v\n", member)
		_, err := db.Exec(`
			INSERT INTO tribe_members (tribe_id, user_id, type)
			VALUES ($1, $2, 'full')
		`, tribe.ID, member.ID)

		if err != nil {
			fmt.Printf("DEBUG: Error adding member to tribe: %v\n", err)
			t.Fatalf("Error adding member to test tribe: %v", err)
		}
	}

	return tribe
}

// CreateTestActivityList creates a test activity list in the database
func CreateTestActivityList(t *testing.T, db *sql.DB, tribeID uuid.UUID) TestActivityList {
	t.Helper()

	list := TestActivityList{
		ID:      uuid.New(),
		TribeID: tribeID,
		Name:    fmt.Sprintf("Test List %s", uuid.New().String()[:8]),
		Type:    "list",
	}

	_, err := db.Exec(`
		INSERT INTO activities (id, type, name, description, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'private', NOW(), NOW())
	`, list.ID, list.Type, list.Name, "Test activity list")

	if err != nil {
		t.Fatalf("Error creating test activity list: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO activity_owners (activity_id, owner_id, owner_type, created_at)
		VALUES ($1, $2, 'tribe', NOW())
	`, list.ID, list.TribeID)

	if err != nil {
		t.Fatalf("Error adding owner to test activity list: %v", err)
	}

	return list
}

// CleanupTestData removes all test data from the database
func CleanupTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"activity_shares",
		"activity_owners",
		"activities",
		"tribe_members",
		"tribes",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Fatalf("Error cleaning up %s table: %v", table, err)
		}
	}
}
