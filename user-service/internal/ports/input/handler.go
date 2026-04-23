package input

import "net/http"

// UserHandler define a superfície HTTP do user-service.
type UserHandler interface {
	// GetProfile trata GET /v1/users/me
	GetProfile(w http.ResponseWriter, r *http.Request)

	// UpdateProfile trata PUT /v1/users/me
	UpdateProfile(w http.ResponseWriter, r *http.Request)
}
