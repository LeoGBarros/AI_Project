package output

import (
	"context"

	"github.com/project/auth-service/internal/domain"
)

// KeycloakPort defines all interactions the application layer has with Keycloak.
// Adapters implementing this interface must not leak Keycloak-specific errors — they must
// translate them into domain errors before returning.
type KeycloakPort interface {
	// ExchangePassword authenticates a user via Resource Owner Password Credentials (ROPC).
	// Used by mobile and desktop app clients.
	ExchangePassword(ctx context.Context, clientID, clientSecret, username, password string) (domain.TokenPair, error)

	// ExchangeCode exchanges an authorization code for a token pair during the PKCE callback.
	// Used by web clients.
	ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (domain.TokenPair, error)

	// GetAuthorizationURL builds the Keycloak authorization endpoint URL that the web client
	// should be redirected to in order to begin the Authorization Code + PKCE flow.
	GetAuthorizationURL(clientID, redirectURI, state, codeChallenge, codeChallengeMethod string) string

	// RefreshToken exchanges a refresh token for a new token pair.
	RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (domain.TokenPair, error)

	// RevokeToken revokes an access or refresh token, effectively logging the user out.
	RevokeToken(ctx context.Context, clientID, clientSecret, token, tokenTypeHint string) error
}
