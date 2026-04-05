package redis

import "time"

type setOptions struct {
	TTL        time.Duration
	IfNotExist bool // NX — only set if key does not exist
	IfExist    bool // XX — only set if key already exists
	KeepTTL    bool // KEEPTTL — retain current TTL (Redis 6.0+)
}

type getOptions struct {
	DeleteAfterGet bool // GETDEL — atomic get-and-delete
}

// SetOption configures a Set or AddToSet call.
type SetOption func(*setOptions)

// GetOption configures a Get call.
type GetOption func(*getOptions)

// WithTTL sets the expiry duration for the key.
func WithTTL(ttl time.Duration) SetOption {
	return func(o *setOptions) { o.TTL = ttl }
}

// WithPersist removes expiry — the key will never expire.
func WithPersist() SetOption {
	return func(o *setOptions) { o.TTL = 0 }
}

// WithNX only writes if the key does not already exist (SET NX).
func WithNX() SetOption {
	return func(o *setOptions) { o.IfNotExist = true }
}

// WithXX only writes if the key already exists (SET XX).
func WithXX() SetOption {
	return func(o *setOptions) { o.IfExist = true }
}

// WithKeepTTL retains the existing TTL when updating a key (Redis 6.0+).
func WithKeepTTL() SetOption {
	return func(o *setOptions) { o.KeepTTL = true }
}

// WithDeleteAfterGet performs an atomic GETDEL — get the value and delete the key.
func WithDeleteAfterGet() GetOption {
	return func(o *getOptions) { o.DeleteAfterGet = true }
}
