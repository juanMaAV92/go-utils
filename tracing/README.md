
# tracing package

Helpers for distributed tracing in Go projects.

## Features
- Utilities to add, manage, and propagate trace information.
- Designed to integrate with logging and monitoring systems.

## Usage example
```go
import (
    "context"
    "github.com/juanMaAV92/go-utils/tracing"
)

// Initialize tracing (usually at app startup)
cfg := tracing.NewTracingConfig("my-service", "localhost:4318", "dev")
shutdown, err := tracing.InitTracing(cfg)
if err != nil {
    panic(err)
}
defer shutdown(context.Background())
```
