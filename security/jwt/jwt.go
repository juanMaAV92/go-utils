package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService handles JWT signing and validation using RSA (RS256).
// Use the package-level generic functions ValidateToken and ValidateTokenIgnoringExpiration
// to get typed claims back.
type TokenService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
}

// New creates a TokenService. Either key can be empty:
//   - omit privateKeyPEM for a validation-only service (e.g. downstream API)
//   - omit publicKeyPEM for a signing-only service (uncommon)
func New(privateKeyPEM, publicKeyPEM, issuer string) (*TokenService, error) {
	svc := &TokenService{issuer: issuer}

	if privateKeyPEM != "" {
		pk, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		svc.privateKey = pk
	}

	if publicKeyPEM != "" {
		pk, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
		svc.publicKey = pk
	}

	return svc, nil
}

// RegisteredClaims returns a jwt.RegisteredClaims pre-filled with the service issuer,
// current issued-at, and the given expiry duration. Use it as the embedded field
// when constructing your custom claims before calling GenerateToken.
func (s *TokenService) RegisteredClaims(expiry time.Duration) jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		Issuer:    s.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
	}
}

// GenerateToken signs a JWT with the provided claims using RS256.
// The caller is responsible for setting RegisteredClaims fields (issuer, expiry, etc.).
// Use TokenService.RegisteredClaims(expiry) as a convenience helper.
func (s *TokenService) GenerateToken(claims jwt.Claims) (string, error) {
	if s.privateKey == nil {
		return "", fmt.Errorf("private key required for signing")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// ValidateToken parses and validates tokenString, returning typed claims.
// C must be a pointer to a struct that embeds jwt.RegisteredClaims.
//
// Usage:
//
//	claims, err := jwt.ValidateToken[MyClaims](svc, tokenString)
func ValidateToken[T any, C interface {
	*T
	jwt.Claims
}](s *TokenService, tokenString string) (C, error) {
	return parseToken[T, C](s, tokenString)
}

// ValidateTokenIgnoringExpiration validates the signature but skips expiration.
// Useful for refresh token flows where the expired token is used to obtain a new one.
func ValidateTokenIgnoringExpiration[T any, C interface {
	*T
	jwt.Claims
}](s *TokenService, tokenString string) (C, error) {
	return parseToken[T, C](s, tokenString, jwt.WithoutClaimsValidation())
}

func parseToken[T any, C interface {
	*T
	jwt.Claims
}](s *TokenService, tokenString string, opts ...jwt.ParserOption) (C, error) {
	var zero C
	if s.publicKey == nil {
		return zero, fmt.Errorf("public key required for validation")
	}
	claims := C(new(T))
	token, err := jwt.ParseWithClaims(tokenString, claims, s.keyFunc(), opts...)
	if err != nil {
		return zero, err
	}
	if !token.Valid {
		return zero, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (s *TokenService) keyFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	}
}
