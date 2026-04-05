package telemetry

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds the configuration for OpenTelemetry initialization.
type Config struct {
	ServiceName string
	Environment string

	// Endpoint is the OTLP gRPC collector address (e.g. "localhost:4317").
	// When empty, no traces are exported — useful for local development.
	Endpoint string

	// Insecure disables TLS for the gRPC connection.
	// Set to true for local collectors and staging environments.
	Insecure bool

	// SampleRate controls the fraction of traces to export (0.0–1.0).
	// Defaults to 1.0 (sample everything) when zero.
	SampleRate float64
}

// InitTelemetry initializes the OTel tracer provider and registers it globally.
// Returns a shutdown function that must be called on service exit to flush pending spans.
func InitTelemetry(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironmentName(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	sampler := resolveSampler(cfg.SampleRate)

	var tp *sdktrace.TracerProvider
	if cfg.Endpoint != "" {
		exporter, err := buildExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sampler),
		)
	} else {
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sampler),
		)
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return func(ctx context.Context) error {
		var errs error
		if err := tp.Shutdown(ctx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("tracer shutdown: %w", err))
		}
		return errs
	}, nil
}

func buildExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	}
	exp, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}
	return exp, nil
}

func resolveSampler(rate float64) sdktrace.Sampler {
	if rate <= 0 || rate >= 1.0 {
		return sdktrace.AlwaysSample()
	}
	return sdktrace.TraceIDRatioBased(rate)
}
