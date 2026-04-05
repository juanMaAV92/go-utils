# middleware/identity

Echo middleware for propagating user identity across microservices via HTTP headers.

## How it works

An API Gateway (Kong, Envoy, AWS API GW) validates the JWT and injects user identity as HTTP headers before forwarding the request. This middleware reads those headers, builds an `Identity`, and stores it in the request context.

Downstream services receiving calls from this service get the same headers forwarded via `ToRequestOptions`.

```
API Gateway
  → validates JWT, injects X-User-* headers
  → Service A: Middleware() extracts identity into ctx
    → handler checks HasPermission(ctx, "orders:create")
    → calls Service B: httpclient.Post(ctx, url, body, identity.ToRequestOptions(ctx)...)
      → Service B receives the same X-User-* headers
```

## Setup

```go
import "github.com/juanmaAV/go-utils/middleware/identity"

e.Use(identity.Middleware(identity.HeaderConfig{
    Extra: []string{"X-User-Nature", "X-Hierarchy-Path", "X-Abac-JWT"},
}))
```

All `HeaderConfig` fields have defaults — only set what differs:

| Field | Default | Description |
|---|---|---|
| `UserCode` | `X-User-Code` | Required — 401 if missing |
| `Email` | `X-User-Email` | User's email address |
| `Roles` | `X-User-Roles` | Comma-separated role names |
| `Permissions` | `X-User-Permissions` | Comma-separated permission codes |
| `Extra` | `[]` | Additional headers → stored in `Identity.Attributes` |

### Custom header names

```go
identity.Middleware(identity.HeaderConfig{
    UserCode: "Authorization-ID",
    Roles:    "X-Roles",
})
```

## RBAC helpers

```go
// Permission check (supports "all:all" superadmin)
identity.HasPermission(ctx, "orders:create")
identity.HasAnyPermission(ctx, "orders:create", "orders:manage")

// Role check
identity.HasRole(ctx, "admin", "manager")

// Wildcard permission filter
identity.FilterPermissions(ctx, "orders:*")            // → []string{"orders:read", "orders:create"}
identity.GetFilteredPermissionsString(ctx, "orders:*") // → "orders:read,orders:create"
```

A user with permission `"all:all"` always passes `HasPermission` and is always included in `FilterPermissions`.

## Enforce permission on a route

```go
e.DELETE("/users/:id", deleteUser, identity.RequireCapability("users:delete"))
```

Returns 403 if the identity in the context does not have the required permission.

## Read identity in handlers

```go
id, ok := identity.FromContext(ctx)

identity.GetUserCode(ctx)
identity.GetEmail(ctx)
identity.GetRoles(ctx)
identity.GetPermissions(ctx)

// Extra headers configured via HeaderConfig.Extra
identity.GetAttribute(ctx, "X-User-Nature")
identity.GetAttribute(ctx, "X-Hierarchy-Path")
```

## Propagate identity to downstream services

```go
// Forward all identity headers
opts := identity.ToRequestOptions(ctx)
httpClient.Post(ctx, url, body, opts...)

// Forward only the permissions relevant to the downstream service
opts := identity.ToRequestOptions(ctx, "inventory:*")
httpClient.Post(ctx, url, body, opts...)
```

`ToRequestOptions` re-emits the same headers the middleware received, including any `Extra` attributes with non-empty values.

## Testing

Set identity manually in tests without running the middleware:

```go
id := &identity.Identity{
    UserCode:    "user-123",
    Permissions: []string{"orders:create"},
    Attributes:  map[string]string{"X-User-Nature": "INDIVIDUAL"},
}
ctx := identity.WithIdentity(context.Background(), id)
```
