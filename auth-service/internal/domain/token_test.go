package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/project/auth-service/internal/domain"
)

// --- ClientType ---

func TestClientType_IsValid_WithKnownTypes_ShouldReturnTrue(t *testing.T) {
	cases := []struct {
		name       string
		clientType domain.ClientType
	}{
		{"web", domain.ClientTypeWeb},
		{"mobile", domain.ClientTypeMobile},
		{"app", domain.ClientTypeApp},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, tc.clientType.IsValid())
		})
	}
}

func TestClientType_IsValid_WithUnknownType_ShouldReturnFalse(t *testing.T) {
	unknown := domain.ClientType("desktop")
	assert.False(t, unknown.IsValid())
}

func TestClientType_IsValid_WithEmptyString_ShouldReturnFalse(t *testing.T) {
	empty := domain.ClientType("")
	assert.False(t, empty.IsValid())
}

func TestClientType_IsROPC_WithROPCTypes_ShouldReturnTrue(t *testing.T) {
	cases := []domain.ClientType{
		domain.ClientTypeMobile,
		domain.ClientTypeApp,
	}
	for _, ct := range cases {
		t.Run(string(ct), func(t *testing.T) {
			assert.True(t, ct.IsROPC())
		})
	}
}

func TestClientType_IsROPC_WithWebType_ShouldReturnFalse(t *testing.T) {
	assert.False(t, domain.ClientTypeWeb.IsROPC())
}

// --- PKCEState ---

func TestPKCEState_IsExpired_WithFutureExpiry_ShouldReturnFalse(t *testing.T) {
	state := domain.PKCEState{
		StateID:   "abc",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	assert.False(t, state.IsExpired())
}

func TestPKCEState_IsExpired_WithPastExpiry_ShouldReturnTrue(t *testing.T) {
	state := domain.PKCEState{
		StateID:   "abc",
		ExpiresAt: time.Now().Add(-1 * time.Second),
	}
	assert.True(t, state.IsExpired())
}

func TestPKCEState_IsExpired_WithZeroValue_ShouldReturnTrue(t *testing.T) {
	state := domain.PKCEState{}
	assert.True(t, state.IsExpired())
}
