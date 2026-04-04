# env

Environment variable parsing with type conversion. Designed for fail-fast startup: required variables panic immediately if missing, not at request time.

## Functions

```go
func MustHave(keys ...string)
```
Panics if any variable is unset or blank. Lists **all** missing variables in a single message.

```go
func GetEnv(key string) string
```
Returns the value of `key`. Panics if unset or blank.

```go
func GetEnvWithDefault(key, defaultValue string) string
```
Returns the value of `key`, or `defaultValue` if unset or blank.

```go
func GetEnvironment() string
```
Returns `ENVIRONMENT` env var, defaulting to `"local"`.

```go
func GetEnvAsIntWithDefault(key string, defaultValue int) int
```
Returns `key` parsed as `int`, or `defaultValue` if unset. Panics if set but not a valid integer.

```go
func GetEnvAsDurationWithDefault(key string, defaultValue time.Duration) time.Duration
```
Returns `key` parsed as `time.Duration`, or `defaultValue` if unset. Panics if set but not a valid duration string (`"30s"`, `"1m"`, etc.).

```go
func GetEnvAsBoolWithDefault(key string, defaultValue bool) bool
```
Returns `key` parsed as `bool`, or `defaultValue` if unset. Accepts `"1"`, `"t"`, `"true"`, `"0"`, `"f"`, `"false"`. Panics if set but not a valid bool string.

```go
func GetEnvAsSliceWithDefault(key, sep string, defaultValue []string) []string
```
Splits `key` by `sep` into a string slice. Trims whitespace and excludes empty elements. Returns `defaultValue` if unset or blank.

## Constants

```go
env.EnvironmentKey    // "ENVIRONMENT"
env.LocalEnvironment  // "local"
```

## Usage

```go
import "github.com/juanmaAV/go-utils/env"

// Validate all required vars at once — fail with a clear message
env.MustHave("PORT", "DATABASE_URL", "JWT_SECRET")

// Required
port := env.GetEnv("PORT")

// Optional with defaults
host     := env.GetEnvWithDefault("HOST", "localhost")
timeout  := env.GetEnvAsDurationWithDefault("TIMEOUT", 30*time.Second)
maxConns := env.GetEnvAsIntWithDefault("MAX_CONNS", 10)
debug    := env.GetEnvAsBoolWithDefault("DEBUG", false)
origins  := env.GetEnvAsSliceWithDefault("ALLOWED_ORIGINS", ",", []string{"http://localhost:3000"})
environ  := env.GetEnvironment() // "local", "production", etc.
```

## Behavior

| Scenario | `GetEnv` | `GetEnvWithDefault` / typed variants |
|---|---|---|
| Variable set, valid | returns value | returns value |
| Variable unset or blank | **panics** | returns default |
| Variable set, invalid type | — | **panics** |

## Notes

- Call all `GetEnv` (required) functions at startup, not inside handlers. A missing required variable should crash the process at boot, not during a request.
- `GetEnvAsDurationWithDefault` accepts a `time.Duration` default (not a string), so invalid defaults are caught at compile time.
- No external dependencies.
