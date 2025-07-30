package pointers

import (
	"testing"
)

func Test_Pointer(t *testing.T) {
	ptr := Pointer(1)
	if *ptr != 1 {
		t.Errorf("Pointer(1) = %v; want 1", *ptr)
	}
}

func Test_FirstNonNil(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []*int
		expected *int
	}{
		{
			name:     "All nil pointers",
			inputs:   []*int{nil, nil, nil},
			expected: nil,
		},
		{
			name:     "First non-nil pointer",
			inputs:   []*int{Pointer(10), nil, nil},
			expected: Pointer(10),
		},
		{
			name:     "Second non-nil pointer",
			inputs:   []*int{nil, Pointer(20), nil},
			expected: Pointer(20),
		},
		{
			name:     "Multiple non-nil pointers",
			inputs:   []*int{Pointer(10), Pointer(20), Pointer(30)},
			expected: Pointer(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstNonNil(tt.inputs...)
			if (result == nil && tt.expected != nil) || (result != nil && tt.expected == nil) || (result != nil && tt.expected != nil && *result != *tt.expected) {
				t.Errorf("FirstNonNil(%v) = %v; want %v", tt.inputs, result, tt.expected)
			}
		})
	}
}

func Test_FirstNotNilOrEmptyString(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []*string
		expected *string
	}{
		{
			name:     "All nil pointers",
			inputs:   []*string{nil, nil, nil},
			expected: nil,
		},
		{
			name:     "All empty strings",
			inputs:   []*string{Pointer(""), Pointer(""), Pointer("")},
			expected: nil,
		},
		{
			name:     "First non-empty string",
			inputs:   []*string{Pointer("hello"), nil, Pointer("")},
			expected: Pointer("hello"),
		},
		{
			name:     "Second non-empty string",
			inputs:   []*string{nil, Pointer("world"), Pointer("")},
			expected: Pointer("world"),
		},
		{
			name:     "Multiple non-empty strings",
			inputs:   []*string{Pointer("foo"), Pointer("bar"), Pointer("baz")},
			expected: Pointer("foo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstNotNilOrEmptyString(tt.inputs...)
			if (result == nil && tt.expected != nil) || (result != nil && tt.expected == nil) || (result != nil && tt.expected != nil && *result != *tt.expected) {
				t.Errorf("FirstNotNilOrEmptyString(%v) = %v; want %v", tt.inputs, result, tt.expected)
			}
		})
	}
}
