# errors

Structured HTTP error responses that implement the `error` interface. Framework-agnostic core — Echo integration is in `errors/echo`.

## Types

```go
type ErrorResponse struct {
    Code     string   `json:"code"`
    Messages []string `json:"messages,omitempty"`
    HttpCode int      `json:"-"`
}
```

`HttpCode` is excluded from JSON. Use `ErrorHTTPCode()` to set the response status code.

## Constructor

```go
func New(httpCode int, code string, messages []string) *ErrorResponse
```

## Methods

```go
func (e *ErrorResponse) Error() string           // implements error
func (e *ErrorResponse) ErrorCode() string
func (e *ErrorResponse) ErrorMessages() []string
func (e *ErrorResponse) ErrorHTTPCode() int
func (e *ErrorResponse) WithMessage(msg string) *ErrorResponse   // returns a copy
func (e *ErrorResponse) WithMessages(msgs []string) *ErrorResponse
```

`WithMessage` / `WithMessages` return a new `*ErrorResponse` — the original is not modified.

## Predefined constructors

Each call returns a fresh `*ErrorResponse`. Modifying one instance does not affect others.

```go
ErrBadRequest(messages ...string)           // 400
ErrUnauthorized(messages ...string)         // 401
ErrForbidden(messages ...string)            // 403
ErrNotFound(messages ...string)             // 404
ErrMethodNotAllowed(messages ...string)     // 405
ErrRequestTimeout(messages ...string)       // 408
ErrTooManyRequests(messages ...string)      // 429
ErrRequestEntityTooLarge(messages ...string)// 413
ErrUnsupportedMediaType(messages ...string) // 415
ErrInternalServer(messages ...string)       // 500
ErrBadGateway(messages ...string)           // 502
ErrServiceUnavailable(messages ...string)   // 503
```

## Error code constants

```go
errors.StatusNotFoundCode        // "NOT_FOUND"
errors.StatusUnauthorizedCode    // "UNAUTHORIZED"
errors.ValidationErrorCode       // "VALIDATION_ERROR"
// ...
```

## Usage

```go
import (
    stderrors "errors"

    apperrors "github.com/juanMaAV92/go-utils/v2/errors"
)

// Return predefined
return apperrors.ErrNotFound()
return apperrors.ErrUnauthorized("token expired")

// Custom message on predefined
return apperrors.ErrBadRequest().WithMessage("email is required")
return apperrors.ErrBadRequest().WithMessages([]string{"email required", "name required"})

// Fully custom
return apperrors.New(http.StatusConflict, "CONFLICT", []string{"resource already exists"})

// Type-check in middleware — use the stdlib errors.As (this package does not
// re-export As/Is), aliasing this package to avoid the name clash.
var appErr *apperrors.ErrorResponse
if stderrors.As(err, &appErr) {
    // appErr.ErrorHTTPCode(), appErr.ErrorCode()
}
```

## Echo integration

```go
import echoerr "github.com/juanMaAV92/go-utils/errors/echo"

e.HTTPErrorHandler = echoerr.HTTPErrorHandler
```

Handles both `*errors.ErrorResponse` and `*echo.HTTPError`, normalizing both into the same JSON shape.
