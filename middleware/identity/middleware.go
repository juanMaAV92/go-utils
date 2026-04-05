package identity

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
)

// Middleware returns an Echo middleware that extracts identity data from HTTP headers
// and stores it in the request context.
//
// Returns 401 if the UserCode header is missing.
//
//	e.Use(identity.Middleware(identity.HeaderConfig{
//	    Extra: []string{"X-User-Nature", "X-Hierarchy-Path"},
//	}))
func Middleware(cfg HeaderConfig) echo.MiddlewareFunc {
	cfg = cfg.withDefaults()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userCode := c.Request().Header.Get(cfg.UserCode)
			if userCode == "" {
				return echo.NewHTTPError(401, "missing user identity")
			}

			id := &Identity{
				UserCode:    userCode,
				Email:       c.Request().Header.Get(cfg.Email),
				Roles:       splitHeader(c.Request().Header.Get(cfg.Roles)),
				Permissions: splitHeader(c.Request().Header.Get(cfg.Permissions)),
				Attributes:  make(map[string]string, len(cfg.Extra)),
			}

			for _, header := range cfg.Extra {
				if v := c.Request().Header.Get(header); v != "" {
					id.Attributes[header] = v
				}
			}

			ctx := context.WithValue(
				context.WithValue(c.Request().Context(), contextIdentityKey{}, id),
				contextConfigKey{}, cfg,
			)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// RequireCapability returns an Echo middleware that enforces a permission check.
// Returns 403 if the identity in the context does not have the required permission.
//
//	e.DELETE("/users/:id", handler, identity.RequireCapability("users:delete"))
func RequireCapability(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !HasPermission(c.Request().Context(), permission) {
				return echo.NewHTTPError(403, "missing required capability")
			}
			return next(c)
		}
	}
}

func splitHeader(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
