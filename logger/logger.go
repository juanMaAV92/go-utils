package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// Level represents the minimum severity to emit.
type Level uint8

const (
	FatalLevel Level = iota
	ErrorLevel
	WarningLevel
	InfoLevel
	DebugLevel
)

// slogLevel maps our Level to slog.Level.
// Fatal uses a custom level above Error so it passes any level filter.
var slogLevel = map[Level]slog.Level{
	FatalLevel:   slog.Level(12),
	ErrorLevel:   slog.LevelError,
	WarningLevel: slog.LevelWarn,
	InfoLevel:    slog.LevelInfo,
	DebugLevel:   slog.LevelDebug,
}

// Logger is the interface all consumers depend on.
// Fatal logs at a critical level and then calls os.Exit(1).
type Logger interface {
	Fatal(ctx context.Context, step, message string, args ...any)
	Error(ctx context.Context, step, message string, args ...any)
	Warning(ctx context.Context, step, message string, args ...any)
	Info(ctx context.Context, step, message string, args ...any)
	Debug(ctx context.Context, step, message string, args ...any)
}

type logger struct {
	sl *slog.Logger
}

// New creates a Logger. serviceName is attached to every log entry.
// environment controls output: "local" → text (human-readable), anything else → JSON.
func New(serviceName, environment string, opts ...Option) Logger {
	cfg := applyOptions(opts...)
	handler := buildHandler(cfg.level, environment)
	sl := slog.New(handler).With("service", serviceName)
	return &logger{sl: sl}
}

func buildHandler(level slog.Level, environment string) slog.Handler {
	return buildHandlerTo(os.Stdout, level, environment)
}

func buildHandlerTo(w io.Writer, level slog.Level, environment string) slog.Handler {
	hopts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Rename "msg" → "message" for consistency across backends
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	}

	otelWrap := func(h slog.Handler) slog.Handler {
		return &otelHandler{Handler: h}
	}

	if environment == "local" {
		return otelWrap(slog.NewTextHandler(w, hopts))
	}
	return otelWrap(slog.NewJSONHandler(w, hopts))
}

// newWithWriter creates a Logger that writes to w. Used in tests.
func newWithWriter(w io.Writer, serviceName string, opts ...Option) Logger {
	cfg := applyOptions(opts...)
	handler := buildHandlerTo(w, cfg.level, "production")
	sl := slog.New(handler).With("service", serviceName)
	return &logger{sl: sl}
}

// otelHandler wraps any slog.Handler and injects OTel trace_id/span_id
// from the context into every log record when a valid span is active.
type otelHandler struct {
	slog.Handler
}

func (h *otelHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (l *logger) Fatal(ctx context.Context, step, message string, args ...any) {
	l.sl.Log(ctx, slog.Level(12), message, buildArgs(step, args)...)
	os.Exit(1)
}

func (l *logger) Error(ctx context.Context, step, message string, args ...any) {
	l.sl.ErrorContext(ctx, message, buildArgs(step, args)...)
}

func (l *logger) Warning(ctx context.Context, step, message string, args ...any) {
	l.sl.WarnContext(ctx, message, buildArgs(step, args)...)
}

func (l *logger) Info(ctx context.Context, step, message string, args ...any) {
	l.sl.InfoContext(ctx, message, buildArgs(step, args)...)
}

func (l *logger) Debug(ctx context.Context, step, message string, args ...any) {
	l.sl.DebugContext(ctx, message, buildArgs(step, args)...)
}

// buildArgs prepends "step" to args so it appears as a flat field in every record.
// args must be key-value pairs. Non-string keys and unpaired keys are skipped by slog.
func buildArgs(step string, args []any) []any {
	out := make([]any, 0, len(args)+2)
	out = append(out, "step", step)
	out = append(out, args...)
	return out
}
