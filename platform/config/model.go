package config

import (
	"time"
)

type BasicConfig struct {
	Port         string
	GracefulTime time.Duration
	Environment  string
	ServerName   string
}

type TelemetryConfig struct {
	OTLPEndpoint string
}
