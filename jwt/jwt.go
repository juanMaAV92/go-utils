package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JWTConfig *JwtConfig

func InitJWTConfig(cfg *JwtConfig) {
	JWTConfig = cfg
}

func generateToken(userCode uuid.UUID, ttl time.Duration, tokenType string) (string, error) {
	if JWTConfig.SigningMethod == nil {
		return "", errors.New("JWTConfig or SigningMethod not initialized")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"user_code": userCode.String(),
		"exp":       now.Add(ttl).Unix(),
		"iat":       now.Unix(),
		"iss":       JWTConfig.Issuer,
		"type":      tokenType,
	}
	token := jwt.NewWithClaims(JWTConfig.SigningMethod, claims)
	return token.SignedString([]byte(JWTConfig.SecretKey))
}

func GenerateAccessToken(userCode uuid.UUID) (string, error) {
	if JWTConfig == nil {
		return "", errors.New("JWTConfig not initialized")
	}
	return generateToken(userCode, JWTConfig.AccessTokenTTL, "access")
}

func GenerateRefreshToken(userCode uuid.UUID) (string, error) {
	if JWTConfig == nil {
		return "", errors.New("JWTConfig not initialized")
	}
	return generateToken(userCode, JWTConfig.RefreshTokenTTL, "refresh")
}

func parseToken(tokenString string) (*jwt.Token, error) {
	if JWTConfig == nil || JWTConfig.SigningMethod == nil {
		return nil, errors.New("JWTConfig or SigningMethod not initialized")
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(JWTConfig.SecretKey), nil
	}

	return jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, keyFunc, jwt.WithoutClaimsValidation())
}

func ParseClaims(tokenString string) (jwt.MapClaims, bool, error) {
	token, err := parseToken(tokenString)
	if err != nil {
		return nil, false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, false, errors.New("invalid claims")
	}

	// Validar manualmente la expiración
	isValid := true
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			isValid = false
		}
	}

	return claims, isValid, nil
}
