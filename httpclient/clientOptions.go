package httpclient

import "time"

type ClientOption func(*ClientConfig)

func WithBaseURL(baseURL string) ClientOption {
	return func(config *ClientConfig) {
		config.BaseURL = baseURL
	}
}

func WithTimeoutClient(timeout time.Duration) ClientOption {
	return func(config *ClientConfig) {
		config.Timeout = timeout
	}
}

func WithRetryCount(count int) ClientOption {
	return func(config *ClientConfig) {
		config.RetryCount = count
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(config *ClientConfig) {
		config.UserAgent = userAgent
	}
}

func WithDefaultHeaders(headers map[string]string) ClientOption {
	return func(config *ClientConfig) {
		config.Headers = headers
	}
}

func WithServiceName(serviceName string) ClientOption {
	return func(config *ClientConfig) {
		config.ServiceName = serviceName
	}
}

func WithLogging(enable bool) ClientOption {
	return func(config *ClientConfig) {
		config.EnableLogging = enable
	}
}
