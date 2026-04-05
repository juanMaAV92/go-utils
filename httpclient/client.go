package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/juanmaAV/go-utils/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

// Client is the interface for making outbound HTTP requests.
type Client interface {
	Get(ctx context.Context, url string, opts ...RequestOption) (*Response, error)
	Post(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)
	Put(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)
	Delete(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)
	Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)
}

// Response holds the result of an HTTP request.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Success    bool
}

// JSON decodes the response body into dest.
func (r *Response) JSON(dest any) error {
	return json.Unmarshal(r.Body, dest)
}

// String returns the response body as a string.
func (r *Response) String() string {
	return string(r.Body)
}

type client struct {
	resty             *resty.Client
	logger            logger.Logger
	tracer            trace.Tracer
	downstreamService string
	config            *clientConfig
}

// New creates a new Client. logger is optional — pass nil to disable request logging.
func New(log logger.Logger, opts ...ClientOption) Client {
	cfg := &clientConfig{
		Timeout:       60 * time.Second,
		RetryCount:    0,
		UserAgent:     "go-http-client/1.0",
		Headers:       make(map[string]string),
		ServiceName:   "upstream",
		EnableLogging: false,
	}
	for _, o := range opts {
		o(cfg)
	}

	rc := resty.New().
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryCount).
		SetHeader("User-Agent", cfg.UserAgent).
		SetHeaders(cfg.Headers)

	if cfg.BaseURL != "" {
		rc.SetBaseURL(cfg.BaseURL)
	}

	return &client{
		resty:             rc,
		logger:            log,
		tracer:            otel.Tracer("github.com/juanmaAV/go-utils/httpclient"),
		downstreamService: cfg.ServiceName,
		config:            cfg,
	}
}

func (c *client) Get(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "GET", url, nil, opts...)
}

func (c *client) Post(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "POST", url, body, opts...)
}

func (c *client) Put(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "PUT", url, body, opts...)
}

func (c *client) Delete(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "DELETE", url, body, opts...)
}

func (c *client) Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "PATCH", url, body, opts...)
}

func (c *client) executeRequest(ctx context.Context, method, url string, body any, opts ...RequestOption) (*Response, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	ctx, span := c.tracer.Start(ctx, "HTTP "+method, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		semconv.HTTPRequestMethodKey.String(method),
		attribute.String("url.full", url),
	)

	reqCfg := &requestConfig{
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}
	for _, o := range opts {
		o(reqCfg)
	}

	shouldLog := c.config.EnableLogging
	if reqCfg.EnableLogging != nil {
		shouldLog = *reqCfg.EnableLogging
	}

	req := c.resty.R().SetContext(ctx)
	otel.GetTextMapPropagator().Inject(ctx, restyCarrier{r: req})
	req.SetHeaders(reqCfg.Headers)
	req.SetQueryParams(reqCfg.QueryParams)

	if body != nil {
		req.SetBody(body)
		if _, exists := reqCfg.Headers["Content-Type"]; !exists {
			req.SetHeader("Content-Type", "application/json")
		}
	}

	if len(reqCfg.FormData) > 0 {
		req.SetBody(reqCfg.FormData.Encode())
		req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	}

	start := time.Now()
	resp, err := dispatch(req, method, url)
	duration := time.Since(start)

	if err != nil {
		c.logRequest(ctx, shouldLog, logParams{
			method: method, url: url,
			statusCode: statusCodeOf(resp), duration: duration,
			body: body, queryParams: reqCfg.QueryParams,
			success: false, err: err,
		})
		span.RecordError(err)
		if resp != nil {
			span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode()))
			return &Response{
				StatusCode: resp.StatusCode(),
				Body:       resp.Body(),
				Headers:    parseHeaders(resp.Header()),
				Success:    false,
			}, fmt.Errorf("request failed: %w; response body: %s", err, resp.Body())
		}
		return nil, err
	}

	span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode()))

	c.logRequest(ctx, shouldLog, logParams{
		method: method, url: url,
		statusCode: resp.StatusCode(), duration: duration,
		body: body, queryParams: reqCfg.QueryParams,
		success: resp.IsSuccess(), responseBody: string(resp.Body()),
	})

	if !resp.IsSuccess() {
		e := fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Body())
		span.RecordError(e)
		return &Response{
			StatusCode: resp.StatusCode(),
			Body:       resp.Body(),
			Headers:    parseHeaders(resp.Header()),
			Success:    false,
		}, e
	}

	return &Response{
		StatusCode: resp.StatusCode(),
		Body:       resp.Body(),
		Headers:    parseHeaders(resp.Header()),
		Success:    true,
	}, nil
}

// dispatch routes to the correct resty method.
func dispatch(req *resty.Request, method, url string) (*resty.Response, error) {
	switch method {
	case "GET":
		return req.Get(url)
	case "POST":
		return req.Post(url)
	case "PUT":
		return req.Put(url)
	case "DELETE":
		return req.Delete(url)
	case "PATCH":
		return req.Patch(url)
	default:
		return nil, errors.New("unsupported HTTP method: " + method)
	}
}

// logParams groups fields for the log call to avoid a long argument list.
type logParams struct {
	method, url  string
	statusCode   int
	duration     time.Duration
	body         any
	queryParams  map[string]string
	responseBody string
	success      bool
	err          error
}

func (c *client) logRequest(ctx context.Context, enabled bool, p logParams) {
	if !enabled || c.logger == nil {
		return
	}
	fields := []any{
		"method", p.method,
		"url", p.url,
		"downstream_service", c.downstreamService,
		"status_code", p.statusCode,
		"response_time", p.duration.String(),
		"success", p.success,
	}
	if p.body != nil {
		fields = append(fields, "request_body", p.body)
	}
	if len(p.queryParams) > 0 {
		fields = append(fields, "query_params", p.queryParams)
	}
	if p.err != nil {
		fields = append(fields, "error", p.err)
		c.logger.Error(ctx, "httpclient.request", "request failed", fields...)
		return
	}
	if p.responseBody != "" {
		fields = append(fields, "response_body", p.responseBody)
	}
	c.logger.Info(ctx, "httpclient.request", "request completed", fields...)
}

// restyCarrier adapts resty.Request to propagation.TextMapCarrier.
type restyCarrier struct{ r *resty.Request }

func (c restyCarrier) Get(key string) string        { return c.r.Header.Get(key) }
func (c restyCarrier) Set(key, value string)        { c.r.SetHeader(key, value) }
func (c restyCarrier) Keys() []string {
	keys := make([]string, 0, len(c.r.Header))
	for k := range c.r.Header {
		keys = append(keys, k)
	}
	return keys
}

func parseHeaders(h map[string][]string) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out
}

func statusCodeOf(r *resty.Response) int {
	if r == nil {
		return 0
	}
	return r.StatusCode()
}
