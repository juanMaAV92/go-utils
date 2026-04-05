package sns

import (
	"fmt"

	"github.com/juanmaAV/go-utils/env"
)

// Config holds the base SNS configuration.
type Config struct {
	Region   string
	Endpoint string // empty = real AWS; set for LocalStack
}

// ConfigFromEnv reads base SNS configuration from environment variables.
//
//	ConfigFromEnv("SNS")       → SNS_REGION, SNS_ENDPOINT
//	ConfigFromEnv("ORDER_SNS") → ORDER_SNS_REGION, ORDER_SNS_ENDPOINT
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
		return Config{}, fmt.Errorf("sns: missing required env var: %s", p+"REGION")
	}
	return cfg, nil
}
