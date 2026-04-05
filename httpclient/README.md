# httpclient

HTTP client backed by [go-resty/resty](https://github.com/go-resty/resty) with OTel trace propagation and structured logging.

## Interface

```go
type Client interface {
    Get(ctx context.Context, url string, opts ...RequestOption) (*Response, error)
    Post(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)
    Put(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)
    Delete(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)
    Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)
}

type Response struct {
    StatusCode int
    Body       []byte
    Headers    map[string]string
    Success    bool
}

func (r *Response) JSON(dest any) error
func (r *Response) String() string
```

## Constructor

```go
func New(log logger.Logger, opts ...ClientOption) Client
```

Pass `nil` for `log` to disable request logging.

## Client options

```go
httpclient.New(log,
    httpclient.WithBaseURL("https://api.example.com"),
    httpclient.WithTimeout(10*time.Second),
    httpclient.WithRetryCount(3),
    httpclient.WithServiceName("payments-api"),   // used in logs and spans
    httpclient.WithDefaultHeaders(map[string]string{"X-API-Key": key}),
    httpclient.WithLogging(true),
)
```

## Request options

```go
resp, err := c.Get(ctx, "/users",
    httpclient.WithQueryParam("page", "1"),
    httpclient.WithQueryParams(map[string]string{"sort": "asc"}),
    httpclient.WithHeader("X-Request-ID", id),
    httpclient.WithAuthToken(token),             // sets Authorization: Bearer
    httpclient.WithRequestLogging(false),        // overrides client-level logging
)

resp, err := c.Post(ctx, "/form", nil,
    httpclient.WithFormData(map[string]string{"grant_type": "client_credentials"}),
)
```

## Usage

```go
import "github.com/juanmaAV/go-utils/httpclient"

c := httpclient.New(log,
    httpclient.WithBaseURL("https://api.example.com"),
    httpclient.WithTimeout(5*time.Second),
    httpclient.WithServiceName("inventory-api"),
    httpclient.WithLogging(true),
)

resp, err := c.Get(ctx, "/products", httpclient.WithQueryParam("limit", "50"))
if err != nil {
    // non-2xx responses also return an error with the status code
    return err
}

var products []Product
if err := resp.JSON(&products); err != nil {
    return err
}
```

## Behavior

- Non-2xx responses return both a non-nil `*Response` and a non-nil `error`
- OTel trace context is propagated via W3C `traceparent` header automatically
- Per-request timeout: use `context.WithTimeout(ctx, d)` — there is no per-request timeout option by design
- `body` is serialized as JSON; `Content-Type: application/json` is set automatically unless overridden
- `WithFormData` overrides `body` and sets `Content-Type: application/x-www-form-urlencoded`
