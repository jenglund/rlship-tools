package testutil

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

// MockRepositories holds all mock repository implementations
type MockRepositories struct {
	Users      *MockUserRepository
	Tribes     *MockTribeRepository
	Activities *MockActivityRepository
	db         *sql.DB
}

// DB returns the underlying database connection
func (r *MockRepositories) DB() *sql.DB {
	return r.db
}

// GetUserRepository returns the user repository
func (r *MockRepositories) GetUserRepository() models.UserRepository {
	return r.Users
}

// NewMockRepositories creates new instances of all mock repositories
func NewMockRepositories() *MockRepositories {
	return &MockRepositories{
		Users:      NewMockUserRepository(),
		Tribes:     NewMockTribeRepository(),
		Activities: NewMockActivityRepository(),
	}
}

// MockUserRepository implements models.UserRepository for testing
type MockUserRepository struct {
	CreateFunc           func(user *models.User) error
	GetByIDFunc          func(id uuid.UUID) (*models.User, error)
	GetByFirebaseUIDFunc func(firebaseUID string) (*models.User, error)
	GetByEmailFunc       func(email string) (*models.User, error)
	UpdateFunc           func(user *models.User) error
	DeleteFunc           func(id uuid.UUID) error
	ListFunc             func(offset, limit int) ([]*models.User, error)
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func (m *MockUserRepository) Create(user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByFirebaseUID(firebaseUID string) (*models.User, error) {
	if m.GetByFirebaseUIDFunc != nil {
		return m.GetByFirebaseUIDFunc(firebaseUID)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, nil
}

func (m *MockUserRepository) Update(user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockUserRepository) List(offset, limit int) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(offset, limit)
	}
	return nil, nil
}

// MockTribeRepository implements models.TribeRepository for testing
type MockTribeRepository struct {
	CreateFunc                     func(tribe *models.Tribe) error
	GetByIDFunc                    func(id uuid.UUID) (*models.Tribe, error)
	UpdateFunc                     func(tribe *models.Tribe) error
	DeleteFunc                     func(id uuid.UUID) error
	ListFunc                       func(offset, limit int) ([]*models.Tribe, error)
	AddMemberFunc                  func(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time, invitedBy *uuid.UUID) error
	UpdateMemberFunc               func(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time) error
	RemoveMemberFunc               func(tribeID, userID uuid.UUID) error
	GetMembersFunc                 func(tribeID uuid.UUID) ([]*models.TribeMember, error)
	GetUserTribesFunc              func(userID uuid.UUID) ([]*models.Tribe, error)
	GetExpiredGuestMembershipsFunc func() ([]*models.TribeMember, error)
}

// NewMockTribeRepository creates a new mock tribe repository
func NewMockTribeRepository() *MockTribeRepository {
	return &MockTribeRepository{}
}

func (m *MockTribeRepository) Create(tribe *models.Tribe) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(tribe)
	}
	return nil
}

func (m *MockTribeRepository) GetByID(id uuid.UUID) (*models.Tribe, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockTribeRepository) Update(tribe *models.Tribe) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(tribe)
	}
	return nil
}

func (m *MockTribeRepository) Delete(id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockTribeRepository) List(offset, limit int) ([]*models.Tribe, error) {
	if m.ListFunc != nil {
		return m.ListFunc(offset, limit)
	}
	return nil, nil
}

func (m *MockTribeRepository) AddMember(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time, invitedBy *uuid.UUID) error {
	if m.AddMemberFunc != nil {
		return m.AddMemberFunc(tribeID, userID, memberType, expiresAt, invitedBy)
	}
	return nil
}

func (m *MockTribeRepository) UpdateMember(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time) error {
	if m.UpdateMemberFunc != nil {
		return m.UpdateMemberFunc(tribeID, userID, memberType, expiresAt)
	}
	return nil
}

func (m *MockTribeRepository) RemoveMember(tribeID, userID uuid.UUID) error {
	if m.RemoveMemberFunc != nil {
		return m.RemoveMemberFunc(tribeID, userID)
	}
	return nil
}

func (m *MockTribeRepository) GetMembers(tribeID uuid.UUID) ([]*models.TribeMember, error) {
	if m.GetMembersFunc != nil {
		return m.GetMembersFunc(tribeID)
	}
	return nil, nil
}

func (m *MockTribeRepository) GetUserTribes(userID uuid.UUID) ([]*models.Tribe, error) {
	if m.GetUserTribesFunc != nil {
		return m.GetUserTribesFunc(userID)
	}
	return nil, nil
}

func (m *MockTribeRepository) GetExpiredGuestMemberships() ([]*models.TribeMember, error) {
	if m.GetExpiredGuestMembershipsFunc != nil {
		return m.GetExpiredGuestMembershipsFunc()
	}
	return nil, nil
}

// MockActivityRepository implements models.ActivityRepository for testing
type MockActivityRepository struct {
	CreateFunc                    func(activity *models.Activity) error
	GetByIDFunc                   func(id uuid.UUID) (*models.Activity, error)
	UpdateFunc                    func(activity *models.Activity) error
	DeleteFunc                    func(id uuid.UUID) error
	ListFunc                      func(offset, limit int) ([]*models.Activity, error)
	AddOwnerFunc                  func(activityID, ownerID uuid.UUID, ownerType string) error
	RemoveOwnerFunc               func(activityID, ownerID uuid.UUID) error
	GetOwnersFunc                 func(activityID uuid.UUID) ([]*models.ActivityOwner, error)
	GetUserActivitiesFunc         func(userID uuid.UUID) ([]*models.Activity, error)
	GetTribeActivitiesFunc        func(tribeID uuid.UUID) ([]*models.Activity, error)
	ShareWithTribeFunc            func(activityID, tribeID, userID uuid.UUID, expiresAt *time.Time) error
	UnshareWithTribeFunc          func(activityID, tribeID uuid.UUID) error
	GetSharedActivitiesFunc       func(tribeID uuid.UUID) ([]*models.Activity, error)
	MarkForDeletionFunc           func(activityID uuid.UUID) error
	CleanupOrphanedActivitiesFunc func() error
}

// NewMockActivityRepository creates a new mock activity repository
func NewMockActivityRepository() *MockActivityRepository {
	return &MockActivityRepository{}
}

func (m *MockActivityRepository) Create(activity *models.Activity) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(activity)
	}
	return nil
}

func (m *MockActivityRepository) GetByID(id uuid.UUID) (*models.Activity, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockActivityRepository) Update(activity *models.Activity) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(activity)
	}
	return nil
}

func (m *MockActivityRepository) Delete(id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockActivityRepository) List(offset, limit int) ([]*models.Activity, error) {
	if m.ListFunc != nil {
		return m.ListFunc(offset, limit)
	}
	return nil, nil
}

func (m *MockActivityRepository) AddOwner(activityID, ownerID uuid.UUID, ownerType string) error {
	if m.AddOwnerFunc != nil {
		return m.AddOwnerFunc(activityID, ownerID, ownerType)
	}
	return nil
}

func (m *MockActivityRepository) RemoveOwner(activityID, ownerID uuid.UUID) error {
	if m.RemoveOwnerFunc != nil {
		return m.RemoveOwnerFunc(activityID, ownerID)
	}
	return nil
}

func (m *MockActivityRepository) GetOwners(activityID uuid.UUID) ([]*models.ActivityOwner, error) {
	if m.GetOwnersFunc != nil {
		return m.GetOwnersFunc(activityID)
	}
	return nil, nil
}

func (m *MockActivityRepository) GetUserActivities(userID uuid.UUID) ([]*models.Activity, error) {
	if m.GetUserActivitiesFunc != nil {
		return m.GetUserActivitiesFunc(userID)
	}
	return nil, nil
}

func (m *MockActivityRepository) GetTribeActivities(tribeID uuid.UUID) ([]*models.Activity, error) {
	if m.GetTribeActivitiesFunc != nil {
		return m.GetTribeActivitiesFunc(tribeID)
	}
	return nil, nil
}

func (m *MockActivityRepository) ShareWithTribe(activityID, tribeID, userID uuid.UUID, expiresAt *time.Time) error {
	if m.ShareWithTribeFunc != nil {
		return m.ShareWithTribeFunc(activityID, tribeID, userID, expiresAt)
	}
	return nil
}

func (m *MockActivityRepository) UnshareWithTribe(activityID, tribeID uuid.UUID) error {
	if m.UnshareWithTribeFunc != nil {
		return m.UnshareWithTribeFunc(activityID, tribeID)
	}
	return nil
}

func (m *MockActivityRepository) GetSharedActivities(tribeID uuid.UUID) ([]*models.Activity, error) {
	if m.GetSharedActivitiesFunc != nil {
		return m.GetSharedActivitiesFunc(tribeID)
	}
	return nil, nil
}

func (m *MockActivityRepository) MarkForDeletion(activityID uuid.UUID) error {
	if m.MarkForDeletionFunc != nil {
		return m.MarkForDeletionFunc(activityID)
	}
	return nil
}

func (m *MockActivityRepository) CleanupOrphanedActivities() error {
	if m.CleanupOrphanedActivitiesFunc != nil {
		return m.CleanupOrphanedActivitiesFunc()
	}
	return nil
}
