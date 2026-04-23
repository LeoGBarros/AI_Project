package domain

import "errors"

var (
	// ErrUserNotFound is returned when a user profile does not exist for the given ID.
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidToken is returned when the JWT is missing the required sub claim.
	ErrInvalidToken = errors.New("invalid token: missing sub claim")

	// ErrValidation is returned when input fields fail business rule validation.
	ErrValidation = errors.New("validation error")
)
