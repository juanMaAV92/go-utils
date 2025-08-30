package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Meraki-Nubia/nubia-go-libs/log"
	"github.com/Meraki-Nubia/nubia-go-libs/tracing"
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	resty       *resty.Client
	logger      log.Logger
	tracer      trace.Tracer
	serviceName string
	config      *ClientConfig
}

type ClientConfig struct {
	BaseURL       string
	Timeout       time.Duration
	RetryCount    int
	UserAgent     string
	Headers       map[string]string
	ServiceName   string
	EnableLogging bool
}

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Success    bool
}

func NewClient(logger log.Logger, options ...ClientOption) HTTPClient {
	config := &ClientConfig{
		Timeout:       30 * time.Second,
		RetryCount:    0,
		UserAgent:     "go-http-client/1.0",
		Headers:       make(map[string]string),
		ServiceName:   "go-server",
		EnableLogging: false,
	}

	for _, option := range options {
		option(config)
	}

	restyClient := resty.New().
		SetTimeout(config.Timeout).
		SetRetryCount(config.RetryCount).
		SetHeader("User-Agent", config.UserAgent)

	if config.BaseURL != "" {
		restyClient.SetBaseURL(config.BaseURL)
	}

	for key, value := range config.Headers {
		restyClient.SetHeader(key, value)
	}

	tracer := tracing.GetTracer(config.ServiceName)

	client := &Client{
		resty:       restyClient,
		logger:      logger,
		tracer:      tracer,
		serviceName: config.ServiceName,
		config:      config,
	}

	client.setupTracingMiddleware()

	return client
}

func (c *Client) setupTracingMiddleware() {
	// Before request - crear child span para HTTP call
	c.resty.OnBeforeRequest(func(client *resty.Client, req *resty.Request) error {
		ctx := req.Context()

		// Crear child span para el HTTP call
		spanName := fmt.Sprintf("HTTP %s", req.Method)
		ctx, span := c.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))

		// Set span attributes
		span.SetAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL),
			attribute.String("http.user_agent", req.Header.Get("User-Agent")),
			attribute.String("component", "http-client"),
			attribute.String("service.name", c.serviceName),
		)

		// Inject trace context into headers para distributed tracing (estándar OpenTelemetry)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

		// ADICIONAL: Inyectar trace_id como header explícito para compatibilidad
		traceID := tracing.GetTraceIDFromContext(ctx)
		if traceID != "" {
			req.SetHeader("X-Trace-Id", traceID)
		}

		// Update request context
		req.SetContext(ctx)

		return nil
	})

	// After response - completar span y log consolidado
	c.resty.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
		ctx := resp.Request.Context()
		span := trace.SpanFromContext(ctx)

		if span != nil {
			// Set response attributes
			span.SetAttributes(
				attribute.Int("http.status_code", resp.StatusCode()),
				attribute.String("http.response.size", fmt.Sprintf("%d", len(resp.Body()))),
				attribute.String("http.response_time", resp.Time().String()),
			)

			// Set span status based on HTTP status
			if resp.StatusCode() >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode()))
				if resp.StatusCode() >= 500 {
					span.RecordError(fmt.Errorf("server error: %d", resp.StatusCode()))
				}
			} else {
				span.SetStatus(codes.Ok, "")
			}

			span.End()
		}

		// Determinar si logging está habilitado
		shouldLog := c.shouldLogRequest(ctx)

		// Log consolidado solo si está habilitado
		if shouldLog {
			traceID := tracing.GetTraceIDFromContext(ctx)

			c.logger.Info(ctx, "http_call_completed", "HTTP request completed",
				log.Field("method", resp.Request.Method),
				log.Field("url", resp.Request.URL),
				log.Field("status_code", resp.StatusCode()),
				log.Field("response_time", resp.Time()),
				log.Field("response_size", len(resp.Body())),
				log.Field("success", resp.IsSuccess()),
				log.Field("trace_id", traceID),
			)
		}

		return nil
	})

	// On error - record error in span y log de error (siempre habilitado)
	c.resty.OnError(func(req *resty.Request, err error) {
		ctx := req.Context()
		span := trace.SpanFromContext(ctx)

		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.End()
		}

		// Log de error siempre habilitado (es importante)
		c.logger.Error(ctx, "http_call_failed", "HTTP request failed",
			log.Field("method", req.Method),
			log.Field("url", req.URL),
			log.Field("error", err.Error()),
			log.Field("trace_id", tracing.GetTraceIDFromContext(ctx)),
		)
	})
}

func (c *Client) Get(ctx context.Context, url string, options ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "GET", url, nil, options...)
}

func (c *Client) Post(ctx context.Context, url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "POST", url, body, options...)
}

func (c *Client) Put(ctx context.Context, url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "PUT", url, body, options...)
}

func (c *Client) Delete(ctx context.Context, url string, options ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "DELETE", url, nil, options...)
}

func (c *Client) Patch(ctx context.Context, url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.executeRequest(ctx, "PATCH", url, body, options...)
}

func (c *Client) executeRequest(ctx context.Context, method, url string, body interface{}, options ...RequestOption) (*Response, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	// Apply request options
	config := &RequestConfig{
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}

	for _, option := range options {
		option(config)
	}

	// Create resty request with context
	req := c.resty.R().SetContext(ctx)

	// Store request config in context for middleware access
	ctx = context.WithValue(ctx, "request_config", config)
	req.SetContext(ctx)

	// Set headers
	for key, value := range config.Headers {
		req.SetHeader(key, value)
	}

	// Set query params
	for key, value := range config.QueryParams {
		req.SetQueryParam(key, value)
	}

	// Set timeout if specified - note: timeout is handled at client level, not request level

	// Set body if provided
	if body != nil {
		req.SetBody(body)
		if _, exists := config.Headers["Content-Type"]; !exists {
			req.SetHeader("Content-Type", "application/json")
		}
	}

	// Execute request (tracing middleware will handle spans)
	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "DELETE":
		resp, err = req.Delete(url)
	case "PATCH":
		resp, err = req.Patch(url)
	default:
		return nil, errors.New("unsupported HTTP method: " + method)
	}

	if err != nil {
		if resp != nil {
			return &Response{
				StatusCode: resp.StatusCode(),
				Body:       resp.Body(),
				Headers:    parseHeaders(resp.Header()),
				Success:    false,
			}, fmt.Errorf("request failed: %w; response body: %s", err, string(resp.Body()))
		}
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode(),
		Body:       resp.Body(),
		Headers:    parseHeaders(resp.Header()),
		Success:    resp.IsSuccess(),
	}, nil
}

func parseHeaders(headers map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// Helper methods for response handling
func (r *Response) JSON(dest interface{}) error {
	return json.Unmarshal(r.Body, dest)
}

func (r *Response) String() string {
	return string(r.Body)
}

// shouldLogRequest determina si debe generar logs basado en la configuración del request o cliente
func (c *Client) shouldLogRequest(ctx context.Context) bool {
	// Intentar extraer configuración del request desde contexto
	if requestConfigValue := ctx.Value("request_config"); requestConfigValue != nil {
		if requestConfig, ok := requestConfigValue.(*RequestConfig); ok {
			// Si hay configuración específica del request, usarla
			if requestConfig.EnableLogging != nil {
				return *requestConfig.EnableLogging
			}
		}
	}

	// Si no hay configuración específica del request, usar la del cliente
	return c.config.EnableLogging
}
