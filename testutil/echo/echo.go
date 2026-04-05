// Package echo provides helpers for table-driven handler tests using Echo.
package echo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

// Case holds the parameters for a single table-driven handler test.
type Case struct {
	Name        string
	Request     Request
	RequestBody any
	Response    ExpectedResponse
}

// Request describes the HTTP request to build.
type Request struct {
	Method     string
	Url        string
	PathParam  []Param
	QueryParam []Param
	Header     map[string]string
}

// Param is a name/value pair used for path parameters and query parameters.
type Param struct {
	Name  string
	Value string
}

// ExpectedResponse describes the expected HTTP response for an assertion.
type ExpectedResponse struct {
	Status int
	Body   *string
}

// PrepareContext builds an Echo context and response recorder from a Case.
// The request body is JSON-encoded unless it is already a string.
// Content-Type is always set to application/json.
func PrepareContext(e *echo.Echo, tc Case) (echo.Context, *httptest.ResponseRecorder) {
	body, _ := json.Marshal(tc.RequestBody)
	if str, ok := tc.RequestBody.(string); ok {
		body = []byte(str)
	}

	rawURL := buildURL(tc.Request.Url, tc.Request.QueryParam)
	req := httptest.NewRequest(tc.Request.Method, rawURL, strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for k, v := range tc.Request.Header {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if len(tc.Request.PathParam) > 0 {
		names := make([]string, len(tc.Request.PathParam))
		values := make([]string, len(tc.Request.PathParam))
		for i, p := range tc.Request.PathParam {
			names[i] = p.Name
			values[i] = p.Value
		}
		ctx.SetParamNames(names...)
		ctx.SetParamValues(values...)
	}

	return ctx, rec
}

// ToJSONString serialises v to a JSON string pointer.
// Returns nil if marshalling fails.
func ToJSONString(v any) *string {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

func buildURL(base string, params []Param) string {
	if len(params) == 0 {
		return base
	}
	q := url.Values{}
	for _, p := range params {
		q.Add(p.Name, p.Value)
	}
	return base + "?" + q.Encode()
}

// NewRequest is a convenience wrapper around httptest.NewRequest that always
// sets Content-Type to application/json and accepts a body of any type.
// Passing nil produces a request with no body.
func NewRequest(method, rawURL string, body any) *http.Request {
	var bodyStr string
	if body != nil {
		if str, ok := body.(string); ok {
			bodyStr = str
		} else {
			b, _ := json.Marshal(body)
			bodyStr = string(b)
		}
	}
	req := httptest.NewRequest(method, rawURL, strings.NewReader(bodyStr))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req
}
