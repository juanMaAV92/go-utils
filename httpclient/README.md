# httpclient

A robust and extensible HTTP client for Go, built on top of [resty](https://github.com/go-resty/resty) and integrated with OpenTelemetry tracing and flexible logging.

## Installation

Add the dependency to your `go.mod`:

```
go get github.com/juanMaAV92/go-utils/httpclient
```

## Features
- Simple API for GET, POST, PUT, DELETE, PATCH requests
- Configurable base URL, headers, timeouts, retries, and user agent
- Per-request options for headers, query params, and logging
- Distributed tracing with OpenTelemetry (automatic span creation and propagation)
- Structured logging for requests and errors
- Response helpers for JSON and string parsing

## Basic Usage

```go
import (
    "context"
    "github.com/juanMaAV92/go-utils/httpclient"
    "github.com/juanMaAV92/go-utils/log"
)

logger := log.NewLogger()
client := httpclient.NewClient(logger, httpclient.WithBaseURL("https://api.example.com"))

resp, err := client.Get(context.Background(), "/users", httpclient.WithQueryParam("active", "true"))
if err != nil {
    // Handle error
}

var users []User
if err := resp.JSON(&users); err != nil {
    // Handle JSON error
}
```

## Per-request Options
You can set headers, query params, and enable/disable logging for each request:

```go
resp, err := client.Post(ctx, "/login", loginPayload,
    httpclient.WithHeader("Authorization", "Bearer ..."),
    httpclient.WithQueryParam("source", "web"),
    httpclient.WithEnableLogging(true),
)
```

## Tracing & Logging
- Tracing spans are created for each request and propagated via headers.
- Logs include method, URL, status code, response time, and trace ID.

## License

MIT
