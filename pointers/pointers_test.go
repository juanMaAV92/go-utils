package pointers

import "testing"

func TestPointer(t *testing.T) {
	p := Pointer(42)
	if p == nil || *p != 42 {
		t.Fatalf("Pointer(42) = %v, want 42", p)
	}

	s := Pointer("hello")
	if s == nil || *s != "hello" {
		t.Fatalf("Pointer(\"hello\") = %v, want \"hello\"", s)
	}
}

func TestValue(t *testing.T) {
	n := 99
	if got := Value(&n); got != 99 {
		t.Errorf("Value(&99) = %d, want 99", got)
	}
	if got := Value[int](nil); got != 0 {
		t.Errorf("Value[int](nil) = %d, want 0", got)
	}
	if got := Value[string](nil); got != "" {
		t.Errorf("Value[string](nil) = %q, want \"\"", got)
	}
}

func TestFirstNonNil(t *testing.T) {
	cases := []struct {
		name   string
		inputs []*int
		want   *int
	}{
		{"all nil", []*int{nil, nil}, nil},
		{"first non-nil", []*int{Pointer(1), nil}, Pointer(1)},
		{"second non-nil", []*int{nil, Pointer(2)}, Pointer(2)},
		{"multiple non-nil returns first", []*int{Pointer(1), Pointer(2)}, Pointer(1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FirstNonNil(c.inputs...)
			if c.want == nil && got != nil {
				t.Errorf("want nil, got %d", *got)
			} else if c.want != nil && (got == nil || *got != *c.want) {
				t.Errorf("got %v, want %d", got, *c.want)
			}
		})
	}
}

func TestFirstNonNilOrEmptyString(t *testing.T) {
	cases := []struct {
		name   string
		inputs []*string
		want   *string
	}{
		{"all nil", []*string{nil, nil}, nil},
		{"all empty", []*string{Pointer(""), Pointer("")}, nil},
		{"first non-empty", []*string{Pointer("a"), Pointer("b")}, Pointer("a")},
		{"skip empty then non-empty", []*string{Pointer(""), nil, Pointer("x")}, Pointer("x")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FirstNonNilOrEmptyString(c.inputs...)
			if c.want == nil && got != nil {
				t.Errorf("want nil, got %q", *got)
			} else if c.want != nil && (got == nil || *got != *c.want) {
				t.Errorf("got %v, want %q", got, *c.want)
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	if StringPtr("") != nil {
		t.Error("StringPtr(\"\") should be nil")
	}
	p := StringPtr("val")
	if p == nil || *p != "val" {
		t.Errorf("StringPtr(\"val\") = %v, want \"val\"", p)
	}
}

func TestStringValue(t *testing.T) {
	if StringValue(nil) != "" {
		t.Error("StringValue(nil) should be \"\"")
	}
	s := "hello"
	if StringValue(&s) != "hello" {
		t.Error("StringValue(&\"hello\") should be \"hello\"")
	}
}
