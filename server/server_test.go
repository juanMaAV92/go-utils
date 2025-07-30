package platform

import (
	"context"
	"testing"
	"time"

	"github.com/juanMaAV92/go-utils/env"
)

func Test_server(t *testing.T) {

	config := &BasicConfig{
		Port:          "8080",
		GracefullTime: 5 * time.Second,
		Environment:   env.LocalEnvironment,
		ServerName:    "test-app",
	}

	server, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	_ = server.Run()

	time.Sleep(2 * time.Second)

	shutdownErr := server.Instance.Shutdown(context.Background())
	if shutdownErr != nil {
		t.Fatalf("Failed to shutdown server: %v", shutdownErr)
	}

}
