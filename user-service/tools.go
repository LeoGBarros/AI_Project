//go:build tools
// +build tools

// Package tools tracks tool and library dependencies for the user-service module.
// This file ensures go mod tidy retains all required dependencies.
package tools

import (
	_ "github.com/go-chi/chi/v5"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/google/uuid"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/redis/go-redis/v9"
	_ "github.com/spf13/viper"
	_ "github.com/stretchr/testify/assert"
	_ "go.opentelemetry.io/otel"
	_ "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	_ "go.opentelemetry.io/otel/sdk"
	_ "go.opentelemetry.io/otel/trace"
	_ "go.uber.org/zap"
	_ "pgregory.net/rapid"
)
