package echo_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	echotest "github.com/juanMaAV92/go-utils/testutil/echo"
)

func TestPrepareContext_Basic(t *testing.T) {
	e := echo.New()
	tc := echotest.Case{
		Request: echotest.Request{
			Method: http.MethodGet,
			Url:    "/users",
		},
	}
	ctx, rec := echotest.PrepareContext(e, tc)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if rec == nil {
		t.Fatal("expected non-nil recorder")
	}
	if ctx.Request().Header.Get(echo.HeaderContentType) != echo.MIMEApplicationJSON {
		t.Error("expected Content-Type application/json")
	}
}

func TestPrepareContext_PathParams(t *testing.T) {
	e := echo.New()
	tc := echotest.Case{
		Request: echotest.Request{
			Method: http.MethodGet,
			Url:    "/users/123",
			PathParam: []echotest.Param{
				{Name: "id", Value: "123"},
			},
		},
	}
	ctx, _ := echotest.PrepareContext(e, tc)
	if ctx.Param("id") != "123" {
		t.Errorf("expected id=123, got %q", ctx.Param("id"))
	}
}

func TestPrepareContext_QueryParams(t *testing.T) {
	e := echo.New()
	tc := echotest.Case{
		Request: echotest.Request{
			Method: http.MethodGet,
			Url:    "/users",
			QueryParam: []echotest.Param{
				{Name: "page", Value: "2"},
				{Name: "size", Value: "10"},
			},
		},
	}
	ctx, _ := echotest.PrepareContext(e, tc)
	if ctx.QueryParam("page") != "2" {
		t.Errorf("expected page=2, got %q", ctx.QueryParam("page"))
	}
	if ctx.QueryParam("size") != "10" {
		t.Errorf("expected size=10, got %q", ctx.QueryParam("size"))
	}
}

func TestPrepareContext_CustomHeaders(t *testing.T) {
	e := echo.New()
	tc := echotest.Case{
		Request: echotest.Request{
			Method: http.MethodPost,
			Url:    "/orders",
			Header: map[string]string{"X-User-Code": "user-abc"},
		},
	}
	ctx, _ := echotest.PrepareContext(e, tc)
	if ctx.Request().Header.Get("X-User-Code") != "user-abc" {
		t.Error("expected X-User-Code header")
	}
}

func TestPrepareContext_JSONBody(t *testing.T) {
	e := echo.New()
	tc := echotest.Case{
		Request:     echotest.Request{Method: http.MethodPost, Url: "/orders"},
		RequestBody: map[string]string{"key": "value"},
	}
	ctx, _ := echotest.PrepareContext(e, tc)
	var body map[string]string
	if err := ctx.Bind(&body); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("expected key=value, got %v", body)
	}
}

func TestPrepareContext_StringBody(t *testing.T) {
	e := echo.New()
	tc := echotest.Case{
		Request:     echotest.Request{Method: http.MethodPost, Url: "/raw"},
		RequestBody: `{"raw":true}`,
	}
	ctx, _ := echotest.PrepareContext(e, tc)
	var body map[string]any
	if err := ctx.Bind(&body); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	if body["raw"] != true {
		t.Errorf("expected raw=true, got %v", body)
	}
}

func TestToJSONString(t *testing.T) {
	s := echotest.ToJSONString(map[string]string{"key": "value"})
	if s == nil {
		t.Fatal("expected non-nil string")
	}
	if *s != `{"key":"value"}` {
		t.Errorf("unexpected JSON: %s", *s)
	}
}

func TestToJSONString_NilOnError(t *testing.T) {
	// channels cannot be marshalled
	s := echotest.ToJSONString(make(chan int))
	if s != nil {
		t.Error("expected nil for unmarshalable value")
	}
}

func TestNewRequest(t *testing.T) {
	req := echotest.NewRequest(http.MethodPost, "/test", map[string]string{"x": "y"})
	if req.Method != http.MethodPost {
		t.Errorf("unexpected method: %s", req.Method)
	}
	if req.Header.Get("Content-Type") != echo.MIMEApplicationJSON {
		t.Error("expected Content-Type application/json")
	}
}
