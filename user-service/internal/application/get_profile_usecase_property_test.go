package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project/user-service/internal/domain"
	"go.uber.org/zap"
	"pgregory.net/rapid"
)

// inMemoryUserProfileRepository is a mock implementation of ports/output.UserProfileRepository
// backed by an in-memory map, used for property-based testing.
type inMemoryUserProfileRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.UserProfile
}

func newInMemoryUserProfileRepository() *inMemoryUserProfileRepository {
	return &inMemoryUserProfileRepository{
		store: make(map[string]*domain.UserProfile),
	}
}

func (r *inMemoryUserProfileRepository) FindByID(_ context.Context, id string) (*domain.UserProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.store[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return p, nil
}

func (r *inMemoryUserProfileRepository) Save(_ context.Context, profile *domain.UserProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[profile.ID] = profile
	return nil
}

func (r *inMemoryUserProfileRepository) Update(_ context.Context, id string, input domain.UpdateProfileInput) (*domain.UserProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.store[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	if input.DisplayName != nil {
		p.DisplayName = *input.DisplayName
	}
	if input.Email != nil {
		p.Email = *input.Email
	}
	if input.Phone != nil {
		p.Phone = *input.Phone
	}
	if input.AvatarURL != nil {
		p.AvatarURL = *input.AvatarURL
	}
	p.UpdatedAt = time.Now()
	return p, nil
}

// seed pre-populates the repository with a profile for the given ID.
func (r *inMemoryUserProfileRepository) seed(profile *domain.UserProfile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[profile.ID] = profile
}

// Feature: user-service, Property 2: GetProfile sempre retorna perfil válido
// **Validates: Requirements 1.2, 1.3**
//
// For any valid user_id (non-empty UUID) and associated JWT claims, the GetProfile
// use case must always return a UserProfile with ID equal to the provided user_id —
// either by querying an existing profile in the repository, or by auto-creating a
// new profile with the JWT claims data when the profile does not exist.
func TestProperty_GetProfileAlwaysReturnsValidProfile(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := newInMemoryUserProfileRepository()
		logger := zap.NewNop()
		uc := NewGetProfileUseCase(repo, logger)

		// Generate a random valid UUID for user_id.
		userID := uuid.New().String()

		// Generate random JWT claim values.
		preferredUsername := rapid.StringMatching(`^[a-zA-Z][a-zA-Z0-9_]{2,19}$`).Draw(t, "preferredUsername")
		email := rapid.StringMatching(`^[a-z]{3,8}@[a-z]{3,8}\.[a-z]{2,4}$`).Draw(t, "email")

		// Decide whether the profile already exists in the repo.
		profileExists := rapid.Bool().Draw(t, "profileExists")

		if profileExists {
			existing := &domain.UserProfile{
				ID:          userID,
				DisplayName: rapid.StringMatching(`^[a-zA-Z ]{1,30}$`).Draw(t, "existingDisplayName"),
				Email:       rapid.StringMatching(`^[a-z]{3,8}@[a-z]{3,8}\.[a-z]{2,4}$`).Draw(t, "existingEmail"),
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			}
			repo.seed(existing)
		}

		input := GetProfileInput{
			UserID:           userID,
			PreferredUsername: preferredUsername,
			Email:            email,
		}

		out, err := uc.Execute(context.Background(), input)

		// Property: Execute must not return an error for a valid user_id.
		if err != nil {
			t.Fatalf("expected no error for valid user_id %q, got: %v", userID, err)
		}

		// Property: output must be non-nil.
		if out == nil {
			t.Fatal("expected non-nil output")
		}

		// Property: the returned profile must be non-nil.
		if out.Profile == nil {
			t.Fatal("expected non-nil profile in output")
		}

		// Property: the returned profile ID must equal the input user_id.
		if out.Profile.ID != userID {
			t.Fatalf("expected profile ID %q, got %q", userID, out.Profile.ID)
		}
	})
}
