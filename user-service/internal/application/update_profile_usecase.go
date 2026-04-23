// Package application contains the use cases that orchestrate domain logic
// and port interactions for the user-service.
package application

import (
	"context"
	"fmt"
	"time"

	"github.com/project/user-service/internal/domain"
	"github.com/project/user-service/internal/ports/output"
	"go.uber.org/zap"
)

// UpdateProfileInput holds the user ID and the fields to update.
type UpdateProfileInput struct {
	UserID               string
	domain.UpdateProfileInput
}

// UpdateProfileOutput wraps the updated user profile.
type UpdateProfileOutput struct {
	Profile *domain.UserProfile
}

// UpdateProfileUseCase handles the update of a user profile.
// It validates input, updates the profile in the repository, and publishes
// an event. If event publishing fails, it logs the error without rolling back.
type UpdateProfileUseCase struct {
	repo      output.UserProfileRepository
	publisher output.EventPublisher
	logger    *zap.Logger
}

// NewUpdateProfileUseCase creates an UpdateProfileUseCase with the given dependencies.
func NewUpdateProfileUseCase(repo output.UserProfileRepository, publisher output.EventPublisher, logger *zap.Logger) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// Execute validates the input, updates the profile in the repository,
// publishes a user.updated event, and returns the updated profile.
// If event publishing fails, it logs the error without rolling back.
func (uc *UpdateProfileUseCase) Execute(ctx context.Context, input UpdateProfileInput) (*UpdateProfileOutput, error) {
	if input.UserID == "" {
		return nil, domain.ErrInvalidToken
	}

	if validationErrors := input.UpdateProfileInput.Validate(); len(validationErrors) > 0 {
		return nil, domain.ErrValidation
	}

	updatedFields := make(map[string]any)
	if input.DisplayName != nil {
		updatedFields["display_name"] = *input.DisplayName
	}
	if input.Email != nil {
		updatedFields["email"] = *input.Email
	}
	if input.Phone != nil {
		updatedFields["phone"] = *input.Phone
	}
	if input.AvatarURL != nil {
		updatedFields["avatar_url"] = *input.AvatarURL
	}

	profile, err := uc.repo.Update(ctx, input.UserID, input.UpdateProfileInput)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	if err := uc.publisher.PublishUserUpdated(ctx, input.UserID, updatedFields); err != nil {
		uc.logger.Error("failed to publish user.updated event",
			zap.String("user_id", input.UserID),
			zap.Error(err),
		)
	} else {
		uc.logger.Info("published user.updated event",
			zap.String("user_id", input.UserID),
			zap.Time("timestamp", time.Now()),
		)
	}

	return &UpdateProfileOutput{Profile: profile}, nil
}