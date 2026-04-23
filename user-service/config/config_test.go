package config

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Feature: user-service, Property 8: Validação de configuração detecta variáveis ausentes
// **Validates: Requirements 8.2**
func TestProperty_ConfigValidationDetectsMissingVars(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		hasDBURL := rapid.Bool().Draw(t, "hasDBURL")
		hasRedisAddr := rapid.Bool().Draw(t, "hasRedisAddr")

		cfg := &Config{
			ServiceName: "user-service",
			Port:        "8080",
			LogLevel:    "info",
		}

		if hasDBURL {
			cfg.DBURL = rapid.StringMatching(`^postgres://[a-z]+:[a-z]+@localhost:\d{4}/[a-z]+$`).Draw(t, "dbURL")
		}
		if hasRedisAddr {
			cfg.RedisAddr = rapid.StringMatching(`^localhost:\d{4,5}$`).Draw(t, "redisAddr")
		}

		err := cfg.validate()

		allSet := hasDBURL && hasRedisAddr
		if allSet {
			if err != nil {
				t.Fatalf("expected no error when all required vars set, got: %v", err)
			}
		} else {
			if err == nil {
				t.Fatal("expected error when required var is missing")
			}
			msg := err.Error()
			mentionsMissing := false
			if !hasDBURL && strings.Contains(msg, "USER_SERVICE_DB_URL") {
				mentionsMissing = true
			}
			if !hasRedisAddr && strings.Contains(msg, "USER_SERVICE_REDIS_ADDR") {
				mentionsMissing = true
			}
			if !mentionsMissing {
				t.Fatalf("error should mention missing variable, got: %s", msg)
			}
		}
	})
}
