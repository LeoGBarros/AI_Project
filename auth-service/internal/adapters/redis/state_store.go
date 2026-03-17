package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/project/auth-service/internal/domain"
)

const (
	pkceStateTTL    = 5 * time.Minute
	keyPrefix       = "auth-service:pkce-state:"
)

// PKCEStateStore implements output.PKCEStateStore using Redis.
// Keys follow the namespacing convention: auth-service:pkce-state:{stateID}
type PKCEStateStore struct {
	client *goredis.Client
	logger *zap.Logger
}

// NewPKCEStateStore constructs a Redis-backed PKCEStateStore.
func NewPKCEStateStore(client *goredis.Client, logger *zap.Logger) *PKCEStateStore {
	return &PKCEStateStore{client: client, logger: logger}
}

func (s *PKCEStateStore) key(stateID string) string {
	return keyPrefix + stateID
}

// Save persists the PKCEState in Redis with a fixed TTL of 5 minutes.
func (s *PKCEStateStore) Save(ctx context.Context, state domain.PKCEState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal pkce state: %w", err)
	}

	if err := s.client.Set(ctx, s.key(state.StateID), data, pkceStateTTL).Err(); err != nil {
		s.logger.Error("redis: failed to save pkce state",
			zap.String("state_id", state.StateID),
			zap.Error(err),
		)
		return fmt.Errorf("save pkce state: %w", err)
	}

	return nil
}

// Get retrieves a PKCEState by its StateID.
// Returns domain.ErrInvalidState when the key does not exist or has expired.
func (s *PKCEStateStore) Get(ctx context.Context, stateID string) (domain.PKCEState, error) {
	data, err := s.client.Get(ctx, s.key(stateID)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.PKCEState{}, domain.ErrInvalidState
		}
		s.logger.Error("redis: failed to get pkce state",
			zap.String("state_id", stateID),
			zap.Error(err),
		)
		return domain.PKCEState{}, fmt.Errorf("get pkce state: %w", err)
	}

	var state domain.PKCEState
	if err := json.Unmarshal(data, &state); err != nil {
		return domain.PKCEState{}, fmt.Errorf("unmarshal pkce state: %w", err)
	}

	return state, nil
}

// Delete removes a PKCEState key after it has been consumed in the callback.
func (s *PKCEStateStore) Delete(ctx context.Context, stateID string) error {
	if err := s.client.Del(ctx, s.key(stateID)).Err(); err != nil {
		s.logger.Warn("redis: failed to delete pkce state",
			zap.String("state_id", stateID),
			zap.Error(err),
		)
		return fmt.Errorf("delete pkce state: %w", err)
	}
	return nil
}
