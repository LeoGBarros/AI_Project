package apierror

import (
	"encoding/json"
	"net/http"
)

// errorDetail represents a single field-level validation problem.
type errorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// errorBody is the inner object of the standard error envelope.
type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []errorDetail `json:"details,omitempty"`
	TraceID string        `json:"trace_id,omitempty"`
}

// envelope is the outer wrapper matching the TECHNICAL_BASE error format.
type envelope struct {
	Error errorBody `json:"error"`
}

// WriteError writes a structured JSON error response following the project's
// envelope format defined in TECHNICAL_BASE section 5.3.
func WriteError(w http.ResponseWriter, statusCode int, code, message, traceID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(envelope{
		Error: errorBody{
			Code:    code,
			Message: message,
			TraceID: traceID,
		},
	})
}

// WriteValidationError writes a 400 error with field-level details.
func WriteValidationError(w http.ResponseWriter, traceID string, details map[string]string) {
	dd := make([]errorDetail, 0, len(details))
	for field, msg := range details {
		dd = append(dd, errorDetail{Field: field, Message: msg})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(envelope{
		Error: errorBody{
			Code:    "VALIDATION_ERROR",
			Message: "one or more fields are invalid",
			Details: dd,
			TraceID: traceID,
		},
	})
}
