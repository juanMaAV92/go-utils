package logger

import "log/slog"

// Option configures the logger at construction time.
type Option func(*options)

type options struct {
	level slog.Level
}

func applyOptions(opts ...Option) *options {
	o := &options{level: slog.LevelInfo}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithLevel sets the minimum log level.
func WithLevel(l Level) Option {
	return func(o *options) {
		if sl, ok := slogLevel[l]; ok {
			o.level = sl
		}
	}
}
