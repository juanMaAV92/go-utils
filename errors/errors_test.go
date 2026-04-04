package errors

import (
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	e := New(http.StatusTeapot, "IM_A_TEAPOT", []string{"short and stout"})
	if e.HttpCode != http.StatusTeapot {
		t.Errorf("HttpCode = %d, want %d", e.HttpCode, http.StatusTeapot)
	}
	if e.Code != "IM_A_TEAPOT" {
		t.Errorf("Code = %q, want \"IM_A_TEAPOT\"", e.Code)
	}
	if len(e.Messages) != 1 || e.Messages[0] != "short and stout" {
		t.Errorf("Messages = %v", e.Messages)
	}
}

func TestErrorInterface(t *testing.T) {
	var err error = ErrNotFound()
	if err == nil {
		t.Fatal("ErrorResponse should implement error")
	}
	if err.Error() == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestPredefinedAreFresh(t *testing.T) {
	a := ErrUnauthorized()
	b := ErrUnauthorized()
	if a == b {
		t.Error("predefined constructors should return different pointers")
	}
	a.Messages = []string{"tampered"}
	if b.Messages[0] == "tampered" {
		t.Error("modifying one instance should not affect another")
	}
}

func TestPredefinedWithCustomMessage(t *testing.T) {
	e := ErrNotFound("user not found")
	if e.Messages[0] != "user not found" {
		t.Errorf("got %q, want \"user not found\"", e.Messages[0])
	}
	if e.HttpCode != http.StatusNotFound {
		t.Errorf("HttpCode = %d, want 404", e.HttpCode)
	}
}

func TestWithMessage(t *testing.T) {
	original := ErrBadRequest()
	cloned := original.WithMessage("email is required")

	if cloned.Messages[0] != "email is required" {
		t.Errorf("WithMessage = %q", cloned.Messages[0])
	}
	if original.Messages[0] == "email is required" {
		t.Error("WithMessage should not modify original")
	}
	if cloned.HttpCode != original.HttpCode {
		t.Error("WithMessage should preserve HttpCode")
	}
}

func TestWithMessages(t *testing.T) {
	e := ErrBadRequest().WithMessages([]string{"field a", "field b"})
	if len(e.Messages) != 2 {
		t.Errorf("WithMessages len = %d, want 2", len(e.Messages))
	}
}

func TestAllPredefinedHaveCorrectStatusCodes(t *testing.T) {
	cases := []struct {
		name string
		err  *ErrorResponse
		code int
	}{
		{"BadRequest", ErrBadRequest(), http.StatusBadRequest},
		{"Unauthorized", ErrUnauthorized(), http.StatusUnauthorized},
		{"Forbidden", ErrForbidden(), http.StatusForbidden},
		{"NotFound", ErrNotFound(), http.StatusNotFound},
		{"MethodNotAllowed", ErrMethodNotAllowed(), http.StatusMethodNotAllowed},
		{"RequestTimeout", ErrRequestTimeout(), http.StatusRequestTimeout},
		{"TooManyRequests", ErrTooManyRequests(), http.StatusTooManyRequests},
		{"RequestEntityTooLarge", ErrRequestEntityTooLarge(), http.StatusRequestEntityTooLarge},
		{"UnsupportedMediaType", ErrUnsupportedMediaType(), http.StatusUnsupportedMediaType},
		{"InternalServer", ErrInternalServer(), http.StatusInternalServerError},
		{"BadGateway", ErrBadGateway(), http.StatusBadGateway},
		{"ServiceUnavailable", ErrServiceUnavailable(), http.StatusServiceUnavailable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.err.HttpCode != c.code {
				t.Errorf("%s: HttpCode = %d, want %d", c.name, c.err.HttpCode, c.code)
			}
			if c.err.Code == "" {
				t.Errorf("%s: Code is empty", c.name)
			}
			if len(c.err.Messages) == 0 {
				t.Errorf("%s: Messages is empty", c.name)
			}
		})
	}
}
