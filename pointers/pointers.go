package pointers

// Pointer returns a pointer to v.
func Pointer[T any](v T) *T {
	return &v
}

// Value dereferences p. Returns zero if p is nil.
func Value[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// FirstNonNil returns the first non-nil pointer in values, or nil if all are nil.
func FirstNonNil[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

// FirstNonNilOrEmptyString returns the first non-nil, non-empty string pointer, or nil.
func FirstNonNilOrEmptyString(values ...*string) *string {
	for _, v := range values {
		if v != nil && *v != "" {
			return v
		}
	}
	return nil
}

// StringPtr returns a pointer to s, or nil if s is empty.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringValue dereferences s. Returns "" if nil.
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
