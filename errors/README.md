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
import "github.com/juanmaAV/go-utils/errors"

// Return predefined
return errors.ErrNotFound()
return errors.ErrUnauthorized("token expired")

// Custom message on predefined
return errors.ErrBadRequest().WithMessage("email is required")
return errors.ErrBadRequest().WithMessages([]string{"email required", "name required"})

// Fully custom
return errors.New(http.StatusConflict, "CONFLICT", []string{"resource already exists"})

// Type-check in middleware
var appErr *errors.ErrorResponse
if errors.As(err, &appErr) {
    // appErr.ErrorHTTPCode(), appErr.ErrorCode()
}
```

## Echo integration

```go
import echoerr "github.com/juanmaAV/go-utils/errors/echo"

e.HTTPErrorHandler = echoerr.HTTPErrorHandler
```

Handles both `*errors.ErrorResponse` and `*echo.HTTPError`, normalizing both into the same JSON shape.
