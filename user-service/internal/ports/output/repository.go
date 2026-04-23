package output

import (
	"context"

	"github.com/project/user-service/internal/domain"
)

// UserProfileRepository define as operações de persistência para perfis de usuário.
type UserProfileRepository interface {
	// FindByID busca um perfil pelo ID (claim sub do JWT).
	// Retorna domain.ErrUserNotFound se não existir.
	FindByID(ctx context.Context, id string) (*domain.UserProfile, error)

	// Save persiste um novo perfil de usuário.
	Save(ctx context.Context, profile *domain.UserProfile) error

	// Update atualiza os campos de um perfil existente.
	// Retorna domain.ErrUserNotFound se o perfil não existir.
	Update(ctx context.Context, id string, input domain.UpdateProfileInput) (*domain.UserProfile, error)
}
