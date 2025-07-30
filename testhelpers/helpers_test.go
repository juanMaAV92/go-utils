package testhelpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func Test_PrepareQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		query    []TestQueryParam
		expected string
	}{
		{
			name:     "No query params",
			baseURL:  "http://example.com",
			query:    []TestQueryParam{},
			expected: "http://example.com",
		},
		{
			name:     "Single query param",
			baseURL:  "http://example.com",
			query:    []TestQueryParam{{Name: "key", Value: "value"}},
			expected: "http://example.com?key=value",
		},
		{
			name:    "Multiple query params",
			baseURL: "http://example.com",
			query: []TestQueryParam{
				{Name: "key1", Value: "value1"},
				{Name: "key2", Value: "value2"},
			},
			expected: "http://example.com?key1=value1&key2=value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prepareQueryParams(tt.baseURL, tt.query)
			if result != tt.expected {
				t.Errorf("prepareQueryParams(%q, %v) = %q; want %q", tt.baseURL, tt.query, result, tt.expected)
			}
		})
	}
}

func Test_PrepareHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected map[string]string
	}{
		{
			name:     "No headers",
			headers:  map[string]string{},
			expected: map[string]string{echo.HeaderContentType: echo.MIMEApplicationJSON},
		},
		{
			name: "Single header",
			headers: map[string]string{
				"Authorization": "Bearer token",
			},
			expected: map[string]string{
				echo.HeaderContentType: echo.MIMEApplicationJSON,
				"Authorization":        "Bearer token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			prepareHeaders(req, tt.headers)

			for key, value := range tt.expected {
				if req.Header.Get(key) != value {
					t.Errorf("Header %q = %q; want %q", key, req.Header.Get(key), value)
				}
			}
		})
	}
}

func Test_PreparePathParams(t *testing.T) {
	tests := []struct {
		name       string
		pathParams []TestPathParam
		expected   []string
	}{
		{
			name:       "No path params",
			pathParams: []TestPathParam{},
			expected:   []string{},
		},
		{
			name: "Single path param",
			pathParams: []TestPathParam{
				{Name: "id", Value: "123"},
			},
			expected: []string{"123"},
		},
		{
			name: "Multiple path params",
			pathParams: []TestPathParam{
				{Name: "id", Value: "123"},
				{Name: "name", Value: "test"},
			},
			expected: []string{"123", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			preparePathParams(ctx, tt.pathParams)

			for i, param := range tt.pathParams {
				if ctx.ParamNames()[i] != param.Name {
					t.Errorf("Param name %d = %q; want %q", i, ctx.ParamNames()[i], param.Name)
				}
				if ctx.ParamValues()[i] != param.Value {
					t.Errorf("Param value %d = %q; want %q", i, ctx.ParamValues()[i], param.Value)
				}
			}
		})
	}
}
