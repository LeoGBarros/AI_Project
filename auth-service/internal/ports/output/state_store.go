package output

import (
	"context"

	"github.com/project/auth-service/internal/domain"
)

// PKCEStateStore defines the persistence contract for OAuth PKCE flow state objects.
// The state is stored transiently (TTL = 5 minutes) and consumed exactly once on callback.
type PKCEStateStore interface {
	// Save persists a PKCEState using its StateID as the key.
	// Returns an error if the underlying store is unavailable.
	Save(ctx context.Context, state domain.PKCEState) error

	// Get retrieves a PKCEState by its StateID.
	// Returns domain.ErrInvalidState when the key does not exist or has expired.
	Get(ctx context.Context, stateID string) (domain.PKCEState, error)

	// Delete removes a PKCEState after it has been consumed.
	Delete(ctx context.Context, stateID string) error
}
