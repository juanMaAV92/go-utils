package platform

import (
	"time"

	"github.com/labstack/echo/v4"
)

type Server struct {
	Instance *echo.Echo
	*BasicConfig
}

type BasicConfig struct {
	Port          string
	GracefullTime time.Duration
	Environment   string
	ServerName    string
}
