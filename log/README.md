# log package

Helpers for structured logging in Go projects.

## Features
- Add fields to logs easily with options.
- Compose log fields and pass them through your application.

## Main types and functions
- `Opts`: Struct to hold log fields.
- `Fields(fields map[string]interface{}) Opts`: Create options from a map of fields.
- `Field(key string, value interface{}) Opts`: Create options for a single field.
- `AddField(key string, value interface{})`: Add a field to an existing Opts.

## Usage example
```go
import "github.com/juanMaAV92/go-utils/log"

logger := log.New("service-name", log.WithLevel(log.InfoLevel))

ctx := context.Background()
    
// Log simple
logger.Info(ctx, "startup", "Service started")

// Log with additional fields
logger.Info(ctx, "user_created", "New user registered",
    log.Field("user_id", "12345"),
    log.Field("email", "user@example.com"))

// Error logging
logger.Error(ctx, "database_error", "Connection failed",
    log.Field("error", err.Error()))

// Log with structured fields
opts := log.Fields(map[string]interface{}{"user_id": 123, "action": "login"})
opts.AddField("ip", "127.0.0.1")
```
