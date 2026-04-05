# testutil/echo

Helpers for table-driven HTTP handler tests using [Echo](https://echo.labstack.com/).

## Usage

```go
import (
    "net/http"
    "testing"

    "github.com/labstack/echo/v4"
    echotest "github.com/juanmaAV/go-utils/testutil/echo"
)

func TestCreateOrder(t *testing.T) {
    e := echo.New()

    tests := []echotest.Case{
        {
            Name: "success",
            Request: echotest.Request{
                Method: http.MethodPost,
                Url:    "/orders",
                Header: map[string]string{"X-User-Code": "user-123"},
            },
            RequestBody: map[string]any{"product_id": "abc"},
            Response:    echotest.ExpectedResponse{Status: http.StatusCreated},
        },
        {
            Name: "missing body",
            Request: echotest.Request{
                Method: http.MethodPost,
                Url:    "/orders",
            },
            Response: echotest.ExpectedResponse{Status: http.StatusBadRequest},
        },
    }

    for _, tc := range tests {
        t.Run(tc.Name, func(t *testing.T) {
            ctx, rec := echotest.PrepareContext(e, tc)
            _ = createOrderHandler(ctx)
            if rec.Code != tc.Response.Status {
                t.Errorf("got %d, want %d", rec.Code, tc.Response.Status)
            }
        })
    }
}
```

### Path and query parameters

```go
echotest.Case{
    Request: echotest.Request{
        Method: http.MethodGet,
        Url:    "/users/123",
        PathParam:  []echotest.Param{{Name: "id", Value: "123"}},
        QueryParam: []echotest.Param{{Name: "include", Value: "roles"}},
    },
}
```

### Helpers

```go
// Serialise a value to a JSON string pointer (returns nil on error).
s := echotest.ToJSONString(myStruct)

// Build a standalone *http.Request with Content-Type: application/json.
req := echotest.NewRequest(http.MethodPost, "/items", myBody)
```
