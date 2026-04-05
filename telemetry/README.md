# telemetry

OpenTelemetry trace provider initialization. Registers the global tracer and text-map propagator so every package in the service can use `otel.Tracer()` without knowing the exporter details.

## Interface

```go
func InitTelemetry(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error)

type Config struct {
    ServiceName string
    Environment string
    Endpoint    string  // OTLP gRPC address; empty → no export (local dev)
    Insecure    bool    // disable TLS — for local collectors and staging
    SampleRate  float64 // 0.0–1.0; 0 or 1.0 → always sample
}
```

## Usage

```go
import "github.com/juanmaAV/go-utils/telemetry"

shutdown, err := telemetry.InitTelemetry(ctx, telemetry.Config{
    ServiceName: "order-service",
    Environment: "production",
    Endpoint:    "collector.internal:4317",
    SampleRate:  0.1, // sample 10% of traces
})
if err != nil {
    log.Fatal(err)
}
defer shutdown(ctx)
```

### Local development (no exporter)

```go
shutdown, err := telemetry.InitTelemetry(ctx, telemetry.Config{
    ServiceName: "order-service",
    Environment: "local",
    // Endpoint empty → noop, no gRPC connection
})
```

### Local collector (no TLS)

```go
shutdown, err := telemetry.InitTelemetry(ctx, telemetry.Config{
    ServiceName: "order-service",
    Environment: "staging",
    Endpoint:    "localhost:4317",
    Insecure:    true,
})
```

## Sampling

| `SampleRate` | Behavior |
|---|---|
| `0` (zero value) | Always sample |
| `1.0` | Always sample |
| `0.1` | Sample 10% (TraceIDRatioBased) |
| `0.01` | Sample 1% |

## Notes

- Must be called once at service startup, before any spans are created
- The returned `shutdown` func flushes pending spans — always defer it
- After `InitTelemetry`, use `otel.Tracer("component-name")` anywhere in the service
- `trace_id` and `span_id` are automatically injected into logs when using `logger.New` from this module
