package server

import (
	"github.com/juanMaAV92/go-utils/platform/config"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Instance *echo.Echo
	*config.BasicConfig
}
