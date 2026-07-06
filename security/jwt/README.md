# security/jwt

JWT signing and validation using RSA (RS256). Generic — works with any claims struct you define.

## Why generics

`ValidateToken` returns your concrete claims type directly, with full type safety. No casting required.

## Usage

### 1. Define your claims

```go
type MyClaims struct {
    jwt.RegisteredClaims               // required: provides standard JWT fields
    UserID string   `json:"user_id"`
    Roles  []string `json:"roles"`
}
```

### 2. Create the service

```go
import "github.com/juanMaAV92/go-utils/security/jwt"

svc, err := jwt.New(privateKeyPEM, publicKeyPEM, "my-service")
```

Either key can be empty:
- omit `privateKeyPEM` for a validation-only instance (downstream services)
- omit `publicKeyPEM` for sign-only (uncommon)

### 3. Sign

```go
token, err := svc.GenerateToken(&MyClaims{
    RegisteredClaims: svc.RegisteredClaims(24 * time.Hour), // issuer + iat + exp
    UserID: "usr_123",
    Roles:  []string{"admin"},
})
```

`svc.RegisteredClaims(expiry)` is a helper that fills `Issuer`, `IssuedAt`, and `ExpiresAt`. You can set them manually instead.

### 4. Validate

```go
claims, err := jwt.ValidateToken[MyClaims, *MyClaims](svc, tokenString)
// claims is *MyClaims — no casting
```

### 5. Validate ignoring expiration (refresh flows)

Skips the time-based claims (`exp`/`nbf`/`iat`) but still enforces the signature,
the RS256 algorithm, and the issuer.

```go
claims, err := jwt.ValidateTokenIgnoringExpiration[MyClaims, *MyClaims](svc, tokenString)
```

## API

```go
func New(privateKeyPEM, publicKeyPEM, issuer string) (*TokenService, error)

func (s *TokenService) RegisteredClaims(expiry time.Duration) jwt.RegisteredClaims
func (s *TokenService) GenerateToken(claims jwt.Claims) (string, error)

func ValidateToken[T any, C interface{ *T; jwt.Claims }](
    s *TokenService, tokenString string,
) (C, error)

func ValidateTokenIgnoringExpiration[T any, C interface{ *T; jwt.Claims }](
    s *TokenService, tokenString string,
) (C, error)
```

## Security

- Algorithm is restricted to `RS256` on every validation (via `WithValidMethods`). Tokens signed with any other algorithm — including RS384/RS512 or HMAC — are rejected, preventing algorithm-substitution attacks.
- Expiration is required: a token without an `exp` claim is rejected (`ValidateToken`).
- Issuer is validated against the service's configured issuer when one is set.
- Keys are parsed at construction time; invalid PEM fails fast at startup.

## Dependencies

- `github.com/golang-jwt/jwt/v5`
