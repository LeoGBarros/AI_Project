package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/project/auth-service/internal/application"
	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/internal/ports/output/mocks"
)

const (
	testWebClientID     = "web-client"
	testWebClientSecret = "web-secret"
	testMobileClientID  = "mobile-client"
	testAppClientID     = "app-client"
)

func newLoginUseCase(t *testing.T, kc *mocks.KeycloakPort) *application.LoginUseCase {
	return application.NewLoginUseCase(
		kc,
		testWebClientID, testWebClientSecret,
		testMobileClientID,
		testAppClientID,
		zaptest.NewLogger(t),
	)
}

func TestLoginUseCase_Execute_WithMobileClientAndValidCredentials_ShouldReturnTokenPair(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	expectedTokens := domain.TokenPair{AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", ExpiresIn: 300}
	kc.On("ExchangePassword", mock.Anything, testMobileClientID, "", "user", "pass").Return(expectedTokens, nil)

	uc := newLoginUseCase(t, kc)
	out, err := uc.Execute(context.Background(), application.LoginInput{
		Username:   "user",
		Password:   "pass",
		ClientType: domain.ClientTypeMobile,
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, out.Tokens)
	kc.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WithAppClientAndValidCredentials_ShouldReturnTokenPair(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	expectedTokens := domain.TokenPair{AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer"}
	kc.On("ExchangePassword", mock.Anything, testAppClientID, "", "user", "pass").Return(expectedTokens, nil)

	uc := newLoginUseCase(t, kc)
	out, err := uc.Execute(context.Background(), application.LoginInput{
		Username:   "user",
		Password:   "pass",
		ClientType: domain.ClientTypeApp,
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, out.Tokens)
	kc.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WithWebClientType_ShouldReturnInvalidClientTypeError(t *testing.T) {
	kc := &mocks.KeycloakPort{}

	uc := newLoginUseCase(t, kc)
	_, err := uc.Execute(context.Background(), application.LoginInput{
		Username:   "user",
		Password:   "pass",
		ClientType: domain.ClientTypeWeb,
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidClientType))
	kc.AssertNotCalled(t, "ExchangePassword")
}

func TestLoginUseCase_Execute_WithUnknownClientType_ShouldReturnInvalidClientTypeError(t *testing.T) {
	kc := &mocks.KeycloakPort{}

	uc := newLoginUseCase(t, kc)
	_, err := uc.Execute(context.Background(), application.LoginInput{
		Username:   "user",
		Password:   "pass",
		ClientType: domain.ClientType("unknown"),
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidClientType))
	kc.AssertNotCalled(t, "ExchangePassword")
}

func TestLoginUseCase_Execute_WithInvalidCredentials_ShouldReturnInvalidCredentialsError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	kc.On("ExchangePassword", mock.Anything, testMobileClientID, "", "user", "wrong").
		Return(domain.TokenPair{}, domain.ErrInvalidCredentials)

	uc := newLoginUseCase(t, kc)
	_, err := uc.Execute(context.Background(), application.LoginInput{
		Username:   "user",
		Password:   "wrong",
		ClientType: domain.ClientTypeMobile,
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
	kc.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WhenKeycloakUnavailable_ShouldReturnUnavailableError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	kc.On("ExchangePassword", mock.Anything, testMobileClientID, "", mock.Anything, mock.Anything).
		Return(domain.TokenPair{}, domain.ErrKeycloakUnavailable)

	uc := newLoginUseCase(t, kc)
	_, err := uc.Execute(context.Background(), application.LoginInput{
		Username:   "user",
		Password:   "pass",
		ClientType: domain.ClientTypeMobile,
	})

	assert.True(t, errors.Is(err, domain.ErrKeycloakUnavailable))
	kc.AssertExpectations(t)
}

// --- Table-driven: all client types that must NOT use ROPC ---

func TestLoginUseCase_Execute_NonROPCClientTypes_ShouldBeRejected(t *testing.T) {
	cases := []domain.ClientType{
		domain.ClientTypeWeb,
		domain.ClientType("tablet"),
		domain.ClientType(""),
	}

	for _, ct := range cases {
		t.Run(string(ct), func(t *testing.T) {
			kc := &mocks.KeycloakPort{}
			uc := newLoginUseCase(t, kc)
			_, err := uc.Execute(context.Background(), application.LoginInput{
				Username:   "user",
				Password:   "pass",
				ClientType: ct,
			})
			assert.Error(t, err)
			kc.AssertNotCalled(t, "ExchangePassword")
		})
	}
}
