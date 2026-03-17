package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/project/auth-service/internal/application"
	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/internal/ports/output/mocks"
)

func newAuthorizeUseCase(t *testing.T, kc *mocks.KeycloakPort, ss *mocks.PKCEStateStore) *application.AuthorizeUseCase {
	return application.NewAuthorizeUseCase(kc, ss, zaptest.NewLogger(t))
}

func TestAuthorizeUseCase_Execute_WithValidInput_ShouldReturnAuthorizationURL(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	expectedURL := "https://keycloak.example.com/realms/test/protocol/openid-connect/auth?..."
	kc.On("GetAuthorizationURL",
		"web-client",
		"https://app.example.com/callback",
		mock.AnythingOfType("string"), // stateID — random, validated below
		mock.AnythingOfType("string"), // codeChallenge
		"S256",
	).Return(expectedURL)
	ss.On("Save", mock.Anything, mock.AnythingOfType("domain.PKCEState")).Return(nil)

	uc := newAuthorizeUseCase(t, kc, ss)
	out, err := uc.Execute(context.Background(), application.AuthorizeInput{
		RedirectURI: "https://app.example.com/callback",
		ClientID:    "web-client",
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, out.AuthorizationURL)
	kc.AssertExpectations(t)
	ss.AssertExpectations(t)
}

func TestAuthorizeUseCase_Execute_WithValidInput_ShouldSaveStateWithCorrectClientID(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	kc.On("GetAuthorizationURL", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("https://kc.example.com/auth?foo=bar")

	ss.On("Save", mock.Anything, mock.MatchedBy(func(s interface{}) bool {
		state, ok := s.(domain.PKCEState)
		if !ok {
			return false
		}
		return state.ClientID == "web-client" && state.RedirectURI == "https://app.example.com/callback"
	})).Return(nil)

	uc := newAuthorizeUseCase(t, kc, ss)
	_, err := uc.Execute(context.Background(), application.AuthorizeInput{
		RedirectURI: "https://app.example.com/callback",
		ClientID:    "web-client",
	})

	assert.NoError(t, err)
	ss.AssertExpectations(t)
}

func TestAuthorizeUseCase_Execute_WhenStateStoreFails_ShouldReturnError(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	storeErr := errors.New("redis connection refused")
	ss.On("Save", mock.Anything, mock.AnythingOfType("domain.PKCEState")).Return(storeErr)

	uc := newAuthorizeUseCase(t, kc, ss)
	_, err := uc.Execute(context.Background(), application.AuthorizeInput{
		RedirectURI: "https://app.example.com/callback",
		ClientID:    "web-client",
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, storeErr))
	kc.AssertNotCalled(t, "GetAuthorizationURL")
}

func TestAuthorizeUseCase_Execute_GeneratedStateIDs_ShouldBeUnique(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	kc.On("GetAuthorizationURL", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("https://kc.example.com/auth")
	ss.On("Save", mock.Anything, mock.AnythingOfType("domain.PKCEState")).Return(nil)

	uc := newAuthorizeUseCase(t, kc, ss)

	for i := 0; i < 10; i++ {
		out, err := uc.Execute(context.Background(), application.AuthorizeInput{
			RedirectURI: "https://app.example.com/callback",
			ClientID:    "web-client",
		})
		assert.NoError(t, err)
		_ = out
	}

	assert.Equal(t, 10, len(ss.Calls), "Save should have been called 10 times")

	seenStateIDs := make(map[string]bool)
	for _, call := range ss.Calls {
		state := call.Arguments[1].(domain.PKCEState)
		assert.False(t, seenStateIDs[state.StateID], "state ID %q was repeated", state.StateID)
		seenStateIDs[state.StateID] = true
	}
}

func TestAuthorizeUseCase_Execute_GeneratedCodeChallenge_ShouldBeBase64URL(t *testing.T) {
	kc := &mocks.KeycloakPort{}
	ss := &mocks.PKCEStateStore{}

	var capturedChallenge string
	kc.On("GetAuthorizationURL",
		mock.Anything, mock.Anything, mock.Anything,
		mock.MatchedBy(func(challenge string) bool {
			capturedChallenge = challenge
			return true
		}),
		"S256",
	).Return("https://kc.example.com/auth")
	ss.On("Save", mock.Anything, mock.AnythingOfType("domain.PKCEState")).Return(nil)

	uc := newAuthorizeUseCase(t, kc, ss)
	_, err := uc.Execute(context.Background(), application.AuthorizeInput{
		RedirectURI: "https://app.example.com/callback",
		ClientID:    "web-client",
	})

	assert.NoError(t, err)
	// S256 code_challenge must be non-empty and must not contain padding '='
	// (raw base64url encoding, per RFC 7636).
	assert.NotEmpty(t, capturedChallenge)
	assert.False(t, strings.Contains(capturedChallenge, "="), "code_challenge must use raw base64url (no padding)")
}
