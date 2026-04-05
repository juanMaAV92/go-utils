package httpclient

import "time"

type clientConfig struct {
	BaseURL       string
	Timeout       time.Duration
	RetryCount    int
	UserAgent     string
	Headers       map[string]string
	ServiceName   string
	EnableLogging bool
}

// ClientOption configures the Client at construction time.
type ClientOption func(*clientConfig)

// WithBaseURL sets the base URL prepended to every request path.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *clientConfig) { c.BaseURL = baseURL }
}

// WithTimeout sets the per-client request timeout. Default: 60s.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) { c.Timeout = timeout }
}

// WithRetryCount sets the number of retries on failure. Default: 0.
func WithRetryCount(count int) ClientOption {
	return func(c *clientConfig) { c.RetryCount = count }
}

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *clientConfig) { c.UserAgent = userAgent }
}

// WithDefaultHeaders sets headers sent on every request.
func WithDefaultHeaders(headers map[string]string) ClientOption {
	return func(c *clientConfig) { c.Headers = headers }
}

// WithServiceName sets the downstream service name used in log fields and spans.
func WithServiceName(name string) ClientOption {
	return func(c *clientConfig) { c.ServiceName = name }
}

// WithLogging enables or disables request/response logging for all requests.
// Can be overridden per-request with WithRequestLogging.
func WithLogging(enable bool) ClientOption {
	return func(c *clientConfig) { c.EnableLogging = enable }
}
