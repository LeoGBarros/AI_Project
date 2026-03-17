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

func newLogoutUseCase(t *testing.T, kc *mocks.KeycloakPort) *application.LogoutUseCase {
	return application.NewLogoutUseCase(
		kc,
		testWebClientID, testWebClientSecret,
		testMobileClientID,
		testAppClientID,
		zaptest.NewLogger(t),
	)
}

// Table-driven: each client type must revoke with its own client_id.
func TestLogoutUseCase_Execute_ShouldUseCorrectClientIDPerClientType(t *testing.T) {
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
			kc.On("RevokeToken", mock.Anything, tc.expectedClientID, tc.expectedSecret, "rt-123", "refresh_token").
				Return(nil)

			uc := newLogoutUseCase(t, kc)
			err := uc.Execute(context.Background(), application.LogoutInput{
				RefreshToken: "rt-123",
				ClientType:   tc.clientType,
			})

			assert.NoError(t, err)
			kc.AssertExpectations(t)
		})
	}
}

func TestLogoutUseCase_Execute_WithInvalidClientType_ShouldReturnError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	uc := newLogoutUseCase(t, kc)

	err := uc.Execute(context.Background(), application.LogoutInput{
		RefreshToken: "rt",
		ClientType:   domain.ClientType("unknown"),
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidClientType))
	kc.AssertNotCalled(t, "RevokeToken")
}

func TestLogoutUseCase_Execute_WithEmptyClientType_ShouldReturnError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	uc := newLogoutUseCase(t, kc)

	err := uc.Execute(context.Background(), application.LogoutInput{
		RefreshToken: "rt",
		ClientType:   domain.ClientType(""),
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidClientType))
}

func TestLogoutUseCase_Execute_WhenRevocationFails_ShouldReturnError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	revokeErr := errors.New("network timeout")
	kc.On("RevokeToken", mock.Anything, testMobileClientID, "", "rt", "refresh_token").
		Return(revokeErr)

	uc := newLogoutUseCase(t, kc)
	err := uc.Execute(context.Background(), application.LogoutInput{
		RefreshToken: "rt",
		ClientType:   domain.ClientTypeMobile,
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, revokeErr))
}

func TestLogoutUseCase_Execute_ShouldAlwaysPassRefreshTokenHintToKeycloak(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	kc.On("RevokeToken", mock.Anything, mock.Anything, mock.Anything, "my-refresh-token", "refresh_token").
		Return(nil)

	uc := newLogoutUseCase(t, kc)
	err := uc.Execute(context.Background(), application.LogoutInput{
		RefreshToken: "my-refresh-token",
		ClientType:   domain.ClientTypeWeb,
	})

	assert.NoError(t, err)
	kc.AssertExpectations(t)
}
