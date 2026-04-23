package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// strPtr is a helper to create a *string from a string literal.
func strPtr(s string) *string { return &s }

// TestValidate_AllFieldsValid_ShouldReturnEmptyMap verifies that when all
// optional fields are provided with valid values, Validate returns no errors.
func TestValidate_AllFieldsValid_ShouldReturnEmptyMap(t *testing.T) {
	input := UpdateProfileInput{
		DisplayName: strPtr("John Doe"),
		Email:       strPtr("john@example.com"),
		Phone:       strPtr("+5511999998888"),
		AvatarURL:   strPtr("https://cdn.example.com/avatar.png"),
	}

	errs := input.Validate()

	assert.Empty(t, errs)
}

// TestValidate_AllFieldsNil_ShouldReturnEmptyMap verifies that when no fields
// are provided (all nil), Validate returns no errors.
func TestValidate_AllFieldsNil_ShouldReturnEmptyMap(t *testing.T) {
	input := UpdateProfileInput{}

	errs := input.Validate()

	assert.Empty(t, errs)
}

// TestValidate_EmptyDisplayName_ShouldReturnError verifies that an empty
// display_name string produces a validation error.
func TestValidate_EmptyDisplayName_ShouldReturnError(t *testing.T) {
	input := UpdateProfileInput{
		DisplayName: strPtr(""),
	}

	errs := input.Validate()

	assert.Contains(t, errs, "display_name")
	assert.Equal(t, "cannot be empty", errs["display_name"])
}

// TestValidate_InvalidEmail_ShouldReturnError verifies that an email with
// bad format produces a validation error.
func TestValidate_InvalidEmail_ShouldReturnError(t *testing.T) {
	input := UpdateProfileInput{
		Email: strPtr("not-an-email"),
	}

	errs := input.Validate()

	assert.Contains(t, errs, "email")
	assert.Equal(t, "invalid email format", errs["email"])
}

// TestValidate_InvalidPhone_ShouldReturnError verifies that a phone number
// not in E.164 format produces a validation error.
func TestValidate_InvalidPhone_ShouldReturnError(t *testing.T) {
	input := UpdateProfileInput{
		Phone: strPtr("12345"),
	}

	errs := input.Validate()

	assert.Contains(t, errs, "phone")
	assert.Equal(t, "invalid phone format", errs["phone"])
}

// TestValidate_InvalidURL_ShouldReturnError verifies that an avatar_url
// not using http/https produces a validation error.
func TestValidate_InvalidURL_ShouldReturnError(t *testing.T) {
	input := UpdateProfileInput{
		AvatarURL: strPtr("ftp://files.example.com/avatar.png"),
	}

	errs := input.Validate()

	assert.Contains(t, errs, "avatar_url")
	assert.Equal(t, "invalid URL format", errs["avatar_url"])
}

// TestValidate_MultipleInvalidFields_ShouldReturnAllErrors verifies that when
// multiple fields are invalid, all of them appear in the error map.
func TestValidate_MultipleInvalidFields_ShouldReturnAllErrors(t *testing.T) {
	input := UpdateProfileInput{
		DisplayName: strPtr(""),
		Email:       strPtr("bad"),
		Phone:       strPtr("abc"),
		AvatarURL:   strPtr("not-a-url"),
	}

	errs := input.Validate()

	assert.Len(t, errs, 4)
	assert.Contains(t, errs, "display_name")
	assert.Contains(t, errs, "email")
	assert.Contains(t, errs, "phone")
	assert.Contains(t, errs, "avatar_url")
}

// TestValidate_ValidEmailFormats_ShouldAccept uses table-driven tests to verify
// that various valid email formats are accepted by Validate.
func TestValidate_ValidEmailFormats_ShouldAccept(t *testing.T) {
	cases := []struct {
		name  string
		email string
	}{
		{"simple", "user@example.com"},
		{"with dot in local", "first.last@example.com"},
		{"with plus tag", "user+tag@example.com"},
		{"with subdomain", "user@mail.example.com"},
		{"numeric local", "123@example.com"},
		{"short domain", "u@ex.co"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := UpdateProfileInput{Email: strPtr(tc.email)}
			errs := input.Validate()
			assert.Empty(t, errs, "email %q should be valid", tc.email)
		})
	}
}

// TestValidate_InvalidEmailFormats_ShouldReject uses table-driven tests to verify
// that various invalid email formats are rejected by Validate.
func TestValidate_InvalidEmailFormats_ShouldReject(t *testing.T) {
	cases := []struct {
		name  string
		email string
	}{
		{"empty string", ""},
		{"no at sign", "userexample.com"},
		{"no domain", "user@"},
		{"no local part", "@example.com"},
		{"spaces", "user @example.com"},
		{"double at", "user@@example.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := UpdateProfileInput{Email: strPtr(tc.email)}
			errs := input.Validate()
			assert.Contains(t, errs, "email", "email %q should be invalid", tc.email)
		})
	}
}

// TestValidate_ValidPhoneFormats_ShouldAccept uses table-driven tests to verify
// that valid E.164 phone numbers are accepted by Validate.
func TestValidate_ValidPhoneFormats_ShouldAccept(t *testing.T) {
	cases := []struct {
		name  string
		phone string
	}{
		{"brazil mobile", "+5511999998888"},
		{"us number", "+14155552671"},
		{"uk number", "+442071234567"},
		{"min digits", "+1234567"},
		{"max digits", "+123456789012345"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := UpdateProfileInput{Phone: strPtr(tc.phone)}
			errs := input.Validate()
			assert.Empty(t, errs, "phone %q should be valid", tc.phone)
		})
	}
}

// TestValidate_InvalidPhoneFormats_ShouldReject uses table-driven tests to verify
// that invalid phone formats are rejected by Validate.
func TestValidate_InvalidPhoneFormats_ShouldReject(t *testing.T) {
	cases := []struct {
		name  string
		phone string
	}{
		{"empty string", ""},
		{"no plus prefix", "5511999998888"},
		{"starts with zero", "+0123456789"},
		{"too few digits", "+12345"},
		{"letters", "+55abc"},
		{"spaces", "+55 11 99999"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := UpdateProfileInput{Phone: strPtr(tc.phone)}
			errs := input.Validate()
			assert.Contains(t, errs, "phone", "phone %q should be invalid", tc.phone)
		})
	}
}

// TestValidate_ValidURLFormats_ShouldAccept uses table-driven tests to verify
// that http and https URLs are accepted by Validate.
func TestValidate_ValidURLFormats_ShouldAccept(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"https with path", "https://cdn.example.com/avatar.png"},
		{"http plain", "http://example.com/img.jpg"},
		{"https root", "https://example.com"},
		{"https with port", "https://example.com:8080/path"},
		{"http with query", "http://example.com/path?v=1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := UpdateProfileInput{AvatarURL: strPtr(tc.url)}
			errs := input.Validate()
			assert.Empty(t, errs, "url %q should be valid", tc.url)
		})
	}
}

// TestValidate_InvalidURLFormats_ShouldReject uses table-driven tests to verify
// that ftp, empty, and malformed URLs are rejected by Validate.
func TestValidate_InvalidURLFormats_ShouldReject(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"empty string", ""},
		{"ftp scheme", "ftp://files.example.com/avatar.png"},
		{"no scheme", "example.com/avatar.png"},
		{"missing scheme separator", "://missing.com"},
		{"just text", "not-a-url"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := UpdateProfileInput{AvatarURL: strPtr(tc.url)}
			errs := input.Validate()
			assert.Contains(t, errs, "avatar_url", "url %q should be invalid", tc.url)
		})
	}
}
