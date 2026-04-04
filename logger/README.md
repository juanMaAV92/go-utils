# logger

Structured logging via `log/slog` (stdlib). Key behaviors:
- Fields are **flat** at the JSON root — queryable directly in Grafana/Loki/Datadog without path notation
- `trace_id` and `span_id` injected automatically from the OTel context when a span is active
- `"local"` environment → colored console output; anything else → JSON to stdout
- `Fatal` calls `os.Exit(1)` after logging — do not use inside request handlers

## Interface

```go
type Logger interface {
    Fatal(ctx context.Context, step, message string, args ...any)
    Error(ctx context.Context, step, message string, args ...any)
    Warning(ctx context.Context, step, message string, args ...any)
    Info(ctx context.Context, step, message string, args ...any)
    Debug(ctx context.Context, step, message string, args ...any)
}
```

## Constructor

```go
func New(serviceName, environment string, opts ...Option) Logger
```

## Options

```go
logger.WithLevel(logger.DebugLevel)   // default: InfoLevel
```

Available levels: `FatalLevel`, `ErrorLevel`, `WarningLevel`, `InfoLevel`, `DebugLevel`

## Usage

```go
import "github.com/juanmaAV/go-utils/logger"

log := logger.New("order-service", "production")
log := logger.New("order-service", "local", logger.WithLevel(logger.DebugLevel))

log.Info(ctx, "order.create", "order created", "orderID", id, "userID", uid)
log.Error(ctx, "payment.charge", "charge failed", "amount", 100, "err", err)
log.Debug(ctx, "cache.hit", "cache hit", "key", cacheKey)
log.Warning(ctx, "rate.limit", "approaching limit", "usage", 0.9)
```

Produces (JSON, production):
```json
{
  "level": "info",
  "service": "order-service",
  "time": "2024-03-15T10:30:00Z",
  "step": "order.create",
  "orderID": "ord_123",
  "userID": "usr_456",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "message": "order created"
}
```

## Conventions

- `step` identifies the operation: `"entity.action"` (e.g. `"user.create"`, `"payment.charge"`)
- `args` are key-value pairs: `"key", value, "key2", value2, ...`
- Non-string keys and unpaired keys are silently skipped
- Avoid using reserved keys as custom fields: `level`, `time`, `message`, `service`, `step`, `trace_id`, `span_id`

## Notes

- `trace_id`/`span_id` are only populated when `telemetry.InitTelemetry` has been called and a span is active in the context
- Dependencies: `log/slog` (stdlib), `go.opentelemetry.io/otel/trace`
