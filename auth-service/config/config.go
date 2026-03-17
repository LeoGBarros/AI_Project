package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration values for the auth-service.
// All values are read from environment variables with the AUTH_SERVICE_ prefix,
// following the naming convention in TECHNICAL_BASE section 9.5.
type Config struct {
	ServiceName string
	Port        string
	LogLevel    string

	// Keycloak settings
	KeycloakBaseURL       string
	KeycloakRealm         string
	KeycloakClientIDWeb   string
	KeycloakClientSecretWeb string
	KeycloakClientIDMobile string
	KeycloakClientIDApp    string

	// Redis
	RedisAddr string

	// OpenTelemetry
	OTLPEndpoint string
}

// Load reads configuration from environment variables and validates required fields.
// The service fails fast on startup if any required variable is missing.
func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetDefault("AUTH_SERVICE_PORT", "8080")
	viper.SetDefault("AUTH_SERVICE_LOG_LEVEL", "info")

	cfg := &Config{
		ServiceName:             "auth-service",
		Port:                    viper.GetString("AUTH_SERVICE_PORT"),
		LogLevel:                viper.GetString("AUTH_SERVICE_LOG_LEVEL"),
		KeycloakBaseURL:         viper.GetString("AUTH_SERVICE_KEYCLOAK_BASE_URL"),
		KeycloakRealm:           viper.GetString("AUTH_SERVICE_KEYCLOAK_REALM"),
		KeycloakClientIDWeb:     viper.GetString("AUTH_SERVICE_KEYCLOAK_CLIENT_ID_WEB"),
		KeycloakClientSecretWeb: viper.GetString("AUTH_SERVICE_KEYCLOAK_CLIENT_SECRET_WEB"),
		KeycloakClientIDMobile:  viper.GetString("AUTH_SERVICE_KEYCLOAK_CLIENT_ID_MOBILE"),
		KeycloakClientIDApp:     viper.GetString("AUTH_SERVICE_KEYCLOAK_CLIENT_ID_APP"),
		RedisAddr:               viper.GetString("AUTH_SERVICE_REDIS_ADDR"),
		OTLPEndpoint:            viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"AUTH_SERVICE_KEYCLOAK_BASE_URL":         c.KeycloakBaseURL,
		"AUTH_SERVICE_KEYCLOAK_REALM":            c.KeycloakRealm,
		"AUTH_SERVICE_KEYCLOAK_CLIENT_ID_WEB":    c.KeycloakClientIDWeb,
		"AUTH_SERVICE_KEYCLOAK_CLIENT_SECRET_WEB": c.KeycloakClientSecretWeb,
		"AUTH_SERVICE_KEYCLOAK_CLIENT_ID_MOBILE": c.KeycloakClientIDMobile,
		"AUTH_SERVICE_KEYCLOAK_CLIENT_ID_APP":    c.KeycloakClientIDApp,
		"AUTH_SERVICE_REDIS_ADDR":                c.RedisAddr,
	}

	for name, val := range required {
		if val == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	return nil
}
