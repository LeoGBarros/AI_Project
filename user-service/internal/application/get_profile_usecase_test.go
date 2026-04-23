package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/project/user-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestGetProfile_WithExistingUser_ShouldReturnProfile verifies that when a
// profile already exists in the repository, Execute returns it directly
// without creating a new one.
// Validates: Requirements 1.1, 1.2
func TestGetProfile_WithExistingUser_ShouldReturnProfile(t *testing.T) {
	repo := newInMemoryUserProfileRepository()
	logger := zap.NewNop()
	uc := NewGetProfileUseCase(repo, logger)

	existing := &domain.UserProfile{
		ID:          "user-123",
		DisplayName: "Alice",
		Email:       "alice@example.com",
		Phone:       "+5511999990000",
		AvatarURL:   "https://cdn.example.com/alice.png",
		CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	repo.seed(existing)

	input := GetProfileInput{
		UserID:           "user-123",
		PreferredUsername: "alice_new",
		Email:            "alice_new@example.com",
	}

	out, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, out.Profile)
	assert.Equal(t, "user-123", out.Profile.ID)
	assert.Equal(t, "Alice", out.Profile.DisplayName, "should return existing display name, not JWT claim")
	assert.Equal(t, "alice@example.com", out.Profile.Email, "should return existing email, not JWT claim")
	assert.Equal(t, "+5511999990000", out.Profile.Phone)
	assert.Equal(t, "https://cdn.example.com/alice.png", out.Profile.AvatarURL)
}

// TestGetProfile_WithNewUser_ShouldAutoCreate verifies that when the profile
// does not exist in the repository, Execute auto-creates one using JWT claims.
// Validates: Requirements 1.2, 1.3
func TestGetProfile_WithNewUser_ShouldAutoCreate(t *testing.T) {
	repo := newInMemoryUserProfileRepository()
	logger := zap.NewNop()
	uc := NewGetProfileUseCase(repo, logger)

	input := GetProfileInput{
		UserID:           "new-user-456",
		PreferredUsername: "bob",
		Email:            "bob@example.com",
	}

	before := time.Now()
	out, err := uc.Execute(context.Background(), input)
	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, out.Profile)

	p := out.Profile
	assert.Equal(t, "new-user-456", p.ID)
	assert.Equal(t, "bob", p.DisplayName, "display name should come from preferred_username claim")
	assert.Equal(t, "bob@example.com", p.Email, "email should come from email claim")
	assert.Empty(t, p.Phone, "phone should be empty for auto-created profile")
	assert.Empty(t, p.AvatarURL, "avatar_url should be empty for auto-created profile")
	assert.False(t, p.CreatedAt.Before(before), "created_at should be >= test start time")
	assert.False(t, p.CreatedAt.After(after), "created_at should be <= test end time")
	assert.False(t, p.UpdatedAt.Before(before), "updated_at should be >= test start time")
	assert.False(t, p.UpdatedAt.After(after), "updated_at should be <= test end time")

	// Verify the profile was actually persisted in the repository.
	stored, err := repo.FindByID(context.Background(), "new-user-456")
	require.NoError(t, err)
	assert.Equal(t, p.ID, stored.ID)
	assert.Equal(t, p.DisplayName, stored.DisplayName)
}

// TestGetProfile_WithEmptyUserID_ShouldReturnInvalidToken verifies that when
// user_id is empty, Execute returns domain.ErrInvalidToken.
// Validates: Requirements 1.1, 4.3
func TestGetProfile_WithEmptyUserID_ShouldReturnInvalidToken(t *testing.T) {
	repo := newInMemoryUserProfileRepository()
	logger := zap.NewNop()
	uc := NewGetProfileUseCase(repo, logger)

	input := GetProfileInput{
		UserID:           "",
		PreferredUsername: "ghost",
		Email:            "ghost@example.com",
	}

	out, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidToken), "expected ErrInvalidToken, got: %v", err)
	assert.Nil(t, out)
}

// TestGetProfile_WithRepoError_ShouldPropagateError verifies that when the
// repository returns an unexpected error (not ErrUserNotFound), Execute
// propagates it to the caller.
// Validates: Requirements 1.2
func TestGetProfile_WithRepoError_ShouldPropagateError(t *testing.T) {
	errDB := errors.New("connection refused")
	repo := &failingRepository{err: errDB}
	logger := zap.NewNop()
	uc := NewGetProfileUseCase(repo, logger)

	input := GetProfileInput{
		UserID:           "user-789",
		PreferredUsername: "charlie",
		Email:            "charlie@example.com",
	}

	out, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errDB), "expected the original repo error, got: %v", err)
	assert.Nil(t, out)
}

// failingRepository is a mock that always returns a configured error from FindByID.
// It is used to test error propagation from the repository layer.
type failingRepository struct {
	err error
}

func (r *failingRepository) FindByID(_ context.Context, _ string) (*domain.UserProfile, error) {
	return nil, r.err
}

func (r *failingRepository) Save(_ context.Context, _ *domain.UserProfile) error {
	return r.err
}

func (r *failingRepository) Update(_ context.Context, _ string, _ domain.UpdateProfileInput) (*domain.UserProfile, error) {
	return nil, r.err
}
