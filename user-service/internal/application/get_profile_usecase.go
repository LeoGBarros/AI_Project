// Package application contains the use cases that orchestrate domain logic
// and port interactions for the user-service.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/project/user-service/internal/domain"
	"github.com/project/user-service/internal/ports/output"
	"go.uber.org/zap"
)

// GetProfileInput holds the data extracted from JWT claims needed to
// retrieve or auto-create a user profile.
type GetProfileInput struct {
	UserID           string // claim "sub"
	PreferredUsername string // claim "preferred_username"
	Email            string // claim "email"
}

// GetProfileOutput wraps the resulting user profile.
type GetProfileOutput struct {
	Profile *domain.UserProfile
}

// GetProfileUseCase handles the retrieval of a user profile.
// If the profile does not exist, it auto-creates one using JWT claims data.
type GetProfileUseCase struct {
	repo   output.UserProfileRepository
	logger *zap.Logger
}

// NewGetProfileUseCase creates a GetProfileUseCase with the given repository and logger.
func NewGetProfileUseCase(repo output.UserProfileRepository, logger *zap.Logger) *GetProfileUseCase {
	return &GetProfileUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute looks up the user profile by UserID. If the profile exists, it is
// returned directly. If it does not exist (domain.ErrUserNotFound), a new
// profile is created from the JWT claims and persisted via the repository.
func (uc *GetProfileUseCase) Execute(ctx context.Context, input GetProfileInput) (*GetProfileOutput, error) {
	if input.UserID == "" {
		return nil, domain.ErrInvalidToken
	}

	profile, err := uc.repo.FindByID(ctx, input.UserID)
	if err == nil {
		return &GetProfileOutput{Profile: profile}, nil
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	// Auto-create profile from JWT claims.
	now := time.Now()
	profile = &domain.UserProfile{
		ID:          input.UserID,
		DisplayName: input.PreferredUsername,
		Email:       input.Email,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.Save(ctx, profile); err != nil {
		return nil, err
	}

	uc.logger.Info("auto-created user profile from JWT claims",
		zap.String("user_id", input.UserID),
	)

	return &GetProfileOutput{Profile: profile}, nil
}
