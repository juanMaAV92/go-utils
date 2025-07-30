# go-utils

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/dl/)

Utilities for working with environment variables, pointers, paths, and Echo server lifecycle in Go.

## 📋 Content table

- [Installation](#installation)
- [Available Packages](#available-packages)
    - [env](#env)
    - [log](#log)
    - [middleware](#middleware)
        - [TraceId](#traceid)
        - [Tracing](#tracing)
        - [Logging](#logging)
    - [Path](#path)
    - [Pointers](#pointers)
    - [server](#server)
    - [Testhelpers](#testhelpers)
    - [Tracing](#tracing)
- [Contributing](#contributing)


## Installation
```bash
go get github.com/juanMaAV92/go-utils
```
---

## Available Packages

### env
Get and validate environment variables easily.
- `GetEnv(key string) string`: Gets an environment variable and panics if empty.
- `GetEnviroment() string`: Returns the current environment.
- `GetEnvAsDurationWithDefault(key, defaultValue string) time.Duration`: Gets an environment variable as a duration, with default.


### log
Helpers for structured logging.
- `Fields(fields map[string]interface{}) Opts`: Create options from a map of fields.
- `Field(key string, value interface{}) Opts`: Create options for a single field.
- `AddField(key string, value interface{})`: Add a field to an existing Opts.

### middleware
Common middleware functions for Echo.
- `TraceId() echo.MiddlewareFunc`: Middleware to add a trace ID to the context.
- `Tracing() echo.MiddlewareFunc`: Middleware to start and stop tracing for requests.
- `Logging() echo.MiddlewareFunc`: Middleware to log requests and responses.

### Path
Path manipulation utilities.

### Pointers
Pointer utilities for working with values.

### server
Configuration and lifecycle management for Echo servers.
- `New(config *BasicConfig) (*Server, error)`: Creates a new Echo server.
- `Run() <-chan error`: Starts the server and handles graceful shutdown.
- `GetBasicServerConfig(serverName string) *BasicConfig`: Gets basic config from environment variables.

### Testhelpers
Helpers for HTTP endpoint testing with Echo. Provides models to define requests, responses, and error expectations for tests.

### Tracing
Distributed tracing helpers for Go applications.

- `InitTracing(config TracingConfig) (func(context.Context) error, error)`: Initializes tracing with the provided configuration.
- `GetTracer(name string) trace.Tracer`: Retrieves a tracer for the specified name.
- `GetTraceIDFromContext(ctx context.Context) string`: Extracts the trace ID from the context.

---
## Contributing
Contributions are welcome! 