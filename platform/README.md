
# platform package

Helpers for configuration and lifecycle management of Echo servers in Go.

## Main functions
- `New(config *BasicConfig) (*Server, error)`: Creates a new Echo server with basic configuration.
- `Run() <-chan error`: Starts the server and handles graceful shutdown.
- `GetBasicServerConfig(serverName string) *BasicConfig`: Gets basic configuration from environment variables.

## `BasicConfig` structure
- `Port`: Server port (must be defined).
- `GracefullTime`: Graceful shutdown timeout.
- `Environment`: Current environment (dev, prod, etc).
- `ServerName`: Server name.

## Usage example
```go
import "github.com/juanMaAV92/go-utils/server"

config := server.GetBasicServerConfig("my-server")
srv, err := server.New(config)
if err != nil {
    log.Fatal(err)
}

errChan := srv.Run()
// Wait for server errors
if err := <-errChan; err != nil {
    log.Fatal(err)
}
```
