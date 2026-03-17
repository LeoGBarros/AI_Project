package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/project/auth-service/internal/domain"
)

// PKCEStateStore is a testify mock for output.PKCEStateStore.
type PKCEStateStore struct {
	mock.Mock
}

func (m *PKCEStateStore) Save(ctx context.Context, state domain.PKCEState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *PKCEStateStore) Get(ctx context.Context, stateID string) (domain.PKCEState, error) {
	args := m.Called(ctx, stateID)
	return args.Get(0).(domain.PKCEState), args.Error(1)
}

func (m *PKCEStateStore) Delete(ctx context.Context, stateID string) error {
	args := m.Called(ctx, stateID)
	return args.Error(0)
}
