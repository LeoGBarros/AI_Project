package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/project/auth-service/internal/application"
	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/internal/ports/output/mocks"
)

func newCallbackUseCase(t *testing.T, kc *mocks.KeycloakPort, ss *mocks.PKCEStateStore) *application.CallbackUseCase {
	return application.NewCallbackUseCase(kc, ss, zaptest.NewLogger(t))
}

var validPKCEState = domain.PKCEState{
	StateID:      "state-abc",
	CodeVerifier: "verifier-xyz",
	RedirectURI:  "https://app.example.com/callback",
	ClientID:     "web-client",
	ExpiresAt:    time.Now().Add(5 * time.Minute),
}

func TestCallbackUseCase_Execute_WithValidStateAndCode_ShouldReturnTokenPair(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	expectedTokens := domain.TokenPair{AccessToken: "at", RefreshToken: "rt", TokenType: "Bearer", ExpiresIn: 300}

	ss.On("Get", mock.Anything, "state-abc").Return(validPKCEState, nil)
	kc.On("ExchangeCode", mock.Anything,
		"web-client", "web-secret",
		"auth-code-123",
		"https://app.example.com/callback",
		"verifier-xyz",
	).Return(expectedTokens, nil)
	ss.On("Delete", mock.Anything, "state-abc").Return(nil)

	uc := newCallbackUseCase(t, kc, ss)
	out, err := uc.Execute(context.Background(), application.CallbackInput{
		Code:         "auth-code-123",
		StateID:      "state-abc",
		ClientSecret: "web-secret",
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, out.Tokens)
	ss.AssertExpectations(t)
	kc.AssertExpectations(t)
}

func TestCallbackUseCase_Execute_WithUnknownStateID_ShouldReturnInvalidStateError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	ss.On("Get", mock.Anything, "bad-state").Return(domain.PKCEState{}, domain.ErrInvalidState)

	uc := newCallbackUseCase(t, kc, ss)
	_, err := uc.Execute(context.Background(), application.CallbackInput{
		Code:    "code",
		StateID: "bad-state",
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidState))
	kc.AssertNotCalled(t, "ExchangeCode")
	ss.AssertNotCalled(t, "Delete")
}

func TestCallbackUseCase_Execute_WithExpiredState_ShouldReturnInvalidStateError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	expiredState := domain.PKCEState{
		StateID:      "state-expired",
		CodeVerifier: "v",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "web-client",
		ExpiresAt:    time.Now().Add(-1 * time.Minute),
	}

	ss.On("Get", mock.Anything, "state-expired").Return(expiredState, nil)
	ss.On("Delete", mock.Anything, "state-expired").Return(nil)

	uc := newCallbackUseCase(t, kc, ss)
	_, err := uc.Execute(context.Background(), application.CallbackInput{
		Code:    "code",
		StateID: "state-expired",
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidState))
	kc.AssertNotCalled(t, "ExchangeCode")
	ss.AssertCalled(t, "Delete", mock.Anything, "state-expired")
}

func TestCallbackUseCase_Execute_WhenKeycloakCodeExchangeFails_ShouldReturnError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	ss.On("Get", mock.Anything, "state-abc").Return(validPKCEState, nil)
	kc.On("ExchangeCode", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(domain.TokenPair{}, domain.ErrInvalidCredentials)

	uc := newCallbackUseCase(t, kc, ss)
	_, err := uc.Execute(context.Background(), application.CallbackInput{
		Code:    "bad-code",
		StateID: "state-abc",
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
	ss.AssertNotCalled(t, "Delete")
}

func TestCallbackUseCase_Execute_AfterSuccessfulExchange_ShouldDeleteState(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	ss.On("Get", mock.Anything, "state-abc").Return(validPKCEState, nil)
	kc.On("ExchangeCode", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(domain.TokenPair{AccessToken: "at"}, nil)
	ss.On("Delete", mock.Anything, "state-abc").Return(nil)

	uc := newCallbackUseCase(t, kc, ss)
	_, err := uc.Execute(context.Background(), application.CallbackInput{
		Code:    "code",
		StateID: "state-abc",
	})

	assert.NoError(t, err)
	ss.AssertCalled(t, "Delete", mock.Anything, "state-abc")
}

func TestCallbackUseCase_Execute_WhenDeleteFails_ShouldStillReturnTokens(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	expectedTokens := domain.TokenPair{AccessToken: "at"}
	ss.On("Get", mock.Anything, "state-abc").Return(validPKCEState, nil)
	kc.On("ExchangeCode", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedTokens, nil)
	ss.On("Delete", mock.Anything, "state-abc").Return(errors.New("redis timeout"))

	uc := newCallbackUseCase(t, kc, ss)
	out, err := uc.Execute(context.Background(), application.CallbackInput{
		Code:    "code",
		StateID: "state-abc",
	})

	// Delete failure is logged but does not abort the successful token response.
	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, out.Tokens)
}
