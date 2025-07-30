package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/juanMaAV92/go-utils/tracing"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TraceId() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			ctxWithOtel := extractOtelContext(req)
			c.SetRequest(req.WithContext(ctxWithOtel))
			traceID := extractTraceIDFromContext(ctxWithOtel)

			if traceID == "" {
				traceID = extractTraceIDFromHeader(req.Header)
			}
			if traceID == "" {
				traceID = generateUUID()
			}

			propagateTraceID(c, traceID)

			return next(c)
		}
	}
}

func extractOtelContext(req *http.Request) context.Context {
	return otel.GetTextMapPropagator().Extract(
		req.Context(),
		propagation.HeaderCarrier(req.Header),
	)
}

func extractTraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func extractTraceIDFromHeader(header http.Header) string {
	return header.Get("X-Trace-Id")
}

func propagateTraceID(c echo.Context, traceID string) {
	c.Set(tracing.ContextTraceIDKey, traceID)
	c.Response().Header().Set("X-Trace-Id", traceID)
}

func generateUUID() string {
	return uuid.New().String()
}
