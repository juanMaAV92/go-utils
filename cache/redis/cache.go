package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Set stores value under key. Structs and slices are JSON-serialized automatically.
// Primitives (string, int, int64, float64, bool, []byte) are stored as-is.
// Default TTL is 7 days — override with WithTTL.
func (c *cache) Set(ctx context.Context, key string, value any, opts ...SetOption) error {
	if err := validateKV(ctx, key, value); err != nil {
		return err
	}

	o := &setOptions{TTL: defaultTTL}
	for _, opt := range opts {
		opt(o)
	}
	if o.IfNotExist && o.IfExist {
		return errors.New("redis: WithNX and WithXX are mutually exclusive")
	}

	payload, err := serialize(value)
	if err != nil {
		return err
	}

	args := goredis.SetArgs{TTL: o.TTL}
	switch {
	case o.IfNotExist:
		args.Mode = "NX"
	case o.IfExist:
		args.Mode = "XX"
	}
	if o.KeepTTL {
		args.KeepTTL = true
	}

	if err := c.instance.SetArgs(ctx, c.key(key), payload, args).Err(); err != nil {
		if err == goredis.Nil {
			return ErrKeyNotSet
		}
		return c.logErr(ctx, "cache.set", err)
	}
	return nil
}

// Get retrieves the value stored under key and deserializes it into dest (must be a pointer).
// Returns ErrKeyNotFound when the key does not exist.
func (c *cache) Get(ctx context.Context, key string, dest any, opts ...GetOption) error {
	if err := validateDest(ctx, key, dest); err != nil {
		return err
	}

	o := &getOptions{}
	for _, opt := range opts {
		opt(o)
	}

	var cmd *goredis.StringCmd
	if o.DeleteAfterGet {
		cmd = c.instance.GetDel(ctx, c.key(key))
	} else {
		cmd = c.instance.Get(ctx, c.key(key))
	}

	val, err := cmd.Result()
	if err != nil {
		if err == goredis.Nil {
			return ErrKeyNotFound
		}
		return c.logErr(ctx, "cache.get", err)
	}

	return deserialize(val, dest)
}

// GetOrSet returns the cached value if the key exists, otherwise calls fn,
// stores the result, and returns it. Uses a single round-trip when the key exists.
func (c *cache) GetOrSet(ctx context.Context, key string, dest any, fn func() (any, error), opts ...SetOption) error {
	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrKeyNotFound) {
		return err
	}

	value, err := fn()
	if err != nil {
		return fmt.Errorf("redis: GetOrSet producer failed: %w", err)
	}

	if setErr := c.Set(ctx, key, value, opts...); setErr != nil {
		return setErr
	}

	return deserialize(mustSerializeString(value), dest)
}

// Delete removes one or more keys. Missing keys are silently ignored.
func (c *cache) Delete(ctx context.Context, keys ...string) error {
	if ctx == nil {
		return errors.New("redis: context is required")
	}
	if len(keys) == 0 {
		return nil
	}
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = c.key(k)
	}
	if err := c.instance.Del(ctx, prefixed...).Err(); err != nil {
		return c.logErr(ctx, "cache.delete", err)
	}
	return nil
}

// Exists reports whether key is present in Redis.
func (c *cache) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(ctx, key); err != nil {
		return false, err
	}
	count, err := c.instance.Exists(ctx, c.key(key)).Result()
	if err != nil {
		return false, c.logErr(ctx, "cache.exists", err)
	}
	return count > 0, nil
}

// Increment atomically adds delta to the integer stored at key (INCRBY).
// If key does not exist it is initialized to 0 before the operation.
// Use negative delta to decrement.
func (c *cache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	if err := validateKey(ctx, key); err != nil {
		return 0, err
	}
	result, err := c.instance.IncrBy(ctx, c.key(key), delta).Result()
	if err != nil {
		return 0, c.logErr(ctx, "cache.increment", err)
	}
	return result, nil
}

// Expire sets or updates the TTL of an existing key.
// Returns no error if the key does not exist (Redis EXPIRE behavior).
func (c *cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := validateKey(ctx, key); err != nil {
		return err
	}
	if err := c.instance.Expire(ctx, c.key(key), ttl).Err(); err != nil {
		return c.logErr(ctx, "cache.expire", err)
	}
	return nil
}

// TTL returns the remaining time-to-live of a key.
// Returns -1 if the key has no expiry, -2 if the key does not exist.
func (c *cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	if err := validateKey(ctx, key); err != nil {
		return 0, err
	}
	ttl, err := c.instance.TTL(ctx, c.key(key)).Result()
	if err != nil {
		return 0, c.logErr(ctx, "cache.ttl", err)
	}
	return ttl, nil
}

// --- Set operations ---

// AddToSet adds members to the Redis set stored at key.
// If opts include WithTTL, the TTL is applied atomically via a Lua script.
func (c *cache) AddToSet(ctx context.Context, key string, members []string, opts ...SetOption) error {
	if err := validateKey(ctx, key); err != nil {
		return err
	}
	if len(members) == 0 {
		return errors.New("redis: at least one member is required")
	}

	o := &setOptions{TTL: 0}
	for _, opt := range opts {
		opt(o)
	}

	if o.TTL <= 0 {
		// No TTL — simple SADD
		if err := c.instance.SAdd(ctx, c.key(key), toAny(members)...).Err(); err != nil {
			return c.logErr(ctx, "cache.set_add", err)
		}
		return nil
	}

	// Atomic SADD + EXPIRE via Lua
	const luaAddWithTTL = `
		redis.call('SADD', KEYS[1], unpack(ARGV, 2))
		redis.call('EXPIRE', KEYS[1], ARGV[1])
		return 1
	`
	args := make([]any, 0, len(members)+1)
	args = append(args, int64(o.TTL.Seconds()))
	for _, m := range members {
		args = append(args, m)
	}
	if err := c.instance.Eval(ctx, luaAddWithTTL, []string{c.key(key)}, args...).Err(); err != nil {
		return c.logErr(ctx, "cache.set_add", err)
	}
	return nil
}

// RemoveFromSet removes members from the set. Deletes the key if the set becomes empty.
func (c *cache) RemoveFromSet(ctx context.Context, key string, members []string) error {
	if err := validateKey(ctx, key); err != nil {
		return err
	}
	if len(members) == 0 {
		return errors.New("redis: at least one member is required")
	}

	const luaRemove = `
		redis.call('SREM', KEYS[1], unpack(ARGV))
		if redis.call('SCARD', KEYS[1]) == 0 then
			redis.call('DEL', KEYS[1])
		end
		return 1
	`
	if err := c.instance.Eval(ctx, luaRemove, []string{c.key(key)}, toAny(members)...).Err(); err != nil {
		return c.logErr(ctx, "cache.set_remove", err)
	}
	return nil
}

// GetSetMembers returns all members of the set. Returns an empty slice if the key does not exist.
func (c *cache) GetSetMembers(ctx context.Context, key string) ([]string, error) {
	if err := validateKey(ctx, key); err != nil {
		return nil, err
	}
	members, err := c.instance.SMembers(ctx, c.key(key)).Result()
	if err != nil {
		if err == goredis.Nil {
			return []string{}, nil
		}
		return nil, c.logErr(ctx, "cache.set_members", err)
	}
	return members, nil
}

// ExistsInSet reports whether member belongs to the set stored at key.
func (c *cache) ExistsInSet(ctx context.Context, key, member string) (bool, error) {
	if err := validateKey(ctx, key); err != nil {
		return false, err
	}
	if member == "" {
		return false, errors.New("redis: member is required")
	}
	exists, err := c.instance.SIsMember(ctx, c.key(key), member).Result()
	if err != nil {
		if err == goredis.Nil {
			return false, nil
		}
		return false, c.logErr(ctx, "cache.set_exists", err)
	}
	return exists, nil
}

// --- Pub/Sub ---

// Publish sends message to channel. message is JSON-serialized if not a primitive.
func (c *cache) Publish(ctx context.Context, channel string, message any) error {
	if ctx == nil {
		return errors.New("redis: context is required")
	}
	if channel == "" {
		return errors.New("redis: channel is required")
	}
	payload, err := serialize(message)
	if err != nil {
		return err
	}
	if err := c.instance.Publish(ctx, channel, payload).Err(); err != nil {
		return c.logErr(ctx, "cache.publish", err)
	}
	return nil
}

// Subscribe returns a *goredis.PubSub subscription to the given channels.
// The caller is responsible for closing the subscription via sub.Close().
//
//	sub := cache.Subscribe(ctx, "notifications:user:123")
//	defer sub.Close()
//	for msg := range sub.Channel() {
//	    // handle msg.Payload
//	}
func (c *cache) Subscribe(ctx context.Context, channels ...string) *goredis.PubSub {
	return c.instance.Subscribe(ctx, channels...)
}

// --- Lifecycle ---

// Close closes the underlying Redis connection.
func (c *cache) Close() error {
	return c.instance.Close()
}

// --- internal helpers ---

func (c *cache) key(k string) string {
	if c.keyPrefix == "" {
		return k
	}
	return c.keyPrefix + ":" + k
}

func (c *cache) logErr(ctx context.Context, step string, err error) error {
	if c.logger != nil {
		c.logger.Error(ctx, step, err.Error())
	}
	return err
}

func serialize(value any) (any, error) {
	switch v := value.(type) {
	case string, int, int64, float64, bool, []byte:
		return v, nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("redis: failed to serialize value: %w", err)
		}
		return b, nil
	}
}

func deserialize(val string, dest any) error {
	switch d := dest.(type) {
	case *string:
		*d = val
		return nil
	case *[]byte:
		*d = []byte(val)
		return nil
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("redis: failed to deserialize value: %w", err)
	}
	return nil
}

// mustSerializeString serializes value to its string representation for GetOrSet.
func mustSerializeString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func toAny(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func validateKey(ctx context.Context, key string) error {
	if ctx == nil {
		return errors.New("redis: context is required")
	}
	if key == "" {
		return errors.New("redis: key is required")
	}
	return nil
}

func validateKV(ctx context.Context, key string, value any) error {
	if err := validateKey(ctx, key); err != nil {
		return err
	}
	if value == nil {
		return errors.New("redis: value is required")
	}
	return nil
}

func validateDest(ctx context.Context, key string, dest any) error {
	if err := validateKey(ctx, key); err != nil {
		return err
	}
	if dest == nil || reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return errors.New("redis: dest must be a non-nil pointer")
	}
	return nil
}
