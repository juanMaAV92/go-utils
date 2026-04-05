package identity

import (
	"context"
	"strings"

	"github.com/juanmaAV/go-utils/httpclient"
)

type contextIdentityKey struct{}
type contextConfigKey struct{}

// WithIdentity returns a new context carrying the provided Identity.
// Use this in tests or when setting identity manually outside of Middleware.
func WithIdentity(ctx context.Context, id *Identity) context.Context {
	return context.WithValue(ctx, contextIdentityKey{}, id)
}

// FromContext extracts the Identity from the context.
func FromContext(ctx context.Context) (*Identity, bool) {
	id, ok := ctx.Value(contextIdentityKey{}).(*Identity)
	return id, ok
}

// GetUserCode returns the UserCode from the context, or empty string if not set.
func GetUserCode(ctx context.Context) string {
	id, ok := FromContext(ctx)
	if !ok {
		return ""
	}
	return id.UserCode
}

// GetEmail returns the Email from the context, or empty string if not set.
func GetEmail(ctx context.Context) string {
	id, ok := FromContext(ctx)
	if !ok {
		return ""
	}
	return id.Email
}

// GetRoles returns the Roles from the context, or nil if not set.
func GetRoles(ctx context.Context) []string {
	id, ok := FromContext(ctx)
	if !ok {
		return nil
	}
	return id.Roles
}

// GetPermissions returns the Permissions from the context, or nil if not set.
func GetPermissions(ctx context.Context) []string {
	id, ok := FromContext(ctx)
	if !ok {
		return nil
	}
	return id.Permissions
}

// GetAttribute returns an extra attribute from the context by header name.
// Returns empty string if the attribute or identity is not present.
//
//	nature := identity.GetAttribute(ctx, "X-User-Nature")
func GetAttribute(ctx context.Context, header string) string {
	id, ok := FromContext(ctx)
	if !ok {
		return ""
	}
	return id.Attributes[header]
}

// HasPermission reports whether the identity in the context has the given permission.
// A superadmin (permission "all:all") always returns true.
func HasPermission(ctx context.Context, permission string) bool {
	id, ok := FromContext(ctx)
	if !ok {
		return false
	}
	for _, p := range id.Permissions {
		if p == superAdminPermission {
			return true
		}
	}
	for _, p := range id.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission reports whether the identity has at least one of the given permissions.
func HasAnyPermission(ctx context.Context, perms ...string) bool {
	for _, p := range perms {
		if HasPermission(ctx, p) {
			return true
		}
	}
	return false
}

// HasRole reports whether the identity has any of the given roles.
func HasRole(ctx context.Context, roles ...string) bool {
	id, ok := FromContext(ctx)
	if !ok || len(roles) == 0 {
		return false
	}
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}
	for _, userRole := range id.Roles {
		if _, exists := roleSet[userRole]; exists {
			return true
		}
	}
	return false
}

// FilterPermissions returns the permissions from context that match any of the given patterns.
// Patterns support a trailing wildcard: "users:*" matches "users:read", "users:write", etc.
// A superadmin ("all:all") is always included when present.
func FilterPermissions(ctx context.Context, patterns ...string) []string {
	userPerms := GetPermissions(ctx)
	if len(userPerms) == 0 || len(patterns) == 0 {
		return nil
	}

	var filtered []string
	hasSuperAdmin := false
	for _, p := range userPerms {
		if p == superAdminPermission {
			hasSuperAdmin = true
			break
		}
	}
	if hasSuperAdmin {
		filtered = append(filtered, superAdminPermission)
	}

	for _, perm := range userPerms {
		if perm == superAdminPermission {
			continue
		}
		for _, pattern := range patterns {
			if matchPattern(perm, pattern) {
				filtered = append(filtered, perm)
				break
			}
		}
	}
	return filtered
}

// GetFilteredPermissionsString returns filtered permissions as a comma-separated string.
func GetFilteredPermissionsString(ctx context.Context, patterns ...string) string {
	filtered := FilterPermissions(ctx, patterns...)
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, ",")
}

// GetRolesString returns the roles from context as a comma-separated string.
func GetRolesString(ctx context.Context) string {
	roles := GetRoles(ctx)
	if len(roles) == 0 {
		return ""
	}
	return strings.Join(roles, ",")
}

// GetPermissionsString returns the permissions from context as a comma-separated string.
func GetPermissionsString(ctx context.Context) string {
	perms := GetPermissions(ctx)
	if len(perms) == 0 {
		return ""
	}
	return strings.Join(perms, ",")
}

// ToRequestOptions returns httpclient options that propagate the identity as headers
// to a downstream service. If permissionPatterns are provided, only matching permissions
// are forwarded.
//
//	opts := identity.ToRequestOptions(ctx)
//	opts := identity.ToRequestOptions(ctx, "orders:*", "inventory:*")
func ToRequestOptions(ctx context.Context, permissionPatterns ...string) []httpclient.RequestOption {
	id, ok := FromContext(ctx)
	if !ok {
		return nil
	}

	cfg := configFromContext(ctx)

	perms := GetPermissionsString(ctx)
	if len(permissionPatterns) > 0 {
		perms = GetFilteredPermissionsString(ctx, permissionPatterns...)
	}

	opts := []httpclient.RequestOption{
		httpclient.WithHeader(cfg.UserCode, id.UserCode),
		httpclient.WithHeader(cfg.Email, id.Email),
		httpclient.WithHeader(cfg.Roles, GetRolesString(ctx)),
		httpclient.WithHeader(cfg.Permissions, perms),
	}

	for header, value := range id.Attributes {
		if value != "" {
			opts = append(opts, httpclient.WithHeader(header, value))
		}
	}

	return opts
}

// matchPattern checks if a permission matches a pattern with wildcard support.
func matchPattern(permission, pattern string) bool {
	if permission == pattern {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(permission, prefix)
	}
	return false
}

// configFromContext retrieves the HeaderConfig stored by Middleware, falling back to defaults.
func configFromContext(ctx context.Context) HeaderConfig {
	if cfg, ok := ctx.Value(contextConfigKey{}).(HeaderConfig); ok {
		return cfg
	}
	return HeaderConfig{}.withDefaults()
}
