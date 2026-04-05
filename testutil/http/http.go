// Package http provides framework-agnostic helpers for HTTP handler tests.
package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// NewRequest builds an *http.Request for use in handler tests.
// body may be a string, []byte, or any JSON-serialisable value; pass nil for no body.
// Content-Type is always set to application/json.
func NewRequest(method, rawURL string, body any) *http.Request {
	var r io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			r = strings.NewReader(v)
		case []byte:
			r = strings.NewReader(string(v))
		default:
			b, _ := json.Marshal(v)
			r = strings.NewReader(string(b))
		}
	} else {
		r = strings.NewReader("")
	}
	req := httptest.NewRequest(method, rawURL, r)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// AssertStatus fails the test if the recorder's status code differs from expected.
func AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if rec.Code != expected {
		t.Errorf("status: got %d, want %d\nbody: %s", rec.Code, expected, rec.Body.String())
	}
}

// DecodeJSON decodes the recorder's response body into v.
// Fails the test if decoding fails.
func DecodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("DecodeJSON: %v\nbody: %s", err, rec.Body.String())
	}
}

// AssertJSONField decodes the response body as a JSON object and checks that
// the top-level string field key equals want.
func AssertJSONField(t *testing.T, rec *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	var m map[string]any
	body := rec.Body.String()
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("AssertJSONField: failed to decode body: %v\nbody: %s", err, body)
	}
	got, ok := m[key]
	if !ok {
		t.Errorf("AssertJSONField: key %q not found in body: %s", key, body)
		return
	}
	if got != want {
		t.Errorf("AssertJSONField[%q]: got %q, want %q", key, got, want)
	}
}

// MustJSON serialises v to a JSON string. Panics on error (safe for test setup only).
func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("testutil/http: MustJSON: " + err.Error())
	}
	return string(b)
}
