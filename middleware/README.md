# Middleware Package

This package provides middlewares for Echo applications in Go, focused on logging, tracing, and trace ID propagation.

## Main Files

- `Tracing Middleware`: Middleware for tracing instrumentation with OpenTelemetry. It records the execution time of requests and propagates the tracing context.
- `TraceId Middleware`: Middleware for trace ID propagation. If no trace ID exists in the context or headers, it generates a new one using UUID. The trace ID is propagated in the context and in the response header `X-Trace-Id`.
- `Logging Middleware`: Middleware for logging HTTP requests and responses, including body, duration, status, and tracing context with OpenTelemetry. It allows logging errors and relevant metrics for observability.

## Usage

### Tracing Middleware
```go
import "github.com/juanMaAV92/go-utils/middleware"
echo.Use(middleware.Tracing("service-name"))
```

### TraceId Middleware
```go
import "github.com/juanMaAV92/go-utils/middleware"
echo.Use(middleware.TraceId())
```

### Logging Middleware
```go
import "github.com/juanMaAV92/go-utils/middleware"
echo.Use(middleware.Logging(logger))
```

## Requirements
- Echo v4
- OpenTelemetry
- github.com/google/uuid (for trace ID generation)

## Integration Example
```go
func main() {
    e := echo.New()
    e.Use(middleware.TraceId())
    e.Use(middleware.Logging(myLogger))
    e.Use(middleware.Tracing("my-service"))
    // ...
    e.Start(":8080")
}
```

## License
MIT
