package application

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/internal/ports/output"
)

// LoginInput holds the parameters for the ROPC login use case.
type LoginInput struct {
	Username   string
	Password   string
	ClientType domain.ClientType
}

// LoginOutput is the result returned to the caller on successful login.
type LoginOutput struct {
	Tokens domain.TokenPair
}

// LoginUseCase authenticates users via Resource Owner Password Credentials (ROPC).
// Intended for mobile and desktop app clients only.
type LoginUseCase struct {
	keycloak         output.KeycloakPort
	webClientID      string
	webClientSecret  string
	mobileClientID   string
	appClientID      string
	logger           *zap.Logger
}

// NewLoginUseCase constructs a LoginUseCase with its required dependencies.
func NewLoginUseCase(
	keycloak output.KeycloakPort,
	webClientID, webClientSecret, mobileClientID, appClientID string,
	logger *zap.Logger,
) *LoginUseCase {
	return &LoginUseCase{
		keycloak:        keycloak,
		webClientID:     webClientID,
		webClientSecret: webClientSecret,
		mobileClientID:  mobileClientID,
		appClientID:     appClientID,
		logger:          logger,
	}
}

// Execute performs ROPC authentication for a given client type.
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (LoginOutput, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "usecase.Login")
	defer span.End()

	span.SetAttributes(attribute.String("client_type", string(input.ClientType)))

	if !input.ClientType.IsValid() {
		return LoginOutput{}, domain.ErrInvalidClientType
	}
	if !input.ClientType.IsROPC() {
		return LoginOutput{}, fmt.Errorf("%w: client type '%s' must use PKCE flow", domain.ErrInvalidClientType, input.ClientType)
	}

	clientID, clientSecret := uc.clientCredentials(input.ClientType)

	tokens, err := uc.keycloak.ExchangePassword(ctx, clientID, clientSecret, input.Username, input.Password)
	if err != nil {
		uc.logger.Warn("login failed",
			zap.String("client_type", string(input.ClientType)),
			zap.Error(err),
		)
		return LoginOutput{}, err
	}

	uc.logger.Info("login successful", zap.String("client_type", string(input.ClientType)))
	return LoginOutput{Tokens: tokens}, nil
}

func (uc *LoginUseCase) clientCredentials(ct domain.ClientType) (clientID, clientSecret string) {
	switch ct {
	case domain.ClientTypeMobile:
		return uc.mobileClientID, ""
	case domain.ClientTypeApp:
		return uc.appClientID, ""
	default:
		return uc.webClientID, uc.webClientSecret
	}
}
