package httpclient

import "time"

type RequestConfig struct {
	Headers       map[string]string
	QueryParams   map[string]string
	Timeout       time.Duration
	EnableLogging *bool
}

type RequestOption func(*RequestConfig)

func WithHeader(key, value string) RequestOption {
	return func(config *RequestConfig) {
		if config.Headers == nil {
			config.Headers = make(map[string]string)
		}
		config.Headers[key] = value
	}
}

func WithHeaders(headers map[string]string) RequestOption {
	return func(config *RequestConfig) {
		if config.Headers == nil {
			config.Headers = make(map[string]string)
		}
		for k, v := range headers {
			config.Headers[k] = v
		}
	}
}

func WithQueryParam(key, value string) RequestOption {
	return func(config *RequestConfig) {
		if config.QueryParams == nil {
			config.QueryParams = make(map[string]string)
		}
		config.QueryParams[key] = value
	}
}

func WithQueryParams(params map[string]string) RequestOption {
	return func(config *RequestConfig) {
		if config.QueryParams == nil {
			config.QueryParams = make(map[string]string)
		}
		for k, v := range params {
			config.QueryParams[k] = v
		}
	}
}

func WithRequestTimeout(timeout time.Duration) RequestOption {
	return func(config *RequestConfig) {
		config.Timeout = timeout
	}
}

func WithAuthToken(token string) RequestOption {
	return WithHeader("Authorization", "Bearer "+token)
}

func WithRequestLogging(enable bool) RequestOption {
	return func(config *RequestConfig) {
		config.EnableLogging = &enable
	}
}
