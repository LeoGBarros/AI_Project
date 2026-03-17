package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Tracing extracts W3C trace context from incoming headers and starts a new span
// for each HTTP request, following TECHNICAL_BASE section 6.5.
func Tracing(next http.Handler) http.Handler {
	tracer := otel.Tracer("auth-service")
	propagator := otel.GetTextMapPropagator()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path)
		defer span.End()

		span.SetAttributes(
			semconv.HTTPMethod(r.Method),
			semconv.HTTPRoute(r.URL.Path),
			attribute.String("correlation_id", GetCorrelationID(ctx)),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
