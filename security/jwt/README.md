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
import "github.com/juanmaAV/go-utils/security/jwt"

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

- Algorithm is verified on every validation (`RS256` only). Tokens signed with other algorithms are rejected — prevents algorithm substitution attacks.
- Keys are parsed at construction time; invalid PEM fails fast at startup.

## Dependencies

- `github.com/golang-jwt/jwt/v5`
