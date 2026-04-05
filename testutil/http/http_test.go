package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httptest2 "github.com/juanmaAV/go-utils/testutil/http"
)

func TestNewRequest_Method(t *testing.T) {
	req := httptest2.NewRequest(http.MethodDelete, "/items/1", nil)
	if req.Method != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("expected Content-Type application/json")
	}
}

func TestNewRequest_StringBody(t *testing.T) {
	req := httptest2.NewRequest(http.MethodPost, "/items", `{"name":"test"}`)
	if req.Body == nil {
		t.Fatal("expected non-nil body")
	}
}

func TestNewRequest_StructBody(t *testing.T) {
	req := httptest2.NewRequest(http.MethodPost, "/items", map[string]string{"name": "test"})
	if req.Body == nil {
		t.Fatal("expected non-nil body")
	}
}

func TestAssertStatus_Pass(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)
	// should not fail
	httptest2.AssertStatus(t, rec, http.StatusOK)
}

func TestAssertStatus_Fail(t *testing.T) {
	inner := &testing.T{}
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusNotFound)
	httptest2.AssertStatus(inner, rec, http.StatusOK)
	if !inner.Failed() {
		t.Error("expected inner test to fail")
	}
}

func TestDecodeJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/json")
	_, _ = rec.WriteString(`{"name":"alice","age":30}`)

	var out struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	httptest2.DecodeJSON(t, rec, &out)
	if out.Name != "alice" {
		t.Errorf("expected name=alice, got %q", out.Name)
	}
	if out.Age != 30 {
		t.Errorf("expected age=30, got %d", out.Age)
	}
}

func TestAssertJSONField_Pass(t *testing.T) {
	rec := httptest.NewRecorder()
	_, _ = rec.WriteString(`{"status":"ok","message":"created"}`)
	httptest2.AssertJSONField(t, rec, "status", "ok")
}

func TestAssertJSONField_WrongValue(t *testing.T) {
	inner := &testing.T{}
	rec := httptest.NewRecorder()
	_, _ = rec.WriteString(`{"status":"error"}`)
	httptest2.AssertJSONField(inner, rec, "status", "ok")
	if !inner.Failed() {
		t.Error("expected inner test to fail")
	}
}

func TestAssertJSONField_MissingKey(t *testing.T) {
	inner := &testing.T{}
	rec := httptest.NewRecorder()
	rec.WriteString(`{"other":"value"}`)
	httptest2.AssertJSONField(inner, rec, "status", "ok")
	if !inner.Failed() {
		t.Error("expected inner test to fail for missing key")
	}
}

func TestMustJSON(t *testing.T) {
	got := httptest2.MustJSON(map[string]int{"count": 3})
	if got != `{"count":3}` {
		t.Errorf("unexpected JSON: %s", got)
	}
}

func TestMustJSON_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unmarshalable value")
		}
	}()
	httptest2.MustJSON(make(chan int))
}
