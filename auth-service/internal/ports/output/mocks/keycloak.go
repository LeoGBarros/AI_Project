package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/project/auth-service/internal/domain"
)

// KeycloakPort is a testify mock for output.KeycloakPort.
type KeycloakPort struct {
	mock.Mock
}

func (m *KeycloakPort) ExchangePassword(ctx context.Context, clientID, clientSecret, username, password string) (domain.TokenPair, error) {
	args := m.Called(ctx, clientID, clientSecret, username, password)
	return args.Get(0).(domain.TokenPair), args.Error(1)
}

func (m *KeycloakPort) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (domain.TokenPair, error) {
	args := m.Called(ctx, clientID, clientSecret, code, redirectURI, codeVerifier)
	return args.Get(0).(domain.TokenPair), args.Error(1)
}

func (m *KeycloakPort) GetAuthorizationURL(clientID, redirectURI, state, codeChallenge, codeChallengeMethod string) string {
	args := m.Called(clientID, redirectURI, state, codeChallenge, codeChallengeMethod)
	return args.String(0)
}

func (m *KeycloakPort) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (domain.TokenPair, error) {
	args := m.Called(ctx, clientID, clientSecret, refreshToken)
	return args.Get(0).(domain.TokenPair), args.Error(1)
}

func (m *KeycloakPort) RevokeToken(ctx context.Context, clientID, clientSecret, token, tokenTypeHint string) error {
	args := m.Called(ctx, clientID, clientSecret, token, tokenTypeHint)
	return args.Error(0)
}
