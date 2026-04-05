package sqs

import (
	"fmt"

	"github.com/juanMaAV92/go-utils/env"
)

// Config holds the base SQS configuration shared by producer and consumer.
type Config struct {
	Region   string
	Endpoint string // empty = real AWS; set for LocalStack
}

// ConfigFromEnv reads base SQS configuration from environment variables.
//
//	ConfigFromEnv("SQS")       → SQS_REGION, SQS_ENDPOINT
//	ConfigFromEnv("ORDER_SQS") → ORDER_SQS_REGION, ORDER_SQS_ENDPOINT
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
		return Config{}, fmt.Errorf("sqs: missing required env var: %s", p+"REGION")
	}
	return cfg, nil
}
