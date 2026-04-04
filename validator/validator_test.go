package validator

import (
	"errors"
	"strings"
	"testing"

	apperrors "github.com/juanmaAV/go-utils/errors"
)

// ---- test structs ----

type userRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=20"`
	Age      int    `json:"age"      validate:"required,gte=18"`
}

type nameRequest struct {
	Name string `json:"name" validate:"required"`
}

type conditionalRequest struct {
	Nature           string `json:"nature"            validate:"required,oneof=INDIVIDUAL ORGANIZATION"`
	IndividualData   string `json:"individual_data"   validate:"required_if=Nature INDIVIDUAL"`
	OrganizationData string `json:"organization_data" validate:"required_if=Nature ORGANIZATION"`
}

type nestedRequest struct {
	Details nameRequest `json:"details"`
}

// ---- helpers ----

func errMessages(t *testing.T, err error) []string {
	t.Helper()
	var e *apperrors.ErrorResponse
	if !errors.As(err, &e) {
		t.Fatalf("expected *ErrorResponse, got %T: %v", err, err)
	}
	return e.Messages
}

func assertContains(t *testing.T, messages []string, substr string) {
	t.Helper()
	for _, m := range messages {
		if strings.Contains(strings.ToLower(m), strings.ToLower(substr)) {
			return
		}
	}
	t.Errorf("no message contains %q, got: %v", substr, messages)
}

func assertHTTPCode(t *testing.T, err error, want int) {
	t.Helper()
	var e *apperrors.ErrorResponse
	if !errors.As(err, &e) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if e.HttpCode != want {
		t.Errorf("HttpCode = %d, want %d", e.HttpCode, want)
	}
}

// ---- Validate tests ----

func TestValidate_Success(t *testing.T) {
	v := New()
	err := v.Validate(userRequest{Email: "user@example.com", Username: "alice", Age: 25})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidate_Nil(t *testing.T) {
	if err := New().Validate(nil); err != nil {
		t.Errorf("expected no error for nil, got: %v", err)
	}
}

func TestValidate_Required(t *testing.T) {
	v := New()
	err := v.Validate(userRequest{Username: "alice", Age: 25})
	assertContains(t, errMessages(t, err), "email is required")
	assertHTTPCode(t, err, 422)
}

func TestValidate_Email(t *testing.T) {
	v := New()
	err := v.Validate(userRequest{Email: "not-an-email", Username: "alice", Age: 25})
	assertContains(t, errMessages(t, err), "must be a valid email")
}

func TestValidate_MinLength(t *testing.T) {
	v := New()
	err := v.Validate(userRequest{Email: "u@e.com", Username: "ab", Age: 25})
	assertContains(t, errMessages(t, err), "at least 3")
}

func TestValidate_MaxLength(t *testing.T) {
	v := New()
	err := v.Validate(userRequest{Email: "u@e.com", Username: "thisusernameistoolong", Age: 25})
	assertContains(t, errMessages(t, err), "at most 20")
}

func TestValidate_Gte(t *testing.T) {
	v := New()
	err := v.Validate(userRequest{Email: "u@e.com", Username: "alice", Age: 15})
	assertContains(t, errMessages(t, err), "greater than or equal to 18")
}

func TestValidate_MultipleErrors(t *testing.T) {
	v := New()
	err := v.Validate(userRequest{Email: "bad", Username: "ab", Age: 5})
	msgs := errMessages(t, err)
	if len(msgs) < 3 {
		t.Errorf("expected at least 3 messages, got %d: %v", len(msgs), msgs)
	}
}

func TestValidate_JsonFieldNames(t *testing.T) {
	v := New()
	err := v.Validate(userRequest{})
	msgs := errMessages(t, err)
	// Must use json tag names, not struct field names
	for _, m := range msgs {
		if strings.Contains(m, "Email") || strings.Contains(m, "Username") || strings.Contains(m, "Age") {
			t.Errorf("message uses struct name instead of json tag: %q", m)
		}
	}
}

func TestValidate_Oneof(t *testing.T) {
	v := New()
	err := v.Validate(conditionalRequest{Nature: "OTHER", IndividualData: "x"})
	assertContains(t, errMessages(t, err), "must be one of")
}

func TestValidate_RequiredIf(t *testing.T) {
	v := New()
	err := v.Validate(conditionalRequest{Nature: "INDIVIDUAL"})
	assertContains(t, errMessages(t, err), "individual_data is required when")
}

func TestValidate_NestedPath(t *testing.T) {
	v := New()
	err := v.Validate(nestedRequest{Details: nameRequest{Name: ""}})
	assertContains(t, errMessages(t, err), "details.name is required")
}

// ---- Slice tests ----

func TestValidate_Slice_Success(t *testing.T) {
	v := New()
	req := []nameRequest{{Name: "Alice"}, {Name: "Bob"}}
	if err := v.Validate(req); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidate_Slice_Failure(t *testing.T) {
	v := New()
	req := []nameRequest{{Name: "Alice"}, {Name: ""}}
	err := v.Validate(req)
	assertContains(t, errMessages(t, err), "[1].name is required")
}

func TestValidate_SlicePointer_Failure(t *testing.T) {
	v := New()
	req := []nameRequest{{Name: "Alice"}, {Name: ""}}
	err := v.Validate(&req)
	assertContains(t, errMessages(t, err), "[1].name is required")
}

func TestValidate_EmptySlice(t *testing.T) {
	v := New()
	if err := v.Validate([]nameRequest{}); err != nil {
		t.Errorf("expected no error for empty slice, got: %v", err)
	}
}

func TestValidate_SliceMultipleErrors(t *testing.T) {
	v := New()
	req := []userRequest{{Email: "bad", Username: "a", Age: 5}}
	err := v.Validate(&req)
	msgs := errMessages(t, err)
	assertContains(t, msgs, "[0].email")
	assertContains(t, msgs, "[0].username")
	assertContains(t, msgs, "[0].age")
}

// ---- Singleton test ----

func TestNew_ReturnsSingleton(t *testing.T) {
	a := New()
	b := New()
	if a != b {
		t.Error("New() should return the same instance")
	}
}

// ---- BindAndValidate tests ----

type mockBinder struct{ err error }

func (m *mockBinder) Bind(v any) error { return m.err }

func TestBindAndValidate_Success(t *testing.T) {
	v := New()
	req := userRequest{Email: "u@e.com", Username: "alice", Age: 25}
	if err := v.BindAndValidate(&mockBinder{}, &req); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestBindAndValidate_BindError(t *testing.T) {
	v := New()
	req := userRequest{}
	err := v.BindAndValidate(&mockBinder{err: errors.New("bad json")}, &req)
	assertContains(t, errMessages(t, err), "invalid request format")
	assertHTTPCode(t, err, 400)
}

func TestBindAndValidate_ValidationError(t *testing.T) {
	v := New()
	req := userRequest{Email: "bad", Username: "ab", Age: 5}
	err := v.BindAndValidate(&mockBinder{}, &req)
	assertHTTPCode(t, err, 422)
}
