package input

import "net/http"

// AuthHandler defines the HTTP surface of the auth-service.
// Each method corresponds to one route registered on the chi router.
type AuthHandler interface {
	// Login handles ROPC authentication for mobile and app clients.
	// POST /v1/auth/login
	Login(w http.ResponseWriter, r *http.Request)

	// Authorize initiates the Authorization Code + PKCE flow for web clients.
	// GET /v1/auth/authorize
	Authorize(w http.ResponseWriter, r *http.Request)

	// Callback processes the Keycloak redirect after successful PKCE authorization.
	// GET /v1/auth/callback
	Callback(w http.ResponseWriter, r *http.Request)

	// Refresh exchanges a refresh_token for a new token pair.
	// POST /v1/auth/refresh
	Refresh(w http.ResponseWriter, r *http.Request)

	// Logout revokes the user's tokens and ends the session.
	// POST /v1/auth/logout
	Logout(w http.ResponseWriter, r *http.Request)
}
