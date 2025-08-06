# JWT Package

This package provides utilities for working with JSON Web Tokens (JWT) in Go, including token generation, validation, and claims parsing. It is designed to simplify authentication and authorization flows in your applications.

## Features

- Generate access and refresh tokens with custom TTLs
- Validate JWT tokens and parse claims
- Configurable signing method and issuer
- Built-in test helpers

## Usage

### Configuration

Before generating or validating tokens, initialize the JWT configuration:

```go
import "github.com/juanMaAV92/go-utils/jwt"
import "github.com/golang-jwt/jwt/v5"

cfg := &jwt.JWTConfig{
    SecretKey:       "your_secret_key",
    AccessTokenTTL:  15 * time.Minute,
    RefreshTokenTTL: 24 * time.Hour,
    Issuer:          "your_issuer",
    SigningMethod:   jwt.SigningMethodHS256,
}
jwt.InitJWTConfig(cfg)
```

### Generate Tokens

```go
userID := "user-uuid-string"
accessToken, err := jwt.GenerateAccessToken(userID)
refreshToken, err := jwt.GenerateRefreshToken(userID)
```

### Validate Token

```go
token, err := jwt.ValidateToken(accessToken)
if err != nil {
    // Handle invalid token
}
```

### Parse Claims

```go
claims, err := jwt.ParseClaims(accessToken)
if err != nil {
    // Handle error
}
userID := claims["user_code"]
```

## Testing

The package includes unit tests for all main functionalities. Run tests with:

```bash
go test ./jwt
```

## Requirements
- Go 1.18+
- github.com/golang-jwt/jwt/v5
- github.com/google/uuid
- github.com/stretchr/testify (for tests)

## License
MIT
