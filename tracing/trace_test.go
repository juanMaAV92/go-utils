package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func Test_GetTraceIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "Contexto sin trace ID",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "Contexto con trace ID",
			ctx:      context.WithValue(context.Background(), TraceIDKey{}, "test-trace-id"),
			expected: "test-trace-id",
		},
		{
			name:     "Contexto con valor incorrecto",
			ctx:      context.WithValue(context.Background(), TraceIDKey{}, 123),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTraceIDFromContext(tt.ctx)
			if got != tt.expected {
				t.Errorf("GetTraceIDFromContext() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func Test_GetTraceID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(c echo.Context)
		expected string
	}{
		{
			name: "Contexto sin trace ID",
			setup: func(c echo.Context) {
				// No se configura ningún trace ID
			},
			expected: "",
		},
		{
			name: "Contexto con trace ID",
			setup: func(c echo.Context) {
				c.Set(ContextTraceIDKey, "test-trace-id")
			},
			expected: "test-trace-id",
		},
		{
			name: "Contexto con valor incorrecto",
			setup: func(c echo.Context) {
				c.Set(ContextTraceIDKey, 123)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear un nuevo contexto de Echo
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Configurar el contexto según el caso de prueba
			tt.setup(c)

			got := GetTraceIDFromEchoContext(c)
			if got != tt.expected {
				t.Errorf("GetTraceID() = %v, want %v", got, tt.expected)
			}
		})
	}
}
