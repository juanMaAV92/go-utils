package scheduler

import (
	"fmt"

	"github.com/juanMaAV92/go-utils/env"
)

// Config holds the configuration for the EventBridge Scheduler client.
type Config struct {
	Region   string
	Endpoint string // empty = real AWS; set for LocalStack
	RoleArn  string // IAM role the scheduler assumes when invoking targets
}

// ConfigFromEnv reads scheduler configuration from environment variables.
//
//	ConfigFromEnv("SCHEDULER") → SCHEDULER_REGION, SCHEDULER_ENDPOINT, SCHEDULER_ROLE_ARN
//
// Required: {prefix}_REGION, {prefix}_ROLE_ARN
// Optional: {prefix}_ENDPOINT (empty = real AWS)
func ConfigFromEnv(prefix string) (Config, error) {
	p := prefix + "_"
	cfg := Config{
		Region:   env.GetEnv(p + "REGION"),
		Endpoint: env.GetEnv(p + "ENDPOINT"),
		RoleArn:  env.GetEnv(p + "ROLE_ARN"),
	}
	var missing []string
	if cfg.Region == "" {
		missing = append(missing, p+"REGION")
	}
	if cfg.RoleArn == "" {
		missing = append(missing, p+"ROLE_ARN")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("scheduler: missing required env vars: %v", missing)
	}
	return cfg, nil
}
