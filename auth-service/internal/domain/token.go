package domain

import "time"

// ClientType identifies the type of client initiating authentication.
type ClientType string

const (
	ClientTypeWeb    ClientType = "web"
	ClientTypeMobile ClientType = "mobile"
	ClientTypeApp    ClientType = "app"
)

// IsValid returns true when the ClientType is one of the accepted values.
func (c ClientType) IsValid() bool {
	switch c {
	case ClientTypeWeb, ClientTypeMobile, ClientTypeApp:
		return true
	}
	return false
}

// IsROPC returns true for client types that use Resource Owner Password Credentials.
func (c ClientType) IsROPC() bool {
	return c == ClientTypeMobile || c == ClientTypeApp
}

// TokenPair holds the token response returned to the client after a successful authentication.
type TokenPair struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

// PKCEState holds the transient state created when starting an Authorization Code + PKCE flow.
// It is persisted in Redis with a short TTL and consumed once on the callback.
type PKCEState struct {
	// StateID is the random opaque value sent as the OAuth `state` parameter (CSRF protection).
	StateID string

	// CodeVerifier is the secret random string whose SHA-256 hash was sent as code_challenge.
	CodeVerifier string

	// RedirectURI is the client-provided URI to redirect to after the Keycloak callback.
	RedirectURI string

	// ClientID is the Keycloak client_id used to initiate this flow.
	ClientID string

	// ExpiresAt is the absolute time after which this state is no longer valid.
	ExpiresAt time.Time
}

// IsExpired returns true when the PKCE state is past its expiration time.
func (p PKCEState) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}
