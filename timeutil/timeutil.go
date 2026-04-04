package timeutil

import "time"

const (
	DateLayout     = "2006-01-02"
	DateTimeLayout = time.RFC3339
)

// ParseDate parses a "2006-01-02" string. Returns nil if value is empty.
func ParseDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(DateLayout, value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ParseDateTime parses an RFC3339 string. Returns nil if value is empty.
func ParseDateTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(DateTimeLayout, value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// IsZeroOrNil reports whether t is nil or the zero time.
func IsZeroOrNil(t *time.Time) bool {
	return t == nil || t.IsZero()
}

// After reports whether t is after u. Returns false if t is nil.
func After(t *time.Time, u time.Time) bool {
	return t != nil && t.After(u)
}

// Before reports whether t is before u. Returns false if t is nil.
func Before(t *time.Time, u time.Time) bool {
	return t != nil && t.Before(u)
}

// StartOfDay returns midnight (00:00:00) of t in t's location.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns 23:59:59.999999999 of t in t's location.
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}
