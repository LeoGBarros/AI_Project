package domain

import (
	"testing"

	"pgregory.net/rapid"
)

// Feature: user-service, Property 3: Validação rejeita entradas inválidas
// **Validates: Requirements 2.1, 2.4**
//
// For any UpdateProfileInput where at least one field contains an invalid value
// (display_name empty, email with invalid format, phone with invalid format,
// avatar_url with invalid format), the Validate function must return a non-empty
// map containing exactly the invalid fields with their respective error messages,
// and no valid field should appear in the error map.
func TestProperty_ValidationRejectsInvalidInputs(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// For each field, decide: nil (not provided), valid, or invalid
		type fieldState int
		const (
			stateNil     fieldState = 0
			stateValid   fieldState = 1
			stateInvalid fieldState = 2
		)

		// Generate states for each field
		dnState := fieldState(rapid.IntRange(0, 2).Draw(t, "displayNameState"))
		emailState := fieldState(rapid.IntRange(0, 2).Draw(t, "emailState"))
		phoneState := fieldState(rapid.IntRange(0, 2).Draw(t, "phoneState"))
		urlState := fieldState(rapid.IntRange(0, 2).Draw(t, "avatarURLState"))

		// Ensure at least one field is invalid
		hasInvalid := dnState == stateInvalid || emailState == stateInvalid ||
			phoneState == stateInvalid || urlState == stateInvalid
		if !hasInvalid {
			switch rapid.IntRange(0, 3).Draw(t, "forceInvalid") {
			case 0:
				dnState = stateInvalid
			case 1:
				emailState = stateInvalid
			case 2:
				phoneState = stateInvalid
			case 3:
				urlState = stateInvalid
			}
		}

		input := UpdateProfileInput{}
		expectedInvalid := make(map[string]bool)

		// DisplayName: empty string is invalid
		switch dnState {
		case stateNil:
			// leave nil — not provided
		case stateValid:
			v := rapid.StringMatching(`^[A-Za-z ]{1,50}$`).Draw(t, "validDisplayName")
			input.DisplayName = &v
		case stateInvalid:
			v := ""
			input.DisplayName = &v
			expectedInvalid["display_name"] = true
		}

		// Email: must be a valid RFC 5322 address
		switch emailState {
		case stateNil:
			// leave nil
		case stateValid:
			v := rapid.StringMatching(`^[a-z]{3,10}@[a-z]{3,10}\.[a-z]{2,4}$`).Draw(t, "validEmail")
			input.Email = &v
		case stateInvalid:
			v := rapid.OneOf(
				rapid.Just(""),
				rapid.Just("notanemail"),
				rapid.Just("missing@"),
				rapid.Just("@nodomain"),
			).Draw(t, "invalidEmail")
			input.Email = &v
			expectedInvalid["email"] = true
		}

		// Phone: must match E.164 format (+[1-9]\d{6,14})
		switch phoneState {
		case stateNil:
			// leave nil
		case stateValid:
			v := rapid.StringMatching(`^\+[1-9]\d{7,13}$`).Draw(t, "validPhone")
			input.Phone = &v
		case stateInvalid:
			v := rapid.OneOf(
				rapid.Just(""),
				rapid.Just("12345"),
				rapid.Just("abc"),
				rapid.Just("+0123456789"),
			).Draw(t, "invalidPhone")
			input.Phone = &v
			expectedInvalid["phone"] = true
		}

		// AvatarURL: must be a valid http or https URL
		switch urlState {
		case stateNil:
			// leave nil
		case stateValid:
			v := rapid.StringMatching(`^https://[a-z]{3,10}\.[a-z]{2,4}/[a-z]{1,10}$`).Draw(t, "validURL")
			input.AvatarURL = &v
		case stateInvalid:
			v := rapid.OneOf(
				rapid.Just(""),
				rapid.Just("not-a-url"),
				rapid.Just("ftp://invalid.com/path"),
				rapid.Just("://missing-scheme"),
			).Draw(t, "invalidURL")
			input.AvatarURL = &v
			expectedInvalid["avatar_url"] = true
		}

		errors := input.Validate()

		// Property: all expected invalid fields must be in the error map
		for field := range expectedInvalid {
			if _, ok := errors[field]; !ok {
				t.Fatalf("expected error for field %q but got none", field)
			}
		}

		// Property: no valid or nil fields should appear in the error map
		for field := range errors {
			if !expectedInvalid[field] {
				t.Fatalf("unexpected error for field %q: %s", field, errors[field])
			}
		}

		// Property: error map must be non-empty (at least one invalid field)
		if len(errors) == 0 {
			t.Fatal("expected non-empty error map for input with invalid fields")
		}
	})
}
