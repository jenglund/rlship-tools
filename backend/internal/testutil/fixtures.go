package testutil

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// TestUser represents a test user
type TestUser struct {
	ID        uuid.UUID
	Email     string
	Name      string
	AvatarURL string
}

// TestRelationship represents a test relationship
type TestRelationship struct {
	ID      uuid.UUID
	Members []TestUser
}

// TestActivityList represents a test activity list
type TestActivityList struct {
	ID             uuid.UUID
	RelationshipID uuid.UUID
	Name           string
	Type           string
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *sql.DB) TestUser {
	t.Helper()

	user := TestUser{
		ID:        uuid.New(),
		Email:     fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		Name:      fmt.Sprintf("Test User %s", uuid.New().String()[:8]),
		AvatarURL: fmt.Sprintf("https://example.com/avatars/%s.jpg", uuid.New().String()),
	}

	_, err := db.Exec(`
		INSERT INTO users (id, email, name, avatar_url)
		VALUES ($1, $2, $3, $4)
	`, user.ID, user.Email, user.Name, user.AvatarURL)

	if err != nil {
		t.Fatalf("Error creating test user: %v", err)
	}

	return user
}

// CreateTestRelationship creates a test relationship with members in the database
func CreateTestRelationship(t *testing.T, db *sql.DB, members []TestUser) TestRelationship {
	t.Helper()

	relationship := TestRelationship{
		ID:      uuid.New(),
		Members: members,
	}

	_, err := db.Exec(`
		INSERT INTO relationships (id)
		VALUES ($1)
	`, relationship.ID)

	if err != nil {
		t.Fatalf("Error creating test relationship: %v", err)
	}

	for _, member := range members {
		_, err := db.Exec(`
			INSERT INTO relationship_members (relationship_id, user_id)
			VALUES ($1, $2)
		`, relationship.ID, member.ID)

		if err != nil {
			t.Fatalf("Error adding member to test relationship: %v", err)
		}
	}

	return relationship
}

// CreateTestActivityList creates a test activity list in the database
func CreateTestActivityList(t *testing.T, db *sql.DB, relationshipID uuid.UUID) TestActivityList {
	t.Helper()

	list := TestActivityList{
		ID:             uuid.New(),
		RelationshipID: relationshipID,
		Name:           fmt.Sprintf("Test List %s", uuid.New().String()[:8]),
		Type:           "restaurants",
	}

	_, err := db.Exec(`
		INSERT INTO activity_lists (id, relationship_id, name, type)
		VALUES ($1, $2, $3, $4)
	`, list.ID, list.RelationshipID, list.Name, list.Type)

	if err != nil {
		t.Fatalf("Error creating test activity list: %v", err)
	}

	return list
}

// CleanupTestData removes all test data from the database
func CleanupTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"interest_button_activations",
		"interest_buttons",
		"activity_event_photos",
		"activity_events",
		"activity_notes",
		"activities",
		"activity_lists",
		"relationship_members",
		"relationships",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Fatalf("Error cleaning up %s table: %v", table, err)
		}
	}
}
