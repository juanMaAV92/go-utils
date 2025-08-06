package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/juanMaAV92/go-utils/env"
)

type jwtConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
	SigningMethod   jwt.SigningMethod
}

func GetJWTConfig(issuer string, signingMethod jwt.SigningMethod) *jwtConfig {
	return &jwtConfig{
		SecretKey:       env.GetEnv(env.JWTSecretKey),
		AccessTokenTTL:  env.GetEnvAsDurationWithDefault(env.JWTAccessTokenTTL, "15m"),
		RefreshTokenTTL: env.GetEnvAsDurationWithDefault(env.JWTRefreshTokenTTL, "24h"),
		Issuer:          issuer,
		SigningMethod:   signingMethod,
	}
}
