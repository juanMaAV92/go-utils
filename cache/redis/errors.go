package redis

import "errors"

// Sentinel errors returned by cache operations.
var (
	// ErrKeyNotFound is returned by Get when the key does not exist in Redis.
	ErrKeyNotFound = errors.New("key not found")

	// ErrKeyNotSet is returned by Set when a conditional write (NX/XX) did not apply.
	ErrKeyNotSet = errors.New("key not set: condition not met")
)
