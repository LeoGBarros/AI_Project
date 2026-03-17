package domain

import "errors"

var (
	// ErrInvalidCredentials is returned when the username or password is wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrTokenExpired is returned when an access or refresh token has expired.
	ErrTokenExpired = errors.New("token expired")

	// ErrInvalidToken is returned when a token cannot be parsed or its signature is invalid.
	ErrInvalidToken = errors.New("invalid token")

	// ErrInvalidState is returned when the OAuth state parameter does not match what was stored.
	ErrInvalidState = errors.New("invalid or expired oauth state")

	// ErrInvalidClientType is returned when an unsupported client_type value is provided.
	ErrInvalidClientType = errors.New("invalid client type")

	// ErrKeycloakUnavailable is returned when the Keycloak server cannot be reached.
	ErrKeycloakUnavailable = errors.New("keycloak is unavailable")
)
