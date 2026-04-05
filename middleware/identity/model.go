package identity

// superAdminPermission grants access to all capabilities.
const superAdminPermission = "all:all"

// HeaderConfig defines which HTTP headers map to identity fields.
// All fields have sensible defaults — only set what differs from the defaults.
//
//	identity.HeaderConfig{
//	    Extra: []string{"X-User-Nature", "X-Tenant-ID"},
//	}
type HeaderConfig struct {
	// UserCode is the header carrying the user's unique identifier.
	// Required — Middleware returns 401 if this header is missing.
	// Default: "X-User-Code"
	UserCode string

	// Email is the header carrying the user's email address.
	// Default: "X-User-Email"
	Email string

	// Roles is the header carrying comma-separated role names.
	// Default: "X-User-Roles"
	Roles string

	// Permissions is the header carrying comma-separated permission codes.
	// Default: "X-User-Permissions"
	Permissions string

	// Extra lists additional headers to capture.
	// Their values are stored in Identity.Attributes keyed by header name.
	// Example: []string{"X-User-Nature", "X-Hierarchy-Path", "X-Abac-JWT"}
	Extra []string
}

// Identity holds the user's identification and authorization data
// extracted from HTTP headers.
type Identity struct {
	UserCode    string
	Email       string
	Roles       []string
	Permissions []string
	// Attributes holds any extra headers configured via HeaderConfig.Extra,
	// keyed by the header name (e.g. "X-User-Nature" → "INDIVIDUAL").
	Attributes map[string]string
}

func (cfg HeaderConfig) withDefaults() HeaderConfig {
	if cfg.UserCode == "" {
		cfg.UserCode = "X-User-Code"
	}
	if cfg.Email == "" {
		cfg.Email = "X-User-Email"
	}
	if cfg.Roles == "" {
		cfg.Roles = "X-User-Roles"
	}
	if cfg.Permissions == "" {
		cfg.Permissions = "X-User-Permissions"
	}
	return cfg
}
