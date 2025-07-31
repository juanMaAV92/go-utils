package config

import (
	"time"
)

type BasicConfig struct {
	Port          string
	GracefullTime time.Duration
	Environment   string
	ServerName    string
}
