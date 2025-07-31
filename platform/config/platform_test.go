package config

import (
	"os"
	"testing"
	"time"

	"github.com/juanMaAV92/go-utils/env"
)

func Test_GetBasicServerConfigt(t *testing.T) {

	os.Setenv(env.Port, "8080")
	defer os.Unsetenv(env.Port)

	os.Setenv(env.GracefulTime, "10m")
	defer os.Unsetenv(env.GracefulTime)

	os.Setenv(env.Enviroment, "dev")
	defer os.Unsetenv(env.Enviroment)

	config := GetBasicServerConfig("test")

	if config.Environment != "dev" {
		t.Errorf("Expected Environment to be 'dev', got %s", config.Environment)
	}
	if config.Port != "8080" {
		t.Errorf("Expected Port to be '8080', got %s", config.Port)
	}
	if config.GracefullTime != 10*time.Minute {
		t.Errorf("Expected GracefullTime to be 10m, got %v", config.GracefullTime)
	}
	if config.ServerName != "test" {
		t.Errorf("Expected AppName to be 'test', got %s", config.ServerName)
	}

}
