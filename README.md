# go-utils

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/dl/)

Utilities for working with environment variables, pointers, paths, and Echo server lifecycle in Go.

## 📋 Tabla de Contenido

- [Installation](#installation)
- [Available Packages](#available-packages)
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

### server
Configuration and lifecycle management for Echo servers.
- `New(config *BasicConfig) (*Server, error)`: Creates a new Echo server.
- `Run() <-chan error`: Starts the server and handles graceful shutdown.
- `GetBasicServerConfig(serverName string) *BasicConfig`: Gets basic config from environment variables.

### Tracing
Distributed tracing helpers for Go applications.

- `InitTracing(config TracingConfig) (func(context.Context) error, error)`: Initializes tracing with the provided configuration.
- `GetTracer(name string) trace.Tracer`: Retrieves a tracer for the specified name.
- `GetTraceIDFromContext(ctx context.Context) string`: Extracts the trace ID from the context.

---
## Contributing
Contributions are welcome! 