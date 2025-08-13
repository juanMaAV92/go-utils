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

func validateToken(tokenString string, shouldSkipValidation bool) (*jwt.Token, error) {
	if JWTConfig == nil || JWTConfig.SigningMethod == nil {
		return nil, errors.New("JWTConfig or SigningMethod not initialized")
	}

	var token *jwt.Token
	var err error

	if shouldSkipValidation {
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(JWTConfig.SecretKey), nil
		}, jwt.WithoutClaimsValidation())
	} else {
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(JWTConfig.SecretKey), nil
		})
		if err != nil {
			return nil, err
		}
		if !token.Valid {
			return nil, errors.New("invalid token")
		}
	}

	return token, err
}

func ParseClaims(tokenString string, skipValidation ...bool) (jwt.MapClaims, error) {
	if JWTConfig == nil || JWTConfig.SigningMethod == nil {
		return nil, errors.New("JWTConfig or SigningMethod not initialized")
	}

	shouldSkipValidation := false
	if len(skipValidation) > 0 {
		shouldSkipValidation = skipValidation[0]
	}

	token, err := validateToken(tokenString, shouldSkipValidation)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
