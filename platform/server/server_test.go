package server

import (
	"context"
	"testing"
	"time"

	"github.com/juanMaAV92/go-utils/platform/config"
	"github.com/labstack/echo/v4"
)

func Test_server(t *testing.T) {

	config := &config.BasicConfig{
		Port:         "8080",
		GracefulTime: 5 * time.Second,
	}

	server := Server{
		Echo: echo.New(),
	}

	_ = server.Run(config.Port, config.GracefulTime)

	time.Sleep(2 * time.Second)

	shutdownErr := server.Echo.Shutdown(context.Background())
	if shutdownErr != nil {
		t.Fatalf("Failed to shutdown server: %v", shutdownErr)
	}

}
