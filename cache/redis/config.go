package redis

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/juanmaAV/go-utils/env"
	goredis "github.com/redis/go-redis/v9"
)

// Config holds the configuration for a Redis client.
type Config struct {
	Host          string
	Port          string        // default "6379"
	Password      string        // empty = no auth
	DB            int           // logical database index; default 0
	TLS           bool          // enable TLS
	TLSServerName string        // TLS SNI; defaults to Host when empty
	ReadTimeout   time.Duration // default 3s
	WriteTimeout  time.Duration // default 3s
	KeyPrefix     string        // prepended to every key: "{prefix}:{key}"
}

// ConfigFromEnv reads Redis configuration from environment variables.
// prefix is prepended to each variable name with an underscore separator.
//
//	ConfigFromEnv("REDIS")        → REDIS_HOST, REDIS_PORT, REDIS_PASSWORD …
//	ConfigFromEnv("SESSION_REDIS") → SESSION_REDIS_HOST …
//
// Required: {prefix}_HOST
// Optional: {prefix}_PORT (6379), {prefix}_PASSWORD, {prefix}_DB (0),
//
//	{prefix}_TLS (false), {prefix}_TLS_SERVER_NAME,
//	{prefix}_READ_TIMEOUT (3s), {prefix}_WRITE_TIMEOUT (3s),
//	{prefix}_KEY_PREFIX
func ConfigFromEnv(prefix string) (Config, error) {
	p := prefix + "_"
	cfg := Config{
		Host:          env.GetEnv(p + "HOST"),
		Port:          env.GetEnvWithDefault(p+"PORT", "6379"),
		Password:      env.GetEnv(p + "PASSWORD"),
		DB:            env.GetEnvAsIntWithDefault(p+"DB", 0),
		TLS:           env.GetEnvWithDefault(p+"TLS", "false") == "true",
		TLSServerName: env.GetEnv(p + "TLS_SERVER_NAME"),
		ReadTimeout:   env.GetEnvAsDurationWithDefault(p+"READ_TIMEOUT", 3*time.Second),
		WriteTimeout:  env.GetEnvAsDurationWithDefault(p+"WRITE_TIMEOUT", 3*time.Second),
		KeyPrefix:     env.GetEnv(p + "KEY_PREFIX"),
	}

	if cfg.Host == "" {
		return Config{}, fmt.Errorf("redis: missing required env var: %s", p+"HOST")
	}
	return cfg, nil
}

func (c Config) toRedisOptions() *goredis.Options {
	opts := &goredis.Options{
		Addr:         c.Host + ":" + c.Port,
		Password:     c.Password,
		DB:           c.DB,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
	}
	if c.TLS {
		serverName := c.TLSServerName
		if serverName == "" {
			serverName = c.Host
		}
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: serverName,
		}
	}
	return opts
}
