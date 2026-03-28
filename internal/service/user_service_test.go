package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/shopspring/decimal"
)

// setupUserService creates a mock repository and a user service for testing.
func setupUserService() (*repository.MockUserRepository, UserService) {
	mock := &repository.MockUserRepository{}
	svc := NewUserService(mock)
	return mock, svc
}

// teardownUserService cleans up after each test (placeholder for future use).
func teardownUserService() {}

func TestCreateOrFetchUser(t *testing.T) {
	userCtx := model.UserContext{
		FirebaseId: "firebase-123",
		Email:      "test@example.com",
		Name:       "Test User",
	}

	expectedUser := &model.User{
		FirebaseId: "firebase-123",
		Name:       "Test User",
		Email:      "test@example.com",
		Balance:    decimal.RequireFromString("500.00"),
	}

	tests := []struct {
		name           string
		setupMock      func(mock *repository.MockUserRepository)
		expectedUser   *model.User
		expectedStatus model.FetchedOrCreated
		expectError    bool
	}{
		{
			name: "user already exists - returns fetched user",
			setupMock: func(mock *repository.MockUserRepository) {
				mock.GetUserFn = func(ctx model.UserContext) (*model.User, error) {
					return expectedUser, nil
				}
			},
			expectedUser:   expectedUser,
			expectedStatus: model.UserFetched,
			expectError:    false,
		},
		{
			name: "user not found - creates new user",
			setupMock: func(mock *repository.MockUserRepository) {
				mock.GetUserFn = func(ctx model.UserContext) (*model.User, error) {
					return nil, nil
				}
				mock.CreateUserFn = func(ctx model.UserContext) (*model.User, error) {
					return expectedUser, nil
				}
			},
			expectedUser:   expectedUser,
			expectedStatus: model.UserCreated,
			expectError:    false,
		},
		{
			name: "GetUser fails - returns error",
			setupMock: func(mock *repository.MockUserRepository) {
				mock.GetUserFn = func(ctx model.UserContext) (*model.User, error) {
					return nil, errors.New("database connection error")
				}
			},
			expectedUser:   nil,
			expectedStatus: model.None,
			expectError:    true,
		},
		{
			name: "CreateUser fails - returns error",
			setupMock: func(mock *repository.MockUserRepository) {
				mock.GetUserFn = func(ctx model.UserContext) (*model.User, error) {
					return nil, nil
				}
				mock.CreateUserFn = func(ctx model.UserContext) (*model.User, error) {
					return nil, errors.New("failed to create user")
				}
			},
			expectedUser:   nil,
			expectedStatus: model.None,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, svc := setupUserService()
			defer teardownUserService()

			tt.setupMock(mock)

			user, status, err := svc.CreateOrFetchUser(userCtx)

			if tt.expectError && err == appErrors.ErrNone {
				t.Fatal("expected error but got none")
			}
			if !tt.expectError && err != appErrors.ErrNone {
				t.Fatalf("expected no error but got: %v", err)
			}
			if status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, status)
			}
			if tt.expectedUser != nil && user.FirebaseId != tt.expectedUser.FirebaseId {
				t.Errorf("expected user ID %s, got %s", tt.expectedUser.FirebaseId, user.FirebaseId)
			}
		})
	}
}

func TestUpdateFields(t *testing.T) {
	userCtx := model.UserContext{
		FirebaseId: "firebase-123",
		Email:      "test@example.com",
		Name:       "Test User",
	}

	newEmail := "new@example.com"
	newName := "New Name"
	whitespace := "   "

	tests := []struct {
		name            string
		fields          model.UpdateFieldsDto
		setupMock       func(mock *repository.MockUserRepository)
		expectedUpdates map[string]interface{}
		expectError     bool
	}{
		{
			name:   "updates both email and name",
			fields: model.UpdateFieldsDto{Email: &newEmail, Name: &newName},
			setupMock: func(mock *repository.MockUserRepository) {
				mock.UpdateUserFn = func(ctx model.UserContext, updates map[string]interface{}) error {
					return nil
				}
				mock.GetUserFn = func(ctx model.UserContext) (*model.User, error) {
					return &model.User{FirebaseId: "firebase-123", Email: newEmail, Name: newName}, nil
				}
			},
			expectedUpdates: map[string]interface{}{"email": newEmail, "name": newName},
			expectError:     false,
		},
		{
			name:   "updates only email when name is nil",
			fields: model.UpdateFieldsDto{Email: &newEmail, Name: nil},
			setupMock: func(mock *repository.MockUserRepository) {
				mock.UpdateUserFn = func(ctx model.UserContext, updates map[string]interface{}) error {
					return nil
				}
				mock.GetUserFn = func(ctx model.UserContext) (*model.User, error) {
					return &model.User{FirebaseId: "firebase-123", Email: newEmail}, nil
				}
			},
			expectedUpdates: map[string]interface{}{"email": newEmail},
			expectError:     false,
		},
		{
			name:   "whitespace-only fields are skipped",
			fields: model.UpdateFieldsDto{Email: &whitespace, Name: &whitespace},
			setupMock: func(mock *repository.MockUserRepository) {
				mock.UpdateUserFn = func(ctx model.UserContext, updates map[string]interface{}) error {
					return errors.New("no fields to update")
				}
			},
			expectedUpdates: nil,
			expectError:     true,
		},
		{
			name:   "UpdateUser fails - returns error",
			fields: model.UpdateFieldsDto{Email: &newEmail, Name: &newName},
			setupMock: func(mock *repository.MockUserRepository) {
				mock.UpdateUserFn = func(ctx model.UserContext, updates map[string]interface{}) error {
					return errors.New("database error")
				}
			},
			expectedUpdates: nil,
			expectError:     true,
		},
		{
			name:   "UpdateUser succeeds but GetUser fails",
			fields: model.UpdateFieldsDto{Email: &newEmail},
			setupMock: func(mock *repository.MockUserRepository) {
				mock.UpdateUserFn = func(ctx model.UserContext, updates map[string]interface{}) error {
					return nil
				}
				mock.GetUserFn = func(ctx model.UserContext) (*model.User, error) {
					return nil, errors.New("database error")
				}
			},
			expectedUpdates: map[string]interface{}{"email": newEmail},
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, svc := setupUserService()
			defer teardownUserService()

			tt.setupMock(mock)

			// Wrap UpdateUserFn to capture what map was passed to the repository
			var capturedUpdates map[string]interface{}
			if mock.UpdateUserFn != nil {
				original := mock.UpdateUserFn
				mock.UpdateUserFn = func(ctx model.UserContext, updates map[string]interface{}) error {
					capturedUpdates = updates
					return original(ctx, updates)
				}
			}

			user, err := svc.UpdateFields(userCtx, tt.fields)

			if tt.expectError && err == appErrors.ErrNone {
				t.Fatal("expected error but got none")
			}
			if !tt.expectError && err != appErrors.ErrNone {
				t.Fatalf("expected no error but got: %v", err)
			}

			if !tt.expectError && user == nil {
				t.Fatal("expected user but got nil")
			}

			// Verify correct fields were sent to repository
			if tt.expectedUpdates != nil && !reflect.DeepEqual(capturedUpdates, tt.expectedUpdates) {
				t.Errorf("expected updates %+v, got %+v", tt.expectedUpdates, capturedUpdates)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	userCtx := model.UserContext{
		FirebaseId: "firebase-123",
		Email:      "test@example.com",
		Name:       "Test User",
	}

	tests := []struct {
		name        string
		setupMock   func(mock *repository.MockUserRepository)
		expectError bool
	}{
		{
			name: "delete succeeds",
			setupMock: func(mock *repository.MockUserRepository) {
				mock.DeleteUserFn = func(ctx model.UserContext) error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "delete fails - returns error",
			setupMock: func(mock *repository.MockUserRepository) {
				mock.DeleteUserFn = func(ctx model.UserContext) error {
					return errors.New("user not found")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, svc := setupUserService()
			defer teardownUserService()

			tt.setupMock(mock)

			err := svc.DeleteUser(userCtx)

			if tt.expectError && err == appErrors.ErrNone {
				t.Fatal("expected error but got none")
			}
			if !tt.expectError && err != appErrors.ErrNone {
				t.Fatalf("expected no error but got: %v", err)
			}
		})
	}
}
