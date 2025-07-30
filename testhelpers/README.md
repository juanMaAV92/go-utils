# testhelpers

Helpers for HTTP endpoint testing in Go using Echo and custom error handling.

## Overview
This package provides models and utilities to simplify writing HTTP tests for Echo servers. It allows you to define test cases, requests, expected responses, and error expectations in a structured way.

## Main Types

### HttpTestCase
Defines a single HTTP test case.
- `TestName string`: Name of the test.
- `Request TestRequest`: The request to send.
- `RequestBody interface{}`: Optional request body.
- `MockFunc func(s *echo.Echo, c echo.Context)`: Function to set up mocks or handlers.
- `Response ExpectedResponse`: Expected response.
- `ExpectError *errors.ErrorResponse`: Expected error response (if any).

### TestRequest
Describes the HTTP request.
- `Method string`: HTTP method (GET, POST, etc).
- `Url string`: Request URL.
- `PathParam []TestPathParam`: Path parameters.
- `Header map[string]string`: Request headers.
- `QueryParam []TestQueryParam`: Query parameters.

### TestPathParam
Represents a path parameter.
- `Name string`: Parameter name.
- `Value string`: Parameter value.

### TestQueryParam
Represents a query parameter.
- `Name string`: Parameter name.
- `Value string`: Parameter value.

### ExpectedResponse
Describes the expected HTTP response.
- `Status int`: Expected status code.
- `Body *string`: Expected response body (optional).

## Usage Example
```go
// Example usage of HttpTestCase
var testCases = HttpTestCase{
    TestName: "Should return 200 OK",
    Request: TestRequest{
        Method: "GET",
        Url: "/api/resource",
        Header: map[string]string{"Authorization": "Bearer token"},
    },
    MockFunc: func(s *echo.Echo, c echo.Context) {
        // Setup handler or mocks
    },
    Response: ExpectedResponse{
        Status: 200,
        Body: nil,
    },
    ExpectError: nil,
}

for _, test := range testCases {
    t.Run(test.TestName, func(t *testing.T) {


        ctx, recorder := PrepareContextFormTestCase(echo.Echo, test)

        if test.MockFunc != nil {
				test.MockFunc(echo.Echo, ctx)
		}
        // Execute the request
        // Execute the assertions
        assert.NoError(t, err)
		assert.Equal(t, test.Response.Status, recorder.Code)
}
```


## License
MIT
