package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/juanMaAV92/go-utils/log"
	"github.com/redis/go-redis/v9"
)

const (
	connectionRetries = 3
	retryDelay        = time.Second
	defaultTTL        = 24 * 7 * time.Hour
)

var (
	singleton    *Cache
	onceCache    sync.Once
	singletonErr error
)

func New(cfg CacheConfig, logger log.Logger) (*Cache, error) {
	onceCache.Do(func() {
		client, err := connect(cfg)
		if err != nil {
			singletonErr = fmt.Errorf("failed to create redis connection: %v", err)
			return
		}
		singleton = &Cache{
			serviceCache: cfg.ServerName,
			instance:     client,
			logger:       logger,
		}
	})

	return singleton, singletonErr
}

func connect(cfg CacheConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	var err error
	for i := 0; i <= connectionRetries; i++ {
		err = client.Ping(ctx).Err()
		if err == nil {
			break
		}
		if i < connectionRetries {
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open cache connection after %d attempts: %w", connectionRetries, err)
	}

	return client, nil
}
