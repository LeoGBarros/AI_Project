package application

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/internal/ports/output"
)

// LogoutInput holds the parameters for the logout operation.
type LogoutInput struct {
	// RefreshToken is revoked to end the Keycloak session.
	RefreshToken string
	ClientType   domain.ClientType
}

// LogoutUseCase revokes the user's refresh token in Keycloak, ending the session.
type LogoutUseCase struct {
	keycloak        output.KeycloakPort
	webClientID     string
	webClientSecret string
	mobileClientID  string
	appClientID     string
	logger          *zap.Logger
}

// NewLogoutUseCase constructs a LogoutUseCase.
func NewLogoutUseCase(
	keycloak output.KeycloakPort,
	webClientID, webClientSecret, mobileClientID, appClientID string,
	logger *zap.Logger,
) *LogoutUseCase {
	return &LogoutUseCase{
		keycloak:        keycloak,
		webClientID:     webClientID,
		webClientSecret: webClientSecret,
		mobileClientID:  mobileClientID,
		appClientID:     appClientID,
		logger:          logger,
	}
}

// Execute revokes the refresh token in Keycloak.
func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "usecase.Logout")
	defer span.End()

	if !input.ClientType.IsValid() {
		return domain.ErrInvalidClientType
	}

	clientID, clientSecret := uc.clientCredentials(input.ClientType)

	if err := uc.keycloak.RevokeToken(ctx, clientID, clientSecret, input.RefreshToken, "refresh_token"); err != nil {
		uc.logger.Warn("logout revocation failed",
			zap.String("client_type", string(input.ClientType)),
			zap.Error(err),
		)
		return err
	}

	uc.logger.Info("logout successful", zap.String("client_type", string(input.ClientType)))
	return nil
}

func (uc *LogoutUseCase) clientCredentials(ct domain.ClientType) (clientID, clientSecret string) {
	switch ct {
	case domain.ClientTypeMobile:
		return uc.mobileClientID, ""
	case domain.ClientTypeApp:
		return uc.appClientID, ""
	default:
		return uc.webClientID, uc.webClientSecret
	}
}
