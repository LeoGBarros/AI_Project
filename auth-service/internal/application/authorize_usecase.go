package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/internal/ports/output"
)

const pkceStateTTL = 5 * time.Minute

// AuthorizeInput holds the parameters required to begin the PKCE flow.
type AuthorizeInput struct {
	RedirectURI string
	ClientID    string
}

// AuthorizeOutput contains the Keycloak authorization URL to redirect the browser to.
type AuthorizeOutput struct {
	// AuthorizationURL is the URL the web client must redirect the browser to.
	AuthorizationURL string
}

// AuthorizeUseCase initiates the Authorization Code + PKCE flow for web clients.
// It generates a cryptographically secure code_verifier, derives the code_challenge (S256),
// stores the state in Redis, and returns the full Keycloak authorization URL.
type AuthorizeUseCase struct {
	keycloak   output.KeycloakPort
	stateStore output.PKCEStateStore
	logger     *zap.Logger
}

// NewAuthorizeUseCase constructs an AuthorizeUseCase.
func NewAuthorizeUseCase(
	keycloak output.KeycloakPort,
	stateStore output.PKCEStateStore,
	logger *zap.Logger,
) *AuthorizeUseCase {
	return &AuthorizeUseCase{
		keycloak:   keycloak,
		stateStore: stateStore,
		logger:     logger,
	}
}

// Execute generates PKCE parameters, persists the state, and builds the authorization URL.
func (uc *AuthorizeUseCase) Execute(ctx context.Context, input AuthorizeInput) (AuthorizeOutput, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "usecase.Authorize")
	defer span.End()

	stateID, err := generateRandomString(32)
	if err != nil {
		return AuthorizeOutput{}, err
	}

	codeVerifier, err := generateRandomString(64)
	if err != nil {
		return AuthorizeOutput{}, err
	}

	codeChallenge := deriveCodeChallenge(codeVerifier)

	pkceState := domain.PKCEState{
		StateID:      stateID,
		CodeVerifier: codeVerifier,
		RedirectURI:  input.RedirectURI,
		ClientID:     input.ClientID,
		ExpiresAt:    time.Now().Add(pkceStateTTL),
	}

	if err := uc.stateStore.Save(ctx, pkceState); err != nil {
		uc.logger.Error("failed to save pkce state", zap.Error(err))
		return AuthorizeOutput{}, err
	}

	authURL := uc.keycloak.GetAuthorizationURL(
		input.ClientID,
		input.RedirectURI,
		stateID,
		codeChallenge,
		"S256",
	)

	uc.logger.Info("pkce flow initiated", zap.String("state_id", stateID))
	return AuthorizeOutput{AuthorizationURL: authURL}, nil
}

// generateRandomString returns a URL-safe base64-encoded random string of n bytes.
func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// deriveCodeChallenge computes the S256 PKCE code challenge from the code verifier.
func deriveCodeChallenge(codeVerifier string) string {
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
