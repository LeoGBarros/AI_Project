package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration values for the user-service.
// All values are read from environment variables with the USER_SERVICE_ prefix,
// following the naming convention in TECHNICAL_BASE section 9.5.
type Config struct {
	ServiceName string
	Port        string
	LogLevel    string

	// Database
	DBURL string

	// Redis
	RedisAddr string

	// OpenTelemetry
	OTLPEndpoint string
}

// Load reads configuration from environment variables and validates required fields.
// The service fails fast on startup if any required variable is missing.
func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetDefault("USER_SERVICE_PORT", "8080")
	viper.SetDefault("USER_SERVICE_LOG_LEVEL", "info")

	cfg := &Config{
		ServiceName:  "user-service",
		Port:         viper.GetString("USER_SERVICE_PORT"),
		LogLevel:     viper.GetString("USER_SERVICE_LOG_LEVEL"),
		DBURL:        viper.GetString("USER_SERVICE_DB_URL"),
		RedisAddr:    viper.GetString("USER_SERVICE_REDIS_ADDR"),
		OTLPEndpoint: viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"USER_SERVICE_DB_URL":     c.DBURL,
		"USER_SERVICE_REDIS_ADDR": c.RedisAddr,
	}

	for name, val := range required {
		if val == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	return nil
}
