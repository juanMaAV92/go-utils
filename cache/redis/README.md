# cache/redis

Redis client with JSON serialization, OTel tracing/metrics, Pub/Sub, and atomic set operations.

## Setup

```go
import "github.com/juanmaAV/go-utils/cache/redis"

cfg, err := redis.ConfigFromEnv("REDIS")
if err != nil {
    log.Fatal(err)
}

cache, err := redis.New(ctx, cfg, logger)
if err != nil {
    log.Fatal(err)
}
defer cache.Close()
```

### Config fields

| Field | Env var (prefix=REDIS) | Default |
|---|---|---|
| `Host` | `REDIS_HOST` | required |
| `Port` | `REDIS_PORT` | `6379` |
| `Password` | `REDIS_PASSWORD` | empty |
| `DB` | `REDIS_DB` | `0` |
| `TLS` | `REDIS_TLS` | `false` |
| `TLSServerName` | `REDIS_TLS_SERVER_NAME` | Host |
| `ReadTimeout` | `REDIS_READ_TIMEOUT` | `3s` |
| `WriteTimeout` | `REDIS_WRITE_TIMEOUT` | `3s` |
| `KeyPrefix` | `REDIS_KEY_PREFIX` | empty |

`KeyPrefix` is prepended to every key with `:` as separator — `"orders"` → `orders:your-key`.

Multiple Redis instances use different prefixes:

```go
sessionCache, _ := redis.ConfigFromEnv("SESSION_REDIS")
rateLimitCache, _ := redis.ConfigFromEnv("RATE_LIMIT_REDIS")
```

## Sentinel errors

```go
err := cache.Get(ctx, "key", &dest)
switch {
case errors.Is(err, redis.ErrKeyNotFound): // key does not exist
case errors.Is(err, redis.ErrKeyNotSet):   // NX/XX condition not met
}
```

## Key-value

### Set

Primitives (`string`, `int`, `int64`, `float64`, `bool`, `[]byte`) are stored as-is.
Any other type is JSON-serialized automatically.

```go
// default TTL: 7 days
cache.Set(ctx, "session:abc", sessionData)

// custom TTL
cache.Set(ctx, "otp:123", "998877", redis.WithTTL(5*time.Minute))

// only set if key does not exist (NX)
err := cache.Set(ctx, "lock:job", "1", redis.WithNX(), redis.WithTTL(30*time.Second))
if errors.Is(err, redis.ErrKeyNotSet) {
    // lock already held by another instance
}

// only set if key exists (XX)
cache.Set(ctx, "session:abc", updated, redis.WithXX(), redis.WithKeepTTL())

// persist — no expiry
cache.Set(ctx, "config:flags", flags, redis.WithPersist())
```

### Get

`dest` must be a pointer of the same type that was stored.

```go
var session SessionData
err := cache.Get(ctx, "session:abc", &session)

// atomic get + delete (one-time tokens, OTP)
var otp string
err := cache.Get(ctx, "otp:123", &otp, redis.WithDeleteAfterGet())
```

### GetOrSet

Cache-aside pattern — fetches from cache if available, otherwise calls the producer function, stores the result, and returns it.

```go
var user User
err := cache.GetOrSet(ctx, "user:123", &user, func() (any, error) {
    return db.FindUser(ctx, 123)
}, redis.WithTTL(10*time.Minute))
```

### Delete / Exists

```go
cache.Delete(ctx, "session:abc", "session:xyz") // multiple keys at once

ok, err := cache.Exists(ctx, "session:abc")
```

### Increment

Atomic counter — key is initialized to 0 if it does not exist.
Use negative delta to decrement.

```go
// rate limiting
count, err := cache.Increment(ctx, "rate:user:123", 1)
if count > 100 {
    // reject request
}

// decrement
cache.Increment(ctx, "credits:user:123", -1)
```

### Expire / TTL

```go
// refresh session TTL on activity
cache.Expire(ctx, "session:abc", 30*time.Minute)

// check remaining TTL
ttl, err := cache.TTL(ctx, "session:abc")
// ttl == -1 → no expiry
// ttl == -2 → key does not exist
```

## Sets

Useful for permissions, tags, memberships — unordered collections of unique strings.

```go
// add members (optional TTL)
cache.AddToSet(ctx, "user:123:roles", []string{"admin", "editor"}, redis.WithTTL(time.Hour))

// check membership
ok, _ := cache.ExistsInSet(ctx, "user:123:roles", "admin")

// list all members
roles, _ := cache.GetSetMembers(ctx, "user:123:roles")

// remove — deletes key automatically if set becomes empty
cache.RemoveFromSet(ctx, "user:123:roles", []string{"editor"})
```

## Pub/Sub

For real-time broadcasting: WebSocket fan-out, live notifications, cache invalidation signals.

**Publisher:**
```go
type OrderEvent struct {
    OrderID string `json:"order_id"`
    Status  string `json:"status"`
}

err := cache.Publish(ctx, "orders:events", OrderEvent{
    OrderID: "ord_123",
    Status:  "shipped",
})
```

**Subscriber** (typically in a goroutine):
```go
sub := cache.Subscribe(ctx, "orders:events")
defer sub.Close()

for msg := range sub.Channel() {
    var event OrderEvent
    if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
        continue
    }
    // broadcast to WebSocket clients, trigger side effects, etc.
}
```

**Multi-channel:**
```go
sub := cache.Subscribe(ctx, "orders:events", "inventory:events")
```

> Pub/Sub is fire-and-forget — messages are not persisted. If a subscriber is offline when a message is published, the message is lost. For reliable delivery with persistence, use a message queue (`messaging/sqs`).

## Notes

- All keys are automatically prefixed: `{KeyPrefix}:{key}` — prevents collisions between services sharing the same Redis instance
- `RemoveFromSet` and `AddToSet` with TTL use Lua scripts for atomicity
- OTel tracing and metrics are enabled automatically on connection
