package testutil

import (
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockListRepository struct {
	mock.Mock
}

// Basic CRUD operations
func (m *MockListRepository) Create(list *models.List) error {
	args := m.Called(list)
	return args.Error(0)
}

func (m *MockListRepository) GetByID(id uuid.UUID) (*models.List, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.List), args.Error(1)
}

func (m *MockListRepository) Update(list *models.List) error {
	args := m.Called(list)
	return args.Error(0)
}

func (m *MockListRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockListRepository) List(offset, limit int) ([]*models.List, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

// List item operations
func (m *MockListRepository) AddItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListRepository) UpdateItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListRepository) RemoveItem(listID, itemID uuid.UUID) error {
	args := m.Called(listID, itemID)
	return args.Error(0)
}

func (m *MockListRepository) GetItems(listID uuid.UUID) ([]*models.ListItem, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListRepository) GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*models.ListItem, error) {
	args := m.Called(listIDs, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListRepository) UpdateItemStats(itemID uuid.UUID, chosen bool) error {
	args := m.Called(itemID, chosen)
	return args.Error(0)
}

func (m *MockListRepository) MarkItemChosen(itemID uuid.UUID) error {
	args := m.Called(itemID)
	return args.Error(0)
}

// Owner management
func (m *MockListRepository) AddOwner(owner *models.ListOwner) error {
	args := m.Called(owner)
	return args.Error(0)
}

func (m *MockListRepository) RemoveOwner(listID, ownerID uuid.UUID) error {
	args := m.Called(listID, ownerID)
	return args.Error(0)
}

func (m *MockListRepository) GetOwners(listID uuid.UUID) ([]*models.ListOwner, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListOwner), args.Error(1)
}

func (m *MockListRepository) GetUserLists(userID uuid.UUID) ([]*models.List, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(tribeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetListsByOwner(ownerID uuid.UUID, ownerType models.OwnerType) ([]*models.List, error) {
	args := m.Called(ownerID, ownerType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

// Share management
func (m *MockListRepository) ShareWithTribe(share *models.ListShare) error {
	args := m.Called(share)
	return args.Error(0)
}

func (m *MockListRepository) UnshareWithTribe(listID, tribeID uuid.UUID) error {
	args := m.Called(listID, tribeID)
	return args.Error(0)
}

func (m *MockListRepository) GetSharedLists(tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(tribeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetSharedTribes(listID uuid.UUID) ([]*models.Tribe, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tribe), args.Error(1)
}

func (m *MockListRepository) GetListShares(listID uuid.UUID) ([]*models.ListShare, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListShare), args.Error(1)
}

// Sync management
func (m *MockListRepository) UpdateSyncStatus(listID uuid.UUID, status models.ListSyncStatus) error {
	args := m.Called(listID, status)
	return args.Error(0)
}

func (m *MockListRepository) GetConflicts(listID uuid.UUID) ([]*models.SyncConflict, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SyncConflict), args.Error(1)
}

func (m *MockListRepository) CreateConflict(conflict *models.SyncConflict) error {
	args := m.Called(conflict)
	return args.Error(0)
}

func (m *MockListRepository) ResolveConflict(conflictID uuid.UUID) error {
	args := m.Called(conflictID)
	return args.Error(0)
}

func (m *MockListRepository) GetListsBySource(source string) ([]*models.List, error) {
	args := m.Called(source)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) AddConflict(conflict *models.SyncConflict) error {
	args := m.Called(conflict)
	return args.Error(0)
}
