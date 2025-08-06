package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"errors"

	"github.com/juanMaAV92/go-utils/log"
	"github.com/redis/go-redis/v9"
)

const (
	// Steps
	setStep    = "setting cache value"
	getStep    = "getting cache value"
	deleteStep = "deleting cache value"
	existsStep = "checking cache key existence"

	// Error Messages
	errContextNil      = "context cannot be nil"
	errKeyEmpty        = "key cannot be empty"
	errValueNil        = "value cannot be nil"
	errSerialization   = "Failed to serialize value to JSON"
	errDeserialization = "failed to deserialize value"
	errGenericCache    = "failed cache"
)

type StatusCommand = redis.StatusCmd
type StringCommand = redis.StringCmd
type IntCommand = redis.IntCmd
type BoolCommand = redis.BoolCmd

type Redis interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *StatusCommand
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *BoolCommand
	SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) *BoolCommand
	Get(ctx context.Context, key string) *StringCommand
	GetDel(ctx context.Context, key string) *StringCommand
	Exists(ctx context.Context, keys ...string) *IntCommand
	Del(ctx context.Context, keys ...string) *IntCommand
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, opts ...SetOption) error {
	if ctx == nil {
		return errors.New(errContextNil)
	}

	if key == "" {
		return errors.New(errKeyEmpty)
	}

	if value == nil {
		return errors.New(errValueNil)
	}

	config := &SetOptions{}
	for _, opt := range opts {
		opt(config)
	}

	jsonData, errMars := json.Marshal(value)
	if errMars != nil {
		return errors.New(errSerialization)
	}

	ttl := validateTTL(config.TTL)

	var err error
	if config.IfNotExist {
		err = c.instance.SetNX(ctx, c.buildKey(key), jsonData, ttl).Err()
	} else if config.IfExist {
		err = c.instance.SetXX(ctx, c.buildKey(key), jsonData, ttl).Err()
	} else {
		err = c.instance.Set(ctx, c.buildKey(key), jsonData, ttl).Err()
	}

	if err != nil {
		return handleCacheError(ctx, c.logger, err, setStep, "Error setting cache value")
	}

	return nil
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}, options ...GetOptions) (bool, error) {
	if ctx == nil {
		return false, errors.New(errContextNil)
	}

	if key == "" {
		return false, errors.New(errKeyEmpty)
	}

	if dest == nil {
		return false, errors.New(errValueNil)
	}

	var opts GetOptions
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	var err error

	if opts.DeleteAfterGet {
		value, err = c.instance.GetDel(ctx, c.buildKey(key)).Result()
	} else {
		value, err = c.instance.Get(ctx, c.buildKey(key)).Result()
	}

	if err != nil {
		return false, handleCacheError(ctx, c.logger, err, getStep, "Error getting cache value")
	}

	if errUnmar := json.Unmarshal([]byte(value), dest); errUnmar != nil {
		return false, errors.New(errDeserialization)
	}

	return true, nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	if ctx == nil {
		return errors.New(errContextNil)
	}

	if key == "" {
		return errors.New(errKeyEmpty)
	}

	_, err := c.instance.Del(ctx, c.buildKey(key)).Result()
	if err != nil {
		return handleCacheError(ctx, c.logger, err, deleteStep, "Error deleting cache key")
	}

	return nil
}

func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errors.New(errContextNil)
	}

	if key == "" {
		return false, errors.New(errKeyEmpty)
	}

	result, err := c.instance.Exists(ctx, c.buildKey(key)).Result()
	if err != nil {
		return false, handleCacheError(ctx, c.logger, err, existsStep, "Error checking key existence")
	}

	return result > 0, nil
}

func validateTTL(ttl time.Duration) time.Duration {
	if ttl == 0 || ttl > defaultTTL {
		return defaultTTL
	}
	return ttl
}

func handleCacheError(ctx context.Context, logger log.Logger, err error, step, message string) error {
	if err == redis.Nil {
		return nil
	}

	logger.Error(ctx, step, message, log.Field("error", err.Error()))
	return errors.New(errGenericCache)
}

func (c *Cache) buildKey(key string) string {
	return fmt.Sprint(c.serviceCache, "_", key)
}
