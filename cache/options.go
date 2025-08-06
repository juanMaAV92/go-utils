package cache

import "time"

type SetOptions struct {
	TTL        time.Duration
	IfNotExist bool
	IfExist    bool
}

type GetOptions struct {
	DeleteAfterGet bool
}

type SetOption func(*SetOptions)
type GetOption func(*GetOptions)

// WithTTL sets the TTL for the cache entry
func WithTTL(ttl time.Duration) SetOption {
	return func(opts *SetOptions) {
		opts.TTL = ttl
	}
}

// WithIfNotExist sets the NX option (only set if key doesn't exist)
func WithIfNotExist(value bool) SetOption {
	return func(opts *SetOptions) {
		opts.IfNotExist = value
	}
}

// WithIfExist sets the XX option (only set if key exists)
func WithIfExist(value bool) SetOption {
	return func(opts *SetOptions) {
		opts.IfExist = value
	}
}

// WithDeleteAfterGet sets the DeleteAfterGet option for GetOptions
func WithDeleteAfterGet(value bool) GetOption {
	return func(opts *GetOptions) {
		opts.DeleteAfterGet = value
	}
}
