# validator

Struct validation via [go-playground/validator](https://github.com/go-playground/validator) with structured error output compatible with the `errors` package.

## Key behaviors

- JSON tag names in error messages — `"email is required"`, not `"Email is required"`
- All failing fields returned in a single response — no fail-fast
- Structs and slices/arrays supported
- Singleton instance — safe to call `New()` anywhere
- Validation failure → HTTP 422 with `VALIDATION_ERROR` code
- Bind failure → HTTP 400 with `INVALID_REQUEST` code

## Interface

```go
type Validator struct{ ... }

func New() *Validator
func (v *Validator) Validate(data any) error
func (v *Validator) BindAndValidate(binder Binder, req any) error

type Binder interface {
    Bind(any) error
}
```

## Usage

```go
import "github.com/juanmaAV/go-utils/validator"

v := validator.New()

// Validate a struct
err := v.Validate(req)

// Bind + validate in one call (works with Echo, Gin, Chi, net/http — any Binder)
err := v.BindAndValidate(c, &req)
```

### Echo setup

```go
import (
    "github.com/juanmaAV/go-utils/validator"
    "github.com/labstack/echo/v4"
)

e := echo.New()
e.Validator = validator.New()   // *Validator implements echo.Validator
```

Wait — to satisfy `echo.Validator`, Echo expects `Validate(i interface{}) error`. `*Validator` already implements that signature, so the assignment works directly.

## Error response

```json
{
  "code": "VALIDATION_ERROR",
  "messages": [
    "email is required",
    "username must be at least 3 characters",
    "age must be greater than or equal to 18"
  ]
}
```

Bind failure:
```json
{
  "code": "INVALID_REQUEST",
  "messages": ["invalid request format"]
}
```

## Supported tags

| Tag | Message |
|---|---|
| `required` | `{field} is required` |
| `email` | `{field} must be a valid email` |
| `url` | `{field} must be a valid URL` |
| `uuid` | `{field} must be a valid UUID` |
| `min` | `{field} must be at least {n} characters` |
| `max` | `{field} must be at most {n} characters` |
| `len` | `{field} must be exactly {n} characters` |
| `gt` / `gte` / `lt` / `lte` | numeric range messages |
| `oneof` | `{field} must be one of: {values}` |
| `numeric` / `alpha` / `alphanum` | format messages |
| `datetime` | `{field} must be a valid datetime` |
| `required_if` / `required_unless` / `required_with` / `required_without` | conditional messages |
| _(unknown)_ | `{field} is invalid ({tag})` |

## Slice validation

```go
type Item struct {
    Name string `json:"name" validate:"required"`
}

items := []Item{{Name: "ok"}, {Name: ""}}
err := v.Validate(items)
// → "[1].name is required"
```
