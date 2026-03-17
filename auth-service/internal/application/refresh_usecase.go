package application

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/internal/ports/output"
)

// RefreshInput holds the parameters for token renewal.
type RefreshInput struct {
	RefreshToken string
	ClientType   domain.ClientType
}

// RefreshOutput holds the new token pair.
type RefreshOutput struct {
	Tokens domain.TokenPair
}

// RefreshUseCase renews an access token using a valid refresh token.
type RefreshUseCase struct {
	keycloak        output.KeycloakPort
	webClientID     string
	webClientSecret string
	mobileClientID  string
	appClientID     string
	logger          *zap.Logger
}

// NewRefreshUseCase constructs a RefreshUseCase.
func NewRefreshUseCase(
	keycloak output.KeycloakPort,
	webClientID, webClientSecret, mobileClientID, appClientID string,
	logger *zap.Logger,
) *RefreshUseCase {
	return &RefreshUseCase{
		keycloak:        keycloak,
		webClientID:     webClientID,
		webClientSecret: webClientSecret,
		mobileClientID:  mobileClientID,
		appClientID:     appClientID,
		logger:          logger,
	}
}

// Execute refreshes the token pair for the given client type.
func (uc *RefreshUseCase) Execute(ctx context.Context, input RefreshInput) (RefreshOutput, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "usecase.Refresh")
	defer span.End()

	if !input.ClientType.IsValid() {
		return RefreshOutput{}, domain.ErrInvalidClientType
	}

	clientID, clientSecret := uc.clientCredentials(input.ClientType)

	tokens, err := uc.keycloak.RefreshToken(ctx, clientID, clientSecret, input.RefreshToken)
	if err != nil {
		uc.logger.Warn("token refresh failed",
			zap.String("client_type", string(input.ClientType)),
			zap.Error(err),
		)
		return RefreshOutput{}, err
	}

	uc.logger.Info("token refreshed", zap.String("client_type", string(input.ClientType)))
	return RefreshOutput{Tokens: tokens}, nil
}

func (uc *RefreshUseCase) clientCredentials(ct domain.ClientType) (clientID, clientSecret string) {
	switch ct {
	case domain.ClientTypeMobile:
		return uc.mobileClientID, ""
	case domain.ClientTypeApp:
		return uc.appClientID, ""
	default:
		return uc.webClientID, uc.webClientSecret
	}
}
