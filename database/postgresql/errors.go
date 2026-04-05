package postgresql

import "errors"

// Sentinel errors returned by database operations.
// Use errors.Is to check for specific database error conditions.
var (
	// ErrDuplicateRecord is returned when an INSERT violates a unique constraint.
	ErrDuplicateRecord = errors.New("duplicate record")

	// ErrConstraintViolation is returned when a CHECK constraint fails.
	ErrConstraintViolation = errors.New("constraint violation")

	// ErrInvalidReference is returned when a foreign key constraint fails.
	ErrInvalidReference = errors.New("invalid reference")
)
