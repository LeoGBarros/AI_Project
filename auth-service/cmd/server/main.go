package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"

	"github.com/project/auth-service/config"
	"github.com/project/auth-service/internal/application"
	adapthttp "github.com/project/auth-service/internal/adapters/http"
	"github.com/project/auth-service/internal/adapters/keycloak"
	adaptredis "github.com/project/auth-service/internal/adapters/redis"
	"github.com/project/auth-service/pkg/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	logger, err := buildLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to build logger: %v", err)
	}
	defer logger.Sync() //nolint:errcheck

	// --- OpenTelemetry setup ---
	shutdown, err := initTracer(cfg.OTLPEndpoint)
	if err != nil {
		logger.Warn("tracing initialization failed — continuing without tracing", zap.Error(err))
	} else {
		defer shutdown()
	}

	// --- Adapters ---
	keycloakClient := keycloak.NewClient(cfg.KeycloakBaseURL, cfg.KeycloakRealm, logger)

	redisClient := goredis.NewClient(&goredis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}

	stateStore := adaptredis.NewPKCEStateStore(redisClient, logger)

	// --- Use cases (wiring) ---
	loginUC := application.NewLoginUseCase(
		keycloakClient,
		cfg.KeycloakClientIDWeb, cfg.KeycloakClientSecretWeb,
		cfg.KeycloakClientIDMobile,
		cfg.KeycloakClientIDApp,
		logger,
	)
	authorizeUC := application.NewAuthorizeUseCase(keycloakClient, stateStore, logger)
	callbackUC := application.NewCallbackUseCase(keycloakClient, stateStore, logger)
	refreshUC := application.NewRefreshUseCase(
		keycloakClient,
		cfg.KeycloakClientIDWeb, cfg.KeycloakClientSecretWeb,
		cfg.KeycloakClientIDMobile,
		cfg.KeycloakClientIDApp,
		logger,
	)
	logoutUC := application.NewLogoutUseCase(
		keycloakClient,
		cfg.KeycloakClientIDWeb, cfg.KeycloakClientSecretWeb,
		cfg.KeycloakClientIDMobile,
		cfg.KeycloakClientIDApp,
		logger,
	)

	// --- HTTP handler ---
	authHandler := adapthttp.NewHandler(
		loginUC, authorizeUC, callbackUC, refreshUC, logoutUC,
		cfg.KeycloakClientSecretWeb,
		logger,
	)

	// --- Router ---
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.CorrelationID)
	r.Use(middleware.Tracing)
	r.Use(middleware.Logging(logger))

	r.Mount("/v1/auth", authHandler.Routes())

	// Health checks are not routed through Kong and require no auth.
	r.Get("/healthz/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})
	r.Get("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := redisClient.Ping(r.Context()).Err(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"degraded","reason":"redis unavailable"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// --- Graceful shutdown ---
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("auth-service started", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down auth-service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
}

// buildLogger constructs a zap production logger.
func buildLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	return cfg.Build()
}

// initTracer configures the OpenTelemetry global tracer and returns a shutdown function.
func initTracer(otlpEndpoint string) (func(), error) {
	if otlpEndpoint == "" {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		return func() {}, nil
	}

	ctx := context.Background()
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(otlpEndpoint), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}, nil
}
