package testutil

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestTestingInfrastructure(t *testing.T) {
	// Set up test database
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// Test data generation
	t.Run("Test_data_generation", func(t *testing.T) {
		// Get current schema
		var schemaName string
		err := db.QueryRow("SELECT current_schema()").Scan(&schemaName)
		assert.NoError(t, err)

		// Create test users
		user1 := CreateTestUser(t, db)
		user2 := CreateTestUser(t, db)

		// Create test tribe
		tribe := CreateTestTribe(t, db, []TestUser{user1, user2})

		// Create test list
		_ = CreateTestList(t, db, tribe)

		// Use fully qualified table names to avoid schema issues
		usersTable := fmt.Sprintf("%s.users", pq.QuoteIdentifier(schemaName))
		tribesTable := fmt.Sprintf("%s.tribes", pq.QuoteIdentifier(schemaName))
		listsTable := fmt.Sprintf("%s.lists", pq.QuoteIdentifier(schemaName))

		// Verify data was created
		var count int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", usersTable)).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)

		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tribesTable)).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", listsTable)).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		// Clean up test data - use a separate function call in a separate transaction
		CleanupTestData(t, db)

		// Verify cleanup
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", usersTable)).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	// Test HTTP helpers
	t.Run("HTTP_helpers", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Add a test endpoint with mock auth
		router.GET("/test",
			MockAuthMiddleware("test-user-id", "test-firebase-uid"),
			func(c *gin.Context) {
				userID, exists := c.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, "test-user-id", userID)

				firebaseUID, exists := c.Get("firebase_uid")
				assert.True(t, exists)
				assert.Equal(t, "test-firebase-uid", firebaseUID)

				c.JSON(http.StatusOK, gin.H{
					"message": "success",
				})
			},
		)

		// Test request execution
		req := TestRequest{
			Method: "GET",
			Path:   "/test",
			Header: map[string]string{
				"Authorization": "Bearer test-token",
			},
		}

		resp := ExecuteRequest(t, router, req)

		// Check response
		expected := TestResponse{
			Code: http.StatusOK,
			Body: gin.H{
				"message": "success",
			},
		}

		CheckResponse(t, resp, expected)
	})
}
