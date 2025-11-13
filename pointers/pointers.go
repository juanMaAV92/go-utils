package pointers

func Pointer[T any](v T) *T {
	return &v
}

func FirstNonNil[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

func FirstNotNilOrEmptyString(values ...*string) *string {
	for _, v := range values {
		if v != nil && *v != "" {
			return v
		}
	}
	return nil
}

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
