package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/juanMaAV92/go-utils/logger"
	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
)

const (
	connectionRetries = 3
	retryDelay        = time.Second
	defaultTTL        = 7 * 24 * time.Hour
)

// New creates a Cache backed by a Redis client with OTel tracing and metrics.
// Call once at service startup and inject the returned Cache where needed.
func New(ctx context.Context, cfg Config, log logger.Logger) (Cache, error) {
	client := goredis.NewClient(cfg.toRedisOptions())

	var err error
	for i := 0; i <= connectionRetries; i++ {
		err = client.Ping(ctx).Err()
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("redis: context cancelled during connection: %w", ctx.Err())
		}
		if i < connectionRetries {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("redis: context cancelled during retry: %w", ctx.Err())
			case <-time.After(retryDelay):
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("redis: failed to connect after %d attempts: %w", connectionRetries+1, err)
	}

	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, fmt.Errorf("redis: failed to enable OTel tracing: %w", err)
	}
	if err := redisotel.InstrumentMetrics(client); err != nil {
		return nil, fmt.Errorf("redis: failed to enable OTel metrics: %w", err)
	}

	return &cache{
		keyPrefix: cfg.KeyPrefix,
		instance:  client,
		logger:    log,
	}, nil
}
