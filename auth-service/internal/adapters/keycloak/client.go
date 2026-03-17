package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/project/auth-service/internal/domain"
)

// circuitState represents the state of the circuit breaker.
type circuitState int

const (
	circuitClosed   circuitState = iota // normal operation
	circuitOpen                         // requests blocked
	circuitHalfOpen                     // probing for recovery
)

const (
	circuitFailureThreshold = 5
	circuitOpenDuration     = 30 * time.Second
	maxRetries              = 3
	retryBaseDelay          = 200 * time.Millisecond
	httpTimeout             = 10 * time.Second
)

// tokenResponse mirrors the JSON body returned by Keycloak token endpoints.
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Client implements output.KeycloakPort using Keycloak's OpenID Connect endpoints.
// It includes a simple circuit breaker and retry logic with exponential backoff.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	realm       string
	logger      *zap.Logger

	// circuit breaker state
	mu           sync.Mutex
	cbState      circuitState
	failureCount int
	openedAt     time.Time
}

// NewClient constructs a Keycloak HTTP adapter.
func NewClient(baseURL, realm string, logger *zap.Logger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: httpTimeout},
		baseURL:    strings.TrimRight(baseURL, "/"),
		realm:      realm,
		logger:     logger,
		cbState:    circuitClosed,
	}
}

// tokenEndpoint returns the Keycloak token endpoint URL for the configured realm.
func (c *Client) tokenEndpoint() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.baseURL, c.realm)
}

// revokeEndpoint returns the Keycloak token revocation endpoint URL.
func (c *Client) revokeEndpoint() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/revoke", c.baseURL, c.realm)
}

// authEndpoint returns the Keycloak authorization endpoint URL.
func (c *Client) authEndpoint() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth", c.baseURL, c.realm)
}

// --- Circuit Breaker ---

func (c *Client) allowRequest() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.cbState {
	case circuitOpen:
		if time.Since(c.openedAt) >= circuitOpenDuration {
			c.cbState = circuitHalfOpen
			return true
		}
		return false
	default:
		return true
	}
}

func (c *Client) recordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cbState = circuitClosed
	c.failureCount = 0
}

func (c *Client) recordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failureCount++
	if c.failureCount >= circuitFailureThreshold || c.cbState == circuitHalfOpen {
		c.cbState = circuitOpen
		c.openedAt = time.Now()
		c.logger.Warn("circuit breaker opened",
			zap.Int("failure_count", c.failureCount),
		)
	}
}

// --- Retry with exponential backoff ---

func (c *Client) postForm(ctx context.Context, endpoint string, form url.Values) (*tokenResponse, error) {
	if !c.allowRequest() {
		return nil, domain.ErrKeycloakUnavailable
	}

	var lastErr error
	delay := retryBaseDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2
		}

		resp, err := c.doPost(ctx, endpoint, form)
		if err != nil {
			lastErr = err
			if isTransient(err) {
				c.recordFailure()
				continue
			}
			c.recordFailure()
			return nil, err
		}

		c.recordSuccess()
		return resp, nil
	}

	return nil, lastErr
}

func (c *Client) doPost(ctx context.Context, endpoint string, form url.Values) (*tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("unexpected keycloak response: %w", err)
	}

	if res.StatusCode >= 500 {
		return nil, fmt.Errorf("keycloak server error: %d", res.StatusCode)
	}

	return &tr, nil
}

func isTransient(err error) bool {
	// Network errors, timeouts, and 5xx are retried.
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	return false
}

// --- KeycloakPort implementation ---

// ExchangePassword implements output.KeycloakPort for the ROPC grant type.
func (c *Client) ExchangePassword(ctx context.Context, clientID, clientSecret, username, password string) (domain.TokenPair, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "keycloak.ExchangePassword")
	defer span.End()
	span.SetAttributes(attribute.String("keycloak.client_id", clientID))

	form := url.Values{
		"grant_type": {"password"},
		"client_id":  {clientID},
		"username":   {username},
		"password":   {password},
		"scope":      {"openid"},
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	tr, err := c.postForm(ctx, c.tokenEndpoint(), form)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return c.mapTokenResponse(tr)
}

// ExchangeCode implements output.KeycloakPort for the Authorization Code + PKCE grant.
func (c *Client) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (domain.TokenPair, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "keycloak.ExchangeCode")
	defer span.End()
	span.SetAttributes(attribute.String("keycloak.client_id", clientID))

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	tr, err := c.postForm(ctx, c.tokenEndpoint(), form)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return c.mapTokenResponse(tr)
}

// GetAuthorizationURL implements output.KeycloakPort.
// It builds the URL the browser must be redirected to in order to start the PKCE flow.
func (c *Client) GetAuthorizationURL(clientID, redirectURI, state, codeChallenge, codeChallengeMethod string) string {
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {codeChallengeMethod},
		"scope":                 {"openid"},
	}
	return c.authEndpoint() + "?" + params.Encode()
}

// RefreshToken implements output.KeycloakPort.
func (c *Client) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (domain.TokenPair, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "keycloak.RefreshToken")
	defer span.End()
	span.SetAttributes(attribute.String("keycloak.client_id", clientID))

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	tr, err := c.postForm(ctx, c.tokenEndpoint(), form)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return c.mapTokenResponse(tr)
}

// RevokeToken implements output.KeycloakPort by calling the Keycloak revocation endpoint.
func (c *Client) RevokeToken(ctx context.Context, clientID, clientSecret, token, tokenTypeHint string) error {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "keycloak.RevokeToken")
	defer span.End()

	form := url.Values{
		"client_id":       {clientID},
		"token":           {token},
		"token_type_hint": {tokenTypeHint},
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	if !c.allowRequest() {
		return domain.ErrKeycloakUnavailable
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.revokeEndpoint(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.recordFailure()
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 500 {
		c.recordFailure()
		return domain.ErrKeycloakUnavailable
	}

	c.recordSuccess()
	return nil
}

// mapTokenResponse translates the raw Keycloak token response to domain.TokenPair.
func (c *Client) mapTokenResponse(tr *tokenResponse) (domain.TokenPair, error) {
	if tr.Error != "" {
		switch tr.Error {
		case "invalid_grant":
			return domain.TokenPair{}, domain.ErrInvalidCredentials
		default:
			return domain.TokenPair{}, fmt.Errorf("keycloak error: %s", tr.Error)
		}
	}

	return domain.TokenPair{
		AccessToken:      tr.AccessToken,
		RefreshToken:     tr.RefreshToken,
		TokenType:        tr.TokenType,
		ExpiresIn:        tr.ExpiresIn,
		RefreshExpiresIn: tr.RefreshExpiresIn,
	}, nil
}
