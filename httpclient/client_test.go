package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer spins up a local HTTP server that responds with the given
// status code and JSON body for any request.
func newTestServer(t *testing.T, statusCode int, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestGet_Success(t *testing.T) {
	srv := newTestServer(t, http.StatusOK, map[string]string{"status": "ok"})
	defer srv.Close()

	c := New(nil, WithBaseURL(srv.URL))
	resp, err := c.Get(context.Background(), "/health")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestPost_Success(t *testing.T) {
	srv := newTestServer(t, http.StatusCreated, map[string]string{"id": "123"})
	defer srv.Close()

	c := New(nil, WithBaseURL(srv.URL))
	resp, err := c.Post(context.Background(), "/items", map[string]string{"name": "test"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
	}
}

func TestGet_NonSuccess_ReturnsError(t *testing.T) {
	srv := newTestServer(t, http.StatusNotFound, map[string]string{"error": "not found"})
	defer srv.Close()

	c := New(nil, WithBaseURL(srv.URL))
	resp, err := c.Get(context.Background(), "/missing")

	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if resp == nil {
		t.Fatal("expected non-nil Response even on error")
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", resp.StatusCode)
	}
	if resp.Success {
		t.Error("expected Success=false")
	}
}

func TestResponse_JSON(t *testing.T) {
	srv := newTestServer(t, http.StatusOK, map[string]string{"key": "value"})
	defer srv.Close()

	c := New(nil, WithBaseURL(srv.URL))
	resp, err := c.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := resp.JSON(&result); err != nil {
		t.Fatalf("JSON decode error: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("key = %q, want \"value\"", result["key"])
	}
}

func TestNilContext_ReturnsError(t *testing.T) {
	c := New(nil, WithBaseURL("http://localhost"))
	//nolint:staticcheck
	_, err := c.Get(nil, "/")
	if err == nil {
		t.Error("expected error for nil context")
	}
}

func TestWithQueryParams(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(nil, WithBaseURL(srv.URL))
	_, _ = c.Get(context.Background(), "/search",
		WithQueryParam("q", "test"),
		WithQueryParam("page", "1"),
	)

	if capturedQuery == "" {
		t.Error("expected query params to be sent")
	}
}

func TestWithAuthToken(t *testing.T) {
	var capturedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(nil, WithBaseURL(srv.URL))
	_, _ = c.Get(context.Background(), "/secure", WithAuthToken("my-token"))

	if capturedAuth != "Bearer my-token" {
		t.Errorf("Authorization = %q, want \"Bearer my-token\"", capturedAuth)
	}
}

func TestWithFormData(t *testing.T) {
	var capturedContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(nil, WithBaseURL(srv.URL))
	_, _ = c.Post(context.Background(), "/form", nil,
		WithFormData(map[string]string{"field": "value"}),
	)

	if capturedContentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", capturedContentType)
	}
}

func TestNew_ReturnsSameInterface(t *testing.T) {
	c := New(nil)
	if c == nil {
		t.Error("New() should not return nil")
	}
}
