package telemetry

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestResolveSampler_Default(t *testing.T) {
	s := resolveSampler(0)
	if s != sdktrace.AlwaysSample() {
		t.Errorf("zero rate should return AlwaysSample, got %v", s)
	}
}

func TestResolveSampler_Full(t *testing.T) {
	s := resolveSampler(1.0)
	if s != sdktrace.AlwaysSample() {
		t.Errorf("rate=1.0 should return AlwaysSample, got %v", s)
	}
}

func TestResolveSampler_Partial(t *testing.T) {
	s := resolveSampler(0.5)
	want := sdktrace.TraceIDRatioBased(0.5)
	if s.Description() != want.Description() {
		t.Errorf("rate=0.5 sampler = %q, want %q", s.Description(), want.Description())
	}
}

func TestInitTelemetry_NoEndpoint(t *testing.T) {
	shutdown, err := InitTelemetry(context.Background(), Config{
		ServiceName: "test-service",
		Environment: "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown func should not be nil")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown returned error: %v", err)
	}
}

func TestInitTelemetry_WithSampleRate(t *testing.T) {
	shutdown, err := InitTelemetry(context.Background(), Config{
		ServiceName: "test-service",
		Environment: "test",
		SampleRate:  0.1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = shutdown(context.Background())
}
