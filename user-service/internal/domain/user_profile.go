// Package domain contains the core business entities and validation rules
// for the user-service. It has zero external dependencies — only Go standard
// library packages are imported.
package domain

import (
	"net/mail"
	"net/url"
	"regexp"
	"time"
)

// UserProfile represents the profile of an authenticated user.
type UserProfile struct {
	ID          string
	DisplayName string
	Email       string
	Phone       string
	AvatarURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpdateProfileInput contains the fields allowed for profile updates.
// Pointer fields indicate optional values — nil means "do not change".
type UpdateProfileInput struct {
	DisplayName *string
	Email       *string
	Phone       *string
	AvatarURL   *string
}

// Validate checks business rules for the update fields.
// It returns a map of field name → error message for invalid fields.
// An empty map means all provided fields are valid.
func (u UpdateProfileInput) Validate() map[string]string {
	details := make(map[string]string)
	if u.DisplayName != nil && *u.DisplayName == "" {
		details["display_name"] = "cannot be empty"
	}
	if u.Email != nil && !isValidEmail(*u.Email) {
		details["email"] = "invalid email format"
	}
	if u.Phone != nil && !isValidPhone(*u.Phone) {
		details["phone"] = "invalid phone format"
	}
	if u.AvatarURL != nil && !isValidURL(*u.AvatarURL) {
		details["avatar_url"] = "invalid URL format"
	}
	return details
}

// phoneRegex matches international phone numbers in E.164 format:
// a leading '+', followed by a country code digit (1-9), then 6 to 14 digits.
var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

// isValidEmail reports whether s is a valid email address.
// It uses net/mail.ParseAddress from the standard library.
func isValidEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

// isValidPhone reports whether s is a valid phone number in E.164 format
// (e.g., +5511999998888). It must start with '+', followed by a country code
// digit 1-9, and then 6 to 14 additional digits.
func isValidPhone(s string) bool {
	return phoneRegex.MatchString(s)
}

// isValidURL reports whether s is a valid HTTP or HTTPS URL.
func isValidURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
