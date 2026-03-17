package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/project/auth-service/internal/application"
	"github.com/project/auth-service/internal/domain"
	"github.com/project/auth-service/pkg/apierror"
)

// Handler implements ports/input.AuthHandler using chi and the application use cases.
type Handler struct {
	login     *application.LoginUseCase
	authorize *application.AuthorizeUseCase
	callback  *application.CallbackUseCase
	refresh   *application.RefreshUseCase
	logout    *application.LogoutUseCase

	webClientSecret string
	logger          *zap.Logger
}

// NewHandler constructs the HTTP handler with all required use cases.
func NewHandler(
	login *application.LoginUseCase,
	authorize *application.AuthorizeUseCase,
	callback *application.CallbackUseCase,
	refresh *application.RefreshUseCase,
	logout *application.LogoutUseCase,
	webClientSecret string,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		login:           login,
		authorize:       authorize,
		callback:        callback,
		refresh:         refresh,
		logout:          logout,
		webClientSecret: webClientSecret,
		logger:          logger,
	}
}

// Routes registers all auth-service routes under the provided chi router.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Get("/authorize", h.Authorize)
	r.Get("/callback", h.Callback)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	return r
}

// --- Request / Response types ---

type loginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ClientType string `json:"client_type"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	ClientType   string `json:"client_type"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	ClientType   string `json:"client_type"`
}

// --- Handlers ---

// Login handles POST /v1/auth/login (ROPC for mobile and app clients).
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	traceID := traceIDFromContext(r)

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON", traceID)
		return
	}

	if details := validateLoginRequest(req); len(details) > 0 {
		apierror.WriteValidationError(w, traceID, details)
		return
	}

	ct := domain.ClientType(req.ClientType)
	out, err := h.login.Execute(r.Context(), application.LoginInput{
		Username:   req.Username,
		Password:   req.Password,
		ClientType: ct,
	})
	if err != nil {
		h.handleAuthError(w, err, traceID)
		return
	}

	writeJSON(w, http.StatusOK, out.Tokens)
}

// Authorize handles GET /v1/auth/authorize and starts the PKCE flow for web clients.
func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	traceID := traceIDFromContext(r)

	redirectURI := r.URL.Query().Get("redirect_uri")
	clientID := r.URL.Query().Get("client_id")

	if redirectURI == "" || clientID == "" {
		apierror.WriteValidationError(w, traceID, map[string]string{
			"redirect_uri": "required",
			"client_id":    "required",
		})
		return
	}

	out, err := h.authorize.Execute(r.Context(), application.AuthorizeInput{
		RedirectURI: redirectURI,
		ClientID:    clientID,
	})
	if err != nil {
		h.logger.Error("authorize failed", zap.Error(err))
		apierror.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to initiate authorization", traceID)
		return
	}

	http.Redirect(w, r, out.AuthorizationURL, http.StatusFound)
}

// Callback handles GET /v1/auth/callback — the Keycloak redirect after PKCE authorization.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	traceID := traceIDFromContext(r)

	code := r.URL.Query().Get("code")
	stateID := r.URL.Query().Get("state")

	if code == "" || stateID == "" {
		apierror.WriteError(w, http.StatusBadRequest, "MISSING_PARAMS", "code and state are required", traceID)
		return
	}

	out, err := h.callback.Execute(r.Context(), application.CallbackInput{
		Code:         code,
		StateID:      stateID,
		ClientSecret: h.webClientSecret,
	})
	if err != nil {
		h.handleAuthError(w, err, traceID)
		return
	}

	writeJSON(w, http.StatusOK, out.Tokens)
}

// Refresh handles POST /v1/auth/refresh.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	traceID := traceIDFromContext(r)

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON", traceID)
		return
	}

	if req.RefreshToken == "" {
		apierror.WriteValidationError(w, traceID, map[string]string{
			"refresh_token": "required",
		})
		return
	}
	if req.ClientType == "" {
		apierror.WriteValidationError(w, traceID, map[string]string{
			"client_type": "required",
		})
		return
	}

	out, err := h.refresh.Execute(r.Context(), application.RefreshInput{
		RefreshToken: req.RefreshToken,
		ClientType:   domain.ClientType(req.ClientType),
	})
	if err != nil {
		h.handleAuthError(w, err, traceID)
		return
	}

	writeJSON(w, http.StatusOK, out.Tokens)
}

// Logout handles POST /v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	traceID := traceIDFromContext(r)

	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON", traceID)
		return
	}

	if req.RefreshToken == "" {
		apierror.WriteValidationError(w, traceID, map[string]string{
			"refresh_token": "required",
		})
		return
	}

	if err := h.logout.Execute(r.Context(), application.LogoutInput{
		RefreshToken: req.RefreshToken,
		ClientType:   domain.ClientType(req.ClientType),
	}); err != nil {
		h.handleAuthError(w, err, traceID)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Helpers ---

func (h *Handler) handleAuthError(w http.ResponseWriter, err error, traceID string) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		apierror.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password", traceID)
	case errors.Is(err, domain.ErrTokenExpired):
		apierror.WriteError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "token has expired", traceID)
	case errors.Is(err, domain.ErrInvalidToken):
		apierror.WriteError(w, http.StatusUnauthorized, "INVALID_TOKEN", "token is invalid", traceID)
	case errors.Is(err, domain.ErrInvalidState):
		apierror.WriteError(w, http.StatusBadRequest, "INVALID_STATE", "oauth state is invalid or expired", traceID)
	case errors.Is(err, domain.ErrInvalidClientType):
		apierror.WriteError(w, http.StatusBadRequest, "INVALID_CLIENT_TYPE", "client_type must be web, mobile, or app", traceID)
	case errors.Is(err, domain.ErrKeycloakUnavailable):
		apierror.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "authentication service is temporarily unavailable", traceID)
	default:
		h.logger.Error("unhandled error", zap.Error(err))
		apierror.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred", traceID)
	}
}

func validateLoginRequest(req loginRequest) map[string]string {
	details := make(map[string]string)
	if req.Username == "" {
		details["username"] = "required"
	}
	if req.Password == "" {
		details["password"] = "required"
	}
	if req.ClientType == "" {
		details["client_type"] = "required"
	} else if ct := domain.ClientType(req.ClientType); !ct.IsROPC() {
		details["client_type"] = "must be 'mobile' or 'app' for password login"
	}
	return details
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func traceIDFromContext(r *http.Request) string {
	span := trace.SpanFromContext(r.Context())
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}
