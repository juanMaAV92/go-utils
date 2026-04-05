package httpclient

import "net/url"

type requestConfig struct {
	Headers       map[string]string
	QueryParams   map[string]string
	EnableLogging *bool
	FormData      url.Values
}

// RequestOption configures a single HTTP request.
type RequestOption func(*requestConfig)

// WithHeader adds a single header to the request.
func WithHeader(key, value string) RequestOption {
	return func(c *requestConfig) {
		c.Headers[key] = value
	}
}

// WithHeaders merges the given headers into the request.
func WithHeaders(headers map[string]string) RequestOption {
	return func(c *requestConfig) {
		for k, v := range headers {
			c.Headers[k] = v
		}
	}
}

// WithQueryParam adds a single query parameter.
func WithQueryParam(key, value string) RequestOption {
	return func(c *requestConfig) {
		c.QueryParams[key] = value
	}
}

// WithQueryParams merges the given query parameters into the request.
func WithQueryParams(params map[string]string) RequestOption {
	return func(c *requestConfig) {
		for k, v := range params {
			c.QueryParams[k] = v
		}
	}
}

// WithAuthToken sets the Authorization: Bearer header.
func WithAuthToken(token string) RequestOption {
	return WithHeader("Authorization", "Bearer "+token)
}

// WithRequestLogging overrides the client-level logging setting for this request.
func WithRequestLogging(enable bool) RequestOption {
	return func(c *requestConfig) { c.EnableLogging = &enable }
}

// WithFormData sets form-encoded body data.
// Automatically sets Content-Type to application/x-www-form-urlencoded.
func WithFormData(data map[string]string) RequestOption {
	return func(c *requestConfig) {
		if c.FormData == nil {
			c.FormData = make(url.Values)
		}
		for k, v := range data {
			c.FormData.Set(k, v)
		}
		c.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}
}
