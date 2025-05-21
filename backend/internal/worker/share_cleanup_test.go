package worker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockListService mocks the ListService interface for testing
type MockListService struct {
	mock.Mock
}

func (m *MockListService) CleanupExpiredShares() error {
	args := m.Called()
	return args.Error(0)
}

// We're only implementing the method we need for testing
// In a real implementation, you would implement all methods of the ListService interface

func TestShareCleanupWorker(t *testing.T) {
	mockService := new(MockListService)

	// Set a short interval for testing
	interval := 50 * time.Millisecond

	t.Run("Initial cleanup happens immediately", func(t *testing.T) {
		mockService.On("CleanupExpiredShares").Return(nil).Once()

		worker := NewShareCleanupWorker(mockService, interval)
		worker.Start()

		// Give it a moment to run the initial cleanup
		time.Sleep(10 * time.Millisecond)

		// Stop the worker to prevent further calls
		worker.Stop()

		mockService.AssertExpectations(t)
	})

	t.Run("Periodic cleanup happens at intervals", func(t *testing.T) {
		// Expect at least 3 calls: initial + 2 periodic
		mockService.On("CleanupExpiredShares").Return(nil).Times(3)

		worker := NewShareCleanupWorker(mockService, interval)
		worker.Start()

		// Wait for a bit more than 2 intervals
		time.Sleep(interval*2 + 20*time.Millisecond)

		// Stop the worker to prevent further calls
		worker.Stop()

		mockService.AssertExpectations(t)
	})

	t.Run("Worker handles errors gracefully", func(t *testing.T) {
		// Initial call returns an error
		mockService.On("CleanupExpiredShares").Return(assert.AnError).Once()
		// Second call succeeds
		mockService.On("CleanupExpiredShares").Return(nil).Once()

		worker := NewShareCleanupWorker(mockService, interval)
		worker.Start()

		// Wait for a bit more than 1 interval
		time.Sleep(interval + 20*time.Millisecond)

		// Stop the worker to prevent further calls
		worker.Stop()

		mockService.AssertExpectations(t)
	})
}
