package redis

import (
	"context"
	"time"

	"github.com/juanMaAV92/go-utils/logger"
	goredis "github.com/redis/go-redis/v9"
)

// Cache is the interface for all cache operations.
type Cache interface {
	// Key-value
	Set(ctx context.Context, key string, value any, opts ...SetOption) error
	Get(ctx context.Context, key string, dest any, opts ...GetOption) error
	GetOrSet(ctx context.Context, key string, dest any, fn func() (any, error), opts ...SetOption) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// TTL management
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Sets
	AddToSet(ctx context.Context, key string, members []string, opts ...SetOption) error
	RemoveFromSet(ctx context.Context, key string, members []string) error
	GetSetMembers(ctx context.Context, key string) ([]string, error)
	ExistsInSet(ctx context.Context, key, member string) (bool, error)

	// Pub/Sub
	Publish(ctx context.Context, channel string, message any) error
	Subscribe(ctx context.Context, channels ...string) *goredis.PubSub

	// Lifecycle
	Close() error
}

type cache struct {
	keyPrefix string
	instance  *goredis.Client
	logger    logger.Logger
}
