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

func newRefreshUseCase(t *testing.T, kc *mocks.KeycloakPort) *application.RefreshUseCase {
	return application.NewRefreshUseCase(
		kc,
		testWebClientID, testWebClientSecret,
		testMobileClientID,
		testAppClientID,
		zaptest.NewLogger(t),
	)
}

// Table-driven test: each client type selects the right Keycloak client_id.
func TestRefreshUseCase_Execute_ShouldUseCorrectClientIDPerClientType(t *testing.T) {
	expectedTokens := domain.TokenPair{AccessToken: "new-at", RefreshToken: "new-rt", ExpiresIn: 300}

	cases := []struct {
		clientType       domain.ClientType
		expectedClientID string
		expectedSecret   string
	}{
		{domain.ClientTypeWeb, testWebClientID, testWebClientSecret},
		{domain.ClientTypeMobile, testMobileClientID, ""},
		{domain.ClientTypeApp, testAppClientID, ""},
	}

	for _, tc := range cases {
		t.Run(string(tc.clientType), func(t *testing.T) {
			kc := &mocks.KeycloakPort{}
			kc.On("RefreshToken", mock.Anything, tc.expectedClientID, tc.expectedSecret, "rt-old").
				Return(expectedTokens, nil)

			uc := newRefreshUseCase(t, kc)
			out, err := uc.Execute(context.Background(), application.RefreshInput{
				RefreshToken: "rt-old",
				ClientType:   tc.clientType,
			})

			assert.NoError(t, err)
			assert.Equal(t, expectedTokens, out.Tokens)
			kc.AssertExpectations(t)
		})
	}
}

func TestRefreshUseCase_Execute_WithInvalidClientType_ShouldReturnError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	uc := newRefreshUseCase(t, kc)

	_, err := uc.Execute(context.Background(), application.RefreshInput{
		RefreshToken: "rt",
		ClientType:   domain.ClientType("unknown"),
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidClientType))
	kc.AssertNotCalled(t, "RefreshToken")
}

func TestRefreshUseCase_Execute_WhenKeycloakReturnsInvalidGrant_ShouldReturnError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	kc.On("RefreshToken", mock.Anything, testMobileClientID, "", "expired-rt").
		Return(domain.TokenPair{}, domain.ErrInvalidCredentials)

	uc := newRefreshUseCase(t, kc)
	_, err := uc.Execute(context.Background(), application.RefreshInput{
		RefreshToken: "expired-rt",
		ClientType:   domain.ClientTypeMobile,
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
	kc.AssertExpectations(t)
}

func TestRefreshUseCase_Execute_WhenKeycloakUnavailable_ShouldReturnUnavailableError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	kc.On("RefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(domain.TokenPair{}, domain.ErrKeycloakUnavailable)

	uc := newRefreshUseCase(t, kc)
	_, err := uc.Execute(context.Background(), application.RefreshInput{
		RefreshToken: "rt",
		ClientType:   domain.ClientTypeApp,
	})

	assert.True(t, errors.Is(err, domain.ErrKeycloakUnavailable))
}
