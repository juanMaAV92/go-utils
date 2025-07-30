
# env package

Utilities to get and validate environment variables in Go.

## Main functions
- `GetEnv(key string) string`: Gets an environment variable and panics if it's empty.
- `GetEnviroment() string`: Returns the current environment (local, prod, etc).
- `GetEnvAsDurationWithDefault(key, defaultValue string) time.Duration`: Gets an environment variable as a duration, with a default value.

## Usage example
```go
import "github.com/juanMaAV92/go-utils/env"

value := env.GetEnv("MY_ENV_VAR")
```
