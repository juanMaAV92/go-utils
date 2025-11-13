package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	customErrors "github.com/juanMaAV92/go-utils/errors"
	"github.com/juanMaAV92/go-utils/log"
)

// Tracing creates an Echo middleware that adds OpenTelemetry tracing
func Tracing(serviceName string) echo.MiddlewareFunc {
	return otelecho.Middleware(serviceName)
}

// CustomResponseWriter wraps the response writer to capture response body
type CustomResponseWriter struct {
	io.Writer
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)          // Save body for logging
	return w.Writer.Write(b) // Write response to client
}

func (w *CustomResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Logging creates an Echo middleware that logs requests with trace context
func Logging(logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			// Get trace context
			span := trace.SpanFromContext(req.Context())

			// Add HTTP attributes to span
			if span.IsRecording() {
				span.SetAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.url", req.URL.String()),
					attribute.String("http.route", c.Path()),
					attribute.String("http.user_agent", req.UserAgent()),
					attribute.String("http.remote_addr", c.RealIP()),
				)
			}

			// Capture request body
			var requestBody string
			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err == nil {
					requestBody = string(bodyBytes)
					// Restore the body for the handler
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			// Setup custom response writer to capture response body and status code
			resBodyBuffer := new(bytes.Buffer)
			customWriter := &CustomResponseWriter{
				Writer:         io.MultiWriter(res.Writer, resBodyBuffer),
				ResponseWriter: res.Writer,
				statusCode:     http.StatusOK, // Default status code
			}
			res.Writer = customWriter

			// Process request and store error for later logging
			var returnErr error
			func() {
				defer func() {
					// This recovers from panics and captures them
					if r := recover(); r != nil {
						returnErr = echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
						panic(r)
					}
				}()
				returnErr = next(c)
			}()

			// Calculate duration
			duration := time.Since(start)

			// Get response status code - Priority: customWriter > response > error > default to 200
			statusCode := customWriter.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}
			// If response was already written by the error handler, get actual status
			if statusCode == http.StatusOK && returnErr != nil {
				// Try to extract HTTP error code from custom ErrorResponse
				var customErr *customErrors.ErrorResponse
				if errors.As(returnErr, &customErr) {
					statusCode = customErr.ErrorHTTPCode()
				} else if he, ok := returnErr.(*echo.HTTPError); ok {
					statusCode = he.Code
				}
			}

			// Get response body
			responseBody := resBodyBuffer.String()

			// Add response attributes to span
			if span.IsRecording() {
				span.SetAttributes(
					attribute.Int("http.status_code", statusCode),
					attribute.Int64("http.response_size", int64(len(responseBody))),
					attribute.String("http.duration", duration.String()),
				)

				if returnErr != nil {
					span.SetAttributes(attribute.String("error.message", returnErr.Error()))
				}
			}

			// Prepare log data
			logData := map[string]interface{}{
				"method":        req.Method,
				"path":          req.URL.Path,
				"remote_ip":     c.RealIP(),
				"user_agent":    req.UserAgent(),
				"status_code":   statusCode,
				"response_size": len(responseBody),
				"duration":      duration.String(),
			}

			// Add request body if present and not too large
			if requestBody != "" && len(requestBody) < 10000 {
				logData["request_body"] = requestBody
			}

			// Add response body if present and not too large
			if responseBody != "" && len(responseBody) < 10000 {
				logData["response_body"] = responseBody
			}

			// Log based on HTTP status code (2xx = success) rather than handler error
			// A handler may return an error for internal logging/tracing while still
			// responding successfully. The HTTP status code is the source of truth.
			isSuccessful := statusCode >= 200 && statusCode < 300
			if isSuccessful {
				logger.Info(req.Context(), "HTTP request completed", "success request", log.Fields(logData))
			} else {
				errorMsg := "HTTP request failed"
				if returnErr != nil {
					errorMsg = returnErr.Error()
				}
				logger.Error(req.Context(), "HTTP request completed with error", errorMsg, log.Fields(logData))
			}

			return returnErr
		}
	}
}
