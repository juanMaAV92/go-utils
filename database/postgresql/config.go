package postgresql

import (
	"fmt"
	"time"

	"github.com/juanMaAV92/go-utils/env"
)

// Config holds the configuration for a PostgreSQL connection pool.
type Config struct {
	Host        string
	Port        string        // default "5432"
	User        string
	Password    string
	Name        string
	SSLMode     string        // "disable" | "require" | "verify-ca" | "verify-full"; default "require"
	MaxPoolSize int           // max idle and open connections; default 2
	MaxLifeTime time.Duration // max connection lifetime; default 5m
	Verbose     bool          // false → silent; true → warn + slow query logging (≥200ms)
}

// ConfigFromEnv reads database configuration from environment variables.
// prefix is prepended to each variable name with an underscore separator.
//
//	ConfigFromEnv("DB")       → DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME …
//	ConfigFromEnv("ORDERS")   → ORDERS_HOST, ORDERS_PORT, ORDERS_USER …
//
// Required: {prefix}_HOST, {prefix}_USER, {prefix}_PASSWORD, {prefix}_NAME
// Optional: {prefix}_PORT (5432), {prefix}_SSLMODE (require),
//
//	{prefix}_MAX_POOL_SIZE (2), {prefix}_MAX_LIFE_TIME (5m)
func ConfigFromEnv(prefix string) (Config, error) {
	p := prefix + "_"
	cfg := Config{
		Host:        env.GetEnv(p + "HOST"),
		Port:        env.GetEnvWithDefault(p+"PORT", "5432"),
		User:        env.GetEnv(p + "USER"),
		Password:    env.GetEnv(p + "PASSWORD"),
		Name:        env.GetEnv(p + "NAME"),
		SSLMode:     env.GetEnvWithDefault(p+"SSLMODE", "require"),
		MaxPoolSize: env.GetEnvAsIntWithDefault(p+"MAX_POOL_SIZE", 2),
		MaxLifeTime: env.GetEnvAsDurationWithDefault(p+"MAX_LIFE_TIME", 5*time.Minute),
	}

	var missing []string
	for _, pair := range []struct{ key, val string }{
		{p + "HOST", cfg.Host},
		{p + "USER", cfg.User},
		{p + "PASSWORD", cfg.Password},
		{p + "NAME", cfg.Name},
	} {
		if pair.val == "" {
			missing = append(missing, pair.key)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("postgresql: missing required env vars: %v", missing)
	}
	return cfg, nil
}
