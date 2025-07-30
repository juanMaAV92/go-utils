package tracing

import (
	"context"

	"github.com/labstack/echo/v4"
)

const ContextTraceIDKey = "trace_id"

type TraceIDKey struct{}

func GetTraceIDFromEchoContext(c echo.Context) string {
	traceID, ok := c.Get(ContextTraceIDKey).(string)
	if !ok {
		return ""
	}
	return traceID
}

func GetTraceIDFromContext(c context.Context) string {
	traceID, ok := c.Value(TraceIDKey{}).(string)
	if !ok {
		return ""
	}
	return traceID
}
