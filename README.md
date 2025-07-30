# go-utils

Utilities for working with environment variables, pointers, paths, and Echo server lifecycle in Go.

## Packages

### env
Get and validate environment variables easily.
- `GetEnv(key string) string`: Gets an environment variable and panics if empty.
- `GetEnviroment() string`: Returns the current environment.
- `GetEnvAsDurationWithDefault(key, defaultValue string) time.Duration`: Gets an environment variable as a duration, with default.

### pointers
Helpers for working with pointers.
- `Pointer[T any](v T) *T`: Returns a pointer to any value.
- `FirstNonNil[T any](values ...*T) *T`: Returns the first non-nil pointer from a list.

### server
Configuration and lifecycle management for Echo servers.
- `New(config *BasicConfig) (*Server, error)`: Creates a new Echo server.
- `Run() <-chan error`: Starts the server and handles graceful shutdown.
- `GetBasicServerConfig(serverName string) *BasicConfig`: Gets basic config from environment variables.

---

## Installation
```bash
go get github.com/juanMaAV92/go-utils