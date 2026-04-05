package s3

import (
	"fmt"

	"github.com/juanmaAV/go-utils/env"
)

// Config holds the configuration for the S3 client.
type Config struct {
	Region   string
	Endpoint string // empty = real AWS; set for LocalStack or custom endpoints
}

// ConfigFromEnv reads S3 configuration from environment variables.
// prefix is prepended to each variable name with an underscore separator.
//
//	ConfigFromEnv("S3")          → S3_REGION, S3_ENDPOINT
//	ConfigFromEnv("MEDIA_S3")    → MEDIA_S3_REGION, MEDIA_S3_ENDPOINT
//
// Required: {prefix}_REGION
// Optional: {prefix}_ENDPOINT (empty = real AWS)
func ConfigFromEnv(prefix string) (Config, error) {
	p := prefix + "_"
	cfg := Config{
		Region:   env.GetEnv(p + "REGION"),
		Endpoint: env.GetEnv(p + "ENDPOINT"),
	}
	if cfg.Region == "" {
		return Config{}, fmt.Errorf("s3: missing required env var: %s", p+"REGION")
	}
	return cfg, nil
}
