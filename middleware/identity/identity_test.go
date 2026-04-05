package identity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// --- helpers ---

func ctxWithIdentity(perms, roles []string) context.Context {
	id := &Identity{
		UserCode:    "user-123",
		Email:       "user@example.com",
		Roles:       roles,
		Permissions: perms,
		Attributes:  map[string]string{},
	}
	return WithIdentity(context.Background(), id)
}

// --- HasPermission ---

func TestHasPermission_Match(t *testing.T) {
	ctx := ctxWithIdentity([]string{"users:read", "users:write"}, nil)
	if !HasPermission(ctx, "users:read") {
		t.Error("expected HasPermission to return true")
	}
}

func TestHasPermission_NoMatch(t *testing.T) {
	ctx := ctxWithIdentity([]string{"users:read"}, nil)
	if HasPermission(ctx, "users:delete") {
		t.Error("expected HasPermission to return false")
	}
}

func TestHasPermission_SuperAdmin(t *testing.T) {
	ctx := ctxWithIdentity([]string{"all:all"}, nil)
	if !HasPermission(ctx, "anything:at:all") {
		t.Error("superadmin should have all permissions")
	}
}

func TestHasPermission_NoIdentity(t *testing.T) {
	if HasPermission(context.Background(), "users:read") {
		t.Error("expected false when no identity in context")
	}
}

// --- HasAnyPermission ---

func TestHasAnyPermission_OneMatches(t *testing.T) {
	ctx := ctxWithIdentity([]string{"users:read"}, nil)
	if !HasAnyPermission(ctx, "events:read", "users:read") {
		t.Error("expected HasAnyPermission to return true")
	}
}

func TestHasAnyPermission_NoneMatch(t *testing.T) {
	ctx := ctxWithIdentity([]string{"users:read"}, nil)
	if HasAnyPermission(ctx, "events:read", "events:write") {
		t.Error("expected HasAnyPermission to return false")
	}
}

// --- HasRole ---

func TestHasRole_Match(t *testing.T) {
	ctx := ctxWithIdentity(nil, []string{"ADMIN", "EDITOR"})
	if !HasRole(ctx, "ADMIN") {
		t.Error("expected HasRole to return true")
	}
}

func TestHasRole_NoMatch(t *testing.T) {
	ctx := ctxWithIdentity(nil, []string{"EDITOR"})
	if HasRole(ctx, "ROOT") {
		t.Error("expected HasRole to return false")
	}
}

func TestHasRole_MultipleOneMatches(t *testing.T) {
	ctx := ctxWithIdentity(nil, []string{"ADMIN", "EDITOR"})
	if !HasRole(ctx, "ROOT", "ADMIN") {
		t.Error("expected HasRole to return true when one of many matches")
	}
}

func TestHasRole_EmptyCheck(t *testing.T) {
	ctx := ctxWithIdentity(nil, []string{"ADMIN"})
	if HasRole(ctx) {
		t.Error("expected HasRole to return false for empty roles check")
	}
}

func TestHasRole_NoIdentity(t *testing.T) {
	if HasRole(context.Background(), "ADMIN") {
		t.Error("expected false when no identity in context")
	}
}

// --- FilterPermissions ---

func TestFilterPermissions_WildcardSinglePattern(t *testing.T) {
	ctx := ctxWithIdentity([]string{"users:read", "users:write", "events:read"}, nil)
	result := FilterPermissions(ctx, "users:*")
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(result), result)
	}
	if result[0] != "users:read" || result[1] != "users:write" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestFilterPermissions_MultiplePatterns(t *testing.T) {
	ctx := ctxWithIdentity([]string{"users:read", "events:write", "orders:read"}, nil)
	result := FilterPermissions(ctx, "users:*", "events:*")
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestFilterPermissions_ExactMatch(t *testing.T) {
	ctx := ctxWithIdentity([]string{"users:read", "users:write"}, nil)
	result := FilterPermissions(ctx, "users:read")
	if len(result) != 1 || result[0] != "users:read" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestFilterPermissions_NoMatch(t *testing.T) {
	ctx := ctxWithIdentity([]string{"users:read"}, nil)
	result := FilterPermissions(ctx, "events:*")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestFilterPermissions_SuperAdminAlwaysIncluded(t *testing.T) {
	ctx := ctxWithIdentity([]string{"all:all", "users:read"}, nil)
	result := FilterPermissions(ctx, "users:*")
	if len(result) != 2 || result[0] != "all:all" {
		t.Errorf("expected superadmin first, got %v", result)
	}
}

func TestFilterPermissions_SuperAdminNoPatternMatch(t *testing.T) {
	ctx := ctxWithIdentity([]string{"all:all", "users:read"}, nil)
	result := FilterPermissions(ctx, "events:*")
	if len(result) != 1 || result[0] != "all:all" {
		t.Errorf("expected only superadmin, got %v", result)
	}
}

func TestFilterPermissions_EmptyPermissions(t *testing.T) {
	ctx := ctxWithIdentity([]string{}, nil)
	if FilterPermissions(ctx, "users:*") != nil {
		t.Error("expected nil for empty permissions")
	}
}

// --- GetFilteredPermissionsString ---

func TestGetFilteredPermissionsString(t *testing.T) {
	ctx := ctxWithIdentity([]string{"all:all", "users:read", "users:write", "events:write"}, nil)
	got := GetFilteredPermissionsString(ctx, "users:*")
	want := "all:all,users:read,users:write"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- GetAttribute ---

func TestGetAttribute(t *testing.T) {
	id := &Identity{
		UserCode:   "user-123",
		Attributes: map[string]string{"X-User-Nature": "INDIVIDUAL"},
	}
	ctx := WithIdentity(context.Background(), id)
	if got := GetAttribute(ctx, "X-User-Nature"); got != "INDIVIDUAL" {
		t.Errorf("got %q, want INDIVIDUAL", got)
	}
}

func TestGetAttribute_Missing(t *testing.T) {
	id := &Identity{UserCode: "user-123", Attributes: map[string]string{}}
	ctx := WithIdentity(context.Background(), id)
	if got := GetAttribute(ctx, "X-Unknown"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGetAttribute_NoIdentity(t *testing.T) {
	if got := GetAttribute(context.Background(), "X-User-Nature"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// --- Middleware ---

func newEchoCtx(headers map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestMiddleware_Success(t *testing.T) {
	c, rec := newEchoCtx(map[string]string{
		"X-User-Code":        "user-123",
		"X-User-Email":       "user@example.com",
		"X-User-Roles":       "admin, editor",
		"X-User-Permissions": "users:read, users:write",
	})

	mw := Middleware(HeaderConfig{})
	called := false
	h := mw(func(c echo.Context) error {
		called = true
		id, ok := FromContext(c.Request().Context())
		if !ok {
			t.Fatal("identity not in context")
		}
		if id.UserCode != "user-123" {
			t.Errorf("unexpected UserCode: %s", id.UserCode)
		}
		if len(id.Roles) != 2 {
			t.Errorf("expected 2 roles, got %d", len(id.Roles))
		}
		if len(id.Permissions) != 2 {
			t.Errorf("expected 2 permissions, got %d", len(id.Permissions))
		}
		return c.String(http.StatusOK, "ok")
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_MissingUserCode(t *testing.T) {
	c, _ := newEchoCtx(map[string]string{})
	mw := Middleware(HeaderConfig{})
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	err := h(c)
	if err == nil {
		t.Fatal("expected error")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %v", err)
	}
}

func TestMiddleware_ExtraHeaders(t *testing.T) {
	c, _ := newEchoCtx(map[string]string{
		"X-User-Code":    "user-123",
		"X-User-Nature":  "INDIVIDUAL",
		"X-Hierarchy-Path": "/org/dept/123",
	})

	mw := Middleware(HeaderConfig{
		Extra: []string{"X-User-Nature", "X-Hierarchy-Path"},
	})
	h := mw(func(c echo.Context) error {
		nature := GetAttribute(c.Request().Context(), "X-User-Nature")
		if nature != "INDIVIDUAL" {
			t.Errorf("expected INDIVIDUAL, got %q", nature)
		}
		path := GetAttribute(c.Request().Context(), "X-Hierarchy-Path")
		if path != "/org/dept/123" {
			t.Errorf("unexpected path: %q", path)
		}
		return c.String(http.StatusOK, "ok")
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMiddleware_CustomHeaderNames(t *testing.T) {
	c, _ := newEchoCtx(map[string]string{
		"Authorization-ID": "user-abc",
		"X-Roles":          "viewer",
	})

	mw := Middleware(HeaderConfig{
		UserCode: "Authorization-ID",
		Roles:    "X-Roles",
	})
	h := mw(func(c echo.Context) error {
		if GetUserCode(c.Request().Context()) != "user-abc" {
			t.Errorf("unexpected UserCode")
		}
		if !HasRole(c.Request().Context(), "viewer") {
			t.Errorf("expected viewer role")
		}
		return c.String(http.StatusOK, "ok")
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- RequireCapability ---

func TestRequireCapability_Allowed(t *testing.T) {
	c, rec := newEchoCtx(map[string]string{
		"X-User-Code":        "user-123",
		"X-User-Permissions": "users:delete",
	})

	mw := Middleware(HeaderConfig{})
	cap := RequireCapability("users:delete")
	h := mw(cap(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}))

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireCapability_Forbidden(t *testing.T) {
	c, _ := newEchoCtx(map[string]string{
		"X-User-Code":        "user-123",
		"X-User-Permissions": "users:read",
	})

	mw := Middleware(HeaderConfig{})
	cap := RequireCapability("users:delete")
	h := mw(cap(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}))

	err := h(c)
	if err == nil {
		t.Fatal("expected 403 error")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %v", err)
	}
}

func TestRequireCapability_SuperAdmin(t *testing.T) {
	c, rec := newEchoCtx(map[string]string{
		"X-User-Code":        "admin",
		"X-User-Permissions": "all:all",
	})

	mw := Middleware(HeaderConfig{})
	cap := RequireCapability("anything:sensitive")
	h := mw(cap(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}))

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
