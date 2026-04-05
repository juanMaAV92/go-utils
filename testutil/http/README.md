# testutil/http

Framework-agnostic helpers for HTTP handler tests. Uses only `net/http` and `net/http/httptest` — works with any Go HTTP handler, not just Echo.

## Usage

```go
import (
    "net/http"
    "net/http/httptest"
    "testing"

    httptest "github.com/juanmaAV/go-utils/testutil/http"
)

func TestGetUser(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    httptest.AssertStatus(t, rec, http.StatusOK)
    httptest.AssertJSONField(t, rec, "email", "user@example.com")
}
```

## API

```go
// Build a request. body can be string, []byte, or any JSON-serialisable value.
// Content-Type is always application/json.
NewRequest(method, url string, body any) *http.Request

// Fail the test if the status code differs from expected.
AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int)

// Decode the response body JSON into v. Fails the test on error.
DecodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v any)

// Check that a top-level JSON string field in the response body equals want.
AssertJSONField(t *testing.T, rec *httptest.ResponseRecorder, key, want string)

// Serialise v to JSON string. Panics on error — safe for test setup only.
MustJSON(v any) string
```

## Example — full test

```go
func TestCreateItem(t *testing.T) {
    tests := []struct {
        name   string
        body   any
        status int
        field  string
        value  string
    }{
        {
            name:   "success",
            body:   map[string]string{"name": "widget"},
            status: http.StatusCreated,
            field:  "id",
            value:  "item-1",
        },
        {
            name:   "empty body",
            body:   nil,
            status: http.StatusBadRequest,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/items", tc.body)
            rec := httptest.NewRecorder()

            handler.ServeHTTP(rec, req)

            httptest.AssertStatus(t, rec, tc.status)
            if tc.field != "" {
                httptest.AssertJSONField(t, rec, tc.field, tc.value)
            }
        })
    }
}
```
