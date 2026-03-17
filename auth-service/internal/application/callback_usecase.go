package application

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/internal/ports/output"
)

// CallbackInput holds the parameters received from the Keycloak redirect.
type CallbackInput struct {
	// Code is the authorization code issued by Keycloak.
	Code string

	// State is the opaque value that must match the stored PKCE state.
	StateID string

	// ClientSecret is the web client's secret, looked up by the service (not sent by the browser).
	ClientSecret string
}

// CallbackOutput holds the token pair returned after a successful code exchange.
type CallbackOutput struct {
	Tokens domain.TokenPair
}

// CallbackUseCase processes the OAuth authorization code callback for the web PKCE flow.
// It validates the state, exchanges the code for tokens, and deletes the consumed state.
type CallbackUseCase struct {
	keycloak   output.KeycloakPort
	stateStore output.PKCEStateStore
	logger     *zap.Logger
}

// NewCallbackUseCase constructs a CallbackUseCase.
func NewCallbackUseCase(
	keycloak output.KeycloakPort,
	stateStore output.PKCEStateStore,
	logger *zap.Logger,
) *CallbackUseCase {
	return &CallbackUseCase{
		keycloak:   keycloak,
		stateStore: stateStore,
		logger:     logger,
	}
}

// Execute validates the state, exchanges the authorization code, and cleans up the state.
func (uc *CallbackUseCase) Execute(ctx context.Context, input CallbackInput) (CallbackOutput, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "usecase.Callback")
	defer span.End()

	pkceState, err := uc.stateStore.Get(ctx, input.StateID)
	if err != nil {
		uc.logger.Warn("invalid or missing oauth state", zap.String("state_id", input.StateID))
		return CallbackOutput{}, domain.ErrInvalidState
	}

	if pkceState.IsExpired() {
		_ = uc.stateStore.Delete(ctx, input.StateID)
		return CallbackOutput{}, domain.ErrInvalidState
	}

	tokens, err := uc.keycloak.ExchangeCode(
		ctx,
		pkceState.ClientID,
		input.ClientSecret,
		input.Code,
		pkceState.RedirectURI,
		pkceState.CodeVerifier,
	)
	if err != nil {
		uc.logger.Error("code exchange failed", zap.Error(err))
		return CallbackOutput{}, err
	}

	// State consumed — delete to prevent replay.
	if delErr := uc.stateStore.Delete(ctx, input.StateID); delErr != nil {
		uc.logger.Warn("failed to delete consumed pkce state", zap.String("state_id", input.StateID), zap.Error(delErr))
	}

	uc.logger.Info("pkce callback completed", zap.String("state_id", input.StateID))
	return CallbackOutput{Tokens: tokens}, nil
}
